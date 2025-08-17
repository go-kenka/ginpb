# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**kratos-gin** is a Protocol Buffer compiler plugin that generates Gin HTTP handlers from protobuf service definitions. It extends the Kratos v2 framework by providing native Gin integration instead of the default HTTP handlers.

### Technology Stack

- **Language**: Go 1.19+
- **Protocol Buffers**: Google Protocol Buffers for API definition
- **Framework**: Kratos v2 microservice framework 
- **HTTP Framework**: Gin HTTP web framework
- **Code Generation**: protoc plugin for Go code generation

## Core Architecture

### Code Generation Flow

1. **Proto Definition**: Services and messages defined in `.proto` files using `google.api.http` annotations
2. **Code Generation**: `protoc-gen-go-gin` plugin processes proto files and generates Gin handlers
3. **Generated Output**: Creates `*_gin.pb.go` files with Gin-compatible HTTP handlers
4. **Integration**: Generated handlers integrate with Gin routers and Kratos services

### Key Components

- **protoc-gen-go-gin**: The main code generator plugin
- **gincontext**: Context wrapper for Gin-Kratos integration
- **HTTP Annotations**: Uses `google.api.http` annotations for REST API mapping
- **Validation**: Integrates with protoc-gen-validate for request validation

## Development Commands

### Prerequisites

Install required tools and dependencies:

```bash
# Install protoc compiler and plugins
make init

# Or manually install dependencies:
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
go install github.com/envoyproxy/protoc-gen-validate@latest
go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
go install github.com/go-kenka/kratos-gin/cmd/protoc-gen-go-gin@latest
```

### Code Generation

```bash
# Generate API code from proto files
make api

# Generate struct tags for validation/binding
make inject-tag
```

### Manual protoc Command

```bash
# Generate Go code, Gin handlers, OpenAPI spec, and validation
protoc --proto_path=./example/api \
       --proto_path=./third_party \
       --go_out=paths=source_relative:./example/api \
       --go-gin_out=paths=source_relative:./example/api \
       --openapi_out=paths=source_relative:. \
       --validate_out=paths=source_relative,lang=go:./example/api \
       --go-errors_out=paths=source_relative:./example/api \
       example/api/example.proto
```

### Build and Test

```bash
# Build the plugin
go build -o bin/protoc-gen-go-gin cmd/protoc-gen-go-gin/main.go

# Run tests
go test ./...

# Install plugin locally
go install ./cmd/protoc-gen-go-gin
```

## Proto File Structure

### Service Definition

```protobuf
service BlogService {
  rpc GetArticles(GetArticlesReq) returns (GetArticlesResp) {
    option (google.api.http) = {
      get: "/v1/articles"
      additional_bindings {
        get: "/v1/author/{author_id}/articles"
      }
    };
  }
  
  rpc CreateArticle(Article) returns (Article) {
    option (google.api.http) = {
      post: "/v1/author/{author_id}/articles"
      body: "*"
    };
  }
}
```

### Message with Binding Tags

```protobuf
message GetArticlesReq {
  // @gotags: form:"title"
  string title = 1;
  
  // @gotags: form:"page_size" binding:"required"
  int32 page_size = 3;
  
  // @gotags: form:"author_id" uri:"author_id"
  int32 author_id = 4;
}
```

## Generated Code Structure

### HTTP Handler Generation

The plugin generates:
- **Service Interface**: `BlogServiceHTTPServer` interface
- **Router Registration**: `RegisterBlogServiceHTTPServer` function  
- **Handler Functions**: Individual HTTP handlers for each RPC method
- **Request Binding**: Automatic binding of query params, URI params, and request body
- **Error Handling**: Integration with Gin's error handling

### Generated Handler Example

```go
func RegisterBlogServiceHTTPServer(r gin.IRouter, srv BlogServiceHTTPServer) {
    r.GET("/v1/articles", _BlogService_GetArticles1_HTTP_Handler(srv))
    r.POST("/v1/author/:author_id/articles", _BlogService_CreateArticle0_HTTP_Handler(srv))
}
```

## HTTP Method Mapping

