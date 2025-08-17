package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-kenka/ginpb/example/api"
	"github.com/go-kenka/ginpb/middleware"
)

// ExampleService implements the generated service interface
type ExampleService struct{}

// TestAllBindings handles the comprehensive binding test
func (s *ExampleService) TestAllBindings(ctx context.Context, req *api.AllBindingsRequest) (*api.AllBindingsResponse, error) {
	return &api.AllBindingsResponse{
		Success: true,
		Message: "Request processed successfully",
		ReceivedData: map[string]string{
			"path_param": req.PathParam,
			"title":      req.Title,
			"email":      req.Email,
		},
	}, nil
}

// TestQueryBinding handles query parameter binding
func (s *ExampleService) TestQueryBinding(ctx context.Context, req *api.QueryBindingRequest) (*api.QueryBindingResponse, error) {
	return &api.QueryBindingResponse{
		TotalCount: int32(len(req.Categories)),
		Results:    req.Categories,
	}, nil
}

// TestPathBinding handles path parameter binding
func (s *ExampleService) TestPathBinding(ctx context.Context, req *api.PathBindingRequest) (*api.PathBindingResponse, error) {
	return &api.PathBindingResponse{
		Id:       req.Id,
		Name:     req.Name,
		FullPath: "/test/path/" + req.Name,
	}, nil
}

// TestHeaderBinding handles header binding
func (s *ExampleService) TestHeaderBinding(ctx context.Context, req *api.HeaderBindingRequest) (*api.HeaderBindingResponse, error) {
	return &api.HeaderBindingResponse{
		Headers: map[string]string{
			"User-Agent": req.UserAgent,
			"X-API-Key":  req.XApiKey,
		},
		RequestInfo: "Headers processed successfully",
	}, nil
}

// TestJSONBinding handles JSON binding
func (s *ExampleService) TestJSONBinding(ctx context.Context, req *api.JSONBindingRequest) (*api.JSONBindingResponse, error) {
	return &api.JSONBindingResponse{
		ProcessedData:   "User: " + req.Name + ", Email: " + req.Email,
		ValidationScore: int32(len(req.Hobbies) * 10),
	}, nil
}

// TestFormBinding handles form binding
func (s *ExampleService) TestFormBinding(ctx context.Context, req *api.FormBindingRequest) (*api.FormBindingResponse, error) {
	return &api.FormBindingResponse{
		RegistrationSuccess: true,
		UserId:              "user_" + req.Username,
		Warnings:            []string{},
	}, nil
}

// TestMixedBinding handles mixed binding types
func (s *ExampleService) TestMixedBinding(ctx context.Context, req *api.MixedBindingRequest) (*api.MixedBindingResponse, error) {
	return &api.MixedBindingResponse{
		OperationId: "op_" + req.ResourceId,
		Status:      "completed",
		Result: map[string]string{
			"action": req.ActionType,
			"name":   req.OperationName,
		},
		ValidationErrors: []string{},
	}, nil
}

// Custom middleware examples
func customLoggingMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log after request
		duration := time.Since(start)
		log.Printf("Custom Log: %s %s - %d - %v",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration)
	})
}

func authMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Simple token validation (in practice, use proper JWT validation)
		if token != "Bearer valid-token" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", "authenticated_user")
		c.Next()
	})
}

func main() {
	// Create gin router
	r := gin.Default()

	// Create service instance
	service := &ExampleService{}

	// Example 1: Basic registration without middleware
	basicGroup := r.Group("/basic")
	api.RegisterGinBindingTestServiceHTTPServer(basicGroup, service)

	// Example 2: Registration with global middleware
	middlewareGroup := r.Group("/middleware")
	api.RegisterGinBindingTestServiceHTTPServerWithMiddleware(
		middlewareGroup,
		service,
		middleware.Logging(),
		middleware.Recovery(),
		customLoggingMiddleware(),
	)

	// Example 3: Registration with CORS middleware
	corsGroup := r.Group("/cors")
	api.RegisterGinBindingTestServiceHTTPServerWithMiddleware(
		corsGroup,
		service,
		middleware.CORS(),
		middleware.Logging(),
	)

	// Example 4: Registration with authentication middleware
	authGroup := r.Group("/auth")
	api.RegisterGinBindingTestServiceHTTPServerWithMiddleware(
		authGroup,
		service,
		authMiddleware(),
		middleware.Logging(),
	)

	// Example 5: Registration with operation-specific middleware
	operationGroup := r.Group("/operation-specific")
	operationMiddlewares := map[string][]gin.HandlerFunc{
		api.OperationGinBindingTestServiceTestJSONBinding: {
			authMiddleware(),
			middleware.LoggingWithConfig(middleware.LoggingConfig{
				LogRequest:  true,
				LogResponse: true,
			}),
		},
		api.OperationGinBindingTestServiceTestMixedBinding: {
			middleware.BearerAuth(),
			middleware.Recovery(),
		},
	}
	api.RegisterGinBindingTestServiceHTTPServerWithOperationMiddleware(
		operationGroup,
		service,
		operationMiddlewares,
	)

	// Example 6: Using middleware selectors
	selectorGroup := r.Group("/selector")
	selectorGroup.Use(
		middleware.NewConditionalMiddleware(
			middleware.PathSelector{Path: "/selector/test/json"},
			authMiddleware(),
		).Apply(),
		middleware.NewOperationMiddleware(
			api.OperationGinBindingTestServiceTestAllBindings,
			middleware.LoggingWithConfig(middleware.LoggingConfig{
				LogRequest:  true,
				LogResponse: true,
			}),
		).Apply(),
	)
	api.RegisterGinBindingTestServiceHTTPServer(selectorGroup, service)

	// Example 7: Using middleware groups
	groupedGroup := r.Group("/grouped")
	middlewareGroup := middleware.NewMiddlewareGroup(
		middleware.Recovery(),
		middleware.Logging(),
		customLoggingMiddleware(),
	)
	groupedGroup.Use(middlewareGroup.Apply()...)
	api.RegisterGinBindingTestServiceHTTPServer(groupedGroup, service)

	// Start server
	log.Println("Server starting on :8080")
	log.Println("Available endpoints:")
	log.Println("  Basic (no middleware):     http://localhost:8080/basic/test/...")
	log.Println("  With middleware:           http://localhost:8080/middleware/test/...")
	log.Println("  With CORS:                 http://localhost:8080/cors/test/...")
	log.Println("  With auth (needs token):   http://localhost:8080/auth/test/...")
	log.Println("  Operation-specific:        http://localhost:8080/operation-specific/test/...")
	log.Println("  With selectors:            http://localhost:8080/selector/test/...")
	log.Println("  Grouped middleware:        http://localhost:8080/grouped/test/...")

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
