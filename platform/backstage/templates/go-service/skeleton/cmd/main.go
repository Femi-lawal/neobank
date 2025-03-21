package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "${{ values.port }}"
	}

	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"service": "${{ values.name }}",
				"version": "1.0.0",
			})
		})
	}

	log.Printf("Starting ${{ values.name }} on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
