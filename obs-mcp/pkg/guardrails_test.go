package pkg

import (
	"strings"
	"testing"
)

func TestValidatePromQL(t *testing.T) {
	guardrails := DefaultPromQLGuardrails()

	tests := []struct {
		name        string
		query       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid simple query",
			query:       "up",
			expectError: false,
		},
		{
			name:        "Valid complex query",
			query:       "rate(http_requests_total[5m])",
			expectError: false,
		},
		{
			name:        "Empty query",
			query:       "",
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "Whitespace only query",
			query:       "   \n\t   ",
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "Query exceeds max length",
			query:       strings.Repeat("a", 10001),
			expectError: true,
			errorMsg:    "exceeds maximum length",
		},
		{
			name:        "Query at max length",
			query:       strings.Repeat("a", 10000),
			expectError: false,
		},
		{
			name:        "Query with aggregation",
			query:       "sum(rate(container_cpu_usage_seconds_total[5m])) by (namespace)",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := guardrails.ValidatePromQL(tt.query)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidatePromQLWithCustomMaxLength(t *testing.T) {
	guardrails := &PromQLGuardrails{
		MaxQueryLength:      100,
		DisallowedFunctions: []string{},
		RequireMetricName:   false,
	}

	tests := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name:        "Query within custom limit",
			query:       strings.Repeat("a", 100),
			expectError: false,
		},
		{
			name:        "Query exceeds custom limit",
			query:       strings.Repeat("a", 101),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := guardrails.ValidatePromQL(tt.query)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateEndpointRequest(t *testing.T) {
	guardrails := NewEndpointGuardrails()

	tests := []struct {
		name        string
		endpoint    PrometheusEndpoint
		params      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid query_range request",
			endpoint: PrometheusEndpoint{
				Path:               PathQueryRange,
				RequiresGuardrails: true,
			},
			params: map[string]interface{}{
				"query": "up",
				"start": "2024-01-01T00:00:00Z",
				"end":   "2024-01-01T01:00:00Z",
				"step":  "15s",
			},
			expectError: false,
		},
		{
			name: "Invalid query_range request (no end)",
			endpoint: PrometheusEndpoint{
				Path:               PathQueryRange,
				RequiresGuardrails: true,
			},
			params: map[string]interface{}{
				"query": "up",
				"start": "2024-01-01T00:00:00Z",
				"step":  "15s",
			},
			expectError: true,
		},
		{
			name: "Query_range with empty query",
			endpoint: PrometheusEndpoint{
				Path:               PathQueryRange,
				RequiresGuardrails: true,
			},
			params: map[string]interface{}{
				"query": "",
			},
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name: "Query_range with query too long",
			endpoint: PrometheusEndpoint{
				Path:               PathQueryRange,
				RequiresGuardrails: true,
			},
			params: map[string]interface{}{
				"query": strings.Repeat("a", 10001),
			},
			expectError: true,
			errorMsg:    "exceeds maximum length",
		},
		{
			name: "Endpoint without guardrails",
			endpoint: PrometheusEndpoint{
				Path:               PathListMetrics,
				RequiresGuardrails: false,
			},
			params:      map[string]interface{}{},
			expectError: false,
		},
		{
			name: "Query endpoint with valid query",
			endpoint: PrometheusEndpoint{
				Path:               PathQuery,
				RequiresGuardrails: true,
			},
			params: map[string]interface{}{
				"query": "rate(http_requests_total[5m])",
			},
			expectError: false,
		},
	}

	pathToPathConfigurationMap := map[string]PrometheusEndpoint{
		PathQuery:         PathQueryEndpointConfig,
		PathQueryRange:    PathQueryRangeEndpointConfig,
		PathListMetrics:   PathListMetricsEndpointConfig,
		PathLabelValues:   PathLabelValuesEndpointConfig,
		PathSeries:        PathSeriesEndpointConfig,
		PathLabels:        PathLabelsEndpointConfig,
		PathAlertmanagers: PathAlertmanagersEndpointConfig,
		PathTargets:       PathTargetsEndpointConfig,
		PathStatusConfig:  PathStatusConfigEndpointConfig,
		PathStatusFlags:   PathStatusFlagsEndpointConfig,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpointConfig := pathToPathConfigurationMap[tt.endpoint.Path]
			endpointConfig.RequiresGuardrails = tt.endpoint.RequiresGuardrails
			err := guardrails.ValidateEndpointRequest(endpointConfig, tt.params)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
