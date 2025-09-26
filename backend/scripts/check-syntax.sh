#!/bin/bash

# Syntax Check Script
# Checks for common Go compilation issues without actually compiling

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Go Syntax Check${NC}"
echo "=================="

# Check for common issues
echo -e "${BLUE}Checking for common compilation issues...${NC}"

# 1. Check for empty imports (causes Go compiler panic)
echo -e "${BLUE}1. Checking for empty imports...${NC}"
empty_imports=0

for file in $(find . -name "*.go" -not -path "./vendor/*"); do
    if grep -n '^\s*""' "$file" >/dev/null 2>&1; then
        echo -e "${RED}❌ Empty import found in $file${NC}"
        grep -n '^\s*""' "$file"
        empty_imports=$((empty_imports + 1))
    fi
done

if [[ $empty_imports -eq 0 ]]; then
    echo -e "${GREEN}✅ No empty imports found${NC}"
fi

# 2. Check for unused imports
echo -e "${BLUE}2. Checking for unused imports...${NC}"
unused_imports=0

for file in $(find . -name "*.go" -not -path "./vendor/*" -not -name "*_test.go"); do
    # Look for imports that might be unused
    if grep -q "imported and not used" <(gofmt -d "$file" 2>&1) 2>/dev/null; then
        echo -e "${YELLOW}⚠️  Potential unused import in $file${NC}"
        unused_imports=$((unused_imports + 1))
    fi
done

if [[ $unused_imports -eq 0 ]]; then
    echo -e "${GREEN}✅ No obvious unused import issues${NC}"
fi

# 3. Check for undefined references
echo -e "${BLUE}3. Checking for undefined references...${NC}"
undefined_refs=0

# Common patterns that cause undefined reference errors
patterns=(
    "\.bucket\."
    "blob\."
    "ContentType"
    "chart\.TextBox"
    "color\.Color.*drawing\.Color"
)

for pattern in "${patterns[@]}"; do
    if grep -r "$pattern" --include="*.go" . >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Found potential undefined reference pattern: $pattern${NC}"
        grep -r "$pattern" --include="*.go" . | head -3
        undefined_refs=$((undefined_refs + 1))
    fi
done

if [[ $undefined_refs -eq 0 ]]; then
    echo -e "${GREEN}✅ No obvious undefined reference issues${NC}"
fi

# 4. Check for type mismatches
echo -e "${BLUE}4. Checking for type mismatches...${NC}"
type_issues=0

# Look for common type mismatch patterns
if grep -r "color\.Color" --include="*.go" . | grep -v "drawing\.Color" >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Found color.Color usage (should be drawing.Color)${NC}"
    type_issues=$((type_issues + 1))
fi

if [[ $type_issues -eq 0 ]]; then
    echo -e "${GREEN}✅ No obvious type mismatch issues${NC}"
fi

# 5. Check for missing imports
echo -e "${BLUE}5. Checking for missing imports...${NC}"
missing_imports=0

# Check if files use certain packages without importing them
files_using_os=$(grep -r "os\." --include="*.go" . | cut -d: -f1 | sort -u)
for file in $files_using_os; do
    if ! grep -q '"os"' "$file"; then
        echo -e "${YELLOW}⚠️  $file uses os. but doesn't import os${NC}"
        missing_imports=$((missing_imports + 1))
    fi
done

files_using_filepath=$(grep -r "filepath\." --include="*.go" . | cut -d: -f1 | sort -u)
for file in $files_using_filepath; do
    if ! grep -q '"path/filepath"' "$file"; then
        echo -e "${YELLOW}⚠️  $file uses filepath. but doesn't import path/filepath${NC}"
        missing_imports=$((missing_imports + 1))
    fi
done

if [[ $missing_imports -eq 0 ]]; then
    echo -e "${GREEN}✅ No obvious missing import issues${NC}"
fi

# 6. Check go.mod for consistency
echo -e "${BLUE}6. Checking go.mod...${NC}"
if [[ -f "go.mod" ]]; then
    if grep -q "gocloud.dev" go.mod; then
        echo -e "${YELLOW}⚠️  go.mod still contains gocloud.dev dependency${NC}"
    else
        echo -e "${GREEN}✅ go.mod looks clean${NC}"
    fi
else
    echo -e "${RED}❌ go.mod not found${NC}"
fi

# Summary
echo ""
echo -e "${BLUE}📊 Summary${NC}"
echo "=========="
total_issues=$((empty_imports + unused_imports + undefined_refs + type_issues + missing_imports))

if [[ $total_issues -eq 0 ]]; then
    echo -e "${GREEN}🎉 No obvious compilation issues found!${NC}"
    echo -e "${GREEN}The code should compile successfully.${NC}"
else
    echo -e "${YELLOW}⚠️  Found $total_issues potential issues${NC}"
    echo -e "${YELLOW}These might cause compilation errors.${NC}"
fi

echo ""
echo -e "${BLUE}💡 To test compilation:${NC}"
echo "make build"
echo ""
echo -e "${BLUE}💡 To test chart generation:${NC}"
echo "./scripts/generate-charts.sh all"
