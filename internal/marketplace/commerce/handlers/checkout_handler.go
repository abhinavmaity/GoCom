package handlers

import (
	"github.com/gin-gonic/gin"
	"gocom/main/internal/marketplace/commerce/services"
	"gorm.io/gorm"
	"net/http"
)

func Checkout(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cartID := c.Param("id")
		addressID := c.Query("address_id")

		// Step 1: Get the cart and validate it
		service := services.NewCheckoutService(db)
		orderDetails, err := service.ValidateCart(cartID, addressID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Step 2: Process payment and order creation
		// Here Razorpay integration happens
		paymentIntent, err := service.CreatePaymentIntent(orderDetails)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"payment_intent": paymentIntent})
	}
}
