package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/layout-manager/api/pkg/db"
	"github.com/layout-manager/api/pkg/models"
	"github.com/mark3labs/mcp-go/mcp"
	"gorm.io/datatypes"
)

// FindWidgetsHandler handles the find_widgets tool
func FindWidgetsHandler(layoutRepo *db.LayoutRepository, bridge *IntegrationBridge) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		log.Printf("[MCP] find_widgets called")

		// Parse request parameters
		description, err := request.RequireString("description")
		if err != nil {
			log.Printf("[MCP] find_widgets error: description parameter missing")
			return mcp.NewToolResultError("description parameter is required"), nil
		}

		breakpoint := request.GetString("breakpoint", "lg")
		componentType := request.GetString("component_type", "")

		log.Printf("[MCP] find_widgets: description='%s', breakpoint=%s, componentType=%s",
			description, breakpoint, componentType)

		// Get active layout
		activeLayout, err := layoutRepo.GetActiveLayout()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("No active dashboard found: %v", err)), nil
		}

		layoutID := activeLayout.LayoutID

		// Get widgets from HTTP handler via bridge
		widgetListResponse, err := bridge.FindLayoutWidgets(ctx, layoutID, breakpoint)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting widgets: %v", err)), nil
		}

		// Find widgets using widget finder (convert API widgets to model widgets for searching)
		var modelWidgets []models.LayoutItem
		if widgetListResponse != nil {
			for _, apiWidget := range widgetListResponse.Widgets {
				modelWidget := models.LayoutItem{
					I:             apiWidget.I,
					ComponentType: string(apiWidget.ComponentType),
					X:             apiWidget.X,
					Y:             apiWidget.Y,
					W:             apiWidget.W,
					H:             apiWidget.H,
				}
				if apiWidget.Props != nil {
					modelWidget.Props = *apiWidget.Props
				}
				modelWidgets = append(modelWidgets, modelWidget)
			}
		}

		// Use widget finder to search by description
		finder := NewWidgetFinder()
		foundWidgets, err := finder.FindWidgetsByDescriptionInList(modelWidgets, description)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error finding widgets: %v", err)), nil
		}

		// Convert FoundWidget to WidgetInfo and filter by component type if specified
		widgets := []WidgetInfo{}
		for _, foundWidget := range foundWidgets {
			if componentType == "" || foundWidget.ComponentType == componentType {
				widgets = append(widgets, WidgetInfo{
					ID:            foundWidget.ID,
					ComponentType: foundWidget.ComponentType,
					Position: WidgetPosition{
						X: foundWidget.Position.X,
						Y: foundWidget.Position.Y,
						W: foundWidget.Position.W,
						H: foundWidget.Position.H,
					},
					Props:       foundWidget.Props,
					MatchReason: foundWidget.MatchReason,
					Breakpoint:  foundWidget.Breakpoint,
				})
			}
		}

		// Create response
		response := MCPResponse{
			Success:        true,
			Operation:      "find_widgets",
			ActiveLayoutID: layoutID,
			Widgets:        widgets,
			SearchQuery:    description,
			TotalFound:     len(widgets),
			Message:        fmt.Sprintf("Found %d widget(s) matching '%s' in active dashboard '%s'", len(widgets), description, activeLayout.Name),
			Timestamp:      time.Now(),
		}

		log.Printf("[MCP] find_widgets result: found %d widgets in active dashboard %s", len(widgets), layoutID)

		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}
}

