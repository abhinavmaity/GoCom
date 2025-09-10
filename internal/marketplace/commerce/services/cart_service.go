package services

import (
	"errors"
	"gocom/main/internal/models"
	"gorm.io/gorm"
)

type CartService struct {
	db *gorm.DB
}

func NewCartService(db *gorm.DB) *CartService {
	return &CartService{db: db}
}

func (s *CartService) AddItem(userID uint, cartID string, skuID uint, qty int) error {
	// Check if cart exists
	var cart models.Cart
	if err := s.db.Where("id = ? AND user_id = ?", cartID, userID).First(&cart).Error; err != nil {
		return errors.New("cart not found")
	}

	// Create a CartItem
	var cartItem models.CartItem
	if err := s.db.Where("cart_id = ? AND sku_id = ?", cartID, skuID).First(&cartItem).Error; err != nil {
		// If item doesn't exist, create a new one
		cartItem = models.CartItem{
			CartID: cart.ID,
			SKUID:  skuID,
			Qty:    qty,
		}
		if err := s.db.Create(&cartItem).Error; err != nil {
			return err
		}
	} else {
		// If item exists, update quantity
		cartItem.Qty += qty
		if err := s.db.Save(&cartItem).Error; err != nil {
			return err
		}
	}
	return nil
}
