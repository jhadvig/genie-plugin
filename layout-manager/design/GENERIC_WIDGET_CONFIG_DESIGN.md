# Generic Widget Configuration Design

## Problem Statement

Users need to modify widget-specific configurations through natural language commands like:
- "Change the filters to show only critically affected systems"
- "Set the chart type to bar chart"
- "Update the data source to use live data"
- "Change the table columns to show name, status, and last updated"

The challenge is that the server doesn't know:
1. What configuration options each widget type supports
2. The structure of widget-specific props
3. Domain-specific terminology and values
4. Current frontend implementations

## Solution Architecture

### Multi-Layered Approach

1. **Generic Intent Extraction** - Parse natural language to identify configuration changes
2. **Schema Discovery** - Optionally discover widget capabilities from frontend
3. **Intelligent Prop Mapping** - Map natural language to likely prop structures
4. **Validation & Feedback Loop** - Validate changes and provide feedback
5. **Extensible Plugin System** - Allow domain-specific handlers

## Core MCP Tool: `configure_widget`

```json
{
  "name": "configure_widget",
  "description": "Configure widget properties and settings using natural language descriptions. Handles complex configuration changes like filters, data sources, display options, and component-specific settings.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "layout_id": {
        "type": "string",
        "description": "The layout identifier"
      },
      "widget_selector": {
        "type": "string",
        "description": "Description of which widget to configure (e.g., 'sales table', 'main chart', 'system status widget')"
      },
      "configuration_request": {
        "type": "string",
        "description": "Natural language description of the configuration change (e.g., 'change filters to show only critical systems', 'set chart type to bar', 'update data source')"
      },
      "domain_context": {
        "type": "string",
        "description": "Optional domain context to help interpret the request (e.g., 'monitoring dashboard', 'sales analytics', 'system health')"
      },
      "current_props": {
        "type": "object",
        "description": "Optional: current widget props to help understand existing configuration",
        "additionalProperties": true
      },
      "suggested_props": {
        "type": "object",
        "description": "Optional: suggested prop changes from frontend if available",
        "additionalProperties": true
      },
      "breakpoint": {
        "type": "string",
        "default": "lg",
        "description": "Breakpoint to modify"
      }
    },
    "required": ["layout_id", "widget_selector", "configuration_request"]
  }
}
```

## Intent Classification System

### Configuration Categories

**Data & Sources:**
- Keywords: "data source", "API", "endpoint", "database", "query", "feed"
- Examples: "change data source to live API", "update query to filter by status"

**Filters & Criteria:**
- Keywords: "filter", "show only", "hide", "criteria", "condition", "where"
- Examples: "filter to show critical systems", "hide inactive users"

**Display & Visualization:**
- Keywords: "chart type", "display as", "show as", "format", "style", "theme"
- Examples: "change to bar chart", "display as table", "use dark theme"

**Columns & Fields:**
- Keywords: "columns", "fields", "show", "display", "include", "remove column"
- Examples: "show name and status columns", "add timestamp field"

**Time & Ranges:**
- Keywords: "time range", "period", "last", "since", "from", "to", "duration"
- Examples: "show last 24 hours", "set range to last month"

**Sorting & Grouping:**
- Keywords: "sort by", "order by", "group by", "arrange", "organize"
- Examples: "sort by severity", "group by department"

## Generic Prop Mapping Strategy

### 1. Common Prop Patterns

