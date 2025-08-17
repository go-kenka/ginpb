package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-kenka/ginpb/example/api"
	"github.com/go-kenka/ginpb/middleware"
	"github.com/stretchr/testify/assert"
)

// TestMiddlewareIntegration tests the middleware integration
func TestMiddlewareIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &ExampleService{}

	tests := []struct {
		name       string
		setupFn    func() *gin.Engine
		method     string
		url        string
		headers    map[string]string
		body       interface{}
		expectCode int
		expectBody interface{}
	}{
		{
			name: "basic_without_middleware",
			setupFn: func() *gin.Engine {
				r := gin.New()
				api.RegisterGinBindingTestServiceHTTPServer(r, service)
				return r
			},
			method:     "GET",
			url:        "/test/query?search=test&page=1",
			expectCode: http.StatusOK,
			expectBody: map[string]interface{}{
				"total_count": float64(0),
				"results":     []interface{}{},
			},
		},
		{
			name: "with_logging_middleware",
			setupFn: func() *gin.Engine {
				r := gin.New()
				api.RegisterGinBindingTestServiceHTTPServerWithMiddleware(
					r,
					service,
					middleware.Logging(),
				)
				return r
			},
			method:     "GET",
			url:        "/test/query?search=test&page=1",
			expectCode: http.StatusOK,
		},
		{
			name: "with_cors_middleware",
			setupFn: func() *gin.Engine {
				r := gin.New()
				api.RegisterGinBindingTestServiceHTTPServerWithMiddleware(
					r,
					service,
					middleware.CORS(),
				)
				return r
			},
			method: "OPTIONS",
			url:    "/test/query",
			headers: map[string]string{
				"Origin": "http://localhost:3000",
			},
			expectCode: http.StatusNoContent,
		},
		{
			name: "with_auth_middleware_valid_token",
			setupFn: func() *gin.Engine {
				r := gin.New()
				api.RegisterGinBindingTestServiceHTTPServerWithMiddleware(
					r,
					service,
					middleware.BearerAuth(),
				)
				return r
			},
			method: "GET",
			url:    "/test/query?search=test",
			headers: map[string]string{
				"Authorization": "Bearer valid-token",
			},
			expectCode: http.StatusOK,
		},
		{
			name: "with_auth_middleware_invalid_token",
			setupFn: func() *gin.Engine {
				r := gin.New()
				api.RegisterGinBindingTestServiceHTTPServerWithMiddleware(
					r,
					service,
					middleware.BearerAuthWithConfig(middleware.AuthConfig{
						Validator: func(c *gin.Context, token string) bool {
							return token == "valid-token"
						},
					}),
				)
				return r
			},
			method: "GET",
			url:    "/test/query?search=test",
			headers: map[string]string{
				"Authorization": "Bearer invalid-token",
			},
			expectCode: http.StatusUnauthorized,
		},
		{
			name: "with_recovery_middleware",
			setupFn: func() *gin.Engine {
				r := gin.New()
				api.RegisterGinBindingTestServiceHTTPServerWithMiddleware(
					r,
					service,
					middleware.Recovery(),
				)
				return r
			},
			method:     "GET",
			url:        "/test/query?search=test",
			expectCode: http.StatusOK,
		},
		{
			name: "operation_specific_middleware",
			setupFn: func() *gin.Engine {
				r := gin.New()
				operationMiddlewares := map[string][]gin.HandlerFunc{
					api.OperationGinBindingTestServiceTestJSONBinding: {
						middleware.LoggingWithConfig(middleware.LoggingConfig{
							LogRequest: true,
						}),
					},
				}
				api.RegisterGinBindingTestServiceHTTPServerWithOperationMiddleware(
					r,
					service,
					operationMiddlewares,
				)
				return r
			},
			method: "POST",
			url:    "/test/json",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			body: map[string]interface{}{
				"name":    "John Doe",
				"age":     30,
				"email":   "john@example.com",
				"hobbies": []string{"reading", "coding"},
			},
			expectCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.setupFn()

			var body bytes.Buffer
			if tt.body != nil {
				jsonBody, _ := json.Marshal(tt.body)
				body = *bytes.NewBuffer(jsonBody)
			}

			req := httptest.NewRequest(tt.method, tt.url, &body)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectCode, w.Code)

			if tt.expectBody != nil {
				var responseBody interface{}
				err := json.Unmarshal(w.Body.Bytes(), &responseBody)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectBody, responseBody)
			}
		})
	}
}

