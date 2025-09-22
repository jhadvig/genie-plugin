# MCP Tools Design for Layout Manager

## Overview

The MCP server provides a precise widget manipulation interface that operates directly on the active dashboard through database operations. The system uses a two-step workflow: first get the current dashboard state with exact widget IDs, then perform precise operations using those IDs.

## MCP Tools Architecture

### Core Workflow (Precise Widget Manipulation)
1. **Get Dashboard State**: Use `get_active_dashboard` to see current layout with exact widget IDs
2. **Identify Targets**: LLM identifies exact widget IDs from the dashboard data
3. **Precise Operations**: Use exact widget IDs for move, resize, remove operations
4. **Database Persistence**: All changes are directly saved to PostgreSQL
5. **Change Tracking**: Return complete widget state and change details

### Dashboard Creation Workflow
1. **Empty Dashboard Creation**: Dashboards are created empty by default unless explicitly requested otherwise
2. **Separate Widget Addition**: When users request widgets during dashboard creation, use the `add_widget` tool instead of embedding widgets in the create request
3. **Workflow Example**:
   - User: "Create a sales dashboard with a revenue chart"
   - System: Creates empty dashboard named "sales dashboard"
   - System: Calls `add_widget` tool to add revenue chart
   - Result: Dashboard with single widget, proper change tracking

## MCP Tool Definitions

### 1. `get_active_dashboard` - Dashboard State Tool

```json
{
  "name": "get_active_dashboard",
  "description": "Get the current active dashboard schema and layout information with exact widget IDs and positions. Essential first step for precise widget manipulation.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "breakpoint": {
        "type": "string",
        "default": "lg",
        "description": "Breakpoint to return layout for"
      },
      "include_all_breakpoints": {
        "type": "boolean",
        "default": false,
        "description": "Whether to include layouts for all breakpoints"
      }
    }
  }
}
```

**Response Format:**
```json
{
  "success": true,
  "operation": "get_active_dashboard",
  "activeLayoutId": "dashboard-20250922-143052",
  "analysis": {
    "layoutId": "dashboard-20250922-143052",
    "name": "Sales Dashboard",
    "description": "Main sales overview",
    "breakpoint": "lg",
    "cols": 12,
    "widgets": [
      {
        "i": "widget-1727012345678",
        "x": 0, "y": 0, "w": 6, "h": 4,
        "componentType": "chart",
        "props": {"title": "Revenue Chart", "chartType": "line"}
      },
      {
        "i": "widget-1727012345679",
        "x": 6, "y": 0, "w": 6, "h": 4,
        "componentType": "table",
        "props": {"title": "Customer Data"}
      }
    ]
  }
}
```

### 2. `find_widgets` - Widget Discovery Tool

```json
{
  "name": "find_widgets",
  "description": "Find widgets in the active dashboard based on natural language descriptions like 'chart widget', 'sales graph', or 'table with customer data'",
  "inputSchema": {
    "type": "object",
    "properties": {
      "description": {
        "type": "string",
        "description": "Natural language description of the widget to find (e.g., 'chart widget', 'sales graph', 'table in top right')"
      },
      "component_type": {
        "type": "string",
        "enum": ["chart", "table", "metric", "text", "image", "iframe"],
        "description": "Optional: specific component type to filter by"
      },
      "breakpoint": {
        "type": "string",
        "default": "lg",
        "description": "Breakpoint to search in"
      }
    },
    "required": ["description"]
  }
}
```

**Response Format:**
```json
{
  "widgets": [
    {
      "id": "widget-1",
      "componentType": "chart",
      "position": {"x": 0, "y": 0, "w": 4, "h": 3},
      "props": {"title": "Sales Chart", "chartType": "line"},
      "matchReason": "Found chart widget with title containing 'sales'"
    }
  ],
  "searchQuery": "chart widget",
  "totalFound": 1
}
```

### 3. `manipulate_widget` - Precise Widget Operations Tool

