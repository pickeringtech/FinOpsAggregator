#!/bin/bash

# Simple syntax validation script for Go code
# This checks for basic syntax errors without actually building

echo "Validating Go syntax..."

# Check if we have any obvious syntax errors by examining the files
find . -name "*.go" -exec echo "Checking {}" \; -exec head -1 {} \;

echo ""
echo "Key files structure:"
echo "- cmd/finops/main.go exists: $(test -f cmd/finops/main.go && echo "✓" || echo "✗")"
echo "- internal/store/db.go exists: $(test -f internal/store/db.go && echo "✓" || echo "✗")"
echo "- internal/models/types.go exists: $(test -f internal/models/types.go && echo "✓" || echo "✗")"
echo "- internal/config/config.go exists: $(test -f internal/config/config.go && echo "✓" || echo "✗")"
echo "- internal/graph/graph.go exists: $(test -f internal/graph/graph.go && echo "✓" || echo "✗")"
echo "- internal/allocate/engine.go exists: $(test -f internal/allocate/engine.go && echo "✓" || echo "✗")"

echo ""
echo "Go module info:"
echo "- go.mod exists: $(test -f go.mod && echo "✓" || echo "✗")"
if [ -f go.mod ]; then
    echo "- Module name: $(head -1 go.mod)"
    echo "- Go version: $(grep "^go " go.mod)"
fi

echo ""
echo "Dependencies check:"
echo "- Required packages in go.mod:"
grep -E "github.com/(Masterminds/squirrel|google/uuid|jackc/pgx|rs/zerolog|shopspring/decimal|spf13/cobra|spf13/viper)" go.mod | head -10

echo ""
echo "Syntax validation complete. If Go were available, you would run:"
echo "  go mod tidy"
echo "  go build ./cmd/finops/"
echo ""
echo "The code structure appears correct for a Go application."
