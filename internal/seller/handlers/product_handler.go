package handlers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    
    "gocom/main/internal/seller/services"
    "gocom/main/internal/common/errors"
)

type ProductHandler struct {
    ProductService *services.ProductService
}

func NewProductHandler() *ProductHandler {
    return &ProductHandler{
        ProductService: services.NewProductService(),
    }
}

// Create product
// POST /v1/sellers/:id/products
func (ph *ProductHandler) CreateProduct(c *gin.Context) {
    sellerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, errors.ErrBadRequest)
        return
    }
    
    // TODO: Validate seller owns this seller_id via JWT middleware
    
    var req services.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    product, err := ph.ProductService.CreateProduct(uint(sellerID), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "success": true,
        "data":    product,
        "message": "Product created successfully",
    })
}

// List seller products
// GET /v1/sellers/:id/products
func (ph *ProductHandler) ListProducts(c *gin.Context) {
    sellerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, errors.ErrBadRequest)
        return
    }
    
    var filters services.ProductFilters
    if err := c.ShouldBindQuery(&filters); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    products, total, err := ph.ProductService.ListProducts(uint(sellerID), filters)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "products": products,
            "total":    total,
            "page":     filters.Page,
            "limit":    filters.Limit,
        },
    })
}

// Get single product
// GET /v1/products/:id
func (ph *ProductHandler) GetProduct(c *gin.Context) {
    productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, errors.ErrBadRequest)
        return
    }
    
    // TODO: Get seller ID from JWT
    sellerID := uint(1) // Placeholder
    
    product, err := ph.ProductService.GetProduct(uint(productID), sellerID)
    if err != nil {
        c.JSON(http.StatusNotFound, errors.ErrNotFound)
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    product,
    })
}

// Publish product
// POST /v1/products/:id/publish
func (ph *ProductHandler) PublishProduct(c *gin.Context) {
    productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, errors.ErrBadRequest)
        return
    }
    
    // TODO: Get seller ID from JWT
    sellerID := uint(1) // Placeholder
    
    if err := ph.ProductService.PublishProduct(uint(productID), sellerID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Product published successfully",
    })
}

