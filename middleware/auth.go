package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthConfig defines the config for authentication middleware
type AuthConfig struct {
	// Skip defines a function to skip middleware
	Skipper func(*gin.Context) bool

	// Realm is used for basic auth
	Realm string

	// Custom validator function
	Validator func(*gin.Context, string) bool

	// Error handler function
	ErrorHandler func(*gin.Context, error)
}

// DefaultAuthConfig returns a default authentication configuration
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Skipper:      nil,
		Realm:        "Restricted",
		Validator:    nil,
		ErrorHandler: defaultAuthErrorHandler,
	}
}

// defaultAuthErrorHandler is the default error handler for authentication middleware
func defaultAuthErrorHandler(c *gin.Context, err error) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"error":   "authentication failed",
		"message": err.Error(),
	})
	c.Abort()
}

// BasicAuth returns a basic authentication middleware
func BasicAuth(accounts gin.Accounts) gin.HandlerFunc {
	return gin.BasicAuth(accounts)
}

// BearerAuth returns a bearer token authentication middleware
func BearerAuth() gin.HandlerFunc {
	return BearerAuthWithConfig(DefaultAuthConfig())
}

// BearerAuthWithConfig returns a bearer token authentication middleware with config
func BearerAuthWithConfig(config AuthConfig) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip middleware if skipper returns true
		if config.Skipper != nil && config.Skipper(c) {
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		if auth == "" {
			config.ErrorHandler(c, fmt.Errorf("authorization header missing"))
			return
		}

		if !strings.HasPrefix(auth, "Bearer ") {
			config.ErrorHandler(c, fmt.Errorf("invalid authorization header format"))
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token == "" {
			config.ErrorHandler(c, fmt.Errorf("bearer token missing"))
			return
		}

		// Use custom validator if provided
		if config.Validator != nil {
			if !config.Validator(c, token) {
				config.ErrorHandler(c, fmt.Errorf("token validation failed"))
				return
			}
		}

		// Store token in context
		c.Set("token", token)
		c.Next()
	})
}

// APIKeyAuth returns an API key authentication middleware
func APIKeyAuth(validKeys map[string]bool) gin.HandlerFunc {
	config := DefaultAuthConfig()
	config.Validator = func(c *gin.Context, key string) bool {
		return validKeys[key]
	}
	return APIKeyAuthWithConfig(config, "X-API-Key")
}

// APIKeyAuthWithConfig returns an API key authentication middleware with config
func APIKeyAuthWithConfig(config AuthConfig, headerName string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip middleware if skipper returns true
		if config.Skipper != nil && config.Skipper(c) {
			c.Next()
			return
		}

		apiKey := c.GetHeader(headerName)
		if apiKey == "" {
			config.ErrorHandler(c, fmt.Errorf("API key missing"))
			return
		}

		// Use custom validator if provided
		if config.Validator != nil {
			if !config.Validator(c, apiKey) {
				config.ErrorHandler(c, fmt.Errorf("invalid API key"))
				return
			}
		}

		// Store API key in context
		c.Set("api_key", apiKey)
		c.Next()
	})
}

// JWTAuth returns a JWT authentication middleware
func JWTAuth(secretKey string) gin.HandlerFunc {
	config := DefaultAuthConfig()
	config.Validator = func(c *gin.Context, token string) bool {
		// This is a simplified JWT validation
		// In production, you should use a proper JWT library
		return validateJWTToken(token, secretKey)
	}
	return JWTAuthWithConfig(config)
}

// JWTAuthWithConfig returns a JWT authentication middleware with config
func JWTAuthWithConfig(config AuthConfig) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip middleware if skipper returns true
		if config.Skipper != nil && config.Skipper(c) {
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		if auth == "" {
			config.ErrorHandler(c, fmt.Errorf("authorization header missing"))
			return
		}

		var token string
		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		} else {
			token = auth
		}

		if token == "" {
			config.ErrorHandler(c, fmt.Errorf("JWT token missing"))
			return
		}

		// Use custom validator if provided
		if config.Validator != nil {
			if !config.Validator(c, token) {
				config.ErrorHandler(c, fmt.Errorf("JWT token validation failed"))
				return
			}
		}

		// Store token in context
		c.Set("jwt_token", token)
		c.Next()
	})
}

// validateJWTToken is a placeholder for JWT token validation
// In production, use a proper JWT library like github.com/golang-jwt/jwt
func validateJWTToken(token, secretKey string) bool {
	// This is a simplified validation - replace with proper JWT validation
	return len(token) > 0 && len(secretKey) > 0
}

// BasicAuthFromConfig creates basic auth middleware from username:password
func BasicAuthFromConfig(username, password string) gin.HandlerFunc {
	accounts := gin.Accounts{
		username: password,
	}
	return gin.BasicAuth(accounts)
}

// extractBasicAuth extracts username and password from Authorization header
func extractBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}

	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}

	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}

	return cs[:s], cs[s+1:], true
}
