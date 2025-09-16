package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/inecas/obs-mcp/pkg/http"
	"github.com/inecas/obs-mcp/pkg/mcp"
	"github.com/inecas/obs-mcp/pkg/prometheus"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Parse command line flags
	var listen = flag.String("listen", "", "Listen address for HTTP mode (e.g., :9100, 127.0.0.1:8080)")
	flag.Parse()

	// Get Prometheus URL from environment variable
	prometheusURL := os.Getenv("PROMETHEUS_URL")

	// Create Prometheus client
	promClient, err := prometheus.NewPrometheusClient(prometheusURL)
	if err != nil {
		log.Fatalf("Failed to create Prometheus client: %v", err)
	}

	// Create MCP server
	mcpServer, err := mcp.NewMCPServer(promClient)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Choose server mode based on flags
	if *listen != "" {
		// HTTP mode
		ctx := context.Background()
		if err := http.Serve(ctx, mcpServer, *listen); err != nil {
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
