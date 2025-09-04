// internal/marketplace/commerce/handlers/cart_handler.go
package handlers

import (
	"github.com/gin-gonic/gin"
	"gocom/main/internal/marketplace/commerce/dto"
	"gocom/main/internal/marketplace/commerce/services"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func CreateCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		var req dto.CreateCartRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		service := services.NewCartService(db)
		cart, err := service.CreateCart(userID, req.Currency)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Cart created successfully",
			"cart_id": cart.ID,
		})
	}
}

func GetCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		cartID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

		service := services.NewCartService(db)
		cart, err := service.GetCart(uint(cartID), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"cart": cart})
	}
}

func AddCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		cartID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

		var req dto.AddCartItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		service := services.NewCartService(db)
		err := service.AddItem(uint(cartID), userID, req.SKUID, req.Qty)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Item added to cart"})
	}
}

func UpdateCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		cartID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
		itemID, _ := strconv.ParseUint(c.Param("itemId"), 10, 32)

		var req dto.UpdateCartItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		service := services.NewCartService(db)
		err := service.UpdateItemQty(uint(cartID), userID, uint(itemID), req.Qty)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Item updated"})
	}
}

func RemoveCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		cartID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
		itemID, _ := strconv.ParseUint(c.Param("itemId"), 10, 32)

		service := services.NewCartService(db)
		err := service.RemoveItem(uint(cartID), userID, uint(itemID))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart"})
	}
}
