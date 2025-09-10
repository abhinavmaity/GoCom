package services

import (
	"errors"
	"gocom/main/internal/models"
	"gorm.io/gorm"
	"strconv" // To convert addressID from string to uint
)

type CheckoutService struct {
	db *gorm.DB
}

func NewCheckoutService(db *gorm.DB) *CheckoutService {
	return &CheckoutService{db: db}
}

// Validate the cart, calculate totals, apply discounts, etc.
func (s *CheckoutService) ValidateCart(cartID string, addressIDStr string) (*models.Order, error) {
	// Convert addressID from string to uint
	addressID, err := strconv.ParseUint(addressIDStr, 10, 32) // Convert string to uint
	if err != nil {
		return nil, errors.New("invalid address_id")
	}

	var cart models.Cart
	if err := s.db.Where("id = ?", cartID).First(&cart).Error; err != nil {
		return nil, errors.New("cart not found")
	}

	// Placeholder for the rest of the validation logic (inventory, totals, etc.)
	// Assuming pricing and inventory validation happens here.

	// Create Order with correct addressID type (uint)
	order := &models.Order{
		UserID:    cart.UserID,
		Total:     1000,            // Placeholder for actual calculation
		AddressID: uint(addressID), // Convert addressID to uint
	}

	return order, nil
}

// Create a payment intent with Razorpay (just a placeholder)
func (s *CheckoutService) CreatePaymentIntent(order *models.Order) (string, error) {
	// Razorpay payment intent logic
	// Here we mock the Razorpay API call
	return "payment_intent_12345", nil
}
