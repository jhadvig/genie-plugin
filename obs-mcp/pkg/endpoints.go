package pkg

import (
	"context"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/mark3labs/mcp-go/mcp"
)

// PrometheusEndpoint represents a configurable Prometheus API endpoint
type PrometheusEndpoint struct {
	// Path is the Prometheus API path (e.g., "/api/v1/query", "/api/v1/series")
	Path string
	// Description explains what this endpoint does
	Description string
	// Method is the HTTP method (GET or POST)
	Method string
	// Parameters defines the expected parameters for this endpoint
	Parameters []EndpointParameter
	// RequiresGuardrails indicates if this endpoint needs PromQL validation
	RequiresGuardrails bool
}

// EndpointParameter defines a parameter for a Prometheus endpoint
type EndpointParameter struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

// EndpointConfig holds the configuration for which endpoints are allowed
type EndpointConfig struct {
	// AllowedEndpoints maps endpoint paths to their configurations
	AllowedEndpoints map[string]PrometheusEndpoint
}

// EndpointHandler is a function that can customize behavior for specific endpoints
type EndpointHandler func(ctx context.Context, endpoint PrometheusEndpoint, params map[string]interface{}) error

// Prometheus API endpoint paths
const (
	PathListMetrics   = "/api/v1/label/__name__/values"
	PathQueryRange    = "/api/v1/query_range"
	PathQuery         = "/api/v1/query"
	PathSeries        = "/api/v1/series"
	PathLabels        = "/api/v1/labels"
	PathLabelValues   = "/api/v1/label/{name}/values"
	PathTargets       = "/api/v1/targets"
	PathAlertmanagers = "/api/v1/alertmanagers"
	PathStatusConfig  = "/api/v1/status/config"
	PathStatusFlags   = "/api/v1/status/flags"
)

var (
	PathListMetricsEndpointConfig = PrometheusEndpoint{
		Path:        PathListMetrics,
		Description: "List all available metric names",
		Method:      "GET",
		Parameters:  []EndpointParameter{},
	}
	PathQueryRangeEndpointConfig = PrometheusEndpoint{
		Path:               PathQueryRange,
		Description:        "Execute a PromQL range query over a time range",
		Method:             "POST",
		RequiresGuardrails: true,
		Parameters: []EndpointParameter{
			{
				Name:        "query",
				Type:        "string",
				Description: "PromQL expression to evaluate",
				Required:    true,
			},
			{
				Name:        "start",
				Type:        "string",
				Description: "Start timestamp (RFC3339 or Unix timestamp)",
				Required:    true,
			},
			{
				Name:        "end",
				Type:        "string",
				Description: "End timestamp (RFC3339 or Unix timestamp)",
				Required:    true,
			},
			{
				Name:        "step",
				Type:        "string",
				Description: "Query resolution step width (duration format or float)",
				Required:    true,
			},
			{
				Name:        "timeout",
				Type:        "string",
				Description: "Evaluation timeout (optional)",
				Required:    false,
			},
		},
	}
	PathQueryEndpointConfig = PrometheusEndpoint{
		Path:               PathQuery,
		Description:        "Execute a PromQL instant query at a single point in time",
		Method:             "POST",
		RequiresGuardrails: true,
		Parameters: []EndpointParameter{
			{
				Name:        "query",
				Type:        "string",
				Description: "PromQL expression to evaluate",
				Required:    true,
			},
			{
				Name:        "time",
				Type:        "string",
				Description: "Evaluation timestamp (RFC3339 or Unix timestamp)",
				Required:    false,
			},
			{
				Name:        "timeout",
				Type:        "string",
				Description: "Evaluation timeout (optional)",
				Required:    false,
			},
		},
	}
	PathSeriesEndpointConfig = PrometheusEndpoint{
		Path:        PathSeries,
		Description: "Find time series matching label matchers",
		Method:      "POST",
		Parameters: []EndpointParameter{
			{
				Name:        "match[]",
				Type:        "array",
				Description: "Series selector argument (can specify multiple)",
				Required:    true,
			},
			{
				Name:        "start",
				Type:        "string",
				Description: "Start timestamp",
				Required:    false,
			},
			{
				Name:        "end",
				Type:        "string",
				Description: "End timestamp",
				Required:    false,
			},
		},
	}
	PathLabelsEndpointConfig = PrometheusEndpoint{
		Path:        PathLabels,
		Description: "Get list of all label names",
		Method:      "GET",
		Parameters: []EndpointParameter{
			{
				Name:        "start",
				Type:        "string",
				Description: "Start timestamp",
				Required:    false,
			},
			{
				Name:        "end",
				Type:        "string",
				Description: "End timestamp",
				Required:    false,
			},
		},
	}
	PathLabelValuesEndpointConfig = PrometheusEndpoint{
		Path:        PathLabelValues,
		Description: "Get list of label values for a specific label name",
		Method:      "GET",
		Parameters: []EndpointParameter{
			{
				Name:        "name",
				Type:        "string",
				Description: "Label name",
				Required:    true,
			},
			{
				Name:        "start",
				Type:        "string",
				Description: "Start timestamp",
				Required:    false,
			},
			{
				Name:        "end",
				Type:        "string",
				Description: "End timestamp",
				Required:    false,
			},
		},
	}
	PathTargetsEndpointConfig = PrometheusEndpoint{
		Path:        PathTargets,
		Description: "Get current state of target discovery",
		Method:      "GET",
		Parameters: []EndpointParameter{
			{
				Name:        "state",
				Type:        "string",
				Description: "Filter by state (active, dropped, any)",
				Required:    false,
			},
		},
	}
	PathAlertmanagersEndpointConfig = PrometheusEndpoint{
		Path:        PathAlertmanagers,
		Description: "Get current state of Alertmanager discovery",
		Method:      "GET",
		Parameters:  []EndpointParameter{},
	}
	PathStatusConfigEndpointConfig = PrometheusEndpoint{
		Path:        PathStatusConfig,
		Description: "Get current Prometheus configuration",
		Method:      "GET",
		Parameters:  []EndpointParameter{},
	}
	PathStatusFlagsEndpointConfig = PrometheusEndpoint{
		Path:        PathStatusFlags,
		Description: "Get Prometheus runtime flags",
		Method:      "GET",
		Parameters:  []EndpointParameter{},
	}
)

