package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gocom/main/internal/common/config"
	"gocom/main/internal/common/db"
	commerce "gocom/main/internal/marketplace/commerce"
	discovery "gocom/main/internal/marketplace/discovery/routes"
	"gocom/main/internal/models"
)

func main() {
	// 1) Load configuration into AppConfig
	config.LoadConfig()

	// 2) Connect DB (ConnectMySQL uses config.GetDatabaseDSN())
	db.ConnectMySQL()

	// 3) Get *gorm.DB
	gormDB := db.GetDB()

	// 4) Run migrations (includes all models for both Dev A & B)
	if err := runAutoMigrate(gormDB); err != nil {
		log.Fatalf("auto-migrate failed: %v", err)
	}

	// 5) Set up Gin router
	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Health check endpoint
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "marketplace-api"})
	})

	// API v1 group
	v1 := r.Group("/v1")

	// Optional: Add global middleware for v1 routes
	// v1.Use(middleware.CORS())
	// v1.Use(middleware.JWTAuth())

	// Developer A routes (Discovery & Product Experience)
	discovery.SetupRoutes(v1, gormDB)

	// Developer B routes (Commerce & Transaction Flow)
	commerce.SetupRoutes(v1, gormDB)

	// 6) Start server
	addr := ":" + getenv("MARKETPLACE_PORT", "8082")
	srvErr := make(chan error, 1)

	go func() {
		log.Printf("🚀 marketplace-api starting on %s", addr)
		log.Println("📋 Available endpoints:")
		log.Println("   Health: GET /healthz")
		log.Println("   Discovery: GET /v1/products, /v1/categories, /v1/products/{id}/reviews")
		log.Println("   Commerce: POST /v1/carts, /v1/orders, /v1/payments, /v1/addresses")

		if err := r.Run(addr); err != nil {
			srvErr <- err
		}
	}()

	// 7) Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("🛑 Shutting down server (signal: %v)...", sig)
		cleanup()
		log.Println("✅ Server shutdown complete")
	case err := <-srvErr:
		log.Fatalf("❌ Server error: %v", err)
	}
}

// runAutoMigrate runs database migrations for all models (Dev A & B)
func runAutoMigrate(gormDB *gorm.DB) error {
	log.Println("🔄 Running database migrations...")

	err := db.AutoMigrate(
		// Core models (shared)
		&models.User{},
		&models.Seller{},
		&models.Category{},
		&models.Media{},

		// Developer A models (Discovery)
		&models.Product{},
		&models.SKU{},
		&models.Inventory{},
		&models.Review{},

		// Developer B models (Commerce)
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.Payment{},
		&models.Address{},

		// Additional models (if needed)
		// &models.Coupon{},
		// &models.Shipment{},
		// &models.Return{},
		// &models.Refund{},
		// &models.AuditLog{},
	)

	if err != nil {
		return err
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}

// cleanup handles graceful shutdown tasks
func cleanup() {
	log.Println("🧹 Starting cleanup...")

	if db.GetDB() == nil {
		return
	}

	sqlDB, err := db.GetDB().DB()
	if err != nil {
		log.Printf("❌ Error getting sql.DB to close: %v", err)
		return
	}

	done := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ Panic during database close: %v", r)
			}
		}()

		if err := sqlDB.Close(); err != nil {
			log.Printf("❌ Error closing database: %v", err)
		} else {
			log.Println("✅ Database connection closed")
		}
		close(done)
	}()

	select {
	case <-done:
		log.Println("✅ Cleanup completed")
	case <-time.After(5 * time.Second):
		log.Println("⚠️  Cleanup timeout - forcing shutdown")
	}
}

// getenv gets environment variable with fallback default
func getenv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Optional: Add development mode check
func isDevelopmentMode() bool {
	return getenv("GIN_MODE", "debug") != "release"
}

// Optional: Add configuration validation
func validateConfig() error {
	required := []string{
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_NAME",
	}

	for _, env := range required {
		if os.Getenv(env) == "" {
			log.Printf("⚠️  Warning: %s environment variable not set", env)
		}
	}

	return nil
}
