package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func CreateListMetricsTool() mcp.Tool {
	return mcp.NewTool("list_metrics",
		mcp.WithDescription("List all available metrics in Prometheus"),
	)
}

func CreateExecuteRangeQueryTool() mcp.Tool {
	return mcp.NewTool("execute_range_query",
		mcp.WithDescription("Execute a PromQL range query with flexible time specification.

For current/recent data queries, use the 'duration' parameter to specify how far back
to look from now (e.g., '1h' for last hour, '30m' for last 30 minutes).

For historical data queries, use explicit 'start' and 'end' times.
"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("PromQL query string"),
		),
		mcp.WithString("step",
			mcp.Required(),
			mcp.Description("Query resolution step width (e.g., '15s', '1m', '1h')"),
		),
		mcp.WithString("start",
			mcp.Description("Start time as RFC3339 or Unix timestamp (optional)"),
		),
		mcp.WithString("end",
			mcp.Description("End time as RFC3339 or Unix timestamp (optional)"),
		),
		mcp.WithString("duration",
			mcp.Description("Duration to look back from now (e.g., '1h', '30m', '1d', '2w') (optional)"),
		),
	)
}
