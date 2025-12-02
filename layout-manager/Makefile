# Layout Manager MCP Server - Development Makefile

# Variables
BINARY_NAME=layout-manager
MAIN_PATH=./cmd/layout-manager
DOCKER_COMPOSE=docker-compose.yml
CONTAINER_CLI ?= docker

# Go commands
GO_BUILD=go build
GO_CLEAN=go clean
GO_TEST=go test
GO_GET=go get
GO_MOD=go mod

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Layout Manager MCP Server - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development Commands
.PHONY: dev
dev: ## Start development environment (database + MCP server)
	@echo "Starting development environment..."
	docker-compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 5
	make run

.PHONY: run
run: ## Run the MCP server
	@echo "Starting Layout Manager MCP Server..."
	$(GO_BUILD) -o $(BINARY_NAME) $(MAIN_PATH)
	./$(BINARY_NAME)

.PHONY: build
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	$(GO_BUILD) -o $(BINARY_NAME) $(MAIN_PATH)

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GO_CLEAN)
	rm -f $(BINARY_NAME)


# Database Management
.PHONY: db-up
db-up: ## Start database container
	@echo "Starting PostgreSQL database..."
	docker-compose up -d postgres

.PHONY: db-down
db-down: ## Stop database container
	@echo "Stopping database..."
	docker-compose down

.PHONY: db-logs
db-logs: ## Show database logs
	docker-compose logs -f postgres

.PHONY: db-shell
db-shell: ## Connect to database shell
	@echo "Connecting to PostgreSQL..."
	docker-compose exec postgres psql -U postgres -d layout_manager

.PHONY: db-reset
db-reset: ## Reset database (WARNING: destroys all data)
	@echo "WARNING: This will destroy all data in the database!"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	docker-compose down -v
	docker-compose up -d postgres
	@echo "Database reset complete"

# Testing
.PHONY: test
test: test-setup test-run test-cleanup ## Run full test suite with fresh database

.PHONY: test-setup
test-setup: ## Set up test environment
	@echo "Setting up test environment..."
	docker-compose -f docker-compose.test.yml down --volumes --remove-orphans
	docker-compose -f docker-compose.test.yml up -d
	@echo "Waiting for test database to be ready..."
	@for i in `seq 1 30`; do \
		if docker-compose -f docker-compose.test.yml exec -T postgres-test pg_isready -U postgres -d layout_manager_test >/dev/null 2>&1; then \
			echo "Test database is ready"; \
			break; \
		fi; \
		echo "Waiting for test database... ($$i/30)"; \
		sleep 2; \
	done

.PHONY: test-run
test-run: ## Run tests with test environment
	@echo "Running tests with fresh database..."
	ENV_FILE=.env.test $(GO_TEST) -v ./pkg/...

.PHONY: test-cleanup
test-cleanup: ## Clean up test environment
	@echo "Cleaning up test environment..."
	docker-compose -f docker-compose.test.yml down --volumes

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	$(GO_TEST) -v ./pkg/...

.PHONY: test-integration
test-integration: ## Run integration tests with test DB
	@echo "Running integration tests..."
	make test-setup
	ENV_FILE=.env.test $(GO_TEST) -v ./tests/integration/...
	make test-cleanup


.PHONY: test-mcp
test-mcp: ## Run MCP tests with test DB
	@echo "Running MCP tests..."
	make test-setup
	ENV_FILE=.env.test $(GO_TEST) -v ./pkg/mcp/...
	make test-cleanup

.PHONY: test-watch
test-watch: ## Run tests in watch mode
	@echo "Running tests in watch mode..."
	find . -name "*.go" | entr -r make test-run

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	make test-setup
	ENV_FILE=.env.test $(GO_TEST) -v -coverprofile=coverage.out ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	make test-cleanup
	@echo "Coverage report generated: coverage.html"

# Dependencies
.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	$(GO_MOD) download
	$(GO_MOD) tidy

.PHONY: deps-upgrade
deps-upgrade: ## Upgrade dependencies
	@echo "Upgrading dependencies..."
	$(GO_GET) -u ./...
	$(GO_MOD) tidy

# Linting and Formatting
.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

# Docker Commands
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	$(CONTAINER_CLI) build -t layout-manager:latest .

.PHONY: docker-up
docker-up: ## Start full stack with Docker Compose
	@echo "Starting full stack..."
	docker-compose up -d

.PHONY: docker-down
docker-down: ## Stop full stack
	@echo "Stopping full stack..."
	docker-compose down

.PHONY: docker-logs
docker-logs: ## Show all container logs
	docker-compose logs -f

# Health Checks
.PHONY: health
health: ## Check MCP server health
	@echo "Checking MCP server health..."
	@curl -f http://localhost:9090/health || echo "MCP server is not responding"

.PHONY: check-env
check-env: ## Validate environment configuration
	@echo "Checking environment configuration..."
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found"; \
		exit 1; \
	fi
	@echo "✓ .env file exists"
	@echo "Environment variables:"
	@grep -v '^#' .env | grep -v '^$$'

# Implementation Stages (for development workflow)
.PHONY: stage-1
stage-1: ## Complete Stage 1: Docker Environment
	@echo "Stage 1: Setting up Docker environment..."
	make db-up
	make check-env
	@echo "✓ Stage 1 complete"

.PHONY: stage-2
stage-2: ## Complete Stage 2: Foundation Setup
	@echo "Stage 2: Foundation setup..."
	make deps
	make build
	@echo "✓ Stage 2 complete"

.PHONY: stage-3
stage-3: ## Complete Stage 3: Database Layer
	@echo "Stage 3: Database layer..."
	make db-up
	@echo "TODO: Implement GORM models"
	@echo "✓ Stage 3 preparation complete"

.PHONY: stage-4
stage-4: ## Complete Stage 4: MCP Server Implementation
	@echo "Stage 4: MCP server implementation..."
	make build
	make test-mcp
	@echo "✓ Stage 4 complete"

# Quick development workflow
.PHONY: quick-start
quick-start: ## Quick start for development (db + build + run)
	@echo "Quick starting development environment..."
	make db-up
	make build
	make run

.PHONY: full-reset
full-reset: ## Full reset and restart
	@echo "Performing full reset..."
	make clean
	make db-reset
	make deps
	make build
	@echo "✓ Full reset complete. Run 'make run' to start."

# Utility commands
.PHONY: logs
logs: ## Show application logs (when running in background)
	tail -f layout-manager.log

.PHONY: status
status: ## Show status of all services
	@echo "=== Docker Services ==="
	docker-compose ps
	@echo ""
	@echo "=== MCP Server Health ==="
	make health

# Install development tools
.PHONY: install-tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✓ Development tools installed"