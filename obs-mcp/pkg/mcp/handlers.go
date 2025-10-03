package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/inecas/obs-mcp/pkg/prometheus"
	"github.com/mark3labs/mcp-go/mcp"
)

// getTokenFromCtx gets the authorization header from the
// provided context
func getTokenFromCtx(ctx context.Context) string {
	k8sToken := ctx.Value(authHeaderStr)
	if k8sToken == nil {
		return ""
	}
	k8TokenStr, ok := k8sToken.(string)
	if !ok {
		slog.Warn("Couldn't parse token... ignoring.")
		return ""
	}
	return k8TokenStr
}

func getPromClient(ctx context.Context, promURL string) (*prometheus.PrometheusClient, error) {
	token := getTokenFromCtx(ctx)

	promClient, err := prometheus.NewPrometheusClient(promURL, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %v", err)
	}

	return promClient, nil
}

func ListMetricsHandler(promURL string) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		promClient, err := getPromClient(ctx, promURL)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create Prometheus client: %s", err.Error())), nil
		}

		metrics, err := promClient.ListMetrics(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to list metrics: %s", err.Error())), nil
		}

		result, err := json.Marshal(metrics)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal metrics: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(string(result)), nil
	}
}

func ExecuteRangeQueryHandler(promURL string) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		promClient, err := getPromClient(ctx, promURL)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create Prometheus client: %s", err.Error())), nil
		}

		// Get required query parameter
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError("query parameter is required and must be a string"), nil
		}

		// Get required step parameter
		step, err := req.RequireString("step")
		if err != nil {
			return mcp.NewToolResultError("step parameter is required and must be a string"), nil
		}

		// Parse step duration
		stepDuration, err := time.ParseDuration(step)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid step format: %s", err.Error())), nil
		}

		// Get optional parameters
		startStr := req.GetString("start", "")
		endStr := req.GetString("end", "")
		durationStr := req.GetString("duration", "")

		// Validate parameter combinations
		if startStr != "" && endStr != "" && durationStr != "" {
			return mcp.NewToolResultError("cannot specify both start/end and duration parameters"), nil
		}

		if (startStr != "" && endStr == "") || (startStr == "" && endStr != "") {
			return mcp.NewToolResultError("both start and end must be provided together"), nil
		}

		var startTime, endTime time.Time

		// Handle duration-based query (default to 1h if nothing specified)
		if durationStr != "" || (startStr == "" && endStr == "") {
			if durationStr == "" {
				durationStr = "1h"
			}

			duration, err := prometheus.ParseDuration(durationStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid duration format: %s", err.Error())), nil
			}

			endTime = time.Now()
			startTime = endTime.Add(-duration)
		} else {
			// Handle explicit start/end times
			startTime, err = prometheus.ParseTimestamp(startStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid start time format: %s", err.Error())), nil
			}

			endTime, err = prometheus.ParseTimestamp(endStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid end time format: %s", err.Error())), nil
			}
		}

		// Execute the range query
		result, err := promClient.ExecuteRangeQuery(ctx, query, startTime, endTime, stepDuration)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to execute range query: %s", err.Error())), nil
		}

		// Convert to JSON
		jsonResult, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	}
}
