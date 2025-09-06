package auth

import (
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine) {
	authHandler := NewAuthHandler()

	v1 := r.Group("/v1")

	// Public auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected auth routes
	protected := v1.Group("/auth")
	protected.Use(JWTAuthMiddleware())
	{
		protected.GET("/me", authHandler.GetProfile)
	}
}
