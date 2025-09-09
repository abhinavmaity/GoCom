package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"gocom/main/internal/common/auth"
	"gocom/main/internal/common/config"
	"gocom/main/internal/common/db"
	"gocom/main/internal/common/errors"
	"gocom/main/internal/integrations/storage" // ✅ Your MinIO setup
	"gocom/main/internal/models"
	"gocom/main/internal/seller"
)

func main() {
	// Load configuration (your existing config works perfectly)
	config.LoadConfig()

	// ✅ Connect to MySQL (using your config)
	db.ConnectMySQL()

	// ✅ FIXED: Auto-migrate ALL models including missing ones
	if err := db.AutoMigrate(
		&models.User{},
		&models.Seller{},
		&models.SellerUser{},
		&models.KYC{},
		&models.Category{},
		&models.Product{},
		&models.SKU{},
		&models.Inventory{},
		&models.Media{},
		&models.Address{},
		// ✅ FIXED: Added missing models
		&models.Order{},
		&models.OrderItem{},
		&models.Payment{},
		&models.Shipment{},
		&models.Return{},
		&models.Refund{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// ✅ Connect MinIO (using your existing setup)
	storage.ConnectMinIO()
	if err := storage.InitializeBuckets(); err != nil {
		log.Printf("MinIO bucket initialization warning: %v", err)
		// Don't fail startup for MinIO issues
	}

	// Setup Gin
	gin.SetMode(config.AppConfig.GinMode)
	r := gin.Default()

	// CORS middleware
	r.Use(corsMiddleware())
	r.Use(errors.ErrorHandler())

	// Setup routes
	auth.SetupAuthRoutes(r)
	seller.SetupRoutes(r)

	// Health check
	r.GET("/health", healthCheck)

	// ✅ Start server
	log.Printf("🚀 Seller API starting on port %s", config.AppConfig.ServerPort)
	log.Printf("🔐 Auth: http://localhost:%s/v1/auth/*", config.AppConfig.ServerPort)
	log.Printf("🏪 Seller: http://localhost:%s/v1/sellers/*", config.AppConfig.ServerPort)
	log.Printf("💊 Health: http://localhost:%s/health", config.AppConfig.ServerPort)
	
	log.Fatal(r.Run(":" + config.AppConfig.ServerPort))
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "ok",
		"service":   "seller-api",
		"database":  "connected",
		"auth":      "enabled",
		"storage":   "minio",
		"version":   "1.0.0",
	})
}
