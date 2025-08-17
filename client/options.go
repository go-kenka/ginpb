package client

import (
	"net/http"
	"time"
)

// ClientOption 客户端配置选项函数类型
type ClientOption func(*clientOptions)

// CallOption 调用选项函数类型
type CallOption func(*callOptions)

// callOptions 调用选项
type callOptions struct {
	operation    string
	pathTemplate string
	headers      map[string]string
}

// WithEndpoint 设置服务端点
func WithEndpoint(endpoint string) ClientOption {
	return func(o *clientOptions) {
		o.endpoint = endpoint
	}
}

// WithTimeout 设置请求超时时间
func WithTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

// WithUserAgent 设置User-Agent
func WithUserAgent(userAgent string) ClientOption {
	return func(o *clientOptions) {
		o.userAgent = userAgent
	}
}

// WithRetry 设置重试参数
func WithRetry(count int, waitTime, maxWaitTime time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.retryCount = count
		o.retryWaitTime = waitTime
		o.retryMaxWaitTime = maxWaitTime
	}
}

// WithRetryCount 设置重试次数
func WithRetryCount(count int) ClientOption {
	return func(o *clientOptions) {
		o.retryCount = count
	}
}

// WithErrorDecoder 设置错误解码器
func WithErrorDecoder(decoder ErrorDecoder) ClientOption {
	return func(o *clientOptions) {
		o.errorDecoder = decoder
	}
}

// WithRequestEncoder 设置请求编码器
func WithRequestEncoder(encoder RequestEncoder) ClientOption {
	return func(o *clientOptions) {
		o.encoder = encoder
	}
}

// WithResponseDecoder 设置响应解码器
func WithResponseDecoder(decoder ResponseDecoder) ClientOption {
	return func(o *clientOptions) {
		o.decoder = decoder
	}
}

// WithTransport 设置HTTP传输
func WithTransport(transport http.RoundTripper) ClientOption {
	return func(o *clientOptions) {
		o.transport = transport
	}
}

// WithHeader 添加默认请求头
func WithHeader(key, value string) ClientOption {
	return func(o *clientOptions) {
		o.headers[key] = value
	}
}

// Operation 设置操作名称
func Operation(operation string) CallOption {
	return func(o *callOptions) {
		o.operation = operation
	}
}

// PathTemplate 设置路径模板
func PathTemplate(pathTemplate string) CallOption {
	return func(o *callOptions) {
		o.pathTemplate = pathTemplate
	}
}

// Header 设置请求头（针对单次调用）
func Header(key, value string) CallOption {
	return func(o *callOptions) {
		o.headers[key] = value
	}
}

// ContentType 设置Content-Type
func ContentType(contentType string) CallOption {
	return func(o *callOptions) {
		o.headers["Content-Type"] = contentType
	}
}

// Accept 设置Accept头
func Accept(accept string) CallOption {
	return func(o *callOptions) {
		o.headers["Accept"] = accept
	}
}

// Authorization 设置Authorization头
func Authorization(auth string) CallOption {
	return func(o *callOptions) {
		o.headers["Authorization"] = auth
	}
}

// BearerToken 设置Bearer Token
func BearerToken(token string) CallOption {
	return func(o *callOptions) {
		o.headers["Authorization"] = "Bearer " + token
	}
}

// BasicAuth 设置基础认证
func BasicAuth(username, password string) CallOption {
	return func(o *callOptions) {
		o.headers["Authorization"] = BasicAuthValue(username, password)
	}
}
