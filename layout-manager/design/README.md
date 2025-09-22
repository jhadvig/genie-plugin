# Layout Manager - MCP Server Design Documentation

This directory contains design documents for the Layout Manager MCP server project.

## Contents

### Core Design Documents
- **[MCP_TOOLS_DESIGN.md](./MCP_TOOLS_DESIGN.md)** - MCP tools specification and natural language interface
- **[DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md)** - PostgreSQL schema design with GORM models and active dashboard support
- **[REACT_GRID_LAYOUT_DOCS.md](./REACT_GRID_LAYOUT_DOCS.md)** - React Grid Layout API reference and data structures
- **[GENERIC_WIDGET_CONFIG_DESIGN.md](./GENERIC_WIDGET_CONFIG_DESIGN.md)** - Generic widget configuration system

## Architecture Overview

### System Components
1. **Go MCP Server** - Natural language interface using mark3labs/mcp-go
2. **PostgreSQL Database** - JSONB storage with active dashboard management
3. **Active Dashboard System** - Single active dashboard workflow
4. **React Grid Layout Compatibility** - Full schema support
5. **Direct Database Integration** - MCP tools directly access database through repository layer

### Key Technologies
- **Backend**: Go 1.21+, GORM, PostgreSQL, mark3labs/mcp-go v0.39.1
- **Database**: PostgreSQL with JSONB, UUID primary keys, GIN indexing, `is_active` constraints
- **MCP**: Natural language tools for dashboard management
- **Development**: Docker Compose, GORM AutoMigrate

### Data Flow
1. **MCP Client** sends natural language commands to MCP server
2. **MCP Tools** interpret commands and operate on active dashboard
3. **Repository Layer** provides database operations for layouts and widgets
4. **Database** stores layout schema as JSONB with active dashboard tracking
5. **Response** provides structured feedback for UI synchronization

### Project Structure
```
layout-manager/
├── design/                    # Design documentation (this directory)
├── cmd/layout-manager/       # Main MCP server application
├── pkg/
│   ├── db/                   # Database layer with active dashboard support
│   ├── models/               # Data models and React Grid Layout schemas
│   ├── mcp/                  # MCP server, tools, and direct database integration
│   └── config/               # Configuration
├── docker-compose.yml        # Development environment
└── go.mod                   # Go dependencies
```

## MCP Tools Overview

### 1. `create_dashboard`
Create new empty dashboard and set as active
- **Parameters**: `name` (optional), `description` (optional)
- **Auto-generates**: Unique layout_id with timestamp

### 2. `get_active_dashboard`
Get current active dashboard schema and widget information
- **Parameters**: `breakpoint` (optional), `include_all_breakpoints` (optional)
- **Returns**: Complete layout schema with exact widget IDs and positions
- **Usage**: Essential first step for precise widget manipulation

### 3. `add_widget`
Add widgets to the active dashboard
- **Parameters**: `widget_description`, `component_type`, `props` (optional)
- **Supports**: All React Grid Layout widget types with custom properties
- **Returns**: Created widget with exact ID and position

### 4. `manipulate_widget`
Perform precise operations on widgets using exact widget IDs
- **Operations**: move, resize, remove
- **Workflow**: Use `get_active_dashboard` first, then operate with exact widget IDs
- **Parameters**: `widget_id` (exact ID), `operation`, coordinates/dimensions
- **Returns**: Updated widget information and change details

### 5. `find_widgets`
Find widgets in active dashboard by description
- **Search**: Natural language widget descriptions
- **Returns**: Matching widgets with positions and properties

### 6. `configure_widget`
Configure widget properties in active dashboard
- **Parameters**: `widget_selector`, `configuration_request`
- **Handles**: Complex property updates, filters, data sources

### 7. `analyze_layout`
Analyze active dashboard structure
- **Questions**: "What widgets are here?", "How many charts?"
- **Returns**: Dashboard insights and widget information

### 8. `batch_widget_operations`
Perform multiple widget operations in sequence
- **Parameters**: `commands_json` (array of commands), `atomic` (optional)
- **Supports**: Complex multi-widget operations

## Active Dashboard System

### Database Schema
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

CREATE INDEX idx_layouts_is_active ON layouts (is_active);
```

### Key Constraints
- Only one dashboard can be `is_active = true` at a time
- New dashboards automatically become active
- All MCP operations work on the active dashboard

### Layout Data Structure
```json
{
  "breakpoints": {"lg": 1200, "md": 996, "sm": 768, "xs": 480, "xxs": 0},
  "cols": {"lg": 12, "md": 10, "sm": 6, "xs": 4, "xxs": 2},
  "layouts": {
    "lg": [{
      "i": "widget-1",
      "x": 0, "y": 0, "w": 4, "h": 3,
      "componentType": "chart",
      "props": {"title": "Sales Chart", "chartType": "line"}
    }]
  },
  "globalConstraints": {
    "maxItems": 20,
    "defaultItemSize": {"w": 4, "h": 3}
  }
}
```

### Component Types
- **chart** - Data visualization widgets
- **table** - Tabular data display
- **metric** - Single KPI/metric display
- **text** - Static/dynamic text content
- **image** - Image widgets
- **iframe** - Embedded content

## Development Workflow

1. **Start Database** - `make db-up`
2. **Build MCP Server** - `make build`
3. **Run Development** - `make dev`
4. **Test MCP Tools** - Connect MCP client and test natural language commands

## Natural Language Examples

### Dashboard Creation
- "Create a new dashboard"
- "Create a sales dashboard"
- "Make a dashboard called System Overview"

### Widget Management
- "Add a chart showing revenue data"
- "Create a table with customer information"
- "Add a metric displaying total users"

### Widget Operations (Precise Workflow)
1. **Get Current State**: "Show me the current dashboard layout"
2. **Identify Target**: Look for exact widget ID (e.g., `widget-1234567890`)
3. **Precise Operation**: "Move widget-1234567890 to position x=6, y=0"
4. **Batch Operations**: "Resize widget-abc to 6x4, then move widget-def to x=0, y=3"

### Configuration
- "Change the chart type to bar"
- "Update the table filters to show only active users"
- "Set the metric to display percentage format"

This MCP server provides a streamlined natural language interface for dashboard management with automatic active dashboard tracking.