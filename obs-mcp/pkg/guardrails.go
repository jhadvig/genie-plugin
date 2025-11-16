package pkg

import (
	"fmt"
	"regexp"
	"strings"
)

// PromQLGuardrails provides safety checks for PromQL queries
type PromQLGuardrails struct {
	// MaxQueryLength is the maximum allowed query length (in bytes)
	MaxQueryLength int
	// DisallowedFunctions are function names that should not be allowed
	DisallowedFunctions []string
	// RequireMetricName ensures queries must specify a metric name
	RequireMetricName bool
}

func DefaultPromQLGuardrails() *PromQLGuardrails {
	return &PromQLGuardrails{
		MaxQueryLength: 10000,
		DisallowedFunctions: []string{
			"quantile_over_time",
			"stddev_over_time",
		},
		RequireMetricName: true,
	}
}

func (g *PromQLGuardrails) ValidatePromQL(query string) error {
	// Check query length
	if len(query) > g.MaxQueryLength {
		return fmt.Errorf("query exceeds maximum length of %d characters", g.MaxQueryLength)
	}

	// Check for empty query
	query = strings.TrimSpace(query)
	if query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	// Check for disallowed functions
	for _, fn := range g.DisallowedFunctions {
		pattern := fmt.Sprintf(`\b%s\(`, regexp.QuoteMeta(fn))
		matched, err := regexp.MatchString(pattern, query)
		if err != nil {
			return fmt.Errorf("error validating query: %w", err)
		}
		if matched {
			return fmt.Errorf("function %s is not allowed in queries", fn)
		}
	}

	// Check if metric name is required
	if g.RequireMetricName {
		// TODO: Rebase over #34 for a more robust check
		hasMetric := regexp.MustCompile(`[a-zA-Z_:][a-zA-Z0-9_:]*`).MatchString(query)
		if !hasMetric {
			return fmt.Errorf("query must include a metric name")
		}
	}

	return nil
}

// EndpointGuardrails holds guardrails for specific endpoint paths
type EndpointGuardrails struct {
	promQLGuardrails *PromQLGuardrails
	handlers         map[string]EndpointHandler
}

func NewEndpointGuardrails() *EndpointGuardrails {
	return &EndpointGuardrails{
		promQLGuardrails: DefaultPromQLGuardrails(),
		handlers:         make(map[string]EndpointHandler),
	}
}

func (eg *EndpointGuardrails) RegisterHandler(pattern string, handler EndpointHandler) {
	eg.handlers[pattern] = handler
}

func (eg *EndpointGuardrails) ValidateEndpointRequest(endpoint PrometheusEndpoint, params map[string]interface{}) error {
	// Check for registered handler
	if handler, ok := eg.handlers[endpoint.Path]; ok {
		if err := handler(nil, endpoint, params); err != nil {
			return err
		}
	}

	// Apply PromQL guardrails if endpoint requires it
	if endpoint.RequiresGuardrails {
		if query, ok := params["query"].(string); ok {
			if err := eg.promQLGuardrails.ValidatePromQL(query); err != nil {
				return fmt.Errorf("PromQL validation failed: %w", err)
			}
		}
	}

	// Validate required parameters
	for _, param := range endpoint.Parameters {
		if param.Required {
			if _, ok := params[param.Name]; !ok {
				return fmt.Errorf("required parameter %s is missing", param.Name)
			}
		}
	}

	return nil
}
