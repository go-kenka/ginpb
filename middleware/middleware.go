package middleware

import (
	"github.com/gin-gonic/gin"
)

// Handler defines the handler used by ginpb middleware as return value
type Handler func(*gin.Context)

// Middleware defines a function to process middleware with gin.HandlerFunc style
type Middleware func(...gin.HandlerFunc) gin.HandlerFunc

// Chain creates a new middleware chain by composing multiple gin.HandlerFunc middleware
func Chain(middleware ...gin.HandlerFunc) []gin.HandlerFunc {
	return middleware
}

// MiddlewareFunc is an adapter to allow the use of ordinary functions as middleware
type MiddlewareFunc func(*gin.Context)

// HandlerFunc returns the gin.HandlerFunc representation
func (m MiddlewareFunc) HandlerFunc() gin.HandlerFunc {
	return gin.HandlerFunc(m)
}

// OperationMiddleware allows applying middleware based on operation name (similar to kratos)
type OperationMiddleware struct {
	middleware gin.HandlerFunc
	operation  string
}

// NewOperationMiddleware creates middleware that only applies to specific operations
func NewOperationMiddleware(operation string, middleware gin.HandlerFunc) *OperationMiddleware {
	return &OperationMiddleware{
		middleware: middleware,
		operation:  operation,
	}
}

// Apply applies the middleware if the operation matches
func (om *OperationMiddleware) Apply() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get operation from context (set by generated handler)
		if op, exists := c.Get("operation"); exists && op == om.operation {
			om.middleware(c)
		} else {
			c.Next()
		}
	})
}

// Selector defines middleware selector for conditional application
type Selector interface {
	Match(c *gin.Context) bool
}

// PathSelector matches based on request path
type PathSelector struct {
	Path string
}

func (ps PathSelector) Match(c *gin.Context) bool {
	return c.Request.URL.Path == ps.Path
}

// MethodSelector matches based on HTTP method
type MethodSelector struct {
	Method string
}

func (ms MethodSelector) Match(c *gin.Context) bool {
	return c.Request.Method == ms.Method
}

// ConditionalMiddleware applies middleware based on selector conditions
type ConditionalMiddleware struct {
	middleware gin.HandlerFunc
	selector   Selector
}

// NewConditionalMiddleware creates middleware that applies conditionally
func NewConditionalMiddleware(selector Selector, middleware gin.HandlerFunc) *ConditionalMiddleware {
	return &ConditionalMiddleware{
		middleware: middleware,
		selector:   selector,
	}
}

// Apply applies the middleware if selector matches
func (cm *ConditionalMiddleware) Apply() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if cm.selector.Match(c) {
			cm.middleware(c)
		} else {
			c.Next()
		}
	})
}

// MiddlewareGroup allows grouping multiple middleware
type MiddlewareGroup struct {
	middleware []gin.HandlerFunc
}

// NewMiddlewareGroup creates a new middleware group
func NewMiddlewareGroup(middleware ...gin.HandlerFunc) *MiddlewareGroup {
	return &MiddlewareGroup{
		middleware: middleware,
	}
}

// Add adds middleware to the group
func (mg *MiddlewareGroup) Add(middleware ...gin.HandlerFunc) *MiddlewareGroup {
	mg.middleware = append(mg.middleware, middleware...)
	return mg
}

// Apply returns all middleware in the group as a slice
func (mg *MiddlewareGroup) Apply() []gin.HandlerFunc {
	return mg.middleware
}

// Wrap wraps multiple middleware into a single HandlerFunc
func (mg *MiddlewareGroup) Wrap() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		for _, m := range mg.middleware {
			if c.IsAborted() {
				return
			}
			m(c)
		}
	})
}
