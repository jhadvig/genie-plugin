package models

import (
	"fmt"
)

// LayoutSchema represents the complete layout configuration with responsive breakpoints
type LayoutSchema struct {
	Breakpoints       Breakpoints       `json:"breakpoints" validate:"required"`
	Cols              Cols              `json:"cols" validate:"required"`
	Layouts           Layouts           `json:"layouts" validate:"required"`
	GlobalConstraints GlobalConstraints `json:"globalConstraints" validate:"required"`
}

// Responsive breakpoint configuration
type Breakpoints map[string]int

// Column configuration per breakpoint
type Cols map[string]int

// Layouts for each breakpoint
type Layouts map[string][]LayoutItem

// LayoutItem represents individual widget configuration in the grid
type LayoutItem struct {
	// Standard react-grid-layout fields
	I           string `json:"i" validate:"required"`
	X           int    `json:"x" validate:"min=0"`
	Y           int    `json:"y" validate:"min=0"`
	W           int    `json:"w" validate:"min=1"`
	H           int    `json:"h" validate:"min=1"`
	Static      bool   `json:"static,omitempty"`
	MinW        *int   `json:"minW,omitempty" validate:"omitempty,min=1"`
	MaxW        *int   `json:"maxW,omitempty" validate:"omitempty,min=1"`
	MinH        *int   `json:"minH,omitempty" validate:"omitempty,min=1"`
	MaxH        *int   `json:"maxH,omitempty" validate:"omitempty,min=1"`
	IsDraggable *bool  `json:"isDraggable,omitempty"`
	IsResizable *bool  `json:"isResizable,omitempty"`

	// Enhanced fields for component management
	ComponentType string                 `json:"componentType" validate:"required"`
	Props         map[string]interface{} `json:"props"`
	Constraints   *ItemConstraints       `json:"constraints,omitempty"`
}

// ItemConstraints represents size and proportion constraints for layout items
type ItemConstraints struct {
	MinW        int          `json:"minW"`
	MaxW        int          `json:"maxW"`
	MinH        int          `json:"minH"`
	MaxH        int          `json:"maxH"`
	AspectRatio *AspectRatio `json:"aspectRatio,omitempty"`
}

// AspectRatio represents aspect ratio constraints (width/height)
type AspectRatio struct {
	Min float64 `json:"min"` // Minimum width/height ratio
	Max float64 `json:"max"` // Maximum width/height ratio
}

// GlobalConstraints represents global layout constraints and defaults
type GlobalConstraints struct {
	MaxItems          int       `json:"maxItems"`
	DefaultItemSize   ItemSize  `json:"defaultItemSize"`
	Margin           [2]int     `json:"margin"`
	ContainerPadding [2]int     `json:"containerPadding"`
}

// ItemSize represents default size for new items
type ItemSize struct {
	W int `json:"w"`
	H int `json:"h"`
}

// Validate validates the layout schema structure
func (ls *LayoutSchema) Validate() error {
	// Check that all breakpoints have corresponding column definitions
	for bp := range ls.Breakpoints {
		if _, exists := ls.Cols[bp]; !exists {
			return fmt.Errorf("breakpoint %s missing column definition", bp)
		}
		if _, exists := ls.Layouts[bp]; !exists {
			return fmt.Errorf("breakpoint %s missing layout definition", bp)
		}
	}

	// Validate each layout item
	for breakpoint, items := range ls.Layouts {
		cols, exists := ls.Cols[breakpoint]
		if !exists {
			return fmt.Errorf("no column definition for breakpoint %s", breakpoint)
		}

		for _, item := range items {
			if err := item.Validate(cols); err != nil {
				return fmt.Errorf("invalid item %s in breakpoint %s: %w", item.I, breakpoint, err)
			}
		}
	}

	return nil
}

// ValidateWithRegistry validates using component registry constraints
func (ls *LayoutSchema) ValidateWithRegistry() error {
	if err := ls.Validate(); err != nil {
		return err
	}

	for _, items := range ls.Layouts {
		for _, item := range items {
			if def, exists := ComponentRegistry[item.ComponentType]; exists {
				if err := item.ValidateAgainstDefinition(def); err != nil {
					return fmt.Errorf("item %s violates %s constraints: %w", item.I, item.ComponentType, err)
				}
			}
		}
	}

	return nil
}

