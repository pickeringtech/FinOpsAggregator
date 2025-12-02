#!/bin/bash

# FinOps Export Demo Script
# Demonstrates exporting cost data, reports, and charts

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

FINOPS_BIN="./bin/finops"
OUTPUT_DIR="./exports"

echo -e "${BLUE}📤 FinOps Export Demo${NC}"
echo "=================================================================="
echo ""

# Check if binary exists
if [[ ! -f "$FINOPS_BIN" ]]; then
    echo -e "${RED}❌ Binary not found: $FINOPS_BIN${NC}"
    echo "Run 'make build' first"
    exit 1
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"
echo -e "${GREEN}✅ Created output directory: $OUTPUT_DIR${NC}"
echo ""

# Get date range
end_date=$(date '+%Y-%m-%d')
start_date=$(date -d '30 days ago' '+%Y-%m-%d' 2>/dev/null || date -v-30d '+%Y-%m-%d')

echo -e "${CYAN}📅 Export Period${NC}"
echo "────────────────────────────────────────────────────────────────"
echo "   From: $start_date"
echo "   To:   $end_date"
echo ""

# Step 1: Export comprehensive HTML report
echo -e "${CYAN}1. 📋 Comprehensive HTML Report${NC}"
echo "────────────────────────────────────────────────────────────────"
echo ""

report_file="$OUTPUT_DIR/finops-report-$(date '+%Y%m%d').html"
echo -e "${YELLOW}Command:${NC} $FINOPS_BIN report generate --output $report_file"
echo ""

if $FINOPS_BIN report generate --output "$report_file" 2>/dev/null; then
    echo -e "${GREEN}✅ HTML report generated${NC}"
    if [[ -f "$report_file" ]]; then
        size=$(du -h "$report_file" | cut -f1)
        echo "   • File: $report_file"
        echo "   • Size: $size"
        echo "   • Contents: Executive summary, cost breakdown, trends, recommendations"
    fi
else
    echo -e "${YELLOW}⚠️  Report generation skipped (may need data)${NC}"
fi
echo ""

# Step 2: Export JSON report
echo -e "${CYAN}2. 📊 JSON Data Export${NC}"
echo "────────────────────────────────────────────────────────────────"
echo ""

json_file="$OUTPUT_DIR/cost-analysis-$(date '+%Y%m%d').json"
echo -e "${YELLOW}Command:${NC} $FINOPS_BIN analyze costs --format json > $json_file"
echo ""

if $FINOPS_BIN analyze costs --format json > "$json_file" 2>/dev/null; then
    echo -e "${GREEN}✅ JSON export generated${NC}"
    if [[ -f "$json_file" ]] && [[ -s "$json_file" ]]; then
        size=$(du -h "$json_file" | cut -f1)
        echo "   • File: $json_file"
        echo "   • Size: $size"
        
        # Show structure
        echo "   • Structure:"
        head -20 "$json_file" | python3 -m json.tool 2>/dev/null | head -15 || head -10 "$json_file"
    fi
else
    echo -e "${YELLOW}⚠️  JSON export skipped (may need data)${NC}"
fi
echo ""

# Step 3: Export optimization recommendations
echo -e "${CYAN}3. 💡 Optimization Recommendations Export${NC}"
echo "────────────────────────────────────────────────────────────────"
echo ""

opt_file="$OUTPUT_DIR/optimizations-$(date '+%Y%m%d').json"
echo -e "${YELLOW}Command:${NC} $FINOPS_BIN analyze optimization --format json > $opt_file"
echo ""

if $FINOPS_BIN analyze optimization --format json > "$opt_file" 2>/dev/null; then
    echo -e "${GREEN}✅ Optimization recommendations exported${NC}"
    if [[ -f "$opt_file" ]] && [[ -s "$opt_file" ]]; then
        size=$(du -h "$opt_file" | cut -f1)
        echo "   • File: $opt_file"
        echo "   • Size: $size"
        
        # Count recommendations
        count=$(cat "$opt_file" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "N/A")
        echo "   • Recommendations: $count"
    fi
else
    echo -e "${YELLOW}⚠️  Optimization export skipped (may need data)${NC}"
fi
echo ""

# Step 4: Export graph structure chart
echo -e "${CYAN}4. 🔗 Graph Structure Chart${NC}"
echo "────────────────────────────────────────────────────────────────"
echo ""

