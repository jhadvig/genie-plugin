package mcp

import (
	"fmt"
	"math"

	"github.com/layout-manager/api/pkg/models"
	"gorm.io/datatypes"
)

// WidgetManipulator handles the actual manipulation of widgets in layouts
type WidgetManipulator struct {
	finder *WidgetFinder
}

// NewWidgetManipulator creates a new widget manipulator
func NewWidgetManipulator() *WidgetManipulator {
	return &WidgetManipulator{
		finder: NewWidgetFinder(),
	}
}

// executeMovement handles moving widgets and returns all changes and affected widgets
func (wm *WidgetManipulator) executeMovement(layout *models.Layout, targetWidgets []FoundWidget, positionParams *PositionParams, breakpoint string) ([]WidgetChange, []WidgetInfo, error) {
	schema := layout.Schema.Data()

	widgets, exists := schema.Layouts[breakpoint]
	if !exists {
		return nil, nil, fmt.Errorf("breakpoint %s not found", breakpoint)
	}

	var allChanges []WidgetChange
	var affectedWidgets []WidgetInfo

	// Get grid columns for this breakpoint
	cols, exists := schema.Cols[breakpoint]
	if !exists {
		cols = 12 // Default to 12 columns
	}

	// Process each target widget
	for _, targetWidget := range targetWidgets {
		// Find the widget in the layout
		var widgetIndex = -1
		for i, widget := range widgets {
			if widget.I == targetWidget.ID {
				widgetIndex = i
				break
			}
		}

		if widgetIndex == -1 {
			continue // Widget not found, skip
		}

		widget := &widgets[widgetIndex]
		previousState := map[string]interface{}{
			"x": widget.X,
			"y": widget.Y,
			"w": widget.W,
			"h": widget.H,
		}

		// Calculate new position based on parameters
		newX, newY, err := wm.calculateNewPosition(widget, positionParams, widgets, cols)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to calculate new position for widget %s: %w", targetWidget.ID, err)
		}

		// Handle collision detection and widget repositioning
		collisionChanges, err := wm.handleCollisions(widgets, widgetIndex, newX, newY, widget.W, widget.H, cols, breakpoint)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to handle collisions: %w", err)
		}

		// Update the target widget position
		widget.X = newX
		widget.Y = newY

		newState := map[string]interface{}{
			"x": widget.X,
			"y": widget.Y,
			"w": widget.W,
			"h": widget.H,
		}

		// Record the primary change
		change := WidgetChange{
			WidgetID:      targetWidget.ID,
			Action:        "moved",
			Breakpoint:    breakpoint,
			WasTargeted:   true,
			Reason:        wm.getMovementReason(positionParams),
			PreviousState: previousState,
			NewState:      newState,
		}
		allChanges = append(allChanges, change)

		// Add collision-related changes
		allChanges = append(allChanges, collisionChanges...)

		// Add this widget to affected widgets
		affectedWidgets = append(affectedWidgets, WidgetInfo{
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
	}

	// Add affected widgets from collisions
	for _, change := range allChanges {
		if !change.WasTargeted {
			// Find the widget and add to affected list
			for _, widget := range widgets {
				if widget.I == change.WidgetID {
					affectedWidgets = append(affectedWidgets, WidgetInfo{
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
		}
	}

	// Update the schema
	schema.Layouts[breakpoint] = widgets
	layout.Schema = datatypes.NewJSONType(schema)

	return allChanges, affectedWidgets, nil
}

// calculateNewPosition calculates the new position for a widget based on position parameters
func (wm *WidgetManipulator) calculateNewPosition(widget *models.LayoutItem, positionParams *PositionParams, widgets []models.LayoutItem, cols int) (int, int, error) {
	if positionParams == nil {
		return widget.X, widget.Y, nil
	}

	// Handle zone-based positioning
	if positionParams.Zone != "" {
		return wm.calculateZonePosition(positionParams.Zone, widget.W, widget.H, cols)
	}

	// Handle relative positioning
	if positionParams.RelativeTo != "" {
		return wm.calculateRelativePosition(widget, positionParams, widgets, cols)
	}

	// Handle simple directional movement (e.g., "move left")
	if positionParams.Direction != "" {
		return wm.calculateDirectionalMovement(widget, positionParams, cols)
	}

	// No specific positioning, keep current position
	return widget.X, widget.Y, nil
}

// calculateZonePosition calculates position for zone-based movement (e.g., "top-left")
func (wm *WidgetManipulator) calculateZonePosition(zone string, width, height, cols int) (int, int, error) {
	switch zone {
	case "top-left":
		return 0, 0, nil
	case "top-right":
		return cols - width, 0, nil
	case "bottom-left":
		return 0, 10, nil // Assume row 10 as "bottom"
	case "bottom-right":
		return cols - width, 10, nil
	case "center":
		return (cols - width) / 2, 2, nil
	default:
		return 0, 0, fmt.Errorf("unknown zone: %s", zone)
	}
}

// calculateRelativePosition calculates position relative to another widget
func (wm *WidgetManipulator) calculateRelativePosition(widget *models.LayoutItem, positionParams *PositionParams, widgets []models.LayoutItem, cols int) (int, int, error) {
	// Find the reference widget
	var refWidget *models.LayoutItem
	for _, w := range widgets {
		if w.I == positionParams.RelativeTo {
			refWidget = &w
			break
		}
	}

	if refWidget == nil {
		return widget.X, widget.Y, fmt.Errorf("reference widget not found: %s", positionParams.RelativeTo)
	}

	// Calculate position based on direction
	switch positionParams.Direction {
	case "left":
		return int(math.Max(0, float64(refWidget.X-widget.W))), refWidget.Y, nil
	case "right":
		return int(math.Min(float64(cols-widget.W), float64(refWidget.X+refWidget.W))), refWidget.Y, nil
	case "above":
		return refWidget.X, int(math.Max(0, float64(refWidget.Y-widget.H))), nil
	case "below":
		return refWidget.X, refWidget.Y + refWidget.H, nil
	default:
		// Default to right side
		return int(math.Min(float64(cols-widget.W), float64(refWidget.X+refWidget.W))), refWidget.Y, nil
	}
}

// calculateDirectionalMovement handles simple directional movement commands
func (wm *WidgetManipulator) calculateDirectionalMovement(widget *models.LayoutItem, positionParams *PositionParams, cols int) (int, int, error) {
	newX := widget.X
	newY := widget.Y

	switch positionParams.Direction {
	case "left":
		// Move left by one grid unit
		if newX > 0 {
			newX = newX - 1
		}
	case "right":
		// Move right by one grid unit, ensuring we don't exceed grid bounds
		if newX+widget.W < cols {
			newX = newX + 1
		}
	case "top":
		// Move up by one grid unit
		if newY > 0 {
			newY = newY - 1
		}
	case "bottom":
		// Move down by one grid unit
		newY = newY + 1
	}

	return newX, newY, nil
}

// handleCollisions detects and resolves widget collisions
func (wm *WidgetManipulator) handleCollisions(widgets []models.LayoutItem, movingWidgetIndex, newX, newY, width, height, cols int, breakpoint string) ([]WidgetChange, error) {
	var changes []WidgetChange

	// Create a map of occupied cells after the move
	occupiedCells := make(map[string]bool)

	// Mark cells that will be occupied by the moving widget
	for x := newX; x < newX+width; x++ {
		for y := newY; y < newY+height; y++ {
			key := fmt.Sprintf("%d,%d", x, y)
			occupiedCells[key] = true
		}
	}

	// Check for collisions with other widgets
	for i, widget := range widgets {
		if i == movingWidgetIndex {
			continue // Skip the moving widget
		}

		// Check if this widget overlaps with the moving widget's new position
		collision := false
		for x := widget.X; x < widget.X+widget.W; x++ {
			for y := widget.Y; y < widget.Y+widget.H; y++ {
				key := fmt.Sprintf("%d,%d", x, y)
				if occupiedCells[key] {
					collision = true
					break
				}
			}
			if collision {
				break
			}
		}

		if collision {
			// Move the colliding widget down to avoid overlap
			previousState := map[string]interface{}{
				"x": widget.X,
				"y": widget.Y,
				"w": widget.W,
				"h": widget.H,
			}

			// Find a new position below the moving widget
			newWidgetY := newY + height
			widgets[i].Y = newWidgetY

			newState := map[string]interface{}{
				"x": widgets[i].X,
				"y": widgets[i].Y,
				"w": widgets[i].W,
				"h": widgets[i].H,
			}

			change := WidgetChange{
				WidgetID:      widget.I,
				Action:        "repositioned",
				Breakpoint:    breakpoint,
				WasTargeted:   false,
				Reason:        "moved to avoid collision",
				PreviousState: previousState,
				NewState:      newState,
			}
			changes = append(changes, change)
		}
	}

	return changes, nil
}

// getMovementReason generates a human-readable reason for the movement
func (wm *WidgetManipulator) getMovementReason(positionParams *PositionParams) string {
	if positionParams == nil {
		return "user requested movement"
	}

	if positionParams.Zone != "" {
		return fmt.Sprintf("moved to %s zone", positionParams.Zone)
	}

	if positionParams.RelativeTo != "" {
		direction := positionParams.Direction
		if direction == "" {
			direction = "next to"
		}
		return fmt.Sprintf("positioned %s %s", direction, positionParams.RelativeTo)
	}

	return "user requested movement"
}

// executeResize handles resizing widgets
func (wm *WidgetManipulator) executeResize(layout *models.Layout, targetWidgets []FoundWidget, sizeParams *SizeParams, breakpoint string) ([]WidgetChange, []WidgetInfo, error) {
	schema := layout.Schema.Data()

	widgets, exists := schema.Layouts[breakpoint]
	if !exists {
		return nil, nil, fmt.Errorf("breakpoint %s not found", breakpoint)
	}

	var allChanges []WidgetChange
	var affectedWidgets []WidgetInfo

	// Process each target widget
	for _, targetWidget := range targetWidgets {
		// Find the widget in the layout
		var widgetIndex = -1
		for i, widget := range widgets {
			if widget.I == targetWidget.ID {
				widgetIndex = i
				break
			}
		}

		if widgetIndex == -1 {
			continue // Widget not found, skip
		}

		widget := &widgets[widgetIndex]
		previousState := map[string]interface{}{
			"x": widget.X,
			"y": widget.Y,
			"w": widget.W,
			"h": widget.H,
		}

		// Calculate new size
		newW, newH := wm.calculateNewSize(widget, sizeParams)

		// Update the widget size
		widget.W = newW
		widget.H = newH

		newState := map[string]interface{}{
			"x": widget.X,
			"y": widget.Y,
			"w": widget.W,
			"h": widget.H,
		}

		// Record the change
		change := WidgetChange{
			WidgetID:      targetWidget.ID,
			Action:        "resized",
			Breakpoint:    breakpoint,
			WasTargeted:   true,
			Reason:        wm.getResizeReason(sizeParams),
			PreviousState: previousState,
			NewState:      newState,
		}
		allChanges = append(allChanges, change)

		// Add this widget to affected widgets
		affectedWidgets = append(affectedWidgets, WidgetInfo{
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
	}

	// Update the schema
	schema.Layouts[breakpoint] = widgets
	layout.Schema = datatypes.NewJSONType(schema)

	return allChanges, affectedWidgets, nil
}

// calculateNewSize calculates new widget dimensions based on size parameters
func (wm *WidgetManipulator) calculateNewSize(widget *models.LayoutItem, sizeParams *SizeParams) (int, int) {
	if sizeParams == nil {
		return widget.W, widget.H
	}

	newW := widget.W
	newH := widget.H

	if sizeParams.Mode == "absolute" {
		if sizeParams.Width != nil {
			newW = *sizeParams.Width
		}
		if sizeParams.Height != nil {
			newH = *sizeParams.Height
		}
	} else if sizeParams.Mode == "larger" || sizeParams.Mode == "smaller" {
		delta := 1
		if sizeParams.Delta != nil {
			delta = *sizeParams.Delta
		}
		if sizeParams.Mode == "smaller" {
			delta = -delta
		}
		newW = int(math.Max(1, float64(widget.W+delta)))
		newH = int(math.Max(1, float64(widget.H+delta)))
	}

	return newW, newH
}

// getResizeReason generates a human-readable reason for the resize
func (wm *WidgetManipulator) getResizeReason(sizeParams *SizeParams) string {
	if sizeParams == nil {
		return "user requested resize"
	}

	if sizeParams.Mode == "absolute" {
		return fmt.Sprintf("resized to %dx%d", *sizeParams.Width, *sizeParams.Height)
	}

	return fmt.Sprintf("resized %s", sizeParams.Mode)
}

// executeRemoval handles removing widgets from the layout
func (wm *WidgetManipulator) executeRemoval(layout *models.Layout, targetWidgets []FoundWidget, breakpoint string) ([]WidgetChange, []WidgetInfo, error) {
	schema := layout.Schema.Data()

	widgets, exists := schema.Layouts[breakpoint]
	if !exists {
		return nil, nil, fmt.Errorf("breakpoint %s not found", breakpoint)
	}

	var allChanges []WidgetChange
	var affectedWidgets []WidgetInfo

	// Process each target widget (in reverse order to maintain indices)
	for i := len(targetWidgets) - 1; i >= 0; i-- {
		targetWidget := targetWidgets[i]

		// Find the widget in the layout
		var widgetIndex = -1
		for j, widget := range widgets {
			if widget.I == targetWidget.ID {
				widgetIndex = j
				break
			}
		}

		if widgetIndex == -1 {
			continue // Widget not found, skip
		}

		widget := widgets[widgetIndex]
		previousState := map[string]interface{}{
			"x": widget.X,
			"y": widget.Y,
			"w": widget.W,
			"h": widget.H,
		}

		// Remove the widget from the layout
		widgets = append(widgets[:widgetIndex], widgets[widgetIndex+1:]...)

		// Record the change
		change := WidgetChange{
			WidgetID:      targetWidget.ID,
			Action:        "removed",
			Breakpoint:    breakpoint,
			WasTargeted:   true,
			Reason:        "user requested removal",
			PreviousState: previousState,
			NewState:      nil,
		}
		allChanges = append(allChanges, change)
	}

	// Update the schema
	schema.Layouts[breakpoint] = widgets
	layout.Schema = datatypes.NewJSONType(schema)

	return allChanges, affectedWidgets, nil
}