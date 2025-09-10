package handlers

import (
	"github.com/gin-gonic/gin"
	"gocom/main/internal/marketplace/commerce/services"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func ReserveInventory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the skuID from the URL parameters (which is passed as a string)
		skuIDStr := c.Param("sku_id")

		// Convert skuID from string to uint
		skuID, err := strconv.ParseUint(skuIDStr, 10, 32) // Parse to uint
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sku_id"})
			return
		}

		// Get the quantity (qty) from the query parameters
		quantity := c.DefaultQuery("qty", "1")

		// Convert quantity from string to int
		qty, err := strconv.Atoi(quantity)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quantity"})
			return
		}

		// Step 1: Call ReserveInventory service
		service := services.NewInventoryService(db)
		if err := service.ReserveInventory(uint(skuID), qty); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Respond with a success message
		c.JSON(http.StatusOK, gin.H{"message": "Inventory reserved successfully"})
	}
}

func CheckInventory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the skuID from the URL parameters (which is passed as a string)
		skuIDStr := c.Param("sku_id")

		// Convert skuID from string to uint
		skuID, err := strconv.ParseUint(skuIDStr, 10, 32) // Parse to uint
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sku_id"})
			return
		}

		// Get the quantity (qty) from the query parameters
		quantity := c.DefaultQuery("qty", "1")

		// Convert quantity from string to int
		qty, err := strconv.Atoi(quantity)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quantity"})
			return
		}

		// Step 1: Call CheckInventory service
		service := services.NewInventoryService(db)
		available, err := service.CheckInventory(uint(skuID), qty)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Step 2: Return whether inventory is sufficient or not
		if available {
			c.JSON(http.StatusOK, gin.H{"message": "Sufficient inventory available"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "Insufficient inventory"})
		}
	}
}

func ReleaseInventory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the skuID from the URL parameters (which is passed as a string)
		skuIDStr := c.Param("sku_id")

		// Convert skuID from string to uint
		skuID, err := strconv.ParseUint(skuIDStr, 10, 32) // Parse to uint
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sku_id"})
			return
		}

		// Get the quantity (qty) from the query parameters
		quantity := c.DefaultQuery("qty", "1")

		// Convert quantity from string to int
		qty, err := strconv.Atoi(quantity)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quantity"})
			return
		}

		// Step 1: Call ReleaseInventory service
		service := services.NewInventoryService(db)
		if err := service.ReleaseInventory(uint(skuID), qty); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Respond with a success message
		c.JSON(http.StatusOK, gin.H{"message": "Inventory released successfully"})
	}
}
