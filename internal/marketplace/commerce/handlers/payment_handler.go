// internal/marketplace/commerce/handlers/payment_handler.go
package handlers

import (
	"github.com/gin-gonic/gin"
	"gocom/main/internal/marketplace/commerce/dto"
	"gocom/main/internal/marketplace/commerce/services"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func CreatePaymentIntent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		var req dto.PaymentIntentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		service := services.NewPaymentService(db)
		response, err := service.CreatePaymentIntent(req.OrderID, userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"payment_intent": response})
	}
}

func CapturePayment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CapturePaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		service := services.NewPaymentService(db)
		err := service.CapturePayment(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Payment captured successfully"})
	}
}
