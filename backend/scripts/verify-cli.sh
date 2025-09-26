#!/bin/bash

# CLI Verification Script
# Verifies the CLI structure matches what the scripts expect

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

FINOPS_BIN="./bin/finops"

echo -e "${BLUE}🔍 Verifying FinOps CLI Structure${NC}"
echo "=================================="

# Check if binary exists
if [[ ! -f "$FINOPS_BIN" ]]; then
    echo -e "${YELLOW}⚠️  Binary not found, attempting to build...${NC}"
    if ! make build >/dev/null 2>&1; then
        echo -e "${RED}❌ Failed to build binary${NC}"
        echo "Please run 'make build' manually to build the application"
        exit 1
    fi
fi

echo -e "${BLUE}📋 Checking command structure...${NC}"

# Test main help
echo -e "${BLUE}Main help:${NC}"
if $FINOPS_BIN --help >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Main command works${NC}"
else
    echo -e "${RED}❌ Main command failed${NC}"
    exit 1
fi

# Test export command
echo -e "${BLUE}Export command:${NC}"
if $FINOPS_BIN export --help >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Export command works${NC}"
else
    echo -e "${RED}❌ Export command failed${NC}"
    exit 1
fi

# Test chart command
echo -e "${BLUE}Chart command:${NC}"
if $FINOPS_BIN export chart --help >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Chart command works${NC}"
else
    echo -e "${RED}❌ Chart command failed${NC}"
    exit 1
fi

# Test graph subcommand
echo -e "${BLUE}Graph subcommand:${NC}"
if $FINOPS_BIN export chart graph --help >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Graph subcommand works${NC}"
    
    # Check if format flag exists
    if $FINOPS_BIN export chart graph --help 2>&1 | grep -q "\-\-format"; then
        echo -e "${GREEN}✅ --format flag available${NC}"
    else
        echo -e "${RED}❌ --format flag missing${NC}"
    fi
    
    # Check if out flag exists
    if $FINOPS_BIN export chart graph --help 2>&1 | grep -q "\-\-out"; then
        echo -e "${GREEN}✅ --out flag available${NC}"
    else
        echo -e "${RED}❌ --out flag missing${NC}"
    fi
else
    echo -e "${RED}❌ Graph subcommand failed${NC}"
fi

# Test trend subcommand
echo -e "${BLUE}Trend subcommand:${NC}"
if $FINOPS_BIN export chart trend --help >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Trend subcommand works${NC}"
    
    # Check required flags
    local required_flags=("node" "from" "to" "format")
    for flag in "${required_flags[@]}"; do
        if $FINOPS_BIN export chart trend --help 2>&1 | grep -q "\-\-$flag"; then
            echo -e "${GREEN}✅ --$flag flag available${NC}"
        else
            echo -e "${RED}❌ --$flag flag missing${NC}"
        fi
    done
else
    echo -e "${RED}❌ Trend subcommand failed${NC}"
fi

# Test waterfall subcommand
echo -e "${BLUE}Waterfall subcommand:${NC}"
if $FINOPS_BIN export chart waterfall --help >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Waterfall subcommand works${NC}"
else
    echo -e "${RED}❌ Waterfall subcommand failed${NC}"
fi

echo ""
echo -e "${BLUE}📊 Testing actual chart generation...${NC}"

# Test graph generation (should work even without data)
echo -e "${BLUE}Testing graph generation:${NC}"
if $FINOPS_BIN export chart graph --format png --out /tmp/test-graph.png 2>/dev/null; then
    if [[ -f "/tmp/test-graph.png" ]]; then
        local size=$(stat -f%z "/tmp/test-graph.png" 2>/dev/null || stat -c%s "/tmp/test-graph.png" 2>/dev/null || echo "0")
        if [[ "$size" -gt 1000 ]]; then
            echo -e "${GREEN}✅ Graph generation works (${size} bytes)${NC}"
            rm -f /tmp/test-graph.png
        else
            echo -e "${YELLOW}⚠️  Graph generated but file is small (${size} bytes)${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  Command succeeded but no file created${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Graph generation failed (expected without data)${NC}"
fi

# Test with demo data
echo -e "${BLUE}Testing with demo data:${NC}"
if $FINOPS_BIN demo seed >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Demo data loaded${NC}"
    
    # Try graph generation again
    if $FINOPS_BIN export chart graph --format png --out /tmp/test-graph-with-data.png 2>/dev/null; then
        if [[ -f "/tmp/test-graph-with-data.png" ]]; then
            local size=$(stat -f%z "/tmp/test-graph-with-data.png" 2>/dev/null || stat -c%s "/tmp/test-graph-with-data.png" 2>/dev/null || echo "0")
            echo -e "${GREEN}✅ Graph generation with data works (${size} bytes)${NC}"
            rm -f /tmp/test-graph-with-data.png
        else
            echo -e "${YELLOW}⚠️  Command succeeded but no file created${NC}"
        fi
    else
        echo -e "${RED}❌ Graph generation with data failed${NC}"
    fi
    
    # Try trend generation
    if $FINOPS_BIN export chart trend --node product_p --dimension instance_hours --from 2024-01-01 --to 2024-01-31 --format png --out /tmp/test-trend.png 2>/dev/null; then
        if [[ -f "/tmp/test-trend.png" ]]; then
            local size=$(stat -f%z "/tmp/test-trend.png" 2>/dev/null || stat -c%s "/tmp/test-trend.png" 2>/dev/null || echo "0")
            echo -e "${GREEN}✅ Trend generation works (${size} bytes)${NC}"
            rm -f /tmp/test-trend.png
        else
            echo -e "${YELLOW}⚠️  Trend command succeeded but no file created${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  Trend generation failed (might be expected)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Demo data loading failed${NC}"
fi

echo ""
echo -e "${GREEN}🎉 CLI verification complete!${NC}"
echo ""
echo -e "${BLUE}💡 Usage examples:${NC}"
echo "  $FINOPS_BIN export chart graph --format png --out graph.png"
echo "  $FINOPS_BIN export chart trend --node product_p --dimension instance_hours --from 2024-01-01 --to 2024-01-31 --format png"
echo "  ./scripts/generate-charts.sh demo"
