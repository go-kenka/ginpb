package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client 是基于resty库的HTTP客户端接口
type Client interface {
	Invoke(ctx context.Context, method, path string, args interface{}, reply interface{}, opts ...CallOption) error
}

// client 是Client接口的实现
type client struct {
	resty *resty.Client
	opts  clientOptions
}

// clientOptions 客户端配置选项
type clientOptions struct {
	endpoint            string
	timeout             time.Duration
	userAgent           string
	errorDecoder        ErrorDecoder
	encoder             RequestEncoder
	decoder             ResponseDecoder
	transport           http.RoundTripper
	headers             map[string]string
	requestMiddlewares  []RestyRequestMiddleware
	responseMiddlewares []RestyResponseMiddleware
	errorMiddlewares    []RestyErrorMiddleware
	retryCount          int
	retryWaitTime       time.Duration
	retryMaxWaitTime    time.Duration
}

// NewClient 创建新的HTTP客户端
func NewClient(opts ...ClientOption) Client {
	o := clientOptions{
		timeout:      30 * time.Second,
		userAgent:    "kratos-gin/1.0",
		errorDecoder: DefaultErrorDecoder,
		encoder:      DefaultRequestEncoder,
		decoder:      DefaultResponseDecoder,
		headers:      make(map[string]string),
	}

	for _, opt := range opts {
		opt(&o)
	}

	// 创建resty客户端
	restyClient := resty.New()

	// 配置基础选项
	if o.endpoint != "" {
		restyClient.SetBaseURL(o.endpoint)
	}
	if o.timeout > 0 {
		restyClient.SetTimeout(o.timeout)
	}
	if o.userAgent != "" {
		restyClient.SetHeader("User-Agent", o.userAgent)
	}
	if o.transport != nil {
		restyClient.SetTransport(o.transport)
	}

	// 设置默认headers
	if len(o.headers) > 0 {
		restyClient.SetHeaders(o.headers)
	}

	// 配置重试
	if o.retryCount > 0 {
		restyClient.SetRetryCount(o.retryCount)
		if o.retryWaitTime > 0 {
			restyClient.SetRetryWaitTime(o.retryWaitTime)
		}
		if o.retryMaxWaitTime > 0 {
			restyClient.SetRetryMaxWaitTime(o.retryMaxWaitTime)
		}
	}

	// 创建客户端实例
	client := &client{
		resty: restyClient,
		opts:  o,
	}

	// 应用中间件
	for _, middleware := range o.requestMiddlewares {
		client.AddRequestMiddleware(middleware)
	}
	for _, middleware := range o.responseMiddlewares {
		client.AddResponseMiddleware(middleware)
	}
	for _, middleware := range o.errorMiddlewares {
		client.AddErrorMiddleware(middleware)
	}

	return client
}

// Invoke 执行HTTP请求
func (c *client) Invoke(ctx context.Context, method, path string, args interface{}, reply interface{}, opts ...CallOption) error {
	// 创建调用上下文
	callOpts := callOptions{
		operation:    "",
		pathTemplate: path,
		headers:      make(map[string]string),
	}

	// 应用调用选项
	for _, opt := range opts {
		opt(&callOpts)
	}

	// 创建请求
	req := c.resty.R().SetContext(ctx)

	// 添加调用特定的headers
	for key, value := range callOpts.headers {
		req.SetHeader(key, value)
	}

	// 设置请求body
	if args != nil {
		req.SetBody(args)
	}

	// 设置响应对象
	if reply != nil {
		req.SetResult(reply)
	}

	// 设置错误响应处理
	req.SetError(&HTTPError{})

	// 执行请求
	var resp *resty.Response
	var err error

	switch strings.ToUpper(method) {
	case http.MethodGet:
		resp, err = req.Get(path)
	case http.MethodPost:
		resp, err = req.Post(path)
	case http.MethodPut:
		resp, err = req.Put(path)
	case http.MethodDelete:
		resp, err = req.Delete(path)
	case http.MethodPatch:
		resp, err = req.Patch(path)
	case http.MethodHead:
		resp, err = req.Head(path)
	case http.MethodOptions:
		resp, err = req.Options(path)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return err
	}

	// 检查HTTP状态码
	if resp.IsError() {
		if errorResp := resp.Error(); errorResp != nil {
			if httpErr, ok := errorResp.(*HTTPError); ok {
				httpErr.Code = resp.StatusCode()
				return httpErr
			}
		}
		return &HTTPError{
			Code:    resp.StatusCode(),
			Message: resp.Status(),
		}
	}

	return nil
}

// AddRequestMiddleware 添加请求中间件
func (c *client) AddRequestMiddleware(middleware RestyRequestMiddleware) {
	c.resty.OnBeforeRequest(func(client *resty.Client, req *resty.Request) error {
		return middleware(client, req)
	})
}

// AddResponseMiddleware 添加响应中间件
func (c *client) AddResponseMiddleware(middleware RestyResponseMiddleware) {
	c.resty.OnAfterResponse(func(client *resty.Client, resp *resty.Response) error {
		return middleware(client, resp)
	})
}

// AddErrorMiddleware 添加错误中间件
func (c *client) AddErrorMiddleware(middleware RestyErrorMiddleware) {
	c.resty.OnError(func(req *resty.Request, err error) {
		middleware(req, err)
	})
}

// GetRestyClient 获取底层的resty客户端（用于高级用法）
func (c *client) GetRestyClient() *resty.Client {
	return c.resty
}

// WithRequestMiddleware 客户端选项：添加请求中间件
func WithRequestMiddleware(middlewares ...RestyRequestMiddleware) ClientOption {
	return func(o *clientOptions) {
		// 这些中间件会在NewClient中应用到resty客户端
		if o.requestMiddlewares == nil {
			o.requestMiddlewares = make([]RestyRequestMiddleware, 0)
		}
		o.requestMiddlewares = append(o.requestMiddlewares, middlewares...)
	}
}

// WithResponseMiddleware 客户端选项：添加响应中间件
func WithResponseMiddleware(middlewares ...RestyResponseMiddleware) ClientOption {
	return func(o *clientOptions) {
		if o.responseMiddlewares == nil {
			o.responseMiddlewares = make([]RestyResponseMiddleware, 0)
		}
		o.responseMiddlewares = append(o.responseMiddlewares, middlewares...)
	}
}

// WithErrorMiddleware 客户端选项：添加错误中间件
func WithErrorMiddleware(middlewares ...RestyErrorMiddleware) ClientOption {
	return func(o *clientOptions) {
		if o.errorMiddlewares == nil {
			o.errorMiddlewares = make([]RestyErrorMiddleware, 0)
		}
		o.errorMiddlewares = append(o.errorMiddlewares, middlewares...)
	}
}
