# Eye in the Sky - Makefile for testing and CI

.PHONY: test test-verbose test-coverage test-race clean build lint help

# Default target
all: clean lint test build

# Run all tests
test:
	@echo "🧪 Running all tests..."
	@go test ./... -short

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running tests with verbose output..."
	@go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	@go test ./... -cover -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📄 Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	@echo "🏃 Running tests with race detection..."
	@go test ./... -race

# Run tests for specific package
test-db:
	@echo "🗄️ Testing database package..."
	@go test ./internal/database -v

test-mcp:
	@echo "🔧 Testing MCP package..."
	@go test ./internal/mcp -v

test-ui:
	@echo "🎨 Testing UI package..."
	@go test ./internal/ui/app -v

# Run integration tests
test-integration:
	@echo "🔗 Running integration tests..."
	@go test ./... -tags=integration -v

# Clean test artifacts
clean:
	@echo "🧹 Cleaning test artifacts..."
	@rm -f coverage.out coverage.html
	@rm -f ./internal/database/*.db
	@rm -f ./internal/mcp/*.db
	@rm -f ./cmd/server/*.db
	@rm -f ./*.db

# Build the application
build:
	@echo "🔨 Building application..."
	@./scripts/build.sh

# Lint the code
lint:
	@echo "🔍 Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed, running go vet instead"; \
		go vet ./...; \
	fi

# Format the code
fmt:
	@echo "✨ Formatting code..."
	@go fmt ./...

# Tidy dependencies
tidy:
	@echo "📦 Tidying dependencies..."
	@go mod tidy

# Run benchmarks
bench:
	@echo "⚡ Running benchmarks..."
	@go test ./... -bench=. -benchmem

# Security scan
security:
	@echo "🔒 Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not installed, skipping security scan"; \
	fi

# Install development tools
dev-tools:
	@echo "🛠️ Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Run CI pipeline
ci: clean lint test-race test-coverage
	@echo "✅ CI pipeline completed"

# Quick development check
check: fmt tidy lint test
	@echo "✅ Development check completed"

# Help
help:
	@echo "Eye in the Sky - Available Commands:"
	@echo ""
	@echo "Testing:"
	@echo "  test              Run all tests"
	@echo "  test-verbose      Run tests with verbose output"
	@echo "  test-coverage     Run tests with coverage report"
	@echo "  test-race         Run tests with race detection"
	@echo "  test-db          Test database package only"
	@echo "  test-mcp         Test MCP package only"
	@echo "  test-integration  Run integration tests"
	@echo "  bench            Run benchmarks"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint             Run linter"
	@echo "  fmt              Format code"
	@echo "  tidy             Tidy dependencies"
	@echo "  security         Run security scan"
	@echo ""
	@echo "Build & CI:"
	@echo "  build            Build application"
	@echo "  clean            Clean test artifacts"
	@echo "  ci               Run full CI pipeline"
	@echo "  check            Quick development check"
	@echo ""
	@echo "Setup:"
	@echo "  dev-tools        Install development tools"
	@echo "  help             Show this help"