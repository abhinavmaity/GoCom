// package adminapi

// import (
// 	"gocom/main/internal/admin"
// 	"gocom/main/internal/common/auth"
// 	"gocom/main/internal/common/config"
// 	"gocom/main/internal/common/db"
// 	"log"

// 	"github.com/gin-gonic/gin"
// )

// func main() {
// 	// Load configuration using your existing function
// 	config.LoadConfig()

// 	// Initialize database using your existing function
// 	database, err := db.InitCommerceDB()
// 	if err != nil {
// 		log.Fatal("Failed to connect to database:", err)
// 	}

// 	// Initialize auth service
// 	authService := auth.NewAuthService(
// 		config.AppConfig.JWTSecret,
// 		database,
// 	)

// 	// Setup Gin
// 	if config.IsProduction() {
// 		gin.SetMode(gin.ReleaseMode)
// 	}
// 	router := gin.Default()

// 	// Setup admin routes
// 	admin.SetupAdminRoutes(router, database, authService)

// 	// Start admin API server on port 8081
// 	adminPort := "8081"
// 	log.Printf("Admin API Server starting on port %s", adminPort)
// 	if err := router.Run(":" + adminPort); err != nil {
// 		log.Fatal("Failed to start admin API server:", err)
// 	}
// }

package main

import (
	"gocom/main/internal/admin"
	"gocom/main/internal/common/auth"
	"gocom/main/internal/common/config"
	"gocom/main/internal/common/db"
	"gocom/main/internal/models"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration using your existing function
	config.LoadConfig()

	// Initialize database using your existing function
	database, err := db.InitCommerceDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate new models (RefreshToken)
	err = database.AutoMigrate(
		&models.User{},
		&models.RefreshToken{}, // Add this for refresh token support
		&models.Seller{},
		&models.SellerUser{},
		&models.KYC{},
		// Add other models as needed
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize auth service
	authService := auth.NewAuthService(
		config.AppConfig.JWTSecret,
		database,
	)

	// Initialize auth handler with database connection
	authHandler := auth.NewAuthHandler(authService, database)

	// Setup Gin
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Setup authentication routes (shared across all APIs)
	authGroup := router.Group("/v1/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/otp", authHandler.VerifyOTP)

		// Protected auth routes
		protected := authGroup.Group("")
		protected.Use(authHandler.JWTMiddleware())
		{
			protected.POST("/otp/generate", authHandler.GenerateOTP)
			protected.GET("/profile", authHandler.GetProfile)
		}
	}

	// Setup admin routes
	admin.SetupAdminRoutes(router, database, authService)

	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "admin-api",
			"port":    "8081",
		})
	})

	// Start admin API server on port 8081
	adminPort := "8081"
	log.Printf("Admin API Server starting on port %s", adminPort)
	log.Printf("Available endpoints:")
	log.Printf("  - POST /v1/auth/register")
	log.Printf("  - POST /v1/auth/login")
	log.Printf("  - POST /v1/auth/refresh")
	log.Printf("  - POST /v1/auth/otp")
	log.Printf("  - GET /v1/admin/kyc/pending")
	log.Printf("  - POST /v1/admin/kyc/{id}/approve")
	log.Printf("  - POST /v1/admin/kyc/{id}/reject")
	log.Printf("  - GET /health")

	if err := router.Run(":" + adminPort); err != nil {
		log.Fatal("Failed to start admin API server:", err)
	}
}
