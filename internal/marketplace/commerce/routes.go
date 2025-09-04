package commerce

import (
	"github.com/gin-gonic/gin"
	"gocom/main/internal/marketplace/commerce/handlers"
	"gocom/main/internal/marketplace/middleware"
	"gorm.io/gorm"
)

func SetupRoutes(v1 *gin.RouterGroup, db *gorm.DB) {
	// Protected routes (require authentication)
	auth := v1.Use(middleware.AuthRequired())

	// Cart routes
	auth.POST("/carts", handlers.CreateCart(db))
	auth.GET("/carts/:id", handlers.GetCart(db))
	auth.POST("/carts/:id/items", handlers.AddCartItem(db))
	auth.PATCH("/carts/:id/items/:itemId", handlers.UpdateCartItem(db))
	auth.DELETE("/carts/:id/items/:itemId", handlers.RemoveCartItem(db))

	// Checkout & Orders
	auth.POST("/checkout", handlers.Checkout(db))
	auth.POST("/orders", handlers.CreateOrder(db))
	auth.GET("/orders", handlers.GetUserOrders(db))
	auth.GET("/orders/:id", handlers.GetOrder(db))

	// Payments
	auth.POST("/payments/intents", handlers.CreatePaymentIntent(db))
	auth.POST("/payments/capture", handlers.CapturePayment(db))

	// Addresses
	auth.POST("/addresses", handlers.CreateAddress(db))
	auth.GET("/addresses", handlers.GetUserAddresses(db))
}
