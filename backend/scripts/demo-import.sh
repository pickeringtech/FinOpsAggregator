#!/bin/bash

# FinOps AWS CUR Import Demo Script
# Demonstrates importing cost data from AWS Cost and Usage Report (CUR) CSV files

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
SAMPLE_CSV="./testdata/sample_aws_cur.csv"

echo -e "${BLUE}📥 FinOps AWS CUR Import Demo${NC}"
echo "=================================================================="
echo ""

# Check if binary exists
if [[ ! -f "$FINOPS_BIN" ]]; then
    echo -e "${RED}❌ Binary not found: $FINOPS_BIN${NC}"
    echo "Run 'make build' first"
    exit 1
fi

# Check if sample CSV exists
if [[ ! -f "$SAMPLE_CSV" ]]; then
    echo -e "${RED}❌ Sample CSV not found: $SAMPLE_CSV${NC}"
    exit 1
fi

# Step 1: Explain the sample data
echo -e "${CYAN}📋 Sample AWS CUR Data Overview${NC}"
echo "────────────────────────────────────────────────────────────────"
echo ""
echo -e "${YELLOW}File:${NC} $SAMPLE_CSV"
echo ""

# Count records
record_count=$(tail -n +2 "$SAMPLE_CSV" | wc -l)
echo -e "${YELLOW}Records:${NC} $record_count cost entries"
echo ""

echo -e "${YELLOW}AWS Services Included:${NC}"
tail -n +2 "$SAMPLE_CSV" | cut -d',' -f2 | sort | uniq -c | sort -rn | head -10 | while read count service; do
    echo "   • $service: $count records"
done
echo ""

echo -e "${YELLOW}Date Range:${NC}"
start_date=$(tail -n +2 "$SAMPLE_CSV" | cut -d',' -f1 | sort | head -1)
end_date=$(tail -n +2 "$SAMPLE_CSV" | cut -d',' -f1 | sort | tail -1)
echo "   • From: $start_date"
echo "   • To:   $end_date"
echo ""

echo -e "${YELLOW}Total Cost:${NC}"
total=$(tail -n +2 "$SAMPLE_CSV" | cut -d',' -f3 | awk '{sum += $1} END {printf "%.2f", sum}')
echo "   • \$$total USD"
echo ""

echo -e "${YELLOW}CSV Columns:${NC}"
head -1 "$SAMPLE_CSV" | tr ',' '\n' | while read col; do
    echo "   • $col"
done
echo ""

# Step 2: Show sample records
echo -e "${CYAN}📄 Sample Records (first 5)${NC}"
echo "────────────────────────────────────────────────────────────────"
echo ""
head -6 "$SAMPLE_CSV" | column -t -s',' 2>/dev/null || head -6 "$SAMPLE_CSV"
echo ""

# Step 3: Import options
echo -e "${CYAN}🔧 Import Options${NC}"
echo "────────────────────────────────────────────────────────────────"
echo ""
echo -e "${YELLOW}Available Flags:${NC}"
echo "   --create-nodes    Create missing nodes from AWS product codes"
echo "   --allocate        Run cost allocation after import"
echo "   --from DATE       Start date for allocation (YYYY-MM-DD)"
echo "   --to DATE         End date for allocation (YYYY-MM-DD)"
echo ""

# Step 4: Run import
echo -e "${CYAN}🚀 Running Import${NC}"
echo "────────────────────────────────────────────────────────────────"
echo ""
echo -e "${YELLOW}Command:${NC} $FINOPS_BIN import costs $SAMPLE_CSV --create-nodes"
echo ""

$FINOPS_BIN import costs "$SAMPLE_CSV" --create-nodes

echo ""

# Step 5: Verify import
echo -e "${CYAN}✅ Verifying Import${NC}"
echo "────────────────────────────────────────────────────────────────"
echo ""

echo -e "${YELLOW}Running cost analysis to verify imported data...${NC}"
echo ""
$FINOPS_BIN analyze costs --top 10 2>/dev/null || echo "Cost analysis not available"
echo ""

# Step 6: Optional allocation
echo -e "${CYAN}⚖️  Running Allocation (Optional)${NC}"
echo "────────────────────────────────────────────────────────────────"
echo ""

# Extract dates from sample data for allocation
alloc_from=$(echo "$start_date" | cut -d'T' -f1)
alloc_to=$(echo "$end_date" | cut -d'T' -f1)

echo -e "${YELLOW}Command:${NC} $FINOPS_BIN allocate --from $alloc_from --to $alloc_to"
echo ""

if $FINOPS_BIN allocate --from "$alloc_from" --to "$alloc_to" 2>/dev/null; then
    echo -e "${GREEN}✅ Allocation completed successfully${NC}"
else
    echo -e "${YELLOW}⚠️  Allocation skipped or had issues${NC}"
fi
echo ""

# Step 7: Summary
echo -e "${CYAN}📊 Import Summary${NC}"
echo "────────────────────────────────────────────────────────────────"
echo ""
echo -e "${GREEN}✅ Import Demo Complete!${NC}"
echo ""
echo -e "${YELLOW}What was demonstrated:${NC}"
echo "   1. Parsed AWS CUR CSV format"
echo "   2. Mapped costs to nodes via resource tags"
echo "   3. Created missing nodes from AWS product codes"
echo "   4. Inserted cost records into the database"
echo "   5. Ran cost allocation to attribute costs to products"
echo ""

echo -e "${YELLOW}Next Steps:${NC}"
echo ""
echo "   • View the dashboard:  Open http://localhost:3000"
echo "   • Analyze costs:       $FINOPS_BIN analyze costs"
echo "   • Generate reports:    $FINOPS_BIN report generate --output report.html"
echo "   • Export charts:       $FINOPS_BIN export chart graph --out graph.png"
echo ""

echo -e "${CYAN}💡 Tips for Production Use:${NC}"
echo ""
echo "   1. Schedule daily imports from your AWS CUR S3 bucket"
echo "   2. Use resource tags (Product, Service, CostCenter) for accurate mapping"
echo "   3. Run allocation after each import to keep costs up-to-date"
echo "   4. Set up alerts for unallocated costs exceeding thresholds"
echo ""
echo -e "${BLUE}🎉 AWS CUR Import Demo Complete!${NC}"

