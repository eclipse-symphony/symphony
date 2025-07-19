#!/bin/bash

# Symphony Remote Agent E2E Test Runner
# This script provides an easy way to run the end-to-end tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Default values
TEST_TYPE="all"
VERBOSE=""
TIMEOUT="25m"  # Increased timeout for minikube startup and Symphony deployment
DOCKER_CHECK=true
KUBE_CHECK=true

# Function to print colored output
print_info() {
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

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -t, --type TYPE         Test type: all, http, mqtt, cert, reconnect (default: all)"
    echo "  -v, --verbose           Enable verbose output"
    echo "  -T, --timeout DURATION  Test timeout (default: 10m)"
    echo "  --skip-docker-check     Skip Docker availability check"
    echo "  --skip-kube-check       Skip Kubernetes connectivity check"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                      # Run all tests"
    echo "  $0 -t http -v           # Run HTTP tests with verbose output"
    echo "  $0 -t mqtt -T 15m       # Run MQTT tests with 15 minute timeout"
    echo "  $0 --skip-docker-check  # Run tests without checking Docker"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            TEST_TYPE="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE="-v"
            shift
            ;;
        -T|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --skip-docker-check)
            DOCKER_CHECK=false
            shift
            ;;
        --skip-kube-check)
            KUBE_CHECK=false
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    print_success "Go found: $(go version)"
    
    # Check minikube (required for fresh cluster creation)
    if ! command -v minikube &> /dev/null; then
        print_error "minikube is not installed or not in PATH"
        print_warning "Tests require minikube for fresh cluster creation"
        exit 1
    fi
    print_success "minikube found: $(minikube version --short)"
    
    # Check Docker (required for minikube driver)
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        print_warning "Tests require Docker as minikube driver"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running"
        print_warning "Please start Docker daemon before running tests"
        exit 1
    fi
    print_success "Docker found and running"
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    print_success "kubectl found: $(kubectl version --client --short)"
    
    # Note: We don't check existing cluster connection since tests create fresh minikube clusters
    print_info "Tests will create fresh minikube clusters automatically"
}

# Function to initialize Go modules
init_modules() {
    print_info "Initializing Go modules..."
    if ! go mod tidy; then
        print_error "Failed to download Go dependencies"
        exit 1
    fi
    print_success "Go modules initialized"
}

# Function to run tests
run_tests() {
    local test_path=""
    local test_name=""
    
    case "$TEST_TYPE" in
        "all")
            test_path="./..."
            test_name="All tests"
            ;;
        "http")
            test_path="./scenarios/http-communication/"
            test_name="HTTP communication tests"
            ;;
        "mqtt")
            test_path="./scenarios/mqtt-communication/"
            test_name="MQTT communication tests"
            ;;
        "cert")
            test_path="./scenarios/mqtt-communication/"
            test_name="MQTT certificate authentication test"
            VERBOSE="$VERBOSE -run TestMQTTCertificateAuthentication"
            ;;
        "reconnect")
            test_path="./scenarios/mqtt-communication/"
            test_name="MQTT reconnection test"
            VERBOSE="$VERBOSE -run TestMQTTReconnection"
            ;;
        *)
            print_error "Invalid test type: $TEST_TYPE"
            print_warning "Valid types: all, http, mqtt, cert, reconnect"
            exit 1
            ;;
    esac
    
    print_info "Running $test_name..."
    print_info "Command: go test $VERBOSE -timeout $TIMEOUT $test_path"
    echo ""
    
    # Run the tests
    if go test $VERBOSE -timeout "$TIMEOUT" $test_path; then
        echo ""
        print_success "$test_name completed successfully!"
    else
        echo ""
        print_error "$test_name failed!"
        exit 1
    fi
}

# Function to cleanup (if needed)
cleanup() {
    print_info "Cleaning up any remaining test resources..."
    
    # Clean up any test namespaces
    kubectl delete namespace test-http-ns test-mqtt-ns test-reconnect-ns 2>/dev/null || true
    
    # Clean up any test MQTT broker containers
    docker ps -q --filter "name=test-mqtt-broker" | xargs -r docker stop 2>/dev/null || true
    docker ps -aq --filter "name=test-mqtt-broker" | xargs -r docker rm 2>/dev/null || true
    
    print_success "Cleanup completed"
}

# Main execution
main() {
    print_info "Symphony Remote Agent E2E Test Runner"
    print_info "====================================="
    echo ""
    
    # Check prerequisites
    check_prerequisites
    echo ""
    
    # Initialize modules
    init_modules
    echo ""
    
    # Run tests
    run_tests
    echo ""
    
    # Optional cleanup
    if [[ "$TEST_TYPE" != "all" ]]; then
        read -p "Do you want to cleanup test resources? (y/N): " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            cleanup
            echo ""
        fi
    fi
    
    print_success "Test execution completed!"
}

# Trap to cleanup on exit
trap cleanup EXIT

# Run main function
main "$@"
