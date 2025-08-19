# GinPB Middleware

GinPB Middleware 提供了类似 Kratos 的中间件功能，但遵循 Gin 的 HandlerFunc 风格。它支持灵活的中间件组合、条件应用和操作级别的中间件配置。

## 特性

- 🚀 **Gin 风格**: 完全兼容 gin.HandlerFunc，遵循 Gin 的中间件模式
- 🎯 **操作级别**: 支持基于 protobuf 操作名称的中间件应用
- 🔧 **条件应用**: 支持基于路径、方法等条件的中间件选择
- 📦 **中间件组**: 支持中间件分组和链式组合
- 🛡️ **内置中间件**: 提供日志、认证、恢复、CORS 等常用中间件

## 快速开始

### 1. 基础使用

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/go-kenka/ginpb/middleware"
)

r := gin.Default()
service := &YourService{}

// 基础注册（无中间件）
api.RegisterYourServiceHTTPServer(r, service)

// 带全局中间件注册
api.RegisterYourServiceHTTPServer(r, service,
    api.WithYourServiceGlobalMiddleware(
        middleware.Logging(),
        middleware.Recovery(),
        middleware.CORS(),
    ),
)
```

### 2. 操作特定中间件

```go
// 为单个操作配置中间件
api.RegisterYourServiceHTTPServer(r, service,
    api.WithYourServiceOperationMiddleware(
        api.OperationYourServiceMethodName,
        middleware.BearerAuth(),
        middleware.LoggingWithConfig(middleware.LoggingConfig{
            LogRequest:  true,
            LogResponse: true,
        }),
    ),
)

// 批量设置操作中间件
operationMiddlewares := map[string][]gin.HandlerFunc{
    api.OperationYourServiceCreateUser: {
        middleware.BearerAuth(),
        middleware.Logging(),
    },
    api.OperationYourServiceDeleteUser: {
        middleware.BearerAuth(),
        middleware.AdminAuth(),
    },
}

api.RegisterYourServiceHTTPServer(r, service,
    api.WithYourServiceOperationMiddlewares(operationMiddlewares),
)
```

## 内置中间件

### 日志中间件

```go
// 默认日志中间件
middleware.Logging()

// 自定义配置的日志中间件
middleware.LoggingWithConfig(middleware.LoggingConfig{
    LogLatency:   true,
    LogMethod:    true,
    LogPath:      true,
    LogStatus:    true,
    LogRequest:   true,  // 记录请求内容
    LogResponse:  true,  // 记录响应内容
    LogOperation: true,  // 记录操作名称
})
```

### 认证中间件

```go
// Bearer Token 认证
middleware.BearerAuth()

// 自定义验证器的 Bearer Token 认证
middleware.BearerAuthWithConfig(middleware.AuthConfig{
    Validator: func(c *gin.Context, token string) bool {
        return validateJWTToken(token)
    },
})

// API Key 认证
middleware.APIKeyAuth(map[string]bool{
    "your-api-key": true,
})

// 基础认证
middleware.BasicAuth(gin.Accounts{
    "admin": "password",
})
```

### 恢复中间件

```go
// 默认恢复中间件
middleware.Recovery()

// 自定义配置的恢复中间件
middleware.RecoveryWithConfig(middleware.RecoveryConfig{
    EnableStackTrace:    true,
    EnableDetailedError: true,
    RecoveryHandler: func(c *gin.Context, err interface{}) {
        // 自定义错误处理
        c.JSON(500, gin.H{"error": "custom error response"})
    },
})
```

### CORS 中间件

```go
// 默认 CORS 中间件
middleware.CORS()

// 自定义 CORS 配置
middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins:     []string{"http://localhost:3000", "https://example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
})
```

## 高级功能

### 条件中间件

```go
// 基于路径的条件中间件
middleware.NewConditionalMiddleware(
    middleware.PathSelector{Path: "/api/v1/users"},
    authMiddleware(),
).Apply()

// 基于 HTTP 方法的条件中间件
middleware.NewConditionalMiddleware(
    middleware.MethodSelector{Method: "POST"},
    loggingMiddleware(),
).Apply()
```

### 操作中间件

```go
// 基于操作名称的中间件
middleware.NewOperationMiddleware(
    api.OperationYourServiceMethodName,
    authMiddleware(),
).Apply()
```

### 中间件组

```go
// 创建中间件组
group := middleware.NewMiddlewareGroup(
    middleware.Recovery(),
    middleware.Logging(),
    customMiddleware(),
)

// 应用中间件组
r.Use(group.Apply()...)

