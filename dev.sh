#!/bin/bash

# FinOps Aggregator - Development Environment Quick Start Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    print_success "Docker is running"
}

# Function to check if ports are available
check_ports() {
    local ports=(3000 8080 5432 2345)
    local unavailable_ports=()

    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            unavailable_ports+=($port)
        fi
    done

    if [ ${#unavailable_ports[@]} -ne 0 ]; then
        print_warning "The following ports are already in use: ${unavailable_ports[*]}"
        print_info "You may need to stop other services or modify docker-compose.dev.yml"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        print_success "All required ports are available"
    fi
}

# Function to start the development environment
start_dev() {
    print_info "Starting FinOps Aggregator development environment..."
    echo ""

    check_docker
    check_ports

    echo ""
    print_info "Building and starting containers..."
    docker-compose -f docker-compose.dev.yml up -d --build

    echo ""
    print_success "Development environment started!"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "  🌐 Frontend:  ${GREEN}http://localhost:3000${NC}"
    echo "  🔧 Backend:   ${GREEN}http://localhost:8080${NC}"
    echo "  🗄️  Postgres:  ${GREEN}localhost:5432${NC}"
    echo "  🐛 Debugger:  ${GREEN}localhost:2345${NC}"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    print_info "Useful commands:"
    echo "  make dev-logs          - View logs"
    echo "  make dev-seed          - Seed demo data and run allocation"
    echo "  make dev-psql          - Connect to database"
    echo "  make dev-down          - Stop environment"
    echo "  make help              - Show all commands"
    echo ""
    print_info "Waiting for services to be ready..."
    sleep 5

    # Check if services are healthy
    if docker-compose -f docker-compose.dev.yml ps | grep -q "Up"; then
        print_success "Services are running!"
        echo ""
        print_info "Would you like to seed the database with demo data and run allocation? (y/N)"
        read -p "" -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "Seeding database and running allocation..."
            sleep 3  # Give backend time to fully start
            docker-compose -f docker-compose.dev.yml exec backend sh -c "cd /app && make demo-seed demo-allocate" || true
            print_success "Database seeded and allocation completed!"
        fi
    else
        print_warning "Some services may not be running correctly"
        print_info "Check logs with: make dev-logs"
    fi
}

# Function to stop the development environment
stop_dev() {
    print_info "Stopping development environment..."
    docker-compose -f docker-compose.dev.yml down
    print_success "Development environment stopped!"
}

# Function to clean the development environment
clean_dev() {
    print_warning "This will remove all containers, volumes, and data!"
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Cleaning development environment..."
        docker-compose -f docker-compose.dev.yml down -v
        print_success "Development environment cleaned!"
    fi
}

# Function to perform full reset + seed + allocation
full_reset_and_seed() {
    print_warning "This will fully reset containers, volumes, and data, then reseed demo data and run allocation!"
    read -p "Are you sure you want to continue? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        return
    fi

    print_info "Performing full reset (clean + rebuild + up)..."
    make dev-reset

    print_info "Waiting for services to become healthy..."
    sleep 5

    print_info "Seeding database with demo data and running allocation..."
    make dev-seed
    print_success "Full reset, seed, and allocation completed!"
}


# Function to show logs
show_logs() {
    print_info "Showing logs (Ctrl+C to exit)..."
    docker-compose -f docker-compose.dev.yml logs -f
}

# Function to show status
show_status() {
    print_info "Service status:"
    docker-compose -f docker-compose.dev.yml ps
}

# Main menu
show_menu() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  FinOps Aggregator - Development Environment"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "  1) Start development environment"
    echo "  2) Stop development environment"
    echo "  3) Clean development environment"
    echo "  4) Show logs"
    echo "  5) Show status"
    echo "  6) Full reset + seed demo data + allocation"
    echo "  7) Exit"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
}

# Parse command line arguments
if [ $# -eq 0 ]; then
    # No arguments, show menu
    while true; do
        show_menu
        read -p "Select an option: " choice
        case $choice in
            1) start_dev ;;
            2) stop_dev ;;
            3) clean_dev ;;
            4) show_logs ;;
            5) show_status ;;
            6) full_reset_and_seed ;;
            7) exit 0 ;;
            *) print_error "Invalid option" ;;
        esac
    done
else
    # Parse arguments
    case "$1" in
        start|up)
            start_dev
            ;;
        stop|down)
            stop_dev
            ;;
        clean)
            clean_dev
            ;;
        logs)
            show_logs
            ;;
        status)
            show_status
            ;;
        reset-seed)
            full_reset_and_seed
            ;;
        *)
            echo "Usage: $0 {start|stop|clean|logs|status|reset-seed}"
            echo ""
            echo "Or run without arguments for interactive menu"
            exit 1
            ;;
    esac
fi

