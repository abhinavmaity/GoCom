package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "gocom/main/internal/common/auth"
    "gocom/main/internal/seller/services"
    "gocom/main/internal/models"
)

type ProductHandler struct {
    ProductService *services.ProductService
}

func NewProductHandler() *ProductHandler {
    return &ProductHandler{
        ProductService: services.NewProductService(),
    }
}

func (ph *ProductHandler) CreateProduct(c *gin.Context) {
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

    if !ph.hasSellerAccess(userID, uint(sellerID)) {
        c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access to seller"})
        return
    }

    var req services.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "validation failed",
            "details": err.Error(),
        })
        return
    }

    product, err := ph.ProductService.CreateProduct(uint(sellerID), &req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "success": true,
        "data":    product,
        "message": "Product created successfully",
    })
}

func (ph *ProductHandler) GetProduct(c *gin.Context) {
    productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }

    product, err := ph.ProductService.GetProduct(uint(productID))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    product,
    })
}

func (ph *ProductHandler) ListProducts(c *gin.Context) {
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

    if !ph.hasSellerAccess(userID, uint(sellerID)) {
        c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access to seller"})
        return
    }
    filters := &services.ProductFilters{}
    
    if statusStr := c.Query("status"); statusStr != "" {
        if status, err := strconv.Atoi(statusStr); err == nil {
            filters.Status = &status
        }
    }

    if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
        if categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32); err == nil {
            categoryIDUint := uint(categoryID)
            filters.CategoryID = &categoryIDUint
        }
    }

    filters.Search = c.Query("search")

    products, err := ph.ProductService.ListProducts(uint(sellerID), filters)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    products,
        "count":   len(products),
    })
}

func (ph *ProductHandler) UpdateProduct(c *gin.Context) {
    productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }
    var product models.Product
    if err := ph.ProductService.DB.First(&product, productID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    sellerID := product.SellerID
    userID := auth.GetUserID(c)
    if userID == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    if !ph.hasSellerAccess(userID, sellerID) {
        c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access to seller"})
        return
    }

    var req services.UpdateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "validation failed",
            "details": err.Error(),
        })
        return
    }

    updatedProduct, err := ph.ProductService.UpdateProduct(uint(productID), sellerID, &req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    updatedProduct,
        "message": "Product updated successfully",
    })
}

func (ph *ProductHandler) DeleteProduct(c *gin.Context) {
    productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }
    var product models.Product
    if err := ph.ProductService.DB.First(&product, productID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    sellerID := product.SellerID
    userID := auth.GetUserID(c)
    if userID == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    if !ph.hasSellerAccess(userID, sellerID) {
        c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access to seller"})
        return
    }

    if err := ph.ProductService.DeleteProduct(uint(productID), sellerID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Product deleted successfully",
    })
}

func (ph *ProductHandler) PublishProduct(c *gin.Context) {
    productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }
    var product models.Product
    if err := ph.ProductService.DB.First(&product, productID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    sellerID := product.SellerID
    userID := auth.GetUserID(c)
    if userID == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    if !ph.hasSellerAccess(userID, sellerID) {
        c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access to seller"})
        return
    }

    publishedProduct, err := ph.ProductService.PublishProduct(uint(productID), sellerID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    publishedProduct,
        "message": "Product published successfully",
    })
}

func (ph *ProductHandler) hasSellerAccess(userID, sellerID uint) bool {
    var count int64
    ph.ProductService.DB.Model(&models.SellerUser{}).
        Where("user_id = ? AND seller_id = ? AND status = 1", userID, sellerID).
        Count(&count)
    return count > 0
}
