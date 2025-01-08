package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method", "path"},
	)

	httpRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
		[]string{"service"},
	)

	// Business metrics
	paymentTransfersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_transfers_total",
			Help: "Total number of payment transfers",
		},
		[]string{"status"}, // success, failed
	)

	accountsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "accounts_created_total",
			Help: "Total number of accounts created",
		},
	)

	cardsIssuedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cards_issued_total",
			Help: "Total number of cards issued",
		},
	)

	authLoginTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_login_total",
			Help: "Total number of login attempts",
		},
		[]string{"status"}, // success, failed
	)

	cacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"service", "type"}, // hit, miss
	)
)

// PrometheusMiddleware returns a Gin middleware for Prometheus metrics
func PrometheusMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint itself
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		// Track in-flight requests
		httpRequestsInFlight.WithLabelValues(serviceName).Inc()
		defer httpRequestsInFlight.WithLabelValues(serviceName).Dec()

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		httpRequestsTotal.WithLabelValues(serviceName, c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(serviceName, c.Request.Method, path).Observe(duration)
	}
}

// MetricsHandler returns the Prometheus metrics handler for Gin
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// Business metric recording functions

// RecordPaymentTransfer records a payment transfer metric
func RecordPaymentTransfer(success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	paymentTransfersTotal.WithLabelValues(status).Inc()
}

// RecordAccountCreated records an account creation
func RecordAccountCreated() {
	accountsCreatedTotal.Inc()
}

// RecordCardIssued records a card issuance
func RecordCardIssued() {
	cardsIssuedTotal.Inc()
}

// RecordLogin records a login attempt
func RecordLogin(success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	authLoginTotal.WithLabelValues(status).Inc()
}

// RecordCacheHit records a cache hit
func RecordCacheHit(serviceName string) {
	cacheHitsTotal.WithLabelValues(serviceName, "hit").Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss(serviceName string) {
	cacheHitsTotal.WithLabelValues(serviceName, "miss").Inc()
}
