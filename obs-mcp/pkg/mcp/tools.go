package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func CreateListMetricsTool() mcp.Tool {
	tool := mcp.NewTool("list_metrics",
		mcp.WithDescription("List all available metrics in Prometheus"),
	)
	// workaround for tool with no parameter
	// see https://github.com/containers/kubernetes-mcp-server/pull/341/files#diff-8f8a99cac7a7cbb9c14477d40539efa1494b62835603244ba9f10e6be1c7e44c
	tool.InputSchema = mcp.ToolInputSchema{}
	tool.RawInputSchema = []byte(`{"type":"object","properties":{}}`)
	return tool
}

func CreateExecuteRangeQueryTool() mcp.Tool {
	return mcp.NewTool("execute_range_query",
		mcp.WithDescription(`Execute a PromQL range query with flexible time specification.

For current time data queries, use only the 'duration' parameter to specify how far back
to look from now (e.g., '1h' for last hour, '30m' for last 30 minutes). In that case
YOU MUST NOT provide neither 'start' NOR 'end' at all.

For historical data queries, use explicit 'start' and 'end' times.
`),
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