```json
{
  "name": "manipulate_widget",
  "description": "Perform precise widget operations using exact widget IDs. WORKFLOW: 1) Call get_active_dashboard to see all widgets, 2) Identify the exact widget ID, 3) Call this tool with the widget ID.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "widget_id": {
        "type": "string",
        "description": "Exact widget ID from get_active_dashboard (e.g., 'widget-1727012345678'). Do NOT use descriptions - use the actual ID."
      },
      "operation": {
        "type": "string",
        "enum": ["move", "resize", "remove"],
        "description": "Operation to perform: 'move', 'resize', or 'remove'"
      },
      "x": {
        "type": "string",
        "description": "New X position as string (required for move operation, e.g., '2')"
      },
      "y": {
        "type": "string",
        "description": "New Y position as string (required for move operation, e.g., '1')"
      },
      "w": {
        "type": "string",
        "description": "New width as string (required for resize operation, e.g., '4')"
      },
      "h": {
        "type": "string",
        "description": "New height as string (required for resize operation, e.g., '3')"
      },
      "breakpoint": {
        "type": "string",
        "default": "lg",
        "description": "Breakpoint to operate on"
      },
      "apply_to_all_breakpoints": {
        "type": "boolean",
        "default": false,
        "description": "Whether to apply changes proportionally to all breakpoints"
      }
    },
    "required": ["widget_id", "operation"]
  }
}
```

**Response Format:**
```json
{
  "success": true,
  "operation": "manipulate_widget",
  "activeLayoutId": "dashboard-20250922-143052",
  "targetedWidgets": ["widget-1727012345678"],
  "widgets": [
    {
      "id": "widget-1727012345678",
      "componentType": "chart",
      "position": {"x": 2, "y": 1, "w": 6, "h": 4},
      "props": {"title": "Revenue Chart", "chartType": "line"},
      "breakpoint": "lg"
    }
  ],
  "allChanges": [
    {
      "widgetId": "widget-1727012345678",
      "action": "moved",
      "breakpoint": "lg",
      "wasTargeted": true,
      "reason": "direct move operation",
      "previousState": {"x": 0, "y": 0, "w": 6, "h": 4},
      "newState": {"x": 2, "y": 1, "w": 6, "h": 4}
    }
  ],
  "message": "Moved widget 'widget-1727012345678' to position (2, 1) in active dashboard 'Sales Dashboard'",
  "timestamp": "2025-09-22T14:30:00Z"
}
```

### 4. `add_widget` - Widget Creation Tool

```json
{
  "name": "add_widget",
  "description": "Add new widgets to the active dashboard based on natural language requests like 'add a chart showing sales data' or 'create a metrics widget in the top right'",
  "inputSchema": {
    "type": "object",
    "properties": {
      "widget_description": {
        "type": "string",
        "description": "Description of widget to add (e.g., 'chart showing sales data', 'table with customer info', 'metric displaying revenue')"
      },
      "position_hint": {
        "type": "string",
        "description": "Optional position hint (e.g., 'top right', 'bottom left', 'next to the sales chart')"
      },
      "component_type": {
        "type": "string",
        "enum": ["chart", "table", "metric", "text", "image", "iframe"],
        "description": "Type of component to create"
      },
      "size_hint": {
        "type": "string",
        "description": "Optional size hint (e.g., 'large', 'small', 'medium', '4x3')"
      },
      "props": {
        "type": "string",
        "description": "Optional JSON string containing widget-specific properties and configuration (e.g., chart settings, data sources, styling options)"
      },
      "breakpoint": {
        "type": "string",
        "default": "lg",
        "description": "Breakpoint to add widget to"
      }
    },
    "required": ["widget_description", "component_type"]
  }
}
```

**Response Format:**
```json
{
  "success": true,
  "operation": "add_widget",
  "activeLayoutId": "dashboard-20250922-143052",
  "widgets": [
    {
      "id": "widget-1727012345680",
      "componentType": "metric",
      "position": {"x": 0, "y": 1, "w": 4, "h": 3},
      "props": {"title": "Total Sales", "value": "$125,000"},
      "breakpoint": "lg"
    }
  ],
  "message": "Added metric widget 'Total Sales' (ID: widget-1727012345680) to active dashboard 'Sales Dashboard' at position (0, 1)",
  "timestamp": "2025-09-22T14:30:00Z"
}
```

### 5. `batch_widget_operations` - Multi-Widget Operations Tool

```json
{
  "name": "batch_widget_operations",
  "description": "Perform multiple widget operations on the active dashboard in a single command like 'remove all chart widgets and add a table' or 'resize all widgets to be smaller'",
  "inputSchema": {
    "type": "object",
    "properties": {
      "commands": {
        "type": "array",
        "items": {
          "type": "string"
        },
        "description": "List of natural language commands to execute in sequence"
      },
      "atomic": {
        "type": "boolean",
        "default": true,
        "description": "Whether all operations must succeed or all fail"
      },
      "breakpoint": {
        "type": "string",
        "default": "lg",
        "description": "Breakpoint to operate on"
      }
    },
    "required": ["commands"]
  }
}
```

