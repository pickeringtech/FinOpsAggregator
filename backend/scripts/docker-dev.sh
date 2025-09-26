#!/bin/bash

# Docker Development Environment Script
# Manages the complete Docker-based development environment

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Determine docker-compose command
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE_CMD="docker compose"
else
    echo -e "${RED}❌ Docker Compose is not available${NC}"
    exit 1
fi

# Function to show usage
show_usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  up          Start all services (PostgreSQL + FinOps app)"
    echo "  down        Stop all services"
    echo "  restart     Restart all services"
    echo "  logs        Show logs from all services"
    echo "  logs-db     Show PostgreSQL logs"
    echo "  logs-app    Show FinOps application logs"
    echo "  shell       Open shell in FinOps container"
    echo "  db-shell    Open PostgreSQL shell"
    echo "  build       Build Docker images"
    echo "  rebuild     Rebuild images from scratch"
    echo "  status      Show service status"
    echo "  clean       Remove all containers and volumes"
    echo "  migrate     Run database migrations"
    echo "  seed        Load demo seed data"
    echo "  demo        Run full demo (seed + validate + allocate)"
    echo ""
    echo "Examples:"
    echo "  $0 up                    # Start development environment"
    echo "  $0 logs-app              # View application logs"
    echo "  $0 shell                 # Open shell in app container"
    echo "  $0 db-shell              # Connect to PostgreSQL"
}

# Function to check if services are running
check_services() {
    if $DOCKER_COMPOSE_CMD ps | grep -q "Up"; then
        return 0
    else
        return 1
    fi
}

# Function to wait for PostgreSQL to be ready
wait_for_postgres() {
    echo -e "${BLUE}⏳ Waiting for PostgreSQL to be ready...${NC}"
    for i in {1..30}; do
        if $DOCKER_COMPOSE_CMD exec -T postgres pg_isready -U finops -d finops &>/dev/null; then
            echo -e "${GREEN}✅ PostgreSQL is ready!${NC}"
            return 0
        fi
        sleep 1
    done
    echo -e "${RED}❌ PostgreSQL failed to start within 30 seconds${NC}"
    return 1
}

# Main command handling
case "${1:-}" in
    "up")
        echo -e "${BLUE}🚀 Starting Docker development environment...${NC}"
        $DOCKER_COMPOSE_CMD up -d
        wait_for_postgres
        echo -e "${GREEN}✅ Development environment is running!${NC}"
        echo ""
        echo -e "${BLUE}Services:${NC}"
        $DOCKER_COMPOSE_CMD ps
        echo ""
        echo -e "${BLUE}Next steps:${NC}"
        echo "  View logs:     $0 logs"
        echo "  Run demo:      $0 demo"
        echo "  Open shell:    $0 shell"
        ;;
    
    "down")
        echo -e "${BLUE}🛑 Stopping Docker development environment...${NC}"
        $DOCKER_COMPOSE_CMD down
        echo -e "${GREEN}✅ Environment stopped${NC}"
        ;;
    
    "restart")
        echo -e "${BLUE}🔄 Restarting Docker development environment...${NC}"
        $DOCKER_COMPOSE_CMD restart
        wait_for_postgres
        echo -e "${GREEN}✅ Environment restarted${NC}"
        ;;
    
    "logs")
        echo -e "${BLUE}📋 Showing logs from all services...${NC}"
        $DOCKER_COMPOSE_CMD logs -f
        ;;
    
    "logs-db")
        echo -e "${BLUE}📋 Showing PostgreSQL logs...${NC}"
        $DOCKER_COMPOSE_CMD logs -f postgres
        ;;
    
    "logs-app")
        echo -e "${BLUE}📋 Showing FinOps application logs...${NC}"
        $DOCKER_COMPOSE_CMD logs -f finops
        ;;
    
    "shell")
        echo -e "${BLUE}🐚 Opening shell in FinOps container...${NC}"
        if check_services; then
            $DOCKER_COMPOSE_CMD exec finops /bin/sh
        else
            echo -e "${RED}❌ Services are not running. Start with: $0 up${NC}"
            exit 1
        fi
        ;;
    
    "db-shell")
        echo -e "${BLUE}🐘 Opening PostgreSQL shell...${NC}"
        if check_services; then
            $DOCKER_COMPOSE_CMD exec postgres psql -U finops -d finops
        else
            echo -e "${RED}❌ Services are not running. Start with: $0 up${NC}"
            exit 1
        fi
        ;;
    
    "build")
        echo -e "${BLUE}🔨 Building Docker images...${NC}"
        $DOCKER_COMPOSE_CMD build
        echo -e "${GREEN}✅ Images built successfully${NC}"
        ;;
    
    "rebuild")
        echo -e "${BLUE}🔨 Rebuilding Docker images from scratch...${NC}"
        $DOCKER_COMPOSE_CMD build --no-cache
        echo -e "${GREEN}✅ Images rebuilt successfully${NC}"
        ;;
    
    "status")
        echo -e "${BLUE}📊 Service status:${NC}"
        $DOCKER_COMPOSE_CMD ps
        ;;
    
    "clean")
        echo -e "${YELLOW}⚠️  This will remove all containers and volumes!${NC}"
        read -p "Continue? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${BLUE}🧹 Cleaning up Docker environment...${NC}"
            $DOCKER_COMPOSE_CMD down -v --remove-orphans
            docker system prune -f
            echo -e "${GREEN}✅ Environment cleaned${NC}"
        else
            echo -e "${YELLOW}Clean cancelled${NC}"
        fi
        ;;
    
    "migrate")
        echo -e "${BLUE}🗄️  Running database migrations...${NC}"
        if check_services; then
            $DOCKER_COMPOSE_CMD run --rm migrate
            echo -e "${GREEN}✅ Migrations completed${NC}"
        else
            echo -e "${RED}❌ Services are not running. Start with: $0 up${NC}"
            exit 1
        fi
        ;;
    
    "seed")
        echo -e "${BLUE}🌱 Loading demo seed data...${NC}"
        if check_services; then
            $DOCKER_COMPOSE_CMD exec finops /finops demo seed
            echo -e "${GREEN}✅ Demo data loaded${NC}"
        else
            echo -e "${RED}❌ Services are not running. Start with: $0 up${NC}"
            exit 1
        fi
        ;;
    
    "demo")
        echo -e "${BLUE}🎬 Running full demo...${NC}"
        if check_services; then
            echo -e "${BLUE}1. Loading seed data...${NC}"
            $DOCKER_COMPOSE_CMD exec finops /finops demo seed
            
            echo -e "${BLUE}2. Validating graph...${NC}"
            $DOCKER_COMPOSE_CMD exec finops /finops graph validate
            
            echo -e "${BLUE}3. Running allocation...${NC}"
            $DOCKER_COMPOSE_CMD exec finops /finops allocate --from 2024-01-01 --to 2024-01-31
            
            echo -e "${GREEN}✅ Demo completed successfully!${NC}"
        else
            echo -e "${RED}❌ Services are not running. Start with: $0 up${NC}"
            exit 1
        fi
        ;;
    
    "help"|"-h"|"--help"|"")
        show_usage
        ;;
    
    *)
        echo -e "${RED}❌ Unknown command: $1${NC}"
        echo ""
        show_usage
        exit 1
        ;;
esac
