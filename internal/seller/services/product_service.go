package services

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gocom/main/internal/common/db"
	"gocom/main/internal/models"
)

type ProductService struct {
	DB *gorm.DB
}

func NewProductService() *ProductService {
	return &ProductService{
		DB: db.GetDB(),
	}
}

// Create new product
func (ps *ProductService) CreateProduct(sellerID uint, req *CreateProductRequest) (*models.Product, error) {
	// Validate category exists
	var category models.Category
	if err := ps.DB.First(&category, req.CategoryID).Error; err != nil {
		return nil, errors.New("invalid category")
	}

	// Create product
	product := &models.Product{
		SellerID:    sellerID,
		CategoryID:  req.CategoryID,
		Title:       req.Title,
		Description: req.Description,
		Brand:       req.Brand,
		Status:      0, // Draft
		Score:       ps.calculateQualityScore(req),
	}

	if err := ps.DB.Create(product).Error; err != nil {
		return nil, err
	}

	return product, nil
}

// Get product by ID
func (ps *ProductService) GetProduct(productID uint) (*ProductResponse, error) {
	var product models.Product
	if err := ps.DB.Preload("Category").First(&product, productID).Error; err != nil {
		return nil, errors.New("product not found")
	}

	// Get SKUs count
	var skuCount int64
	ps.DB.Model(&models.SKU{}).Where("product_id = ?", productID).Count(&skuCount)

	// Get media count
	var mediaCount int64
	ps.DB.Model(&models.Media{}).Where("entity_type = ? AND entity_id = ?", "product", productID).Count(&mediaCount)

	response := &ProductResponse{
		ID:          product.ID,
		SellerID:    product.SellerID,
		CategoryID:  product.CategoryID,
		Title:       product.Title,
		Description: product.Description,
		Brand:       product.Brand,
		Status:      product.Status,
		StatusText:  ps.getStatusText(product.Status),
		Score:       product.Score,
		SKUCount:    int(skuCount),
		MediaCount:  int(mediaCount),
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}

	return response, nil
}

// List products for seller
func (ps *ProductService) ListProducts(sellerID uint, filters *ProductFilters) ([]ProductSummary, error) {
    query := ps.DB.Model(&models.Product{}).Where("seller_id = ?", sellerID)
    
    // Apply filters
    if filters != nil {
        if filters.Status != nil {
            query = query.Where("status = ?", *filters.Status)
        }
        
        if filters.CategoryID != nil {
            query = query.Where("category_id = ?", *filters.CategoryID)
        }
        
        if filters.Search != "" {
            // 🔧 FIX: Use LIKE instead of ILIKE for MySQL compatibility
            query = query.Where("title LIKE ? OR brand LIKE ?", "%"+filters.Search+"%", "%"+filters.Search+"%")
        }
    }

    var products []models.Product
    if err := query.Order("created_at DESC").Find(&products).Error; err != nil {
        return nil, err
    }

    var summary []ProductSummary
    for _, product := range products {
        // Get SKU count for each product
        var skuCount int64
        ps.DB.Model(&models.SKU{}).Where("product_id = ?", product.ID).Count(&skuCount)
        
        summary = append(summary, ProductSummary{
            ID:         product.ID,
            Title:      product.Title,
            Brand:      product.Brand,
            Status:     product.Status,
            StatusText: ps.getStatusText(product.Status),
            SKUCount:   int(skuCount),
            CreatedAt:  product.CreatedAt,
        })
    }

    return summary, nil
}

