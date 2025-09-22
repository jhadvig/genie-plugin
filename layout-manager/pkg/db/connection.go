package db

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/layout-manager/api/pkg/config"
	"github.com/layout-manager/api/pkg/models"
)

// InitDB initializes database connection and runs AutoMigrate
func InitDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	// Build connection string from configuration
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Port,
	)

	// Configure GORM logging based on environment
	logLevel := logger.Error // Default to minimal logging
	if cfg.Host == "localhost" || cfg.Port == "5433" {
		// Development environment - enable verbose logging
		logLevel = logger.Info
	}

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run AutoMigrate for all models (will be implemented in next stage)
	if err := AutoMigrateAll(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connection initialized successfully")
	return db, nil
}

// Close properly closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AutoMigrateAll runs GORM AutoMigrate for all models
func AutoMigrateAll(db *gorm.DB) error {
	// Import models package
	// Note: We'll need to import this at the top when we import the models
	// For now, we'll create the import path assuming it exists

	// Run AutoMigrate for Layout model
	err := db.AutoMigrate(
		&models.Layout{},
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