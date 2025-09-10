package handlers

import (
	"github.com/gin-gonic/gin"
	"gocom/main/internal/marketplace/commerce/services"
	"gorm.io/gorm"
	"net/http"
)

func CapturePayment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the payment intent ID from the request
		intentID := c.Param("intent_id")

		// Step 1: Capture the payment via the Payment Service
		service := services.NewPaymentService(db)
		order, err := service.CapturePayment(intentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Step 2: Return order details (payment confirmed)
		c.JSON(http.StatusOK, gin.H{
			"order_id": order.ID,
			"status":   "confirmed",
		})
	}
}