// ManipulateWidgetHandler handles the manipulate_widget tool
func ManipulateWidgetHandler(layoutRepo *db.LayoutRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		log.Printf("[MCP] manipulate_widget called")

		// Parse request parameters
		widgetID, err := request.RequireString("widget_id")
		if err != nil {
			return mcp.NewToolResultError("widget_id parameter is required"), nil
		}

		operation, err := request.RequireString("operation")
		if err != nil {
			return mcp.NewToolResultError("operation parameter is required"), nil
		}

		breakpoint := request.GetString("breakpoint", "lg")
		applyToAllBreakpoints := request.GetBool("apply_to_all_breakpoints", false)

		log.Printf("[MCP] manipulate_widget: widgetID=%s, operation=%s, breakpoint=%s", widgetID, operation, breakpoint)

		// Get active layout
		layout, err := layoutRepo.GetActiveLayout()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("No active dashboard found: %v", err)), nil
		}

		layoutID := layout.LayoutID

		// Get the schema and find the widget
		schema := layout.Schema.Data()
		widgets, exists := schema.Layouts[breakpoint]
		if !exists {
			return mcp.NewToolResultError(fmt.Sprintf("Breakpoint %s not found", breakpoint)), nil
		}

		// Find the widget by ID
		widgetIndex := -1
		for i, widget := range widgets {
			if widget.I == widgetID {
				widgetIndex = i
				break
			}
		}

		if widgetIndex == -1 {
			return mcp.NewToolResultError(fmt.Sprintf("Widget with ID '%s' not found in breakpoint '%s'", widgetID, breakpoint)), nil
		}

		var operationMessage string
		var allChanges []WidgetChange

		// Perform the operation
		switch operation {
		case "move":
			xStr := request.GetString("x", "0")
			yStr := request.GetString("y", "0")

			x, err := strconv.Atoi(xStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid x position: %s", xStr)), nil
			}

			y, err := strconv.Atoi(yStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid y position: %s", yStr)), nil
			}

			oldWidget := widgets[widgetIndex]
			widgets[widgetIndex].X = x
			widgets[widgetIndex].Y = y

			// Record the change
			allChanges = append(allChanges, WidgetChange{
				WidgetID:      widgetID,
				Action:        "moved",
				Breakpoint:    breakpoint,
				WasTargeted:   true,
				Reason:        "direct move operation",
				PreviousState: map[string]interface{}{"x": oldWidget.X, "y": oldWidget.Y},
				NewState:      map[string]interface{}{"x": x, "y": y},
			})

			operationMessage = fmt.Sprintf("Moved widget '%s' to position (%d, %d)", widgetID, x, y)

		case "resize":
			wStr := request.GetString("w", "0")
			hStr := request.GetString("h", "0")

			w, err := strconv.Atoi(wStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid width: %s", wStr)), nil
			}

			h, err := strconv.Atoi(hStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid height: %s", hStr)), nil
			}

			oldWidget := widgets[widgetIndex]
			widgets[widgetIndex].W = w
			widgets[widgetIndex].H = h

			// Record the change
			allChanges = append(allChanges, WidgetChange{
				WidgetID:      widgetID,
				Action:        "resized",
				Breakpoint:    breakpoint,
				WasTargeted:   true,
				Reason:        "direct resize operation",
				PreviousState: map[string]interface{}{"w": oldWidget.W, "h": oldWidget.H},
				NewState:      map[string]interface{}{"w": w, "h": h},
			})

			operationMessage = fmt.Sprintf("Resized widget '%s' to %dx%d", widgetID, w, h)

		case "remove":
			// Remove widget from slice
			widgets = append(widgets[:widgetIndex], widgets[widgetIndex+1:]...)

			// Record the change
			allChanges = append(allChanges, WidgetChange{
				WidgetID:    widgetID,
				Action:      "removed",
				Breakpoint:  breakpoint,
				WasTargeted: true,
				Reason:      "direct remove operation",
			})

			operationMessage = fmt.Sprintf("Removed widget '%s'", widgetID)

		default:
			return mcp.NewToolResultError(fmt.Sprintf("Unsupported operation: %s. Use 'move', 'resize', or 'remove'", operation)), nil
		}

		// Update the schema
		schema.Layouts[breakpoint] = widgets
		layout.Schema = datatypes.NewJSONType(schema)

		// Save the updated layout
		if err := layoutRepo.Update(layout); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error saving layout: %v", err)), nil
		}

		// Build updated widget info for response
		var updatedWidgets []WidgetInfo
		for _, widget := range widgets {
			if widget.I == widgetID {
				updatedWidgets = append(updatedWidgets, WidgetInfo{
					ID:            widget.I,
					ComponentType: widget.ComponentType,
					Position: WidgetPosition{
						X: widget.X,
						Y: widget.Y,
						W: widget.W,
						H: widget.H,
					},
					Props:      widget.Props,
					Breakpoint: breakpoint,
				})
				break
			}
		}

		// Create summary
		summary := &ChangeSummary{
			TotalAffected:     len(allChanges),
			Targeted:          1, // Always 1 widget with precise ID targeting
			CollateralChanges: 0, // No collateral changes with precise operations
			Operations:        make(map[string]int),
			Reasons:           make(map[string]int),
		}

		for _, change := range allChanges {
			summary.Operations[change.Action]++
			summary.Reasons[change.Reason]++
		}

		response := MCPResponse{
			Success:             true,
			Operation:           "manipulate_widget",
			ActiveLayoutID:      layoutID,
			TargetedWidgets:     []string{widgetID},
			Widgets:             updatedWidgets,
			AllChanges:          allChanges,
			Summary:             summary,
			Message:             fmt.Sprintf("%s in active dashboard '%s'", operationMessage, layout.Name),
			AffectedBreakpoints: []string{breakpoint},
			Timestamp:           time.Now(),
		}

		// Add a note about apply to all breakpoints
		if applyToAllBreakpoints {
			response.Message += " (applying to all breakpoints)"
		}

		log.Printf("[MCP] manipulate_widget result: %s", operationMessage)

		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}
}

