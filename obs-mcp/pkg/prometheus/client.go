package prometheus

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promcfg "github.com/prometheus/common/config"
)

type PrometheusClient struct {
	client v1.API
}

func NewPrometheusClient(prometheusURL, token string) (*PrometheusClient, error) {
	if prometheusURL == "" {
		prometheusURL = "http://localhost:9090"
	}

	apiConfig := api.Config{
		Address: prometheusURL,
	}

	use_tls := strings.HasPrefix(prometheusURL, "https://")
	if use_tls {
		certs, err := createCertPool()
		if err != nil {
			return nil, err
		}

		defaultRt := api.DefaultRoundTripper.(*http.Transport)
		defaultRt.TLSClientConfig = &tls.Config{RootCAs: certs}

		apiConfig.RoundTripper = promcfg.NewAuthorizationCredentialsRoundTripper(
			"Bearer", promcfg.NewInlineSecret(token), defaultRt)
	} else {
		slog.Warn("Connecting to Prometheus without TLS")
	}

	client, err := api.NewClient(apiConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating prometheus client: %w", err)
	}

	v1api := v1.NewAPI(client)
	return &PrometheusClient{client: v1api}, nil
}

func (p *PrometheusClient) ListMetrics(ctx context.Context) ([]string, error) {
	labelValues, _, err := p.client.LabelValues(ctx, "__name__", []string{}, time.Now().Add(-time.Hour), time.Now())
	if err != nil {
		return nil, fmt.Errorf("error fetching metric names: %w", err)
	}

	metrics := make([]string, len(labelValues))
	for i, value := range labelValues {
		metrics[i] = string(value)
	}
	return metrics, nil
}

func (p *PrometheusClient) ExecuteRangeQuery(ctx context.Context, query string, start, end time.Time, step time.Duration) (map[string]interface{}, error) {
	r := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	result, warnings, err := p.client.QueryRange(ctx, query, r, v1.WithTimeout(30*time.Second))
	if err != nil {
		return nil, fmt.Errorf("error executing range query: %w", err)
	}

	response := map[string]interface{}{
		"resultType": "matrix",
		"result":     result,
	}

	if len(warnings) > 0 {
		response["warnings"] = warnings
	}

	return response, nil
}

func createCertPool() (*x509.CertPool, error) {
	certs := x509.NewCertPool()

	pemData, err := os.ReadFile(`/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt`)
	if err != nil {
		slog.Error("Failed to read the CA certificate", "err", err)
		return nil, err
	}
	certs.AppendCertsFromPEM(pemData)
	return certs, nil
}