The plugin supports all standard HTTP methods:
- `GET` - Read operations
- `POST` - Create operations  
- `PUT` - Update operations (full replacement)
- `PATCH` - Partial updates
- `DELETE` - Delete operations
- `Custom` - Custom HTTP methods

## Path Parameter Handling

### Proto Route Definition
```protobuf
option (google.api.http) = {
  get: "/v1/author/{author_id}/articles/{article_id}"
};
```

### Generated Gin Route
```go
r.GET("/v1/author/:author_id/articles/:article_id", handler)
```

## Request Binding Features

### Automatic Binding
- **Query Parameters**: Bound to struct fields with `form` tags
- **URI Parameters**: Bound to struct fields with `uri` tags  
- **Request Body**: JSON body binding for POST/PUT/PATCH requests
- **Validation**: Integrated with struct validation tags

### @gotags Integration
Use `@gotags` comments in proto files to add Go struct tags:
```protobuf
// @gotags: form:"name" binding:"required" validate:"min=1,max=100"
string name = 1;
```

## Development Workflow

### 1. Define Proto Services
Create `.proto` files with service definitions and HTTP annotations

### 2. Generate Code
Run `make api` to generate Go code and Gin handlers

### 3. Implement Service
Implement the generated service interface:
```go
type blogService struct{}

func (s *blogService) GetArticles(ctx context.Context, req *api.GetArticlesReq) (*api.GetArticlesResp, error) {
    // Implementation
    return &api.GetArticlesResp{}, nil
}
```

### 4. Register Routes
Register the generated handlers with Gin router:
```go
r := gin.Default()
api.RegisterBlogServiceHTTPServer(r, &blogService{})
```

## Plugin Configuration

### Command Line Options
- `--omitempty`: Skip generation if no HTTP rules are defined (default: true)
- `--version`: Display plugin version

### Supported Features
- **Proto3 Optional**: Full support for proto3 optional fields
- **Additional Bindings**: Multiple HTTP bindings per RPC method
- **Custom HTTP Methods**: Support for custom HTTP verbs
- **Path Parameters**: URL parameter extraction and validation
- **Request Body Mapping**: Flexible body field mapping

## Error Handling

The generated code integrates with Gin's error handling:
- Parameter binding errors are automatically handled
- Service errors are propagated through Gin's error system
- Compatible with Kratos error handling patterns

## Integration with Kratos

### Service Registration
```go
// Register with Kratos HTTP server
httpSrv := http.NewServer(http.Address(":8000"))
api.RegisterBlogServiceHTTPServer(httpSrv.Router, service)
```

### Middleware Support
Generated handlers work with standard Gin middleware:
```go
r.Use(gin.Logger())
r.Use(gin.Recovery())
api.RegisterBlogServiceHTTPServer(r, service)
```

## Directory Structure

```
kratos-gin/
├── cmd/protoc-gen-go-gin/    # Plugin main entry point
│   ├── main.go               # CLI and plugin setup
│   ├── gin.go                # Core code generation logic  
│   ├── template.go           # Go code templates
│   └── version.go            # Version information
├── gincontext/               # Gin-Kratos context bridge
├── example/                  # Usage examples
│   └── api/                  # Example proto and generated code
├── third_party/              # Third-party proto definitions
└── Makefile                  # Build automation
```

## Best Practices

1. **Use HTTP annotations**: Always define `google.api.http` options for REST APIs
2. **Validate inputs**: Use `@gotags` for request validation
3. **Handle errors properly**: Implement proper error responses in service methods
4. **Use additional_bindings**: Provide multiple endpoints for the same operation when needed
5. **Follow REST conventions**: Use appropriate HTTP methods for operations
6. **Test generated code**: Verify generated handlers work correctly with your services

## Troubleshooting

### Common Issues
- **Missing HTTP annotations**: Ensure all RPC methods have `google.api.http` options
- **Path parameter mismatches**: Verify proto field names match URL parameters
- **Binding errors**: Check `@gotags` syntax for form/uri binding
- **Import issues**: Ensure all required proto dependencies are in `third_party/`

### Debug Commands
```bash
# Check plugin version
protoc-gen-go-gin --version

# Verify proto syntax
protoc --proto_path=. --go_out=. example.proto

# Test generated code compilation
go build ./example/api
```