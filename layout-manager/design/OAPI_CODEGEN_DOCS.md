# oapi-codegen Documentation

## Overview

**oapi-codegen** is a command-line tool and library to convert OpenAPI 3.0 specifications to Go code, generating server-side implementations, API clients, or HTTP models with minimal boilerplate.

## Installation & Setup

### For Go 1.24+ (Recommended)
```bash
go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

Use in go:generate comments:
```go
//go:generate go tool oapi-codegen -config cfg.yaml ../../api.yaml
```

### Prior to Go 1.24
Create `tools/tools.go`:
```go
//go:build tools
// +build tools

package main

import (
    _ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
```

## Configuration File

Use YAML configuration with JSON Schema support:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/HEAD/configuration-schema.json
package: api
generate:
  models: true
  chi-server: true
  strict-server: true
output: gen.go
```

### Key Configuration Options

```yaml
package: "api"                    # Go package name
output: "gen.go"                 # Output file
output-options:
  skip-fmt: false                # Skip gofmt
  skip-prune: false              # Skip pruning unused code

generate:
  models: true                   # Generate model structs
  chi-server: true              # Generate Chi router server
  echo-server: false            # Generate Echo server
  gin-server: false             # Generate Gin server
  gorilla-server: false         # Generate Gorilla/mux server
  strict-server: true           # Generate strict server wrapper
  client: false                 # Generate client code

import-mapping:
  "github.com/google/uuid": "googleuuid"

compatibility:
  always-prefix-enum-values: true
  apply-chi-middleware-first-to-last: true
```

## Key Extensions for GORM Integration

### x-go-type Extension

Override generated types with custom Go types:

```yaml
properties:
  id:
    type: string
    x-go-type: uuid.UUID
    x-go-type-import:
      path: github.com/google/uuid
      name: uuid
  created_at:
    type: string
    format: date-time
    x-go-type: time.Time
    x-go-type-import:
      path: time
```

### x-oapi-codegen-extra-tags Extension

Add arbitrary struct tags for validation, GORM, logging:

```yaml
properties:
  id:
    type: string
    x-go-type: uuid.UUID
    x-oapi-codegen-extra-tags:
      gorm: "type:uuid;default:gen_random_uuid();primaryKey"
      validate: "required"
  email:
    type: string
    x-oapi-codegen-extra-tags:
      gorm: "uniqueIndex;not null"
      validate: "required,email"
      json: "email"
```

Generates:
```go
type User struct {
    Id    uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id" validate:"required"`
    Email string    `gorm:"uniqueIndex;not null" json:"email" validate:"required,email"`
}
```

## GORM Integration Patterns

### Complete Model Example

```yaml
components:
  schemas:
    Layout:
      type: object
      required: ["layout_id", "schema"]
      properties:
        id:
          type: string
          format: uuid
          x-go-type: uuid.UUID
          x-go-type-import:
            path: github.com/google/uuid
            name: uuid
          x-oapi-codegen-extra-tags:
            gorm: "type:uuid;default:gen_random_uuid();primaryKey"
        layout_id:
          type: string
          maxLength: 255
          x-oapi-codegen-extra-tags:
            gorm: "uniqueIndex;not null"
            validate: "required,min=1,max=255"
        name:
          type: string
          maxLength: 255
          x-oapi-codegen-extra-tags:
            gorm: "size:255"
            validate: "max=255"
        description:
          type: string
          maxLength: 1000
          x-oapi-codegen-extra-tags:
            gorm: "size:1000"
            validate: "max=1000"
        schema:
          $ref: '#/components/schemas/LayoutSchema'
          x-oapi-codegen-extra-tags:
            gorm: "type:jsonb;not null"
            validate: "required"
        created_at:
          type: string
          format: date-time
          readOnly: true
          x-go-type: time.Time
          x-go-type-import:
            path: time
          x-oapi-codegen-extra-tags:
            gorm: "autoCreateTime"
        updated_at:
          type: string
          format: date-time
          readOnly: true
          x-go-type: time.Time
          x-go-type-import:
            path: time
          x-oapi-codegen-extra-tags:
            gorm: "autoUpdateTime"
```

### JSON/JSONB Fields

For complex JSON fields, use datatypes.JSON:

```yaml
schema:
  type: object
  x-go-type: datatypes.JSON
  x-go-type-import:
    path: gorm.io/datatypes
    name: datatypes
  x-oapi-codegen-extra-tags:
    gorm: "type:jsonb;not null"
```

## Server Generation

### Basic Server Interface

Generated interfaces:
```go
type ServerInterface interface {
    // Create layout
    // (POST /layouts)
    CreateLayout(w http.ResponseWriter, r *http.Request)

    // Get layout by ID
    // (GET /layouts/{layoutId})
    GetLayout(w http.ResponseWriter, r *http.Request, layoutId string)
}
```

### Strict Server Mode

For reduced boilerplate and better error handling:

```yaml
generate:
  strict-server: true
```

Generates:
```go
type StrictServerInterface interface {
    CreateLayout(ctx context.Context, request CreateLayoutRequestObject) (CreateLayoutResponseObject, error)
    GetLayout(ctx context.Context, request GetLayoutRequestObject) (GetLayoutResponseObject, error)
}
```

## Request/Response Objects

### Request Objects
```go
type CreateLayoutRequestObject struct {
    Body *CreateLayoutJSONRequestBody
}

type GetLayoutRequestObject struct {
    LayoutId string `json:"layoutId"`
}
```

### Response Objects
```go
type CreateLayoutResponseObject interface {
    VisitCreateLayoutResponse(w http.ResponseWriter) error
}

type CreateLayout201JSONResponse Layout

func (response CreateLayout201JSONResponse) VisitCreateLayoutResponse(w http.ResponseWriter) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(201)
    return json.NewEncoder(w).Encode(response)
}
```

## Validation Integration

### go-playground/validator

```yaml
properties:
  email:
    type: string
    format: email
    x-oapi-codegen-extra-tags:
      validate: "required,email"
  age:
    type: integer
    minimum: 0
    maximum: 150
    x-oapi-codegen-extra-tags:
      validate: "min=0,max=150"
```

## File Structure

```
project/
├── api/
│   ├── openapi.yaml           # OpenAPI specification
│   ├── config.yaml            # oapi-codegen config
│   └── generated.go           # Generated code
├── internal/
│   ├── handlers/             # Implementation
│   └── models/              # Additional model logic
└── main.go
```

## Best Practices

1. **Separate concerns**: Use composition to separate API models from business logic
2. **Validation**: Leverage both OpenAPI validation and Go struct tags
3. **GORM tags**: Use x-oapi-codegen-extra-tags for database constraints
4. **Imports**: Use x-go-type-import for external dependencies
5. **Versioning**: Keep API spec and generated code in sync
6. **Testing**: Generate mock interfaces for testing

## Example Configuration for Layout Manager

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/HEAD/configuration-schema.json
package: api
output: pkg/api/generated.go

generate:
  models: true
  chi-server: true
  strict-server: true

import-mapping:
  "github.com/google/uuid": uuid
  "gorm.io/datatypes": datatypes
  "time": time

output-options:
  skip-fmt: false
  user-templates:
    - template-file: custom.tmpl
```

This approach ensures API specifications and GORM models stay synchronized while providing type safety and validation.