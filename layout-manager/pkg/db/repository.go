package db

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/layout-manager/api/pkg/models"
)

// LayoutRepository provides database operations for layouts
type LayoutRepository struct {
	db *gorm.DB
}

// NewLayoutRepository creates a new layout repository
func NewLayoutRepository(db *gorm.DB) *LayoutRepository {
	return &LayoutRepository{db: db}
}

// Create creates a new layout
func (r *LayoutRepository) Create(layout *models.Layout) error {
	return r.db.Create(layout).Error
}

// CreateLayoutWithActivation creates a new layout and sets it as active
func (r *LayoutRepository) CreateLayoutWithActivation(layout *models.Layout) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// First, deactivate all existing layouts
		if err := tx.Model(&models.Layout{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			return err
		}

		// Set the new layout as active
		layout.IsActive = true

		// Create the new layout
		return tx.Create(layout).Error
	})
}

// GetByID retrieves a layout by its UUID
func (r *LayoutRepository) GetByID(id string) (*models.Layout, error) {
	var layout models.Layout
	err := r.db.Where("id = ?", id).First(&layout).Error
	if err != nil {
		return nil, err
	}
	return &layout, nil
}

// GetByLayoutID retrieves a layout by its layout_id
func (r *LayoutRepository) GetByLayoutID(layoutID string) (*models.Layout, error) {
	var layout models.Layout
	err := r.db.Where("layout_id = ?", layoutID).First(&layout).Error
	if err != nil {
		return nil, err
	}
	return &layout, nil
}

// Update updates an existing layout
func (r *LayoutRepository) Update(layout *models.Layout) error {
	return r.db.Save(layout).Error
}

// Delete deletes a layout by layout_id
func (r *LayoutRepository) Delete(layoutID string) error {
	return r.db.Where("layout_id = ?", layoutID).Delete(&models.Layout{}).Error
}

// List retrieves layouts with pagination, ordered by active status first
func (r *LayoutRepository) List(limit, offset int) ([]models.Layout, error) {
	var layouts []models.Layout
	err := r.db.Order("is_active DESC, created_at DESC").Limit(limit).Offset(offset).Find(&layouts).Error
	return layouts, err
}

// SetActiveLayout sets a specific layout as active and deactivates others
func (r *LayoutRepository) SetActiveLayout(layoutID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Deactivate all layouts
		if err := tx.Model(&models.Layout{}).Where("1 = 1").Update("is_active", false).Error; err != nil {
			return err
		}

		// Activate the specified layout
		result := tx.Model(&models.Layout{}).Where("layout_id = ?", layoutID).Update("is_active", true)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

// GetActiveLayout returns the currently active layout
func (r *LayoutRepository) GetActiveLayout() (*models.Layout, error) {
	var layout models.Layout
	err := r.db.Where("is_active = ?", true).First(&layout).Error
	return &layout, err
}

// Count returns the total number of layouts
func (r *LayoutRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Layout{}).Count(&count).Error
	return count, err
}

// FindByComponentType finds layouts containing specific component types using JSONB queries
func (r *LayoutRepository) FindByComponentType(componentType string) ([]models.Layout, error) {
	var layouts []models.Layout

	// Query using JSONB path operations to find layouts containing the component type
	query := `schema -> 'layouts' -> 'lg' @> ?`
	componentQuery := fmt.Sprintf(`[{"componentType": "%s"}]`, componentType)

	err := r.db.Where(query, componentQuery).Find(&layouts).Error
	return layouts, err
}

// FindByWidgetID finds layouts containing a specific widget ID
func (r *LayoutRepository) FindByWidgetID(widgetID string) ([]models.Layout, error) {
	var layouts []models.Layout

	// Query using JSONB path operations to find layouts containing the widget ID
	query := `schema -> 'layouts' -> 'lg' @> ?`
	widgetQuery := fmt.Sprintf(`[{"i": "%s"}]`, widgetID)

	err := r.db.Where(query, widgetQuery).Find(&layouts).Error
	return layouts, err
}

// Exists checks if a layout with the given layout_id exists
func (r *LayoutRepository) Exists(layoutID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Layout{}).Where("layout_id = ?", layoutID).Count(&count).Error
	return count > 0, err
}