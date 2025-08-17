package main

import (
	"context"
	"log"
	"time"

	"github.com/go-kenka/ginpb/client"
	"github.com/go-kenka/ginpb/example/api"
)

func main() {
	// 创建使用resty-based client的服务客户端
	blogClient := api.NewBlogServiceHTTPClient(
		client.WithEndpoint("http://localhost:8080"),
		client.WithTimeout(30*time.Second),
		client.WithRetryCount(3),
		client.WithRequestMiddleware(
			client.LoggingRequestMiddleware(func(format string, args ...interface{}) {
				log.Printf("[REQUEST] "+format, args...)
			}),
			client.AuthRequestMiddleware("your-api-token"),
		),
		client.WithResponseMiddleware(
			client.LoggingResponseMiddleware(func(format string, args ...interface{}) {
				log.Printf("[RESPONSE] "+format, args...)
			}),
		),
	)

	ctx := context.Background()

	// 获取文章列表
	log.Println("=== 获取文章列表 ===")
	articlesResp, err := blogClient.GetArticles(ctx, &api.GetArticlesReq{
		Title:    "golang",
		PageSize: 10,
		AuthorId: 1,
	})
	if err != nil {
		log.Printf("获取文章失败: %v", err)
	} else {
		log.Printf("获取到 %d 篇文章", len(articlesResp.Articles))
	}

	// 创建文章
	log.Println("\n=== 创建文章 ===")
	newArticle, err := blogClient.CreateArticle(ctx, &api.Article{
		Title:    "使用resty-based client的新文章",
		Content:  "这是一篇使用新的resty-based client生成的代码创建的文章",
		AuthorId: 1,
	},
		// 可以为单个请求添加额外的选项
		client.Header("X-Custom-Header", "custom-value"),
		client.BearerToken("request-specific-token"),
	)
	if err != nil {
		log.Printf("创建文章失败: %v", err)
	} else {
		log.Printf("创建文章成功: %s", newArticle.Title)
	}

	log.Println("\n=== Demo 完成 ===")
}
