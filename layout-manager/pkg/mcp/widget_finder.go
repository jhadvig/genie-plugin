package mcp

import (
	"fmt"
	"strings"

	"github.com/layout-manager/api/pkg/models"
)

// WidgetFinder provides utilities for finding widgets based on various criteria
type WidgetFinder struct {
	parser *IntentParser
}

// NewWidgetFinder creates a new widget finder
func NewWidgetFinder() *WidgetFinder {
	return &WidgetFinder{
		parser: NewIntentParser(),
	}
}

// FindWidgets finds widgets in a layout based on a matcher
func (wf *WidgetFinder) FindWidgets(layout *models.Layout, matcher *WidgetMatcher, breakpoint string) ([]FoundWidget, error) {
	var foundWidgets []FoundWidget

	// Get the layout schema
	schema := layout.Schema.Data()

	// Get widgets for the specified breakpoint
	widgets, exists := schema.Layouts[breakpoint]
	if !exists {
		return foundWidgets, nil // No widgets in this breakpoint
	}

	// Check each widget against the matcher
	for _, widget := range widgets {
		if wf.matchesWidget(widget, matcher) {
			foundWidget := FoundWidget{
				ID:            widget.I,
				ComponentType: string(widget.ComponentType),
				Position: WidgetPosition{
					X: widget.X,
					Y: widget.Y,
					W: widget.W,
					H: widget.H,
				},
				Props:       widget.Props,
				MatchReason: wf.getMatchReason(widget, matcher),
				Breakpoint:  breakpoint,
			}
			foundWidgets = append(foundWidgets, foundWidget)
		}
	}

	return foundWidgets, nil
}

// FindWidgetsByDescription finds widgets using natural language description
func (wf *WidgetFinder) FindWidgetsByDescription(layout *models.Layout, description, breakpoint string) ([]FoundWidget, error) {
	// Parse the description into a matcher
	matcher := wf.parser.ParseWidgetSelector(description)

	// Find matching widgets
	return wf.FindWidgets(layout, matcher, breakpoint)
}

// matchesWidget checks if a widget matches the given criteria
func (wf *WidgetFinder) matchesWidget(widget models.LayoutItem, matcher *WidgetMatcher) bool {
	// Check component type
	if matcher.ComponentType != "" && string(widget.ComponentType) != matcher.ComponentType {
		return false
	}

	// Check widget ID exact match
	if matcher.WidgetID != "" && widget.I != matcher.WidgetID {
		return false
	}

	// Check title contains
	if matcher.TitleContains != "" {
		if !wf.checkTitleContains(widget, matcher.TitleContains) {
			return false
		}
	}

	// Check props contain
	if len(matcher.PropsContain) > 0 {
		if !wf.checkPropsContain(widget, matcher.PropsContain) {
			return false
		}
	}

	// Check position zone
	if matcher.PositionZone != "" {
		if !wf.checkPositionZone(widget, matcher.PositionZone) {
			return false
		}
	}

	// Check size range
	if matcher.SizeRange != nil {
		if !wf.checkSizeRange(widget, matcher.SizeRange) {
			return false
		}
	}

	return true
}

// checkTitleContains checks if widget title contains the specified text
func (wf *WidgetFinder) checkTitleContains(widget models.LayoutItem, contains string) bool {
	if widget.Props == nil {
		return false
	}

	// Check title property
	if title, exists := widget.Props["title"]; exists {
		if titleStr, ok := title.(string); ok {
			return strings.Contains(strings.ToLower(titleStr), strings.ToLower(contains))
		}
	}

	// Check other text properties that might contain the search term
	textFields := []string{"name", "label", "description", "dataSource"}
	for _, field := range textFields {
		if value, exists := widget.Props[field]; exists {
			if valueStr, ok := value.(string); ok {
				if strings.Contains(strings.ToLower(valueStr), strings.ToLower(contains)) {
					return true
				}
			}
		}
	}

	return false
}

