.PHONY: help dev-up dev-down dev-logs dev-restart dev-clean dev-rebuild dev-seed dev-migrate dev-psql dev-backend-shell dev-frontend-shell dev-test dev-debug

# Default target
help:
	@echo "FinOps Aggregator - Development Environment"
	@echo ""
	@echo "Available targets:"
	@echo "  make dev-up              - Start development environment"
	@echo "  make dev-down            - Stop development environment"
	@echo "  make dev-logs            - View logs from all services"
	@echo "  make dev-restart         - Restart all services"
	@echo "  make dev-clean           - Stop and remove all containers and volumes"
	@echo "  make dev-rebuild         - Rebuild all containers from scratch"
	@echo "  make dev-seed            - Seed database with demo data"
	@echo "  make dev-migrate         - Run database migrations"
	@echo "  make dev-psql            - Connect to PostgreSQL database"
	@echo "  make dev-backend-shell   - Open shell in backend container"
	@echo "  make dev-frontend-shell  - Open shell in frontend container"
	@echo "  make dev-test            - Run backend tests"
	@echo "  make dev-debug           - Start backend with debugger attached"
	@echo ""

# Start development environment
dev-up:
	@echo "Starting development environment..."
	docker-compose -f docker-compose.dev.yml up -d
	@echo ""
	@echo "✅ Development environment started!"
	@echo ""
	@echo "Services:"
	@echo "  Frontend:  http://localhost:3000"
	@echo "  Backend:   http://localhost:8080"
	@echo "  Postgres:  localhost:5432"
	@echo "  Debugger:  localhost:2345"
	@echo ""
	@echo "View logs: make dev-logs"

# Stop development environment
dev-down:
	@echo "Stopping development environment..."
	docker-compose -f docker-compose.dev.yml down
	@echo "✅ Development environment stopped!"

# View logs
dev-logs:
	docker-compose -f docker-compose.dev.yml logs -f

# View backend logs only
dev-logs-backend:
	docker-compose -f docker-compose.dev.yml logs -f backend

# View frontend logs only
dev-logs-frontend:
	docker-compose -f docker-compose.dev.yml logs -f frontend

# View postgres logs only
dev-logs-postgres:
	docker-compose -f docker-compose.dev.yml logs -f postgres

# Restart all services
dev-restart:
	@echo "Restarting all services..."
	docker-compose -f docker-compose.dev.yml restart
	@echo "✅ Services restarted!"

# Restart backend only
dev-restart-backend:
	@echo "Restarting backend..."
	docker-compose -f docker-compose.dev.yml restart backend
	@echo "✅ Backend restarted!"

# Restart frontend only
dev-restart-frontend:
	@echo "Restarting frontend..."
	docker-compose -f docker-compose.dev.yml restart frontend
	@echo "✅ Frontend restarted!"

# Clean everything (stop and remove volumes)
dev-clean:
	@echo "Cleaning development environment..."
	docker-compose -f docker-compose.dev.yml down -v
	@echo "✅ Development environment cleaned!"

# Rebuild all containers from scratch
dev-rebuild:
	@echo "Rebuilding all containers..."
	docker-compose -f docker-compose.dev.yml build --no-cache
	@echo "✅ Containers rebuilt!"

# Rebuild and restart
dev-rebuild-up:
	@echo "Rebuilding and starting..."
	docker-compose -f docker-compose.dev.yml up -d --build
	@echo "✅ Containers rebuilt and started!"

# Seed database with demo data
dev-seed:
	@echo "Seeding database with demo data..."
	docker-compose -f docker-compose.dev.yml exec backend sh -c "cd /app && make demo-seed"
	@echo "✅ Database seeded!"

# Run database migrations
dev-migrate:
	@echo "Running database migrations..."
	docker-compose -f docker-compose.dev.yml run --rm migrate
	@echo "✅ Migrations complete!"

# Connect to PostgreSQL
dev-psql:
	docker exec -it finops-postgres-dev psql -U finops -d finops

# Open shell in backend container
dev-backend-shell:
	docker exec -it finops-backend-dev sh

# Open shell in frontend container
dev-frontend-shell:
	docker exec -it finops-frontend-dev sh

# Run backend tests
dev-test:
	@echo "Running backend tests..."
	docker-compose -f docker-compose.dev.yml exec backend go test ./...

# Run backend tests with coverage
dev-test-coverage:
	@echo "Running backend tests with coverage..."
	docker-compose -f docker-compose.dev.yml exec backend go test -cover ./...

# Start backend with debugger
dev-debug:
	@echo "Starting backend with debugger..."
	@echo "Debugger will be available on localhost:2345"
	@echo ""
	@echo "VS Code: Press F5 to attach debugger"
	@echo "GoLand: Run > Attach to Process > localhost:2345"
	@echo ""
	docker-compose -f docker-compose.dev.yml up backend

# Check service status
dev-status:
	docker-compose -f docker-compose.dev.yml ps

# View resource usage
dev-stats:
	docker stats

# Prune Docker system
dev-prune:
	@echo "Pruning Docker system..."
	docker system prune -f
	@echo "✅ Docker system pruned!"

# Full reset (clean + rebuild + up)
dev-reset:
	@echo "Performing full reset..."
	$(MAKE) dev-clean
	$(MAKE) dev-rebuild
	$(MAKE) dev-up
	@echo "✅ Full reset complete!"

# Install backend dependencies
dev-backend-deps:
	@echo "Installing backend dependencies..."
	docker-compose -f docker-compose.dev.yml exec backend go mod download
	docker-compose -f docker-compose.dev.yml exec backend go mod tidy
	@echo "✅ Backend dependencies installed!"

# Install frontend dependencies
dev-frontend-deps:
	@echo "Installing frontend dependencies..."
	docker-compose -f docker-compose.dev.yml exec frontend npm install
	@echo "✅ Frontend dependencies installed!"

# Run backend linter
dev-lint-backend:
	@echo "Running backend linter..."
	docker-compose -f docker-compose.dev.yml exec backend go vet ./...

# Run frontend linter
dev-lint-frontend:
	@echo "Running frontend linter..."
	docker-compose -f docker-compose.dev.yml exec frontend npm run lint

# Format backend code
dev-fmt-backend:
	@echo "Formatting backend code..."
	docker-compose -f docker-compose.dev.yml exec backend go fmt ./...
	@echo "✅ Backend code formatted!"

# Build backend binary
dev-build-backend:
	@echo "Building backend binary..."
	docker-compose -f docker-compose.dev.yml exec backend go build -o bin/finops ./cmd/finops
	@echo "✅ Backend binary built!"

# Build frontend
dev-build-frontend:
	@echo "Building frontend..."
	docker-compose -f docker-compose.dev.yml exec frontend npm run build
	@echo "✅ Frontend built!"