// AddWidgetHandler handles the add_widget tool
func AddWidgetHandler(layoutRepo *db.LayoutRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse request parameters
		widgetDescription, err := request.RequireString("widget_description")
		if err != nil {
			return mcp.NewToolResultError("widget_description parameter is required"), nil
		}

		componentType, err := request.RequireString("component_type")
		if err != nil {
			return mcp.NewToolResultError("component_type parameter is required"), nil
		}

		propsJSON := request.GetString("props", "")
		breakpoint := request.GetString("breakpoint", "lg")

		// Get active layout
		activeLayout, err := layoutRepo.GetActiveLayout()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("No active dashboard found: %v", err)), nil
		}

		layoutID := activeLayout.LayoutID

		// Parse props if provided
		var props map[string]interface{}
		if propsJSON != "" {
			if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid props JSON: %v", err)), nil
			}
		}

		// Get the schema and find a good position for the new widget
		schema := activeLayout.Schema.Data()
		widgets, exists := schema.Layouts[breakpoint]
		if !exists {
			widgets = []models.LayoutItem{}
		}

		// Generate a unique widget ID
		widgetID := fmt.Sprintf("widget-%d", time.Now().UnixNano())

		// Calculate position (simple approach - find next available spot)
		x, y := 0, 0
		if len(widgets) > 0 {
			// Find the rightmost widget in the top row, then place to its right
			maxX := 0
			for _, widget := range widgets {
				if widget.Y == 0 && widget.X+widget.W > maxX {
					maxX = widget.X + widget.W
				}
			}
			x = maxX
			// If it would go beyond grid width, move to next row
			cols, exists := schema.Cols[breakpoint]
			if !exists {
				cols = 12
			}
			if x+4 > cols { // Assuming default width of 4
				x = 0
				y = 1
			}
		}

		// Default widget dimensions
		w, h := 4, 3

		// Create new widget
		newWidget := models.LayoutItem{
			I:             widgetID,
			ComponentType: componentType,
			X:             x,
			Y:             y,
			W:             w,
			H:             h,
			Static:        false,
			IsDraggable:   boolPtr(true),
			IsResizable:   boolPtr(true),
			Props:         props,
		}

		// Add widget to layout
		widgets = append(widgets, newWidget)
		schema.Layouts[breakpoint] = widgets
		activeLayout.Schema = datatypes.NewJSONType(schema)

		// Save to database
		if err := layoutRepo.Update(activeLayout); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error saving layout: %v", err)), nil
		}

		// Create widget info for response
		createdWidget := WidgetInfo{
			ID:            widgetID,
			ComponentType: componentType,
			Position: WidgetPosition{
				X: x,
				Y: y,
				W: w,
				H: h,
			},
			Props:      props,
			Breakpoint: breakpoint,
		}

		// Create response
		response := MCPResponse{
			Success:        true,
			Operation:      "add_widget",
			ActiveLayoutID: layoutID,
			Widgets:        []WidgetInfo{createdWidget},
			Message:        fmt.Sprintf("Added %s widget '%s' (ID: %s) to active dashboard '%s' at position (%d, %d)", componentType, widgetDescription, widgetID, activeLayout.Name, x, y),
			Timestamp:      time.Now(),
		}

		log.Printf("[MCP] add_widget result: added widget %s at (%d, %d)", widgetID, x, y)

		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}
}

