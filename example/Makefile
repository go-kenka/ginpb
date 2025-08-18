# Complete Example Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Binary names
SERVER_BINARY=complete_server
CLIENT_BINARY=test_client

# Paths
SERVER_PATH=./server
CLIENT_PATH=./client
API_PATH=./api

.PHONY: all build clean run-server run-client test-api proto help

# Default target
all: build

# Build all binaries
build: build-server build-client

# Build server
build-server:
	@echo "🔨 Building complete example server..."
	cd $(SERVER_PATH) && $(GOBUILD) -o $(SERVER_BINARY) .

# Build client
build-client:
	@echo "🔨 Building test client..."
	cd $(CLIENT_PATH) && $(GOBUILD) -o $(CLIENT_BINARY) .

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	cd $(SERVER_PATH) && $(GOCLEAN) && rm -f $(SERVER_BINARY)
	cd $(CLIENT_PATH) && $(GOCLEAN) && rm -f $(CLIENT_BINARY)

# Run the server
run-server:
	@echo "🚀 Starting complete example server..."
	cd $(SERVER_PATH) && $(GORUN) .

# Run the test client (server must be running)
run-client:
	@echo "🧪 Running comprehensive API tests..."
	@echo "⚠️  Make sure the server is running on localhost:8080"
	cd $(CLIENT_PATH) && $(GORUN) .

# Test API with curl (basic tests)
test-api:
	@echo "🧪 Running basic API tests with curl..."
	@echo "Testing health endpoint..."
	@curl -s http://localhost:8080/health | jq .
	@echo ""
	@echo "Testing API info endpoint..."
	@curl -s http://localhost:8080/ | jq .
	@echo ""
	@echo "Testing list users..."
	@curl -s "http://localhost:8080/api/v1/users?page=1&page_size=5" | jq .
	@echo ""
	@echo "Testing search users..."
	@curl -s -H "X-Client-ID: test-client" -H "X-API-Key: test-key" \
		"http://localhost:8080/api/v1/users/search?q=test&limit=5" | jq .

# Generate protobuf code
proto:
	@echo "📋 Generating protobuf code..."
	protoc --proto_path=./api \
	       --proto_path=../tag \
	       --proto_path=../third_party \
	       --experimental_allow_proto3_optional \
	       --go_out=paths=source_relative:./api \
	       --go-gin_out=paths=source_relative:./api \
	       complete_example.proto

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	$(GOGET) github.com/gin-gonic/gin
	$(GOGET) github.com/go-kenka/ginpb/binding
	$(GOGET) github.com/go-kenka/ginpb/middleware

# Run comprehensive tests
test-comprehensive: 
	@echo "🎯 Running comprehensive tests..."
	@echo "1️⃣  Starting server in background..."
	@cd $(SERVER_PATH) && $(GORUN) . &
	@echo "⏳ Waiting for server to start..."
	@sleep 3
	@echo "2️⃣  Running test client..."
	@cd $(CLIENT_PATH) && $(GORUN) . || true
	@echo "3️⃣  Stopping server..."
	@pkill -f "go run.*complete_server" || true
	@echo "✅ Comprehensive tests completed!"

# Quick demo
demo:
	@echo "🎬 Running quick demo..."
	@echo "Starting server and running basic tests..."
	@cd $(SERVER_PATH) && $(GORUN) . &
	@sleep 2
	@echo "Server started, testing endpoints..."
	@curl -s http://localhost:8080/health
	@echo ""
	@curl -s http://localhost:8080/
	@echo ""
	@pkill -f "go run.*complete_server" || true
	@echo "Demo completed!"

# Development mode - run server with auto-restart
dev:
	@echo "🔧 Development mode - server will auto-restart on changes"
	@echo "Install 'air' for hot reload: go install github.com/cosmtrek/air@latest"
	cd $(SERVER_PATH) && air -c ../.air.toml || $(GORUN) .

# Show help
help:
	@echo "📖 Complete Example Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build              - Build server and client binaries"
	@echo "  build-server       - Build only the server binary"
	@echo "  build-client       - Build only the client binary"
	@echo "  clean              - Clean build artifacts"
	@echo "  run-server         - Run the complete example server"
	@echo "  run-client         - Run the comprehensive test client"
	@echo "  test-api           - Test API with curl (requires jq)"
	@echo "  test-comprehensive - Run full automated test suite"
	@echo "  proto              - Regenerate protobuf code"
	@echo "  deps               - Install Go dependencies"
	@echo "  demo               - Run quick demo"
	@echo "  dev                - Development mode with hot reload"
	@echo "  help               - Show this help message"
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "  make run-server    # In one terminal"
	@echo "  make run-client    # In another terminal"
	@echo ""
	@echo "  Or run automated tests:"
	@echo "  make test-comprehensive"