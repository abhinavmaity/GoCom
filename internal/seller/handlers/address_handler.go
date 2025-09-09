package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "gocom/main/internal/common/auth"
    "gocom/main/internal/seller/services"
    "gocom/main/internal/models"
)

type AddressHandler struct {
    AddressService *services.AddressService
}

func NewAddressHandler() *AddressHandler {
    return &AddressHandler{AddressService: services.NewAddressService()}
}

func (ah *AddressHandler) AddAddress(c *gin.Context) {
    sellerID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    userID := auth.GetUserID(c)
    if userID == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    if !ah.hasSellerAccess(userID, uint(sellerID)) {
        c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access to seller"})
        return
    }

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

func (ah *AddressHandler) GetSellerAddresses(c *gin.Context) {
    sellerID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    userID := auth.GetUserID(c)
    if userID == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    if !ah.hasSellerAccess(userID, uint(sellerID)) {
        c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access to seller"})
        return
    }

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

func (ah *AddressHandler) UpdateAddress(c *gin.Context) {
    addressID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    userID := auth.GetUserID(c)
    if userID == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
    var address models.Address
    if err := ah.AddressService.DB.First(&address, addressID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
        return
    }

    if address.SellerID == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address"})
        return
    }

    sellerID := *address.SellerID
    if !ah.hasSellerAccess(userID, sellerID) {
        c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access to seller"})
        return
    }

    var req services.UpdateAddressRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    updatedAddress, err := ah.AddressService.UpdateAddress(uint(addressID), sellerID, &req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    updatedAddress,
        "message": "Address updated successfully",
    })
}

func (ah *AddressHandler) DeleteAddress(c *gin.Context) {
    addressID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    userID := auth.GetUserID(c)
    if userID == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
    var address models.Address
    if err := ah.AddressService.DB.First(&address, addressID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
        return
    }
    if address.SellerID == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address"})
        return
    }

    sellerID := *address.SellerID
    if !ah.hasSellerAccess(userID, sellerID) {
        c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access to seller"})
        return
    }

    if err := ah.AddressService.DeleteAddress(uint(addressID), sellerID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Address deleted successfully",
    })
}

func (ah *AddressHandler) hasSellerAccess(userID, sellerID uint) bool {
    var count int64
    ah.AddressService.DB.Model(&models.SellerUser{}).
        Where("user_id = ? AND seller_id = ? AND status = 1", userID, sellerID).
        Count(&count)
    return count > 0
}
