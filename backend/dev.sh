#!/bin/bash

# FinOps Aggregator Development Script
# This script starts the PostgreSQL database and runs the backend application

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DB_CONTAINER_NAME="finops-postgres"
DB_PORT="5432"
DB_NAME="finops"
DB_USER="finops"
DB_PASSWORD="finops"
API_PORT="8080"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  FinOps Aggregator Development Setup  ${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        echo -e "${RED}Error: Docker is not running${NC}"
        echo "Please start Docker and try again"
        exit 1
    fi
    echo -e "${GREEN}✓ Docker is running${NC}"
}

# Function to check if container exists
container_exists() {
    docker ps -a --format '{{.Names}}' | grep -q "^${DB_CONTAINER_NAME}$"
}

# Function to check if container is running
container_running() {
    docker ps --format '{{.Names}}' | grep -q "^${DB_CONTAINER_NAME}$"
}

# Function to start the database
start_database() {
    echo ""
    echo -e "${YELLOW}Starting PostgreSQL database...${NC}"
    
    if container_exists; then
        if container_running; then
            echo -e "${GREEN}✓ Database container is already running${NC}"
        else
            echo "Starting existing container..."
            docker start ${DB_CONTAINER_NAME}
            echo -e "${GREEN}✓ Database container started${NC}"
        fi
    else
        echo "Creating new database container..."
        docker run -d \
            --name ${DB_CONTAINER_NAME} \
            -e POSTGRES_DB=${DB_NAME} \
            -e POSTGRES_USER=${DB_USER} \
            -e POSTGRES_PASSWORD=${DB_PASSWORD} \
            -p ${DB_PORT}:5432 \
            postgres:15-alpine
        
        echo -e "${GREEN}✓ Database container created and started${NC}"
        echo "Waiting for database to be ready..."
        sleep 3
    fi
    
    # Wait for database to be ready
    echo "Checking database connection..."
    for i in {1..30}; do
        if docker exec ${DB_CONTAINER_NAME} pg_isready -U ${DB_USER} > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Database is ready${NC}"
            return 0
        fi
        echo -n "."
        sleep 1
    done
    
    echo -e "${RED}Error: Database failed to start${NC}"
    exit 1
}

# Function to run migrations
run_migrations() {
    echo ""
    echo -e "${YELLOW}Running database migrations...${NC}"
    
    if [ -f "migrations/schema.sql" ]; then
        docker exec -i ${DB_CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} < migrations/schema.sql > /dev/null 2>&1
        echo -e "${GREEN}✓ Migrations completed${NC}"
    else
        echo -e "${YELLOW}⚠ No migrations found (migrations/schema.sql)${NC}"
    fi
}

# Function to check if demo data should be seeded
check_demo_data() {
    echo ""
    echo -e "${YELLOW}Checking for demo data...${NC}"
    
    # Check if cost_nodes table has data
    NODE_COUNT=$(docker exec ${DB_CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -t -c "SELECT COUNT(*) FROM cost_nodes;" 2>/dev/null | tr -d ' ' || echo "0")
    
    if [ "$NODE_COUNT" -eq "0" ]; then
        echo -e "${YELLOW}No demo data found. Would you like to seed demo data? (y/n)${NC}"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            echo "Seeding demo data..."
            go run ./cmd/finops demo seed
            echo ""
            echo "Running allocation..."
            go run ./cmd/finops allocate --from 2025-09-10 --to 2025-10-10
            echo -e "${GREEN}✓ Demo data seeded${NC}"
        fi
    else
        echo -e "${GREEN}✓ Database has data ($NODE_COUNT nodes)${NC}"
    fi
}

# Function to start the API server
start_api() {
    echo ""
    echo -e "${YELLOW}Starting API server...${NC}"
    echo -e "${BLUE}API will be available at: http://localhost:${API_PORT}${NC}"
    echo -e "${BLUE}Press Ctrl+C to stop${NC}"
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo ""
    
    # Run the API server
    go run ./cmd/finops api
}

# Function to cleanup on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}Shutting down...${NC}"
    echo -e "${GREEN}Database container is still running. Use 'docker stop ${DB_CONTAINER_NAME}' to stop it.${NC}"
    exit 0
}

# Trap Ctrl+C
trap cleanup INT TERM

# Main execution
main() {
    check_docker
    start_database
    run_migrations
    check_demo_data
    start_api
}

# Run main function
main

