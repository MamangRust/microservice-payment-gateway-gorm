#!/bin/bash

# Docker Build Script for Payment Gateway Services
# This script builds Docker images for all services.
#
# IMPORTANT: Build context is the PROJECT ROOT, not the service directory.
# The Dockerfiles COPY shared/, pb/, pkg/, service/ from the repository root,
# so building from the service dir will fail.

set -e

# Resolve the project root (directory containing this script)
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🐳 Building Docker images for Payment Gateway services..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Number of images to build concurrently
BUILD_CONCURRENCY="${BUILD_CONCURRENCY:-4}"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if docker is installed and running
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! docker info &> /dev/null; then
        print_error "Docker is not running. Please start Docker."
        exit 1
    fi
    print_status "Docker is installed and running"
}

# Build Docker image for a service.
# Writes a per-service log into $1 (tmp dir) to keep parallel output readable.
build_service_image() {
    local service="$1"
    local log_dir="$2"

    if [ -z "$service" ]; then
        print_error "Service name is required"
        return 1
    fi

    local service_dir="${ROOT_DIR}/service/${service}"
    local dockerfile="${service_dir}/Dockerfile"
    local image="microservice-payment-gateway-grpc/${service}:latest"
    local log_file="${log_dir}/${service}.log"

    print_status "Building ${service} service image..."

    if [ ! -d "$service_dir" ]; then
        print_error "Service directory not found: $service_dir"
        return 1
    fi

    if [ ! -f "$dockerfile" ]; then
        print_error "Dockerfile not found: $dockerfile"
        return 1
    fi

    # Go services must have a go.mod; Python services (e.g. ai-security) don't.
    if [ ! -f "${service_dir}/go.mod" ] && [ ! -f "${service_dir}/requirements.txt" ]; then
        print_error "Neither go.mod nor requirements.txt found in service: $service_dir"
        return 1
    fi

    # Build with the PROJECT ROOT as context so COPY shared/, pb/, pkg/, service/ works.
    # Retry once on failure: module downloads can die on transient network resets.
    local attempts=0
    local max_attempts=2
    local ok=0

    while [ "$attempts" -lt "$max_attempts" ]; do
        attempts=$((attempts + 1))
        if docker build \
            --progress=plain \
            -f "$dockerfile" \
            -t "$image" \
            "$ROOT_DIR" > "$log_file" 2>&1; then
            ok=1
            break
        fi
        if [ "$attempts" -lt "$max_attempts" ]; then
            print_warning "Retrying ${service} (attempt ${attempts}/${max_attempts} failed)..."
            sleep 3
        fi
    done

    if [ "$ok" -eq 1 ]; then
        print_status "Successfully built ${image}"
        return 0
    fi

    print_error "Failed to build ${image}"
    if [ -f "$log_file" ]; then
        echo "--- last 15 lines of ${service}.log ---"
        tail -15 "$log_file"
    fi
    return 1
}

export -f build_service_image
export -f print_status
export -f print_warning
export -f print_error
export ROOT_DIR

# Build all service images
build_all_images() {
    print_status "Building all service images (concurrency: ${BUILD_CONCURRENCY})..."

    # List of services to build (must match the services in docker-compose.yml)
    services=(
        "auth" "user" "card" "merchant" "role" "saldo"
        "topup" "transaction" "transfer" "withdraw" "apigateway"
        "email" "stats-reader" "stats-writer" "ai-security"
    )

    local failed_builds=0

    # Build images in parallel, then collect results.
    # NOTE: fail.$$ is expanded INSIDE each bash -c child (unique PID per child)
    # so parallel failures never overwrite each other.
    local tmp_dir
    tmp_dir="$(mktemp -d)"
    TMP_DIR="$tmp_dir"
    if ! printf '%s\n' "${services[@]}" | xargs -P "${BUILD_CONCURRENCY}" -I {} \
        bash -c 'build_service_image "$1" "$2" || echo "$1" > "$2/fail.$$"' _ {} "${tmp_dir}"; then
        print_warning "⚠️  xargs pipeline failed; some services may not have been built"
    fi

    # Collect failed services
    for fail_file in "${tmp_dir}"/fail.*; do
        if [ -f "$fail_file" ]; then
            while IFS= read -r service; do
                print_error "❌ ${service} image failed to build"
                ((failed_builds+=1))
            done < "$fail_file"
        fi
    done
    rm -rf "$tmp_dir"

    if [ $failed_builds -eq 0 ]; then
        print_status "🎉 All images built successfully!"
    else
        print_warning "⚠️  $failed_builds images failed to build"
        return 1
    fi
}

# Show built images
show_built_images() {
    print_status "Built Docker images:"
    echo ""
    docker images | grep -E "microservice-payment-gateway-grpc" | head -25
    echo ""
}

# Cleanup function
cleanup() {
    if [ -n "${TMP_DIR:-}" ] && [ -d "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR"
    fi
    print_status "Build process finished"
}

# Main execution
main() {
    # Set trap for cleanup
    trap cleanup EXIT

    # Run checks
    check_docker

    # Build all images.
    # NOTE: must stay in a conditional context - a bare call returning non-zero
    # would trip set -e and skip the summary below.
    rc=0
    if ! build_all_images; then
        rc=1
    fi

    # Show built images
    show_built_images

    if [ "$rc" -ne 0 ]; then
        exit "$rc"
    fi
    print_status "Docker build process completed! 🎉"
}

# Handle script interruption
trap 'print_error "Build interrupted"; exit 1' INT

# Run main function
main "$@"