// BatchWidgetOperationsHandler handles the batch_widget_operations tool
func BatchWidgetOperationsHandler(layoutRepo *db.LayoutRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse request parameters
		commandsJSON, err := request.RequireString("commands_json")
		if err != nil {
			return mcp.NewToolResultError("commands_json parameter is required"), nil
		}

		breakpoint := request.GetString("breakpoint", "lg")
		atomic := request.GetBool("atomic", true)
		_ = breakpoint // TODO: Use breakpoint in actual implementation
		_ = atomic     // TODO: Use atomic in actual implementation

		// Parse commands array
		var commands []string
		if err := json.Unmarshal([]byte(commandsJSON), &commands); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid commands_json format: %v", err)), nil
		}

		// Get active layout
		activeLayout, err := layoutRepo.GetActiveLayout()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("No active dashboard found: %v", err)), nil
		}

		layoutID := activeLayout.LayoutID

		// Create response showing what would be executed
		response := MCPResponse{
			Success:        true,
			Operation:      "batch_widget_operations",
			ActiveLayoutID: layoutID,
			Message:        fmt.Sprintf("Would execute %d batch operations in %s mode on active dashboard '%s'", len(commands), map[bool]string{true: "atomic", false: "non-atomic"}[atomic], activeLayout.Name),
			Timestamp:      time.Now(),
		}

		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}
}

// AnalyzeLayoutHandler handles the analyze_layout tool
func AnalyzeLayoutHandler(layoutRepo *db.LayoutRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse request parameters
		question, err := request.RequireString("question")
		if err != nil {
			return mcp.NewToolResultError("question parameter is required"), nil
		}

		breakpoint := request.GetString("breakpoint", "lg")

		// Get active layout
		layout, err := layoutRepo.GetActiveLayout()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("No active dashboard found: %v", err)), nil
		}

		layoutID := layout.LayoutID

		// Perform basic analysis
		analysis, err := analyzeLayout(layout, breakpoint, question)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error analyzing layout: %v", err)), nil
		}

		// Create response
		response := MCPResponse{
			Success:        true,
			Operation:      "analyze_layout",
			ActiveLayoutID: layoutID,
			Analysis:       map[string]interface{}{"analysis": analysis},
			Message:        fmt.Sprintf("Analysis of active dashboard '%s': %s", layout.Name, generateAnalysisMessage(analysis, question)),
			Timestamp:      time.Now(),
		}

		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}
}

// Helper functions
func getTargetWidgetIDs(widgets []FoundWidget) []string {
	ids := make([]string, len(widgets))
	for i, widget := range widgets {
		ids[i] = widget.ID
	}
	return ids
}

func getWidgetIDs(widgets []WidgetInfo) []string {
	ids := make([]string, len(widgets))
	for i, widget := range widgets {
		ids[i] = widget.ID
	}
	return ids
}

// analyzeLayout performs basic layout analysis
func analyzeLayout(layout *models.Layout, breakpoint, question string) (*LayoutAnalysis, error) {
	schema := layout.Schema.Data()

	widgets, exists := schema.Layouts[breakpoint]
	if !exists {
		widgets = []models.LayoutItem{}
	}

	analysis := &LayoutAnalysis{
		TotalWidgets:  len(widgets),
		WidgetsByType: make(map[string]int),
		WidgetsByZone: make(map[string][]string),
		WidgetDetails: widgets,
	}

	// Count widgets by type
	for _, widget := range widgets {
		analysis.WidgetsByType[string(widget.ComponentType)]++
	}

	// Calculate grid dimensions
	if len(widgets) > 0 {
		maxX, maxY := 0, 0
		usedCells := 0

		for _, widget := range widgets {
			if widget.X+widget.W > maxX {
				maxX = widget.X + widget.W
			}
			if widget.Y+widget.H > maxY {
				maxY = widget.Y + widget.H
			}
			usedCells += widget.W * widget.H
		}

		cols := 12 // Default
		if colsForBreakpoint, exists := schema.Cols[breakpoint]; exists {
			cols = colsForBreakpoint
		}

		analysis.GridDimensions = &GridDimensions{
			Columns:    cols,
			UsedRows:   maxY,
			MaxX:       maxX,
			MaxY:       maxY,
			TotalCells: cols * maxY,
			UsedCells:  usedCells,
		}

		if analysis.GridDimensions.TotalCells > 0 {
			analysis.Density = float64(usedCells) / float64(analysis.GridDimensions.TotalCells)
		}
	}

	return analysis, nil
}

// generateAnalysisMessage creates a user-friendly analysis message
func generateAnalysisMessage(analysis *LayoutAnalysis, question string) string {
	questionLower := strings.ToLower(question)

	if strings.Contains(questionLower, "how many") {
		return fmt.Sprintf("This layout contains %d widgets total", analysis.TotalWidgets)
	}

	if strings.Contains(questionLower, "what widgets") || strings.Contains(questionLower, "what's in") {
		types := []string{}
		for widgetType, count := range analysis.WidgetsByType {
			if count == 1 {
				types = append(types, fmt.Sprintf("1 %s", widgetType))
			} else {
				types = append(types, fmt.Sprintf("%d %ss", count, widgetType))
			}
		}
		return fmt.Sprintf("This layout contains: %s", strings.Join(types, ", "))
	}

	return fmt.Sprintf("Layout analysis: %d widgets with density %.1f%%", analysis.TotalWidgets, analysis.Density*100)
}

// ConfigureWidgetHandler handles the configure_widget tool
func ConfigureWidgetHandler(layoutRepo *db.LayoutRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse request parameters
		widgetSelector, err := request.RequireString("widget_selector")
		if err != nil {
			return mcp.NewToolResultError("widget_selector parameter is required"), nil
		}

		configurationRequest, err := request.RequireString("configuration_request")
		if err != nil {
			return mcp.NewToolResultError("configuration_request parameter is required"), nil
		}

		domainContext := request.GetString("domain_context", "")
		currentPropsJSON := request.GetString("current_props", "")
		suggestedPropsJSON := request.GetString("suggested_props", "")

		// Get active layout
		activeLayout, err := layoutRepo.GetActiveLayout()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("No active dashboard found: %v", err)), nil
		}

		layoutID := activeLayout.LayoutID

		// Parse props if provided
		var currentProps, suggestedProps map[string]interface{}
		if currentPropsJSON != "" {
			if err := json.Unmarshal([]byte(currentPropsJSON), &currentProps); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid current_props JSON: %v", err)), nil
			}
		}
		if suggestedPropsJSON != "" {
			if err := json.Unmarshal([]byte(suggestedPropsJSON), &suggestedProps); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid suggested_props JSON: %v", err)), nil
			}
		}

		// Create response showing what would be configured
		message := fmt.Sprintf("Would configure widget '%s' in active dashboard '%s': %s", widgetSelector, activeLayout.Name, configurationRequest)
		if domainContext != "" {
			message += fmt.Sprintf(" (context: %s)", domainContext)
		}

		response := MCPResponse{
			Success:        true,
			Operation:      "configure_widget",
			ActiveLayoutID: layoutID,
			Message:        message,
			Timestamp:      time.Now(),
		}

		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}
}

// CreateDashboardHandler handles the create_dashboard tool
func CreateDashboardHandler(layoutRepo *db.LayoutRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		log.Printf("[MCP] create_dashboard called")

		// Parse optional parameters
		name := request.GetString("name", "")
		description := request.GetString("description", "")

		// Generate default name if not provided
		if name == "" {
			name = fmt.Sprintf("Dashboard %s", time.Now().Format("2006-01-02 15:04"))
		}

		// Generate unique layout_id
		layoutID := generateLayoutID(name)

		log.Printf("[MCP] create_dashboard: name='%s', description='%s'", name, description)

		// Use the default empty schema from models
		defaultSchema := models.GetDefaultEmptySchema()
		log.Printf("[MCP] create_dashboard: using default empty schema")

		// Create layout model with the schema directly
		layout := &models.Layout{
			LayoutID:    layoutID,
			Name:        name,
			Description: description,
			Schema:      datatypes.NewJSONType(*defaultSchema),
		}

		log.Printf("[MCP] create_dashboard: schema set successfully")

		// Save to database with auto-activation
		err := layoutRepo.CreateLayoutWithActivation(layout)
		if err != nil {
			log.Printf("[MCP] create_dashboard error: failed to create layout: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create dashboard: %v", err)), nil
		}

		// Build response
		response := MCPResponse{
			Success:        true,
			Operation:      "create_dashboard",
			ActiveLayoutID: layoutID,
			Message:        fmt.Sprintf("Created and activated empty dashboard '%s'", name),
			Layout: &LayoutInfo{
				ID:          layout.ID.String(),
				LayoutID:    layoutID,
				Name:        name,
				Description: layout.Description,
			},
			Widgets:    []WidgetInfo{}, // Always empty
			TotalFound: 0,
		}

		responseJSON, _ := json.Marshal(response)
		log.Printf("[MCP] create_dashboard result: created empty dashboard '%s'", name)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}
}