graph_file="$OUTPUT_DIR/graph-structure-$(date '+%Y%m%d').png"
echo -e "${YELLOW}Command:${NC} $FINOPS_BIN export chart graph --out $graph_file --format png"
echo ""

if $FINOPS_BIN export chart graph --out "$graph_file" --format png 2>/dev/null; then
    echo -e "${GREEN}✅ Graph structure chart exported${NC}"
    if [[ -f "$graph_file" ]]; then
        size=$(du -h "$graph_file" | cut -f1)
        echo "   • File: $graph_file"
        echo "   • Size: $size"
        echo "   • Shows: Node relationships and cost flow direction"
    fi
else
    echo -e "${YELLOW}⚠️  Graph chart export skipped${NC}"
fi
echo ""

# Step 5: Export cost trend charts
echo -e "${CYAN}5. 📈 Cost Trend Charts${NC}"
echo "────────────────────────────────────────────────────────────────"
echo ""

# Get list of nodes for trend charts
echo -e "${YELLOW}Generating trend charts for top nodes...${NC}"
echo ""

# Try to get node list from analysis
nodes=$($FINOPS_BIN analyze costs --format json 2>/dev/null | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if isinstance(data, dict) and 'nodes' in data:
        for n in data['nodes'][:5]:
            print(n.get('name', n.get('id', '')))
    elif isinstance(data, list):
        for n in data[:5]:
            print(n.get('name', n.get('id', '')))
except:
    pass
" 2>/dev/null || echo "")

if [[ -n "$nodes" ]]; then
    while IFS= read -r node; do
        if [[ -n "$node" ]]; then
            trend_file="$OUTPUT_DIR/trend-${node}-$(date '+%Y%m%d').png"
            echo "   Exporting trend for: $node"
            if $FINOPS_BIN export chart trend --node "$node" --from "$start_date" --to "$end_date" --out "$trend_file" 2>/dev/null; then
                echo -e "   ${GREEN}✅ $trend_file${NC}"
            fi
        fi
    done <<< "$nodes"
else
    echo -e "${YELLOW}   No nodes found for trend charts${NC}"
fi
echo ""

# Step 6: Summary
echo -e "${CYAN}📊 Export Summary${NC}"
echo "────────────────────────────────────────────────────────────────"
echo ""

echo -e "${GREEN}✅ Export Demo Complete!${NC}"
echo ""

echo -e "${YELLOW}Generated Files:${NC}"
find "$OUTPUT_DIR" -type f -newer /tmp 2>/dev/null | while read -r file; do
    size=$(du -h "$file" | cut -f1)
    echo "   • $(basename "$file") ($size)"
done || ls -lh "$OUTPUT_DIR" 2>/dev/null | tail -n +2 | awk '{print "   • " $NF " (" $5 ")"}'
echo ""

echo -e "${YELLOW}Export Formats Available:${NC}"
echo "   • HTML  - Interactive reports for stakeholders"
echo "   • JSON  - Machine-readable data for integrations"
echo "   • PNG   - Charts for presentations"
echo "   • SVG   - Scalable charts for web embedding"
echo ""

echo -e "${YELLOW}Integration Examples:${NC}"
echo ""
echo "   # Export to S3 for dashboards"
echo "   aws s3 cp $OUTPUT_DIR/ s3://my-bucket/finops-reports/ --recursive"
echo ""
echo "   # Send to Slack"
echo "   curl -F file=@$report_file https://slack.com/api/files.upload"
echo ""
echo "   # Import JSON into data warehouse"
echo "   cat $json_file | jq -c '.[]' | bq load --source_format=NEWLINE_DELIMITED_JSON"
echo ""

echo -e "${CYAN}💡 Automation Tips:${NC}"
echo ""
echo "   1. Schedule daily exports with cron:"
echo "      0 6 * * * cd /path/to/finops && ./scripts/demo-export.sh"
echo ""
echo "   2. Integrate with CI/CD for cost tracking:"
echo "      - Export after each deployment"
echo "      - Compare costs before/after changes"
echo ""
echo "   3. Set up alerts on exported data:"
echo "      - Monitor for cost spikes"
echo "      - Track unallocated cost trends"
echo ""

echo -e "${BLUE}🎉 Export Demo Complete!${NC}"
echo ""
echo "Output directory: $OUTPUT_DIR"

