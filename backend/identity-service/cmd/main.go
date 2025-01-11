package main

import (
	"log/slog"
	"os"

	"github.com/femi-lawal/new_bank/backend/identity-service/internal/handler"
	"github.com/femi-lawal/new_bank/backend/identity-service/internal/model"
	"github.com/femi-lawal/new_bank/backend/identity-service/internal/repository"
	"github.com/femi-lawal/new_bank/backend/identity-service/internal/service"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/db"
	apperrors "github.com/femi-lawal/new_bank/backend/shared-lib/pkg/errors"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/logger"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/metrics"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/middleware"
	"github.com/gin-gonic/gin"
)

const serviceName = "identity-service"

func main() {
	// Initialize Logger
	logger.InitLogger(serviceName, true)
	slog.Info("Starting Identity Service")

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
	if err := database.AutoMigrate(&model.User{}); err != nil {
		slog.Error("Failed to migrate database", "error", err)
	}

	// Wiring
	userRepo := repository.NewUserRepository(database)
	jwtSecret := getEnv("JWT_SECRET", "my-secret-key")
	authService := service.NewAuthService(userRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(authService)

	// Setup Router
	r := gin.Default()

	// ============================================
	// Global Middleware (applied to ALL routes)
	// ============================================
	r.Use(apperrors.ErrorMiddleware())               // Panic recovery with structured errors
	r.Use(middleware.RequestLogger(serviceName))     // Request logging with request ID
	r.Use(middleware.CORS())                         // CORS handling
	r.Use(middleware.RateLimit())                    // Rate limiting
	r.Use(metrics.PrometheusMiddleware(serviceName)) // Prometheus metrics

	// ============================================
	// Public endpoints (no auth required)
	// ============================================
	r.GET("/metrics", metrics.MetricsHandler())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": serviceName,
		})
	})

	// Auth endpoints (public - for login/register)
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// ============================================
	// Protected endpoints (auth required)
	// ============================================
	protected := r.Group("/api/v1")
	protected.Use(middleware.JWTAuth(jwtSecret))
	{
		// User profile endpoints
		protected.GET("/me", func(c *gin.Context) {
			userID := middleware.GetUserID(c)
			email := middleware.GetEmail(c)
			c.JSON(200, gin.H{
				"user_id": userID,
				"email":   email,
			})
		})
	}

	port := getEnv("PORT", "8081")
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
