package main

import (
    "log"

    "github.com/gin-gonic/gin"

    "gocom/main/internal/common/auth"
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

    // Auto-migrate database schemas
    if err := db.GetDB().AutoMigrate(
        &models.User{},
        &models.Seller{},
        &models.SellerUser{},
        &models.KYC{},
        &models.SKU{},
        &models.Inventory{},
        &models.Media{},
        &models.Category{},
        &models.Product{},
        &models.Address{},
    ); err != nil {
        log.Fatal("Failed to migrate database:", err)
    }

    storage.ConnectMinIO()
    if err := storage.InitializeBuckets(); err != nil {
        log.Fatalf("Failed to initialize buckets: %v", err)
    }

    // Setup Gin
    gin.SetMode(config.AppConfig.GinMode)
    r := gin.Default()

    // CORS middleware
    r.Use(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    })

    // Add middleware
    r.Use(errors.ErrorHandler())

    // Setup authentication routes
    auth.SetupAuthRoutes(r)

    // Setup seller routes
    seller.SetupRoutes(r)

    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status":    "ok",
            "service":   "seller-api",
            "database":  "connected",
            "auth":      "enabled",
        })
    })

    // Start server
    log.Printf("🚀 Server starting on port %s", config.AppConfig.ServerPort)
    log.Printf("🔐 Auth: http://localhost:%s/v1/auth/*", config.AppConfig.ServerPort)
    log.Printf("💊 Health: http://localhost:%s/health", config.AppConfig.ServerPort)
    
    log.Fatal(r.Run(":" + config.AppConfig.ServerPort))
}
