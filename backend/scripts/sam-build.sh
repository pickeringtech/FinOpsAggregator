#!/bin/bash
# Build Lambda functions using SAM CLI
set -e

cd "$(dirname "$0")/.."

echo "Building Lambda functions with SAM..."

# Build the Lambda binary
echo "Building Go binary for Lambda..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap ./cmd/lambda

# Run SAM build
echo "Running SAM build..."
sam build --use-container

echo "Build complete!"

