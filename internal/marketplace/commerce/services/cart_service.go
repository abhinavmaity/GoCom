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

func (s *CartService) CreateCart(userID uint) (*models.Cart, error) {
	cart := models.Cart{
		UserID:   userID,
		Currency: "INR",
	}

	if err := s.db.Create(&cart).Error; err != nil {
		return nil, err
	}

	return &cart, nil
}

func (s *CartService) GetCart(cartID string, userID uint) (*models.Cart, error) {
	var cart models.Cart
	if err := s.db.Preload("Items").Where("id = ? AND user_id = ?", cartID, userID).First(&cart).Error; err != nil {
		return nil, errors.New("cart not found")
	}

	return &cart, nil
}

func (s *CartService) UpdateItem(userID uint, cartID string, itemID string, qty int) error {
	// First check if cart belongs to user
	var cart models.Cart
	if err := s.db.Where("id = ? AND user_id = ?", cartID, userID).First(&cart).Error; err != nil {
		return errors.New("cart not found")
	}

	// Update the cart item
	var cartItem models.CartItem
	if err := s.db.Where("id = ? AND cart_id = ?", itemID, cartID).First(&cartItem).Error; err != nil {
		return errors.New("cart item not found")
	}

	if qty <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	cartItem.Qty = qty
	if err := s.db.Save(&cartItem).Error; err != nil {
		return err
	}

	return nil
}

func (s *CartService) RemoveItem(userID uint, cartID string, itemID string) error {
	// First check if cart belongs to user
	var cart models.Cart
	if err := s.db.Where("id = ? AND user_id = ?", cartID, userID).First(&cart).Error; err != nil {
		return errors.New("cart not found")
	}

	// Delete the cart item
	if err := s.db.Where("id = ? AND cart_id = ?", itemID, cartID).Delete(&models.CartItem{}).Error; err != nil {
		return errors.New("failed to remove cart item")
	}

	return nil
}
