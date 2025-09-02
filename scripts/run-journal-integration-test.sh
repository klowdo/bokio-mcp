#!/bin/bash

# Bokio MCP Journal Integration Test Runner
# This script helps you run the journal integration tests with proper setup

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_header() {
    echo -e "\n${BLUE}===============================================${NC}"
    echo -e "${BLUE} $1${NC}"
    echo -e "${BLUE}===============================================${NC}\n"
}

# Check if we're in the right directory
if [[ ! -f "journal_integration_test.go" ]]; then
    print_error "Please run this script from the bokio-mcp project root directory"
    exit 1
fi

print_header "Bokio MCP Journal Integration Test Runner"

# Check for required environment variables
REQUIRED_VARS=("BOKIO_CLIENT_ID" "BOKIO_CLIENT_SECRET")
MISSING_VARS=()

for var in "${REQUIRED_VARS[@]}"; do
    if [[ -z "${!var:-}" ]]; then
        MISSING_VARS+=("$var")
    fi
done

# Configuration options
READ_ONLY=${BOKIO_READ_ONLY:-false}
SKIP_AUTH=${SKIP_AUTH_TESTS:-false}
BASE_URL=${BOKIO_BASE_URL:-https://api.bokio.se}

print_info "Configuration:"
print_info "  Base URL: $BASE_URL"
print_info "  Read-only mode: $READ_ONLY"
print_info "  Skip auth tests: $SKIP_AUTH"

if [[ ${#MISSING_VARS[@]} -gt 0 ]] && [[ "$SKIP_AUTH" != "true" ]]; then
    print_warning "Missing required environment variables:"
    for var in "${MISSING_VARS[@]}"; do
        print_warning "  - $var"
    done
    echo
    print_info "Options:"
    print_info "1. Set the missing environment variables and re-run"
    print_info "2. Run with SKIP_AUTH_TESTS=true for client setup tests only"
    print_info "3. Create a .env file with your credentials (see .env.example)"
    echo
    print_info "Example:"
    print_info "  export BOKIO_CLIENT_ID=\"your_client_id\""
    print_info "  export BOKIO_CLIENT_SECRET=\"your_client_secret\""
    print_info "  $0"
    echo
    print_info "Or for setup tests only:"
    print_info "  SKIP_AUTH_TESTS=true $0"
    exit 1
fi

# Check if nix is available
if command -v nix &> /dev/null; then
    print_success "Using Nix development environment"
    RUN_CMD="nix develop --command bash -c"
else
    print_warning "Nix not available, using system Go"
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.24+ or use Nix."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ "$GO_VERSION" < "1.24" ]]; then
        print_warning "Go version $GO_VERSION detected. Go 1.24+ recommended."
    fi
    RUN_CMD="bash -c"
fi

print_header "Running Integration Tests"

# Test selection menu
echo "Select tests to run:"
echo "1) All integration tests (requires authentication)"
echo "2) Client configuration test only"
echo "3) Authentication flow test only"
echo "4) Journal listing tests only"
echo "5) Specific journal entry test only"
echo "6) Read-only mode validation"
echo "7) Error handling tests"
echo "8) Performance/rate limiting tests"
echo "9) Custom test pattern"

read -p "Enter your choice (1-9): " choice

case $choice in
    1)
        TEST_PATTERN="TestJournalIntegrationSuite"
        print_info "Running all integration tests..."
        ;;
    2)
        TEST_PATTERN="TestJournalIntegrationSuite/Test_01_ClientConfiguration"
        export SKIP_AUTH_TESTS=true
        print_info "Running client configuration test only..."
        ;;
    3)
        TEST_PATTERN="TestJournalIntegrationSuite/Test_02_AuthenticationFlow"
        print_info "Running authentication flow test only..."
        ;;
    4)
        TEST_PATTERN="TestJournalIntegrationSuite/Test_03_ListJournalEntries"
        print_info "Running journal listing tests only..."
        ;;
    5)
        TEST_PATTERN="TestJournalIntegrationSuite/Test_04_GetSpecificJournalEntry"
        print_info "Running specific journal entry test only..."
        ;;
    6)
        TEST_PATTERN="TestJournalIntegrationSuite/Test_06_ReadOnlyModeValidation"
        export BOKIO_READ_ONLY=true
        print_info "Running read-only mode validation (enabling read-only mode)..."
        ;;
    7)
        TEST_PATTERN="TestJournalIntegrationSuite/Test_07_ErrorHandling"
        print_info "Running error handling tests only..."
        ;;
    8)
        TEST_PATTERN="TestJournalIntegrationSuite/Test_08_RateLimitingAndRetries"
        print_info "Running performance/rate limiting tests only..."
        ;;
    9)
        read -p "Enter custom test pattern: " TEST_PATTERN
        print_info "Running custom test pattern: $TEST_PATTERN"
        ;;
    *)
        print_error "Invalid choice"
        exit 1
        ;;
esac

# Authentication instructions if needed
if [[ "$choice" == "3" ]] || [[ "$choice" == "1" && "$SKIP_AUTH" != "true" ]]; then
    print_header "Authentication Instructions"
    print_warning "Authentication tests require manual OAuth2 authorization:"
    print_info "1. The test will display an authorization URL"
    print_info "2. Open the URL in your browser and authorize the app"
    print_info "3. Copy the 'code' parameter from the redirect URL"
    print_info "4. Set TEST_AUTH_CODE environment variable with that code"
    print_info "5. Re-run the test"
    echo
    read -p "Press Enter to continue..."
fi

# Export environment variables
export BOKIO_CLIENT_ID="${BOKIO_CLIENT_ID:-}"
export BOKIO_CLIENT_SECRET="${BOKIO_CLIENT_SECRET:-}"
export BOKIO_BASE_URL="$BASE_URL"
export BOKIO_READ_ONLY="${BOKIO_READ_ONLY:-$READ_ONLY}"
export SKIP_AUTH_TESTS="${SKIP_AUTH_TESTS:-$SKIP_AUTH}"

# Additional test configuration
export TEST_FROM_DATE="${TEST_FROM_DATE:-2024-01-01}"
export TEST_TO_DATE="${TEST_TO_DATE:-2024-12-31}"
export TEST_ACCOUNT_ID="${TEST_ACCOUNT_ID:-1930}"
export EXPECTED_MIN_JOURNAL_ENTRIES="${EXPECTED_MIN_JOURNAL_ENTRIES:-0}"

print_info "Starting tests..."
echo

# Run the test
if $RUN_CMD "go test -v -run \"$TEST_PATTERN\""; then
    echo
    print_success "Tests completed successfully!"
    
    # Show next steps if authentication is needed
    if [[ "$choice" == "1" ]] && [[ -z "${TEST_AUTH_CODE:-}" ]]; then
        echo
        print_header "Next Steps"
        print_info "To complete the OAuth2 authentication flow:"
        print_info "1. Set TEST_AUTH_CODE with your authorization code"
        print_info "2. Re-run the authentication test:"
        print_info "   TEST_AUTH_CODE=your_code $0"
    fi
else
    echo
    print_error "Tests failed. Check the output above for details."
    exit 1
fi

print_header "Test Summary"
print_success "Integration test execution completed"

if [[ "$READ_ONLY" == "true" ]]; then
    print_info "✓ Read-only mode was enabled - no write operations performed"
fi

if [[ "$SKIP_AUTH" == "true" ]]; then
    print_info "✓ Authentication was skipped - client setup tests only"
fi

print_info "For more testing options, see JOURNAL_INTEGRATION_TEST.md"