package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingConfig defines the config for Logging middleware
type LoggingConfig struct {
	// Optional. Default value is gin.DefaultWriter
	Output io.Writer

	// Skip defines a function to skip middleware
	Skipper func(*gin.Context) bool

	// Fields to log
	LogLatency   bool
	LogMethod    bool
	LogPath      bool
	LogStatus    bool
	LogUserAgent bool
	LogClientIP  bool
	LogReferer   bool
	LogOperation bool
	LogRequest   bool
	LogResponse  bool
}

// DefaultLoggingConfig returns a default logging configuration
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Output:       gin.DefaultWriter,
		Skipper:      nil,
		LogLatency:   true,
		LogMethod:    true,
		LogPath:      true,
		LogStatus:    true,
		LogUserAgent: true,
		LogClientIP:  true,
		LogReferer:   false,
		LogOperation: true,
		LogRequest:   false,
		LogResponse:  false,
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Latency   string      `json:"latency,omitempty"`
	Method    string      `json:"method,omitempty"`
	Path      string      `json:"path,omitempty"`
	Status    int         `json:"status,omitempty"`
	UserAgent string      `json:"user_agent,omitempty"`
	ClientIP  string      `json:"client_ip,omitempty"`
	Referer   string      `json:"referer,omitempty"`
	Operation string      `json:"operation,omitempty"`
	Request   interface{} `json:"request,omitempty"`
	Response  interface{} `json:"response,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// responseBodyWriter wraps gin.ResponseWriter to capture response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Logging returns a gin middleware for logging requests and responses
func Logging() gin.HandlerFunc {
	return LoggingWithConfig(DefaultLoggingConfig())
}

// LoggingWithConfig returns a gin middleware for logging with custom config
func LoggingWithConfig(config LoggingConfig) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip middleware if skipper returns true
		if config.Skipper != nil && config.Skipper(c) {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Capture request body if needed
		var requestBody interface{}
		if config.LogRequest {
			if c.Request.Body != nil {
				bodyBytes, err := io.ReadAll(c.Request.Body)
				if err == nil {
					// Restore request body for further processing
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

					// Try to parse JSON
					var jsonBody interface{}
					if json.Unmarshal(bodyBytes, &jsonBody) == nil {
						requestBody = jsonBody
					} else {
						requestBody = string(bodyBytes)
					}
				}
			}
		}

		// Capture response body if needed
		var responseWriter *responseBodyWriter
		if config.LogResponse {
			responseWriter = &responseBodyWriter{
				body:           bytes.NewBufferString(""),
				ResponseWriter: c.Writer,
			}
			c.Writer = responseWriter
		}

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Create log entry
		entry := LogEntry{
			Timestamp: start.Format(time.RFC3339),
		}

		if config.LogLatency {
			entry.Latency = latency.String()
		}
		if config.LogMethod {
			entry.Method = method
		}
		if config.LogPath {
			entry.Path = path
		}
		if config.LogStatus {
			entry.Status = c.Writer.Status()
		}
		if config.LogUserAgent {
			entry.UserAgent = c.Request.UserAgent()
		}
		if config.LogClientIP {
			entry.ClientIP = c.ClientIP()
		}
		if config.LogReferer {
			entry.Referer = c.Request.Referer()
		}
		if config.LogOperation {
			if op, exists := c.Get("operation"); exists {
				entry.Operation = fmt.Sprintf("%v", op)
			}
		}
		if config.LogRequest && requestBody != nil {
			entry.Request = requestBody
		}
		if config.LogResponse && responseWriter != nil {
			var responseBody interface{}
			if responseWriter.body.Len() > 0 {
				bodyBytes := responseWriter.body.Bytes()
				// Try to parse JSON
				var jsonBody interface{}
				if json.Unmarshal(bodyBytes, &jsonBody) == nil {
					responseBody = jsonBody
				} else {
					responseBody = string(bodyBytes)
				}
			}
			entry.Response = responseBody
		}

		// Log errors if any
		if len(c.Errors) > 0 {
			entry.Error = c.Errors.String()
		}

		// Write log
		logBytes, _ := json.Marshal(entry)
		fmt.Fprintln(config.Output, string(logBytes))
	})
}
