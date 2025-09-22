# React Grid Layout Manager - Backend System

## Project Overview
**Feature:** Go-based API server with PostgreSQL persistence for react-grid-layout management
**Current Branch:** widget-dashboard-manager
**Architecture:** Go API + PostgreSQL + Docker Compose + MCP Integration

## System Architecture

### Components
1. **Go API Server** - RESTful API for layout management
2. **PostgreSQL Database** - Persistent storage for layouts and configurations
3. **MCP Server Integration** - Using mark3labs/mcp-go (same as obs-mcp)
4. **Docker Compose** - Local development environment
5. **React Grid Layout Frontend** - UI component management

### Technical Stack
- **Backend:** Go 1.24.6, mark3labs/mcp-go v0.39.1
- **Database:** PostgreSQL with migrations
- **Infrastructure:** Docker Compose
- **Frontend Integration:** React Grid Layout

## Implementation Plan

### 1. Database Schema Design ✓
```sql
-- Layouts table for storing grid configurations
-- Layout items table for individual widget positions
-- Users/Organizations for multi-tenancy
-- Breakpoint-specific layouts for responsive design
```

### 2. Go API Server Structure
```
layout-manager-api/
├── cmd/layout-manager/           # Main application
├── pkg/api/                      # HTTP handlers
├── pkg/db/                       # Database layer
├── pkg/models/                   # Data models
├── pkg/mcp/                      # MCP server integration
└── pkg/config/                   # Configuration
```

### 3. React Grid Layout Data Models
Based on documentation analysis:
```go
type LayoutItem struct {
    I      string  `json:"i"`      // Item identifier
    X      int     `json:"x"`      // Grid column position
    Y      int     `json:"y"`      // Grid row position
    W      int     `json:"w"`      // Width in grid units
    H      int     `json:"h"`      // Height in grid units
    Static bool    `json:"static"` // Cannot be dragged/resized
    MinW   *int    `json:"minW"`   // Minimum width
    MaxW   *int    `json:"maxW"`   // Maximum width
    MinH   *int    `json:"minH"`   // Minimum height
    MaxH   *int    `json:"maxH"`   // Maximum height
}
```

### 4. MCP Integration Points
- Layout persistence tools
- Layout retrieval tools
- Configuration management tools
- Real-time layout synchronization

## Development Tasks

### Phase 1: Foundation
- [x] Analyze obs-mcp MCP patterns
- [x] Research react-grid-layout API
- [ ] Design PostgreSQL schema
- [ ] Create Docker Compose setup
- [ ] Initialize Go module structure

### Phase 2: Core Backend
- [ ] Implement database models and migrations
- [ ] Build REST API endpoints
- [ ] Add MCP server integration
- [ ] Implement layout persistence logic

### Phase 3: Integration & Testing
- [ ] Frontend integration testing
- [ ] MCP tool validation
- [ ] Performance optimization
- [ ] Documentation

---
*Created: 2025-09-19*