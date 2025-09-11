package routes

import (
	"github.com/gin-gonic/gin"
	"gocom/main/internal/marketplace/commerce/handlers"
	"gorm.io/gorm"
)

func SetupRoutes(v1 *gin.RouterGroup, db *gorm.DB) {
	// Cart management
	v1.POST("/carts", handlers.CreateCart(db))
	v1.GET("/carts/:id", handlers.GetCart(db))
	v1.POST("/carts/:id/items", handlers.AddCartItem(db))
	v1.PATCH("/carts/:id/items/:item_id", handlers.UpdateCartItem(db))
	v1.DELETE("/carts/:id/items/:item_id", handlers.RemoveCartItem(db))

	// Address management
	v1.POST("/addresses", handlers.CreateAddress(db))
	v1.GET("/addresses", handlers.GetAddresses(db))
	v1.PATCH("/addresses/:id", handlers.UpdateAddress(db))
	v1.DELETE("/addresses/:id", handlers.DeleteAddress(db))

	// Checkout process
	v1.POST("/checkout", handlers.Checkout(db))

	// Order management
	v1.POST("/orders", handlers.PlaceOrder(db))
	v1.GET("/orders", handlers.GetOrders(db))
	v1.GET("/orders/:id", handlers.GetOrder(db))

	// Payment handling
	v1.POST("/payments/intents", handlers.CreatePaymentIntent(db))
	v1.POST("/payments/capture/:intent_id", handlers.CapturePayment(db))

	// Inventory operations
	v1.POST("/inventory/:sku_id/reserve", handlers.ReserveInventory(db))
	v1.POST("/inventory/:sku_id/release", handlers.ReleaseInventory(db))
	v1.GET("/inventory/:sku_id/check", handlers.CheckInventory(db))
}
