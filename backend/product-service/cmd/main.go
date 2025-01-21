package main

import (
	"log/slog"
	"os"

	"github.com/femi-lawal/new_bank/backend/product-service/internal/handler"
	"github.com/femi-lawal/new_bank/backend/product-service/internal/model"
	"github.com/femi-lawal/new_bank/backend/product-service/internal/repository"
	"github.com/femi-lawal/new_bank/backend/product-service/internal/service"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/db"
	apperrors "github.com/femi-lawal/new_bank/backend/shared-lib/pkg/errors"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/logger"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/metrics"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/middleware"
	"github.com/gin-gonic/gin"
)

const serviceName = "product-service"

func main() {
	// Initialize Logger
	logger.InitLogger(serviceName, true)
	slog.Info("Starting Product Service")

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
	if err := database.AutoMigrate(&model.Product{}); err != nil {
		slog.Error("Failed to migrate database", "error", err)
	}

	// Wiring
	repo := repository.NewProductRepository(database)
	svc := service.NewProductService(repo)
	h := handler.NewProductHandler(svc)

	// Get JWT secret
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
		})
	})

	// Products can be viewed without auth, but apply/create requires auth
	r.GET("/api/v1/products", h.ListProducts)

	// ============================================
	// Protected endpoints
	// ============================================
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(jwtSecret))
	{
		api.POST("/products", h.CreateProduct)
	}

	port := getEnv("PORT", "8084")
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
