package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/femi-lawal/new_bank/backend/payment-service/internal/handler"
	"github.com/femi-lawal/new_bank/backend/payment-service/internal/model"
	"github.com/femi-lawal/new_bank/backend/payment-service/internal/repository"
	"github.com/femi-lawal/new_bank/backend/payment-service/internal/service"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/db"
	apperrors "github.com/femi-lawal/new_bank/backend/shared-lib/pkg/errors"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/kafka"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/logger"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/metrics"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/middleware"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/tracing"
	"github.com/gin-gonic/gin"
)

const serviceName = "payment-service"

func main() {
	// Initialize Logger
	logger.InitLogger(serviceName, true)
	slog.Info("Starting Payment Service")

	// Initialize Tracing
	tp, err := tracing.InitTracing(context.Background(), tracing.DefaultConfig(serviceName))
	if err != nil {
		slog.Warn("Failed to initialize tracing", "error", err)
	} else {
		defer func() { _ = tp.Shutdown(context.Background()) }()
	}

	// Connect to Database
	dbConfig := db.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5433"),
		User:     getEnv("DB_USER", "user"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "newbank_core"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	database, err := db.Connect(dbConfig)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		panic(err)
	}

	// Auto Migrate
	if err := database.AutoMigrate(&model.Payment{}); err != nil {
		slog.Error("Failed to migrate database", "error", err)
	}

	// Initialize Kafka Producer
	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	var producer *kafka.Producer

	producer = kafka.NewProducer(kafkaBrokers)
	if producer != nil {
		slog.Info("Kafka producer initialized")
	}

	// Wiring
	repo := repository.NewPaymentRepository(database)
	var svc *service.PaymentService
	if producer != nil {
		svc = service.NewPaymentServiceWithKafka(repo, producer)
	} else {
		svc = service.NewPaymentService(repo)
	}
	h := handler.NewPaymentHandler(svc)

	// Get JWT secret
	jwtSecret := requireEnv("JWT_SECRET")

	// Setup Router
	r := gin.Default()

	// ============================================
	// Global Middleware
	// ============================================
	r.Use(apperrors.ErrorMiddleware())
	r.Use(middleware.RequestLogger(serviceName))
	r.Use(middleware.Tracing(serviceName))
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit())
	r.Use(metrics.PrometheusMiddleware(serviceName))

	// ============================================
	// Public endpoints
	// ============================================
	r.GET("/metrics", metrics.MetricsHandler())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": serviceName,
			"kafka":   producer != nil,
		})
	})

	// ============================================
	// Protected endpoints
	// ============================================
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(jwtSecret))
	{
		api.POST("/transfer", h.MakeTransfer)
	}

	port := getEnv("PORT", "8083")
	slog.Info("Server listening", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// requireEnv returns the value of an environment variable or panics if not set.
func requireEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	slog.Error("Required environment variable not set", "key", key)
	panic("Required environment variable " + key + " is not set")
}