### 5. `configure_widget` - Widget Configuration Tool

```json
{
  "name": "configure_widget",
  "description": "Configure widget properties and settings in the active dashboard using natural language descriptions. Handles complex configuration changes like filters, data sources, display options, and component-specific settings.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "widget_selector": {
        "type": "string",
        "description": "Description of which widget to configure (e.g., 'sales table', 'main chart', 'system status widget')"
      },
      "configuration_request": {
        "type": "string",
        "description": "Natural language description of the configuration change (e.g., 'change filters to show only critical systems', 'set chart type to bar', 'update data source')"
      },
      "domain_context": {
        "type": "string",
        "description": "Optional domain context to help interpret the request (e.g., 'monitoring dashboard', 'sales analytics', 'system health')"
      },
      "current_props": {
        "type": "object",
        "description": "Optional: current widget props to help understand existing configuration",
        "additionalProperties": true
      },
      "suggested_props": {
        "type": "object",
        "description": "Optional: suggested prop changes from frontend if available",
        "additionalProperties": true
      },
      "breakpoint": {
        "type": "string",
        "default": "lg",
        "description": "Breakpoint to modify"
      }
    },
    "required": ["widget_selector", "configuration_request"]
  }
}
```

### 6. `analyze_layout` - Layout Analysis Tool

```json
{
  "name": "analyze_layout",
  "description": "Analyze the active dashboard structure and provide insights like 'what widgets are in this layout' or 'how is the layout organized'",
  "inputSchema": {
    "type": "object",
    "properties": {
      "question": {
        "type": "string",
        "description": "Question about the layout (e.g., 'what widgets are here?', 'how many charts?', 'what's in the top row?')"
      },
      "breakpoint": {
        "type": "string",
        "default": "lg",
        "description": "Breakpoint to analyze"
      }
    },
    "required": ["question"]
  }
}
```

## Intent Parsing Logic

### Command Classification

**Remove Operations:**
- Keywords: "remove", "delete", "get rid of", "take away"
- Examples: "remove the chart widget", "delete sales table"

**Resize Operations:**
- Keywords: "make larger", "make smaller", "resize", "bigger", "smaller", "expand", "shrink"
- Examples: "make the table larger", "resize chart to 6x4"

**Move Operations:**
- Keywords: "move", "relocate", "put", "place"
- Position indicators: "top left", "bottom right", "next to", "above", "below"
- Examples: "move chart to top right", "put table next to metrics"

**Add Operations:**
- Keywords: "add", "create", "insert", "place"
- Examples: "add a chart", "create metrics widget"

**Update Properties:**
- Keywords: "change", "update", "modify", "set"
- Examples: "change chart title to 'Sales'", "update table data source"

### Widget Identification Strategies

1. **By Component Type**: "chart", "table", "metric", "text"
2. **By Title/Props**: "sales chart", "customer table", "revenue metric"
3. **By Position**: "widget in top left", "chart on the right"
4. **By Size**: "large widget", "small chart"
5. **By Index**: "first widget", "second chart", "last table"

## Database Integration Pattern

```go
// Example handler structure for precise widget manipulation
func ManipulateWidgetHandler(layoutRepo *db.LayoutRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // 1. Get exact widget ID from request (no natural language parsing)
        widgetID, err := request.RequireString("widget_id")
        if err != nil {
            return mcp.NewToolResultError("widget_id parameter is required"), nil
        }

        operation, err := request.RequireString("operation")
        if err != nil {
            return mcp.NewToolResultError("operation parameter is required"), nil
        }

        // 2. Get active layout directly from database
        layout, err := layoutRepo.GetActiveLayout()
        if err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("No active dashboard found: %v", err)), nil
        }

        // 3. Get the schema and find the widget by exact ID
        schema := layout.Schema.Data()
        widgets, exists := schema.Layouts[breakpoint]
        if !exists {
            return mcp.NewToolResultError(fmt.Sprintf("Breakpoint %s not found", breakpoint)), nil
        }

        // 4. Find widget by exact ID match
        widgetIndex := -1
        for i, widget := range widgets {
            if widget.I == widgetID {
                widgetIndex = i
                break
            }
        }

        if widgetIndex == -1 {
            return mcp.NewToolResultError(fmt.Sprintf("Widget with ID '%s' not found", widgetID)), nil
        }

        // 5. Perform the operation directly on the schema
        switch operation {
        case "move":
            x, _ := strconv.Atoi(request.GetString("x", "0"))
            y, _ := strconv.Atoi(request.GetString("y", "0"))
            widgets[widgetIndex].X = x
            widgets[widgetIndex].Y = y
        case "resize":
            w, _ := strconv.Atoi(request.GetString("w", "0"))
            h, _ := strconv.Atoi(request.GetString("h", "0"))
            widgets[widgetIndex].W = w
            widgets[widgetIndex].H = h
        case "remove":
            widgets = append(widgets[:widgetIndex], widgets[widgetIndex+1:]...)
        }

        // 6. Update schema and save to database
        schema.Layouts[breakpoint] = widgets
        layout.Schema = datatypes.NewJSONType(schema)

        if err := layoutRepo.Update(layout); err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("Error saving layout: %v", err)), nil
        }

        // 7. Return precise response with widget state
        return mcp.NewToolResultText(responseJSON), nil
    }
}
```

