package client

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

// RestyRequestMiddleware Resty请求中间件类型
type RestyRequestMiddleware func(*resty.Client, *resty.Request) error

// RestyResponseMiddleware Resty响应中间件类型
type RestyResponseMiddleware func(*resty.Client, *resty.Response) error

// RestyErrorMiddleware Resty错误中间件类型
type RestyErrorMiddleware func(*resty.Request, error)

// LoggingRequestMiddleware 日志请求中间件
func LoggingRequestMiddleware(logger func(format string, args ...interface{})) RestyRequestMiddleware {
	return func(c *resty.Client, req *resty.Request) error {
		logger("🚀 Request: %s %s", req.Method, req.URL)

		// 记录请求开始时间
		req.SetContext(req.Context())
		req.Header.Set("X-Request-Start", fmt.Sprintf("%d", time.Now().UnixNano()))

		return nil
	}
}

// LoggingResponseMiddleware 日志响应中间件
func LoggingResponseMiddleware(logger func(format string, args ...interface{})) RestyResponseMiddleware {
	return func(c *resty.Client, resp *resty.Response) error {
		startTime := resp.Request.Header.Get("X-Request-Start")
		if startTime != "" {
			// 计算耗时
			if startNano, err := strconv.ParseInt(startTime, 10, 64); err == nil {
				start := time.Unix(0, startNano)
				duration := time.Since(start)
				logger("✅ Response: %s %s - %d (%v)",
					resp.Request.Method, resp.Request.URL, resp.StatusCode(), duration)
			} else {
				logger("✅ Response: %s %s - %d",
					resp.Request.Method, resp.Request.URL, resp.StatusCode())
			}
		} else {
			logger("✅ Response: %s %s - %d",
				resp.Request.Method, resp.Request.URL, resp.StatusCode())
		}
		return nil
	}
}

// LoggingErrorMiddleware 日志错误中间件
func LoggingErrorMiddleware(logger func(format string, args ...interface{})) RestyErrorMiddleware {
	return func(req *resty.Request, err error) {
		logger("❌ Request Error: %s %s - %v", req.Method, req.URL, err)
	}
}

// AuthRequestMiddleware 认证请求中间件
func AuthRequestMiddleware(token string) RestyRequestMiddleware {
	return func(c *resty.Client, req *resty.Request) error {
		if token != "" {
			req.SetHeader("Authorization", "Bearer "+token)
		}
		return nil
	}
}

// RetryMiddleware 重试中间件（使用resty内置重试）
func RetryMiddleware(maxRetries int) ClientOption {
	return func(o *clientOptions) {
		// 这个会在NewClient中设置到resty客户端
	}
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() RestyRequestMiddleware {
	return func(c *resty.Client, req *resty.Request) error {
		if req.Header.Get("X-Request-ID") == "" {
			// 生成请求ID（简单实现）
			requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())
			req.SetHeader("X-Request-ID", requestID)
		}
		return nil
	}
}

// TimingMiddleware 计时中间件
func TimingMiddleware(logger func(format string, args ...interface{})) struct {
	Request  RestyRequestMiddleware
	Response RestyResponseMiddleware
} {
	return struct {
		Request  RestyRequestMiddleware
		Response RestyResponseMiddleware
	}{
		Request: func(c *resty.Client, req *resty.Request) error {
			req.Header.Set("X-Timing-Start", fmt.Sprintf("%d", time.Now().UnixNano()))
			return nil
		},
		Response: func(c *resty.Client, resp *resty.Response) error {
			startHeader := resp.Request.Header.Get("X-Timing-Start")
			if startHeader != "" {
				if startNano, err := strconv.ParseInt(startHeader, 10, 64); err == nil {
					start := time.Unix(0, startNano)
					duration := time.Since(start)
					logger("⏱️  Request Duration: %s %s took %v",
						resp.Request.Method, resp.Request.URL, duration)
				}
			}
			return nil
		},
	}
}

// CircuitBreakerMiddleware 熔断中间件
func CircuitBreakerMiddleware(threshold int) RestyRequestMiddleware {
	failures := 0
	lastFailTime := time.Time{}

	return func(c *resty.Client, req *resty.Request) error {
		// 简单的熔断逻辑
		if failures >= threshold {
			if time.Since(lastFailTime) < 30*time.Second {
				return fmt.Errorf("circuit breaker open: too many failures")
			}
			// 重置计数器
			failures = 0
		}

		// 在错误中间件中处理失败计数
		c.OnError(func(req *resty.Request, err error) {
			failures++
			lastFailTime = time.Now()
		})

		return nil
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(requestsPerSecond int) RestyRequestMiddleware {
	lastRequest := time.Now()
	minInterval := time.Duration(1000/requestsPerSecond) * time.Millisecond

	return func(c *resty.Client, req *resty.Request) error {
		elapsed := time.Since(lastRequest)
		if elapsed < minInterval {
			time.Sleep(minInterval - elapsed)
		}
		lastRequest = time.Now()
		return nil
	}
}

// HeaderMiddleware 添加自定义头部的中间件
func HeaderMiddleware(headers map[string]string) RestyRequestMiddleware {
	return func(c *resty.Client, req *resty.Request) error {
		for key, value := range headers {
			req.SetHeader(key, value)
		}
		return nil
	}
}

// 预定义的中间件组合
func DefaultMiddleware() struct {
	Request  []RestyRequestMiddleware
	Response []RestyResponseMiddleware
	Error    []RestyErrorMiddleware
} {
	logger := log.Printf

	return struct {
		Request  []RestyRequestMiddleware
		Response []RestyResponseMiddleware
		Error    []RestyErrorMiddleware
	}{
		Request: []RestyRequestMiddleware{
			RequestIDMiddleware(),
			LoggingRequestMiddleware(logger),
		},
		Response: []RestyResponseMiddleware{
			LoggingResponseMiddleware(logger),
		},
		Error: []RestyErrorMiddleware{
			LoggingErrorMiddleware(logger),
		},
	}
}
