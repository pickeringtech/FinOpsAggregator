# FinOps DAG Cost Attribution Tool

A dimension-aware FinOps aggregation tool that models cost attribution as a weighted directed acyclic graph (DAG) and provides both TUI and API interfaces for operational visibility.

## Features

- **DAG-based Cost Attribution**: Model cost relationships as a directed acyclic graph with weighted edges
- **Multi-dimensional Costs**: Support for multiple cost dimensions (instance_hours, storage_gb_month, egress_gb, etc.)
- **Flexible Allocation Strategies**: Multiple weighting strategies including proportional, equal, fixed_percent, capped_proportional, and residual_to_max
- **Terminal User Interface**: Interactive TUI for cost exploration and management
- **Background Jobs**: PostgreSQL-backed job system using River for reliable computation and export tasks
- **Chart Generation**: Automated generation of trend, waterfall, and attribution charts
- **Flexible Storage**: Support for local filesystem, S3, and GCS storage backends
- **Comprehensive CLI**: Full command-line interface for all operations

## Architecture

### Core Components

- **Graph Engine**: DAG operations, topological sorting, cycle detection
- **Allocation Engine**: Cost propagation with configurable weighting strategies
- **Data Store**: PostgreSQL-backed repositories with transaction support
- **Job System**: River-based background processing for computations and exports
- **Chart Generation**: PNG/SVG chart export using go-chart and gonum/plot
- **TUI**: Bubble Tea-based terminal interface
- **CLI**: Cobra-based command-line interface

### Database Schema

The system uses PostgreSQL with the following core tables:

- `cost_nodes`: Nodes in the cost attribution graph
- `dependency_edges`: Relationships between nodes with effective dating
- `edge_strategies`: Dimension-specific allocation strategy overrides
- `node_costs_by_dimension`: Direct costs per node/date/dimension
- `node_usage_by_dimension`: Usage metrics for allocation calculations
- `computation_runs`: Allocation computation metadata
- `allocation_results_by_dimension`: Computed allocation results
- `contribution_results_by_dimension`: Detailed contribution tracking

## Getting Started

### Prerequisites

- Go 1.22 or later
- PostgreSQL 14 or later
- (Optional) Docker for containerized deployment

### Installation

#### Option 1: Local Development

1. Clone the repository:
```bash
git clone https://github.com/pickeringtech/FinOpsAggregator.git
cd FinOpsAggregator/backend
```

2. Install dependencies:
```bash
make deps
```

3. Set up PostgreSQL database:
```bash
make dev-db-start
make migrate-up
```

4. Copy and configure the config file:
```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your database connection and preferences
```

#### Option 2: Docker Compose

1. Clone the repository:
```bash
git clone https://github.com/pickeringtech/FinOpsAggregator.git
cd FinOpsAggregator/backend
```

2. Start all services:
```bash
docker-compose up -d
```

This will start PostgreSQL, run migrations, and build the application.

### Basic Usage

#### Quick Start with Demo Data

1. Build the application:
```bash
make build
```

2. Load demo seed data:
```bash
make demo-seed
```

3. Validate the graph structure:
```bash
make demo-validate
```

4. Run cost allocation:
```bash
make demo-allocate
```

Or run all demo steps at once:
```bash
make demo-full
```

#### Manual Commands

#### Graph Operations

Validate the cost attribution graph:
```bash
./bin/finops graph validate
```

#### Run Allocations

Execute cost allocation for a date range:
```bash
./bin/finops allocate --from 2024-01-01 --to 2024-01-31
```

#### Demo Data

Load demo seed data:
```bash
./bin/finops demo seed
```

Generate synthetic data for testing (not yet implemented):
```bash
./bin/finops demo synth --nodes 1000 --edges 3000 --days 30 --dimensions 6
```

#### Import Data (not yet implemented)

Import cost data from CSV:
```bash
./bin/finops import costs ./data/costs.csv
```

Import usage data from CSV:
```bash
./bin/finops import usage ./data/usage.csv
```

#### Export Charts (not yet implemented)

Generate trend charts:
```bash
./bin/finops export chart trend --node my-product --out ./charts/my-product-trend.png
```

#### Launch TUI (not yet implemented)

Start the interactive terminal interface:
```bash
./bin/finops tui
```

## Configuration

The application uses YAML configuration with environment variable overrides. Key configuration sections:

- `postgres`: Database connection settings
- `compute`: Computation parameters (base currency, active dimensions)
- `charts`: Chart generation settings
- `storage`: Storage backend configuration (file://, s3://, gs://)
- `jobs`: Background job system settings
- `logging`: Logging configuration

Environment variables use the `FINOPS_` prefix with underscores replacing dots (e.g., `FINOPS_POSTGRES_DSN`).

## Development Status

This is currently a work-in-progress implementation. Completed components:

- [x] Database schema and migrations
- [x] Core data models and types
- [x] Configuration management
- [x] Logging infrastructure
- [x] Database connection and repository layer
- [x] Node, edge, cost, usage, and run repositories
- [x] Graph operations and validation
- [x] Allocation engine core
- [x] Basic weighting strategies (equal, proportional, fixed_percent, etc.)
- [x] Demo data seeding system
- [x] CLI command structure with working commands
- [x] Docker and docker-compose setup
- [x] Makefile for development workflow

In progress:
- [ ] Advanced weighting strategies (capped_proportional, residual_to_max)
- [ ] Job system integration with River
- [ ] TUI implementation
- [ ] Chart generation
- [ ] Data import/export (CSV)
- [ ] Comprehensive testing
- [ ] Performance optimization

## Contributing

This project follows standard Go conventions. Key guidelines:

- Use `go fmt` for code formatting
- Write tests for all new functionality
- Follow the repository pattern for data access
- Use structured logging with zerolog
- Maintain database transaction safety

## License

[License details to be added]
