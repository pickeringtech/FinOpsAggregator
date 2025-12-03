#!/bin/bash

# Quick CSV Summary Export Script for FinOps Aggregator
# Exports only summary/aggregated data (fast, small files)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BACKEND_URL="http://localhost:8080"
EXPORTS_DIR="./exports"
DATE_RANGE_START="2024-12-01"
DATE_RANGE_END="2026-12-31"
CURRENCY="USD"

# Create exports directory if it doesn't exist
mkdir -p "$EXPORTS_DIR"

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to check if backend is ready
check_backend() {
    print_info "Checking if backend is ready..."
    if curl -s "$BACKEND_URL/health" > /dev/null 2>&1; then
        print_success "Backend is ready"
        return 0
    else
        print_error "Backend is not responding"
        exit 1
    fi
}

# Function to export via API
export_api() {
    local export_type="$1"
    local additional_params="$2"
    local description="$3"
    local filename="$4"
    
    print_info "Exporting: $description"
    
    local url="${BACKEND_URL}/api/v1/export/csv?type=${export_type}&start_date=${DATE_RANGE_START}&end_date=${DATE_RANGE_END}&currency=${CURRENCY}"
    if [ -n "$additional_params" ]; then
        url="${url}&${additional_params}"
    fi
    
    local output_file="${EXPORTS_DIR}/${filename}"
    
    if curl -s "$url" -o "$output_file"; then
        local file_size=$(stat -f%z "$output_file" 2>/dev/null || stat -c%s "$output_file" 2>/dev/null || echo "unknown")
        local line_count=$(wc -l < "$output_file" 2>/dev/null || echo "unknown")
        
        print_success "✓ $filename - $line_count lines, $file_size bytes"
        return 0
    else
        print_error "✗ Failed to export $filename"
        return 1
    fi
}

echo "⚡ FinOps Quick CSV Summary Export"
echo "================================="
print_info "Date range: $DATE_RANGE_START to $DATE_RANGE_END"
print_info "Currency: $CURRENCY"
print_info "Output directory: $EXPORTS_DIR"
echo ""

# Check backend availability
check_backend

echo ""
print_info "Exporting summary data (fast, small files)..."
echo ""

# Summary Exports Only
export_api "products" "" "Products summary" "products_summary.csv"
export_api "nodes" "" "All nodes summary" "nodes_summary.csv"
export_api "nodes" "node_type=product" "Product nodes summary" "product_nodes_summary.csv"
export_api "costs_by_type" "" "Costs by type summary" "costs_by_type_summary.csv"
export_api "recommendations" "" "Recommendations" "recommendations.csv"

echo ""
print_success "Quick summary export complete!"
print_info "Files saved to: $EXPORTS_DIR/"
echo ""
ls -lh "$EXPORTS_DIR"/*.csv 2>/dev/null | grep -E "(products_summary|nodes_summary|product_nodes_summary|costs_by_type_summary|recommendations)" | while read -r line; do
    echo "  $line"
done
echo ""
print_info "For detailed exports (779K+ rows), use: ./export-csv-data.sh"
