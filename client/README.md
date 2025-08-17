# HTTP Client 封装

基于 `github.com/carlmjohnson/requests` 库实现的类似 Kratos HTTP Client 的封装，提供了丰富的功能和灵活的配置选项。

## 特性

- 🚀 **简洁API**: 类似Kratos的`Invoke`方法设计
- ⚙️ **灵活配置**: 支持各种ClientOption和CallOption
- 🔄 **中间件支持**: 内置日志、重试、认证等中间件
- 🎯 **错误处理**: 完善的HTTP错误处理机制  
- 📝 **编解码器**: 可自定义请求/响应编解码器
- 🏷️ **结构体标签**: 支持uri和form标签自动编码URL
- ✅ **类型安全**: 完全的Go类型安全

## 安装

```bash
go get github.com/carlmjohnson/requests
```

## 基本用法

### 创建客户端

```go
package main

import (
    "context"
    "time"
    
    "github.com/go-kenka/kratos-gin/client"
)

func main() {
    // 基础客户端
    c := client.NewClient(
        client.WithEndpoint("http://localhost:8080"),
        client.WithTimeout(30*time.Second),
        client.WithUserAgent("my-app/1.0"),
    )

    // 使用中间件的客户端
    c := client.NewClient(
        client.WithEndpoint("http://localhost:8080"),
        client.WithMiddleware(
            client.LoggingMiddleware(log.Printf),
            client.RetryMiddleware(3),
            client.AuthMiddleware("your-token"),
        ),
    )
}
```

### 发送请求

```go
// GET请求
var resp GetArticlesResp
err := c.Invoke(ctx, "GET", "/v1/articles", nil, &resp)

// POST请求
req := &CreateArticleReq{Title: "New Article"}
var resp CreateArticleResp
err := c.Invoke(ctx, "POST", "/v1/articles", req, &resp,
    client.ContentType("application/json"),
    client.BearerToken("jwt-token"),
)
```

## 配置选项

### ClientOption (客户端级别配置)

| 选项 | 说明 | 示例 |
|------|------|------|
| `WithEndpoint` | 设置服务端点 | `WithEndpoint("http://api.example.com")` |
| `WithTimeout` | 设置请求超时 | `WithTimeout(30*time.Second)` |
| `WithUserAgent` | 设置User-Agent | `WithUserAgent("my-app/1.0")` |
| `WithMiddleware` | 添加中间件 | `WithMiddleware(LoggingMiddleware(...))` |
| `WithErrorDecoder` | 自定义错误解码器 | `WithErrorDecoder(customDecoder)` |
| `WithTransport` | 自定义HTTP传输 | `WithTransport(customTransport)` |
| `WithHeader` | 添加默认请求头 | `WithHeader("API-Key", "secret")` |

### CallOption (单次调用配置)

| 选项 | 说明 | 示例 |
|------|------|------|
| `Operation` | 设置操作名称 | `Operation("/api.Service/Method")` |
| `PathTemplate` | 设置路径模板 | `PathTemplate("/users/{id}")` |
| `Header` | 添加请求头 | `Header("Content-Type", "application/json")` |
| `ContentType` | 设置Content-Type | `ContentType("application/json")` |
| `BearerToken` | 设置Bearer Token | `BearerToken("jwt-token")` |
| `BasicAuth` | 设置基础认证 | `BasicAuth("user", "pass")` |

## 中间件

### 内置中间件

```go
// 日志中间件
client.LoggingMiddleware(log.Printf)

// 重试中间件 (最多重试3次)
client.RetryMiddleware(3)

// 认证中间件
client.AuthMiddleware("bearer-token")

// 超时中间件
client.TimeoutMiddleware(10*time.Second)

// 自定义头部中间件
client.HeaderMiddleware(map[string]string{
    "API-Version": "v1",
    "Accept": "application/json",
})
```

### 自定义中间件

```go
func MyMiddleware() client.Middleware {
    return func(next client.Handler) client.Handler {
        return func(ctx context.Context, req *http.Request) (*http.Response, error) {
            // 请求前处理
            start := time.Now()
            
            // 调用下一个处理器
            resp, err := next(ctx, req)
            
            // 请求后处理
            duration := time.Since(start)
            log.Printf("Request took %v", duration)
            
            return resp, err
        }
    }
}
```

## 错误处理

### HTTP错误

