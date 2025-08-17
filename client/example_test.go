package client_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-kenka/ginpb/client"
	"github.com/go-kenka/ginpb/example/api"
)

// 示例：基本用法
func ExampleClient_basic() {
	// 创建客户端
	c := client.NewClient(
		client.WithEndpoint("http://localhost:8080"),
		client.WithTimeout(10*time.Second),
		client.WithUserAgent("my-app/1.0"),
	)

	ctx := context.Background()

	// 发送GET请求
	var resp api.GetArticlesResp
	err := c.Invoke(ctx, http.MethodGet, "/v1/articles", nil, &resp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d articles\n", resp.Total)
}

// 示例：使用CallOption
func ExampleClient_withCallOptions() {
	c := client.NewClient(
		client.WithEndpoint("http://localhost:8080"),
	)

	ctx := context.Background()
	req := &api.GetArticlesReq{
		Title:    "golang",
		Page:     1,
		PageSize: 10,
	}

	var resp api.GetArticlesResp
	err := c.Invoke(ctx, http.MethodGet, "/v1/articles", req, &resp,
		client.Operation("/example.BlogService/GetArticles"),
		client.PathTemplate("/v1/articles"),
		client.ContentType("application/json"),
		client.Header("X-Request-ID", "12345"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Articles: %+v\n", resp)
}

// 示例：使用中间件
func ExampleClient_withMiddleware() {
	// 日志中间件
	logger := func(format string, args ...interface{}) {
		log.Printf("[HTTP] "+format, args...)
	}

	// 创建带中间件的客户端
	c := client.NewClient(
		client.WithEndpoint("http://localhost:8080"),
		client.WithRetryCount(3),
		client.WithRequestMiddleware(
			client.LoggingRequestMiddleware(logger),
			client.AuthRequestMiddleware("your-api-token"),
		),
		client.WithResponseMiddleware(
			client.LoggingResponseMiddleware(logger),
		),
		client.WithErrorMiddleware(
			client.LoggingErrorMiddleware(logger),
		),
	)

	ctx := context.Background()

	// 发送请求
	var resp api.GetArticlesResp
	err := c.Invoke(ctx, http.MethodGet, "/v1/articles", nil, &resp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response: %+v\n", resp)
}

// 示例：POST请求
func ExampleClient_post() {
	c := client.NewClient(
		client.WithEndpoint("http://localhost:8080"),
	)

	ctx := context.Background()
	req := &api.Article{
		Title:    "My New Article",
		Content:  "This is the content",
		AuthorId: 123,
	}

	var resp api.Article
	err := c.Invoke(ctx, http.MethodPost, "/v1/author/123/articles", req, &resp,
		client.ContentType("application/json"),
		client.BearerToken("your-jwt-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created article: %s\n", resp.Title)
}

// 示例：错误处理
func ExampleClient_errorHandling() {
	c := client.NewClient(
		client.WithEndpoint("http://localhost:8080"),
		client.WithErrorDecoder(func(resp *http.Response) error {
			// 自定义错误解码
			if resp.StatusCode == 404 {
				return fmt.Errorf("resource not found")
			}
			return client.DefaultErrorDecoder(resp)
		}),
	)

	ctx := context.Background()

	var resp api.GetArticlesResp
	err := c.Invoke(ctx, http.MethodGet, "/v1/articles/999", nil, &resp)
	if err != nil {
		// 检查错误类型
		if client.IsHTTPError(err) {
			statusCode := client.GetHTTPStatusCode(err)
			fmt.Printf("HTTP Error %d: %v\n", statusCode, err)
		} else {
			fmt.Printf("Other error: %v\n", err)
		}
		return
	}
}

// 示例：自定义传输
func ExampleClient_customTransport() {
	// 自定义HTTP传输配置
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	c := client.NewClient(
		client.WithEndpoint("https://api.example.com"),
		client.WithTransport(transport),
		client.WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	var resp map[string]interface{}
	err := c.Invoke(ctx, http.MethodGet, "/api/status", nil, &resp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("API Status: %+v\n", resp)
}

// 博客服务客户端包装示例
type BlogServiceClient struct {
	client client.Client
}

func NewBlogServiceClient(endpoint string) *BlogServiceClient {
	c := client.NewClient(
		client.WithEndpoint(endpoint),
		client.WithTimeout(30*time.Second),
		client.WithUserAgent("blog-service-client/1.0"),
	)

	return &BlogServiceClient{client: c}
}

func (c *BlogServiceClient) GetArticles(ctx context.Context, req *api.GetArticlesReq) (*api.GetArticlesResp, error) {
	var resp api.GetArticlesResp

	// 使用EncodeURL处理路径和查询参数
	path := client.EncodeURL("/v1/articles", req, true)
	if req.AuthorId > 0 {
		path = client.EncodeURL("/v1/author/{author_id}/articles", req, true)
	}

	err := c.client.Invoke(ctx, http.MethodGet, path, nil, &resp,
		client.Operation("/example.BlogService/GetArticles"),
		client.PathTemplate(path),
	)

	return &resp, err
}

func (c *BlogServiceClient) CreateArticle(ctx context.Context, req *api.Article) (*api.Article, error) {
	var resp api.Article

	path := fmt.Sprintf("/v1/author/%d/articles", req.AuthorId)

	err := c.client.Invoke(ctx, http.MethodPost, path, req, &resp,
		client.Operation("/example.BlogService/CreateArticle"),
		client.PathTemplate("/v1/author/{author_id}/articles"),
		client.ContentType("application/json"),
	)

	return &resp, err
}

// 示例：使用服务客户端
func ExampleBlogServiceClient() {
	client := NewBlogServiceClient("http://localhost:8080")
	ctx := context.Background()

	// 获取文章
	articles, err := client.GetArticles(ctx, &api.GetArticlesReq{
		Title:    "golang",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d articles\n", articles.Total)

	// 创建文章
	newArticle, err := client.CreateArticle(ctx, &api.Article{
		Title:    "New Article",
		Content:  "Article content",
		AuthorId: 123,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created: %s\n", newArticle.Title)
}