// 或者包装为单个中间件
r.Use(group.Wrap())
```

## 代码生成增强

GinPB 的代码生成器会自动为每个生成的服务创建统一的注册函数和选项函数：

### 统一注册函数

```go
// 统一的注册函数，支持函数选项模式
func RegisterYourServiceHTTPServer(
    r gin.IRouter, 
    srv YourServiceHTTPServer, 
    opts ...YourServiceRegisterOption,
)
```

### 中间件选项函数

```go
// 全局中间件选项
func WithYourServiceGlobalMiddleware(middlewares ...gin.HandlerFunc) YourServiceRegisterOption

// 单个操作中间件选项
func WithYourServiceOperationMiddleware(operation string, middlewares ...gin.HandlerFunc) YourServiceRegisterOption

// 批量操作中间件选项  
func WithYourServiceOperationMiddlewares(middlewares map[string][]gin.HandlerFunc) YourServiceRegisterOption
```

## 完整示例

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/go-kenka/ginpb/middleware"
)

func main() {
    r := gin.Default()
    service := &YourService{}
    
    // 方式1: 基础注册（无中间件）
    basicGroup := r.Group("/api/v1/basic")
    api.RegisterYourServiceHTTPServer(basicGroup, service)
    
    // 方式2: 全局中间件
    publicGroup := r.Group("/api/v1/public")
    api.RegisterYourServiceHTTPServer(publicGroup, service,
        api.WithYourServiceGlobalMiddleware(
            middleware.Recovery(),
            middleware.CORS(),
            middleware.Logging(),
        ),
    )
    
    // 方式3: 操作特定中间件
    authGroup := r.Group("/api/v1/auth")
    api.RegisterYourServiceHTTPServer(authGroup, service,
        api.WithYourServiceOperationMiddleware(
            api.OperationYourServiceCreateUser,
            middleware.BearerAuth(),
            middleware.Logging(),
        ),
        api.WithYourServiceOperationMiddleware(
            api.OperationYourServiceDeleteUser,
            middleware.BearerAuth(),
            middleware.AdminAuth(),
        ),
    )
    
    // 方式4: 混合使用
    adminGroup := r.Group("/api/v1/admin")
    operationMiddlewares := map[string][]gin.HandlerFunc{
        api.OperationYourServiceCreateUser: {
            middleware.AdminAuth(),
            middleware.AuditLog(),
        },
        api.OperationYourServiceDeleteUser: {
            middleware.AdminAuth(),
            middleware.ConfirmationRequired(),
            middleware.AuditLog(),
        },
    }
    
    api.RegisterYourServiceHTTPServer(adminGroup, service,
        // 全局中间件（所有操作都会应用）
        api.WithYourServiceGlobalMiddleware(
            middleware.Recovery(),
            middleware.RateLimiting(),
        ),
        // 批量操作中间件
        api.WithYourServiceOperationMiddlewares(operationMiddlewares),
    )
    
    r.Run(":8080")
}
```

## 自定义中间件

创建自定义中间件非常简单，只需要返回 `gin.HandlerFunc`：

```go
func customMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        // 预处理
        start := time.Now()
        
        // 处理请求
        c.Next()
        
        // 后处理
        latency := time.Since(start)
        log.Printf("Request processed in %v", latency)
    })
}
```

## 操作名称

GinPB 会为每个 protobuf 方法生成操作常量：

```go
const (
    OperationYourServiceMethodName = "/package.YourService/MethodName"
)
```

这些常量可以用于操作特定的中间件配置。

## 与 Kratos 的对比

| 特性 | Kratos | GinPB Middleware |
|------|--------|------------------|
| 中间件类型 | `func(Handler) Handler` | `gin.HandlerFunc` |
| 操作匹配 | 内置支持 | 通过 operation 常量支持 |
| 条件应用 | Selector 机制 | Selector + Conditional 中间件 |
| 中间件链 | `middleware.Chain()` | `middleware.NewMiddlewareGroup()` |
| HTTP 集成 | 抽象传输层 | 原生 Gin 集成 |

## 最佳实践

1. **分层使用中间件**: 全局中间件 → 操作特定中间件，优先级从低到高
2. **函数选项组合**: 灵活组合多种中间件选项，满足不同场景需求  
3. **合理使用日志**: 在开发环境启用详细日志，生产环境关闭请求/响应日志
4. **认证中间件**: 将认证中间件应用在需要的操作上，避免全局应用
5. **错误处理**: 使用 Recovery 中间件防止 panic 导致服务崩溃
6. **性能监控**: 使用自定义中间件进行性能监控和指标收集

