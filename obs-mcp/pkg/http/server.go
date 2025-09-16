package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

const (
	mcpEndpoint    = "/mcp"
	healthEndpoint = "/health"
)

// loggingMiddleware logs incoming HTTP requests with debug information
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] Incoming request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		log.Printf("[DEBUG] Request headers: %v", r.Header)
		if r.ContentLength > 0 {
			log.Printf("[DEBUG] Content-Length: %d", r.ContentLength)
		}
		next.ServeHTTP(w, r)
	})
}

// Serve starts an HTTP server with the MCP server mounted
func Serve(ctx context.Context, mcpServer *server.MCPServer, listenAddr string) error {
	mux := http.NewServeMux()

	// Create streamable HTTP server from MCP server with logging middleware
	httpServer := &http.Server{
		Addr:    listenAddr,
		Handler: loggingMiddleware(mux),
	}

	// Mount the MCP server on the /mcp endpoint
	streamableHTTPServer := server.NewStreamableHTTPServer(mcpServer,
		server.WithStreamableHTTPServer(httpServer),
		server.WithStateLess(true),
	)
	mux.Handle(mcpEndpoint, streamableHTTPServer)

	// It seems Lightspeed-stack needs the server on / as well
	mux.Handle("/", streamableHTTPServer)

	// Add health check endpoint
	mux.HandleFunc(healthEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("HTTP server starting on %s with MCP endpoint at %s", listenAddr, mcpEndpoint)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, initiating graceful shutdown", sig)
		cancel()
	case <-ctx.Done():
		log.Printf("Context cancelled, initiating graceful shutdown")
	case err := <-serverErr:
		log.Printf("HTTP server error: %v", err)
		return err
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	log.Printf("Shutting down HTTP server gracefully...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
		return err
	}

	log.Printf("HTTP server shutdown complete")
	return nil
}
