package models

// ComponentDefinition represents a component type definition with constraints
type ComponentDefinition struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	DefaultSize ItemSize               `json:"defaultSize"`
	Constraints ItemConstraints        `json:"constraints"`
	PropSchema  map[string]interface{} `json:"propSchema"` // JSON Schema for props validation
}

// ComponentRegistry holds predefined component types with their constraints
var ComponentRegistry = map[string]ComponentDefinition{
	"chart": {
		Type:        "chart",
		Name:        "Chart Widget",
		Description: "Data visualization charts",
		DefaultSize: ItemSize{W: 4, H: 3},
		Constraints: ItemConstraints{
			MinW: 2, MaxW: 12,
			MinH: 2, MaxH: 8,
			AspectRatio: &AspectRatio{Min: 0.5, Max: 3.0},
		},
		PropSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Chart title",
				},
				"chartType": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"line", "bar", "pie", "area", "scatter"},
					"description": "Type of chart visualization",
				},
				"dataSource": map[string]interface{}{
					"type":        "string",
					"description": "Data source URL or endpoint",
				},
				"refreshInterval": map[string]interface{}{
					"type":        "integer",
					"minimum":     1000,
					"description": "Refresh interval in milliseconds",
				},
			},
		},
	},
	"table": {
		Type:        "table",
		Name:        "Data Table",
		Description: "Tabular data display",
		DefaultSize: ItemSize{W: 6, H: 4},
		Constraints: ItemConstraints{
			MinW: 4, MaxW: 12,
			MinH: 3, MaxH: 10,
		},
		PropSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Table title",
				},
				"dataSource": map[string]interface{}{
					"type":        "string",
					"description": "Data source URL or endpoint",
				},
				"columns": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "Column names to display",
				},
				"filters": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"field":    map[string]string{"type": "string"},
							"operator": map[string]string{"type": "string"},
							"value":    map[string]string{"type": "string"},
						},
					},
					"description": "Table filters",
				},
				"sortable": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether columns are sortable",
				},
			},
		},
	},
	"metric": {
		Type:        "metric",
		Name:        "Metric Display",
		Description: "Single metric or KPI",
		DefaultSize: ItemSize{W: 2, H: 2},
		Constraints: ItemConstraints{
			MinW: 1, MaxW: 4,
			MinH: 1, MaxH: 3,
			AspectRatio: &AspectRatio{Min: 0.8, Max: 2.0},
		},
		PropSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Metric title",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "Current metric value",
				},
				"unit": map[string]interface{}{
					"type":        "string",
					"description": "Unit of measurement",
				},
				"dataSource": map[string]interface{}{
					"type":        "string",
					"description": "Data source URL or endpoint",
				},
				"threshold": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"warning":  map[string]string{"type": "number"},
						"critical": map[string]string{"type": "number"},
					},
					"description": "Alert thresholds",
				},
			},
		},
	},
	"text": {
		Type:        "text",
		Name:        "Text Widget",
		Description: "Static or dynamic text content",
		DefaultSize: ItemSize{W: 3, H: 2},
		Constraints: ItemConstraints{
			MinW: 1, MaxW: 8,
			MinH: 1, MaxH: 6,
		},
		PropSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Text content (supports markdown)",
				},
				"fontSize": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"small", "medium", "large", "xlarge"},
					"description": "Font size",
				},
				"alignment": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"left", "center", "right"},
					"description": "Text alignment",
				},
				"color": map[string]interface{}{
					"type":        "string",
					"description": "Text color (CSS color value)",
				},
			},
		},
	},
	"image": {
		Type:        "image",
		Name:        "Image Widget",
		Description: "Static or dynamic image display",
		DefaultSize: ItemSize{W: 3, H: 3},
		Constraints: ItemConstraints{
			MinW: 1, MaxW: 8,
			MinH: 1, MaxH: 6,
			AspectRatio: &AspectRatio{Min: 0.5, Max: 2.0},
		},
		PropSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"src": map[string]interface{}{
					"type":        "string",
					"description": "Image source URL",
				},
				"alt": map[string]interface{}{
					"type":        "string",
					"description": "Alternative text",
				},
				"fit": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"cover", "contain", "fill", "none", "scale-down"},
					"description": "How image should fit in container",
				},
			},
		},
	},
	"iframe": {
		Type:        "iframe",
		Name:        "Embedded Content",
		Description: "Iframe for embedding external content",
		DefaultSize: ItemSize{W: 4, H: 4},
		Constraints: ItemConstraints{
			MinW: 2, MaxW: 12,
			MinH: 2, MaxH: 8,
		},
		PropSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"src": map[string]interface{}{
					"type":        "string",
					"description": "Iframe source URL",
				},
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Iframe title",
				},
				"allowFullscreen": map[string]interface{}{
					"type":        "boolean",
					"description": "Allow fullscreen mode",
				},
				"sandbox": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "Sandbox restrictions",
				},
			},
		},
	},
}

// GetComponentTypes returns a list of all available component types
func GetComponentTypes() []string {
	types := make([]string, 0, len(ComponentRegistry))
	for t := range ComponentRegistry {
		types = append(types, t)
	}
	return types
}

// GetComponentDefinition returns the definition for a specific component type
func GetComponentDefinition(componentType string) (ComponentDefinition, bool) {
	def, exists := ComponentRegistry[componentType]
	return def, exists
}

// ValidateComponentType checks if a component type exists in the registry
func ValidateComponentType(componentType string) bool {
	_, exists := ComponentRegistry[componentType]
	return exists
}