## Response Format for UI Synchronization

```json
{
  "success": true,
  "operation": "manipulate_widget",
  "activeLayoutId": "dashboard-main",
  "targetedWidgets": ["widget-2"],
  "allChanges": [
    {
      "widgetId": "widget-1",
      "action": "removed",
      "breakpoint": "lg",
      "wasTargeted": false,
      "reason": "explicit_action",
      "previousState": {
        "x": 0, "y": 0, "w": 4, "h": 3,
        "componentType": "chart",
        "props": {"title": "Sales Chart"}
      }
    },
    {
      "widgetId": "widget-2",
      "action": "moved",
      "breakpoint": "lg",
      "wasTargeted": true,
      "reason": "user_requested",
      "previousState": {"x": 4, "y": 0, "w": 2, "h": 2},
      "newState": {"x": 0, "y": 0, "w": 2, "h": 2}
    },
    {
      "widgetId": "widget-3",
      "action": "repositioned",
      "breakpoint": "lg",
      "wasTargeted": false,
      "reason": "collision_avoidance",
      "previousState": {"x": 0, "y": 2, "w": 3, "h": 2},
      "newState": {"x": 0, "y": 1, "w": 3, "h": 2}
    },
    {
      "widgetId": "widget-4",
      "action": "repositioned",
      "breakpoint": "lg",
      "wasTargeted": false,
      "reason": "layout_compaction",
      "previousState": {"x": 6, "y": 4, "w": 2, "h": 1},
      "newState": {"x": 6, "y": 3, "w": 2, "h": 1}
    }
  ],
  "summary": {
    "totalAffected": 4,
    "targeted": 1,
    "collateralChanges": 3,
    "operations": {
      "removed": 1,
      "moved": 1,
      "repositioned": 2
    }
  },
  "message": "Moved customer table to top left. This caused 3 other widgets to reposition automatically.",
  "layoutVersion": "1.2.4",
  "affectedBreakpoints": ["lg"],
  "timestamp": "2025-09-19T10:30:00Z"
}
```

## Precise Widget Manipulation Examples

### Workflow Pattern
1. **Get Dashboard State**: "Show me the current dashboard"
   → Calls `get_active_dashboard` → Returns widgets with exact IDs
2. **Identify Target**: Review widget list, find `widget-1727012345678`
3. **Precise Operation**: Use exact widget ID for manipulation

### Remove Operations
- User: "Remove the revenue chart"
- LLM: Calls `get_active_dashboard`, identifies `widget-1727012345678` as revenue chart
- LLM: Calls `manipulate_widget` with `widget_id: "widget-1727012345678"`, `operation: "remove"`

### Resize Operations
- User: "Make the customer table larger"
- LLM: Calls `get_active_dashboard`, identifies `widget-1727012345679` as customer table
- LLM: Calls `manipulate_widget` with `widget_id: "widget-1727012345679"`, `operation: "resize"`, `w: "8"`, `h: "6"`

### Move Operations
- User: "Move the sales chart to position x=6, y=0"
- LLM: Calls `get_active_dashboard`, identifies `widget-1727012345680` as sales chart
- LLM: Calls `manipulate_widget` with `widget_id: "widget-1727012345680"`, `operation: "move"`, `x: "6"`, `y: "0"`

### Add Operations
- User: "Add a metrics widget showing total revenue"
- LLM: Calls `add_widget` with `widget_description: "total revenue metrics"`, `component_type: "metric"`, `props: "{\"title\": \"Total Revenue\", \"value\": \"$125,000\"}"`

