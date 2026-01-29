# CMS-AI Server Makefile

.PHONY: test test-smart test-industry-themes test-unit build clean help

# Default target
help:
	@echo "🚀 CMS-AI Server Development Commands"
	@echo "======================================"
	@echo ""
	@echo "Testing:"
	@echo "  make test              Run all tests"
	@echo "  make test-unit         Run unit tests only"
	@echo "  make test-smart        Run smart features test suite"
	@echo "  make test-industry     Run industry themes test script"
	@echo ""
	@echo "Building:"
	@echo "  make build             Build the server binary"
	@echo "  make clean             Clean build artifacts"
	@echo ""

# Run all tests
test: test-unit test-smart

# Run unit tests
test-unit:
	@echo "🧪 Running unit tests..."
	@go test -v ./...

# Run smart features test suite
test-smart:
	@echo "🎨 Running smart features test suite..."
	@cd scripts && go run test_smart_features.go

# Run industry themes test script (like olama)
test-industry:
	@echo "🏭 Running industry-specific theme tests..."
	@./scripts/test_industry_themes.sh

# Test Python PPTX renderer with rich visuals
test-python:
	@echo "🐍 Testing Python PPTX Renderer with Rich Visuals..."
	@cd scripts && go run test_python_renderer.go

# Test olama AI integration
test-ai:
	@echo "🧠 Testing Olama AI Bridge Integration..."
	@cd scripts && go run test_olama_ai_bridge.go

# Test all renderers
test-all:
	@echo "🚀 Running All Renderer Tests..."
	@make test-smart
	@make test-python
	@make test-ai
	@echo "✅ All tests complete!"

# Build server binary
build:
	@echo "🔨 Building server..."
	@go build -o bin/server ./cmd/server

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf test_outputs/

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy