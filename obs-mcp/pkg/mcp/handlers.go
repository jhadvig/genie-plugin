package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/inecas/obs-mcp/pkg/prometheus"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/prometheus/common/model"
)

// errorResult is a helper to log and return an error result.
func errorResult(msg string) (*mcp.CallToolResult, error) {
	slog.Info("Query execution error: " + msg)
	return mcp.NewToolResultError(msg), nil
}

// ListMetricsHandler handles the listing of available Prometheus metrics.
func ListMetricsHandler(opts ObsMCPOptions) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.Info("ListMetricsHandler called")
		slog.Debug("ListMetricsHandler params", "params", req.Params)
		promClient, err := getPromClient(ctx, opts)
		if err != nil {
			return errorResult(fmt.Sprintf("failed to create Prometheus client: %s", err.Error()))
		}

		metrics, err := promClient.ListMetrics(ctx)
		if err != nil {
			return errorResult(fmt.Sprintf("failed to list metrics: %s", err.Error()))
		}

		slog.Info("ListMetricsHandler executed successfully", "resultLenght", len(metrics))
		slog.Debug("ListMetricsHandler results", "results", metrics)

		result, err := json.Marshal(metrics)
		if err != nil {
			return errorResult(fmt.Sprintf("failed to marshal metrics: %s", err.Error()))
		}

		return mcp.NewToolResultText(string(result)), nil
	}
}

// ExecuteRangeQueryHandler handles the execution of Prometheus range queries.
func ExecuteRangeQueryHandler(opts ObsMCPOptions) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.Info("ExecuteRangeQueryHandler called")
		slog.Debug("ExecuteRangeQueryHandler params", "params", req.Params)

		promClient, err := getPromClient(ctx, opts)
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
			return errorResult("cannot specify both start/end and duration parameters")
		}

		if (startStr != "" && endStr == "") || (startStr == "" && endStr != "") {
			return errorResult("both start and end must be provided together")
		}

		var startTime, endTime time.Time

		// Handle duration-based query (default to 1h if nothing specified)
		if durationStr != "" || (startStr == "" && endStr == "") {
			if durationStr == "" {
				durationStr = "1h"
			}

			duration, err := prometheus.ParseDuration(durationStr)
			if err != nil {
				return errorResult(fmt.Sprintf("invalid duration format: %s", err.Error()))
			}

			endTime = time.Now()
			startTime = endTime.Add(-duration)
		} else {
			// Handle explicit start/end times
			startTime, err = prometheus.ParseTimestamp(startStr)
			if err != nil {
				return errorResult(fmt.Sprintf("invalid start time format: %s", err.Error()))
			}

			endTime, err = prometheus.ParseTimestamp(endStr)
			if err != nil {
				return errorResult(fmt.Sprintf("invalid end time format: %s", err.Error()))
			}
		}

		// Execute the range query
		result, err := promClient.ExecuteRangeQuery(ctx, query, startTime, endTime, stepDuration, opts.UseGuardrails)
		if err != nil {
			return errorResult(fmt.Sprintf("failed to execute range query: %s", err.Error()))
		}

		resMatrix, ok := result["result"].(model.Matrix)
		if ok {
			slog.Info("ExecuteRangeQueryHandler executed successfully", "resultLenght", resMatrix.Len())
			slog.Debug("ExecuteRangeQueryHandler results", "results", resMatrix)
		} else {
			slog.Info("ExecuteRangeQueryHandler executed successfully (unknown format)", "result", result)
		}

		// Convert to JSON
		jsonResult, err := json.Marshal(result)
		if err != nil {
			return errorResult(fmt.Sprintf("failed to marshal result: %s", err.Error()))
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	}
}
