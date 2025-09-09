#!/bin/bash

# API Route Status Analyzer
# Usage: ./analyze_routes.sh [test_output_file]

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Input file (default to paste-30.txt or from argument)
INPUT_FILE="${1:-paste-30.txt}"

if [[ ! -f "$INPUT_FILE" ]]; then
    echo -e "${RED}❌ Error: File '$INPUT_FILE' not found!${NC}"
    echo "Usage: $0 [test_output_file]"
    exit 1
fi

echo -e "${BLUE}🔍 Analyzing API Route Test Results from: $INPUT_FILE${NC}\n"

# Create temporary files
TEMP_TESTS="/tmp/route_tests.tmp"
TEMP_STATUS="/tmp/route_status.tmp"
TEMP_RESULTS="/tmp/route_results.tmp"

# Clean up temp files on exit
trap 'rm -f $TEMP_TESTS $TEMP_STATUS $TEMP_RESULTS' EXIT

# Extract test names (lines starting with 📋)
grep "^📋" "$INPUT_FILE" | sed 's/📋 Testing: //' | sed 's/📋 //' > "$TEMP_TESTS"

# Extract status indicators (✅ or ❌)
grep -E "^[✅❌]|✅|❌" "$INPUT_FILE" | \
    sed 's/.*✅.*/PASS/' | \
    sed 's/.*❌.*/FAIL/' > "$TEMP_STATUS"

# Combine test names with their status
paste "$TEMP_TESTS" "$TEMP_STATUS" > "$TEMP_RESULTS"

# Calculate statistics
TOTAL_TESTS=$(wc -l < "$TEMP_RESULTS")
PASSED_TESTS=$(grep -c "PASS" "$TEMP_RESULTS" || echo 0)
FAILED_TESTS=$(grep -c "FAIL" "$TEMP_RESULTS" || echo 0)

# Create formatted table
echo -e "${YELLOW}📊 API Route Test Results Summary${NC}"
echo "=============================================="
printf "%-50s | %-6s\n" "API Route Test" "Status"
echo "------------------------------------------------|-------"

# Read results and format table
while IFS=$'\t' read -r test_name status; do
    if [[ "$status" == "PASS" ]]; then
        printf "%-50s | ${GREEN}%-6s${NC}\n" "$test_name" "✅ PASS"
    elif [[ "$status" == "FAIL" ]]; then
        printf "%-50s | ${RED}%-6s${NC}\n" "$test_name" "❌ FAIL"
    else
        printf "%-50s | ${YELLOW}%-6s${NC}\n" "$test_name" "⚠️ UNKNOWN"
    fi
done < "$TEMP_RESULTS"

echo "------------------------------------------------|-------"

# Display summary statistics
echo -e "\n${BLUE}📈 Test Statistics:${NC}"
echo "  Total Tests: $TOTAL_TESTS"
echo -e "  Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "  Failed: ${RED}$FAILED_TESTS${NC}"

if [[ $TOTAL_TESTS -gt 0 ]]; then
    SUCCESS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo "  Success Rate: ${SUCCESS_RATE}%"
fi

# Save detailed results to CSV
CSV_FILE="api_route_results_$(date +%Y%m%d_%H%M%S).csv"
echo "Test Name,Status" > "$CSV_FILE"
while IFS=$'\t' read -r test_name status; do
    echo "\"$test_name\",$status" >> "$CSV_FILE"
done < "$TEMP_RESULTS"

echo -e "\n💾 Detailed results saved to: ${YELLOW}$CSV_FILE${NC}"

# Show only failed tests if any exist
if [[ $FAILED_TESTS -gt 0 ]]; then
    echo -e "\n${RED}🚨 Failed Tests Only:${NC}"
    echo "========================="
    while IFS=$'\t' read -r test_name status; do
        if [[ "$status" == "FAIL" ]]; then
            echo -e "${RED}❌${NC} $test_name"
        fi
    done < "$TEMP_RESULTS"
fi
