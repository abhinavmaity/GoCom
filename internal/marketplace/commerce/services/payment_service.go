package services

import (
	"errors"
	"gocom/main/internal/models"
	"gorm.io/gorm"
	// Add Razorpay SDK or mock the logic for PoC
)

type PaymentService struct {
	db *gorm.DB
}

func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{db: db}
}

// CapturePayment simulates capturing payment through Razorpay (in PoC, mock capture)
func (s *PaymentService) CapturePayment(intentID string) (*models.Order, error) {
	// Find the payment intent
	var payment models.Payment
	if err := s.db.Where("intent_id = ?", intentID).First(&payment).Error; err != nil {
		return nil, errors.New("payment intent not found")
	}

	// Simulate payment capture via Razorpay API (or sandbox)
	if payment.Status != 0 {
		return nil, errors.New("payment already captured or failed")
	}

	// Mark payment as captured
	payment.Status = 1
	if err := s.db.Save(&payment).Error; err != nil {
		return nil, err
	}

	// Mark the order as confirmed
	var order models.Order
	if err := s.db.Where("id = ?", payment.OrderID).First(&order).Error; err != nil {
		return nil, err
	}
	order.Status = "1"
	if err := s.db.Save(&order).Error; err != nil {
		return nil, err
	}

	// Return the confirmed order
	return &order, nil
}