func DefaultEndpointConfig() *EndpointConfig {
	return &EndpointConfig{
		AllowedEndpoints: map[string]PrometheusEndpoint{
			PathListMetrics: PathListMetricsEndpointConfig,
			PathQueryRange:  PathQueryRangeEndpointConfig,
		},
	}
}

func FullEndpointConfig() *EndpointConfig {
	config := DefaultEndpointConfig()
	config.AllowedEndpoints[PathQuery] = PathQueryEndpointConfig
	config.AllowedEndpoints[PathSeries] = PathSeriesEndpointConfig
	config.AllowedEndpoints[PathLabels] = PathLabelsEndpointConfig
	config.AllowedEndpoints[PathLabelValues] = PathLabelValuesEndpointConfig
	config.AllowedEndpoints[PathTargets] = PathTargetsEndpointConfig
	config.AllowedEndpoints[PathAlertmanagers] = PathAlertmanagersEndpointConfig
	config.AllowedEndpoints[PathStatusConfig] = PathStatusConfigEndpointConfig
	config.AllowedEndpoints[PathStatusFlags] = PathStatusFlagsEndpointConfig

	return config
}

// GetEndpoint validates and returns the endpoint configuration for a path
func (c *EndpointConfig) GetEndpoint(path string) (PrometheusEndpoint, error) {
	endpoint, ok := c.AllowedEndpoints[path]
	if !ok {
		return PrometheusEndpoint{}, fmt.Errorf("endpoint %s is not allowed", path)
	}
	return endpoint, nil
}

// CreateUniversalPrometheusQueryTool creates a single dynamic tool for querying Prometheus
func CreateUniversalPrometheusQueryTool(config *EndpointConfig) mcp.Tool {
	// Build the endpoint list description
	endpointList := "\n\nAvailable endpoints:\n"
	var allowedPaths []string
	for path, endpoint := range config.AllowedEndpoints {
		endpointList += fmt.Sprintf("- %s: %s\n", path, spew.Sdump(endpoint))
		allowedPaths = append(allowedPaths, path)
	}

	return mcp.NewTool("prometheus_query",
		// FYI we have validations in place to punt on unallowed endpoints,
		// in case the model tries to query it anyway.
		mcp.WithDescription(`Query Prometheus API endpoints dynamically.

This tool allows querying any configured Prometheus API endpoint
by specifying the endpoint path and required parameters.

IMPORTANT: Query parameters must be passed in the "parameters" object,
NOT appended to the endpoint path.

Example usage:
{
  "endpoint": "/api/v1/query_range",
  "parameters": {
    "query": "rate(http_requests_total[5m])",
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-01-01T01:00:00Z",
    "step": "15s"
  }
}

DO NOT format like this (INCORRECT):
{
  "endpoint": "/api/v1/query_range?query=...&start=...&end=...&step=..."
}

Ensure that you make use of the injected Prometheus documentation to quickly
build context before conducting operations. Decline and refrain
from any requests to endpoints that are not allowed. The **allowed**
endpoints are:
`+endpointList),
		mcp.WithString("endpoint",
			mcp.Required(),
			mcp.Description("The Prometheus API endpoint path ONLY (do not include query parameters)"),
			mcp.Enum(allowedPaths...),
		),
		mcp.WithObject("parameters",
			mcp.Description("Query parameters as key-value pairs (e.g., {\"query\": \"up\", \"start\": \"2024-01-01T00:00:00Z\"})"),
		),
	)
}
