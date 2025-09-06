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
	config.LoadConfig()

	database, err := db.InitCommerceDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate new models
	err = database.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Seller{},
		&models.SellerUser{},
		&models.KYC{},
		&models.Category{},
		&models.AuditLog{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	authService := auth.NewAuthService(
		config.AppConfig.JWTSecret,
		database,
	)

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
		protected.Use(authHandler.JWTMiddleware()) // This should work now
		{
			protected.POST("/otp/generate", authHandler.GenerateOTP)
			protected.GET("/profile", authHandler.GetProfile)
		}
	}

	// Setup admin routes with proper middleware
	adminGroup := router.Group("/v1/admin")
	adminGroup.Use(authHandler.JWTMiddleware())
	adminGroup.Use(authHandler.RequireRole(string(auth.RoleAdmin)))
	{
		adminGroup.GET("/kyc/pending", admin.GetKYCPending(database))
		adminGroup.POST("/kyc/:id/approve", admin.ApproveKYC(database))
		adminGroup.POST("/kyc/:id/reject", admin.RejectKYC(database))
		adminGroup.POST("/catalog/categories", admin.CreateCategory(database))
		adminGroup.PATCH("/catalog/attributes", admin.UpdateCategoryAttributes(database))
	}

	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "admin-api",
			"port":    "8081",
		})
	})

	// Start admin API server on port 8081
	adminPort := "8082"
	log.Printf("🚀 Admin API Server starting on port %s", adminPort)
	log.Printf("Available endpoints:")
	log.Printf("  - GET /health")
	log.Printf("  - POST /v1/auth/register")
	log.Printf("  - POST /v1/auth/login")
	log.Printf("  - GET /v1/admin/kyc/pending (Admin only)")
	log.Printf("  - POST /v1/admin/kyc/{id}/approve (Admin only)")

	if err := router.Run(":" + adminPort); err != nil {
		log.Fatal("Failed to start admin API server:", err)
	}
}
