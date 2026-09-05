#!/bin/bash

# Hurl API Test Runner
# This script runs all Hurl test files for the payment gateway API

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Base URL for the API
BASE_URL="http://localhost:5000"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if hurl is installed
check_hurl() {
    if ! command -v hurl &> /dev/null; then
        print_error "Hurl is not installed. Please install it first:"
        echo "Visit: https://hurl.dev/docs/installation.html"
        exit 1
    fi
}

# Check if API is running
check_api() {
    print_status "Checking if API Gateway is running on $BASE_URL..."
    if curl -s --head --request GET "$BASE_URL/api/auth/hello" | grep "200 OK" > /dev/null; then
        print_success "API Gateway is running!"
    else
        print_error "API Gateway is not responding on $BASE_URL"
        print_warning "Please start the API Gateway first"
        exit 1
    fi
}

# Obtain an auth token by registering + logging in a test user
# NOTE: stdout must contain ONLY the token (it is captured via $(...)),
# so all status messages are sent to stderr.
obtain_auth_token() {
    print_status "Registering test user..." >&2
    curl -s --max-time 10 -X POST "$BASE_URL/api/auth/register" \
        -H "Content-Type: application/json" \
        -d '{"firstname":"John","lastname":"Doe","email":"john.doe@hellodota.com","password":"password123","confirm_password":"password123"}' \
        > /dev/null 2>&1 || true

    print_status "Logging in to obtain auth_token..." >&2
    local login_response
    login_response=$(curl -s --max-time 10 -X POST "$BASE_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"email":"john.doe@hellodota.com","password":"password123"}') || true

    local token=""
    if command -v python3 >/dev/null 2>&1; then
        token=$(echo "$login_response" | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("data",{}).get("access_token",""))' 2>/dev/null || echo "")
    fi
    if [[ -z "$token" ]]; then
        token=$(echo "$login_response" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')
    fi

    if [[ -z "$token" ]]; then
        print_error "Failed to obtain auth_token"
        echo "$login_response" | head -5
        exit 1
    fi
    echo "$token"
}

# Run a single test file with the auth token injected
# Capture hurl output (stdout+stderr) so failures show the real error instead of a bare summary.
run_test_file() {
    local file=$1
    local output
    print_status "Running tests in $file..."
    
    if output=$(hurl --variable "auth_token=$AUTH_TOKEN" --variable "unique_suffix=$UNIQUE_SUFFIX" --variable "year=$YEAR" "$file" 2>&1); then
        print_success "All tests in $file passed!"
        return 0
    else
        print_error "Some tests in $file failed!"
        echo "$output" | head -30
        return 1
    fi
}

# Main execution
main() {
    echo "========================================"
    echo "  Payment Gateway API Test Runner"
    echo "========================================"
    echo

    # Check prerequisites
    check_hurl
    check_api

    echo
    print_status "Starting API tests..."
    echo

    # Get an auth token for authenticated endpoints
    AUTH_TOKEN=$(obtain_auth_token)
    print_success "Auth token obtained"

    # Unique suffix so each run creates fresh, non-colliding test data
    UNIQUE_SUFFIX=$(date +%s%N)

    # Current year for stats query coverage (?year=...)
    YEAR=$(date +%Y)

    # Get all .hurl files
    test_files=( *.hurl )
    failed_tests=()
    passed_tests=()

    # Run each test file
    for file in "${test_files[@]}"; do
        if [[ -f "$file" ]]; then
            echo "----------------------------------------"
            if run_test_file "$file"; then
                passed_tests+=("$file")
            else
                failed_tests+=("$file")
            fi
            echo
        fi
    done

    # Summary
    echo "========================================"
    echo "  Test Summary"
    echo "========================================"
    echo

    if [[ ${#passed_tests[@]} -gt 0 ]]; then
        print_success "Passed tests (${#passed_tests[@]}):"
        for file in "${passed_tests[@]}"; do
            echo "  ✓ $file"
        done
        echo
    fi

    if [[ ${#failed_tests[@]} -gt 0 ]]; then
        print_error "Failed tests (${#failed_tests[@]}):"
        for file in "${failed_tests[@]}"; do
            echo "  ✗ $file"
        done
        echo
    fi

    # Exit with appropriate code
    if [[ ${#failed_tests[@]} -eq 0 ]]; then
        print_success "All tests passed! 🎉"
        exit 0
    else
        print_error "Some tests failed. Please check the output above."
        exit 1
    fi
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [options]"
        echo
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --check         Only check prerequisites (hurl installation and API status)"
        echo
        echo "Examples:"
        echo "  $0              # Run all tests"
        echo "  $0 --check      # Check prerequisites only"
        exit 0
        ;;
    --check)
        check_hurl
        check_api
        print_success "All prerequisites met!"
        exit 0
        ;;
    "")
        main
        ;;
    *)
        print_error "Unknown option: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac