#!/bin/bash

# Development Environment Reset Script
# This script completely resets the development environment

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

CONTAINER_NAME="finops-postgres"
VOLUME_NAME="finops_postgres_data"

echo -e "${BLUE}🔄 Resetting FinOps Development Environment${NC}"
echo "=============================================="

# Confirm reset
read -p "This will destroy all data and reset the environment. Continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Reset cancelled.${NC}"
    exit 0
fi

echo -e "${BLUE}🛑 Stopping and removing containers...${NC}"

# Stop and remove PostgreSQL container
if docker ps -a --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
    docker stop $CONTAINER_NAME 2>/dev/null || true
    docker rm $CONTAINER_NAME 2>/dev/null || true
    echo -e "${GREEN}✅ PostgreSQL container removed${NC}"
else
    echo -e "${YELLOW}⚠️  PostgreSQL container not found${NC}"
fi

# Remove Docker volume
if docker volume ls --format "table {{.Name}}" | grep -q "^${VOLUME_NAME}$"; then
    docker volume rm $VOLUME_NAME 2>/dev/null || true
    echo -e "${GREEN}✅ PostgreSQL data volume removed${NC}"
else
    echo -e "${YELLOW}⚠️  PostgreSQL data volume not found${NC}"
fi

# Clean up build artifacts
echo -e "${BLUE}🧹 Cleaning up build artifacts...${NC}"
rm -rf bin/
rm -f coverage.out coverage.html
echo -e "${GREEN}✅ Build artifacts cleaned${NC}"

# Remove config file (optional)
read -p "Remove config.yaml? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -f config.yaml
    echo -e "${GREEN}✅ Config file removed${NC}"
fi

# Clean Go module cache (optional)
if command -v go &> /dev/null; then
    read -p "Clean Go module cache? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        go clean -modcache
        echo -e "${GREEN}✅ Go module cache cleaned${NC}"
    fi
fi

echo ""
echo -e "${GREEN}🎉 Environment reset complete!${NC}"
echo "================================"
echo -e "${BLUE}To set up again:${NC}"
echo "  ./scripts/dev-setup.sh"