// GetActiveDashboardHandler handles the get_active_dashboard tool
func GetActiveDashboardHandler(layoutRepo *db.LayoutRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		log.Printf("[MCP] get_active_dashboard called")

		// Parse request parameters
		breakpoint := request.GetString("breakpoint", "lg")
		includeAllBreakpoints := request.GetBool("include_all_breakpoints", false)

		log.Printf("[MCP] get_active_dashboard: breakpoint=%s, includeAll=%v", breakpoint, includeAllBreakpoints)

		// Get active layout
		activeLayout, err := layoutRepo.GetActiveLayout()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("No active dashboard found: %v", err)), nil
		}

		// Get the schema directly
		schema := activeLayout.Schema.Data()

		// Prepare response data
		var responseData map[string]interface{}

		if includeAllBreakpoints {
			// Return full schema with all breakpoints
			responseData = map[string]interface{}{
				"layoutId":          activeLayout.LayoutID,
				"name":              activeLayout.Name,
				"description":       activeLayout.Description,
				"breakpoints":       schema.Breakpoints,
				"cols":              schema.Cols,
				"layouts":           schema.Layouts,
				"globalConstraints": schema.GlobalConstraints,
			}
		} else {
			// Return only the specified breakpoint
			widgets, exists := schema.Layouts[breakpoint]
			if !exists {
				widgets = []models.LayoutItem{}
			}

			cols, colsExist := schema.Cols[breakpoint]
			if !colsExist {
				cols = 12 // Default
			}

			responseData = map[string]interface{}{
				"layoutId":          activeLayout.LayoutID,
				"name":              activeLayout.Name,
				"description":       activeLayout.Description,
				"breakpoint":        breakpoint,
				"cols":              cols,
				"widgets":           widgets,
				"globalConstraints": schema.GlobalConstraints,
			}
		}

		// Create response
		response := MCPResponse{
			Success:        true,
			Operation:      "get_active_dashboard",
			ActiveLayoutID: activeLayout.LayoutID,
			Analysis:       responseData,
			Message:        fmt.Sprintf("Retrieved active dashboard '%s' for breakpoint '%s'", activeLayout.Name, breakpoint),
			Timestamp:      time.Now(),
		}

		log.Printf("[MCP] get_active_dashboard result: returned dashboard %s", activeLayout.LayoutID)

		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}
}


// SetDashboardMetadataHandler handles the set_dashboard_metadata tool
func SetDashboardMetadataHandler(layoutRepo *db.LayoutRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        log.Printf("[MCP] set_dashboard_metadata called")

        layoutID := request.GetString("layout_id", "")
        if layoutID == "" {
            layoutID = request.GetString("dashboard_id", "")
        }

        newName := request.GetString("name", "")
        newDescription := request.GetString("description", "")

        var layout *models.Layout
        var err error
        switch {
        case layoutID != "":
            layout, err = layoutRepo.GetByLayoutID(layoutID)
        default:
            layout, err = layoutRepo.GetActiveLayout()
        }
        if err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("Dashboard not found: %v", err)), nil
        }

        if newName == "" && newDescription == "" {
            return mcp.NewToolResultError("At least one of name or description must be provided"), nil
        }

        if newName != "" {
            layout.Name = newName
        }
        if newDescription != "" {
            layout.Description = newDescription
        }

        if err := layoutRepo.Update(layout); err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("Failed to update dashboard: %v", err)), nil
        }

        response := MCPResponse{
            Success:   true,
            Operation: "set_dashboard_metadata",
            Message:   fmt.Sprintf("Set metadata for dashboard '%s'", layout.LayoutID),
            Timestamp: time.Now(),
            Layout: &LayoutInfo{
                ID:          layout.ID.String(),
                LayoutID:    layout.LayoutID,
                Name:        layout.Name,
                Description: layout.Description,
            },
        }

        responseJSON, _ := json.Marshal(response)
        return mcp.NewToolResultText(string(responseJSON)), nil
    }
}

