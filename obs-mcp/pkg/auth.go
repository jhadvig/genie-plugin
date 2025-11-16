package pkg

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	promapi "github.com/prometheus/client_golang/api"
	promcfg "github.com/prometheus/common/config"
	"k8s.io/client-go/rest"
)

// AuthMode defines the authentication mode for Prometheus client
type AuthMode string

const (
	AuthModeKubeConfig     AuthMode = "kubeconfig"
	AuthModeServiceAccount AuthMode = "serviceaccount"
	AuthModeHeader         AuthMode = "header"
	AuthModeNone           AuthMode = "none"

	// AuthHeaderKey is the context key for the Kubernetes authorization header
	AuthHeaderKey string = "kubernetes-authorization"
)

// ParseAuthMode validates and converts a string to AuthMode
func ParseAuthMode(mode string) (AuthMode, error) {
	switch mode {
	case string(AuthModeKubeConfig):
		return AuthModeKubeConfig, nil
	case string(AuthModeServiceAccount):
		return AuthModeServiceAccount, nil
	case string(AuthModeHeader):
		return AuthModeHeader, nil
	case string(AuthModeNone):
		return AuthModeNone, nil
	default:
		return "", fmt.Errorf("invalid auth mode: %s (valid options: kubeconfig, serviceaccount, header, none)", mode)
	}
}

func createAPIConfig(ctx context.Context, opts ObsMCPOptions) (promapi.Config, error) {
	switch opts.AuthMode {
	case AuthModeKubeConfig:
		return createKubeconfigAPIConfig(opts)
	case AuthModeServiceAccount:
		return createServiceAccountAPIConfig(opts)
	case AuthModeHeader:
		return createHeaderAPIConfig(ctx, opts)
	case AuthModeNone:
		return createNoAuthAPIConfig(opts)
	default:
		return promapi.Config{}, fmt.Errorf("unsupported auth mode: %s", opts.AuthMode)
	}
}

func createKubeconfigAPIConfig(opts ObsMCPOptions) (promapi.Config, error) {
	restConfig, err := GetClientConfig()
	if err != nil {
		return promapi.Config{}, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	if restConfig.BearerToken == "" {
		return promapi.Config{}, fmt.Errorf("kubeconfig doesn't contain a bearer token for Prometheus authentication")
	}

	// Set TLS verification based on opts
	restConfig.TLSClientConfig.Insecure = opts.Insecure

	// Create HTTP client with kubeconfig authentication
	rt, err := rest.TransportFor(restConfig)
	if err != nil {
		return promapi.Config{}, fmt.Errorf("failed to create transport from kubeconfig: %w", err)
	}

	return promapi.Config{
		Address:      opts.PromURL,
		RoundTripper: rt,
	}, nil
}

func createServiceAccountAPIConfig(opts ObsMCPOptions) (promapi.Config, error) {
	slog.Info("Using service account token for authentication")
	tokenBytes, err := readTokenFromFile()
	if err != nil {
		slog.Error("Failed to read the service account token", "err", err)
		return promapi.Config{}, err
	}
	token := string(tokenBytes)

	return createAPIConfigWithToken(opts.PromURL, token, opts.Insecure)
}

func createHeaderAPIConfig(ctx context.Context, opts ObsMCPOptions) (promapi.Config, error) {
	token := getTokenFromCtx(ctx)
	if token == "" {
		slog.Warn("No token provided in context for header auth mode")
	}

	return createAPIConfigWithToken(opts.PromURL, token, opts.Insecure)
}

func createNoAuthAPIConfig(opts ObsMCPOptions) (promapi.Config, error) {
	slog.Info("Using no authentication (suitable for localhost port-forward)")
	return createAPIConfigWithToken(opts.PromURL, "", opts.Insecure)
}

func createAPIConfigWithToken(prometheusURL, token string, insecure bool) (promapi.Config, error) {
	apiConfig := promapi.Config{
		Address: prometheusURL,
	}

	useTLS := strings.HasPrefix(prometheusURL, "https://")
	if useTLS {
		defaultRt := promapi.DefaultRoundTripper.(*http.Transport)

		if insecure {
			defaultRt.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		} else {
			certs, err := createCertPool()
			if err != nil {
				return promapi.Config{}, err
			}
			defaultRt.TLSClientConfig = &tls.Config{RootCAs: certs}
		}

		if token != "" {
			apiConfig.RoundTripper = promcfg.NewAuthorizationCredentialsRoundTripper(
				"Bearer", promcfg.NewInlineSecret(token), defaultRt)
		} else {
			apiConfig.RoundTripper = defaultRt
		}
	} else {
		slog.Warn("Connecting to Prometheus without TLS")
	}

	return apiConfig, nil
}

func getTokenFromCtx(_ context.Context) string {

	// TODO: we're ignoring user auth for now, just to see if something improves
	return ""
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

func readTokenFromFile() ([]byte, error) {
	return os.ReadFile(`/var/run/secrets/kubernetes.io/serviceaccount/token`)
}
