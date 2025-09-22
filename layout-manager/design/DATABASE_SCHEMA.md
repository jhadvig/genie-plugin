# Database Schema Design

## PostgreSQL + GORM Schema

### Single Table: `layouts`

```sql
CREATE TABLE layouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    layout_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    description TEXT,
    schema JSONB NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for fast JSON queries
CREATE INDEX idx_layouts_schema ON layouts USING GIN (schema);
CREATE INDEX idx_layouts_layout_id ON layouts (layout_id);
CREATE INDEX idx_layouts_is_active ON layouts (is_active);
```

### GORM Model with Validation

```go
package models

import (
    "encoding/json"
    "fmt"
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
    "gorm.io/datatypes"
)

type Layout struct {
    ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
    LayoutID    string         `gorm:"uniqueIndex;not null" json:"layout_id" validate:"required,min=1,max=255"`
    Name        string         `json:"name" validate:"max=255"`
    Description string         `json:"description" validate:"max=1000"`
    Schema      datatypes.JSON `gorm:"type:jsonb;not null" json:"schema" validate:"required"`
    IsActive    bool           `gorm:"default:false;index" json:"is_active"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}

func (Layout) TableName() string {
    return "layouts"
}

// BeforeSave hook to validate JSON schema structure and handle active status
func (l *Layout) BeforeSave(tx *gorm.DB) error {
    // Validate that Schema can be unmarshaled into LayoutSchema
    var layoutSchema LayoutSchema
    if err := json.Unmarshal(l.Schema, &layoutSchema); err != nil {
        return fmt.Errorf("invalid schema format: %w", err)
    }

    // Validate schema structure
    if err := layoutSchema.Validate(); err != nil {
        return fmt.Errorf("schema validation failed: %w", err)
    }

    // Handle single active dashboard constraint
    if l.IsActive {
        // Deactivate all other layouts before saving this one as active
        err := tx.Model(&Layout{}).Where("id != ? AND is_active = ?", l.ID, true).Update("is_active", false).Error
        if err != nil {
            return fmt.Errorf("failed to deactivate other layouts: %w", err)
        }
    }

    return nil
}

// GetLayoutSchema unmarshals the JSON schema into a structured type
func (l *Layout) GetLayoutSchema() (*LayoutSchema, error) {
    var schema LayoutSchema
    if err := json.Unmarshal(l.Schema, &schema); err != nil {
        return nil, err
    }
    return &schema, nil
}

// SetLayoutSchema marshals a LayoutSchema into the JSON field
func (l *Layout) SetLayoutSchema(schema *LayoutSchema) error {
    if err := schema.Validate(); err != nil {
        return err
    }

    data, err := json.Marshal(schema)
    if err != nil {
        return err
    }

    l.Schema = datatypes.JSON(data)
    return nil
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
            MaxItems:          20,
            DefaultItemSize:   ItemSize{W: 4, H: 3},
            Margin:           [2]int{10, 10},
            ContainerPadding: [2]int{10, 10},
        },
    }
}
```

## Enhanced Layout Schema Structure

### JSON Schema stored in `schema` field:

```json
{
  "breakpoints": {
    "lg": 1200,
    "md": 996,
    "sm": 768,
    "xs": 480,
    "xxs": 0
  },
  "cols": {
    "lg": 12,
    "md": 10,
    "sm": 6,
    "xs": 4,
    "xxs": 2
  },
  "layouts": {
    "lg": [
      {
        "i": "widget-1",
        "x": 0,
        "y": 0,
        "w": 6,
        "h": 4,
        "minW": 2,
        "maxW": 8,
        "minH": 2,
        "maxH": 6,
        "static": false,
        "componentType": "chart",
        "props": {
          "title": "Sales Chart",
          "chartType": "line",
          "dataSource": "/api/sales",
          "refreshInterval": 30000
        },
        "constraints": {
          "minW": 2,
          "maxW": 8,
          "minH": 2,
          "maxH": 6,
          "aspectRatio": {
            "min": 0.5,
            "max": 2.0
          }
        }
      }
    ],
    "md": [...],
    "sm": [...],
    "xs": [...],
    "xxs": [...]
  },
  "globalConstraints": {
    "maxItems": 20,
    "defaultItemSize": { "w": 4, "h": 3 },
    "margin": [10, 10],
    "containerPadding": [10, 10]
  }
}
```

## Go Data Structures

```go
package models

