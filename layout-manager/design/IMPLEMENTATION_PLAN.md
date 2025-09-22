# Layout Manager POC - Implementation Plan

## Overview

This document outlines the step-by-step implementation plan for the Layout Manager POC based on our complete design documentation.

## Design Foundation Summary

### What We Have Implemented ✅
- **Database Schema**: PostgreSQL + GORM with JSONB storage and `is_active` constraints (`DATABASE_SCHEMA.md`)
- **MCP Tools**: 8 natural language tools for precise widget manipulation (`MCP_TOOLS_DESIGN.md`)
- **Active Dashboard System**: Single active dashboard with automatic activation
- **Precise Widget Manipulation**: Direct database operations using exact widget IDs
- **Generic Configuration**: Handle unknown widget props (`GENERIC_WIDGET_CONFIG_DESIGN.md`)
- **React Grid Layout Integration**: Complete data structures (`REACT_GRID_LAYOUT_DOCS.md`)

### Project Architecture (MCP-Only)
```
layout-manager/
├── design/                    # Design docs (updated for MCP-only)
├── cmd/layout-manager/        # MCP server main application
├── pkg/
│   ├── db/                    # Database connection & repository layer
│   ├── models/                # Data models with GORM tags and active dashboard support
│   ├── mcp/                   # MCP server, tools, and direct database integration
│   └── config/                # Configuration management
├── docker-compose.yml         # PostgreSQL development environment
└── go.mod                    # Go dependencies (mark3labs/mcp-go)
```

## Implementation Methodology

### Design Review and Validation Process

**After Each Stage:**
1. **Validate Against Design Documents** - Ensure implementation matches design specifications
2. **Test Core Functionality** - Verify stage deliverables work as designed
3. **Document Implementation Progress** - Record what was completed and any deviations
4. **Review Next Stage Requirements** - Check if any design assumptions need updating
5. **Update Design if Needed** - Document any deviations or improvements discovered during implementation
6. **Create Session Checkpoint** - Document current state for easy resumption

### Implementation Feedback Loop

```
Design Docs → Implementation → Testing → Design Review → Next Stage
     ↑                                          ↓
     ←──────── Update if needed ←───────────────┘
```

**Review Checkpoints:**
- **After Stage 2 (Database)**: Validate GORM models match OpenAPI schema exactly
- **After Stage 3 (Code Gen)**: Verify generated code integrates with GORM models properly
- **After Stage 4 (MCP Server)**: Test natural language parsing against design examples
- **After Stage 6 (Integration)**: Validate end-to-end flow matches MCP tool designs
- **After Stage 8 (Testing)**: Final review of all design assumptions and document learnings

### Documentation Updates Required

**During Implementation:**
- Note any design assumptions that prove incorrect
- Document implementation-specific decisions not covered in design
- Update design docs if significant changes are needed
- Record lessons learned for future iterations

**Post-Implementation:**
- Create implementation notes document
- Update design docs with any corrections
- Document performance characteristics discovered
- Note frontend integration requirements discovered

### Session Tracking and Resumption

**Progress Tracking File:** `IMPLEMENTATION_PROGRESS.md`

**After Each Stage, Update Progress File With:**
1. **Stage Status** - Completed, in-progress, blocked, or not started
2. **Implementation Decisions** - Any choices made during implementation
3. **Code Changes** - Files created, modified, or generated
4. **Issues Encountered** - Problems and their solutions
5. **Next Steps** - What to do when resuming
6. **Design Updates** - Any changes to original design docs

**Session Checkpoint Format:**
```markdown
## Session Checkpoint - [Date/Time]

### Completed Stages: 1, 2, 3
### Current Stage: 4 (MCP Server Implementation)
### Progress: 60% - MCP tools defined, working on handlers

### Files Created/Modified:
- go.mod, go.sum
- pkg/models/layout.go
- pkg/db/connection.go
- pkg/api/generated.go (via oapi-codegen)

### Current Issues:
- None blocking

### Next Session Tasks:
1. Complete MCP tool handlers
2. Test natural language parsing
3. Move to Stage 5

### Design Deviations:
- None so far

### Environment Status:
- Database: Running via Docker Compose
- Dependencies: All installed
- Code: Compiles successfully
```

## Implementation Stages

### Stage 1: Docker Development Environment (15-30 minutes)

**Design Review Focus:** Database connectivity and development environment setup

**Priority Rationale:** Setting up the development environment first allows us to:
- Test database connections immediately
- Validate GORM models against real PostgreSQL
- Have a consistent development environment
- Verify Docker configuration works