// Validate validates the layout item
func (li *LayoutItem) Validate(maxCols int) error {
	// Basic bounds checking
	if li.X < 0 || li.Y < 0 || li.W < 1 || li.H < 1 {
		return fmt.Errorf("invalid position or size: x=%d, y=%d, w=%d, h=%d", li.X, li.Y, li.W, li.H)
	}

	// Check grid bounds
	if li.X+li.W > maxCols {
		return fmt.Errorf("item extends beyond grid: x=%d, w=%d, maxCols=%d", li.X, li.W, maxCols)
	}

	// Validate min/max constraints
	if li.MinW != nil && li.W < *li.MinW {
		return fmt.Errorf("width %d below minimum %d", li.W, *li.MinW)
	}
	if li.MaxW != nil && li.W > *li.MaxW {
		return fmt.Errorf("width %d exceeds maximum %d", li.W, *li.MaxW)
	}
	if li.MinH != nil && li.H < *li.MinH {
		return fmt.Errorf("height %d below minimum %d", li.H, *li.MinH)
	}
	if li.MaxH != nil && li.H > *li.MaxH {
		return fmt.Errorf("height %d exceeds maximum %d", li.H, *li.MaxH)
	}

	// Validate min/max relationship
	if li.MinW != nil && li.MaxW != nil && *li.MinW > *li.MaxW {
		return fmt.Errorf("minW %d greater than maxW %d", *li.MinW, *li.MaxW)
	}
	if li.MinH != nil && li.MaxH != nil && *li.MinH > *li.MaxH {
		return fmt.Errorf("minH %d greater than maxH %d", *li.MinH, *li.MaxH)
	}

	// Validate constraints if present
	if li.Constraints != nil {
		if err := li.Constraints.Validate(); err != nil {
			return fmt.Errorf("constraint validation failed: %w", err)
		}

		// Check aspect ratio if defined
		if li.Constraints.AspectRatio != nil {
			ratio := float64(li.W) / float64(li.H)
			if ratio < li.Constraints.AspectRatio.Min || ratio > li.Constraints.AspectRatio.Max {
				return fmt.Errorf("aspect ratio %.2f outside allowed range %.2f-%.2f",
					ratio, li.Constraints.AspectRatio.Min, li.Constraints.AspectRatio.Max)
			}
		}
	}

	return nil
}

// ValidateAgainstDefinition validates item against component definition
func (li *LayoutItem) ValidateAgainstDefinition(def ComponentDefinition) error {
	// Check size constraints from component definition
	if li.W < def.Constraints.MinW || li.W > def.Constraints.MaxW {
		return fmt.Errorf("width %d outside component range %d-%d", li.W, def.Constraints.MinW, def.Constraints.MaxW)
	}
	if li.H < def.Constraints.MinH || li.H > def.Constraints.MaxH {
		return fmt.Errorf("height %d outside component range %d-%d", li.H, def.Constraints.MinH, def.Constraints.MaxH)
	}

	// Check aspect ratio constraints from component definition
	if def.Constraints.AspectRatio != nil {
		ratio := float64(li.W) / float64(li.H)
		if ratio < def.Constraints.AspectRatio.Min || ratio > def.Constraints.AspectRatio.Max {
			return fmt.Errorf("aspect ratio %.2f violates component constraints %.2f-%.2f",
				ratio, def.Constraints.AspectRatio.Min, def.Constraints.AspectRatio.Max)
		}
	}

	return nil
}

// Validate validates the item constraints
func (ic *ItemConstraints) Validate() error {
	if ic.MinW > ic.MaxW {
		return fmt.Errorf("minW %d greater than maxW %d", ic.MinW, ic.MaxW)
	}
	if ic.MinH > ic.MaxH {
		return fmt.Errorf("minH %d greater than maxH %d", ic.MinH, ic.MaxH)
	}
	if ic.AspectRatio != nil && ic.AspectRatio.Min > ic.AspectRatio.Max {
		return fmt.Errorf("aspect ratio min %.2f greater than max %.2f", ic.AspectRatio.Min, ic.AspectRatio.Max)
	}
	return nil
}