package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// CreateFindWidgetsTool creates the find_widgets tool definition
func CreateFindWidgetsTool() mcp.Tool {
	return mcp.NewTool("find_widgets",
		mcp.WithDescription("Find widgets in the active dashboard based on natural language descriptions. TIP: Use get_active_dashboard first to see the current layout state, then use this tool to locate specific widgets by description."),
		mcp.WithString("description",
			mcp.Required(),
			mcp.Description("Natural language description of the widget to find (e.g., 'chart widget', 'sales graph', 'table in top right')"),
		),
		mcp.WithString("component_type",
			mcp.Description("Optional: specific component type to filter by (chart, table, metric, text, image, iframe)"),
		),
		mcp.WithString("breakpoint",
			mcp.Description("Breakpoint to search in (default: lg)"),
		),
	)
}

// CreateManipulateWidgetTool creates the manipulate_widget tool definition
func CreateManipulateWidgetTool() mcp.Tool {
	return mcp.NewTool("manipulate_widget",
		mcp.WithDescription("Perform widget operations on the active dashboard like move, resize, or remove. WORKFLOW: 1) Call get_active_dashboard to see all widgets, 2) Identify the exact widget ID, 3) Call this tool with the widget ID."),
		mcp.WithString("widget_id",
			mcp.Required(),
			mcp.Description("Exact widget ID from get_active_dashboard (e.g., 'widget-1', 'chart-abc123'). Do NOT use descriptions - use the actual ID."),
		),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("Operation to perform: 'move', 'resize', or 'remove'"),
		),
		mcp.WithString("x",
			mcp.Description("New X position as string (required for move operation, e.g., '2')"),
		),
		mcp.WithString("y",
			mcp.Description("New Y position as string (required for move operation, e.g., '1')"),
		),
		mcp.WithString("w",
			mcp.Description("New width as string (required for resize operation, e.g., '4')"),
		),
		mcp.WithString("h",
			mcp.Description("New height as string (required for resize operation, e.g., '3')"),
		),
		mcp.WithString("breakpoint",
			mcp.Description("Breakpoint to operate on (default: lg)"),
		),
		mcp.WithBoolean("apply_to_all_breakpoints",
			mcp.Description("Whether to apply changes proportionally to all breakpoints (default: false)"),
		),
	)
}

// CreateAddWidgetTool creates the add_widget tool definition
func CreateAddWidgetTool() mcp.Tool {
	return mcp.NewTool("add_widget",
		mcp.WithDescription("Add new widgets to the active dashboard. RECOMMENDED: Call get_active_dashboard first to understand the current layout and find the best position for the new widget to avoid overlaps."),
		mcp.WithString("widget_description",
			mcp.Required(),
			mcp.Description("Description of widget to add (e.g., 'chart showing sales data', 'table with customer info', 'metric displaying revenue')"),
		),
		mcp.WithString("component_type",
			mcp.Required(),
			mcp.Description("Type of component to create (chart, table, metric, text, image, iframe)"),
		),
		mcp.WithString("position_hint",
			mcp.Description("Optional position hint (e.g., 'top right', 'bottom left', 'next to the sales chart')"),
		),
		mcp.WithString("size_hint",
			mcp.Description("Optional size hint (e.g., 'large', 'small', 'medium', '4x3')"),
		),
		mcp.WithString("props",
			mcp.Description("Optional JSON string containing widget-specific properties and configuration (e.g., chart settings, data sources, styling options)"),
		),
		mcp.WithString("breakpoint",
			mcp.Description("Breakpoint to add widget to (default: lg)"),
		),
	)
}

// CreateBatchWidgetOperationsTool creates the batch_widget_operations tool definition
func CreateBatchWidgetOperationsTool() mcp.Tool {
	return mcp.NewTool("batch_widget_operations",
		mcp.WithDescription("Perform multiple widget operations on the active dashboard in a single command. ESSENTIAL: Call get_active_dashboard first to understand the current state before planning batch operations."),
		// Note: For array parameters, we'll handle them as JSON strings for now
		mcp.WithString("commands_json",
			mcp.Required(),
			mcp.Description("JSON array of natural language commands to execute in sequence"),
		),
		mcp.WithBoolean("atomic",
			mcp.Description("Whether all operations must succeed or all fail (default: true)"),
		),
		mcp.WithString("breakpoint",
			mcp.Description("Breakpoint to operate on (default: lg)"),
		),
	)
}

// CreateDashboardTool creates the create_dashboard tool definition
func CreateDashboardTool() mcp.Tool {
	return mcp.NewTool("create_dashboard",
		mcp.WithDescription("Create a new empty dashboard and set it as active"),
		mcp.WithString("name",
			mcp.Description("Optional display name for the dashboard (e.g., 'Sales Dashboard', 'System Overview')"),
		),
		mcp.WithString("description",
			mcp.Description("Optional description of the dashboard purpose"),
		),
	)
}

// CreateAnalyzeLayoutTool creates the analyze_layout tool definition
func CreateAnalyzeLayoutTool() mcp.Tool {
	return mcp.NewTool("analyze_layout",
		mcp.WithDescription("Analyze the active dashboard structure and provide insights. NOTE: For basic layout information, get_active_dashboard is often more direct and provides raw data instead of analysis."),
		mcp.WithString("question",
			mcp.Required(),
			mcp.Description("Question about the layout (e.g., 'what widgets are here?', 'how many charts?', 'what's in the top row?')"),
		),
		mcp.WithString("breakpoint",
			mcp.Description("Breakpoint to analyze (default: lg)"),
		),
	)
}

// CreateConfigureWidgetTool creates the configure_widget tool definition
func CreateConfigureWidgetTool() mcp.Tool {
	return mcp.NewTool("configure_widget",
		mcp.WithDescription("Configure widget properties and settings in the active dashboard. BEST PRACTICE: Use get_active_dashboard first to see current widget properties, then make informed configuration changes."),
		mcp.WithString("widget_selector",
			mcp.Required(),
			mcp.Description("Description of which widget to configure (e.g., 'sales table', 'main chart', 'system status widget')"),
		),
		mcp.WithString("configuration_request",
			mcp.Required(),
			mcp.Description("Natural language description of the configuration change (e.g., 'change filters to show only critical systems', 'set chart type to bar', 'update data source')"),
		),
		mcp.WithString("domain_context",
			mcp.Description("Optional domain context to help interpret the request (e.g., 'monitoring dashboard', 'sales analytics', 'system health')"),
		),
		mcp.WithString("current_props",
			mcp.Description("Optional JSON string of current widget props to help understand existing configuration"),
		),
		mcp.WithString("suggested_props",
			mcp.Description("Optional JSON string of suggested prop changes from frontend if available"),
		),
		mcp.WithString("breakpoint",
			mcp.Description("Breakpoint to modify (default: lg)"),
		),
	)
}

// CreateGetActiveDashboardTool creates the get_active_dashboard tool definition
func CreateGetActiveDashboardTool() mcp.Tool {
	return mcp.NewTool("get_active_dashboard",
		mcp.WithDescription("Get the current active dashboard schema and layout information. This provides the LLM with the current state of widgets, their positions, and properties to make informed decisions about layout modifications."),
		mcp.WithString("breakpoint",
			mcp.Description("Breakpoint to return layout for (default: lg)"),
		),
		mcp.WithBoolean("include_all_breakpoints",
			mcp.Description("Whether to include layouts for all breakpoints (default: false)"),
		),
	)
}