#### 1.1 Docker Compose Setup
- **Base on**: Docker configuration from `DATABASE_SCHEMA.md`
- **File**: `docker-compose.yml`
- **Services**: PostgreSQL database only (API server added later)
- **Purpose**: Provide consistent development database

#### 1.2 Environment Configuration
- **File**: `.env` (optional for POC)
- **Purpose**: Database connection parameters
- **Content**: PostgreSQL credentials, database name, port

#### 1.3 Database Verification
- Test PostgreSQL container starts
- Verify database connectivity
- Confirm database creation

**Completion Criteria:**
- Docker Compose starts PostgreSQL successfully
- Database is accessible on localhost:5432
- Can connect with psql or database client

**Design Review Checkpoint:**
- [ ] PostgreSQL container starts without errors
- [ ] Database accepts connections
- [ ] Environment configuration works
- [ ] Docker setup matches design specifications

### Stage 2: Foundation Setup (30-45 minutes)

**Design Review Focus:** Project structure aligns with planned architecture

#### 2.1 Development Workflow Setup
```bash
# Create comprehensive Makefile for development
make help          # Show all available commands
make dev           # Start development environment
make quick-start   # Quick start (db + build + run)
make stage-N       # Run specific implementation stages
```

#### 2.2 Project Structure & Dependencies
```bash
# Create Go module
go mod init github.com/layout-manager/api

# Use Makefile for dependency management
make deps          # Download and tidy dependencies
make install-tools # Install development tools

# Core dependencies (automatically managed):
# - gorm.io/gorm, gorm.io/driver/postgres, gorm.io/datatypes
# - github.com/google/uuid, github.com/mark3labs/mcp-go
# - github.com/go-chi/chi/v5 + middleware ecosystem
# - github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
# - github.com/joho/godotenv (for .env loading)
```

#### 2.2 Environment Configuration
- **File**: `.env` - Environment variables for database and server
- **Integration**: godotenv for automatic .env loading
- **Validation**: `make check-env` to verify configuration

#### 2.3 oapi-codegen Configuration
- **File**: `pkg/api/config.yaml`
- **Purpose**: Generate Go code from OpenAPI spec
- **Output**: `pkg/api/generated.go`
- **Dependencies**: Uses our OpenAPI spec with x-go-type extensions

#### 2.4 Basic Project Structure
- Create all necessary directories (pkg/api, pkg/db, pkg/models, etc.)
- Add main.go with Chi router and middleware setup
- Set up configuration management with .env integration
- Create comprehensive Makefile for development workflow

**Completion Criteria:**
- Project compiles without errors (`make build`)
- All directories created with proper structure
- Dependencies resolved (`make deps`)
- Environment configuration works (`make check-env`)
- Development workflow functional (`make help`)

**Design Review Checkpoint:**
- [ ] Project structure matches planned architecture
- [ ] oapi-codegen configuration is correct
- [ ] All required dependencies are present
- [ ] No unexpected issues with Go module setup

### Stage 3: Database Layer (45-60 minutes)

**Design Review Focus:** GORM models exactly match OpenAPI schema with proper validation

**Priority Note:** Now we can test database connections against the running PostgreSQL container

#### 3.1 GORM Models (`pkg/models/`)
- **Base on**: `DATABASE_SCHEMA.md` designs
- **Files**:
  - `layout.go` - Main Layout model with UUID and JSONB
  - `layout_schema.go` - LayoutSchema, LayoutItem, and related structs
  - `component_registry.go` - Component definitions and constraints
- **Key Features**:
  - UUID primary keys
  - JSONB validation with BeforeSave hooks
  - Component type registry

#### 3.2 Database Connection (`pkg/db/`)
- **Base on**: AutoMigrate setup from `DATABASE_SCHEMA.md`
- **Files**:
  - `connection.go` - PostgreSQL connection with GORM
  - `migrate.go` - AutoMigrate setup
- **Configuration**: Connect to Docker Compose PostgreSQL

#### 3.3 Repository Layer (`pkg/db/`)
- **File**: `repository.go`
- **Purpose**: CRUD operations for layouts
- **Methods**: Create, Get, Update, Delete, List layouts
- **Features**: JSONB querying for component types

#### 3.4 Database Testing
- **Verify**: Connection to Docker PostgreSQL works
- **Test**: AutoMigrate creates correct table structure
- **Validate**: JSONB storage and retrieval works

**Completion Criteria:**
- Database connects successfully
- AutoMigrate creates tables
- Basic CRUD operations work
- JSON validation functions

