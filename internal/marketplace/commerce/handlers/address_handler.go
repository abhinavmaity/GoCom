package handlers

import (
	"github.com/gin-gonic/gin"
	"gocom/main/internal/marketplace/commerce/services"
	"gocom/main/internal/models"
	"gorm.io/gorm"
	"net/http"
)

func CreateAddress(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.Address
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		userID := c.GetUint("user_id") // Get user_id from JWT

		service := services.NewAddressService(db)
		address, err := service.CreateAddress(userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"address_id": address.ID})
	}
}

func GetAddresses(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id") // Get user_id from JWT

		service := services.NewAddressService(db)
		addresses, err := service.GetAddresses(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"addresses": addresses})
	}
}