// checkPropsContain checks if widget props contain the specified key-value pairs
func (wf *WidgetFinder) checkPropsContain(widget models.LayoutItem, propsContain map[string]interface{}) bool {
	if widget.Props == nil {
		return false
	}

	for key, expectedValue := range propsContain {
		actualValue, exists := widget.Props[key]
		if !exists {
			return false
		}

		// Simple equality check (could be enhanced for more complex matching)
		if actualValue != expectedValue {
			return false
		}
	}

	return true
}

// checkPositionZone checks if widget is in the specified position zone
func (wf *WidgetFinder) checkPositionZone(widget models.LayoutItem, zone string) bool {
	// For simplicity, assume a 12-column grid
	gridCols := 12

	// Define zones based on grid position
	switch zone {
	case "top-left":
		return widget.X < gridCols/2 && widget.Y < 2
	case "top-right":
		return widget.X >= gridCols/2 && widget.Y < 2
	case "bottom-left":
		return widget.X < gridCols/2 && widget.Y >= 2
	case "bottom-right":
		return widget.X >= gridCols/2 && widget.Y >= 2
	case "left":
		return widget.X < gridCols/2
	case "right":
		return widget.X >= gridCols/2
	case "top":
		return widget.Y < 2
	case "bottom":
		return widget.Y >= 2
	case "center":
		return widget.X >= gridCols/4 && widget.X < 3*gridCols/4
	}

	return false
}

// checkSizeRange checks if widget size is within the specified range
func (wf *WidgetFinder) checkSizeRange(widget models.LayoutItem, sizeRange *SizeRange) bool {
	// Check width constraints
	if sizeRange.MinWidth != nil && widget.W < *sizeRange.MinWidth {
		return false
	}
	if sizeRange.MaxWidth != nil && widget.W > *sizeRange.MaxWidth {
		return false
	}

	// Check height constraints
	if sizeRange.MinHeight != nil && widget.H < *sizeRange.MinHeight {
		return false
	}
	if sizeRange.MaxHeight != nil && widget.H > *sizeRange.MaxHeight {
		return false
	}

	return true
}

// getMatchReason generates a human-readable reason for why the widget matched
func (wf *WidgetFinder) getMatchReason(widget models.LayoutItem, matcher *WidgetMatcher) string {
	reasons := []string{}

	if matcher.ComponentType != "" {
		reasons = append(reasons, fmt.Sprintf("type is %s", matcher.ComponentType))
	}

	if matcher.TitleContains != "" {
		reasons = append(reasons, fmt.Sprintf("contains '%s' in title or properties", matcher.TitleContains))
	}

	if matcher.PositionZone != "" {
		reasons = append(reasons, fmt.Sprintf("located in %s area", matcher.PositionZone))
	}

	if matcher.SizeRange != nil {
		if matcher.SizeRange.MinWidth != nil || matcher.SizeRange.MinHeight != nil {
			reasons = append(reasons, "matches size criteria")
		}
	}

	if matcher.WidgetID != "" {
		reasons = append(reasons, "exact ID match")
	}

	if len(reasons) == 0 {
		return "matches all criteria"
	}

	return "Found " + strings.Join(reasons, " and ")
}