**Design Review Checkpoint:**
- [ ] GORM models match OpenAPI schema exactly (UUIDs, JSONB, validation tags)
- [ ] AutoMigrate creates correct table structure
- [ ] JSON validation hooks work as designed
- [ ] Repository methods handle JSONB queries correctly
- [ ] Component registry matches design specification
- [ ] Database connection to Docker PostgreSQL works
- [ ] JSONB storage and retrieval functions correctly

### Stage 4: API Code Generation (15-30 minutes)

**Design Review Focus:** Generated code integrates perfectly with GORM models

#### 4.1 Generate API Code
- **Input**: `design/api/openapi.yaml`
- **Config**: `pkg/api/config.yaml`
- **Output**: `pkg/api/generated.go`
- **Features**:
  - **Chi router integration** (chi-server: true)
  - **Strict server interfaces** (strict-server: true)
  - **GORM-compatible models** with x-go-type extensions
  - **Chi middleware support** for CORS, logging, etc.

#### 4.2 Verify Generated Code
- Models match GORM structure
- Handler interfaces are correct
- Validation tags are present

**Completion Criteria:**
- Code generation succeeds
- Generated models compile
- Handler interfaces defined

**Design Review Checkpoint:**
- [ ] Generated models are identical to GORM models (types, tags, structure)
- [ ] x-go-type extensions work correctly (UUID, time.Time, datatypes.JSON)
- [ ] x-oapi-codegen-extra-tags produce correct GORM and validation tags
- [ ] Handler interfaces match expected API design
- [ ] No conflicts between generated and manual code

### Stage 5: MCP Server Implementation (60-90 minutes)

**Design Review Focus:** Natural language parsing works as designed for all 6 tools

#### 4.1 Core MCP Infrastructure (`pkg/mcp/`)
- **Base on**: `MCP_TOOLS_DESIGN.md` and obs-mcp patterns
- **Files**:
  - `server.go` - MCP server setup (based on obs-mcp)
  - `tools.go` - Tool definitions
  - `handlers.go` - Tool handler implementations

#### 4.2 Natural Language Tools
Implement the 6 core tools:

1. **`find_widgets`** - Widget discovery
   - Parse natural language descriptions
   - Search layouts by component type, props, position
   - Return matching widgets with reasons

2. **`manipulate_widget`** - Primary operations tool
   - Parse commands (remove, resize, move)
   - Find target widgets
   - Execute API calls
   - Track collateral changes

3. **`add_widget`** - Widget creation
   - Parse widget descriptions
   - Calculate positions (auto-position)
   - Create with appropriate props

4. **`configure_widget`** - Generic configuration
   - **Base on**: `GENERIC_WIDGET_CONFIG_DESIGN.md`
   - Parse configuration requests
   - Map to prop structures
   - Apply intelligent defaults

5. **`batch_widget_operations`** - Multi-widget operations
   - Parse multiple commands
   - Execute atomically
   - Return combined results

6. **`analyze_layout`** - Layout insights
   - Answer questions about layouts
   - Provide structure analysis

#### 4.3 Intent Parsing System
- **Base on**: Intent classification from `MCP_TOOLS_DESIGN.md`
- Command parsing utilities
- Entity extraction (widget selectors, positions, sizes)
- Context-aware mapping

#### 4.4 Change Tracking System
- **Base on**: Collateral movement tracking design
- Before/after state comparison
- Change reason classification
- Structured response generation

**Completion Criteria:**
- MCP server starts successfully
- All 6 tools register correctly
- Basic intent parsing works
- Can execute simple commands

### Stage 5: HTTP Handlers (30-45 minutes)

#### 5.1 API Handler Implementation (`pkg/handlers/`)
- **Base on**: Generated interfaces from Stage 3
- **Files**:
  - `layouts.go` - Layout CRUD endpoints
  - `widgets.go` - Widget manipulation endpoints
  - `mcp.go` - Bridge to MCP server

#### 5.2 MCP Integration
- Bridge HTTP requests to MCP tools
- Handle natural language input
- Return structured responses

#### 5.3 Error Handling
- Simple HTTP error responses
- Basic validation error handling
- JSON error format

**Completion Criteria:**
- All API endpoints respond
- MCP integration works
- Basic error handling functions

### Stage 6: Server Setup & Integration (30 minutes)

#### 6.1 Main Application (`cmd/layout-manager/main.go`)
- Initialize database connection
- Set up MCP server
- Configure Chi HTTP router
- Add Chi middleware stack

#### 6.2 Chi Router Configuration
- Mount generated API handlers (using chi-server from oapi-codegen)
- Add Chi middleware stack:
  - `middleware.Logger` - Request logging
  - `middleware.Recoverer` - Panic recovery
  - `cors.Handler()` - CORS support
  - `middleware.RequestID` - Request tracing
