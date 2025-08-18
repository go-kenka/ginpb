# Complete Example - Comprehensive HTTP API Demo

This example demonstrates all features of the ginpb framework including:

- ✅ All HTTP methods (GET, POST, PUT, PATCH, DELETE)
- ✅ All parameter binding types (Query, Path, Header, JSON, Form, Multipart)
- ✅ Validation rules and custom tags
- ✅ Middleware integration (Auth, Logging, Recovery, CORS)
- ✅ Complex scenarios (nested paths, multiple bindings)
- ✅ Error handling and response formatting

## 📁 Project Structure

```
example/
├── api/                          # Protocol Buffer definitions
│   ├── complete_example.proto   # Comprehensive API definition
│   ├── complete_example.pb.go   # Generated Go structs
│   └── complete_example_gin.pb.go # Generated Gin handlers
├── server/                       # Server implementation
│   └── complete_server.go       # Complete service implementation
├── client/                       # Test client
│   └── test_client.go           # Comprehensive API test client
├── Makefile                     # Build and test automation
└── README.md                    # This file
```

## 🚀 Quick Start

### Option 1: Automated Testing
```bash
# Run comprehensive automated tests
make test-comprehensive
```

### Option 2: Manual Testing
```bash
# Terminal 1: Start the server
make run-server

# Terminal 2: Run the test client
make run-client
```

### Option 3: Individual Commands
```bash
# Start server
cd server && go run .

# Run tests
cd client && go run .
```

## 🌐 API Endpoints

### User Management
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/users` | List users with pagination and filtering | No |
| GET | `/api/v1/users/{user_id}` | Get specific user details | No |
| GET | `/api/v1/users/search` | Search users with advanced filters | No |
| POST | `/api/v1/users` | Create new user | Yes (Bearer) |
| PUT | `/api/v1/users/{user_id}` | Update user completely | Yes (Bearer) |
| PATCH | `/api/v1/users/{user_id}` | Partial user update | Yes (Bearer) |
| DELETE | `/api/v1/users/{user_id}` | Delete user | Yes (Admin Bearer) |
| DELETE | `/api/v1/users` | Batch delete users | Yes (Admin Bearer) |

### User Registration
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/users/register` | User registration with form data | No |

### Posts Management
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/users/{user_id}/posts` | Create user post | No |
| GET | `/api/v1/users/{user_id}/posts/{post_id}/comments` | Get post comments | No |

### Profiles
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/users/{user_id}/profile` | Get user profile | No |
| PUT | `/api/v1/users/{user_id}/profile` | Update user profile | No |
| GET | `/api/v1/profiles/{user_id}/data` | Alternative profile route | No |

### Utility
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/` | API documentation |

## 🔗 Parameter Binding Examples

### Query Parameters (Form Binding)
```protobuf
// Proto definition
int32 page = 1 [(tag.tags) = { form: "page", binding: "min=1" }];
string sort_by = 2 [(tag.tags) = { form: "sort_by", binding: "oneof=id name email" }];

// HTTP request
GET /api/v1/users?page=1&sort_by=name
```

### Path Parameters (URI Binding)
```protobuf
// Proto definition
string user_id = 1 [(tag.tags) = { uri: "user_id", binding: "required,uuid" }];

// HTTP request
GET /api/v1/users/550e8400-e29b-41d4-a716-446655440000
```

### Header Parameters
```protobuf
// Proto definition
string client_id = 1 [(tag.tags) = { header: "X-Client-ID", binding: "required" }];
string api_key = 2 [(tag.tags) = { header: "X-API-Key", binding: "required,min=32" }];

// HTTP request
GET /api/v1/users/search
X-Client-ID: demo-client-001
X-API-Key: demo-api-key-12345
```

### JSON Body Binding
```protobuf
// Proto definition
string username = 1 [(tag.tags) = { json: "username", binding: "required,min=3,max=50,alphanum" }];
string email = 2 [(tag.tags) = { json: "email", binding: "required,email" }];

// HTTP request
POST /api/v1/users
Content-Type: application/json
{
  "username": "testuser123",
  "email": "test@example.com"
}
```

### Form Data Binding
```protobuf
// Proto definition
string username = 1 [(tag.tags) = { form: "username", binding: "required,min=3,max=30,alphanum" }];
string password = 2 [(tag.tags) = { form: "password", binding: "required,min=8" }];

// HTTP request
POST /api/v1/users/register
Content-Type: application/x-www-form-urlencoded
username=formuser123&password=password123
```

### Multipart Form Binding
```protobuf
// Proto definition
string filename = 1 [(tag.tags) = { multipart: "filename", binding: "required" }];
string content_type = 2 [(tag.tags) = { multipart: "content_type", binding: "required,oneof=image/jpeg image/png" }];

// HTTP request
POST /api/v1/users/user-123/avatar
Content-Type: multipart/form-data
(multipart form data with file upload)
```

## 🔒 Authentication Examples

### Bearer Token Authentication
```bash
# Create user (requires authentication)
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer demo-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"username": "newuser", "email": "new@example.com", "password": "password123"}'
```

### Admin Operations
```bash
# Delete user (requires admin token)
curl -X DELETE http://localhost:8080/api/v1/users/user-123 \
  -H "Authorization: Bearer admin-secret-key" \
  -H "X-Confirm-Delete: DELETE"
