package seller

import (
	"github.com/gin-gonic/gin"

	"gocom/main/internal/seller/handlers"
)

func SetupRoutes(r *gin.Engine) {
	// Initialize handlers
	productHandler := handlers.NewProductHandler()

	// API v1 group
	v1 := r.Group("/v1")

	// Product routes
	{
		// Seller-specific product routes
		v1.POST("/sellers/:id/products", productHandler.CreateProduct)
		v1.GET("/sellers/:id/products", productHandler.ListProducts)

		// Product management routes
		v1.GET("/products/:id", productHandler.GetProduct)
		v1.POST("/products/:id/publish", productHandler.PublishProduct)
	}
}