- Health check endpoint
- API versioning with sub-routers

#### 6.3 Configuration
- Environment variables for database
- Hardcoded defaults for POC
- Simple flag parsing

**Completion Criteria:**
- Server starts successfully
- All endpoints accessible
- Database connects
- MCP server integrated

### Stage 7: Docker Environment (15-30 minutes)

#### 7.1 Dockerfile
- Multi-stage build for Go application
- Based on existing Docker Compose design

#### 7.2 Docker Compose Update
- **Base on**: Existing design in `DATABASE_SCHEMA.md`
- PostgreSQL service
- Layout Manager API service
- Environment configuration
- Volume mounts for development

**Completion Criteria:**
- Docker build succeeds
- Compose stack starts
- Database initializes
- API responds to requests

### Stage 8: Testing & Validation (30-45 minutes)

#### 8.1 Basic API Testing
- Test layout CRUD operations
- Test widget manipulation
- Verify JSONB storage/retrieval

#### 8.2 MCP Tool Testing
- Test each MCP tool individually
- Test natural language parsing
- Verify change tracking

#### 8.3 Integration Testing
- End-to-end widget operations
- Collateral movement verification
- Cross-breakpoint operations

**Completion Criteria:**
- All API endpoints work
- MCP tools respond correctly
- Natural language commands execute
- Change tracking accurate

## Implementation Dependencies

### External Dependencies
```go
// Core dependencies
gorm.io/gorm
gorm.io/driver/postgres
gorm.io/datatypes
github.com/google/uuid

// MCP integration
github.com/mark3labs/mcp-go

// HTTP server with Chi
github.com/go-chi/chi/v5
github.com/go-chi/chi/v5/middleware  // Built-in middleware
github.com/go-chi/cors               // CORS middleware
github.com/go-chi/render            // JSON rendering utilities

// Code generation
github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
```

### Internal Dependencies
- Each stage builds on previous stages
- Database layer required before handlers
- MCP server needed for API integration
- Generated code needed for handlers

## Key Implementation Notes

### From Design Documentation

1. **GORM Integration** (`DATABASE_SCHEMA.md`):
   - Use `datatypes.JSON` for schema field
   - Implement `BeforeSave` validation hooks
   - UUID primary keys with `gen_random_uuid()`

2. **OpenAPI Integration** (`OAPI_CODEGEN_DOCS.md`):
   - Use `x-go-type` for custom types (UUID, time.Time, datatypes.JSON)
   - Use `x-oapi-codegen-extra-tags` for GORM and validation tags
   - Generate **Chi router server** interfaces with strict mode
   - Leverage Chi middleware ecosystem

3. **MCP Patterns** (obs-mcp analysis):
   - Follow same server setup pattern
   - Use mark3labs/mcp-go library
   - Implement tool definitions and handlers separately

4. **React Grid Layout Integration** (`REACT_GRID_LAYOUT_DOCS.md`):
   - Preserve exact data structure compatibility
   - Support all breakpoint configurations
   - Handle responsive layout variations

5. **Change Tracking** (`MCP_TOOLS_DESIGN.md`):
   - Capture before/after states
   - Classify change reasons
   - Track targeted vs collateral changes

## Success Criteria

### MVP Functionality
- ✅ Create, read, update, delete layouts
- ✅ Add, remove, move, resize widgets via API
- ✅ Natural language widget operations via MCP
- ✅ Generic widget configuration handling
- ✅ Collateral movement tracking
- ✅ Multi-breakpoint support
- ✅ Docker development environment

### Demonstration Capabilities
- User can say "remove the chart widget" → finds and removes chart
- User can say "make the table larger" → identifies table, increases size
- User can say "move sales chart to top right" → finds chart, calculates position
- User can say "change filters to show only critical systems" → updates widget props
- All operations return complete change tracking for UI sync

## Estimated Timeline

**Total Implementation Time: 4-6 hours**

- Stage 1 (Docker Environment): 15-30 min
- Stage 2 (Foundation): 30-45 min
- Stage 3 (Database): 45-60 min
- Stage 4 (API Generation): 15-30 min
- Stage 5 (MCP Server): 60-90 min
- Stage 6 (HTTP Handlers): 30-45 min
- Stage 7 (Server Setup): 30 min
- Stage 8 (Testing): 30-45 min

**Note:** Docker environment setup moved to Stage 1 for immediate database connectivity validation.

This timeline assumes following the detailed design documentation and reusing patterns from obs-mcp analysis.