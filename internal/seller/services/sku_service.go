package services

import (
	"errors"
	"fmt"
	"time"
	"encoding/json"

	"gorm.io/gorm"
	"gocom/main/internal/common/db"
	"gocom/main/internal/models"
	"github.com/shopspring/decimal"
)

type SKUService struct {
	DB *gorm.DB
}

func NewSKUService() *SKUService {
	return &SKUService{
		DB: db.GetDB(),
	}
}

// Add SKU to product
func (ss *SKUService) CreateSKU(productID uint, req *CreateSKURequest) (*models.SKU, error) {
	// Verify product exists and get seller ID for authorization
	var product models.Product
	if err := ss.DB.First(&product, productID).Error; err != nil {
		return nil, errors.New("product not found")
	}

	// Generate unique SKU code if not provided
	skuCode := req.SKUCode
	if skuCode == "" {
		skuCode = ss.generateSKUCode(product.ID)
	}

	// Check SKU code uniqueness
	var existingSKU models.SKU
	if err := ss.DB.Where("sku_code = ?", skuCode).First(&existingSKU).Error; err == nil {
		return nil, errors.New("SKU code already exists")
	}

	// Validate pricing
	if req.PriceSell.GreaterThan(req.PriceMRP) {
		return nil, errors.New("selling price cannot be greater than MRP")
	}

	// Convert attributes to JSON
	attributesJSON, err := json.Marshal(req.Attributes)
	if err != nil {
		return nil, err
	}

	// Create SKU
	sku := &models.SKU{
		ProductID:  productID,
		SKUCode:    skuCode,
		Attributes: attributesJSON,
		PriceMRP:   req.PriceMRP,
		PriceSell:  req.PriceSell,
		TaxPct:     req.TaxPct,
		Barcode:    req.Barcode,
	}

	if err := ss.DB.Create(sku).Error; err != nil {
		return nil, err
	}

	// Create initial inventory record
	inventory := &models.Inventory{
		SKUID:      sku.ID,
		LocationID: 1, // Default location
		OnHand:     0,
		Reserved:   0,
		Threshold:  req.Threshold,
	}

	ss.DB.Create(inventory)

	return sku, nil
}

// Get all SKUs for a product
func (ss *SKUService) GetProductSKUs(productID uint) ([]SKUResponse, error) {
	var skus []models.SKU
	err := ss.DB.Where("product_id = ?", productID).Find(&skus).Error
	if err != nil {
		return nil, err
	}

	var response []SKUResponse
	for _, sku := range skus {
		// Get inventory info
		var inventory models.Inventory
		ss.DB.Where("sku_id = ?", sku.ID).First(&inventory)

		// Parse attributes
		var attributes map[string]interface{}
		json.Unmarshal(sku.Attributes, &attributes)

		skuResp := SKUResponse{
			ID:         sku.ID,
			ProductID:  sku.ProductID,
			SKUCode:    sku.SKUCode,
			Attributes: attributes,
			PriceMRP:   sku.PriceMRP,
			PriceSell:  sku.PriceSell,
			TaxPct:     sku.TaxPct,
			Barcode:    sku.Barcode,
			Inventory: InventoryInfo{
				OnHand:    inventory.OnHand,
				Reserved:  inventory.Reserved,
				Available: inventory.OnHand - inventory.Reserved,
				Threshold: inventory.Threshold,
			},
			CreatedAt: sku.CreatedAt,
		}
		response = append(response, skuResp)
	}

	return response, nil
}

// Update SKU
func (ss *SKUService) UpdateSKU(skuID uint, req *UpdateSKURequest) (*models.SKU, error) {
	var sku models.SKU
	if err := ss.DB.First(&sku, skuID).Error; err != nil {
		return nil, errors.New("SKU not found")
	}

	// Update fields if provided
	if req.Attributes != nil {
		attributesJSON, _ := json.Marshal(req.Attributes)
		sku.Attributes = attributesJSON
	}
	if req.PriceMRP != nil {
		sku.PriceMRP = *req.PriceMRP
	}
	if req.PriceSell != nil {
		// Validate pricing
		if req.PriceSell.GreaterThan(sku.PriceMRP) {
			return nil, errors.New("selling price cannot be greater than MRP")
		}
		sku.PriceSell = *req.PriceSell
	}
	if req.TaxPct != nil {
		sku.TaxPct = *req.TaxPct
	}
	if req.Barcode != nil {
		sku.Barcode = *req.Barcode
	}

	if err := ss.DB.Save(&sku).Error; err != nil {
		return nil, err
	}

	return &sku, nil
}

// Delete SKU
func (ss *SKUService) DeleteSKU(skuID uint) error {
	// Check if SKU has pending orders
	var orderItemCount int64
	ss.DB.Model(&models.OrderItem{}).Where("sku_id = ?", skuID).Count(&orderItemCount)
	
	if orderItemCount > 0 {
		return errors.New("cannot delete SKU with existing orders")
	}

	// Delete inventory first (cascade)
	ss.DB.Where("sku_id = ?", skuID).Delete(&models.Inventory{})

	// Delete SKU
	return ss.DB.Delete(&models.SKU{}, skuID).Error
}

// Generate unique SKU code
func (ss *SKUService) generateSKUCode(productID uint) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("SKU-%d-%d", productID, timestamp)
}

// Request/Response DTOs
type CreateSKURequest struct {
	SKUCode    string                 `json:"sku_code"`
	Attributes map[string]interface{} `json:"attributes" binding:"required"`
	PriceMRP   decimal.Decimal        `json:"price_mrp" binding:"required"`
	PriceSell  decimal.Decimal        `json:"price_sell" binding:"required"`
	TaxPct     decimal.Decimal        `json:"tax_pct" binding:"required"`
	Barcode    string                 `json:"barcode"`
	Threshold  int                    `json:"threshold"`
}

type UpdateSKURequest struct {
	Attributes *map[string]interface{} `json:"attributes,omitempty"`
	PriceMRP   *decimal.Decimal        `json:"price_mrp,omitempty"`
	PriceSell  *decimal.Decimal        `json:"price_sell,omitempty"`
	TaxPct     *decimal.Decimal        `json:"tax_pct,omitempty"`
	Barcode    *string                 `json:"barcode,omitempty"`
}

type SKUResponse struct {
	ID         uint                   `json:"id"`
	ProductID  uint                   `json:"product_id"`
	SKUCode    string                 `json:"sku_code"`
	Attributes map[string]interface{} `json:"attributes"`
	PriceMRP   decimal.Decimal        `json:"price_mrp"`
	PriceSell  decimal.Decimal        `json:"price_sell"`
	TaxPct     decimal.Decimal        `json:"tax_pct"`
	Barcode    string                 `json:"barcode"`
	Inventory  InventoryInfo          `json:"inventory"`
	CreatedAt  time.Time              `json:"created_at"`
}

type InventoryInfo struct {
	OnHand    int `json:"on_hand"`
	Reserved  int `json:"reserved"`
	Available int `json:"available"`
	Threshold int `json:"threshold"`
}

