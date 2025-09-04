// internal/marketplace/commerce/services/cart_service.go
package services

import (
	"errors"
	"github.com/shopspring/decimal"
	"gocom/main/internal/marketplace/commerce/dto"
	"gocom/main/internal/models"
	"gorm.io/gorm"
)

type CartService struct {
	db *gorm.DB
}

func NewCartService(db *gorm.DB) *CartService {
	return &CartService{db: db}
}

func (s *CartService) CreateCart(userID uint, currency string) (*models.Cart, error) {
	cart := &models.Cart{
		UserID:   userID,
		Currency: currency,
	}

	err := s.db.Create(cart).Error
	return cart, err
}

func (s *CartService) GetCart(cartID, userID uint) (*dto.CartResponse, error) {
	var cart models.Cart
	err := s.db.Where("id = ? AND user_id = ?", cartID, userID).
		Preload("Items").
		Preload("Items.SKU").
		Preload("Items.SKU.Product").
		First(&cart).Error

	if err != nil {
		return nil, err
	}

	return s.mapCartToResponse(&cart), nil
}

func (s *CartService) AddItem(cartID, userID, skuID uint, qty int) error {
	// Verify cart belongs to user
	var cart models.Cart
	err := s.db.Where("id = ? AND user_id = ?", cartID, userID).First(&cart).Error
	if err != nil {
		return errors.New("cart not found")
	}

	// Check if item already exists
	var existingItem models.CartItem
	err = s.db.Where("cart_id = ? AND sku_id = ?", cartID, skuID).First(&existingItem).Error

	if err == gorm.ErrRecordNotFound {
		// Get SKU price
		var sku models.SKU
		err = s.db.Where("id = ?", skuID).First(&sku).Error
		if err != nil {
			return errors.New("SKU not found")
		}

		// Create new item
		item := &models.CartItem{
			CartID: cartID,
			SKUID:  skuID,
			Qty:    qty,
			Price:  sku.PriceSell,
		}
		return s.db.Create(item).Error
	} else if err != nil {
		return err
	}

	// Update existing item
	existingItem.Qty += qty
	return s.db.Save(&existingItem).Error
}

func (s *CartService) UpdateItemQty(cartID, userID, itemID uint, qty int) error {
	return s.db.Model(&models.CartItem{}).
		Where("id = ? AND cart_id IN (SELECT id FROM carts WHERE id = ? AND user_id = ?)",
			itemID, cartID, userID).
		Update("qty", qty).Error
}

func (s *CartService) RemoveItem(cartID, userID, itemID uint) error {
	return s.db.Where("id = ? AND cart_id IN (SELECT id FROM carts WHERE id = ? AND user_id = ?)",
		itemID, cartID, userID).
		Delete(&models.CartItem{}).Error
}

func (s *CartService) mapCartToResponse(cart *models.Cart) *dto.CartResponse {
	var items []dto.CartItemResponse
	var total decimal.Decimal

	for _, item := range cart.Items {
		subtotal := item.Price.Mul(decimal.NewFromInt(int64(item.Qty)))
		total = total.Add(subtotal)

		items = append(items, dto.CartItemResponse{
			ID:       item.ID,
			SKUID:    item.SKUID,
			SKUCode:  item.SKU.SKUCode,
			Title:    item.SKU.Product.Title,
			Price:    item.Price,
			Qty:      item.Qty,
			Subtotal: subtotal,
		})
	}

	return &dto.CartResponse{
		ID:        cart.ID,
		UserID:    cart.UserID,
		Currency:  cart.Currency,
		Items:     items,
		Total:     total,
		CreatedAt: cart.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
