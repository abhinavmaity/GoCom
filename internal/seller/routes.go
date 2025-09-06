package seller

import (
	"github.com/gin-gonic/gin"
	"gocom/main/internal/seller/handlers"
)

func SetupRoutes(r *gin.Engine) {
	// Initialize all handlers
	sellerHandler := handlers.NewSellerHandler()       // Phase 2
	

	kycHandler := handlers.NewKYCHandler()             // Phase 3  
	

	productHandler := handlers.NewProductHandler()     // Phase 4
	skuHandler := handlers.NewSKUHandler()             // Phase 5
	inventoryHandler := handlers.NewInventoryHandler() // Phase 6
	addressHandler := handlers.NewAddressHandler()     // Phase 7
	// orderHandler := handlers.NewOrderHandler()      // Phase 8 - TODO

	// API v1 group
	v1 := r.Group("/v1")

	// Phase 2: Seller Core routes
	{
		v1.POST("/sellers", sellerHandler.CreateSeller)
		v1.GET("/sellers/:id", sellerHandler.GetSeller)
		v1.PATCH("/sellers/:id", sellerHandler.UpdateSeller)
	}

	// Phase 3: KYC Management routes
	{
		v1.POST("/sellers/:id/kyc", kycHandler.UploadKYC)
		v1.GET("/sellers/:id/kyc", kycHandler.GetKYCDocuments)
		v1.GET("/sellers/:id/kyc/:docId", kycHandler.GetKYCDocument)
		v1.DELETE("/sellers/:id/kyc/:docId", kycHandler.DeleteKYC)
	}

	// Phase 4: Product Catalog routes
	{
		v1.POST("/sellers/:id/products", productHandler.CreateProduct)
		v1.GET("/sellers/:id/products", productHandler.ListProducts)
		v1.GET("/products/:id", productHandler.GetProduct)
		v1.PATCH("/products/:id", productHandler.UpdateProduct)
		v1.DELETE("/products/:id", productHandler.DeleteProduct)
		v1.POST("/products/:id/publish", productHandler.PublishProduct)
	}

	// Phase 5: SKU & Variants routes
	{
		v1.POST("/products/:id/skus", skuHandler.CreateSKU)
		v1.GET("/products/:id/skus", skuHandler.GetProductSKUs)
		v1.PATCH("/skus/:id", skuHandler.UpdateSKU)
		v1.DELETE("/skus/:id", skuHandler.DeleteSKU)
	}

	// Phase 6: Inventory Management routes
	{
		v1.GET("/skus/:id/inventory", inventoryHandler.GetInventory)
		v1.PATCH("/skus/:id/inventory", inventoryHandler.UpdateInventory)
		v1.GET("/sellers/:id/inventory/alerts", inventoryHandler.GetLowStockAlerts)
		v1.POST("/inventory/bulk-update", inventoryHandler.BulkUpdateInventory)
	}

	// Phase 7: Address Management routes
	{
		v1.POST("/sellers/:id/addresses", addressHandler.AddAddress)
		v1.GET("/sellers/:id/addresses", addressHandler.GetSellerAddresses)
		v1.PATCH("/addresses/:id", addressHandler.UpdateAddress)
		v1.DELETE("/addresses/:id", addressHandler.DeleteAddress)
	}
}

