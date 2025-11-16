package pkg

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	prometheusVersion     = "v3.7.3" // -ldflags "-X 'github.com/inecas/obs-mcp/pkg/mcp.prometheusVersion=...'"
	prometheusDocsBaseURL = "https://raw.githubusercontent.com/prometheus/prometheus/refs/tags/" + prometheusVersion + "/docs"
)

// PrometheusDocResource represents a Prometheus documentation file
type PrometheusDocResource struct {
	URI         string
	Name        string
	Description string
	FilePath    string // w.r.t. docs/ directory in Prometheus repo
}

func GetPrometheusDocResources() []PrometheusDocResource {
	return []PrometheusDocResource{
		{
			URI:         "prometheus://docs/querying-basics",
			Name:        "Prometheus Querying Basics",
			Description: "Introduction to querying Prometheus using PromQL",
			FilePath:    "querying/basics.md",
		},
		{
			URI:         "prometheus://docs/querying-operators",
			Name:        "Prometheus Query Operators",
			Description: "PromQL operators and their usage",
			FilePath:    "querying/operators.md",
		},
		{
			URI:         "prometheus://docs/querying-functions",
			Name:        "Prometheus Query Functions",
			Description: "Available functions in PromQL",
			FilePath:    "querying/functions.md",
		},
		{
			URI:         "prometheus://docs/querying-examples",
			Name:        "Prometheus Query Examples",
			Description: "Practical examples of PromQL queries",
			FilePath:    "querying/examples.md",
		},
		{
			URI:         "prometheus://docs/http-api",
			Name:        "Prometheus HTTP API",
			Description: "Complete reference for the Prometheus HTTP API",
			FilePath:    "querying/api.md",
		},
	}
}

func SetupResources(mcpServer *server.MCPServer) error {
	resources := GetPrometheusDocResources()
	for _, docResource := range resources {
		resource := mcp.NewResource(
			docResource.URI,
			docResource.Name,
			mcp.WithResourceDescription(docResource.Description),
			mcp.WithMIMEType("text/markdown"),
		)

		filePath := docResource.FilePath
		mcpServer.AddResource(resource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			slog.Info("Fetching resource", "filePath", filePath, "uri", request.Params.URI)

			return fetchPrometheusDoc(ctx, filePath, request.Params.URI)
		})
	}

	return nil
}

// TODO: Vendor a local copy of the docs (+a Makefile target to update it, based on version.GetVersion) instead of fetching over HTTP each time
func fetchPrometheusDoc(ctx context.Context, filePath, uri string) ([]mcp.ResourceContents, error) {
	url := fmt.Sprintf("%s/%s", prometheusDocsBaseURL, filePath)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch documentation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch documentation: status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read documentation: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/markdown",
			Text:     string(content),
		},
	}, nil
}