```

## 🧪 Test Scenarios

The test client demonstrates:

1. **GET Requests**: Query parameters, path parameters, headers
2. **POST Requests**: JSON body, form data, multipart uploads
3. **PUT Requests**: Complete resource updates, partial body updates
4. **PATCH Requests**: Partial updates, nested object patches
5. **DELETE Requests**: Single deletion, batch operations
6. **Complex Scenarios**: Nested paths, multiple route bindings
7. **Content Types**: JSON, XML, YAML support (simulated)
8. **Authentication**: Bearer tokens, admin operations
9. **Validation**: All validation rules and error responses

## 🛡️ Middleware Features

### Global Middleware
- **Logging**: JSON structured logging with request/response capture
- **Recovery**: Panic recovery with stack traces
- **CORS**: Cross-origin resource sharing with configurable policies

### Operation-Specific Middleware
- **Authentication**: Bearer token validation for protected endpoints
- **Authorization**: Admin-level operations with enhanced security

### Middleware Configuration Example
```go
// Global middleware
r.Use(middleware.LoggingWithConfig(middleware.LoggingConfig{
    Format: middleware.LogFormatJSON,
}))
r.Use(middleware.Recovery())
r.Use(middleware.CORS())

// Operation-specific middleware
operationMiddleware := map[string][]gin.HandlerFunc{
    "CreateUser": {middleware.BearerAuth("demo-secret-key")},
    "DeleteUser": {middleware.BearerAuth("admin-secret-key")},
}
```

## 📊 Validation Rules Examples

### Basic Validation
```protobuf
string username = 1 [(tag.tags) = { json: "username", binding: "required,min=3,max=50,alphanum" }];
string email = 2 [(tag.tags) = { json: "email", binding: "required,email" }];
int32 age = 3 [(tag.tags) = { json: "age", binding: "min=13,max=120" }];
```

### Advanced Validation
```protobuf
string gender = 1 [(tag.tags) = { json: "gender", binding: "oneof=male female other prefer_not_to_say" }];
string phone = 2 [(tag.tags) = { json: "phone", binding: "len=11,numeric" }];
string birth_date = 3 [(tag.tags) = { form: "birth_date", binding: "required,datetime=2006-01-02" }];
```

### Array and Object Validation
```protobuf
repeated string hobbies = 1 [(tag.tags) = { json: "hobbies", binding: "min=1,max=10" }];
Address address = 2 [(tag.tags) = { json: "address", binding: "required" }];
bool agree_terms = 3 [(tag.tags) = { json: "agree_terms", binding: "required,eq=true" }];
```

### Custom Validation Tags
```protobuf
string referral_code = 1 [(tag.tags) = { json: "referral_code", custom: "referral_format" }];
repeated string tags = 2 [(tag.tags) = { json: "tags", custom: "max_length:20" }];
map<string, string> metadata = 3 [(tag.tags) = { json: "metadata", custom: "max_keys:10" }];
```

## 🐛 Error Handling

The server demonstrates comprehensive error handling:

- **Validation Errors**: Detailed field-level validation messages
- **Authentication Errors**: Token validation and authorization failures
- **Business Logic Errors**: Custom application errors with proper HTTP status codes
- **System Errors**: Panic recovery with structured error responses

## 🔧 Development Commands

```bash
# Build everything
make build

# Clean build artifacts
make clean

# Generate protobuf code
make proto

# Install dependencies
make deps

# Run quick demo
make demo

# Development mode with hot reload
make dev

# Show all available commands
make help
```

## 📈 Performance Features

- **Connection Pooling**: Efficient HTTP client with connection reuse
- **JSON Streaming**: Large response handling with streaming
- **Middleware Optimization**: Conditional middleware execution
- **Memory Management**: Efficient request/response processing

## 🔍 Debugging

### Enable Debug Logging
```bash
# Set Gin to debug mode
export GIN_MODE=debug

# Run server with verbose logging
make run-server
```

### API Testing with curl
```bash
# Test basic endpoints
make test-api

# Manual curl examples
curl -v http://localhost:8080/health
curl -v "http://localhost:8080/api/v1/users?page=1&page_size=5"
```

## 📚 Learning Resources

This example demonstrates:

1. **Protocol Buffers**: Advanced proto3 features and annotations
2. **Gin Framework**: All binding types, middleware, and routing
3. **HTTP API Design**: RESTful patterns and best practices
4. **Go Development**: Clean code architecture and testing
5. **API Documentation**: Self-documenting APIs with structured responses

## 🤝 Contributing

To extend this example:

1. Add new endpoints to `complete_example.proto`
2. Regenerate code with `make proto`
3. Implement service methods in `complete_server.go`
4. Add test cases in `test_client.go`
5. Update documentation

## 📄 License

This example is part of the ginpb framework and follows the same license terms.