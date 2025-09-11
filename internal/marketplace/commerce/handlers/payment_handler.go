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

func CreatePaymentIntent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OrderID uint    `json:"order_id"`
			Amount  float64 `json:"amount"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		service := services.NewPaymentService(db)
		intentID, err := service.CreatePaymentIntent(req.OrderID, req.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"intent_id": intentID})
	}
}
