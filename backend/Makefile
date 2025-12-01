.PHONY: build run test clean deps migrate-up migrate-down docker-build docker-run

# Variables
BINARY_NAME=finops
DOCKER_IMAGE=finops-aggregator
POSTGRES_URL=postgresql://finops:finops@localhost:5432/finops?sslmode=disable

# Build the application
build:
	go build -o bin/$(BINARY_NAME) ./cmd/finops

# Run the application
run: build
	./bin/$(BINARY_NAME)

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Database migrations
migrate-up:
	migrate -path migrations -database "$(POSTGRES_URL)" up

migrate-down:
	migrate -path migrations -database "$(POSTGRES_URL)" down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations $$name

# Development database setup
dev-db-start:
	docker run --name finops-postgres -e POSTGRES_USER=finops -e POSTGRES_PASSWORD=finops -e POSTGRES_DB=finops -p 5432:5432 -d postgres:14

dev-db-stop:
	docker stop finops-postgres || true
	docker rm finops-postgres || true

dev-db-reset: dev-db-stop dev-db-start
	sleep 3
	$(MAKE) migrate-up

# Demo commands
demo-seed: build
	./bin/$(BINARY_NAME) demo seed

demo-validate: build
	./bin/$(BINARY_NAME) graph validate

demo-allocate: build
	@echo "Running demo allocation over seeded period..."
	./bin/$(BINARY_NAME) demo allocate

demo-charts: build
	chmod +x scripts/generate-charts.sh
	./scripts/generate-charts.sh demo

demo-full: demo-seed demo-validate demo-allocate demo-charts

# Chart generation commands
charts-graph: build
	chmod +x scripts/generate-charts.sh
	./scripts/generate-charts.sh graph

charts-trends: build
	chmod +x scripts/generate-charts.sh
	./scripts/generate-charts.sh trends

charts-all: build
	chmod +x scripts/generate-charts.sh
	./scripts/generate-charts.sh all

# Testing commands
test-charts: build
	chmod +x scripts/test-charts.sh
	./scripts/test-charts.sh

test-unit:
	go test -v ./internal/...

test-integration: build
	./scripts/test-charts.sh basic

debug-charts: build
	chmod +x scripts/debug-charts.sh
	./scripts/debug-charts.sh

verify-cli: build
	chmod +x scripts/verify-cli.sh
	./scripts/verify-cli.sh

test-chart-step-by-step:
	chmod +x scripts/test-chart-step-by-step.sh
	./scripts/test-chart-step-by-step.sh

test-chart-direct:
	go run test-chart-direct.go

check-syntax:
	chmod +x scripts/check-syntax.sh
	./scripts/check-syntax.sh

diagnose-data: build
	chmod +x scripts/diagnose-data.sh
	./scripts/diagnose-data.sh

finops-demo: build
	chmod +x scripts/finops-demo.sh
	./scripts/finops-demo.sh

# Docker commands
docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker run --rm -it $(DOCKER_IMAGE)

# Linting and formatting
fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet
	golangci-lint run

# Install development tools
install-tools:
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Build and run the application"
	@echo "  deps           - Install dependencies"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean          - Clean build artifacts"
	@echo "  migrate-up     - Run database migrations up"
	@echo "  migrate-down   - Run database migrations down"
	@echo "  migrate-create - Create a new migration"
	@echo "  dev-db-start   - Start development PostgreSQL container"
	@echo "  dev-db-stop    - Stop development PostgreSQL container"
	@echo "  dev-db-reset   - Reset development database"
	@echo "  demo-seed      - Load demo seed data"
	@echo "  demo-validate  - Validate graph structure"
	@echo "  demo-allocate  - Run demo allocation"
	@echo "  demo-charts    - Generate demo charts"
	@echo "  demo-full      - Run full demo (seed + validate + allocate + charts)"
	@echo "  charts-graph   - Generate graph structure chart"
	@echo "  charts-trends  - Generate cost trend charts"
	@echo "  charts-all     - Generate all charts"
	@echo "  test-unit      - Run unit tests"
	@echo "  test-charts    - Test chart generation"
	@echo "  test-integration - Run integration tests"
	@echo "  debug-charts   - Debug chart generation issues"
	@echo "  verify-cli     - Verify CLI structure and commands"
	@echo "  test-chart-step-by-step - Step-by-step chart generation test"
	@echo "  test-chart-direct - Direct chart generation test (no CLI)"
	@echo "  check-syntax   - Check for common Go compilation issues"
	@echo "  diagnose-data  - Diagnose database data and chart generation issues"
	@echo "  finops-demo    - Run complete FinOps demonstration"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  fmt            - Format Go code"
	@echo "  vet            - Run go vet"
	@echo "  lint           - Run linting and formatting"
	@echo "  install-tools  - Install development tools"
	@echo "  help           - Show this help message"
