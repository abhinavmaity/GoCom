package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"gocom/main/internal/common/config"
	"gocom/main/internal/common/db"
	"gocom/main/internal/common/errors"
	"gocom/main/internal/integrations/storage"
	"gocom/main/internal/models"
	"gocom/main/internal/seller"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Connect to services
	db.ConnectMySQL()
	
	/*
	TODO: Connect redis
	db.ConnectRedis()
	*/

	// Auto-migrate database schemas
	if err := db.GetDB().AutoMigrate(
		&models.User{},
		&models.Seller{},
		&models.Category{},
		&models.Product{},
		&models.SKU{},
		&models.Media{},
		&models.Inventory{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	storage.ConnectMinIO()
	if err := storage.InitializeBuckets(); err != nil {
		log.Fatalf("Failed to initialize buckets: %v", err)
	} else {
		log.Printf("Initialized Buckets!")
	}

	// Setup Gin
	gin.SetMode(config.AppConfig.GinMode)
	r := gin.Default()

	// Add middleware
	r.Use(errors.ErrorHandler())

	// Setup routes
	seller.SetupRoutes(r)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "seller-api"})
	})

	// Start server
	log.Printf("🚀 Seller API server starting on port %s", config.AppConfig.ServerPort)
	log.Fatal(r.Run(":" + config.AppConfig.ServerPort))
}
