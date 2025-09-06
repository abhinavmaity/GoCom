package seller

import (
	"github.com/gin-gonic/gin"
	"gocom/main/internal/seller/handlers"
)

func SetupRoutes(r *gin.Engine) {
	// Initialize handlers
	sellerHandler := handlers.NewSellerHandler()
	productHandler := handlers.NewProductHandler() // Existing

	// API v1 group
	v1 := r.Group("/v1")

	// Seller routes
	{
		v1.POST("/sellers", sellerHandler.CreateSeller)
		v1.GET("/sellers/:id", sellerHandler.GetSeller)
		v1.PATCH("/sellers/:id", sellerHandler.UpdateSeller)
	}

	// Product routes (existing)
	{
		v1.POST("/sellers/:id/products", productHandler.CreateProduct)
		v1.GET("/sellers/:id/products", productHandler.ListProducts)
		v1.GET("/products/:id", productHandler.GetProduct)
		v1.POST("/products/:id/publish", productHandler.PublishProduct)
	}
}

