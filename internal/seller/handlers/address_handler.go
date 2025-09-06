package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gocom/main/internal/seller/services"
)

type AddressHandler struct {
	AddressService *services.AddressService
}

func NewAddressHandler() *AddressHandler {
	return &AddressHandler{AddressService: services.NewAddressService()}
}

// Add seller address
// POST /sellers/:id/addresses
func (ah *AddressHandler) AddAddress(c *gin.Context) {
	sellerID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req services.AddAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	address, err := ah.AddressService.AddAddress(uint(sellerID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    address,
		"message": "Address added successfully",
	})
}

// Get seller addresses
// GET /sellers/:id/addresses
func (ah *AddressHandler) GetSellerAddresses(c *gin.Context) {
	sellerID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	addresses, err := ah.AddressService.GetSellerAddresses(uint(sellerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    addresses,
		"count":   len(addresses),
	})
}

// Update address
// PATCH /addresses/:id
func (ah *AddressHandler) UpdateAddress(c *gin.Context) {
	addressID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	sellerID := uint(1) // TODO: Get from JWT

	var req services.UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	address, err := ah.AddressService.UpdateAddress(uint(addressID), sellerID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    address,
		"message": "Address updated successfully",
	})
}

// Delete address
// DELETE /addresses/:id
func (ah *AddressHandler) DeleteAddress(c *gin.Context) {
	addressID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	sellerID := uint(1) // TODO: Get from JWT

	if err := ah.AddressService.DeleteAddress(uint(addressID), sellerID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Address deleted successfully",
	})
}

