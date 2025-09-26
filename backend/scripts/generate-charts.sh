#!/bin/bash

# Chart Generation Script
# Generates various charts and visualizations from FinOps data

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
FINOPS_BIN="./bin/finops"
CHARTS_DIR="./charts"
FORMAT="png"
DATE_RANGE_START="2024-01-01"
DATE_RANGE_END="2024-01-31"

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS] [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  all         Generate all available charts"
    echo "  graph       Generate graph structure chart"
    echo "  trends      Generate cost trend charts for all nodes"
    echo "  waterfalls  Generate allocation waterfall charts"
    echo "  demo        Generate demo charts with sample data"
    echo ""
    echo "Options:"
    echo "  -f, --format FORMAT     Output format (png, svg) [default: png]"
    echo "  -d, --dir DIR          Output directory [default: ./charts]"
    echo "  -s, --start DATE       Start date for trends (YYYY-MM-DD) [default: 2024-01-01]"
    echo "  -e, --end DATE         End date for trends (YYYY-MM-DD) [default: 2024-01-31]"
    echo "  -h, --help             Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 demo                           # Generate demo charts"
    echo "  $0 graph                          # Generate graph structure"
    echo "  $0 trends --format svg            # Generate trend charts as SVG"
    echo "  $0 all --dir /tmp/charts          # Generate all charts to /tmp/charts"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--format)
            FORMAT="$2"
            shift 2
            ;;
        -d|--dir)
            CHARTS_DIR="$2"
            shift 2
            ;;
        -s|--start)
            DATE_RANGE_START="$2"
            shift 2
            ;;
        -e|--end)
            DATE_RANGE_END="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            COMMAND="$1"
            shift
            ;;
    esac
done

# Validate format
if [[ "$FORMAT" != "png" && "$FORMAT" != "svg" ]]; then
    echo -e "${RED}❌ Invalid format: $FORMAT (supported: png, svg)${NC}"
    exit 1
fi

# Create charts directory
mkdir -p "$CHARTS_DIR"

# Check if finops binary exists
if [[ ! -f "$FINOPS_BIN" ]]; then
    echo -e "${RED}❌ FinOps binary not found: $FINOPS_BIN${NC}"
    echo "Run 'make build' first to build the application."
    exit 1
fi

# Function to generate graph structure chart
generate_graph_chart() {
    echo -e "${BLUE}📊 Generating graph structure chart...${NC}"
    
    local output_file="$CHARTS_DIR/graph-structure.${FORMAT}"
    
    if $FINOPS_BIN export chart graph --format "$FORMAT" --out "$output_file"; then
        echo -e "${GREEN}✅ Graph structure chart saved to: $output_file${NC}"
    else
        echo -e "${RED}❌ Failed to generate graph structure chart${NC}"
        return 1
    fi
}

# Function to get all node names
get_node_names() {
    # This would ideally query the database, but for now we'll use demo node names
    echo "product_p product_q rds_shared platform_pool ec2_p s3_p"
}

# Function to generate trend charts for all nodes
generate_trend_charts() {
    echo -e "${BLUE}📈 Generating cost trend charts...${NC}"
    
    local nodes=$(get_node_names)
    local dimensions=("instance_hours" "storage_gb_month" "egress_gb")
    
    for node in $nodes; do
        for dimension in "${dimensions[@]}"; do
            echo -e "${BLUE}  Generating trend for $node ($dimension)...${NC}"
            
            local output_file="$CHARTS_DIR/trend-${node}-${dimension}.${FORMAT}"
            
            if $FINOPS_BIN export chart trend \
                --node "$node" \
                --dimension "$dimension" \
                --from "$DATE_RANGE_START" \
                --to "$DATE_RANGE_END" \
                --format "$FORMAT" \
                --out "$output_file" 2>/dev/null; then
                echo -e "${GREEN}    ✅ Saved: $output_file${NC}"
            else
                echo -e "${YELLOW}    ⚠️  Skipped $node ($dimension) - no data or error${NC}"
            fi
        done
    done
}

# Function to generate waterfall charts
generate_waterfall_charts() {
    echo -e "${BLUE}🌊 Generating allocation waterfall charts...${NC}"
    
    # Get the latest allocation run ID
    # This is a simplified approach - in practice you'd query the database
    echo -e "${YELLOW}⚠️  Waterfall charts require a specific allocation run ID${NC}"
    echo "   Run an allocation first: $FINOPS_BIN allocate --from $DATE_RANGE_START --to $DATE_RANGE_END"
    echo "   Then use: $FINOPS_BIN export chart waterfall --node NODE_NAME --date DATE --run RUN_ID"
}

# Function to generate demo charts
generate_demo_charts() {
    echo -e "${BLUE}🎬 Generating demo charts...${NC}"
    
    # Ensure demo data is loaded
    echo -e "${BLUE}1. Loading demo data...${NC}"
    if ! $FINOPS_BIN demo seed >/dev/null 2>&1; then
        echo -e "${RED}❌ Failed to load demo data${NC}"
        return 1
    fi
    
    # Generate graph structure
    echo -e "${BLUE}2. Generating graph structure...${NC}"
    generate_graph_chart
    
    # Generate trend charts for key nodes
    echo -e "${BLUE}3. Generating trend charts for key nodes...${NC}"
    local key_nodes=("product_p" "product_q" "rds_shared")
    local key_dimension="instance_hours"
    
    for node in "${key_nodes[@]}"; do
        echo -e "${BLUE}  Generating trend for $node...${NC}"
        
        local output_file="$CHARTS_DIR/demo-trend-${node}.${FORMAT}"
        
        if $FINOPS_BIN export chart trend \
            --node "$node" \
            --dimension "$key_dimension" \
            --from "$DATE_RANGE_START" \
            --to "$DATE_RANGE_END" \
            --format "$FORMAT" \
            --out "$output_file" 2>/dev/null; then
            echo -e "${GREEN}    ✅ Saved: $output_file${NC}"
        else
            echo -e "${YELLOW}    ⚠️  Skipped $node - no data or error${NC}"
        fi
    done
    
    echo -e "${GREEN}✅ Demo charts generated successfully!${NC}"
}

# Main command handling
case "${COMMAND:-}" in
    "graph")
        generate_graph_chart
        ;;
    
    "trends")
        generate_trend_charts
        ;;
    
    "waterfalls")
        generate_waterfall_charts
        ;;
    
    "demo")
        generate_demo_charts
        ;;
    
    "all")
        echo -e "${BLUE}🚀 Generating all charts...${NC}"
        generate_graph_chart
        generate_trend_charts
        generate_waterfall_charts
        echo -e "${GREEN}✅ All charts generated!${NC}"
        ;;
    
    "help"|"-h"|"--help"|"")
        show_usage
        ;;
    
    *)
        echo -e "${RED}❌ Unknown command: $COMMAND${NC}"
        echo ""
        show_usage
        exit 1
        ;;
esac

echo ""
echo -e "${BLUE}📁 Charts saved to: $CHARTS_DIR${NC}"
echo -e "${BLUE}📊 Format: $FORMAT${NC}"

# List generated files
if [[ -d "$CHARTS_DIR" ]]; then
    echo -e "${BLUE}📋 Generated files:${NC}"
    find "$CHARTS_DIR" -name "*.${FORMAT}" -type f | sort | while read -r file; do
        size=$(du -h "$file" | cut -f1)
        echo "  $file ($size)"
    done
fi