// CalculateGridPosition calculates a position for a new widget
func (wf *WidgetFinder) CalculateGridPosition(layout *models.Layout, breakpoint string, width, height int, positionHint string) (int, int, error) {
	// Get the layout schema
	schema := layout.Schema.Data()

	widgets, exists := schema.Layouts[breakpoint]
	if !exists {
		// No widgets yet, place at top-left
		return 0, 0, nil
	}

	// Get grid columns for this breakpoint
	cols, exists := schema.Cols[breakpoint]
	if !exists {
		cols = 12 // Default to 12 columns
	}

	// Create a grid map to track occupied cells
	occupiedCells := make(map[string]bool)
	for _, widget := range widgets {
		for x := widget.X; x < widget.X+widget.W; x++ {
			for y := widget.Y; y < widget.Y+widget.H; y++ {
				key := fmt.Sprintf("%d,%d", x, y)
				occupiedCells[key] = true
			}
		}
	}

	// Parse position hint
	if positionHint != "" {
		if x, y := wf.parsePositionHint(positionHint, cols); x >= 0 && y >= 0 {
			if wf.canPlaceWidget(x, y, width, height, cols, occupiedCells) {
				return x, y, nil
			}
		}
	}

	// Find the first available position
	for y := 0; y < 20; y++ { // Limit search to 20 rows
		for x := 0; x <= cols-width; x++ {
			if wf.canPlaceWidget(x, y, width, height, cols, occupiedCells) {
				return x, y, nil
			}
		}
	}

	// If no space found, place at the bottom
	maxY := 0
	for _, widget := range widgets {
		if widget.Y+widget.H > maxY {
			maxY = widget.Y + widget.H
		}
	}

	return 0, maxY, nil
}

// parsePositionHint converts position hints to coordinates
func (wf *WidgetFinder) parsePositionHint(hint string, cols int) (int, int) {
	hint = strings.ToLower(hint)

	switch hint {
	case "top left", "top-left":
		return 0, 0
	case "top right", "top-right":
		return cols - 4, 0 // Assume 4-width widget
	case "bottom left", "bottom-left":
		return 0, 10 // Assume placing at row 10
	case "bottom right", "bottom-right":
		return cols - 4, 10
	case "center":
		return cols/2 - 2, 2 // Center horizontally, row 2
	}

	return -1, -1 // Invalid hint
}

// canPlaceWidget checks if a widget can be placed at the given position
func (wf *WidgetFinder) canPlaceWidget(x, y, width, height, cols int, occupiedCells map[string]bool) bool {
	// Check if widget goes beyond grid boundaries
	if x+width > cols {
		return false
	}

	// Check if any cells are occupied
	for px := x; px < x+width; px++ {
		for py := y; py < y+height; py++ {
			key := fmt.Sprintf("%d,%d", px, py)
			if occupiedCells[key] {
				return false
			}
		}
	}

	return true
}

// FindWidgetsByDescriptionInList finds widgets in a list based on natural language description
func (wf *WidgetFinder) FindWidgetsByDescriptionInList(widgets []models.LayoutItem, description string) ([]FoundWidget, error) {
	var foundWidgets []FoundWidget

	// Simple matching logic for now - can be enhanced later
	descriptionLower := strings.ToLower(description)

	for _, widget := range widgets {
		// Check component type match
		if strings.Contains(descriptionLower, strings.ToLower(widget.ComponentType)) {
			foundWidget := FoundWidget{
				ID:            widget.I,
				ComponentType: widget.ComponentType,
				Position: WidgetPosition{
					X: widget.X,
					Y: widget.Y,
					W: widget.W,
					H: widget.H,
				},
				Props: widget.Props,
			}
			foundWidgets = append(foundWidgets, foundWidget)
			continue
		}

		// Check properties for title or other descriptive fields
		if widget.Props != nil {
			for key, value := range widget.Props {
				if strings.Contains(strings.ToLower(key), "title") ||
				   strings.Contains(strings.ToLower(key), "name") ||
				   strings.Contains(strings.ToLower(key), "label") {
					if valueStr, ok := value.(string); ok {
						if strings.Contains(strings.ToLower(valueStr), descriptionLower) ||
						   strings.Contains(descriptionLower, strings.ToLower(valueStr)) {
							foundWidget := FoundWidget{
								ID:            widget.I,
								ComponentType: widget.ComponentType,
								Position: WidgetPosition{
									X: widget.X,
									Y: widget.Y,
									W: widget.W,
									H: widget.H,
								},
								Props: widget.Props,
							}
							foundWidgets = append(foundWidgets, foundWidget)
							break
						}
					}
				}
			}
		}
	}

	return foundWidgets, nil
}