```go
// Standard prop structures we can intelligently map to
type CommonPropPatterns struct {
    DataSource    *DataSourceConfig    `json:"dataSource,omitempty"`
    Filters       []FilterConfig       `json:"filters,omitempty"`
    DisplayConfig *DisplayConfig       `json:"display,omitempty"`
    Columns       []ColumnConfig       `json:"columns,omitempty"`
    TimeRange     *TimeRangeConfig     `json:"timeRange,omitempty"`
    Sorting       *SortConfig          `json:"sorting,omitempty"`
}

type DataSourceConfig struct {
    URL      string            `json:"url,omitempty"`
    Type     string            `json:"type,omitempty"` // "api", "static", "database"
    Params   map[string]interface{} `json:"params,omitempty"`
    Headers  map[string]string `json:"headers,omitempty"`
}

type FilterConfig struct {
    Field    string      `json:"field"`
    Operator string      `json:"operator"` // "equals", "contains", "greater", "less"
    Value    interface{} `json:"value"`
    Label    string      `json:"label,omitempty"`
}

type DisplayConfig struct {
    Type       string                 `json:"type,omitempty"`    // "table", "chart", "list"
    ChartType  string                 `json:"chartType,omitempty"` // "bar", "line", "pie"
    Theme      string                 `json:"theme,omitempty"`
    Options    map[string]interface{} `json:"options,omitempty"`
}
```

### 2. Natural Language → Prop Mapping

```go
type ConfigurationParser struct {
    patterns map[string]PropMappingRule
}

type PropMappingRule struct {
    Keywords    []string                   `json:"keywords"`
    PropPath    string                     `json:"propPath"`
    ValueMapper func(string) interface{}   `json:"-"`
    Examples    []string                   `json:"examples"`
}

// Example mapping rules
var DefaultMappingRules = []PropMappingRule{
    {
        Keywords: []string{"filter", "show only", "where", "criteria"},
        PropPath: "filters",
        ValueMapper: parseFilterExpression,
        Examples: []string{
            "show only critical systems" → filters: [{field: "severity", operator: "equals", value: "critical"}]
            "filter by status active"    → filters: [{field: "status", operator: "equals", value: "active"}]
        },
    },
    {
        Keywords: []string{"chart type", "display as", "visualization"},
        PropPath: "display.chartType",
        ValueMapper: parseChartType,
        Examples: []string{
            "change to bar chart"   → display: {chartType: "bar"}
            "show as line graph"    → display: {chartType: "line"}
        },
    },
    {
        Keywords: []string{"data source", "API", "endpoint"},
        PropPath: "dataSource.url",
        ValueMapper: parseDataSource,
        Examples: []string{
            "change data source to /api/v1/systems" → dataSource: {url: "/api/v1/systems", type: "api"}
        },
    },
}
```

## Advanced Configuration Tool: `discover_widget_capabilities`

```json
{
  "name": "discover_widget_capabilities",
  "description": "Discover what configuration options are available for a widget by analyzing its current props and making intelligent suggestions",
  "inputSchema": {
    "type": "object",
    "properties": {
      "layout_id": {
        "type": "string",
        "description": "The layout identifier"
      },
      "widget_selector": {
        "type": "string",
        "description": "Description of which widget to analyze"
      },
      "exploration_query": {
        "type": "string",
        "description": "What you want to know about the widget's capabilities (e.g., 'what filters can I apply?', 'what chart types are supported?', 'what data sources can I use?')"
      }
    },
    "required": ["layout_id", "widget_selector", "exploration_query"]
  }
}
```

**Response Example:**
```json
{
  "widget": {
    "id": "table-1",
    "componentType": "table",
    "currentProps": {
      "dataSource": "/api/systems",
      "columns": ["name", "status", "lastUpdate"],
      "filters": [{"field": "status", "operator": "equals", "value": "active"}]
    }
  },
  "capabilities": {
    "supportedFilters": [
      {"field": "status", "operators": ["equals", "not_equals"], "values": ["active", "inactive", "critical"]},
      {"field": "severity", "operators": ["equals", "greater", "less"], "values": ["low", "medium", "high", "critical"]},
      {"field": "lastUpdate", "operators": ["since", "before"], "type": "datetime"}
    ],
    "availableColumns": ["name", "status", "severity", "lastUpdate", "description", "owner"],
    "dataSources": ["/api/systems", "/api/systems/live", "/api/systems/historical"],
    "sortableFields": ["name", "status", "severity", "lastUpdate"]
  },
  "suggestions": [
    "You can filter by severity (low, medium, high, critical)",
    "Available columns include: name, status, severity, lastUpdate, description, owner",
    "You can sort by any displayed column",
    "Live data is available at /api/systems/live"
  ]
}
```

