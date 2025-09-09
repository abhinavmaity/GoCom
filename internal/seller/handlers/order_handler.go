package handlers

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"gocom/main/internal/common/auth"
	"gocom/main/internal/seller/services"
)

type OrderHandler struct {
	OrderService *services.OrderService
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{OrderService: services.NewOrderService()}
}

func (oh *OrderHandler) getSellerIDFromUser(userID uint) (uint, error) {
	return oh.OrderService.GetSellerIDByUserID(userID)
}

func (oh *OrderHandler) GetSellerOrders(c *gin.Context) {
	sellerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid seller ID"})
		return
	}

	userID := auth.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	orders, err := oh.OrderService.GetSellerOrders(uint(sellerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
		"count":   len(orders),
	})
}

func (oh *OrderHandler) GetOrderDetails(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	userID := auth.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	sellerID, err := oh.getSellerIDFromUser(userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No seller account found for user"})
		return
	}

	details, err := oh.OrderService.GetOrderDetails(uint(orderID), sellerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    details,
	})
}

func (oh *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	userID := auth.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sellerID, err := oh.getSellerIDFromUser(userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No seller account found for user"})
		return
	}

	if err := oh.OrderService.UpdateOrderStatus(uint(orderID), sellerID, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Order status updated successfully",
	})
}

func (oh *OrderHandler) ShipOrder(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	userID := auth.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Provider string `json:"provider" binding:"required"`
		AWB      string `json:"awb" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sellerID, err := oh.getSellerIDFromUser(userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No seller account found for user"})
		return
	}

	if err := oh.OrderService.ShipOrder(uint(orderID), sellerID, req.Provider, req.AWB); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Order shipped successfully"})
}
