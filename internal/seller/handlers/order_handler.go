package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gocom/main/internal/seller/services"
)

type OrderHandler struct {
	OrderService *services.OrderService
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{OrderService: services.NewOrderService()}
}

// GET /sellers/:id/orders
func (oh *OrderHandler) GetSellerOrders(c *gin.Context) {
	sellerID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	
	orders, err := oh.OrderService.GetSellerOrders(uint(sellerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": orders})
}

// GET /orders/:id  
func (oh *OrderHandler) GetOrderDetails(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	sellerID := uint(1) // TODO: Get from JWT

	details, err := oh.OrderService.GetOrderDetails(uint(orderID), sellerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": details})
}

// PATCH /orders/:id/status
func (oh *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	sellerID := uint(1) // TODO: Get from JWT

	var req struct {
		Status int `json:"status" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := oh.OrderService.UpdateOrderStatus(uint(orderID), sellerID, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Status updated"})
}

// POST /orders/:id/ship
func (oh *OrderHandler) ShipOrder(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	sellerID := uint(1) // TODO: Get from JWT

	var req struct {
		Provider string `json:"provider" binding:"required"`
		AWB      string `json:"awb" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := oh.OrderService.ShipOrder(uint(orderID), sellerID, req.Provider, req.AWB); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Order shipped"})
}

