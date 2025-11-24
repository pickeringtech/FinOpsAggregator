# Development Scripts

This directory contains scripts to help with development and operations of the FinOps Aggregator application.

## Available Scripts

### `dev` - Development Environment Startup

Starts the complete development environment including:
- PostgreSQL database (Docker container)
- Backend API server (Go)
- Frontend development server (Next.js)

#### Prerequisites

Before running the dev script, ensure you have the following installed:

- **Docker** - For running the PostgreSQL database
- **Go 1.22+** - For the backend API server
- **Node.js 18+** - For the frontend development server
- **npm** - Node package manager

#### Usage

From the project root directory:

```bash
./scripts/dev
```

Or from anywhere:

```bash
/path/to/FinOpsAggregator/scripts/dev
```

#### What It Does

1. **Checks Prerequisites**
   - Verifies Docker, Go, and Node.js are installed
   - Ensures Docker daemon is running

2. **Starts PostgreSQL Database**
   - Creates a Docker container named `finops-postgres` if it doesn't exist
   - Starts the container if it's stopped
   - Waits for the database to be ready
   - Uses credentials: `finops/finops` on port `5432`

3. **Runs Database Migrations**
   - Automatically applies all pending database migrations
   - Creates the necessary schema and tables

4. **Checks for Demo Data**
   - Prompts to load demo data if the database is empty
   - Useful for testing and development

5. **Starts Backend API Server**
   - Builds the Go application if needed
   - Creates `config.yaml` from example if it doesn't exist
   - Starts the API server on port `8080`

6. **Starts Frontend Development Server**
   - Installs npm dependencies if needed
   - Starts Next.js dev server on port `3000`

#### Services

Once running, you can access:

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Database**: postgresql://finops:finops@localhost:5432/finops

#### Stopping the Environment

Press `Ctrl+C` to gracefully shut down all services.

**Note**: The database container will continue running after shutdown. To stop it:

```bash
docker stop finops-postgres
```

To remove it completely:

```bash
docker rm finops-postgres
```

#### Troubleshooting

**Port Already in Use**

If you get errors about ports already in use:

- Port 5432: Another PostgreSQL instance is running
  ```bash
  docker stop finops-postgres
  # or stop your local PostgreSQL service
  ```

- Port 8080: Another backend instance is running
  ```bash
  # Find and kill the process using port 8080
  lsof -ti:8080 | xargs kill -9
  ```

- Port 3000: Another Next.js instance is running
  ```bash
  # Find and kill the process using port 3000
  lsof -ti:3000 | xargs kill -9
  ```

**Database Connection Issues**

If the backend can't connect to the database:

1. Check if the container is running:
   ```bash
   docker ps | grep finops-postgres
   ```

2. Check the logs:
   ```bash
   docker logs finops-postgres
   ```

3. Verify connectivity:
   ```bash
   docker exec finops-postgres pg_isready -U finops -d finops
   ```

**Migration Errors**

If migrations fail:

1. Install the migrate tool:
   ```bash
   # macOS
   brew install golang-migrate
   
   # Linux
   curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz
   sudo mv migrate /usr/local/bin/
   ```

2. Run migrations manually:
   ```bash
   cd backend
   make migrate-up
   ```

**Frontend Dependencies Issues**

If npm install fails:

1. Clear npm cache:
   ```bash
   cd frontend
   rm -rf node_modules package-lock.json
   npm cache clean --force
   npm install
   ```

## Manual Development

If you prefer to run services individually:

### Backend Only

```bash
cd backend
./dev.sh
```

### Frontend Only

```bash
cd frontend
npm run dev
```

### Database Only

```bash
docker run -d \
  --name finops-postgres \
  -e POSTGRES_DB=finops \
  -e POSTGRES_USER=finops \
  -e POSTGRES_PASSWORD=finops \
  -p 5432:5432 \
  postgres:15-alpine
```

## Additional Resources

- Backend documentation: `backend/README.md`
- Frontend documentation: `frontend/README.md`
- API specification: `api-specification.yaml`
- Quick start guide: `QUICK_START_GUIDE.md`

