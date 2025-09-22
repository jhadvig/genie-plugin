# Layout Manager MCP Server

A Model Context Protocol (MCP) server for precise dashboard layout management. This server provides natural language tools for creating, managing, and manipulating dashboard widgets with exact widget ID targeting and database persistence.

## Quick Start

```bash
# Start the development environment
make dev

# Or start components separately:
make db-up     # Start PostgreSQL database
make run       # Start the MCP server
```

## MCP Server

### Connection Details
- **Server Name**: `layout-manager-mcp`
- **Version**: `1.0.0`
- **Base URL**: `http://localhost:9081`
- **Health Check**: `GET /health`
- **MCP Endpoint**: `/mcp` or `/`

### Key Features
- **Active Dashboard System**: Single active dashboard workflow
- **Precise Widget Manipulation**: Exact widget ID targeting
- **Database Persistence**: All changes saved to PostgreSQL
- **Complete State Tracking**: Returns full widget information and change details

## MCP Tools

The server provides 8 natural language tools for precise dashboard management:

### 1. `create_dashboard`
Create new empty dashboard and set as active
```
"Create a new sales dashboard"
"Make a dashboard called System Overview"
```

### 2. `get_active_dashboard`
Get current dashboard state with exact widget IDs (essential first step)
```
"Show me the current dashboard layout"
"Get the active dashboard state"
```

### 3. `add_widget`
Add widgets to active dashboard with automatic positioning
```
"Add a chart showing revenue data"
"Create a metric widget for total users"
```

### 4. `manipulate_widget`
Precise widget operations using exact widget IDs
```
# Workflow: First get dashboard state, then use exact IDs
"Move widget-1727012345678 to position x=6, y=0"
"Resize widget-1727012345679 to 8x4"
"Remove widget-1727012345680"
```

### 5. `find_widgets`
Search widgets by natural language descriptions
```
"Find the revenue chart"
"Find all table widgets"
```

### 6. `configure_widget`
Update widget properties and settings
```
"Change the chart type to bar for the sales widget"
"Update the table filters to show only active users"
```

### 7. `analyze_layout`
Get insights about dashboard structure
```
"How many widgets are in this layout?"
"What widget types are present?"
```

### 8. `batch_widget_operations`
Execute multiple operations in sequence
```
"Resize widget-abc to 6x4, then move widget-def to x=0, y=3"
```

## Precise Widget Manipulation Workflow

### 1. Create Dashboard
```
User: "Create a sales dashboard"
→ MCP Tool: create_dashboard
→ Result: Empty dashboard created and activated
```

### 2. Add Widgets
```
User: "Add a chart showing revenue data"
→ MCP Tool: add_widget
→ Parameters: widget_description="revenue chart", component_type="chart"
→ Result: Widget created with ID "widget-1727012345678"
```

### 3. Get Current State (Essential)
```
User: "Show me the current dashboard"
→ MCP Tool: get_active_dashboard
→ Result: Complete layout with exact widget IDs and positions
```

### 4. Precise Manipulation
```
User: "Move the revenue chart to the top right"
→ LLM: Identifies widget-1727012345678 from dashboard state
→ MCP Tool: manipulate_widget
→ Parameters: widget_id="widget-1727012345678", operation="move", x="6", y="0"
→ Result: Widget moved with complete change tracking
```

### 5. Batch Operations
```
User: "Resize the chart to 8x4 and add a metrics widget"
→ Step 1: manipulate_widget (resize chart)
→ Step 2: add_widget (new metrics widget)
→ Result: Multiple widgets updated with full state tracking
```

## Configuration

### Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=layout_manager

# MCP Server
MCP_HOST=0.0.0.0
MCP_PORT=9081

# Environment
ENV=development
```

## Development

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL (via Docker)

### Available Commands

```bash
# Development
make dev          # Start full development environment (PostgreSQL + MCP server)
make run          # Run MCP server only
make build        # Build the binary

# Database
make db-up        # Start PostgreSQL
make db-down      # Stop database
make db-reset     # Reset database (destroys data)

# Code Quality
make lint         # Run linter
make fmt          # Format code
make vet          # Run go vet
```

### Project Structure

```
layout-manager/
├── cmd/layout-manager/     # MCP server main application
├── pkg/
│   ├── config/            # Configuration management
│   ├── db/                # Database layer with active dashboard support
│   ├── mcp/               # MCP server, tools, and database integration
│   └── models/            # Data models with GORM tags
├── design/                # Design documents (updated for MCP-only)
├── docker-compose.yml     # Development PostgreSQL database
└── Makefile              # Development commands
```

## Widget Types

The system supports these widget types:

- **chart** - Data visualizations (line, bar, pie charts)
- **table** - Tabular data display
- **metric** - Single value displays with formatting
- **text** - Rich text content
- **image** - Image displays
- **iframe** - Embedded external content

## Active Dashboard System

### Database Schema Features
- **Single Active Dashboard**: Only one dashboard can be active at a time
- **Automatic Activation**: New dashboards become active automatically
- **JSONB Storage**: Flexible React Grid Layout schema storage
- **Widget ID Generation**: Unique timestamped widget IDs
- **Complete State Tracking**: Full widget position and property management

### Database Constraints
```sql
-- Only one active dashboard constraint
CREATE UNIQUE INDEX idx_layouts_single_active
ON layouts (is_active) WHERE is_active = true;

-- JSONB indexing for performance
CREATE INDEX idx_layouts_schema_gin ON layouts USING GIN (schema);
```

## MCP Integration

This MCP server demonstrates:

1. **Precise Widget Manipulation**: Exact widget ID targeting eliminates ambiguity
2. **Direct Database Operations**: No REST API layer - direct PostgreSQL access
3. **Active Dashboard Workflow**: Simplified single-dashboard operations
4. **Complete State Management**: Full widget lifecycle with change tracking
5. **Natural Language Interface**: Complex operations via simple commands

### MCP Client Configuration

Add to your MCP client configuration:

```json
{
  "mcpServers": {
    "layout-manager": {
      "command": "curl",
      "args": ["-X", "POST", "http://localhost:9081/mcp"],
      "env": {}
    }
  }
}
```

The system provides a streamlined natural language interface for dashboard management with precise widget control and comprehensive database persistence.