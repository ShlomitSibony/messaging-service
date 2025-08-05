package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

const (
	RequestIDHeader = "X-Request-ID"
	RequestIDKey    = "request_id"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Header(RequestIDHeader, requestID)
		c.Set(RequestIDKey, requestID)
		c.Next()
	}
}

// LoggingMiddleware logs HTTP requests with request ID and timing
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Get request ID from context
		requestID, exists := c.Get(RequestIDKey)
		if !exists {
			requestID = "unknown"
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		status := c.Writer.Status()

		// Log with request ID
		logger.Info("HTTP Request",
			zap.String("request_id", requestID.(string)),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("remote_addr", c.ClientIP()),
		)
	}
}

// MetricsMiddleware records metrics using OpenTelemetry
func MetricsMiddleware() gin.HandlerFunc {
	meter := otel.GetMeterProvider().Meter("messaging-service")

	// Create metrics
	requestCounter, _ := meter.Int64Counter("http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("1"),
	)

	requestDuration, _ := meter.Float64Histogram("http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)

	requestSize, _ := meter.Int64Histogram("http_request_size_bytes",
		metric.WithDescription("HTTP request size in bytes"),
		metric.WithUnit("By"),
	)

	responseSize, _ := meter.Int64Histogram("http_response_size_bytes",
		metric.WithDescription("HTTP response size in bytes"),
		metric.WithUnit("By"),
	)

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Record request size
		if c.Request.ContentLength > 0 {
			requestSize.Record(context.Background(), c.Request.ContentLength,
				metric.WithAttributes(
					attribute.String("method", method),
					attribute.String("path", path),
				),
			)
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		// Record metrics
		requestCounter.Add(context.Background(), 1,
			metric.WithAttributes(
				attribute.String("method", method),
				attribute.String("path", path),
				attribute.Int("status", status),
			),
		)

		requestDuration.Record(context.Background(), duration,
			metric.WithAttributes(
				attribute.String("method", method),
				attribute.String("path", path),
				attribute.Int("status", status),
			),
		)

		// Record response size
		if c.Writer.Size() > 0 {
			responseSize.Record(context.Background(), int64(c.Writer.Size()),
				metric.WithAttributes(
					attribute.String("method", method),
					attribute.String("path", path),
					attribute.Int("status", status),
				),
			)
		}
	}
}
