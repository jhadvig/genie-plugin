package mcp

import (
	"context"

	"github.com/layout-manager/api/pkg/db"
	"github.com/layout-manager/api/pkg/models"
)

// IntegrationBridge connects MCP tools to database operations
type IntegrationBridge struct {
	layoutRepo *db.LayoutRepository
}

// NewIntegrationBridge creates a new MCP-database integration bridge
func NewIntegrationBridge(layoutRepo *db.LayoutRepository) *IntegrationBridge {
	return &IntegrationBridge{
		layoutRepo: layoutRepo,
	}
}

// WidgetListResponse represents widget list response
type WidgetListResponse struct {
	Widgets    []LayoutItem `json:"widgets"`
	Breakpoint string       `json:"breakpoint"`
	Total      int          `json:"total"`
}

// LayoutItem represents a layout item for API responses
type LayoutItem struct {
	I             string                  `json:"i"`
	ComponentType ComponentType           `json:"componentType"`
	X             int                     `json:"x"`
	Y             int                     `json:"y"`
	W             int                     `json:"w"`
	H             int                     `json:"h"`
	Static        *bool                   `json:"static,omitempty"`
	IsDraggable   *bool                   `json:"isDraggable,omitempty"`
	IsResizable   *bool                   `json:"isResizable,omitempty"`
	MinW          *int                    `json:"minW,omitempty"`
	MaxW          *int                    `json:"maxW,omitempty"`
	MinH          *int                    `json:"minH,omitempty"`
	MaxH          *int                    `json:"maxH,omitempty"`
	Props         *map[string]interface{} `json:"props,omitempty"`
}

// ComponentType represents component types
type ComponentType string

const (
	ComponentTypeChart  ComponentType = "chart"
	ComponentTypeTable  ComponentType = "table"
	ComponentTypeMetric ComponentType = "metric"
	ComponentTypeText   ComponentType = "text"
	ComponentTypeImage  ComponentType = "image"
	ComponentTypeIframe ComponentType = "iframe"
)

// Layout represents a layout for API responses
type Layout struct {
	ID          string                 `json:"id"`
	LayoutID    string                 `json:"layoutId"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	IsActive    bool                   `json:"isActive"`
}

// FindLayoutWidgets lists widgets in a layout
func (b *IntegrationBridge) FindLayoutWidgets(ctx context.Context, layoutID, breakpoint string) (*WidgetListResponse, error) {
	// Get layout
	layout, err := b.layoutRepo.GetByLayoutID(layoutID)
	if err != nil {
		return nil, err
	}

	// Get schema directly (no need to unmarshal with JSONType)
	schema := layout.Schema.Data()

	// Get widgets for breakpoint
	widgets, exists := schema.Layouts[breakpoint]
	if !exists {
		widgets = []models.LayoutItem{}
	}

	// Convert to API format
	var apiWidgets []LayoutItem
	for _, widget := range widgets {
		staticPtr := &widget.Static
		isDraggablePtr := widget.IsDraggable
		isResizablePtr := widget.IsResizable

		apiWidget := LayoutItem{
			I:             widget.I,
			ComponentType: ComponentType(widget.ComponentType),
			X:             widget.X,
			Y:             widget.Y,
			W:             widget.W,
			H:             widget.H,
			Static:        staticPtr,
			IsDraggable:   isDraggablePtr,
			IsResizable:   isResizablePtr,
			MinW:          widget.MinW,
			MaxW:          widget.MaxW,
			MinH:          widget.MinH,
			MaxH:          widget.MaxH,
		}

		// Include properties if they exist
		if len(widget.Props) > 0 {
			apiWidget.Props = &widget.Props
		}

		apiWidgets = append(apiWidgets, apiWidget)
	}

	response := &WidgetListResponse{
		Widgets:    apiWidgets,
		Breakpoint: breakpoint,
		Total:      len(apiWidgets),
	}

	return response, nil
}