// Enhanced layout item with component information and validation
type LayoutItem struct {
    // Standard react-grid-layout fields
    I             string  `json:"i" validate:"required"`
    X             int     `json:"x" validate:"min=0"`
    Y             int     `json:"y" validate:"min=0"`
    W             int     `json:"w" validate:"min=1"`
    H             int     `json:"h" validate:"min=1"`
    Static        bool    `json:"static,omitempty"`
    MinW          *int    `json:"minW,omitempty" validate:"omitempty,min=1"`
    MaxW          *int    `json:"maxW,omitempty" validate:"omitempty,min=1"`
    MinH          *int    `json:"minH,omitempty" validate:"omitempty,min=1"`
    MaxH          *int    `json:"maxH,omitempty" validate:"omitempty,min=1"`
    IsDraggable   *bool   `json:"isDraggable,omitempty"`
    IsResizable   *bool   `json:"isResizable,omitempty"`

    // Enhanced fields for component management
    ComponentType string                 `json:"componentType" validate:"required"`   // e.g., "chart", "table", "metric", "text"
    Props         map[string]interface{} `json:"props"`                               // Component-specific properties
    Constraints   *ItemConstraints       `json:"constraints,omitempty"`               // Size and proportion constraints
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

// Constraints for individual layout items
type ItemConstraints struct {
    MinW        int              `json:"minW"`
    MaxW        int              `json:"maxW"`
    MinH        int              `json:"minH"`
    MaxH        int              `json:"maxH"`
    AspectRatio *AspectRatio     `json:"aspectRatio,omitempty"`
}

// Aspect ratio constraints
type AspectRatio struct {
    Min float64 `json:"min"`  // Minimum width/height ratio
    Max float64 `json:"max"`  // Maximum width/height ratio
}

// Responsive breakpoint configuration
type Breakpoints map[string]int

// Column configuration per breakpoint
type Cols map[string]int

// Layouts for each breakpoint
type Layouts map[string][]LayoutItem

// Global layout constraints
type GlobalConstraints struct {
    MaxItems          int       `json:"maxItems"`
    DefaultItemSize   ItemSize  `json:"defaultItemSize"`
    Margin           [2]int     `json:"margin"`
    ContainerPadding [2]int     `json:"containerPadding"`
}

type ItemSize struct {
    W int `json:"w"`
    H int `json:"h"`
}

// Complete layout schema structure with validation
type LayoutSchema struct {
    Breakpoints       Breakpoints       `json:"breakpoints" validate:"required"`
    Cols             Cols              `json:"cols" validate:"required"`
    Layouts          Layouts           `json:"layouts" validate:"required"`
    GlobalConstraints GlobalConstraints `json:"globalConstraints" validate:"required"`
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

    for breakpoint, items := range ls.Layouts {
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
```

## Component Types Registry

Predefined component types with their constraints:

```go
package models

type ComponentDefinition struct {
    Type         string            `json:"type"`
    Name         string            `json:"name"`
    Description  string            `json:"description"`
    DefaultSize  ItemSize          `json:"defaultSize"`
    Constraints  ItemConstraints   `json:"constraints"`
    PropSchema   map[string]interface{} `json:"propSchema"` // JSON Schema for props validation
}

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
    },
}
```

## Database Operations with GORM

```go
package db

import (
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
    "github.com/your-org/layout-manager/pkg/models"
)

type LayoutRepository struct {
    db *gorm.DB
}

func NewLayoutRepository(db *gorm.DB) *LayoutRepository {
    return &LayoutRepository{db: db}
}

// Create or update layout
func (r *LayoutRepository) SaveLayout(layout *models.Layout) error {
    return r.db.Save(layout).Error
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

// SetActiveLayout sets a specific layout as active and deactivates others
func (r *LayoutRepository) SetActiveLayout(layoutID string) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // Deactivate all layouts
        if err := tx.Model(&models.Layout{}).Update("is_active", false).Error; err != nil {
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

// Get layout by ID
func (r *LayoutRepository) GetLayout(layoutID string) (*models.Layout, error) {
    var layout models.Layout
    err := r.db.Where("layout_id = ?", layoutID).First(&layout).Error
    return &layout, err
}

// List all layouts
func (r *LayoutRepository) ListLayouts() ([]models.Layout, error) {
    var layouts []models.Layout
    err := r.db.Order("is_active DESC, created_at DESC").Find(&layouts).Error
    return layouts, err
}

// Delete layout
func (r *LayoutRepository) DeleteLayout(layoutID string) error {
    return r.db.Where("layout_id = ?", layoutID).Delete(&models.Layout{}).Error
}

// Query layouts by component type
func (r *LayoutRepository) FindLayoutsByComponentType(componentType string) ([]models.Layout, error) {
    var layouts []models.Layout
    err := r.db.Where("schema -> 'layouts' -> 'lg' @> ?",
        fmt.Sprintf(`[{"componentType": "%s"}]`, componentType)).Find(&layouts).Error
    return layouts, err
}
```

## AutoMigrate Setup for POC

```go
package db

import (
    "fmt"
    "log"
    "os"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "github.com/your-org/layout-manager/pkg/models"
)

// InitDB initializes database connection and runs AutoMigrate
func InitDB() (*gorm.DB, error) {
    // Build connection string from environment variables
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
        getEnv("DB_HOST", "localhost"),
        getEnv("DB_USER", "postgres"),
        getEnv("DB_PASSWORD", "password"),
        getEnv("DB_NAME", "layout_manager"),
        getEnv("DB_PORT", "5432"),
    )

    // Configure GORM with detailed logging for POC
    config := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info), // Verbose logging for development
    }

    // Connect to database
    db, err := gorm.Open(postgres.Open(dsn), config)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Run AutoMigrate for all models
    if err := AutoMigrateAll(db); err != nil {
        return nil, fmt.Errorf("failed to migrate database: %w", err)
    }

    log.Println("Database connected and migrated successfully")
    return db, nil
}

