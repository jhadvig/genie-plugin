package mcp_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/layout-manager/api/pkg/mcp"
	"github.com/layout-manager/api/pkg/models"
	"github.com/layout-manager/api/pkg/testutil"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestFindWidgetsTool(t *testing.T) {
	// Setup test environment
	ctx := testutil.SetupTestEnvironment(t)
	defer ctx.CleanupTestEnvironment(t)

	// Create a test layout with widgets
	testLayout := createTestLayoutWithWidgets(t, ctx)

	t.Run("Find Chart Widget", func(t *testing.T) {
		// Create MCP request to find chart widgets
		arguments := map[string]interface{}{
			"layout_id":      testLayout.LayoutID,
			"description":    "chart widget",
			"breakpoint":     "lg",
			"component_type": "chart",
		}
		request := mcpgo.CallToolRequest{
			Params: mcpgo.CallToolParams{
				Name:      "find_widgets",
				Arguments: arguments,
			},
		}

		// Create handler
		handler := mcp.FindWidgetsHandler(ctx.LayoutRepo, ctx.IntegrationBridge)

		// Execute handler
		result, err := handler(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Parse response - result should be text content
		textContent, ok := mcpgo.AsTextContent(result.Content[0])
		require.True(t, ok, "Expected text content")

		var response mcp.MCPResponse
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err)

		// Verify results
		assert.True(t, response.Success)
		assert.Equal(t, "find_widgets", response.Operation)
		assert.Equal(t, 1, response.TotalFound)
		assert.Len(t, response.Widgets, 1)

		widget := response.Widgets[0]
		assert.Equal(t, "sales-chart", widget.ID)
		assert.Equal(t, "chart", widget.ComponentType)
		assert.Contains(t, response.Message, "Found 1 widget(s)")
	})

	t.Run("Find Table Widget by Title", func(t *testing.T) {
		arguments := map[string]interface{}{
			"layout_id":   testLayout.LayoutID,
			"description": "customer table",
			"breakpoint":  "lg",
		}
		request := mcpgo.CallToolRequest{
			Params: mcpgo.CallToolParams{
				Name:      "find_widgets",
				Arguments: arguments,
			},
		}

		handler := mcp.FindWidgetsHandler(ctx.LayoutRepo, ctx.IntegrationBridge)
		result, err := handler(context.Background(), request)
		require.NoError(t, err)

		textContent, ok := mcpgo.AsTextContent(result.Content[0])
		require.True(t, ok, "Expected text content")

		var response mcp.MCPResponse
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, 1, response.TotalFound)

		widget := response.Widgets[0]
		assert.Equal(t, "customer-table", widget.ID)
		assert.Equal(t, "table", widget.ComponentType)
	})

	t.Run("Find No Matching Widgets", func(t *testing.T) {
		arguments := map[string]interface{}{
			"layout_id":   testLayout.LayoutID,
			"description": "nonexistent widget",
			"breakpoint":  "lg",
		}
		request := mcpgo.CallToolRequest{
			Params: mcpgo.CallToolParams{
				Name:      "find_widgets",
				Arguments: arguments,
			},
		}

		handler := mcp.FindWidgetsHandler(ctx.LayoutRepo, ctx.IntegrationBridge)
		result, err := handler(context.Background(), request)
		require.NoError(t, err)

		textContent, ok := mcpgo.AsTextContent(result.Content[0])
		require.True(t, ok, "Expected text content")

		var response mcp.MCPResponse
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, 0, response.TotalFound)
		assert.Len(t, response.Widgets, 0)
		assert.Contains(t, response.Message, "Found 0 widget(s)")
	})

	t.Run("Invalid Layout ID", func(t *testing.T) {
		arguments := map[string]interface{}{
			"layout_id":   "nonexistent-layout",
			"description": "any widget",
			"breakpoint":  "lg",
		}
		request := mcpgo.CallToolRequest{
			Params: mcpgo.CallToolParams{
				Name:      "find_widgets",
				Arguments: arguments,
			},
		}

		handler := mcp.FindWidgetsHandler(ctx.LayoutRepo, ctx.IntegrationBridge)
		result, err := handler(context.Background(), request)
		require.NoError(t, err)

		// Should return error result
		textContent, ok := mcpgo.AsTextContent(result.Content[0])
		require.True(t, ok, "Expected text content")
		assert.Contains(t, textContent.Text, "Layout not found")
	})
}

// Helper function to create a test layout with sample widgets
func createTestLayoutWithWidgets(t *testing.T, ctx *testutil.TestContext) *models.Layout {
	layout := &models.Layout{
		LayoutID:    "test-layout-mcp",
		Name:        "MCP Test Layout",
		Description: "Layout for testing MCP operations",
		Schema: datatypes.JSONType[models.LayoutSchema]{
			Data: models.LayoutSchema{
				Breakpoints: models.Breakpoints{
					"lg": 1200,
				},
				Cols: models.Cols{
					"lg": 12,
				},
				Layouts: models.Layouts{
					"lg": []models.LayoutItem{
						{
							I:             "sales-chart",
							X:             0,
							Y:             0,
							W:             6,
							H:             4,
							ComponentType: "chart",
							Props: map[string]interface{}{
								"title":      "Sales Chart",
								"chartType":  "line",
								"dataSource": "sales",
							},
						},
						{
							I:             "customer-table",
							X:             6,
							Y:             0,
							W:             6,
							H:             4,
							ComponentType: "table",
							Props: map[string]interface{}{
								"title":      "Customer Table",
								"dataSource": "customers",
							},
						},
						{
							I:             "revenue-metric",
							X:             0,
							Y:             4,
							W:             3,
							H:             2,
							ComponentType: "metric",
							Props: map[string]interface{}{
								"title":  "Revenue",
								"value":  125000,
								"format": "currency",
							},
						},
					},
				},
				GlobalConstraints: models.GlobalConstraints{
					MaxItems:         20,
					DefaultItemSize:  models.ItemSize{W: 4, H: 3},
					Margin:           [2]int{10, 10},
					ContainerPadding: [2]int{10, 10},
				},
			},
		},
	}

	err := ctx.LayoutRepo.Create(layout)
	require.NoError(t, err)
	return layout
}
