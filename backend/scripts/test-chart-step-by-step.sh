#!/bin/bash

# Step-by-step chart generation test
# Tests each component individually to isolate issues

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

FINOPS_BIN="./bin/finops"

echo -e "${BLUE}🧪 Step-by-Step Chart Generation Test${NC}"
echo "====================================="

# Step 1: Check if binary exists
echo -e "${BLUE}Step 1: Check binary${NC}"
if [[ ! -f "$FINOPS_BIN" ]]; then
    echo -e "${YELLOW}⚠️  Binary not found, attempting to build...${NC}"
    if command -v go >/dev/null 2>&1; then
        echo "Building with Go..."
        if go build -o bin/finops ./cmd/finops; then
            echo -e "${GREEN}✅ Build successful${NC}"
        else
            echo -e "${RED}❌ Build failed${NC}"
            exit 1
        fi
    else
        echo -e "${RED}❌ Go not available and binary not found${NC}"
        echo "Please build the binary manually: make build"
        exit 1
    fi
else
    echo -e "${GREEN}✅ Binary found${NC}"
fi

# Step 2: Test basic CLI
echo -e "${BLUE}Step 2: Test basic CLI${NC}"
if $FINOPS_BIN --help >/dev/null 2>&1; then
    echo -e "${GREEN}✅ CLI works${NC}"
else
    echo -e "${RED}❌ CLI failed${NC}"
    exit 1
fi

# Step 3: Test export command
echo -e "${BLUE}Step 3: Test export command${NC}"
if $FINOPS_BIN export --help >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Export command works${NC}"
else
    echo -e "${RED}❌ Export command failed${NC}"
    exit 1
fi

# Step 4: Test chart command
echo -e "${BLUE}Step 4: Test chart command${NC}"
if $FINOPS_BIN export chart --help >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Chart command works${NC}"
else
    echo -e "${RED}❌ Chart command failed${NC}"
    exit 1
fi

# Step 5: Test graph subcommand help
echo -e "${BLUE}Step 5: Test graph subcommand help${NC}"
if $FINOPS_BIN export chart graph --help >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Graph subcommand help works${NC}"
    
    # Show the help to see available flags
    echo -e "${BLUE}Available flags:${NC}"
    $FINOPS_BIN export chart graph --help | grep -E "^\s*--" || echo "No flags found"
else
    echo -e "${RED}❌ Graph subcommand help failed${NC}"
    exit 1
fi

# Step 6: Test database connection (if available)
echo -e "${BLUE}Step 6: Test database connection${NC}"
if $FINOPS_BIN demo validate >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Database connection works${NC}"
    HAS_DB=true
else
    echo -e "${YELLOW}⚠️  Database connection failed (expected if not configured)${NC}"
    HAS_DB=false
fi

# Step 7: Test chart generation without database
echo -e "${BLUE}Step 7: Test chart generation (no database)${NC}"
mkdir -p test-output

# Try to generate a chart - this should create a "no data" chart
echo "Running: $FINOPS_BIN export chart graph --format png --out test-output/test-no-db.png"
if $FINOPS_BIN export chart graph --format png --out test-output/test-no-db.png 2>&1; then
    if [[ -f "test-output/test-no-db.png" ]]; then
        size=$(stat -f%z "test-output/test-no-db.png" 2>/dev/null || stat -c%s "test-output/test-no-db.png" 2>/dev/null || echo "0")
        if [[ "$size" -gt 1000 ]]; then
            echo -e "${GREEN}✅ Chart generated successfully (${size} bytes)${NC}"
        else
            echo -e "${YELLOW}⚠️  Chart file is very small (${size} bytes)${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  Command succeeded but no file created${NC}"
    fi
else
    echo -e "${RED}❌ Chart generation failed${NC}"
    echo "Let's try with verbose output:"
    $FINOPS_BIN export chart graph --format png --out test-output/test-no-db-verbose.png 2>&1 || true
fi

# Step 8: Test with demo data (if database works)
if [[ "$HAS_DB" == "true" ]]; then
    echo -e "${BLUE}Step 8: Test with demo data${NC}"
    
    echo "Loading demo data..."
    if $FINOPS_BIN demo seed >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Demo data loaded${NC}"
        
        echo "Generating chart with data..."
        if $FINOPS_BIN export chart graph --format png --out test-output/test-with-data.png 2>&1; then
            if [[ -f "test-output/test-with-data.png" ]]; then
                size=$(stat -f%z "test-output/test-with-data.png" 2>/dev/null || stat -c%s "test-output/test-with-data.png" 2>/dev/null || echo "0")
                echo -e "${GREEN}✅ Chart with data generated (${size} bytes)${NC}"
            else
                echo -e "${YELLOW}⚠️  Command succeeded but no file created${NC}"
            fi
        else
            echo -e "${RED}❌ Chart generation with data failed${NC}"
        fi
        
        # Test trend chart
        echo "Testing trend chart..."
        if $FINOPS_BIN export chart trend --node product_p --dimension instance_hours --from 2024-01-01 --to 2024-01-31 --format png --out test-output/test-trend.png 2>&1; then
            if [[ -f "test-output/test-trend.png" ]]; then
                size=$(stat -f%z "test-output/test-trend.png" 2>/dev/null || stat -c%s "test-output/test-trend.png" 2>/dev/null || echo "0")
                echo -e "${GREEN}✅ Trend chart generated (${size} bytes)${NC}"
            else
                echo -e "${YELLOW}⚠️  Trend command succeeded but no file created${NC}"
            fi
        else
            echo -e "${YELLOW}⚠️  Trend chart generation failed${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  Demo data loading failed${NC}"
    fi
else
    echo -e "${BLUE}Step 8: Skipped (no database)${NC}"
fi

# Step 9: Show results
echo -e "${BLUE}Step 9: Results${NC}"
if [[ -d "test-output" ]]; then
    echo "Generated files:"
    find test-output -type f | while read -r file; do
        size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")
        echo "  $file ($size bytes)"
    done
else
    echo "No output directory created"
fi

echo ""
echo -e "${GREEN}🎉 Step-by-step test complete!${NC}"
echo ""
echo -e "${BLUE}💡 Next steps:${NC}"
echo "1. Check the generated files in test-output/"
echo "2. If files are very small, there might be an issue with chart rendering"
echo "3. If no files are generated, check the CLI flag configuration"
echo "4. Try running individual commands manually to see detailed error messages"