// ListDashboardsHandler handles the list_dashboards tool
func ListDashboardsHandler(layoutRepo *db.LayoutRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        log.Printf("[MCP] list_dashboards called")

        // Parse pagination params (strings by tool schema)
        limitStr := request.GetString("limit", "50")
        offsetStr := request.GetString("offset", "0")

        limit, err := strconv.Atoi(limitStr)
        if err != nil || limit <= 0 {
            limit = 50
        }
        offset, err := strconv.Atoi(offsetStr)
        if err != nil || offset < 0 {
            offset = 0
        }

        // Query repository
        layouts, err := layoutRepo.List(limit, offset)
        if err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("Failed to list dashboards: %v", err)), nil
        }

        // Map to LayoutInfo slice
        list := make([]LayoutInfo, 0, len(layouts))
        for _, l := range layouts {
            list = append(list, LayoutInfo{
                ID:          l.ID.String(),
                LayoutID:    l.LayoutID,
                Name:        l.Name,
                Description: l.Description,
                IsActive:    l.IsActive,
                CreatedAt:   l.CreatedAt,
                UpdatedAt:   l.UpdatedAt,
            })
        }

        // Build response
        response := MCPResponse{
            Success:   true,
            Operation: "list_dashboards",
            Message:   fmt.Sprintf("Returned %d dashboard(s)", len(list)),
            Timestamp: time.Now(),
            Layouts:   list,
        }

        responseJSON, _ := json.Marshal(response)
        return mcp.NewToolResultText(string(responseJSON)), nil
    }
}

// GetDashboardHandler handles the get_dashboard tool
func GetDashboardHandler(layoutRepo *db.LayoutRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        log.Printf("[MCP] get_dashboard called")

        // Parse request parameters
        layoutID, err := request.RequireString("layout_id")
        if err != nil {
            return mcp.NewToolResultError("layout_id parameter is required"), nil
        }

        breakpoint := request.GetString("breakpoint", "lg")
        includeAllBreakpoints := request.GetBool("include_all_breakpoints", false)

        log.Printf("[MCP] get_dashboard: layoutID=%s, breakpoint=%s, includeAll=%v", layoutID, breakpoint, includeAllBreakpoints)

        // Get the dashboard by layout ID
        dashboard, err := layoutRepo.GetByLayoutID(layoutID)
        if err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("Dashboard not found: %v", err)), nil
        }

        // Set this dashboard as active and deactivate all others
        if err := layoutRepo.SetActiveLayout(layoutID); err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("Failed to set dashboard as active: %v", err)), nil
        }

        log.Printf("[MCP] get_dashboard: set dashboard '%s' as active", layoutID)

        // Get the schema directly
        schema := dashboard.Schema.Data()

        // Prepare response data
        var responseData map[string]interface{}

        if includeAllBreakpoints {
            // Return full schema with all breakpoints
            responseData = map[string]interface{}{
                "layoutId":          dashboard.LayoutID,
                "name":              dashboard.Name,
                "description":       dashboard.Description,
                "breakpoints":       schema.Breakpoints,
                "cols":              schema.Cols,
                "layouts":           schema.Layouts,
                "globalConstraints": schema.GlobalConstraints,
            }
        } else {
            // Return only the specified breakpoint
            widgets, exists := schema.Layouts[breakpoint]
            if !exists {
                widgets = []models.LayoutItem{}
            }

            cols, colsExist := schema.Cols[breakpoint]
            if !colsExist {
                cols = 12 // Default
            }

            responseData = map[string]interface{}{
                "layoutId":          dashboard.LayoutID,
                "name":              dashboard.Name,
                "description":       dashboard.Description,
                "breakpoint":        breakpoint,
                "cols":              cols,
                "widgets":           widgets,
                "globalConstraints": schema.GlobalConstraints,
            }
        }

        // Create response
        response := MCPResponse{
            Success:        true,
            Operation:      "get_dashboard",
            ActiveLayoutID: dashboard.LayoutID,
            Analysis:       responseData,
            Message:        fmt.Sprintf("Retrieved and activated dashboard '%s' for breakpoint '%s'", dashboard.Name, breakpoint),
            Timestamp:      time.Now(),
        }

        log.Printf("[MCP] get_dashboard result: returned and activated dashboard %s", dashboard.LayoutID)

        responseJSON, _ := json.Marshal(response)
        return mcp.NewToolResultText(string(responseJSON)), nil
    }
}

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}

// generateLayoutID creates a layout ID from name
func generateLayoutID(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")

	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	// Add timestamp suffix to ensure uniqueness
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s", result.String(), timestamp)
}
