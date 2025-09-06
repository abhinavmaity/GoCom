package services

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gocom/main/internal/common/db"
	"gocom/main/internal/models"
	"github.com/shopspring/decimal"
)

type OrderService struct {
	DB *gorm.DB
}

func NewOrderService() *OrderService {
	return &OrderService{DB: db.GetDB()}
}

// Get seller orders (simplified)
func (os *OrderService) GetSellerOrders(sellerID uint) ([]OrderSummary, error) {
	var orders []OrderSummary
	
	query := `
		SELECT DISTINCT o.id, o.total, o.status, o.payment_status, o.created_at,
		       u.name as customer_name
		FROM orders o
		JOIN order_items oi ON o.id = oi.order_id  
		JOIN users u ON o.user_id = u.id
		WHERE oi.seller_id = ?
		ORDER BY o.created_at DESC`

	if err := os.DB.Raw(query, sellerID).Scan(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

// Get order details
func (os *OrderService) GetOrderDetails(orderID, sellerID uint) (*OrderDetails, error) {
	// Verify seller access
	var count int64
	os.DB.Model(&models.OrderItem{}).
		Where("order_id = ? AND seller_id = ?", orderID, sellerID).
		Count(&count)
	
	if count == 0 {
		return nil, errors.New("order not found")
	}

	var order models.Order
	if err := os.DB.Preload("Address").First(&order, orderID).Error; err != nil {
		return nil, err
	}

	var customer models.User
	os.DB.First(&customer, order.UserID)

	return &OrderDetails{
		ID:           order.ID,
		CustomerName: customer.Name,
		Total:        order.Total,
		Status:       order.Status,
		StatusText:   os.getStatusText(order.Status),
		CreatedAt:    order.CreatedAt,
	}, nil
}

// Update order status
func (os *OrderService) UpdateOrderStatus(orderID, sellerID uint, status int) error {
	// Verify seller access
	var count int64
	os.DB.Model(&models.OrderItem{}).
		Where("order_id = ? AND seller_id = ?", orderID, sellerID).
		Count(&count)
	
	if count == 0 {
		return errors.New("order not found")
	}

	return os.DB.Model(&models.Order{}).
		Where("id = ?", orderID).
		Update("status", status).Error
}

// Ship order
func (os *OrderService) ShipOrder(orderID, sellerID uint, provider, awb string) error {
	tx := os.DB.Begin()

	// Create shipment
	shipment := &models.Shipment{
		OrderID:  orderID,
		Provider: provider,
		AWB:      awb,
		Status:   1,
	}
	
	if err := tx.Create(shipment).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update order status to shipped
	if err := tx.Model(&models.Order{}).Where("id = ?", orderID).
		Update("status", 2).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (os *OrderService) getStatusText(status int) string {
	statusMap := map[int]string{0: "New", 1: "Confirmed", 2: "Shipped", 3: "Delivered"}
	if text, exists := statusMap[status]; exists {
		return text
	}
	return "Unknown"
}

// Simple DTOs
type OrderSummary struct {
	ID           uint            `json:"id"`
	CustomerName string          `json:"customer_name"`
	Total        decimal.Decimal `json:"total"`
	Status       int             `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
}

type OrderDetails struct {
	ID           uint            `json:"id"`
	CustomerName string          `json:"customer_name"`
	Total        decimal.Decimal `json:"total"`
	Status       int             `json:"status"`
	StatusText   string          `json:"status_text"`
	CreatedAt    time.Time       `json:"created_at"`
}

