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

func CreateCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		service := services.NewCartService(db)
		cart, err := service.CreateCart(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"cart_id": cart.ID})
	}
}

func GetCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		cartID := c.Param("id")

		service := services.NewCartService(db)
		cart, err := service.GetCart(cartID, userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"cart": cart})
	}
}

func UpdateCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		cartID := c.Param("id")
		itemID := c.Param("item_id")

		var req struct {
			Qty int `json:"qty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		service := services.NewCartService(db)
		if err := service.UpdateItem(userID, cartID, itemID, req.Qty); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cart item updated"})
	}
}

func RemoveCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		cartID := c.Param("id")
		itemID := c.Param("item_id")

		service := services.NewCartService(db)
		if err := service.RemoveItem(userID, cartID, itemID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cart item removed"})
	}
}
