package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPError HTTP错误类型
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error 实现error接口
func (e *HTTPError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("HTTP %d: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

// IsHTTPError 检查是否为HTTP错误
func IsHTTPError(err error) bool {
	_, ok := err.(*HTTPError)
	return ok
}

// GetHTTPStatusCode 获取HTTP状态码
func GetHTTPStatusCode(err error) int {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.Code
	}
	return 0
}

// NewHTTPError 创建HTTP错误
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
	}
}

// NewHTTPErrorWithDetails 创建带详细信息的HTTP错误
func NewHTTPErrorWithDetails(code int, message, details string) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// 常见的HTTP错误
var (
	ErrBadRequest          = NewHTTPError(http.StatusBadRequest, "Bad Request")
	ErrUnauthorized        = NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	ErrForbidden           = NewHTTPError(http.StatusForbidden, "Forbidden")
	ErrNotFound            = NewHTTPError(http.StatusNotFound, "Not Found")
	ErrMethodNotAllowed    = NewHTTPError(http.StatusMethodNotAllowed, "Method Not Allowed")
	ErrConflict            = NewHTTPError(http.StatusConflict, "Conflict")
	ErrInternalServerError = NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	ErrBadGateway          = NewHTTPError(http.StatusBadGateway, "Bad Gateway")
	ErrServiceUnavailable  = NewHTTPError(http.StatusServiceUnavailable, "Service Unavailable")
	ErrGatewayTimeout      = NewHTTPError(http.StatusGatewayTimeout, "Gateway Timeout")
)

// IsClientError 检查是否为客户端错误（4xx）
func IsClientError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.Code >= 400 && httpErr.Code < 500
	}
	return false
}

// IsServerError 检查是否为服务端错误（5xx）
func IsServerError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.Code >= 500 && httpErr.Code < 600
	}
	return false
}

// IsRetryableError 检查错误是否可重试
func IsRetryableError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		// 5xx错误通常可重试
		return httpErr.Code >= 500
	}
	return false
}

// ErrorFromResponse 从HTTP响应创建错误
func ErrorFromResponse(resp *http.Response) error {
	message := resp.Status
	if resp.StatusCode >= 400 {
		return &HTTPError{
			Code:    resp.StatusCode,
			Message: message,
		}
	}
	return nil
}

// 编码器和解码器类型定义
type (
	// ErrorDecoder 错误解码器
	ErrorDecoder func(resp *http.Response) error

	// RequestEncoder 请求编码器
	RequestEncoder func(ctx context.Context, contentType string, v interface{}) ([]byte, error)

	// ResponseDecoder 响应解码器
	ResponseDecoder func(resp *http.Response, v interface{}) error
)

// DefaultErrorDecoder 默认错误解码器
func DefaultErrorDecoder(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		return &HTTPError{
			Code:    resp.StatusCode,
			Message: resp.Status,
		}
	}
	return nil
}

// DefaultRequestEncoder 默认请求编码器
func DefaultRequestEncoder(ctx context.Context, contentType string, v interface{}) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

// DefaultResponseDecoder 默认响应解码器
func DefaultResponseDecoder(resp *http.Response, v interface{}) error {
	if v == nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return nil
	}

	return json.Unmarshal(body, v)
}