## Handling Collateral Widget Movements

### Problem: Cascade Effects in Grid Layouts

When performing operations like move, resize, or remove, React Grid Layout automatically adjusts other widgets:

1. **Collision Avoidance**: Moving widget A pushes widget B out of the way
2. **Layout Compaction**: Removing widget A causes widgets below to move up
3. **Boundary Constraints**: Resizing widget A forces widget B to wrap to next row
4. **Responsive Adjustments**: Changes in one breakpoint affect widget positions in others

### Solution: Complete Change Tracking

#### API Integration Pattern

```go
type WidgetOperationResult struct {
    TargetedWidgets   []string      `json:"targetedWidgets"`   // Widgets user explicitly wanted to change
    AllChanges        []WidgetChange `json:"allChanges"`        // All widgets that actually changed
    CollateralChanges []WidgetChange `json:"collateralChanges"` // Only the unintended changes
    Summary           ChangeSummary  `json:"summary"`
}

type WidgetChange struct {
    WidgetID      string                 `json:"widgetId"`
    Action        string                 `json:"action"`        // "moved", "resized", "removed", "repositioned"
    Breakpoint    string                 `json:"breakpoint"`
    WasTargeted   bool                   `json:"wasTargeted"`   // Was this widget explicitly targeted by user?
    Reason        string                 `json:"reason"`        // Why did this change occur?
    PreviousState map[string]interface{} `json:"previousState"`
    NewState      map[string]interface{} `json:"newState,omitempty"`
}

type ChangeSummary struct {
    TotalAffected     int                    `json:"totalAffected"`
    Targeted          int                    `json:"targeted"`
    CollateralChanges int                    `json:"collateralChanges"`
    Operations        map[string]int         `json:"operations"`     // Count by operation type
    Reasons           map[string]int         `json:"reasons"`        // Count by reason
}
```

#### Change Tracking Implementation

```go
func (h *LayoutHandler) MoveWidget(layoutID, widgetID string, newPos Position) (*WidgetOperationResult, error) {
    // 1. Capture current layout state
    beforeLayout, err := h.layoutRepo.GetLayout(layoutID)
    if err != nil {
        return nil, err
    }

    // 2. Apply the requested change
    err = h.apiClient.MoveWidget(layoutID, widgetID, newPos)
    if err != nil {
        return nil, err
    }

    // 3. Capture new layout state
    afterLayout, err := h.layoutRepo.GetLayout(layoutID)
    if err != nil {
        return nil, err
    }

    // 4. Compare states to identify all changes
    changes := h.analyzeLayoutChanges(beforeLayout, afterLayout, []string{widgetID})

    return &WidgetOperationResult{
        TargetedWidgets:   []string{widgetID},
        AllChanges:        changes.All,
        CollateralChanges: changes.Collateral,
        Summary:           changes.Summary,
    }, nil
}

func (h *LayoutHandler) analyzeLayoutChanges(before, after *Layout, targetedIDs []string) *ChangeAnalysis {
    changes := []WidgetChange{}
    targeted := make(map[string]bool)

    for _, id := range targetedIDs {
        targeted[id] = true
    }

    // Compare each breakpoint
    for breakpoint := range after.Schema.Layouts {
        beforeItems := before.Schema.Layouts[breakpoint]
        afterItems := after.Schema.Layouts[breakpoint]

        // Track changes
        for _, afterItem := range afterItems {
            beforeItem := findWidget(beforeItems, afterItem.I)

            if beforeItem == nil {
                // Widget was added
                changes = append(changes, WidgetChange{
                    WidgetID:    afterItem.I,
                    Action:      "added",
                    Breakpoint:  breakpoint,
                    WasTargeted: targeted[afterItem.I],
                    Reason:      h.determineChangeReason(afterItem, nil, targeted[afterItem.I]),
                    NewState:    widgetToMap(afterItem),
                })
            } else if !widgetsEqual(beforeItem, afterItem) {
                // Widget was modified
                action := h.determineAction(beforeItem, afterItem)
                changes = append(changes, WidgetChange{
                    WidgetID:      afterItem.I,
                    Action:        action,
                    Breakpoint:    breakpoint,
                    WasTargeted:   targeted[afterItem.I],
                    Reason:        h.determineChangeReason(afterItem, beforeItem, targeted[afterItem.I]),
                    PreviousState: widgetToMap(beforeItem),
                    NewState:      widgetToMap(afterItem),
                })
            }
        }

        // Check for removed widgets
        for _, beforeItem := range beforeItems {
            if findWidget(afterItems, beforeItem.I) == nil {
                changes = append(changes, WidgetChange{
                    WidgetID:      beforeItem.I,
                    Action:        "removed",
                    Breakpoint:    breakpoint,
                    WasTargeted:   targeted[beforeItem.I],
                    Reason:        h.determineChangeReason(nil, beforeItem, targeted[beforeItem.I]),
                    PreviousState: widgetToMap(beforeItem),
                })
            }
        }
    }

    return h.categorizeChanges(changes, targetedIDs)
}

func (h *LayoutHandler) determineChangeReason(after, before *LayoutItem, wasTargeted bool) string {
    if wasTargeted {
        return "user_requested"
    }

    if before == nil {
        return "user_added"
    }

    if after == nil {
        return "user_removed"
    }

    // Analyze what type of automatic change occurred
    if after.X != before.X || after.Y != before.Y {
        if after.Y < before.Y {
            return "layout_compaction"  // Widget moved up due to space freed above
        }
        return "collision_avoidance"   // Widget moved due to collision
    }

    if after.W != before.W || after.H != before.H {
        return "boundary_constraint"   // Widget resized due to layout constraints
    }

    return "layout_optimization"       // Other automatic layout adjustments
}
```

