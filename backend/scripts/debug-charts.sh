#!/bin/bash

# Debug Chart Generation Script
# Shows exactly what commands are being run and their output

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

FINOPS_BIN="./bin/finops"
CHARTS_DIR="./debug-charts"
FORMAT="png"

echo -e "${BLUE}🐛 Debug Chart Generation${NC}"
echo "========================="

# Create debug directory
mkdir -p "$CHARTS_DIR"

# Check binary
if [[ ! -f "$FINOPS_BIN" ]]; then
    echo -e "${RED}❌ Binary not found: $FINOPS_BIN${NC}"
    echo "Run 'make build' first"
    exit 1
fi

echo -e "${BLUE}📋 Available commands:${NC}"
echo "Main help:"
$FINOPS_BIN --help | head -20

echo ""
echo "Export help:"
$FINOPS_BIN export --help

echo ""
echo "Chart help:"
$FINOPS_BIN export chart --help

echo ""
echo "Graph help:"
$FINOPS_BIN export chart graph --help

echo ""
echo -e "${BLUE}🧪 Testing Commands${NC}"
echo "==================="

# Test 1: Graph structure
echo -e "${BLUE}Test 1: Graph structure chart${NC}"
cmd="$FINOPS_BIN export chart graph --format $FORMAT --out $CHARTS_DIR/debug-graph.$FORMAT"
echo -e "${YELLOW}Command: $cmd${NC}"

if $cmd; then
    echo -e "${GREEN}✅ Success${NC}"
    if [[ -f "$CHARTS_DIR/debug-graph.$FORMAT" ]]; then
        size=$(stat -f%z "$CHARTS_DIR/debug-graph.$FORMAT" 2>/dev/null || stat -c%s "$CHARTS_DIR/debug-graph.$FORMAT" 2>/dev/null || echo "0")
        echo -e "${GREEN}File created: $size bytes${NC}"
    else
        echo -e "${YELLOW}⚠️  Command succeeded but no file found${NC}"
    fi
else
    echo -e "${RED}❌ Failed${NC}"
fi

echo ""

# Test 2: Load demo data first
echo -e "${BLUE}Test 2: Loading demo data${NC}"
cmd="$FINOPS_BIN demo seed"
echo -e "${YELLOW}Command: $cmd${NC}"

if $cmd; then
    echo -e "${GREEN}✅ Demo data loaded${NC}"
else
    echo -e "${RED}❌ Demo data loading failed${NC}"
fi

echo ""

# Test 3: Graph structure with data
echo -e "${BLUE}Test 3: Graph structure with data${NC}"
cmd="$FINOPS_BIN export chart graph --format $FORMAT --out $CHARTS_DIR/debug-graph-with-data.$FORMAT"
echo -e "${YELLOW}Command: $cmd${NC}"

if $cmd; then
    echo -e "${GREEN}✅ Success${NC}"
    if [[ -f "$CHARTS_DIR/debug-graph-with-data.$FORMAT" ]]; then
        size=$(stat -f%z "$CHARTS_DIR/debug-graph-with-data.$FORMAT" 2>/dev/null || stat -c%s "$CHARTS_DIR/debug-graph-with-data.$FORMAT" 2>/dev/null || echo "0")
        echo -e "${GREEN}File created: $size bytes${NC}"
    fi
else
    echo -e "${RED}❌ Failed${NC}"
fi

echo ""

# Test 4: Trend chart
echo -e "${BLUE}Test 4: Trend chart${NC}"
cmd="$FINOPS_BIN export chart trend --node product_p --dimension instance_hours --from 2024-01-01 --to 2024-01-31 --format $FORMAT --out $CHARTS_DIR/debug-trend.$FORMAT"
echo -e "${YELLOW}Command: $cmd${NC}"

if $cmd; then
    echo -e "${GREEN}✅ Success${NC}"
    if [[ -f "$CHARTS_DIR/debug-trend.$FORMAT" ]]; then
        size=$(stat -f%z "$CHARTS_DIR/debug-trend.$FORMAT" 2>/dev/null || stat -c%s "$CHARTS_DIR/debug-trend.$FORMAT" 2>/dev/null || echo "0")
        echo -e "${GREEN}File created: $size bytes${NC}"
    fi
else
    echo -e "${RED}❌ Failed${NC}"
fi

echo ""

# Test 5: Try with verbose output
echo -e "${BLUE}Test 5: Verbose graph generation${NC}"
cmd="$FINOPS_BIN export chart graph --format $FORMAT --out $CHARTS_DIR/debug-verbose.$FORMAT"
echo -e "${YELLOW}Command: $cmd${NC}"
echo -e "${BLUE}Full output:${NC}"

$cmd 2>&1 || echo -e "${RED}Command failed${NC}"

echo ""
echo -e "${BLUE}📁 Generated files:${NC}"
if [[ -d "$CHARTS_DIR" ]]; then
    find "$CHARTS_DIR" -type f | while read -r file; do
        size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")
        echo "  $file ($size bytes)"
    done
else
    echo "  No files generated"
fi

echo ""
echo -e "${BLUE}🔍 Configuration check:${NC}"
echo "Config file exists: $(test -f config.yaml && echo "✅ Yes" || echo "❌ No")"
echo "Charts directory: $CHARTS_DIR"
echo "Format: $FORMAT"

if [[ -f config.yaml ]]; then
    echo ""
    echo -e "${BLUE}Config content:${NC}"
    cat config.yaml
fi

echo ""
echo -e "${GREEN}🎉 Debug complete!${NC}"
echo ""
echo -e "${BLUE}💡 Next steps:${NC}"
echo "1. Check if the binary was built correctly: make build"
echo "2. Verify database connection: make demo-validate"
echo "3. Try the working commands manually"
echo "4. Check logs for more detailed error messages"
