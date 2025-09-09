package handlers

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "gocom/main/internal/common/auth"
    "gocom/main/internal/common/errors"
    "gocom/main/internal/seller/services"
)

type SellerHandler struct {
    SellerService *services.SellerService
}

func NewSellerHandler() *SellerHandler {
    return &SellerHandler{
        SellerService: services.NewSellerService(),
    }
}


func (sh *SellerHandler) CreateSeller(c *gin.Context) {
    userID := auth.GetUserID(c)
    if userID == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    var req services.CreateSellerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "validation failed",
            "details": err.Error(),
        })
        return
    }

    seller, err := sh.SellerService.CreateSeller(userID, &req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "success": true,
        "data": seller,
        "message": "Seller registered successfully. Awaiting admin approval.",
    })
}

func (sh *SellerHandler) GetSeller(c *gin.Context) {
    sellerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, errors.ErrBadRequest)
        return
    }
    userID := auth.GetUserID(c)
    if userID == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    seller, err := sh.SellerService.GetSeller(uint(sellerID), userID)
    if err != nil {
        if err.Error() == "unauthorized access to seller" {
            c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        } else {
            c.JSON(http.StatusNotFound, gin.H{"error": "seller not found"})
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": seller,
    })
}

func (sh *SellerHandler) UpdateSeller(c *gin.Context) {
    sellerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, errors.ErrBadRequest)
        return
    }

    userID := auth.GetUserID(c)
    if userID == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    var req services.UpdateSellerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "validation failed",
            "details": err.Error(),
        })
        return
    }

    seller, err := sh.SellerService.UpdateSeller(uint(sellerID), userID, &req)
    if err != nil {
        if err.Error() == "unauthorized access to seller" {
            c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        } else {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": seller,
        "message": "Seller profile updated successfully",
    })
}
