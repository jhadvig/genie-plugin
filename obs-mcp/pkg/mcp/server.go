package mcp

import (
	"github.com/inecas/obs-mcp/pkg/prometheus"
	"github.com/mark3labs/mcp-go/server"
)

func NewMCPServer(promClient *prometheus.PrometheusClient) (*server.MCPServer, error) {
	mcpServer := server.NewMCPServer(
		"obs-mcp",
		"1.0.0",
		server.WithLogging(),
		server.WithToolCapabilities(true),
	)

	if err := SetupTools(mcpServer, promClient); err != nil {
		return nil, err
	}

	return mcpServer, nil
}

func SetupTools(mcpServer *server.MCPServer, promClient *prometheus.PrometheusClient) error {
	// Create tool definitions
	listMetricsTool := CreateListMetricsTool()
	executeRangeQueryTool := CreateExecuteRangeQueryTool()

	// Create handlers
	listMetricsHandler := ListMetricsHandler(promClient)
	executeRangeQueryHandler := ExecuteRangeQueryHandler(promClient)

	// Add tools to server
	mcpServer.AddTool(listMetricsTool, listMetricsHandler)
	mcpServer.AddTool(executeRangeQueryTool, executeRangeQueryHandler)

	return nil
}
