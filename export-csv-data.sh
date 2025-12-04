#!/bin/bash

# CSV Data Export Script for FinOps Aggregator
# Exports various types of financial data to CSV files in the exports/ directory

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
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$BACKEND_URL/health" > /dev/null 2>&1; then
            print_success "Backend is ready"
            return 0
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            print_error "Backend is not responding after $max_attempts attempts"
            exit 1
        fi
        
        sleep 2
        ((attempt++))
    done
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
    
    print_info "URL: $url"
    print_info "Output: $output_file"
    
    if curl -s "$url" -o "$output_file"; then
        local file_size=$(stat -f%z "$output_file" 2>/dev/null || stat -c%s "$output_file" 2>/dev/null || echo "unknown")
        local line_count=$(wc -l < "$output_file" 2>/dev/null || echo "unknown")
        
        print_success "Export completed - Size: $file_size bytes, Lines: $line_count"
        
        # Show sample data (first 3 lines)
        echo "Sample data:"
        head -3 "$output_file" | sed 's/^/  /'
        
        # Validate that we have actual data (more than just headers)
        if [ "$line_count" -gt 1 ] 2>/dev/null; then
            print_success "✓ Contains actual data ($line_count total lines)"
        else
            print_warning "⚠ Only contains headers (no data for date range)"
        fi
        
        return 0
    else
        print_error "Export failed"
        return 1
    fi
}

echo "🚀 FinOps CSV Data Export"
echo "========================="
print_info "Date range: $DATE_RANGE_START to $DATE_RANGE_END"
print_info "Currency: $CURRENCY"
print_info "Output directory: $EXPORTS_DIR"
echo ""

# Check backend availability
check_backend

echo ""
echo "📊 Starting CSV Data Exports"
echo "============================="
echo ""

# Summary Exports (Aggregated Data)
print_info "=== SUMMARY EXPORTS (Aggregated Data) ==="
echo ""

# Export 1: Products with aggregated costs
export_api "products" "" "Products with aggregated costs" "products_summary.csv"
echo ""

# Export 2: All nodes with aggregated costs
export_api "nodes" "" "All nodes with aggregated costs" "nodes_summary.csv"
echo ""

# Export 3: Product nodes only
export_api "nodes" "node_type=product" "Product nodes with aggregated costs" "product_nodes_summary.csv"
echo ""

# Export 4: Costs breakdown by type
export_api "costs_by_type" "" "Cost breakdown by node type" "costs_by_type_summary.csv"
echo ""

# Export 5: Recommendations
export_api "recommendations" "" "Cost optimization recommendations" "recommendations.csv"
echo ""

# Detailed Exports (Individual Records)
print_info "=== DETAILED EXPORTS (Individual Records) ==="
echo ""

# Export 6: Detailed cost records (WARNING: Large file - 779K+ rows)
print_warning "The following exports contain hundreds of thousands of records and may take time..."
echo ""

export_api "detailed_costs" "" "Detailed cost allocation records (779K+ rows)" "detailed_costs_full.csv"
echo ""

# Export 7: Detailed costs for products only
export_api "detailed_costs" "node_type=product" "Detailed product cost records" "detailed_costs_products.csv"
echo ""

# Export 8: Raw ingested cost data (WARNING: Large file - 368K+ rows)
export_api "raw_costs" "" "Raw ingested cost data (368K+ rows)" "raw_costs_full.csv"
echo ""

# Export 9: Raw costs for products only
export_api "raw_costs" "node_type=product" "Raw product cost data" "raw_costs_products.csv"
echo ""

# Export 10: Product hierarchy with downstream relationships (NEW!)
export_api "product_hierarchy" "" "Product hierarchy with downstream relationships (475K+ rows)" "product_hierarchy_full.csv"
echo ""

# Summary
echo "🎉 CSV Export Complete!"
echo "======================="
print_success "All exports completed successfully!"
print_info "Files saved to: $EXPORTS_DIR/"
echo ""
print_info "File summary:"
ls -lh "$EXPORTS_DIR"/*.csv 2>/dev/null | while read -r line; do
    echo "  $line"
done
echo ""
print_info "Total disk usage:"
du -sh "$EXPORTS_DIR" 2>/dev/null || echo "  Unable to calculate disk usage"
