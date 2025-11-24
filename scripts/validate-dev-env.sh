#!/bin/bash

# Script to validate the development environment setup

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  FinOps Aggregator - Development Environment Validation"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check Docker
print_info "Checking Docker..."
if command -v docker &> /dev/null; then
    if docker info &> /dev/null; then
        print_success "Docker is installed and running"
        docker --version
    else
        print_error "Docker is installed but not running"
        exit 1
    fi
else
    print_error "Docker is not installed"
    exit 1
fi

echo ""

# Check Docker Compose
print_info "Checking Docker Compose..."
if docker-compose --version &> /dev/null; then
    print_success "Docker Compose is installed"
    docker-compose --version
else
    print_error "Docker Compose is not installed"
    exit 1
fi

echo ""

# Check required files
print_info "Checking required files..."
files=(
    "docker-compose.dev.yml"
    "backend/Dockerfile.dev"
    "backend/.air.toml"
    "backend/config.dev.yaml"
    "frontend/Dockerfile.dev"
    "Makefile"
    "dev.sh"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        print_success "$file exists"
    else
        print_error "$file is missing"
        exit 1
    fi
done

echo ""

# Check ports
print_info "Checking port availability..."
ports=(3000 8080 5432 2345)
all_available=true

for port in "${ports[@]}"; do
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        print_warning "Port $port is already in use"
        all_available=false
    else
        print_success "Port $port is available"
    fi
done

if [ "$all_available" = false ]; then
    print_warning "Some ports are in use. You may need to stop other services."
fi

echo ""

# Validate docker-compose file
print_info "Validating docker-compose.dev.yml..."
if docker-compose -f docker-compose.dev.yml config > /dev/null 2>&1; then
    print_success "docker-compose.dev.yml is valid"
else
    print_error "docker-compose.dev.yml has errors"
    exit 1
fi

echo ""

# Check if services are running
print_info "Checking if services are running..."
if docker-compose -f docker-compose.dev.yml ps | grep -q "Up"; then
    print_success "Development environment is running"
    echo ""
    docker-compose -f docker-compose.dev.yml ps
else
    print_warning "Development environment is not running"
    print_info "Start it with: make dev-up"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
print_success "Validation complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

