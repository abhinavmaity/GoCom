package services

import (
    "errors"
    "fmt"
    "strings"
    "gorm.io/gorm"
    "github.com/shopspring/decimal"
    
    "gocom/main/internal/models"
    "gocom/main/internal/common/db"
)

type ProductService struct {
    DB *gorm.DB
}

func NewProductService() *ProductService {
    return &ProductService{
        DB: db.GetDB(),
    }
}

// Create product with SKUs
func (ps *ProductService) CreateProduct(sellerID uint, req *CreateProductRequest) (*models.Product, error) {
    // Validate category exists
    var category models.Category
    if err := ps.DB.First(&category, req.CategoryID).Error; err != nil {
        return nil, errors.New("invalid category")
    }
    
    // Generate content quality score
    score := ps.calculateContentScore(req)
    
    // Create product
    product := &models.Product{
        SellerID:    sellerID,
        CategoryID:  req.CategoryID,
        Title:       req.Title,
        Description: req.Description,
        Brand:       req.Brand,
        Status:      models.ProductStatusDraft,
        Score:       score,
    }
    
    // Begin transaction
    tx := ps.DB.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    
    // Save product
    if err := tx.Create(product).Error; err != nil {
        tx.Rollback()
        return nil, err
    }
    
    // Create SKUs
    for _, skuReq := range req.SKUs {
        sku := &models.SKU{
            ProductID: product.ID,
            SKUCode:   ps.generateSKUCode(product.ID, skuReq.Attributes),
            PriceMRP:  skuReq.PriceMRP,
            PriceSell: skuReq.PriceSell,
            TaxPct:    skuReq.TaxPct,
            Barcode:   skuReq.Barcode,
        }
        
        // Set attributes
        if err := sku.SetAttributes(skuReq.Attributes); err != nil {
            tx.Rollback()
            return nil, err
        }
        
        if err := tx.Create(sku).Error; err != nil {
            tx.Rollback()
            return nil, err
        }
    }
    
    // Commit transaction
    if err := tx.Commit().Error; err != nil {
        return nil, err
    }
    
    // Return product with relations
    return ps.GetProduct(product.ID, sellerID)
}

// Get product by ID
func (ps *ProductService) GetProduct(productID, sellerID uint) (*models.Product, error) {
    var product models.Product
    
    err := ps.DB.
        Preload("Category").
        Preload("SKUs").
        Preload("Media").
        Where("id = ? AND seller_id = ?", productID, sellerID).
        First(&product).Error
        
    return &product, err
}

// List seller products
func (ps *ProductService) ListProducts(sellerID uint, filters ProductFilters) ([]models.Product, int64, error) {
    var products []models.Product
    var total int64
    
    query := ps.DB.Model(&models.Product{}).Where("seller_id = ?", sellerID)
    
    // Apply filters
    if filters.Status != nil {
        query = query.Where("status = ?", *filters.Status)
    }
    if filters.CategoryID != nil {
        query = query.Where("category_id = ?", *filters.CategoryID)
    }
    if filters.Search != "" {
        query = query.Where("title ILIKE ? OR description ILIKE ?", 
            "%"+filters.Search+"%", "%"+filters.Search+"%")
    }
    
    // Get total count
    query.Count(&total)
    
    // Apply pagination
    offset := (filters.Page - 1) * filters.Limit
    err := query.
        Preload("Category").
        Preload("SKUs").
        Offset(offset).
        Limit(filters.Limit).
        Order("created_at DESC").
        Find(&products).Error
        
    return products, total, err
}

// Publish product
func (ps *ProductService) PublishProduct(productID, sellerID uint) error {
    // Validate product can be published
    var product models.Product
    if err := ps.DB.Where("id = ? AND seller_id = ?", productID, sellerID).First(&product).Error; err != nil {
        return err
    }
    
    // Check minimum quality score
    if product.Score < 60 {
        return errors.New("product quality score too low for publishing")
    }
    
    // Update status
    return ps.DB.Model(&product).Update("status", models.ProductStatusPublished).Error
}

// Calculate content quality score
func (ps *ProductService) calculateContentScore(req *CreateProductRequest) int {
    score := 0
    
    // Title quality (0-20 points)
    if len(req.Title) >= 10 && len(req.Title) <= 100 {
        score += 20
    } else if len(req.Title) >= 5 {
        score += 10
    }
    
    // Description quality (0-25 points)
    if len(req.Description) >= 100 && len(req.Description) <= 1000 {
        score += 25
    } else if len(req.Description) >= 50 {
        score += 15
    }
    
    // Brand presence (0-10 points)
    if req.Brand != "" {
        score += 10
    }
    
    // SKU completeness (0-25 points)
    if len(req.SKUs) > 0 {
        score += 15
        for _, sku := range req.SKUs {
            if sku.PriceMRP.GreaterThan(sku.PriceSell) && sku.PriceSell.GreaterThan(decimal.Zero) {
                score += 10
                break
            }
        }
    }
    
    // Media presence (0-20 points) - will be updated when media is added
    
    return score
}

// Generate SKU code
func (ps *ProductService) generateSKUCode(productID uint, attrs models.SKUAttributes) string {
    base := fmt.Sprintf("PRD%d", productID)
    var parts []string
    
    if attrs.Color != "" {
        parts = append(parts, strings.ToUpper(attrs.Color[:1]))
    }
    if attrs.Size != "" {
        parts = append(parts, strings.ToUpper(attrs.Size))
    }
    
    if len(parts) > 0 {
        return fmt.Sprintf("%s-%s", base, strings.Join(parts, ""))
    }
    
    return base
}

// Request DTOs
type CreateProductRequest struct {
    CategoryID  uint                    `json:"category_id" binding:"required"`
    Title       string                  `json:"title" binding:"required,min=5,max=100"`
    Description string                  `json:"description" binding:"required,min=10"`
    Brand       string                  `json:"brand"`
    SKUs        []CreateSKURequest      `json:"skus" binding:"required,min=1"`
}

type CreateSKURequest struct {
    Attributes models.SKUAttributes    `json:"attributes"`
    PriceMRP   decimal.Decimal         `json:"price_mrp" binding:"required"`
    PriceSell  decimal.Decimal         `json:"price_sell" binding:"required"`
    TaxPct     decimal.Decimal         `json:"tax_pct"`
    Barcode    string                  `json:"barcode"`
}

type ProductFilters struct {
    Status     *int   `form:"status"`
    CategoryID *uint  `form:"category_id"`
    Search     string `form:"search"`
    Page       int    `form:"page,default=1"`
    Limit      int    `form:"limit,default=20"`
}

