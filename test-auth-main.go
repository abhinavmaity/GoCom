// // cmd/test-auth/main.go
// package main

// import (
// 	"gocom/main/internal/common/auth"
// 	"gocom/main/internal/common/config"
// 	"gocom/main/internal/common/db"
// 	"log"

// 	"github.com/gin-gonic/gin"
// )

// func main() {
// 	// Load your config
// 	config.LoadConfig()

// 	// Connect to database
// 	// database, err := db.InitCommerceDB()
// 	// if err != nil {
// 	// 	log.Fatal("Failed to connect to database:", err)
// 	// }

// 	// Initialize auth service
// 	database := db.GetDB()
// 	authService := auth.NewAuthService()
// 	authHandler := auth.NewAuthHandler(authService, database)

// 	// Setup test routes
// 	router := gin.Default()

// 	// Public routes
// 	router.POST("/register", authHandler.Register)
// 	router.POST("/login", authHandler.Login)

// 	// Protected routes
// 	protected := router.Group("/protected")
// 	protected.Use(authService.JWTMiddleware())
// 	protected.GET("/profile", authHandler.GetProfile)

// 	// Admin only route
// 	admin := router.Group("/admin")
// 	admin.Use(authService.JWTMiddleware())
// 	admin.Use(authService.RequireRole(auth.RoleAdmin))
// 	admin.GET("/dashboard", func(c *gin.Context) {
// 		c.JSON(200, gin.H{"message": "Admin dashboard"})
// 	})

// 	log.Println("🚀 Test auth server starting on :8080")
// 	router.Run(":8080")
// }

package main

import (
	"gocom/main/internal/common/auth"
	"gocom/main/internal/common/config"
	"gocom/main/internal/common/db"
	"gocom/main/internal/models"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load your config
	config.LoadConfig()

	// Connect to database
	database, err := db.InitCommerceDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models to ensure database schema is up to date
	err = database.AutoMigrate(
		&models.User{},
		&models.RefreshToken{}, // Add refresh token support
		&models.Seller{},
		&models.SellerUser{},
		&models.KYC{},
		// Add other models as needed
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize auth service with JWT secret
	authService := auth.NewAuthService(
		config.AppConfig.JWTSecret,
		database,
	)

	// Initialize auth handler with both service and database
	authHandler := auth.NewAuthHandler(authService, database)

	// Setup test routes
	router := gin.Default()

	// Add CORS middleware for testing
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Public auth routes
	authGroup := router.Group("/v1/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/otp", authHandler.VerifyOTP)
	}

	// Protected routes requiring authentication
	protected := router.Group("/protected")
	protected.Use(authService.JWTMiddleware())
	{
		protected.GET("/profile", authHandler.GetProfile)
		protected.POST("/otp/generate", authHandler.GenerateOTP)
		protected.GET("/user-info", func(c *gin.Context) {
			userID, exists := auth.GetUserIDFromContext(c)
			if !exists {
				c.JSON(401, gin.H{"error": "User not found in context"})
				return
			}

			var user models.User
			if err := database.First(&user, userID).Error; err != nil {
				c.JSON(404, gin.H{"error": "User not found"})
				return
			}

			c.JSON(200, gin.H{
				"user": gin.H{
					"id":          user.ID,
					"name":        user.Name,
					"email":       user.Email,
					"phone":       user.Phone,
					"status":      user.Status,
					"otp_enabled": user.OTPEnabled,
				},
			})
		})
	}

	// Admin only routes
	adminGroup := router.Group("/admin")
	adminGroup.Use(authService.JWTMiddleware())
	adminGroup.Use(authService.RequireRole(auth.RoleAdmin))
	{
		adminGroup.GET("/dashboard", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message":     "Welcome to Admin Dashboard",
				"user":        c.MustGet("user"),
				"permissions": c.MustGet("permissions"),
			})
		})

		adminGroup.GET("/users", func(c *gin.Context) {
			var users []models.User
			database.Find(&users)
			c.JSON(200, gin.H{
				"users": users,
				"count": len(users),
			})
		})
	}

	// Seller routes
	sellerGroup := router.Group("/seller")
	sellerGroup.Use(authService.JWTMiddleware())
	sellerGroup.Use(authService.RequireRole(auth.RoleSeller))
	{
		sellerGroup.GET("/dashboard", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Welcome to Seller Dashboard",
				"user":    c.MustGet("user"),
			})
		})
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "test-auth",
			"endpoints": []string{
				"POST /v1/auth/register",
				"POST /v1/auth/login",
				"POST /v1/auth/refresh",
				"POST /v1/auth/otp",
				"GET /protected/profile",
				"POST /protected/otp/generate",
				"GET /admin/dashboard",
				"GET /admin/users",
				"GET /seller/dashboard",
			},
		})
	})

	// Test endpoint to check if server is running
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":   "pong",
			"timestamp": "2025-01-01T00:00:00Z", // You can use time.Now() here
		})
	})

	log.Println("🚀 Test auth server starting on :8080")
	log.Println("Available endpoints:")
	log.Println("  - POST /v1/auth/register - Register new user")
	log.Println("  - POST /v1/auth/login - Login user")
	log.Println("  - POST /v1/auth/refresh - Refresh JWT token")
	log.Println("  - POST /v1/auth/otp - Verify OTP")
	log.Println("  - GET /protected/profile - Get user profile (protected)")
	log.Println("  - POST /protected/otp/generate - Generate OTP (protected)")
	log.Println("  - GET /admin/dashboard - Admin dashboard (admin only)")
	log.Println("  - GET /admin/users - List all users (admin only)")
	log.Println("  - GET /seller/dashboard - Seller dashboard (seller only)")
	log.Println("  - GET /health - Health check")
	log.Println("  - GET /ping - Server status")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
