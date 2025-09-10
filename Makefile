# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run

# Binary names
SERVER_BINARY=server
MIGRATE_BINARY=migrate

# Directories
CMD_DIR=cmd
SERVER_CMD=$(CMD_DIR)/server
MIGRATE_CMD=$(CMD_DIR)/migrate

# Build the server
.PHONY: build
build:
	$(GOBUILD) -o $(SERVER_BINARY) $(SERVER_CMD)/main.go

# Build the migrate tool
.PHONY: build-migrate
build-migrate:
	$(GOBUILD) -o $(MIGRATE_BINARY) $(MIGRATE_CMD)/main.go

# Run the server
.PHONY: run
run:
	$(GORUN) $(SERVER_CMD)/main.go

# Database backup
.PHONY: db-backup
db-backup:
	@echo "Creating database backup..."
	docker exec gin-api-postgres pg_dump -U postgres gin_api_server > backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "Backup created: backup_$$(date +%Y%m%d_%H%M%S).sql"

# Database migration commands
.PHONY: migrate-up
migrate-up:
	$(GORUN) $(MIGRATE_CMD)/main.go -action=up

.PHONY: migrate-down
migrate-down:
	$(GORUN) $(MIGRATE_CMD)/main.go -action=down

.PHONY: migrate-down-1
migrate-down-1:
	$(GORUN) $(MIGRATE_CMD)/main.go -action=down -steps=1

.PHONY: migrate-up-1
migrate-up-1:
	$(GORUN) $(MIGRATE_CMD)/main.go -action=up -steps=1

.PHONY: migrate-version
migrate-version:
	$(GORUN) $(MIGRATE_CMD)/main.go -action=version

.PHONY: migrate-force
migrate-force:
	@echo "Usage: make migrate-force VERSION=<version_number>"
	@echo "Example: make migrate-force VERSION=1"

# Test database migration commands
.PHONY: migrate-test-up
migrate-test-up:
	$(GORUN) $(MIGRATE_CMD)/main.go -action=up -test

.PHONY: migrate-test-down
migrate-test-down:
	$(GORUN) $(MIGRATE_CMD)/main.go -action=down -test

.PHONY: migrate-test-down-1
migrate-test-down-1:
	$(GORUN) $(MIGRATE_CMD)/main.go -action=down -steps=1 -test

.PHONY: migrate-test-up-1
migrate-test-up-1:
	$(GORUN) $(MIGRATE_CMD)/main.go -action=up -steps=1 -test

.PHONY: migrate-test-version
migrate-test-version:
	$(GORUN) $(MIGRATE_CMD)/main.go -action=version -test

.PHONY: migrate-test-force
migrate-test-force:
	@echo "Usage: make migrate-test-force VERSION=<version_number>"
	@echo "Example: make migrate-test-force VERSION=1"

# Create a new migration file
.PHONY: migrate-create
migrate-create:
	@echo "Usage: make migrate-create NAME=<migration_name>"
	@echo "Example: make migrate-create NAME=add_user_table"

# Test commands
.PHONY: test
test:
	$(GOTEST) -v ./...

.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(SERVER_BINARY) $(MIGRATE_BINARY)
	rm -f coverage.out

# Install dependencies
.PHONY: deps
deps:
	$(GOGET) -u ./...

# Help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build          - Build the server binary"
	@echo "  build-migrate  - Build the migrate tool binary"
	@echo "  run            - Run the server"
	@echo "  db-backup      - Create database backup"
	@echo "  migrate-up     - Run all pending migrations"
	@echo "  migrate-down   - Rollback all migrations"
	@echo "  migrate-down-1 - Rollback 1 migration"
	@echo "  migrate-up-1   - Run 1 migration"
	@echo "  migrate-version- Show current migration version"
	@echo "  migrate-force  - Force migration to specific version (use VERSION=<n>)"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install/update dependencies"
	@echo "  help           - Show this help message"
