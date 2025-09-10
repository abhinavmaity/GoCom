package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"gocom/main/internal/marketplace/commerce/services"
)

func PlaceOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cartID := c.Param("cart_id")
		addressIDStr := c.Query("address_id") // Get address_id as a string

		// Step 1: Convert addressID from string to uint
		addressID, err := strconv.ParseUint(addressIDStr, 10, 32) // Convert to uint
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address_id"})
			return
		}

		// Step 2: Validate Cart and create Order
		service := services.NewOrderService(db)
		order, err := service.CreateOrderFromCart(cartID, uint(addressID)) // Pass the uint addressID
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Step 3: Return order details (including payment status)
		c.JSON(http.StatusOK, gin.H{"order_id": order.ID, "order_status": order.Status})
	}
}
