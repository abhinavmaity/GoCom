package seller

import (
	"github.com/gin-gonic/gin"
	"gocom/main/internal/common/auth"
	"gocom/main/internal/seller/handlers"
)

func SetupRoutes(r *gin.Engine) {
	// Initialize all handlers
	sellerHandler := handlers.NewSellerHandler()
	kycHandler := handlers.NewKYCHandler()
	productHandler := handlers.NewProductHandler()
	skuHandler := handlers.NewSKUHandler()
	inventoryHandler := handlers.NewInventoryHandler()
	addressHandler := handlers.NewAddressHandler()
	orderHandler := handlers.NewOrderHandler() // ✅ FIXED: Uncommented!

	// API v1 group
	v1 := r.Group("/v1")

	// ✅ Public seller routes (no auth needed for registration)
	v1.POST("/sellers", sellerHandler.CreateSeller)

	// ✅ Protected routes (require JWT)
	protected := v1.Group("")
	protected.Use(auth.JWTAuthMiddleware())
	{
		// Seller Core routes
		protected.GET("/sellers/:id", sellerHandler.GetSeller)
		protected.PATCH("/sellers/:id", sellerHandler.UpdateSeller)

		// KYC Management routes
		protected.POST("/sellers/:id/kyc", kycHandler.UploadKYC)
		protected.GET("/sellers/:id/kyc", kycHandler.GetKYCDocuments)
		protected.GET("/sellers/:id/kyc/:docId", kycHandler.GetKYCDocument)
		protected.DELETE("/sellers/:id/kyc/:docId", kycHandler.DeleteKYC)

		// Product Catalog routes
		protected.POST("/sellers/:id/products", productHandler.CreateProduct)
		protected.GET("/sellers/:id/products", productHandler.ListProducts)
		protected.GET("/products/:id", productHandler.GetProduct)
		protected.PATCH("/products/:id", productHandler.UpdateProduct)
		protected.DELETE("/products/:id", productHandler.DeleteProduct)
		protected.POST("/products/:id/publish", productHandler.PublishProduct)

		// SKU & Variants routes
		protected.POST("/products/:id/skus", skuHandler.CreateSKU)
		protected.GET("/products/:id/skus", skuHandler.GetProductSKUs)
		protected.PATCH("/skus/:id", skuHandler.UpdateSKU)
		protected.DELETE("/skus/:id", skuHandler.DeleteSKU)

		// Inventory Management routes
		protected.GET("/skus/:id/inventory", inventoryHandler.GetInventory)
		protected.PATCH("/skus/:id/inventory", inventoryHandler.UpdateInventory)
		protected.GET("/sellers/:id/inventory/alerts", inventoryHandler.GetLowStockAlerts)
		protected.POST("/inventory/bulk-update", inventoryHandler.BulkUpdateInventory)

		// Address Management routes
		protected.POST("/sellers/:id/addresses", addressHandler.AddAddress)
		protected.GET("/sellers/:id/addresses", addressHandler.GetSellerAddresses)
		protected.PATCH("/addresses/:id", addressHandler.UpdateAddress)
		protected.DELETE("/addresses/:id", addressHandler.DeleteAddress)

		// ✅ FIXED: Order Management routes (uncommented and working)
		protected.GET("/sellers/:id/orders", orderHandler.GetSellerOrders)
		protected.GET("/orders/:id", orderHandler.GetOrderDetails)
		protected.PATCH("/orders/:id/status", orderHandler.UpdateOrderStatus)
		protected.POST("/orders/:id/ship", orderHandler.ShipOrder)
	}
}
