package services

import (
	"errors"
	"github.com/shopspring/decimal"
	"gocom/main/internal/models"
	"gorm.io/gorm"
)

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

// CreateOrderFromCart creates an order based on cart items and applies necessary data
func (s *OrderService) CreateOrderFromCart(cartID string, addressID uint) (*models.Order, error) {
	// Fetch cart from DB
	var cart models.Cart
	if err := s.db.Where("id = ?", cartID).First(&cart).Error; err != nil {
		return nil, errors.New("cart not found")
	}

	// Initialize total for order calculation
	var totalAmount decimal.Decimal

	// Create the order from the cart
	order := &models.Order{
		UserID:    cart.UserID,
		AddressID: addressID,
		Status:    "pending", // "pending" means waiting for payment capture
		Total:     0,         // Placeholder for now; we'll calculate the total
	}

	// Create order in the database
	if err := s.db.Create(order).Error; err != nil {
		return nil, err
	}

	// Create OrderItems from CartItems
	var cartItems []models.CartItem
	if err := s.db.Where("cart_id = ?", cartID).Find(&cartItems).Error; err != nil {
		return nil, err
	}

	for _, cartItem := range cartItems {
		// We need to convert cartItem.Price (decimal) to float64
		itemTotal := decimal.NewFromFloat(float64(cartItem.Qty)).Mul(cartItem.Price) // Total = Qty * Price

		// Save the OrderItem
		orderItem := models.OrderItem{
			OrderID:  order.ID,
			SKUID:    cartItem.SKUID,
			Qty:      cartItem.Qty,
			Price:    cartItem.Price.InexactFloat64(), // No conversion needed here; Price is already decimal
			Total:    itemTotal.InexactFloat64(),      // Convert decimal to float64 safely
			Tax:      0.0,                             // Placeholder; you can calculate tax if needed
			SellerID: 0,                               // Placeholder; link to seller if applicable
		}

		if err := s.db.Create(&orderItem).Error; err != nil {
			return nil, err
		}

		// Add to totalAmount (used for order total)
		totalAmount = totalAmount.Add(itemTotal)
	}

	// Update the order with the total amount
	order.Total = totalAmount.InexactFloat64() // Update the order total
	if err := s.db.Save(&order).Error; err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetOrdersByUser(userID uint) ([]models.Order, error) {
	var orders []models.Order
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *OrderService) GetOrderByID(userID uint, orderID string) (*models.Order, error) {
	var order models.Order
	if err := s.db.Preload("Items").Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return nil, errors.New("order not found")
	}

	return &order, nil
}
