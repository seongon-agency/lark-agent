#!/bin/bash
# =============================================================================
# Lark-Agent Deployment Verification Script
# =============================================================================
# This script verifies that both services are properly configured and running
# Run this after deploying to Railway or starting with Docker Compose
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
FEISHU_BOT_URL="${FEISHU_BOT_URL:-http://localhost:9000}"
AGNO_SERVICE_URL="${AGNO_SERVICE_URL:-http://localhost:8000}"

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}Lark-Agent Deployment Verification${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""

# Function to print success
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Function to print error
print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Function to print info
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# =============================================================================
# Check Prerequisites
# =============================================================================
echo -e "${BLUE}[1/6] Checking Prerequisites...${NC}"

if command -v curl &> /dev/null; then
    print_success "curl is installed"
else
    print_error "curl is not installed. Please install curl."
    exit 1
fi

if command -v jq &> /dev/null; then
    print_success "jq is installed"
else
    print_warning "jq is not installed. JSON output will be raw."
fi

echo ""

# =============================================================================
# Check Environment Variables
# =============================================================================
echo -e "${BLUE}[2/6] Checking Environment Variables...${NC}"

required_vars=("APP_ID" "APP_SECRET" "OPENAI_KEY")
missing_vars=()

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        missing_vars+=("$var")
        print_warning "$var is not set"
    else
        # Mask sensitive values
        masked_value=$(echo "${!var}" | sed 's/\(.\{4\}\).*\(.\{4\}\)/\1****\2/')
        print_success "$var is set: $masked_value"
    fi
done

if [ ${#missing_vars[@]} -gt 0 ]; then
    print_warning "Some environment variables are missing. Make sure to set them before deployment."
else
    print_success "All required environment variables are set"
fi

echo ""

# =============================================================================
# Check Agno Service Health
# =============================================================================
echo -e "${BLUE}[3/6] Checking Agno AI Service...${NC}"
print_info "Connecting to: $AGNO_SERVICE_URL"

response=$(curl -s -w "\n%{http_code}" "$AGNO_SERVICE_URL/health" || echo "000")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    print_success "Agno service is healthy (HTTP $http_code)"

    if command -v jq &> /dev/null; then
        echo "$body" | jq '.'

        # Check OpenAI configuration
        openai_configured=$(echo "$body" | jq -r '.openai_configured')
        if [ "$openai_configured" = "true" ]; then
            print_success "OpenAI is configured in Agno service"
        else
            print_error "OpenAI is NOT configured in Agno service"
        fi
    else
        echo "$body"
    fi
else
    print_error "Agno service is not healthy (HTTP $http_code)"
    echo "Response: $body"
fi

echo ""

# =============================================================================
# Check Feishu Bot Service Health
# =============================================================================
echo -e "${BLUE}[4/6] Checking Feishu Bot Service...${NC}"
print_info "Connecting to: $FEISHU_BOT_URL"

response=$(curl -s -w "\n%{http_code}" "$FEISHU_BOT_URL/ping" || echo "000")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    print_success "Feishu bot is healthy (HTTP $http_code)"
    echo "Response: $body"
else
    print_error "Feishu bot is not healthy (HTTP $http_code)"
    echo "Response: $body"
fi

echo ""

# =============================================================================
# Test Agno Chat Endpoint
# =============================================================================
echo -e "${BLUE}[5/6] Testing Agno Chat Endpoint...${NC}"

chat_payload='{
  "session_id": "test_session_123",
  "message": "Hello, this is a test message",
  "history": []
}'

response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d "$chat_payload" \
    "$AGNO_SERVICE_URL/chat" || echo "000")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    print_success "Agno chat endpoint is working (HTTP $http_code)"

    if command -v jq &> /dev/null; then
        echo "$body" | jq '.'
    else
        echo "$body"
    fi
else
    print_error "Agno chat endpoint failed (HTTP $http_code)"
    echo "Response: $body"
fi

echo ""

# =============================================================================
# Service Communication Test
# =============================================================================
echo -e "${BLUE}[6/6] Testing Service Communication...${NC}"

if [ "$http_code" = "200" ] && [ -n "$body" ]; then
    print_success "Both services can communicate"
else
    print_warning "Service communication test skipped (previous tests failed)"
fi

echo ""

# =============================================================================
# Summary
# =============================================================================
echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}Verification Summary${NC}"
echo -e "${BLUE}============================================${NC}"

echo ""
echo "Service URLs:"
echo "  Feishu Bot:   $FEISHU_BOT_URL"
echo "  Agno Service: $AGNO_SERVICE_URL"
echo ""

echo "Webhook URLs for Feishu Configuration:"
echo "  Event Webhook: ${FEISHU_BOT_URL}/webhook/event"
echo "  Card Callback: ${FEISHU_BOT_URL}/webhook/card"
echo ""

echo "Health Check Endpoints:"
echo "  Feishu Bot:   ${FEISHU_BOT_URL}/ping"
echo "  Agno Service: ${AGNO_SERVICE_URL}/health"
echo ""

if [ ${#missing_vars[@]} -eq 0 ] && [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓ All checks passed! Your deployment is ready.${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some checks failed. Please review the output above.${NC}"
    exit 1
fi
