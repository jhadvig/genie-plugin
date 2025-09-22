package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Layout represents the main layout table with JSONB schema storage
type Layout struct {
	ID          uuid.UUID                        `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	LayoutID    string                           `gorm:"uniqueIndex;not null" json:"layout_id" validate:"required,min=1,max=255"`
	Name        string                           `json:"name" validate:"max=255"`
	Description string                           `json:"description" validate:"max=1000"`
	Schema      datatypes.JSONType[LayoutSchema] `gorm:"type:jsonb;not null" json:"schema" validate:"required"`
	IsActive    bool                             `gorm:"default:false;index" json:"is_active"`
	CreatedAt   time.Time                        `json:"created_at"`
	UpdatedAt   time.Time                        `json:"updated_at"`
}

// TableName returns the table name for GORM
func (Layout) TableName() string {
	return "layouts"
}

// GetDefaultEmptySchema returns a default empty layout schema
func GetDefaultEmptySchema() *LayoutSchema {
	return &LayoutSchema{
		Breakpoints: Breakpoints{
			"lg":  1200,
			"md":  996,
			"sm":  768,
			"xs":  480,
			"xxs": 0,
		},
		Cols: Cols{
			"lg":  12,
			"md":  10,
			"sm":  6,
			"xs":  4,
			"xxs": 2,
		},
		Layouts: Layouts{
			"lg":  []LayoutItem{},
			"md":  []LayoutItem{},
			"sm":  []LayoutItem{},
			"xs":  []LayoutItem{},
			"xxs": []LayoutItem{},
		},
		GlobalConstraints: GlobalConstraints{
			MaxItems:         20,
			DefaultItemSize:  ItemSize{W: 4, H: 3},
			Margin:           [2]int{10, 10},
			ContainerPadding: [2]int{10, 10},
		},
	}
}
