# Symphony Remote Agent E2E Tests Makefile

.PHONY: help test test-http test-mqtt test-cert test-reconnect clean deps check

# Default target
help:
	@echo "Symphony Remote Agent E2E Tests"
	@echo "================================"
	@echo ""
	@echo "Available targets:"
	@echo "  help           Show this help message"
	@echo "  deps           Download Go dependencies"
	@echo "  check          Check prerequisites"
	@echo "  test           Run all tests"
	@echo "  test-http      Run HTTP communication tests"
	@echo "  test-mqtt      Run MQTT communication tests"
	@echo "  test-cert      Run MQTT certificate authentication tests"
	@echo "  test-reconnect Run MQTT reconnection tests"
	@echo "  clean          Clean up test resources"
	@echo ""
	@echo "Examples:"
	@echo "  make test           # Run all tests"
	@echo "  make test-http      # Run only HTTP tests"
	@echo "  make test-mqtt      # Run only MQTT tests"
	@echo "  make deps           # Download dependencies"

# Check prerequisites
check:
	@echo "Checking prerequisites..."
	@which go > /dev/null || (echo "Error: Go not found" && exit 1)
	@which docker > /dev/null || (echo "Error: Docker not found" && exit 1)
	@which kubectl > /dev/null || (echo "Error: kubectl not found" && exit 1)
	@kubectl cluster-info > /dev/null || (echo "Error: Cannot connect to Kubernetes" && exit 1)
	@echo "All prerequisites satisfied"

# Download dependencies
deps:
	@echo "Downloading Go dependencies..."
	@go mod tidy
	@echo "Dependencies updated"

# Run all tests
test: deps
	@echo "Running all E2E tests..."
	@go test -v -timeout 15m ./...

# Run HTTP communication tests
test-http: deps
	@echo "Running HTTP communication tests..."
	@go test -v -timeout 10m ./scenarios/http-communication/

# Run MQTT communication tests
test-mqtt: deps
	@echo "Running MQTT communication tests..."
	@go test -v -timeout 10m ./scenarios/mqtt-communication/

# Run MQTT certificate authentication tests
test-cert: deps
	@echo "Running MQTT certificate authentication tests..."
	@go test -v -timeout 5m -run TestMQTTCertificateAuthentication ./scenarios/mqtt-communication/

# Run MQTT reconnection tests
test-reconnect: deps
	@echo "Running MQTT reconnection tests..."
	@go test -v -timeout 5m -run TestMQTTReconnection ./scenarios/mqtt-communication/

# Clean up test resources
clean:
	@echo "Cleaning up test resources..."
	@kubectl delete namespace test-http-ns test-mqtt-ns test-reconnect-ns 2>/dev/null || true
	@docker ps -q --filter "name=test-mqtt-broker" | xargs -r docker stop 2>/dev/null || true
	@docker ps -aq --filter "name=test-mqtt-broker" | xargs -r docker rm 2>/dev/null || true
	@echo "Cleanup completed"

# Advanced targets with custom parameters

# Run tests with verbose output
test-verbose: deps
	@echo "Running all tests with verbose output..."
	@go test -v -timeout 15m ./...

# Run tests with custom timeout
test-long: deps
	@echo "Running all tests with extended timeout..."
	@go test -v -timeout 30m ./...

# Run only quick tests (no Docker dependencies)
test-quick: deps
	@echo "Running quick tests (HTTP only)..."
	@go test -v -timeout 5m ./scenarios/http-communication/

# Development helpers
dev-setup: check deps
	@echo "Development environment setup complete"

# CI/CD target
ci-test: deps
	@echo "Running tests in CI mode..."
	@go test -timeout 20m ./...

# Test with coverage
test-coverage: deps
	@echo "Running tests with coverage..."
	@go test -v -timeout 15m -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format and lint
fmt:
	@echo "Formatting Go code..."
	@go fmt ./...

# Verify tests compile
verify: deps
	@echo "Verifying tests compile..."
	@go test -c ./scenarios/http-communication/
	@go test -c ./scenarios/mqtt-communication/
	@rm -f *.test
	@echo "All tests compile successfully"
