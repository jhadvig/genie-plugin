package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// PromClient defines the interface for querying Prometheus
type PromClient interface {
	ListMetrics(ctx context.Context) ([]string, error)
	ExecuteRangeQuery(ctx context.Context, query string, start, end time.Time, step time.Duration) (map[string]interface{}, error)
	ExecuteInstantQuery(ctx context.Context, query string, time time.Time) (map[string]interface{}, error)
}

// PrometheusClient implements PromClient
type Client struct {
	client     v1.API
	guardrails *Guardrails
}

// Ensure PrometheusClient implements PromClient at compile time
var _ PromClient = (*Client)(nil)

func NewPrometheusClient(apiConfig api.Config) (*Client, error) {
	client, err := api.NewClient(apiConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating prometheus client: %w", err)
	}

	v1api := v1.NewAPI(client)
	return &Client{
		client:     v1api,
		guardrails: DefaultGuardrails(),
	}, nil
}

// WithGuardrails sets a custom Guardrails configuration for the client.
func (p *Client) WithGuardrails(g *Guardrails) *Client {
	p.guardrails = g
	return p
}

func (p *Client) ListMetrics(ctx context.Context) ([]string, error) {
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

func (p *Client) ExecuteRangeQuery(ctx context.Context, query string, start, end time.Time, step time.Duration) (map[string]interface{}, error) {
	if p.guardrails != nil {
		isSafe, err := p.guardrails.IsSafeQuery(query)
		if err != nil {
			return nil, fmt.Errorf("failed to parse query: %w", err)
		}

		if !isSafe {
			return nil, fmt.Errorf("query is not safe")
		}
	}

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
		"resultType": result.Type().String(),
		"result":     result,
	}

	if len(warnings) > 0 {
		response["warnings"] = warnings
	}

	return response, nil
}

func (p *Client) ExecuteInstantQuery(ctx context.Context, query string, time time.Time) (map[string]interface{}, error) {
	if p.guardrails != nil {
		isSafe, err := p.guardrails.IsSafeQuery(query)
		if err != nil {
			return nil, fmt.Errorf("failed to parse query: %w", err)
		}
		if !isSafe {
			return nil, fmt.Errorf("query is not safe")
		}
	}

	result, warnings, err := p.client.Query(ctx, query, time)
	if err != nil {
		return nil, fmt.Errorf("error executing instant query: %w", err)
	}

	response := map[string]interface{}{
		"resultType": result.Type().String(),
		"result":     result,
	}

	if len(warnings) > 0 {
		response["warnings"] = warnings
	}

	return response, nil
}
