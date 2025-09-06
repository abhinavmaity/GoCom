package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gocom/main/internal/seller/services"
)

type InventoryHandler struct {
	InventoryService *services.InventoryService
}

func NewInventoryHandler() *InventoryHandler {
	return &InventoryHandler{
		InventoryService: services.NewInventoryService(),
	}
}

// Get inventory for SKU
// GET /skus/:id/inventory
func (ih *InventoryHandler) GetInventory(c *gin.Context) {
	skuID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SKU ID"})
		return
	}

	inventory, err := ih.InventoryService.GetInventory(uint(skuID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    inventory,
	})
}

// Update inventory
// PATCH /skus/:id/inventory
func (ih *InventoryHandler) UpdateInventory(c *gin.Context) {
	skuID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SKU ID"})
		return
	}

	var req services.UpdateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	inventory, err := ih.InventoryService.UpdateInventory(uint(skuID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    inventory,
		"message": "Inventory updated successfully",
	})
}

// Get low stock alerts for seller
// GET /sellers/:id/inventory/alerts
func (ih *InventoryHandler) GetLowStockAlerts(c *gin.Context) {
	sellerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid seller ID"})
		return
	}

	alerts, err := ih.InventoryService.GetLowStockAlerts(uint(sellerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    alerts,
		"count":   len(alerts),
	})
}

// Bulk inventory update
// POST /inventory/bulk-update
func (ih *InventoryHandler) BulkUpdateInventory(c *gin.Context) {
	var updates []services.BulkInventoryUpdate
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	if err := ih.InventoryService.BulkUpdateInventory(updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Inventory updated successfully",
	})
}

