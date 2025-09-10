package handlers

import (
	"github.com/gin-gonic/gin"
	"gocom/main/internal/marketplace/commerce/services"
	"gorm.io/gorm"
	"net/http"
)

func AddCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		cartID := c.Param("id")

		var req struct {
			SKUID uint `json:"sku_id"`
			Qty   int  `json:"qty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		service := services.NewCartService(db)
		if err := service.AddItem(userID, cartID, req.SKUID, req.Qty); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Item added to cart"})
	}
}
