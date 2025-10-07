package mcp

import (
	"time"

	"github.com/layout-manager/api/pkg/models"
)

// MCPRequest represents a request to an MCP tool
type MCPRequest struct {
	LayoutID                string `json:"layout_id"`
	Description             string `json:"description,omitempty"`
	Command                 string `json:"command,omitempty"`
	WidgetSelector          string `json:"widget_selector,omitempty"`
	WidgetDescription       string `json:"widget_description,omitempty"`
	Question                string `json:"question,omitempty"`
	Commands                []string `json:"commands,omitempty"`
	ComponentType           string `json:"component_type,omitempty"`
	PositionHint            string `json:"position_hint,omitempty"`
	SizeHint                string `json:"size_hint,omitempty"`
	Breakpoint              string `json:"breakpoint,omitempty"`
	ApplyToAllBreakpoints   bool   `json:"apply_to_all_breakpoints,omitempty"`
	Atomic                  bool   `json:"atomic,omitempty"`
}

// MCPResponse represents a response from an MCP tool
type MCPResponse struct {
	Success              bool                   `json:"success"`
	Operation            string                 `json:"operation"`
	LayoutID             string                 `json:"layoutId,omitempty"`            // Deprecated: use ActiveLayoutID
	ActiveLayoutID       string                 `json:"activeLayoutId,omitempty"`     // Active dashboard ID
	TargetedWidgets      []string               `json:"targetedWidgets,omitempty"`
	AllChanges           []WidgetChange         `json:"allChanges,omitempty"`
	Summary              *ChangeSummary         `json:"summary,omitempty"`
	Message              string                 `json:"message"`
	LayoutVersion        string                 `json:"layoutVersion,omitempty"`
	AffectedBreakpoints  []string               `json:"affectedBreakpoints,omitempty"`
	Timestamp            time.Time              `json:"timestamp"`

	// For find operations
	Widgets              []WidgetInfo           `json:"widgets,omitempty"`
	SearchQuery          string                 `json:"searchQuery,omitempty"`
	TotalFound           int                    `json:"totalFound,omitempty"`

	// For dashboard creation operations
	Layout               *LayoutInfo            `json:"layout,omitempty"`

    // For listing dashboards
    Layouts              []LayoutInfo           `json:"layouts,omitempty"`

	// For analysis operations
	Analysis             map[string]interface{} `json:"analysis,omitempty"`

	// Error handling
	Error                string                 `json:"error,omitempty"`
	Details              []string               `json:"details,omitempty"`
}

// WidgetInfo represents widget information in responses (unified for all operations)
type WidgetInfo struct {
	ID            string                 `json:"id"`
	ComponentType string                 `json:"componentType"`
	Position      WidgetPosition         `json:"position"`
	Props         map[string]interface{} `json:"props,omitempty"`
	MatchReason   string                 `json:"matchReason,omitempty"`
	Breakpoint    string                 `json:"breakpoint"`
}

// LayoutInfo represents layout information in responses
type LayoutInfo struct {
	ID          string    `json:"id"`          // UUID
	LayoutID    string    `json:"layoutId"`    // Human-readable ID
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// FoundWidget represents a widget found by search (deprecated - use WidgetInfo)
type FoundWidget struct {
	ID            string                 `json:"id"`
	ComponentType string                 `json:"componentType"`
	Position      WidgetPosition         `json:"position"`
	Props         map[string]interface{} `json:"props,omitempty"`
	MatchReason   string                 `json:"matchReason"`
	Breakpoint    string                 `json:"breakpoint"`
}

// WidgetPosition represents position and size in the grid
type WidgetPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// WidgetChange represents a change to a widget
type WidgetChange struct {
	WidgetID      string                 `json:"widgetId"`
	Action        string                 `json:"action"`        // "moved", "resized", "removed", "repositioned", "added"
	Breakpoint    string                 `json:"breakpoint"`
	WasTargeted   bool                   `json:"wasTargeted"`   // Was this widget explicitly targeted by user?
	Reason        string                 `json:"reason"`        // Why did this change occur?
	PreviousState map[string]interface{} `json:"previousState,omitempty"`
	NewState      map[string]interface{} `json:"newState,omitempty"`
}

// ChangeSummary provides statistics about changes
type ChangeSummary struct {
	TotalAffected     int            `json:"totalAffected"`
	Targeted          int            `json:"targeted"`
	CollateralChanges int            `json:"collateralChanges"`
	Operations        map[string]int `json:"operations"`     // Count by operation type
	Reasons           map[string]int `json:"reasons"`        // Count by reason
}

// Intent represents parsed user intention
type Intent struct {
	Action        string             // "remove", "resize", "move", "add", "update"
	Target        string             // Widget selector description
	Params        map[string]interface{} // Operation-specific parameters
	SizeParams    *SizeParams        // For resize operations
	PositionParams *PositionParams   // For move operations
	PropsParams   map[string]interface{} // For property updates
}

// SizeParams represents size change parameters
type SizeParams struct {
	Width  *int   `json:"width,omitempty"`
	Height *int   `json:"height,omitempty"`
	Delta  *int   `json:"delta,omitempty"`  // Relative change (+1, -1, etc.)
	Mode   string `json:"mode,omitempty"`   // "absolute", "relative", "larger", "smaller"
}

// PositionParams represents position change parameters
type PositionParams struct {
	X            *int   `json:"x,omitempty"`
	Y            *int   `json:"y,omitempty"`
	RelativeTo   string `json:"relativeTo,omitempty"`   // Widget ID to position relative to
	Direction    string `json:"direction,omitempty"`    // "above", "below", "left", "right"
	Zone         string `json:"zone,omitempty"`         // "top-left", "top-right", "bottom-left", "bottom-right"
}

// WidgetMatcher represents criteria for finding widgets
type WidgetMatcher struct {
	ComponentType    string                 `json:"componentType,omitempty"`
	TitleContains    string                 `json:"titleContains,omitempty"`
	PropsContain     map[string]interface{} `json:"propsContain,omitempty"`
	PositionZone     string                 `json:"positionZone,omitempty"`
	SizeRange        *SizeRange            `json:"sizeRange,omitempty"`
	WidgetID         string                 `json:"widgetId,omitempty"`
}

// SizeRange represents a range of widget sizes
type SizeRange struct {
	MinWidth  *int `json:"minWidth,omitempty"`
	MaxWidth  *int `json:"maxWidth,omitempty"`
	MinHeight *int `json:"minHeight,omitempty"`
	MaxHeight *int `json:"maxHeight,omitempty"`
}

// LayoutAnalysis represents analysis results
type LayoutAnalysis struct {
	TotalWidgets     int                            `json:"totalWidgets"`
	WidgetsByType    map[string]int                 `json:"widgetsByType"`
	WidgetsByZone    map[string][]string            `json:"widgetsByZone"`
	GridDimensions   *GridDimensions                `json:"gridDimensions"`
	Density          float64                        `json:"density"`
	Issues           []string                       `json:"issues,omitempty"`
	Suggestions      []string                       `json:"suggestions,omitempty"`
	WidgetDetails    []models.LayoutItem            `json:"widgetDetails,omitempty"`
}

// GridDimensions represents the layout's grid characteristics
type GridDimensions struct {
	Columns     int `json:"columns"`
	UsedRows    int `json:"usedRows"`
	MaxX        int `json:"maxX"`
	MaxY        int `json:"maxY"`
	TotalCells  int `json:"totalCells"`
	UsedCells   int `json:"usedCells"`
}