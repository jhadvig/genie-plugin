package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/inecas/obs-mcp/pkg"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	var listen = flag.String("listen", "", "Listen address for HTTP mode")
	var authMode = flag.String("auth-mode", "", "Authentication mode: kubeconfig, serviceaccount, or header")
	var insecure = flag.Bool("insecure", false, "Skip TLS certificate verification")
	var enableAllEndpoints = flag.Bool("enable-all-endpoints", false, "Enable all Prometheus API endpoints (by default, only /api/v1/label/__name__/values and /api/v1/query_range are permitted)")
	var dev = flag.Bool("dev", false, "Development mode: target Prometheus instead of Thanos Querier (for single-node non-HA clusters)")
	flag.Parse()

	parsedAuthMode, err := pkg.ParseAuthMode(*authMode)
	if err != nil {
		log.Fatalf("Invalid auth mode: %v", err)
	}

	promURL := determinePrometheusURL(parsedAuthMode, *dev)

	var enabledEndpoints []string
	var endpointConfig *pkg.EndpointConfig
	if *enableAllEndpoints {
		endpointConfig = pkg.FullEndpointConfig()
		for path := range endpointConfig.AllowedEndpoints {
			enabledEndpoints = append(enabledEndpoints, path)
		}
		slog.Info("All endpoints enabled", "endpoints", enabledEndpoints)
	} else {
		endpointConfig = pkg.DefaultEndpointConfig()
		for path := range endpointConfig.AllowedEndpoints {
			enabledEndpoints = append(enabledEndpoints, path)
		}
		slog.Info("Default endpoints enabled", "endpoints", enabledEndpoints)
	}

	// Create MCP options
	opts := pkg.ObsMCPOptions{
		AuthMode:       parsedAuthMode,
		PromURL:        promURL,
		Insecure:       *insecure,
		EndpointConfig: endpointConfig,
	}

	// Create MCP server
	mcpServer, err := pkg.NewMCPServer(opts)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	slog.Info("Starting server",
		"PromURL", opts.PromURL,
		"AuthMode", opts.AuthMode,
		"Insecure", opts.Insecure,
		"Listen", *listen,
	)

	// Choose server mode based on flags
	if *listen != "" {
		// HTTP mode
		ctx := context.Background()
		if err := pkg.Serve(ctx, mcpServer, *listen); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	} else {
		// Start server on stdio (default mode)
		stdioServer := server.NewStdioServer(mcpServer)
		if err := stdioServer.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}
}

// determinePrometheusURL determines the Prometheus URL based on auth mode, dev mode, and environment
func determinePrometheusURL(authMode pkg.AuthMode, devMode bool) string {
	// Get Prometheus URL from environment variable
	promURL := os.Getenv("PROMETHEUS_URL")

	// If URL is provided, use it
	if promURL != "" {
		return promURL
	}

	// For kubeconfig mode, attempt service discovery
	if authMode == pkg.AuthModeKubeConfig {
		if devMode {
			// Dev mode: discover Prometheus in kind cluster
			slog.Info("Dev mode: attempting to discover Prometheus service in kind cluster")
			url, err := pkg.GetPrometheusURL()
			if err != nil {
				slog.Warn("Failed to discover Prometheus via kubeconfig, falling back to localhost:9090", "err", err)
				return "http://localhost:9090"
			}
			slog.Info("Discovered Prometheus URL", "url", url)
			return url
		} else {
			// Production mode: discover Thanos Querier in OpenShift
			slog.Info("No Prometheus URL provided, attempting to use kubeconfig to discover Thanos Querier")
			url, err := pkg.GetThanosQuerierURL()
			if err != nil {
				slog.Warn("Failed to discover Thanos Querier via kubeconfig, falling back to localhost:9090", "err", err)
				return "http://localhost:9090"
			}
			slog.Info("Discovered Thanos Querier URL", "url", url)
			return url
		}
	}

	// Default to localhost for all other auth modes
	return "http://localhost:9090"
}
