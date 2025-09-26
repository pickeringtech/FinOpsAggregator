#!/bin/bash

# Development Environment Setup Script
# This script sets up the complete local development environment

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
POSTGRES_USER="finops"
POSTGRES_PASSWORD="finops"
POSTGRES_DB="finops"
POSTGRES_PORT="5432"
CONTAINER_NAME="finops-postgres"

echo -e "${BLUE}🚀 Setting up FinOps Development Environment${NC}"
echo "=================================================="

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo -e "${YELLOW}⚠️  docker-compose not found, trying docker compose...${NC}"
    if ! docker compose version &> /dev/null; then
        echo -e "${RED}❌ Docker Compose is not installed. Please install Docker Compose first.${NC}"
        exit 1
    fi
    DOCKER_COMPOSE_CMD="docker compose"
else
    DOCKER_COMPOSE_CMD="docker-compose"
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}⚠️  Go is not installed. You'll need Go 1.22+ to build the application.${NC}"
    echo "   Download from: https://golang.org/dl/"
fi

# Function to check if container is running
is_container_running() {
    docker ps --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"
}

# Function to check if container exists (running or stopped)
container_exists() {
    docker ps -a --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"
}

echo -e "${BLUE}📦 Setting up PostgreSQL database...${NC}"

# Stop and remove existing container if it exists
if container_exists; then
    echo -e "${YELLOW}🔄 Removing existing PostgreSQL container...${NC}"
    docker stop $CONTAINER_NAME 2>/dev/null || true
    docker rm $CONTAINER_NAME 2>/dev/null || true
fi

# Start PostgreSQL container
echo -e "${BLUE}🐘 Starting PostgreSQL container...${NC}"
docker run -d \
    --name $CONTAINER_NAME \
    -e POSTGRES_USER=$POSTGRES_USER \
    -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
    -e POSTGRES_DB=$POSTGRES_DB \
    -p $POSTGRES_PORT:5432 \
    -v finops_postgres_data:/var/lib/postgresql/data \
    postgres:14

# Wait for PostgreSQL to be ready
echo -e "${BLUE}⏳ Waiting for PostgreSQL to be ready...${NC}"
for i in {1..30}; do
    if docker exec $CONTAINER_NAME pg_isready -U $POSTGRES_USER -d $POSTGRES_DB &>/dev/null; then
        echo -e "${GREEN}✅ PostgreSQL is ready!${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}❌ PostgreSQL failed to start within 30 seconds${NC}"
        exit 1
    fi
    sleep 1
done

# Install Go dependencies if Go is available
if command -v go &> /dev/null; then
    echo -e "${BLUE}📚 Installing Go dependencies...${NC}"
    go mod tidy
    echo -e "${GREEN}✅ Go dependencies installed${NC}"
else
    echo -e "${YELLOW}⚠️  Skipping Go dependencies (Go not installed)${NC}"
fi

# Install golang-migrate if not present
if ! command -v migrate &> /dev/null; then
    echo -e "${BLUE}🔧 Installing golang-migrate...${NC}"
    if command -v go &> /dev/null; then
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        echo -e "${GREEN}✅ golang-migrate installed${NC}"
    else
        echo -e "${YELLOW}⚠️  Cannot install golang-migrate without Go. Please install manually.${NC}"
        echo "   Instructions: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate"
    fi
fi

# Run database migrations
echo -e "${BLUE}🗄️  Running database migrations...${NC}"
if command -v migrate &> /dev/null; then
    migrate -path migrations -database "postgresql://$POSTGRES_USER:$POSTGRES_PASSWORD@localhost:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable" up
    echo -e "${GREEN}✅ Database migrations completed${NC}"
else
    echo -e "${YELLOW}⚠️  Skipping migrations (migrate command not available)${NC}"
    echo "   Run manually: make migrate-up"
fi

# Create config file if it doesn't exist
if [ ! -f config.yaml ]; then
    echo -e "${BLUE}⚙️  Creating config file...${NC}"
    cp config.yaml.example config.yaml
    echo -e "${GREEN}✅ Config file created (config.yaml)${NC}"
else
    echo -e "${YELLOW}⚠️  Config file already exists${NC}"
fi

# Build the application if Go is available
if command -v go &> /dev/null; then
    echo -e "${BLUE}🔨 Building application...${NC}"
    go build -o bin/finops ./cmd/finops
    echo -e "${GREEN}✅ Application built successfully${NC}"
else
    echo -e "${YELLOW}⚠️  Skipping build (Go not installed)${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Development environment setup complete!${NC}"
echo "=================================================="
echo -e "${BLUE}Database Info:${NC}"
echo "  Host: localhost"
echo "  Port: $POSTGRES_PORT"
echo "  Database: $POSTGRES_DB"
echo "  Username: $POSTGRES_USER"
echo "  Password: $POSTGRES_PASSWORD"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
if command -v go &> /dev/null; then
    echo "  1. Load demo data:     ./bin/finops demo seed"
    echo "  2. Validate graph:     ./bin/finops graph validate"
    echo "  3. Run allocation:     ./bin/finops allocate --from 2024-01-01 --to 2024-01-31"
    echo "  4. Or run all:         make demo-full"
else
    echo "  1. Install Go 1.22+:   https://golang.org/dl/"
    echo "  2. Run setup again:    ./scripts/dev-setup.sh"
    echo "  3. Load demo data:     make demo-full"
fi
echo ""
echo -e "${BLUE}Useful Commands:${NC}"
echo "  Stop database:         docker stop $CONTAINER_NAME"
echo "  Start database:        docker start $CONTAINER_NAME"
echo "  View logs:             docker logs $CONTAINER_NAME"
echo "  Connect to DB:         psql postgresql://$POSTGRES_USER:$POSTGRES_PASSWORD@localhost:$POSTGRES_PORT/$POSTGRES_DB"
echo "  Reset environment:     ./scripts/dev-reset.sh"