```go
err := c.Invoke(ctx, "GET", "/api/resource", nil, &response)
if err != nil {
    if client.IsHTTPError(err) {
        statusCode := client.GetHTTPStatusCode(err)
        fmt.Printf("HTTP %d: %v", statusCode, err)
        
        // 检查错误类型
        if client.IsClientError(err) {
            // 4xx错误
        }
        if client.IsServerError(err) {
            // 5xx错误
        }
        if client.IsRetryableError(err) {
            // 可重试错误
        }
    }
}
```

### 自定义错误解码器

```go
customErrorDecoder := func(resp *http.Response) error {
    // 解析自定义错误格式
    var apiErr struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
    }
    
    json.NewDecoder(resp.Body).Decode(&apiErr)
    
    return &client.HTTPError{
        Code:    resp.StatusCode,
        Message: apiErr.Message,
    }
}

c := client.NewClient(
    client.WithEndpoint("http://api.example.com"),
    client.WithErrorDecoder(customErrorDecoder),
)
```

## URL编码

支持结构体标签自动编码URL路径参数和查询参数：

```go
type GetUserReq struct {
    UserID   int32  `uri:"user_id"`      // 路径参数
    Page     int32  `form:"page"`        // 查询参数  
    PageSize int32  `form:"page_size"`   // 查询参数
    Keyword  string `form:"keyword"`     // 查询参数
}

req := &GetUserReq{
    UserID:   123,
    Page:     1, 
    PageSize: 10,
    Keyword:  "golang",
}

// 自动编码为: /users/123?page=1&page_size=10&keyword=golang
path := client.EncodeURL("/users/{user_id}", req, true)
```

## 服务客户端封装

推荐为每个服务创建专门的客户端封装：

```go
type BlogServiceClient struct {
    client client.Client
}

func NewBlogServiceClient(endpoint string) *BlogServiceClient {
    c := client.NewClient(
        client.WithEndpoint(endpoint),
        client.WithTimeout(30*time.Second),
        client.WithMiddleware(
            client.LoggingMiddleware(log.Printf),
            client.RetryMiddleware(3),
        ),
    )
    
    return &BlogServiceClient{client: c}
}

func (c *BlogServiceClient) GetArticles(ctx context.Context, req *GetArticlesReq) (*GetArticlesResp, error) {
    var resp GetArticlesResp
    
    path := client.EncodeURL("/v1/articles", req, true)
    if req.AuthorId > 0 {
        path = client.EncodeURL("/v1/author/{author_id}/articles", req, true)
    }
    
    err := c.client.Invoke(ctx, "GET", path, nil, &resp,
        client.Operation("/example.BlogService/GetArticles"),
    )
    
    return &resp, err
}
```

## 完整示例

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/go-kenka/kratos-gin/client"
)

func main() {
    // 创建客户端
    c := client.NewClient(
        client.WithEndpoint("http://localhost:8080"),
        client.WithTimeout(10*time.Second),
        client.WithMiddleware(
            client.LoggingMiddleware(log.Printf),
            client.RetryMiddleware(2),
        ),
    )

    ctx := context.Background()

    // GET请求
    var articles GetArticlesResp
    err := c.Invoke(ctx, "GET", "/v1/articles", nil, &articles,
        client.Header("Accept", "application/json"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d articles", articles.Total)

    // POST请求
    newArticle := &Article{
        Title:   "My Article",
        Content: "Article content",
    }
    
    var created Article
    err = c.Invoke(ctx, "POST", "/v1/articles", newArticle, &created,
        client.ContentType("application/json"),
        client.BearerToken("your-jwt-token"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Created article: %s", created.Title)
}
```

## 与Kratos HTTP Client的对比

| 特性 | Kratos HTTP Client | 本封装 |
|------|-------------------|--------|
| 依赖 | 完整Kratos框架 | 仅requests库 |
| API设计 | `Invoke(ctx, method, path, args, reply, ...opts)` | 相同 |
| 中间件 | 支持 | 支持 |
| 错误处理 | Kratos错误系统 | HTTP错误 + 自定义 |
| 配置方式 | 功能选项 | 功能选项 |
| URL编码 | `binding.EncodeURL` | `client.EncodeURL` |
| 大小 | 较大 | 轻量级 |

这个封装在保持Kratos HTTP Client API兼容性的同时，提供了更轻量级的实现方案。