// Update product
func (ps *ProductService) UpdateProduct(productID, sellerID uint, req *UpdateProductRequest) (*models.Product, error) {
	var product models.Product
	if err := ps.DB.Where("id = ? AND seller_id = ?", productID, sellerID).First(&product).Error; err != nil {
		return nil, errors.New("product not found")
	}

	// Update fields if provided
	if req.Title != nil {
		product.Title = *req.Title
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Brand != nil {
		product.Brand = *req.Brand
	}
	if req.CategoryID != nil {
		// Validate category
		var category models.Category
		if err := ps.DB.First(&category, *req.CategoryID).Error; err != nil {
			return nil, errors.New("invalid category")
		}
		product.CategoryID = *req.CategoryID
	}

	product.UpdatedAt = time.Now()

	if err := ps.DB.Save(&product).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

// Delete product
func (ps *ProductService) DeleteProduct(productID, sellerID uint) error {
	// Check for existing orders
	var orderCount int64
	ps.DB.Table("order_items").
		Joins("JOIN skus ON order_items.sku_id = skus.id").
		Where("skus.product_id = ?", productID).
		Count(&orderCount)

	if orderCount > 0 {
		return errors.New("cannot delete product with existing orders")
	}

	// Delete in transaction
	tx := ps.DB.Begin()

	// Delete inventories
	if err := tx.Where("sku_id IN (SELECT id FROM skus WHERE product_id = ?)", productID).Delete(&models.Inventory{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete SKUs
	if err := tx.Where("product_id = ?", productID).Delete(&models.SKU{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete media
	if err := tx.Where("entity_type = ? AND entity_id = ?", "product", productID).Delete(&models.Media{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete product
	if err := tx.Where("id = ? AND seller_id = ?", productID, sellerID).Delete(&models.Product{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Publish product
func (ps *ProductService) PublishProduct(productID, sellerID uint) (*models.Product, error) {
	var product models.Product
	if err := ps.DB.Where("id = ? AND seller_id = ?", productID, sellerID).First(&product).Error; err != nil {
		return nil, errors.New("product not found")
	}

	// Validate for publishing
	if err := ps.validateForPublishing(&product); err != nil {
		return nil, err
	}

	product.Status = 1 // Published
	product.UpdatedAt = time.Now()

	if err := ps.DB.Save(&product).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

// Helper methods
func (ps *ProductService) validateForPublishing(product *models.Product) error {
	if product.Title == "" {
		return errors.New("product title is required")
	}
	if product.Description == "" {
		return errors.New("product description is required")
	}

	// Check SKUs
	var skuCount int64
	ps.DB.Model(&models.SKU{}).Where("product_id = ?", product.ID).Count(&skuCount)
	if skuCount == 0 {
		return errors.New("product must have at least one SKU")
	}

	// Check images
	var mediaCount int64
	ps.DB.Model(&models.Media{}).Where("entity_type = ? AND entity_id = ? AND type = ?", "product", product.ID, "image").Count(&mediaCount)
	if mediaCount == 0 {
		return errors.New("product must have at least one image")
	}

	return nil
}

func (ps *ProductService) calculateQualityScore(req *CreateProductRequest) int {
	score := 50 // Base score

	if len(req.Title) > 10 {
		score += 15
	}
	if len(req.Description) > 50 {
		score += 20
	}
	if req.Brand != "" {
		score += 15
	}

	return score
}

func (ps *ProductService) getStatusText(status int) string {
	switch status {
	case 0:
		return "Draft"
	case 1:
		return "Published"
	case 2:
		return "Suspended"
	default:
		return "Unknown"
	}
}

// DTOs
type CreateProductRequest struct {
	CategoryID  uint   `json:"category_id" binding:"required"`
	Title       string `json:"title" binding:"required,min=3,max=200"`
	Description string `json:"description" binding:"max=2000"`
	Brand       string `json:"brand" binding:"max=100"`
}

type UpdateProductRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Brand       *string `json:"brand,omitempty"`
	CategoryID  *uint   `json:"category_id,omitempty"`
}

type ProductFilters struct {
	Status     *int   `json:"status,omitempty"`
	CategoryID *uint  `json:"category_id,omitempty"`
	Search     string `json:"search,omitempty"`
}

type ProductResponse struct {
	ID          uint      `json:"id"`
	SellerID    uint      `json:"seller_id"`
	CategoryID  uint      `json:"category_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Brand       string    `json:"brand"`
	Status      int       `json:"status"`
	StatusText  string    `json:"status_text"`
	Score       int       `json:"score"`
	SKUCount    int       `json:"sku_count"`
	MediaCount  int       `json:"media_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductSummary struct {
	ID         uint      `json:"id"`
	Title      string    `json:"title"`
	Brand      string    `json:"brand"`
	Status     int       `json:"status"`
	StatusText string    `json:"status_text"`
	SKUCount   int       `json:"sku_count"`
	CreatedAt  time.Time `json:"created_at"`
}