// TestMiddlewareSelectors tests middleware selectors
func TestMiddlewareSelectors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &ExampleService{}
	r := gin.New()

	// Setup middleware with selectors
	r.Use(
		middleware.NewConditionalMiddleware(
			middleware.PathSelector{Path: "/test/json"},
			gin.HandlerFunc(func(c *gin.Context) {
				c.Header("X-JSON-Middleware", "applied")
				c.Next()
			}),
		).Apply(),
		middleware.NewConditionalMiddleware(
			middleware.MethodSelector{Method: "POST"},
			gin.HandlerFunc(func(c *gin.Context) {
				c.Header("X-POST-Middleware", "applied")
				c.Next()
			}),
		).Apply(),
	)

	api.RegisterGinBindingTestServiceHTTPServer(r, service)

	tests := []struct {
		name             string
		method           string
		url              string
		expectHeaders    map[string]string
		notExpectHeaders []string
	}{
		{
			name:   "json_path_selector",
			method: "POST",
			url:    "/test/json",
			expectHeaders: map[string]string{
				"X-JSON-Middleware": "applied",
				"X-POST-Middleware": "applied",
			},
		},
		{
			name:   "post_method_selector_only",
			method: "POST",
			url:    "/test/form",
			expectHeaders: map[string]string{
				"X-POST-Middleware": "applied",
			},
			notExpectHeaders: []string{"X-JSON-Middleware"},
		},
		{
			name:             "no_selectors_match",
			method:           "GET",
			url:              "/test/query",
			notExpectHeaders: []string{"X-JSON-Middleware", "X-POST-Middleware"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if tt.method == "POST" {
				jsonBody, _ := json.Marshal(map[string]interface{}{
					"name":  "test",
					"email": "test@example.com",
				})
				body = *bytes.NewBuffer(jsonBody)
			}

			req := httptest.NewRequest(tt.method, tt.url, &body)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			for key, value := range tt.expectHeaders {
				assert.Equal(t, value, w.Header().Get(key), "Expected header %s to be %s", key, value)
			}

			for _, header := range tt.notExpectHeaders {
				assert.Empty(t, w.Header().Get(header), "Expected header %s to not be present", header)
			}
		})
	}
}

// TestMiddlewareChaining tests middleware chaining
func TestMiddlewareChaining(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &ExampleService{}
	r := gin.New()

	// Create middleware group
	middlewareGroup := middleware.NewMiddlewareGroup(
		gin.HandlerFunc(func(c *gin.Context) {
			c.Header("X-Middleware-1", "applied")
			c.Next()
		}),
		gin.HandlerFunc(func(c *gin.Context) {
			c.Header("X-Middleware-2", "applied")
			c.Next()
		}),
		gin.HandlerFunc(func(c *gin.Context) {
			c.Header("X-Middleware-3", "applied")
			c.Next()
		}),
	)

	api.RegisterGinBindingTestServiceHTTPServerWithMiddleware(
		r,
		service,
		middlewareGroup.Apply()...,
	)

	req := httptest.NewRequest("GET", "/test/query?search=test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "applied", w.Header().Get("X-Middleware-1"))
	assert.Equal(t, "applied", w.Header().Get("X-Middleware-2"))
	assert.Equal(t, "applied", w.Header().Get("X-Middleware-3"))
}
