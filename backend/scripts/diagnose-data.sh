#!/bin/bash

# Data Diagnosis Script
# Checks what data is actually in the database

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

FINOPS_BIN="./bin/finops"

echo -e "${BLUE}🔍 Database Data Diagnosis${NC}"
echo "=========================="

# Check if binary exists
if [[ ! -f "$FINOPS_BIN" ]]; then
    echo -e "${RED}❌ Binary not found: $FINOPS_BIN${NC}"
    echo "Run 'make build' first"
    exit 1
fi

# Step 1: Check database connection
echo -e "${BLUE}1. Testing database connection...${NC}"
if $FINOPS_BIN demo validate >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Database connection works${NC}"
else
    echo -e "${RED}❌ Database connection failed${NC}"
    echo "Make sure PostgreSQL is running and configured correctly"
    echo "Try: make dev-db-start && make migrate-up"
    exit 1
fi

# Step 2: Load demo data
echo -e "${BLUE}2. Loading demo data...${NC}"
if $FINOPS_BIN demo seed 2>&1; then
    echo -e "${GREEN}✅ Demo data loaded${NC}"
else
    echo -e "${RED}❌ Demo data loading failed${NC}"
    exit 1
fi

# Step 3: Check what nodes exist
echo -e "${BLUE}3. Checking nodes in database...${NC}"
echo "Available nodes:"

# We'll use a simple approach - try to generate charts for known nodes and see what happens
test_nodes=("product_p" "product_q" "rds_shared" "ec2_p" "s3_p" "platform_pool")

for node in "${test_nodes[@]}"; do
    echo -e "${BLUE}  Testing node: $node${NC}"
    
    # Try to generate a chart for this node - capture the output
    output=$(mktemp)
    if $FINOPS_BIN export chart trend \
        --node "$node" \
        --dimension "instance_hours" \
        --from "2024-01-01" \
        --to "2024-01-31" \
        --format "png" \
        --out "/tmp/test-$node.png" 2>&1 | tee "$output"; then
        
        if [[ -f "/tmp/test-$node.png" ]]; then
            size=$(stat -f%z "/tmp/test-$node.png" 2>/dev/null || stat -c%s "/tmp/test-$node.png" 2>/dev/null || echo "0")
            if [[ "$size" -gt 5000 ]]; then
                echo -e "${GREEN}    ✅ $node: Has data (${size} bytes)${NC}"
            else
                echo -e "${YELLOW}    ⚠️  $node: Chart generated but small (${size} bytes) - might be 'no data' chart${NC}"
            fi
            rm -f "/tmp/test-$node.png"
        else
            echo -e "${RED}    ❌ $node: Chart generation failed${NC}"
        fi
    else
        # Check if it's a "no data" error or something else
        if grep -q "No cost data found" "$output"; then
            echo -e "${YELLOW}    ⚠️  $node: Node exists but no cost data${NC}"
        elif grep -q "Failed to get node" "$output"; then
            echo -e "${RED}    ❌ $node: Node not found in database${NC}"
        else
            echo -e "${RED}    ❌ $node: Unknown error${NC}"
            cat "$output"
        fi
    fi
    rm -f "$output"
done

# Step 4: Test with different dimensions
echo -e "${BLUE}4. Testing different cost dimensions...${NC}"
dimensions=("instance_hours" "storage_gb_month" "egress_gb" "iops" "backups_gb_month")

for dim in "${dimensions[@]}"; do
    echo -e "${BLUE}  Testing dimension: $dim${NC}"
    
    output=$(mktemp)
    if $FINOPS_BIN export chart trend \
        --node "product_p" \
        --dimension "$dim" \
        --from "2024-01-01" \
        --to "2024-01-31" \
        --format "png" \
        --out "/tmp/test-dim-$dim.png" 2>&1 | tee "$output"; then
        
        if [[ -f "/tmp/test-dim-$dim.png" ]]; then
            size=$(stat -f%z "/tmp/test-dim-$dim.png" 2>/dev/null || stat -c%s "/tmp/test-dim-$dim.png" 2>/dev/null || echo "0")
            if [[ "$size" -gt 5000 ]]; then
                echo -e "${GREEN}    ✅ $dim: Has data (${size} bytes)${NC}"
            else
                echo -e "${YELLOW}    ⚠️  $dim: No data (${size} bytes)${NC}"
            fi
            rm -f "/tmp/test-dim-$dim.png"
        fi
    else
        if grep -q "No cost data found" "$output"; then
            echo -e "${YELLOW}    ⚠️  $dim: No cost data${NC}"
        else
            echo -e "${RED}    ❌ $dim: Error${NC}"
        fi
    fi
    rm -f "$output"
done

# Step 5: Test date ranges
echo -e "${BLUE}5. Testing different date ranges...${NC}"
date_ranges=(
    "2024-01-01 2024-01-31"
    "$(date -d '30 days ago' '+%Y-%m-%d') $(date '+%Y-%m-%d')"
    "$(date -d '7 days ago' '+%Y-%m-%d') $(date '+%Y-%m-%d')"
)

for range in "${date_ranges[@]}"; do
    read -r start_date end_date <<< "$range"
    echo -e "${BLUE}  Testing range: $start_date to $end_date${NC}"
    
    output=$(mktemp)
    if $FINOPS_BIN export chart trend \
        --node "product_p" \
        --dimension "instance_hours" \
        --from "$start_date" \
        --to "$end_date" \
        --format "png" \
        --out "/tmp/test-range.png" 2>&1 | tee "$output"; then
        
        if [[ -f "/tmp/test-range.png" ]]; then
            size=$(stat -f%z "/tmp/test-range.png" 2>/dev/null || stat -c%s "/tmp/test-range.png" 2>/dev/null || echo "0")
            if [[ "$size" -gt 5000 ]]; then
                echo -e "${GREEN}    ✅ $start_date to $end_date: Has data (${size} bytes)${NC}"
            else
                echo -e "${YELLOW}    ⚠️  $start_date to $end_date: No data (${size} bytes)${NC}"
            fi
            rm -f "/tmp/test-range.png"
        fi
    else
        if grep -q "No cost data found" "$output"; then
            echo -e "${YELLOW}    ⚠️  $start_date to $end_date: No cost data${NC}"
        else
            echo -e "${RED}    ❌ $start_date to $end_date: Error${NC}"
        fi
    fi
    rm -f "$output"
done

echo ""
echo -e "${BLUE}📊 Summary${NC}"
echo "=========="
echo -e "${BLUE}If you see 'No cost data found' messages:${NC}"
echo "1. The nodes exist but cost data is missing"
echo "2. Try running: $FINOPS_BIN demo seed"
echo "3. Check the date ranges match when the data was seeded"
echo ""
echo -e "${BLUE}If you see 'Node not found' messages:${NC}"
echo "1. The demo seeding failed"
echo "2. Check database connection and permissions"
echo ""
echo -e "${BLUE}💡 Next steps:${NC}"
echo "1. If data exists, charts should work: ./scripts/generate-charts.sh demo"
echo "2. If no data, re-run seeding: $FINOPS_BIN demo seed"
echo "3. Check database directly if issues persist"