// AutoMigrateAll runs GORM AutoMigrate for all models
func AutoMigrateAll(db *gorm.DB) error {
    // AutoMigrate will create the table, missing columns, missing indexes
    // It WON'T delete unused columns to protect your data
    err := db.AutoMigrate(
        &models.Layout{},
        // Add more models here as we expand
    )

    if err != nil {
        return fmt.Errorf("auto migration failed: %w", err)
    }

    // Create additional indexes that AutoMigrate might miss
    if err := createAdditionalIndexes(db); err != nil {
        return fmt.Errorf("failed to create additional indexes: %w", err)
    }

    log.Println("AutoMigrate completed successfully")
    return nil
}

// createAdditionalIndexes creates indexes that AutoMigrate doesn't handle
func createAdditionalIndexes(db *gorm.DB) error {
    // GIN index for JSONB queries (if not created by AutoMigrate)
    if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_layouts_schema_gin ON layouts USING GIN (schema)").Error; err != nil {
        return fmt.Errorf("failed to create GIN index: %w", err)
    }

    return nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}

// For POC: Quick database reset function (USE WITH CAUTION)
func ResetDatabase(db *gorm.DB) error {
    log.Println("WARNING: Resetting database - all data will be lost!")

    // Drop all tables
    if err := db.Migrator().DropTable(&models.Layout{}); err != nil {
        return fmt.Errorf("failed to drop tables: %w", err)
    }

    // Re-run migration
    return AutoMigrateAll(db)
}
```

## Simplified Database Connection for Main

```go
package main

import (
    "log"
    "github.com/your-org/layout-manager/pkg/db"
)

func main() {
    // Initialize database
    database, err := db.InitDB()
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    // Get underlying SQL DB for connection management
    sqlDB, err := database.DB()
    if err != nil {
        log.Fatalf("Failed to get underlying SQL DB: %v", err)
    }
    defer sqlDB.Close()

    // Configure connection pool for production later
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)

    log.Println("Layout Manager API starting...")
    // Continue with server setup...
}
```

## Environment Variables for POC

Create a `.env` file:
```env
# Database Configuration
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=layout_manager
DB_PORT=5432

# Server Configuration
SERVER_PORT=8080
```

## Docker Compose Integration

The AutoMigrate will run automatically when the Go server starts, so we just need to ensure the database is ready:

```yaml
# docker-compose.yml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: layout_manager
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  layout-manager-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      DB_USER: postgres
      DB_PASSWORD: password
      DB_NAME: layout_manager
      DB_PORT: 5432
    depends_on:
      - postgres
    restart: unless-stopped

volumes:
  postgres_data:
```

This schema provides:
1. **Single table simplicity** with UUID primary keys
2. **JSONB storage** for flexible layout schemas with indexing
3. **Enhanced layout items** with componentType and props
4. **Constraint system** for size and proportion limits
5. **Component registry** for predefined widget types
6. **GORM integration** for easy database operations
7. **JSON querying capabilities** for complex searches