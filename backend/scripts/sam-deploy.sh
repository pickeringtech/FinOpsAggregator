#!/bin/bash
# Deploy Lambda functions using SAM CLI
set -e

cd "$(dirname "$0")/.."

ENVIRONMENT=${1:-"dev"}
POSTGRES_DSN=${FINOPS_POSTGRES_DSN:-""}

if [ -z "$POSTGRES_DSN" ]; then
    echo "Error: FINOPS_POSTGRES_DSN environment variable must be set"
    echo "Usage: FINOPS_POSTGRES_DSN='postgresql://...' ./scripts/sam-deploy.sh [dev|staging|prod]"
    exit 1
fi

echo "Deploying to environment: $ENVIRONMENT"

# Build first
./scripts/sam-build.sh

# Deploy with the appropriate config
sam deploy \
    --config-env "$ENVIRONMENT" \
    --parameter-overrides "PostgresDSN=$POSTGRES_DSN Environment=$ENVIRONMENT"

echo "Deployment complete!"

