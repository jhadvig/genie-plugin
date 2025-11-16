package pkg

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/inecas/obs-mcp/version"
	"github.com/mark3labs/mcp-go/server"
)

// ObsMCPOptions contains configuration options for the MCP server
type ObsMCPOptions struct {
	AuthMode       AuthMode
	PromURL        string
	Insecure       bool
	EndpointConfig *EndpointConfig
}

const (
	mcpEndpoint    = "/mcp"
	healthEndpoint = "/health"
)

func NewMCPServer(opts ObsMCPOptions) (*server.MCPServer, error) {
	if opts.EndpointConfig == nil {
		opts.EndpointConfig = DefaultEndpointConfig()
	}

	serverOpts := []server.ServerOption{
		server.WithLogging(),
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
	}

	mcpServer := server.NewMCPServer(
		"obs-mcp",
		version.GetVersion(),
		serverOpts...,
	)
	if err := SetupUniversalTool(mcpServer, opts); err != nil {
		return nil, err
	}
	if err := SetupResources(mcpServer); err != nil {
		return nil, err
	}

	return mcpServer, nil
}

// SetupUniversalTool sets up the universal Prometheus query tool in the MCP server, with guardrails
func SetupUniversalTool(mcpServer *server.MCPServer, opts ObsMCPOptions) error {
	guardrails := NewEndpointGuardrails()
	guardrails.RegisterHandler(PathQueryRange, func(ctx context.Context, endpoint PrometheusEndpoint, params map[string]interface{}) error {
		fmt.Println("custom handler for /api/v1/query_range invoked")
		fmt.Println("Endpoint Path:", endpoint.Path)
		fmt.Println("Parameters:", params)
		fmt.Println("Endpoint Config:", opts.EndpointConfig)
		fmt.Println("this runs in addition to the default PromQL validation guardrails")

		return nil
	})
	tool := CreateUniversalPrometheusQueryTool(opts.EndpointConfig)
	handler := DynamicPrometheusHandler(opts, guardrails)
	mcpServer.AddTool(tool, handler)

	return nil
}

func Serve(ctx context.Context, mcpServer *server.MCPServer, listenAddr string) error {
	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:    listenAddr,
		Handler: loggingMiddleware(mux),
	}
	streamableHTTPServer := server.NewStreamableHTTPServer(mcpServer,
		server.WithStreamableHTTPServer(httpServer),
		server.WithStateLess(true), // TODO: This affects session handling; consider implications
		server.WithHTTPContextFunc(authFromRequest),
	)
	mux.Handle(mcpEndpoint, streamableHTTPServer)
	mux.Handle("/", streamableHTTPServer)
	mux.HandleFunc(healthEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
		if err != nil {
			log.Printf("Health check response write error: %v", err)
			return
		}
	})

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

func authFromRequest(ctx context.Context, r *http.Request) context.Context {
	authHeaderValue := r.Header.Get(AuthHeaderKey)
	token, found := strings.CutPrefix(authHeaderValue, "Bearer ")
	if !found {
		return ctx
	}

	//nolint:staticcheck //SA1029
	return context.WithValue(ctx, AuthHeaderKey, token)
}

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
