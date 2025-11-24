# FinOps Aggregator

A comprehensive FinOps cost attribution and analysis platform with DAG-based cost allocation, product hierarchy visualization, and cost optimization recommendations.

## Features

- 🎯 **DAG-Based Cost Allocation** - Flexible cost attribution through directed acyclic graphs
- 📊 **Product Hierarchy View** - Visualize cost flows from resources to products
- 💰 **Cost Recommendations** - AI-powered suggestions for cost optimization
- 🔄 **Multi-Strategy Allocation** - Support for proportional, weighted, and custom allocation strategies
- 📈 **Trend Analysis** - Track cost trends over time with interactive charts
- 🌐 **Multi-Currency Support** - Handle costs in multiple currencies
- 🔍 **Detailed Cost Breakdown** - Drill down into cost dimensions and allocations

## Architecture

- **Backend**: Go with Gin framework, PostgreSQL database
- **Frontend**: Next.js with TypeScript, React, Tailwind CSS
- **Database**: PostgreSQL with pgx driver
- **Visualization**: Recharts for interactive charts

## Quick Start

### Option 1: Docker Compose (Recommended)

The easiest way to get started is using Docker Compose with hot reload:

```bash
# Start the development environment
./dev.sh start

# Or use make
make dev-up
```

This will start:
- Frontend at http://localhost:3000
- Backend API at http://localhost:8080
- PostgreSQL at localhost:5432
- Delve debugger at localhost:2345

### Option 2: Manual Setup

#### Prerequisites

- Go 1.24+
- Node.js 20+
- PostgreSQL 14+

#### Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Copy config
cp config.yaml.example config.yaml

# Edit config.yaml with your database connection

# Run migrations
make migrate

# Seed demo data
make demo-seed

# Start API server
make api
```

#### Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

## Development Environment

### Docker Compose Development

The Docker Compose development environment provides:

✅ **Hot Reload** - Automatic reload on code changes (backend & frontend)
✅ **Remote Debugging** - Debug Go code with Delve
✅ **Volume Caching** - Fast rebuilds with Go module and build caching
✅ **Isolated Network** - All services on dedicated Docker network

See [DOCKER_DEV_GUIDE.md](./DOCKER_DEV_GUIDE.md) for detailed documentation.

### Quick Commands

```bash
# Start environment
make dev-up

# View logs
make dev-logs

# Seed demo data
make dev-seed

# Connect to database
make dev-psql

# Stop environment
make dev-down

# Clean everything
make dev-clean

# See all commands
make help
```

### Remote Debugging

1. Start the development environment:
   ```bash
   make dev-up
   ```

2. In VS Code:
   - Set breakpoints in your Go code
   - Press F5 or Run > Start Debugging
   - Select "Connect to Backend Container"

3. The debugger will attach to the running container!

See [DOCKER_DEV_GUIDE.md](./DOCKER_DEV_GUIDE.md#remote-debugging) for more details.

## Project Structure

```
.
├── backend/                 # Go backend
│   ├── cmd/                # CLI commands
│   ├── internal/           # Internal packages
│   │   ├── api/           # REST API handlers
│   │   ├── allocate/      # Cost allocation engine
│   │   ├── analyzer/      # Recommendation analyzer
│   │   ├── models/        # Data models
│   │   └── store/         # Database layer
│   ├── migrations/        # Database migrations
│   └── config.yaml        # Configuration
├── frontend/               # Next.js frontend
│   ├── src/
│   │   ├── components/    # React components
│   │   ├── pages/         # Next.js pages
│   │   ├── types/         # TypeScript types
│   │   └── lib/           # Utilities
│   └── package.json
├── docker-compose.dev.yml  # Development environment
├── Makefile               # Development commands
└── dev.sh                 # Quick start script
```

## API Documentation

### Core Endpoints

- `GET /api/v1/products/hierarchy` - Product hierarchy with cost allocations
- `GET /api/v1/platform/services` - Platform and shared services
- `GET /api/v1/recommendations` - Cost optimization recommendations
- `GET /api/v1/nodes` - All cost nodes
- `GET /api/v1/allocations` - Cost allocation details

See [api-specification.yaml](./api-specification.yaml) for full API documentation.

## Key Concepts

### Node Types

- **Product** - Business products that consume resources
- **Resource** - Direct cloud resources (EC2, RDS, S3, etc.)
- **Platform** - Shared platform services (Kubernetes, API Gateway)
- **Shared** - Shared services (databases, caches, queues)

### Allocation Strategies

- **Proportional** - Allocate based on usage metrics
- **Weighted** - Allocate based on predefined weights
- **Equal** - Split costs equally
- **Custom** - Custom allocation logic

### Cost Dimensions

- `instance_hours` - Compute instance hours
- `storage_gb_month` - Storage in GB-months
- `egress_gb` - Network egress in GB
- `iops` - I/O operations per second
- `backups_gb_month` - Backup storage

## Configuration

### Backend Configuration

Edit `backend/config.yaml`:

```yaml
postgres:
  dsn: postgresql://user:pass@localhost:5432/finops

compute:
  base_currency: USD
  active_dimensions:
    - instance_hours
    - storage_gb_month
    - egress_gb

logging:
  level: info
```

### Frontend Configuration

The frontend automatically connects to the backend API. For custom configuration, create `frontend/.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Testing

### Backend Tests

```bash
# Run all tests
make dev-test

# Run with coverage
make dev-test-coverage

# Run specific package
cd backend
go test ./internal/allocate/...
```

### Frontend Tests

```bash
cd frontend
npm run test
```

## Database Management

### Migrations

```bash
# Run migrations
make dev-migrate

# Create new migration
cd backend
migrate create -ext sql -dir migrations -seq migration_name
```

### Seed Data

```bash
# Seed demo data
make dev-seed

# Reset database
make dev-clean
make dev-up
make dev-seed
```

### Access Database

```bash
# Using make
make dev-psql

# Or directly
PGPASSWORD=finops psql -h localhost -U finops -d finops
```

## Troubleshooting

### Backend Not Reloading

```bash
# Check Air is running
make dev-logs-backend | grep -i air

# Restart backend
make dev-restart-backend
```

### Frontend Not Reloading

```bash
# Clear Next.js cache
docker-compose -f docker-compose.dev.yml exec frontend rm -rf .next
make dev-restart-frontend
```

### Port Already in Use

```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Database Connection Issues

```bash
# Check postgres is healthy
make dev-status

# Restart postgres
docker-compose -f docker-compose.dev.yml restart postgres
```

See [DOCKER_DEV_GUIDE.md](./DOCKER_DEV_GUIDE.md#troubleshooting) for more troubleshooting tips.

## Documentation

- [Docker Development Guide](./DOCKER_DEV_GUIDE.md) - Comprehensive Docker development setup
- [Backend README](./backend/README.md) - Backend architecture and API
- [Frontend README](./frontend/README.md) - Frontend components and pages
- [API Specification](./api-specification.yaml) - OpenAPI specification
- [Quick Start Guide](./QUICK_START_GUIDE.md) - Getting started guide

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make dev-test`
5. Submit a pull request

## License

Proprietary - PickeringTech

## Support

For issues and questions:
- Check the [Docker Development Guide](./DOCKER_DEV_GUIDE.md)
- Review existing documentation
- Contact the development team

