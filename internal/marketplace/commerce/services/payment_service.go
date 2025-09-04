// internal/marketplace/commerce/services/payment_service.go
package services

import (
	"fmt"
	"github.com/shopspring/decimal"
	"gocom/main/internal/integrations/payment"
	"gocom/main/internal/marketplace/commerce/dto"
	"gocom/main/internal/models"
	"gorm.io/gorm"
	"strconv"
)

type PaymentService struct {
	db             *gorm.DB
	razorpayClient *payment.RazorpayClient
}

func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{
		db:             db,
		razorpayClient: payment.NewRazorpayClient(),
	}
}

func (s *PaymentService) CreatePaymentIntent(orderID, userID uint) (*dto.PaymentIntentResponse, error) {
	// Get order details
	var order models.Order
	err := s.db.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error
	if err != nil {
		return nil, err
	}

	if order.PaymentStatus != 0 { // Not pending
		return nil, fmt.Errorf("order payment already processed")
	}

	// Calculate amount in paise
	grandTotal := order.Total.Add(order.Tax).Add(order.Shipping)
	amountPaise, _ := grandTotal.Mul(decimal.NewFromInt(100)).Float64()

	// Create payment intent with Razorpay
	receipt := fmt.Sprintf("order_%d", orderID)
	razorpayOrder, err := s.razorpayClient.CreateOrder(int(amountPaise), receipt)
	if err != nil {
		return nil, err
	}

	// Save payment record
	payment := &models.Payment{
		OrderID:  orderID,
		IntentID: razorpayOrder.ID,
		Provider: "razorpay",
		Amount:   grandTotal,
		Status:   0, // Pending
	}

	err = s.db.Create(payment).Error
	if err != nil {
		return nil, err
	}

	return &dto.PaymentIntentResponse{
		PaymentID:       payment.ID,
		RazorpayOrderID: razorpayOrder.ID,
		Amount:          grandTotal,
		Currency:        "INR",
	}, nil
}

func (s *PaymentService) CapturePayment(req dto.CapturePaymentRequest) error {
	// Get payment record
	var payment models.Payment
	err := s.db.Where("intent_id = ?", req.RazorpayOrderID).First(&payment).Error
	if err != nil {
		return err
	}

	// Verify payment signature
	isValid := s.razorpayClient.VerifyPayment(
		req.RazorpayPaymentID,
		req.RazorpayOrderID,
		req.RazorpaySignature,
	)

	if !isValid {
		// Mark payment as failed
		payment.Status = 2
		s.db.Save(&payment)
		return fmt.Errorf("payment verification failed")
	}

	// Update payment status
	payment.Status = 1 // Captured
	payment.TxnRef = req.RazorpayPaymentID

	tx := s.db.Begin()
	err = tx.Save(&payment).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Update order payment status
	err = tx.Model(&models.Order{}).
		Where("id = ?", payment.OrderID).
		Updates(map[string]interface{}{
			"payment_status": 1, // Captured
			"status":         1, // Confirmed
		}).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
