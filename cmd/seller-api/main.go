package main

import (
	"fmt"
	"log"
	"time"
	"gorm.io/gorm"
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
	log.Println("Loading configuration...")
	config.LoadConfig()

	log.Println("Connecting to MySQL database...")
	db.ConnectMySQL()
	if err := initializeDatabase(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	log.Println("Connecting to MinIO storage...")
	storage.ConnectMinIO()
	if err := storage.InitializeBuckets(); err != nil {
		log.Fatalf("Failed to initialize storage buckets: %v", err)
	}

	setupServer()
}

func initializeDatabase() error {
	log.Println("Starting database schema migration...")
	
	database := db.GetDB()
	if database == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}
	log.Println("Database connection verified")

	if err := database.AutoMigrate(
		&models.User{},
		&models.Seller{},
		&models.SellerUser{},
		&models.Category{},    
		&models.Product{},
		&models.SKU{},
		&models.Inventory{},   
		&models.Media{},       
		&models.KYC{},        
		&models.Address{},      
		&models.Order{},    
	    &models.OrderItem{},   
	    &models.Shipment{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}


	if err := insertInitialData(database); err != nil {
		log.Printf("Warning: Failed to insert initial data: %v", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

func insertInitialData(database *gorm.DB) error {
	var categoryCount int64
	database.Model(&models.Category{}).Count(&categoryCount)
	
	if categoryCount == 0 {
		categories := []models.Category{
			{Name: "Electronics", SEOSlug: "electronics", IsActive: true},
			{Name: "Fashion", SEOSlug: "fashion", IsActive: true},
			{Name: "Books", SEOSlug: "books", IsActive: true},
			{Name: "Home & Garden", SEOSlug: "home-garden", IsActive: true},
			{Name: "Sports", SEOSlug: "sports", IsActive: true},
		}
		
		for _, category := range categories {
			if err := database.Create(&category).Error; err != nil {
				log.Printf("Failed to create category %s: %v", category.Name, err)
			}
		}
		log.Println("Initial categories created")
	}

	return nil
}

func setupServer() {
	gin.SetMode(config.AppConfig.GinMode)
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type, X-Requested-With")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.Use(errors.ErrorHandler())
	auth.SetupAuthRoutes(r)
	seller.SetupRoutes(r)
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.GetDB().DB()
		dbStatus := "connected"
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "disconnected"
		}

		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "seller-api",
			"version":   "1.0.0",
			"database":  dbStatus,
			"auth":      "enabled",
			"storage":   "minio",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	port := config.AppConfig.ServerPort
	log.Printf("Server starting on port %s", port)
	log.Printf("Auth: http://localhost:%s/v1/auth/*", port)
	log.Printf("Seller: http://localhost:%s/v1/sellers/*", port)
	log.Printf("Health: http://localhost:%s/health", port)
	
	log.Fatal(r.Run(":" + port))
}