### Enhanced MCP Response Messages

#### Detailed User Feedback

```go
func (h *MCPHandler) generateUserMessage(result *WidgetOperationResult) string {
    target := len(result.TargetedWidgets)
    total := result.Summary.TotalAffected
    collateral := result.Summary.CollateralChanges

    if collateral == 0 {
        return fmt.Sprintf("Successfully modified %d widget(s) as requested.", target)
    }

    msg := fmt.Sprintf("Successfully modified %d widget(s). ", target)

    if collateral == 1 {
        msg += "This caused 1 other widget to automatically reposition."
    } else {
        msg += fmt.Sprintf("This caused %d other widgets to automatically reposition.", collateral)
    }

    // Add specific details about why widgets moved
    reasons := []string{}
    if count := result.Summary.Reasons["collision_avoidance"]; count > 0 {
        reasons = append(reasons, fmt.Sprintf("%d moved to avoid collisions", count))
    }
    if count := result.Summary.Reasons["layout_compaction"]; count > 0 {
        reasons = append(reasons, fmt.Sprintf("%d moved up to fill empty space", count))
    }
    if count := result.Summary.Reasons["boundary_constraint"]; count > 0 {
        reasons = append(reasons, fmt.Sprintf("%d adjusted due to layout boundaries", count))
    }

    if len(reasons) > 0 {
        msg += " (" + strings.Join(reasons, ", ") + ")"
    }

    return msg
}
```

#### Example Enhanced Messages

**Simple case (no collateral changes):**
> "Moved sales chart to top right as requested."

**With collateral changes:**
> "Moved sales chart to top right. This caused 3 other widgets to automatically reposition (2 moved to avoid collisions, 1 moved up to fill empty space)."

**Complex case:**
> "Removed analytics widget. This caused 5 widgets to reposition automatically (3 moved up due to layout compaction, 2 adjusted to avoid new collisions)."

### Frontend Integration Requirements

#### Full State Synchronization

The frontend must be prepared to update ALL affected widgets, not just the targeted ones:

```javascript
// Handle MCP response with collateral changes
function handleWidgetOperationResponse(response) {
    const { allChanges, targetedWidgets, summary } = response;

    // Update all affected widgets in the layout
    allChanges.forEach(change => {
        const widget = findWidget(change.widgetId);

        switch (change.action) {
            case 'removed':
                removeWidget(change.widgetId);
                break;
            case 'moved':
            case 'repositioned':
                updateWidgetPosition(change.widgetId, change.newState);
                break;
            case 'resized':
                updateWidgetSize(change.widgetId, change.newState);
                break;
        }

        // Visual feedback for different types of changes
        if (change.wasTargeted) {
            highlightWidget(change.widgetId, 'primary-change');
        } else {
            highlightWidget(change.widgetId, 'collateral-change');
        }
    });

    // Show summary message to user
    showNotification(response.message, summary.collateralChanges > 0 ? 'info' : 'success');
}
```

This comprehensive approach ensures that users understand the full impact of their actions and the UI stays perfectly synchronized with all layout changes, both intended and automatic.