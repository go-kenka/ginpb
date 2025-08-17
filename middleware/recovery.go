package middleware

import (
	"fmt"
	"io"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// RecoveryConfig defines the config for Recovery middleware
type RecoveryConfig struct {
	// Skip defines a function to skip middleware
	Skipper func(*gin.Context) bool

	// Recovery handler function
	RecoveryHandler func(*gin.Context, interface{})

	// Enable stack trace in response
	EnableStackTrace bool

	// Enable detailed error information
	EnableDetailedError bool
}

// DefaultRecoveryConfig returns a default recovery configuration
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		Skipper:             nil,
		RecoveryHandler:     defaultRecoveryHandler,
		EnableStackTrace:    false,
		EnableDetailedError: false,
	}
}

// defaultRecoveryHandler is the default recovery handler
func defaultRecoveryHandler(c *gin.Context, err interface{}) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "internal server error",
		"message": "an unexpected error occurred",
	})
	c.Abort()
}

// Recovery returns a gin middleware that recovers from panics
func Recovery() gin.HandlerFunc {
	return RecoveryWithConfig(DefaultRecoveryConfig())
}

// RecoveryWithConfig returns a gin middleware that recovers from panics with config
func RecoveryWithConfig(config RecoveryConfig) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip middleware if skipper returns true
		if config.Skipper != nil && config.Skipper(c) {
			c.Next()
			return
		}

		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := debug.Stack()

				// Log the panic
				fmt.Printf("[Recovery] panic recovered:\n%s\n%s\n", err, stack)

				// Create detailed error response if enabled
				if config.EnableDetailedError {
					response := gin.H{
						"error":     "panic recovered",
						"message":   fmt.Sprintf("%v", err),
						"operation": c.GetString("operation"),
						"path":      c.Request.URL.Path,
						"method":    c.Request.Method,
					}

					if config.EnableStackTrace {
						response["stack_trace"] = string(stack)
					}

					c.JSON(http.StatusInternalServerError, response)
				} else {
					// Use custom recovery handler
					config.RecoveryHandler(c, err)
				}

				// Abort the request
				c.Abort()
			}
		}()

		c.Next()
	})
}

// RecoveryWithWriter returns a gin middleware that recovers from panics and writes to specified writer
func RecoveryWithWriter(out io.Writer) gin.HandlerFunc {
	return gin.RecoveryWithWriter(out)
}

// RecoveryFunc returns a gin middleware that recovers from panics with custom recovery function
func RecoveryFunc(f gin.RecoveryFunc) gin.HandlerFunc {
	return gin.Recovery()
}
