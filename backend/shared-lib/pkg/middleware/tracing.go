package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Tracing returns a middleware that adds OpenTelemetry tracing to requests
func Tracing(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		// Extract trace context from incoming request
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Start a new span
		spanName := c.Request.Method + " " + c.FullPath()
		if spanName == " " {
			spanName = c.Request.Method + " " + c.Request.URL.Path
		}

		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.target", c.Request.URL.Path),
				attribute.String("http.host", c.Request.Host),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("net.peer.ip", c.ClientIP()),
			),
		)
		defer span.End()

		// Add request ID if present
		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			span.SetAttributes(attribute.String("request.id", requestID))
		}

		// Store span in context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Record response info
		status := c.Writer.Status()
		span.SetAttributes(
			attribute.Int("http.status_code", status),
			attribute.Int("http.response_content_length", c.Writer.Size()),
		)

		// Record errors
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("error.message", c.Errors.String()))
			span.RecordError(c.Errors[0])
		}

		// Mark span as error if status >= 400
		if status >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	}
}
