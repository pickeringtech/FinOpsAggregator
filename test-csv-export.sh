#!/bin/bash

# Test script for CSV export functionality
# This script tests the CSV export API endpoint in the docker-compose environment

set -e

API_BASE_URL="http://localhost:8080/api/v1"
EXPORT_ENDPOINT="$API_BASE_URL/export/csv"

echo "Testing CSV Export Functionality"
echo "================================="

# Function to test CSV export
test_csv_export() {
    local export_type="$1"
    local additional_params="$2"
    local description="$3"
    
    echo ""
    echo "Testing: $description"
    echo "Type: $export_type"
    echo "Additional params: $additional_params"
    
    # Build URL with parameters (using date range that has data)
    local url="$EXPORT_ENDPOINT?type=$export_type&start_date=2024-12-01&end_date=2026-12-31&currency=USD"
    if [ -n "$additional_params" ]; then
        url="$url&$additional_params"
    fi
    
    echo "URL: $url"
    
    # Make request and save response
    local output_file="test_${export_type}_export.csv"
    if [ -n "$additional_params" ]; then
        output_file="test_${export_type}_$(echo $additional_params | tr '&=' '_')_export.csv"
    fi
    
    echo "Saving to: $output_file"
    
    # Test the API endpoint
    if curl -s -f -o "$output_file" "$url"; then
        echo "✅ SUCCESS: CSV export completed"
        echo "File size: $(wc -c < "$output_file") bytes"
        echo "First few lines:"
        head -5 "$output_file" || echo "Could not read file content"
    else
        echo "❌ FAILED: CSV export failed"
        # Try to get error response
        curl -s "$url" || echo "No error response available"
    fi
}

# Wait for backend to be ready
echo "Waiting for backend to be ready..."
for i in {1..30}; do
    if curl -s -f "http://localhost:8080/health" > /dev/null 2>&1; then
        echo "✅ Backend is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ Backend not ready after 30 attempts"
        exit 1
    fi
    echo "Attempt $i/30: Backend not ready, waiting..."
    sleep 2
done

# Test different CSV export types
test_csv_export "products" "" "Products with costs export"
test_csv_export "nodes" "node_type=compute" "Compute nodes export"
test_csv_export "costs_by_type" "" "Costs by type export"
test_csv_export "recommendations" "" "All recommendations export"

echo ""
echo "================================="
echo "CSV Export Tests Completed"
echo "Check the generated CSV files for results"
