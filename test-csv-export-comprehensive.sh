#!/bin/bash

# Comprehensive CSV Export Test Script for Docker Compose Environment
# Tests both API endpoints and CLI commands within the docker environment

set -e

echo "🧪 Comprehensive CSV Export Testing"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
API_BASE_URL="http://localhost:8080"
EXPORT_ENDPOINT="$API_BASE_URL/api/v1/export/csv"
START_DATE="2024-12-01"
END_DATE="2026-12-31"
CURRENCY="USD"

# Function to print colored output
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }
print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }

# Function to test API endpoint
test_api_export() {
    local export_type="$1"
    local additional_params="$2"
    local description="$3"
    
    echo ""
    print_info "Testing API: $description"
    
    # Build URL
    local url="$EXPORT_ENDPOINT?type=$export_type&start_date=$START_DATE&end_date=$END_DATE&currency=$CURRENCY"
    if [ -n "$additional_params" ]; then
        url="$url&$additional_params"
    fi
    
    # Test the endpoint
    local output_file="api_${export_type}_export.csv"
    if [ -n "$additional_params" ]; then
        output_file="api_${export_type}_$(echo $additional_params | tr '&=' '_')_export.csv"
    fi
    
    print_info "URL: $url"
    print_info "Output: $output_file"
    
    if curl -s -f -o "$output_file" "$url"; then
        local file_size=$(wc -c < "$output_file")
        local line_count=$(wc -l < "$output_file")
        print_success "API Export completed - Size: ${file_size} bytes, Lines: ${line_count}"
        
        # Show first few lines
        echo "Sample data:"
        head -3 "$output_file" | sed 's/^/  /'
        
        # Check if we have data beyond headers
        if [ "$line_count" -gt 1 ]; then
            print_success "✓ Contains actual data (${line_count} total lines)"
        else
            print_warning "⚠ Only contains headers (no data for date range)"
        fi
    else
        print_error "API Export failed"
        return 1
    fi
}

# Function to test CLI command within docker container
test_cli_export() {
    local export_type="$1"
    local additional_params="$2"
    local description="$3"
    
    echo ""
    print_info "Testing CLI: $description"
    
    # Build CLI command
    local cli_cmd="./finops export csv --type=$export_type --start-date=$START_DATE --end-date=$END_DATE --currency=$CURRENCY"
    if [ -n "$additional_params" ]; then
        cli_cmd="$cli_cmd $additional_params"
    fi
    
    local output_file="cli_${export_type}_export.csv"
    if [ -n "$additional_params" ]; then
        output_file="cli_${export_type}_$(echo $additional_params | tr ' -=' '_')_export.csv"
    fi
    
    print_info "Command: $cli_cmd --out=$output_file"
    
    # Execute within the backend container
    if docker-compose -f docker-compose.dev.yml exec -T backend sh -c "cd /app && [ ! -f ./finops ] && go build -o finops ./cmd/finops; $cli_cmd --out=/tmp/$output_file && cat /tmp/$output_file" > "$output_file" 2>/dev/null; then
        local file_size=$(wc -c < "$output_file")
        local line_count=$(wc -l < "$output_file")
        print_success "CLI Export completed - Size: ${file_size} bytes, Lines: ${line_count}"
        
        # Show first few lines
        echo "Sample data:"
        head -3 "$output_file" | sed 's/^/  /'
        
        # Check if we have data beyond headers
        if [ "$line_count" -gt 1 ]; then
            print_success "✓ Contains actual data (${line_count} total lines)"
        else
            print_warning "⚠ Only contains headers (no data for date range)"
        fi
    else
        print_error "CLI Export failed"
        return 1
    fi
}

# Wait for backend to be ready
print_info "Checking if backend is ready..."
for i in {1..10}; do
    if curl -s -f "http://localhost:8080/health" > /dev/null 2>&1; then
        print_success "Backend is ready"
        break
    fi
    if [ $i -eq 10 ]; then
        print_error "Backend not ready after 10 attempts"
        exit 1
    fi
    echo "Attempt $i/10: Backend not ready, waiting..."
    sleep 2
done

# Check what date range has data
print_info "Checking available data date range..."
DATE_RANGE=$(docker-compose -f docker-compose.dev.yml exec -T postgres psql -U finops -d finops -t -c "SELECT MIN(cost_date) || ' to ' || MAX(cost_date) FROM node_costs_by_dimension;" 2>/dev/null | tr -d ' ')
print_info "Available data range: $DATE_RANGE"

echo ""
echo "🔬 Starting CSV Export Tests"
echo "============================="

# Test 1: Products Export
test_api_export "products" "" "Products with costs export"
test_cli_export "products" "" "Products with costs export"

# Test 2: Nodes Export (all nodes)
test_api_export "nodes" "" "All nodes export"
test_cli_export "nodes" "" "All nodes export"

# Test 3: Nodes Export (filtered by type)
test_api_export "nodes" "node_type=product" "Product nodes export"
test_cli_export "nodes" "--node-type=product" "Product nodes export"

# Test 4: Costs by Type Export
test_api_export "costs_by_type" "" "Costs by type export"
test_cli_export "costs_by_type" "" "Costs by type export"

# Test 5: Recommendations Export
test_api_export "recommendations" "" "All recommendations export"
test_cli_export "recommendations" "" "All recommendations export"

# Test 6: Detailed Costs Export (NEW - shows individual cost records)
test_api_export "detailed_costs" "" "Detailed cost records export"
test_cli_export "detailed_costs" "" "Detailed cost records export"

# Test 7: Raw Costs Export (NEW - shows original ingested data)
test_api_export "raw_costs" "" "Raw cost records export"
test_cli_export "raw_costs" "" "Raw cost records export"

# Test 8: Detailed Costs Export (filtered by node type)
test_api_export "detailed_costs" "node_type=product" "Detailed product cost records export"
test_cli_export "detailed_costs" "--node-type=product" "Detailed product cost records export"

echo ""
echo "📊 Test Summary"
echo "==============="

# Count successful exports
api_files=$(ls api_*_export.csv 2>/dev/null | wc -l)
cli_files=$(ls cli_*_export.csv 2>/dev/null | wc -l)

print_success "API exports completed: $api_files"
print_success "CLI exports completed: $cli_files"

# Show file sizes
echo ""
print_info "Generated files:"
ls -la *_export.csv 2>/dev/null | while read line; do
    echo "  $line"
done

echo ""
print_success "🎉 CSV Export testing completed!"
print_info "All CSV files are ready for inspection."
