package seller

import (
	"github.com/gin-gonic/gin"
	"gocom/main/internal/seller/handlers"
)

func SetupRoutes(r *gin.Engine) {
	// Initialize handlers
	sellerHandler := handlers.NewSellerHandler()
	productHandler := handlers.NewProductHandler()
	skuHandler := handlers.NewSKUHandler()
	inventoryHandler := handlers.NewInventoryHandler()

	// API v1 group
	v1 := r.Group("/v1")

	// Seller routes
	{
		v1.POST("/sellers", sellerHandler.CreateSeller)
		v1.GET("/sellers/:id", sellerHandler.GetSeller)
		v1.PATCH("/sellers/:id", sellerHandler.UpdateSeller)
	}

	// Product routes
	{
		v1.POST("/sellers/:id/products", productHandler.CreateProduct)
		v1.GET("/sellers/:id/products", productHandler.ListProducts)
		v1.GET("/products/:id", productHandler.GetProduct)
		v1.PATCH("/products/:id", productHandler.UpdateProduct)      // Phase 4
		v1.DELETE("/products/:id", productHandler.DeleteProduct)     // Phase 4
		v1.POST("/products/:id/publish", productHandler.PublishProduct) // Phase 4
	}

	// SKU routes (Phase 5)
	{
		v1.POST("/products/:id/skus", skuHandler.CreateSKU)
		v1.GET("/products/:id/skus", skuHandler.GetProductSKUs)
		v1.PATCH("/skus/:id", skuHandler.UpdateSKU)
		v1.DELETE("/skus/:id", skuHandler.DeleteSKU)
	}

	// Inventory routes (Phase 5)
	{
		v1.GET("/skus/:id/inventory", inventoryHandler.GetInventory)
		v1.PATCH("/skus/:id/inventory", inventoryHandler.UpdateInventory)
		v1.GET("/sellers/:id/inventory/alerts", inventoryHandler.GetLowStockAlerts)
		v1.POST("/inventory/bulk-update", inventoryHandler.BulkUpdateInventory)
	}
}

