package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/femi-lawal/new_bank/backend/ledger-service/internal/consumer"
	"github.com/femi-lawal/new_bank/backend/ledger-service/internal/handler"
	"github.com/femi-lawal/new_bank/backend/ledger-service/internal/model"
	"github.com/femi-lawal/new_bank/backend/ledger-service/internal/repository"
	"github.com/femi-lawal/new_bank/backend/ledger-service/internal/service"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/cache"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/db"
	apperrors "github.com/femi-lawal/new_bank/backend/shared-lib/pkg/errors"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/kafka"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/logger"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/metrics"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/middleware"
	"github.com/gin-gonic/gin"
)

const serviceName = "ledger-service"

func main() {
	// Initialize Logger
	logger.InitLogger(serviceName, true)
	slog.Info("Starting Ledger Service")

	// Connect to Database
	dbConfig := db.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5433"),
		User:     getEnv("DB_USER", "user"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "newbank_core"),
		SSLMode:  "disable",
	}

	database, err := db.Connect(dbConfig)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		panic(err)
	}

	// Auto Migrate
	if err := database.AutoMigrate(&model.Account{}, &model.JournalEntry{}, &model.Posting{}); err != nil {
		slog.Error("Failed to migrate database", "error", err)
	}

	// Initialize Redis Cache
	var redisClient *cache.RedisClient
	redisCfg := cache.Config{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	}
	redisClient, err = cache.NewRedisClient(redisCfg)
	if err != nil {
		slog.Warn("Redis connection failed, caching disabled", "error", err)
	} else {
		slog.Info("Redis cache connected")
	}

	// Wiring
	repo := repository.NewLedgerRepository(database)
	var svc *service.LedgerService
	if redisClient != nil {
		svc = service.NewLedgerServiceWithCache(repo, redisClient)
	} else {
		svc = service.NewLedgerService(repo)
	}
	h := handler.NewLedgerHandler(svc)

	// Initialize Kafka
	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	var producer *kafka.Producer

	producer = kafka.NewProducer(kafkaBrokers)
	if producer != nil {
		slog.Info("Kafka producer initialized")
	}

	// Start Kafka consumer for payment events
	go func() {
		paymentConsumer := consumer.NewPaymentConsumer(kafkaBrokers, svc, producer)
		if paymentConsumer != nil {
			if err := paymentConsumer.Start(context.Background()); err != nil {
				slog.Error("Kafka consumer error", "error", err)
			}
		}
	}()

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		slog.Info("Shutting down gracefully...")
		if producer != nil {
			producer.Close()
		}
		os.Exit(0)
	}()

	// Get JWT secret for auth
	jwtSecret := getEnv("JWT_SECRET", "my-secret-key")

	// Setup Router
	r := gin.Default()

	// ============================================
	// Global Middleware
	// ============================================
	r.Use(apperrors.ErrorMiddleware())
	r.Use(middleware.RequestLogger(serviceName))
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
			"redis":   redisClient != nil,
			"kafka":   producer != nil,
		})
	})

	// ============================================
	// Protected endpoints
	// ============================================
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(jwtSecret))
	{
		api.POST("/accounts", h.CreateAccount)
		api.GET("/accounts", h.ListAccounts)
		api.POST("/transactions", h.PostTransaction)
	}

	port := getEnv("PORT", "8082")
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
