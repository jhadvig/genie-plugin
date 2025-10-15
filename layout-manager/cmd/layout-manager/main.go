package main

import (
	"log"
	"net/http"

	"github.com/mark3labs/mcp-go/server"

	"github.com/layout-manager/api/pkg/config"
	"github.com/layout-manager/api/pkg/db"
	"github.com/layout-manager/api/pkg/mcp"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set headers to allow all origins, methods, and headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Expose-Headers", "*")
		// Handle preflight requests (OPTIONS method)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass control to the next handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Println("Layout Manager MCP Server starting...")

	// Load configuration from .env file and environment variables
	cfg := config.Load()
	log.Printf("Running in %s environment", cfg.Environment)
	log.Printf("MCP Server will start on %s:%s", cfg.MCP.Host, cfg.MCP.Port)
	log.Printf("Database connection: %s@%s:%s/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	// Initialize database connection
	database, err := db.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close(database)

	log.Println("Database connected and migrated successfully")

	// Initialize repository
	layoutRepo := db.NewLayoutRepository(database)

	// Create integration bridge
	integrationBridge := mcp.NewIntegrationBridge(layoutRepo)

	// Initialize MCP server
	mcpServer, err := mcp.NewMCPServer(layoutRepo, integrationBridge)
	if err != nil {
		log.Fatalf("Failed to initialize MCP server: %v", err)
	}
	log.Println("MCP server initialized with 7 natural language tools")

	// Start MCP server
	mcpAddr := cfg.MCP.Host + ":" + cfg.MCP.Port
	log.Printf("Starting MCP HTTP server on %s", mcpAddr)

	// Create MCP HTTP server
	mcpMux := http.NewServeMux()

	corsHandler := corsMiddleware(mcpMux)
	mcpHTTPServer := &http.Server{
		Addr:    mcpAddr,
		Handler: corsHandler,
	}

	// Create streamable HTTP server from MCP server
	streamableHTTPServer := server.NewStreamableHTTPServer(mcpServer,
		server.WithStreamableHTTPServer(mcpHTTPServer),
	)

	// Mount MCP server on both /mcp and / endpoints
	mcpMux.Handle("/mcp", streamableHTTPServer)
	mcpMux.Handle("/", streamableHTTPServer)

	// Add health check for MCP server
	mcpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("MCP Server OK"))
	})

	if err := mcpHTTPServer.ListenAndServe(); err != nil {
		log.Fatalf("MCP server failed to start: %v", err)
	}
}
