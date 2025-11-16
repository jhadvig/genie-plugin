package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// DynamicPrometheusHandler creates a handler that can query any allowed Prometheus endpoint
func DynamicPrometheusHandler(opts ObsMCPOptions, guardrails *EndpointGuardrails) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		endpointPath, err := req.RequireString("endpoint")
		if err != nil {
			return mcp.NewToolResultError("endpoint parameter is required and must be a string"), nil
		}
		endpoint, err := opts.EndpointConfig.GetEndpoint(endpointPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("endpoint not allowed: %s", err.Error())), nil
		}
		params := make(map[string]interface{})
		if req.Params.Arguments != nil {
			if argsMap, ok := req.Params.Arguments.(map[string]interface{}); ok {
				if paramsRaw, exists := argsMap["parameters"]; exists {
					if paramsMap, ok := paramsRaw.(map[string]interface{}); ok {
						params = paramsMap
					} else {
						return mcp.NewToolResultError("parameters must be an object"), nil
					}
				}
			}
		}
		if err := guardrails.ValidateEndpointRequest(endpoint, params); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("validation failed: %s", err.Error())), nil
		}
		result, err := executePrometheusEndpoint(ctx, opts, endpoint, params)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to execute query: %s", err.Error())), nil
		}
		jsonResult, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	}
}

func executePrometheusEndpoint(ctx context.Context, opts ObsMCPOptions, endpoint PrometheusEndpoint, params map[string]interface{}) (map[string]interface{}, error) {
	promURL := opts.PromURL
	if !strings.HasSuffix(promURL, "/") {
		promURL += "/"
	}

	// Handle path parameters (like {name} in /api/v1/label/{name}/values)
	endpointPath := endpoint.Path
	for key, value := range params {
		placeholder := fmt.Sprintf("{%s}", key)
		if strings.Contains(endpointPath, placeholder) {
			endpointPath = strings.ReplaceAll(endpointPath, placeholder, fmt.Sprintf("%v", value))
			delete(params, key) // Remove from params since it's now in the path
		}
	}
	fullURL := promURL + strings.TrimPrefix(endpointPath, "/")

	// Create properly configured API client with authentication and TLS
	apiConfig, err := createAPIConfig(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create API config: %w", err)
	}

	// Use the configured RoundTripper for authentication and TLS
	client := &http.Client{
		Timeout:   60 * time.Second,
		Transport: apiConfig.RoundTripper,
	}

	var httpReq *http.Request
	if endpoint.Method == "GET" || endpoint.Method == "" {
		if len(params) > 0 {
			urlParams := url.Values{}
			for key, value := range params {
				urlParams.Add(key, fmt.Sprintf("%v", value))
			}
			fullURL += "?" + urlParams.Encode()
		}

		httpReq, err = http.NewRequestWithContext(ctx, "GET", fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
	} else if endpoint.Method == "POST" {
		formData := url.Values{}
		for key, value := range params {
			formData.Add(key, fmt.Sprintf("%v", value))
		}
		httpReq, err = http.NewRequestWithContext(ctx, "POST", fullURL, strings.NewReader(formData.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// Authentication is handled by the RoundTripper from createAPIConfig
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if status, ok := result["status"].(string); ok && status != "success" {
		if errorMsg, ok := result["error"].(string); ok {
			return nil, fmt.Errorf("prometheus error: %s", errorMsg)
		}
		return nil, fmt.Errorf("prometheus returned non-success status: %s", status)
	}

	return result, nil
}
