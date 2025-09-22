package tests

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/layout-manager/api/pkg/config"
	"github.com/layout-manager/api/pkg/db"
	"github.com/layout-manager/api/pkg/mcp"
	"gorm.io/gorm"
)

// TestContext holds all test dependencies
type TestContext struct {
	DB               *gorm.DB
	LayoutRepo       *db.LayoutRepository
	IntegrationBridge *mcp.IntegrationBridge
	MCPServer        interface{} // MCP server instance
	Config           *config.Config
}

// SetupTestEnvironment initializes a fresh test environment
func SetupTestEnvironment(t *testing.T) *TestContext {
	// Load test environment variables
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = "../.env.test"
	}

	if err := godotenv.Load(envFile); err != nil {
		t.Logf("Warning: env file not found at %s: %v", envFile, err)
	}

	// Load test configuration
	cfg := config.Load()

	// Ensure we're using test database
	if cfg.Database.Name != "layout_manager_test" {
		t.Fatal("Test must use layout_manager_test database")
	}

	// Wait for test database to be ready
	waitForDatabase(t, cfg.Database)

	// Initialize database connection
	database, err := db.InitDB(cfg.Database)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Initialize repository
	layoutRepo := db.NewLayoutRepository(database)

	// Create integration bridge
	integrationBridge := mcp.NewIntegrationBridge(layoutRepo)

	// Initialize MCP server
	mcpServer, err := mcp.NewMCPServer(layoutRepo, integrationBridge)
	if err != nil {
		t.Fatalf("Failed to initialize MCP server: %v", err)
	}

	log.Printf("[TEST] Test environment initialized successfully")
	log.Printf("[TEST] Database: %s@%s:%s/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	return &TestContext{
		DB:                database,
		LayoutRepo:        layoutRepo,
		IntegrationBridge: integrationBridge,
		MCPServer:         mcpServer,
		Config:            cfg,
	}
}

// CleanupTestEnvironment cleans up test resources
func (ctx *TestContext) CleanupTestEnvironment(t *testing.T) {
	if ctx.DB != nil {
		// Clean up all test data
		ctx.DB.Exec("TRUNCATE TABLE layouts RESTART IDENTITY CASCADE")

		// Close database connection
		if sqlDB, err := ctx.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}
	log.Printf("[TEST] Test environment cleaned up")
}

// waitForDatabase waits for the test database to be ready
func waitForDatabase(t *testing.T, dbConfig config.DatabaseConfig) {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		database, err := db.InitDB(dbConfig)
		if err == nil {
			if sqlDB, err := database.DB(); err == nil {
				sqlDB.Close()
				log.Printf("[TEST] Test database is ready")
				return
			}
		}

		if i == maxRetries-1 {
			t.Fatalf("Test database not ready after %d retries", maxRetries)
		}

		log.Printf("[TEST] Waiting for test database... (attempt %d/%d)", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}
}

// TestMain handles setup and teardown for all tests
func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("[TEST] Starting test suite...")

	// Run all tests
	code := m.Run()

	log.Println("[TEST] Test suite completed")
	os.Exit(code)
}