#!/bin/bash

# Chart Testing Script
# Tests chart generation functionality to ensure it works

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
FINOPS_BIN="./bin/finops"
TEST_DIR="./test-charts"
FAILED_TESTS=0
TOTAL_TESTS=0

echo -e "${BLUE}🧪 Testing FinOps Chart Generation${NC}"
echo "=================================="

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_file="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${BLUE}Testing: $test_name${NC}"
    
    # Run the command
    if eval "$command" >/dev/null 2>&1; then
        # Check if expected file was created
        if [[ -n "$expected_file" && -f "$expected_file" ]]; then
            local file_size=$(stat -f%z "$expected_file" 2>/dev/null || stat -c%s "$expected_file" 2>/dev/null || echo "0")
            if [[ "$file_size" -gt 1000 ]]; then
                echo -e "${GREEN}  ✅ PASS - File created ($file_size bytes)${NC}"
                return 0
            else
                echo -e "${RED}  ❌ FAIL - File too small ($file_size bytes)${NC}"
                FAILED_TESTS=$((FAILED_TESTS + 1))
                return 1
            fi
        else
            echo -e "${GREEN}  ✅ PASS - Command succeeded${NC}"
            return 0
        fi
    else
        echo -e "${RED}  ❌ FAIL - Command failed${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Function to setup test environment
setup_test_env() {
    echo -e "${BLUE}🔧 Setting up test environment...${NC}"
    
    # Create test directory
    mkdir -p "$TEST_DIR"
    
    # Check if binary exists
    if [[ ! -f "$FINOPS_BIN" ]]; then
        echo -e "${YELLOW}⚠️  Building FinOps binary...${NC}"
        if ! make build >/dev/null 2>&1; then
            echo -e "${RED}❌ Failed to build FinOps binary${NC}"
            exit 1
        fi
    fi
    
    # Load demo data if needed
    echo -e "${BLUE}📊 Loading demo data...${NC}"
    if ! $FINOPS_BIN demo seed >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Demo data loading failed, continuing anyway...${NC}"
    fi
    
    echo -e "${GREEN}✅ Test environment ready${NC}"
    echo ""
}

# Function to cleanup test environment
cleanup_test_env() {
    echo -e "${BLUE}🧹 Cleaning up test files...${NC}"
    rm -rf "$TEST_DIR"
}

# Function to test basic chart functionality
test_basic_charts() {
    echo -e "${BLUE}📊 Testing Basic Chart Generation${NC}"
    echo "--------------------------------"
    
    # Test 1: Graph structure chart
    run_test "Graph Structure (PNG)" \
        "$FINOPS_BIN export chart graph --format png --out $TEST_DIR/graph-test.png" \
        "$TEST_DIR/graph-test.png"
    
    # Test 2: Graph structure chart (SVG)
    run_test "Graph Structure (SVG)" \
        "$FINOPS_BIN export chart graph --format svg --out $TEST_DIR/graph-test.svg" \
        "$TEST_DIR/graph-test.svg"
    
    # Test 3: Cost trend chart (this might fail if no data)
    run_test "Cost Trend Chart" \
        "$FINOPS_BIN export chart trend --node product_p --dimension instance_hours --from 2024-01-01 --to 2024-01-31 --format png --out $TEST_DIR/trend-test.png" \
        "$TEST_DIR/trend-test.png"
    
    echo ""
}

# Function to test error conditions
test_error_conditions() {
    echo -e "${BLUE}🚨 Testing Error Conditions${NC}"
    echo "----------------------------"
    
    # Test 1: Invalid format
    run_test "Invalid Format" \
        "$FINOPS_BIN export chart graph --format invalid --out $TEST_DIR/invalid.invalid 2>&1 | grep -q 'unsupported format'" \
        ""
    
    # Test 2: Invalid node
    run_test "Invalid Node" \
        "$FINOPS_BIN export chart trend --node nonexistent --dimension instance_hours --from 2024-01-01 --to 2024-01-31 --format png --out $TEST_DIR/invalid-node.png 2>&1 | grep -q 'invalid node'" \
        ""
    
    echo ""
}

# Function to test batch generation
test_batch_generation() {
    echo -e "${BLUE}📦 Testing Batch Generation${NC}"
    echo "---------------------------"
    
    # Test batch script
    run_test "Batch Demo Charts" \
        "./scripts/generate-charts.sh demo --dir $TEST_DIR --format png" \
        ""
    
    echo ""
}

# Function to show test results
show_results() {
    echo -e "${BLUE}📋 Test Results${NC}"
    echo "==============="
    echo "Total Tests: $TOTAL_TESTS"
    echo "Failed Tests: $FAILED_TESTS"
    echo "Passed Tests: $((TOTAL_TESTS - FAILED_TESTS))"
    
    if [[ $FAILED_TESTS -eq 0 ]]; then
        echo -e "${GREEN}🎉 All tests passed!${NC}"
        
        # Show generated files
        if [[ -d "$TEST_DIR" ]]; then
            echo ""
            echo -e "${BLUE}📁 Generated test files:${NC}"
            find "$TEST_DIR" -type f | sort | while read -r file; do
                size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")
                echo "  $file ($(numfmt --to=iec $size 2>/dev/null || echo "${size} bytes"))"
            done
        fi
        
        return 0
    else
        echo -e "${RED}❌ Some tests failed${NC}"
        return 1
    fi
}

# Main execution
main() {
    setup_test_env
    
    test_basic_charts
    test_error_conditions
    test_batch_generation
    
    show_results
    local result=$?
    
    cleanup_test_env
    
    return $result
}

# Handle script arguments
case "${1:-}" in
    "basic")
        setup_test_env
        test_basic_charts
        show_results
        cleanup_test_env
        ;;
    "errors")
        setup_test_env
        test_error_conditions
        show_results
        cleanup_test_env
        ;;
    "batch")
        setup_test_env
        test_batch_generation
        show_results
        cleanup_test_env
        ;;
    "clean")
        cleanup_test_env
        echo -e "${GREEN}✅ Test files cleaned${NC}"
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [COMMAND]"
        echo ""
        echo "Commands:"
        echo "  basic    - Test basic chart generation"
        echo "  errors   - Test error conditions"
        echo "  batch    - Test batch generation"
        echo "  clean    - Clean up test files"
        echo "  (none)   - Run all tests"
        ;;
    *)
        main
        ;;
esac
