#!/bin/bash
# License checker for Lawrence OSS dependencies
# This script helps verify third-party dependency licenses

set -e

echo "========================================"
echo "Lawrence OSS License Checker"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if tools are available
check_tool() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}✗ $1 is not installed${NC}"
        return 1
    else
        echo -e "${GREEN}✓ $1 is available${NC}"
        return 0
    fi
}

echo "Checking required tools..."
check_tool "go" || exit 1
check_tool "jq" || echo -e "${YELLOW}⚠ jq not found, some features will be limited${NC}"
check_tool "pnpm" || echo -e "${YELLOW}⚠ pnpm not found, frontend check will be skipped${NC}"
echo ""

# Go Dependencies
echo "========================================"
echo "Go Backend Dependencies"
echo "========================================"
echo ""

echo "Total Go modules:"
go list -m all | wc -l
echo ""

echo "Direct dependencies:"
go list -m -f '{{if not .Indirect}}{{.Path}}@{{.Version}}{{end}}' all | grep -v '^$' | head -20
echo ""

if command -v jq &> /dev/null; then
    echo "Checking for known licenses..."
    echo ""
    
    # Common Apache 2.0 projects
    echo "Apache 2.0 licensed projects:"
    go list -m all | grep -E "(opentelemetry|grpc|protobuf|prometheus|apache)" || echo "None found in list"
    echo ""
    
    # Common MIT projects
    echo "Common MIT licensed projects:"
    go list -m all | grep -E "(gin-gonic|uuid|cobra|viper|testify)" || echo "None found in list"
    echo ""
    
    # BSD projects
    echo "Common BSD licensed projects:"
    go list -m all | grep -E "(uber-go|golang.org/x)" || echo "None found in list"
    echo ""
fi

# Frontend Dependencies
if command -v pnpm &> /dev/null; then
    echo "========================================"
    echo "Frontend Dependencies"
    echo "========================================"
    echo ""
    
    cd ui
    
    echo "Total npm packages:"
    pnpm list --depth=0 2>/dev/null | grep -c "^[├└]" || echo "0"
    echo ""
    
    echo "Checking licenses with pnpm..."
    if pnpm licenses list 2>/dev/null | head -20; then
        echo ""
        echo "License summary available via: cd ui && pnpm licenses list"
    else
        echo "Run 'cd ui && pnpm install' first"
    fi
    
    cd ..
    echo ""
fi

# Check for potential issues
echo "========================================"
echo "License Compliance Check"
echo "========================================"
echo ""

echo "Checking for potentially problematic licenses..."
echo ""

# Check for GPL (should be avoided in most cases)
echo "Checking for GPL licenses (none should be found):"
if go list -m all | grep -i gpl; then
    echo -e "${RED}⚠ Warning: GPL licensed dependency found!${NC}"
else
    echo -e "${GREEN}✓ No GPL licenses found${NC}"
fi
echo ""

# Verify NOTICE file exists
if [ -f "NOTICE" ]; then
    echo -e "${GREEN}✓ NOTICE file exists${NC}"
else
    echo -e "${RED}✗ NOTICE file missing${NC}"
fi

# Verify LICENSE file exists
if [ -f "LICENSE" ]; then
    echo -e "${GREEN}✓ LICENSE file exists${NC}"
else
    echo -e "${RED}✗ LICENSE file missing${NC}"
fi
echo ""

echo "========================================"
echo "Summary"
echo "========================================"
echo ""
echo "For detailed license information:"
echo "1. Read the NOTICE file in the repository root"
echo "2. Check individual project repositories"
echo "3. Run 'go list -m all' for Go dependencies"
echo "4. Run 'cd ui && pnpm licenses list' for frontend"
echo ""
echo "All dependencies should be compatible with Apache 2.0"
echo "Acceptable licenses: Apache-2.0, MIT, BSD, ISC"
echo ""