## Intelligent Prop Generation

### Context-Aware Value Inference

```go
type ConfigurationContext struct {
    WidgetType      string                 `json:"widgetType"`
    DomainContext   string                 `json:"domainContext"`
    CurrentProps    map[string]interface{} `json:"currentProps"`
    AvailableFields []string               `json:"availableFields"`
    CommonValues    map[string][]string    `json:"commonValues"`
}

func (p *ConfigurationParser) GenerateProps(request string, context ConfigurationContext) PropChanges {
    // 1. Extract intent and entities from natural language
    intent := p.extractIntent(request)
    entities := p.extractEntities(request, context)

    // 2. Map to prop structure based on widget type and context
    propChanges := p.mapToProps(intent, entities, context)

    // 3. Validate against known patterns
    validated := p.validateProps(propChanges, context)

    return validated
}
```

### Example Transformations

**Input**: "change the filters to show only critically affected systems"

**Processing**:
1. Intent: "modify_filters"
2. Entities: ["critically", "affected", "systems"]
3. Context: table widget, monitoring domain
4. Mapping: filters → [{field: "severity", operator: "equals", value: "critical"}]

**Output**:
```json
{
  "propChanges": {
    "filters": [
      {
        "field": "severity",
        "operator": "equals",
        "value": "critical",
        "label": "Critical Systems Only"
      }
    ]
  },
  "confidence": 0.85,
  "alternatives": [
    {"field": "status", "value": "critical"},
    {"field": "alertLevel", "value": "high"}
  ]
}
```

## Feedback and Learning System

### Configuration Validation Tool: `validate_widget_config`

```json
{
  "name": "validate_widget_config",
  "description": "Validate proposed widget configuration changes before applying them, with suggestions for corrections",
  "inputSchema": {
    "type": "object",
    "properties": {
      "layout_id": {"type": "string"},
      "widget_id": {"type": "string"},
      "proposed_props": {
        "type": "object",
        "additionalProperties": true
      },
      "dry_run": {
        "type": "boolean",
        "default": true,
        "description": "Test the configuration without applying it"
      }
    },
    "required": ["layout_id", "widget_id", "proposed_props"]
  }
}
```

## Frontend Integration Patterns

### 1. Capability Registration

Allow widgets to register their configuration capabilities:

```javascript
// Frontend widget registers its capabilities
window.layoutManager.registerWidgetCapabilities('table', {
  filters: {
    supportedFields: ['status', 'severity', 'lastUpdate'],
    operators: {
      status: ['equals', 'not_equals'],
      severity: ['equals', 'greater', 'less'],
      lastUpdate: ['since', 'before']
    }
  },
  dataSources: ['/api/systems', '/api/systems/live'],
  columns: ['name', 'status', 'severity', 'lastUpdate', 'description']
});
```

### 2. Configuration Schema Introspection

```javascript
// Widget provides its current schema for introspection
const getWidgetSchema = (widgetId) => {
  const widget = findWidget(widgetId);
  return {
    currentProps: widget.props,
    schema: widget.getConfigSchema(),
    examples: widget.getConfigExamples()
  };
};
```

## Implementation Strategy

### Phase 1: Basic Prop Mapping
- Implement common prop patterns (filters, display, data source)
- Build natural language → prop mapping engine
- Support basic configuration changes

### Phase 2: Context Intelligence
- Add domain-specific vocabulary
- Implement capability discovery
- Build validation and suggestion system

### Phase 3: Learning & Adaptation
- Track successful configuration changes
- Learn from user corrections
- Improve mapping accuracy over time

This approach provides a flexible foundation for handling unknown widget configurations while being extensible enough to handle complex domain-specific requirements.