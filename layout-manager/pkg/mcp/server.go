package mcp

import (
	"github.com/layout-manager/api/pkg/db"
	"github.com/mark3labs/mcp-go/server"
)

// NewMCPServer creates a new MCP server for layout management
func NewMCPServer(layoutRepo *db.LayoutRepository, bridge *IntegrationBridge) (*server.MCPServer, error) {
	mcpServer := server.NewMCPServer(
		"layout-manager-mcp",
		"1.0.0",
		server.WithLogging(),
		server.WithToolCapabilities(true),
	)

	if err := SetupTools(mcpServer, layoutRepo, bridge); err != nil {
		return nil, err
	}

	return mcpServer, nil
}

// SetupTools configures all MCP tools with their handlers
func SetupTools(mcpServer *server.MCPServer, layoutRepo *db.LayoutRepository, bridge *IntegrationBridge) error {
	// Create tool definitions
	findWidgetsTool := CreateFindWidgetsTool()
	manipulateWidgetTool := CreateManipulateWidgetTool()
	addWidgetTool := CreateAddWidgetTool()
	batchOperationsTool := CreateBatchWidgetOperationsTool()
	analyzeLayoutTool := CreateAnalyzeLayoutTool()
	configureWidgetTool := CreateConfigureWidgetTool()
	createDashboardTool := CreateDashboardTool()
    getActiveDashboardTool := CreateGetActiveDashboardTool()
    setDashboardMetadataTool := CreateSetDashboardMetadataTool()
    listDashboardsTool := CreateListDashboardsTool()
    getDashboardTool := CreateGetDashboardTool()

	// Create handlers with repository access
	findWidgetsHandler := FindWidgetsHandler(layoutRepo, bridge)
	manipulateWidgetHandler := ManipulateWidgetHandler(layoutRepo)
	addWidgetHandler := AddWidgetHandler(layoutRepo)
	batchOperationsHandler := BatchWidgetOperationsHandler(layoutRepo)
	analyzeLayoutHandler := AnalyzeLayoutHandler(layoutRepo)
	configureWidgetHandler := ConfigureWidgetHandler(layoutRepo)
	createDashboardHandler := CreateDashboardHandler(layoutRepo)
    getActiveDashboardHandler := GetActiveDashboardHandler(layoutRepo)
    setDashboardMetadataHandler := SetDashboardMetadataHandler(layoutRepo)
    listDashboardsHandler := ListDashboardsHandler(layoutRepo)
    getDashboardHandler := GetDashboardHandler(layoutRepo)

	// Add tools to server
	mcpServer.AddTool(findWidgetsTool, findWidgetsHandler)
	mcpServer.AddTool(manipulateWidgetTool, manipulateWidgetHandler)
	mcpServer.AddTool(addWidgetTool, addWidgetHandler)
	mcpServer.AddTool(batchOperationsTool, batchOperationsHandler)
	mcpServer.AddTool(analyzeLayoutTool, analyzeLayoutHandler)
	mcpServer.AddTool(configureWidgetTool, configureWidgetHandler)
	mcpServer.AddTool(createDashboardTool, createDashboardHandler)
	mcpServer.AddTool(getActiveDashboardTool, getActiveDashboardHandler)
    mcpServer.AddTool(setDashboardMetadataTool, setDashboardMetadataHandler)
    mcpServer.AddTool(listDashboardsTool, listDashboardsHandler)
    mcpServer.AddTool(getDashboardTool, getDashboardHandler)

	return nil
}