package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSConfig defines the config for CORS middleware
type CORSConfig struct {
	// Skip defines a function to skip middleware
	Skipper func(*gin.Context) bool

	// AllowOrigins defines the allowed origins
	AllowOrigins []string

	// AllowMethods defines the allowed HTTP methods
	AllowMethods []string

	// AllowHeaders defines the allowed headers
	AllowHeaders []string

	// ExposeHeaders defines the headers that are safe to expose
	ExposeHeaders []string

	// MaxAge indicates how long the results of a preflight request can be cached
	MaxAge time.Duration

	// AllowCredentials indicates whether the request can include user credentials
	AllowCredentials bool

	// AllowAllOrigins allows any origin
	AllowAllOrigins bool

	// AllowWildcard allows wildcard in AllowOrigins
	AllowWildcard bool
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		Skipper:          nil,
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{},
		MaxAge:           12 * time.Hour,
		AllowCredentials: false,
		AllowAllOrigins:  false,
		AllowWildcard:    false,
	}
}

// CORS returns a CORS middleware with default configuration
func CORS() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS middleware with custom configuration
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	// Normalize configuration
	if config.AllowAllOrigins {
		config.AllowOrigins = []string{"*"}
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip middleware if skipper returns true
		if config.Skipper != nil && config.Skipper(c) {
			c.Next()
			return
		}

		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := config.AllowAllOrigins || isOriginAllowed(origin, config.AllowOrigins, config.AllowWildcard)

		if allowed {
			// Set Access-Control-Allow-Origin header
			if config.AllowAllOrigins || contains(config.AllowOrigins, "*") {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		// Set Access-Control-Allow-Credentials header
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Set Access-Control-Expose-Headers header
		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ","))
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			// Set Access-Control-Allow-Methods header
			if len(config.AllowMethods) > 0 {
				c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ","))
			}

			// Set Access-Control-Allow-Headers header
			if len(config.AllowHeaders) > 0 {
				c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ","))
			} else {
				// If no specific headers configured, use requested headers
				requestedHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
				if requestedHeaders != "" {
					c.Header("Access-Control-Allow-Headers", requestedHeaders)
				}
			}

			// Set Access-Control-Max-Age header
			if config.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", strconv.Itoa(int(config.MaxAge.Seconds())))
			}

			c.Status(http.StatusNoContent)
			return
		}

		c.Next()
	})
}

// isOriginAllowed checks if origin is in allowed origins list
func isOriginAllowed(origin string, allowedOrigins []string, allowWildcard bool) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
		if allowWildcard && matchWildcard(origin, allowed) {
			return true
		}
	}
	return false
}

// matchWildcard matches origin against wildcard pattern
func matchWildcard(origin, pattern string) bool {
	if !strings.Contains(pattern, "*") {
		return origin == pattern
	}

	// Simple wildcard matching - replace * with .*
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	matched, _ := regexp.MatchString("^"+pattern+"$", origin)
	return matched
}

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
