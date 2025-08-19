# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GinPB is a protobuf code generation tool that generates Gin HTTP framework bindings from Protocol Buffer definitions. It provides:

- **protoc-gen-gin**: A protoc plugin that generates Gin server handlers and HTTP client implementations
- **Binding utilities**: Auto-detection of content types for request binding (JSON, XML, YAML, etc.)
- **HTTP client**: Resty-based client implementation for generated services
- **Middleware support**: Authentication, CORS, logging, recovery, and custom middleware
- **Custom field tags**: Support for gin binding tags in protobuf field definitions

## Key Architecture

### Core Components

- **cmd/protoc-gen-gin/**: Main protoc plugin entry point
- **internal/gen/gin.go**: Core code generation logic with templates for servers/clients
- **binding/**: Content-type aware request binding utilities  
- **client/**: HTTP client implementation using go-resty library
- **middleware/**: Common middleware implementations (auth, CORS, logging, recovery)
- **metadata/**: Gin context metadata handling
- **tag/**: Custom protobuf field tag definitions for gin binding

### Generated Code Structure

The plugin generates `.pb.gin.go` files containing:
1. HTTP server interfaces and registration functions
2. HTTP client interfaces and implementations  
3. Internal structs with gin binding tags for request handling
4. Operation constants for middleware identification

## Development Commands

### Initial Setup
```bash
# Install required protoc plugins
make init
```

### Code Generation
```bash
# Generate protobuf code for examples
make api

# Generate example-specific protobuf code
cd example && make proto
```

### Building and Testing
```bash
# Build the protoc-gen-gin plugin
go build -o bin/protoc-gen-gin ./cmd/protoc-gen-gin

# Run example server
cd example && make run-server

# Run comprehensive client tests  
cd example && make run-client

# Run automated test suite
cd example && make test-comprehensive

# Quick API demo
cd example && make demo
```

### Development Workflow
```bash
# Development mode with hot reload (requires air)
cd example && make dev

# Test individual endpoints
cd example && make test-api
```

## Testing

- **Unit tests**: Use `go test ./...` for package-level testing
- **Integration tests**: Located in example/ directory with comprehensive client tests
- **API testing**: Manual curl-based testing via `make test-api` in example/
- **Test framework**: Uses testify for assertions (github.com/stretchr/testify)

## Important Dependencies

- **Gin**: HTTP web framework (github.com/gin-gonic/gin)
- **Resty**: HTTP client library (github.com/go-resty/resty/v2)  
- **Protobuf**: Protocol buffer runtime (google.golang.org/protobuf)
- **Google APIs**: HTTP annotations (google.golang.org/genproto/googleapis/api)
- **Testify**: Testing framework (github.com/stretchr/testify)

## File Patterns

- **Proto files**: `*.proto` - Protocol buffer definitions
- **Generated files**: `*.pb.gin.go` - Generated Gin bindings (do not edit manually)
- **Standard proto**: `*.pb.go` - Standard protobuf generated code
- **Validation**: `*.pb.validate.go` - Generated validation code

## Code Generation Workflow

1. Write `.proto` files with HTTP annotations
2. Run `make api` to generate `.pb.gin.go` files
3. Implement the generated interfaces in your server code
4. Use generated client interfaces for calling services
5. Apply middleware as needed using the provided utilities

## 规则

- 使用中文回复
- 注释使用英文
- 代码风格遵循 Go 语言规范
- 避免重复代码,超过10行的重复代码抽象公共方法
- 变量名尽量简短,见名思意
- 接口名尽量简短,见名思意
- 错误信息尽量清晰,包括错误原因,解决方案,以及可能的影响