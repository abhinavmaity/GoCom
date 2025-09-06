apackage services

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gocom/main/internal/common/db"
	"gocom/main/internal/models"
)

type OrderService struct {
	DB *gorm.DB
}

func NewOrderService() *OrderService {
	return &OrderService{DB: db.GetDB()}
}

// Get all orders for seller using GORM
func (os *OrderService) GetSellerOrders(sellerID uint, page, limit int) ([]OrderResponse, error) {
	var orders []models.Order
	offset := (page - 1) * limit

	// Get orders where seller has items using GORM joins
	err := os.DB.
		Joins("JOIN order_items oi ON orders.id = oi.order_id").
		Where("oi.seller_id = ?", sellerID).
		Group("orders.id").
		Order("orders.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	// Convert to response format and get item counts
	var response []OrderResponse
	for _, order := range orders {
		// Get item count for this order and seller
		var itemCount int64
		os.DB.Model(&models.OrderItem{}).
			Where("order_id = ? AND seller_id = ?", order.ID, sellerID).
			Count(&itemCount)

		response = append(response, OrderResponse{
			ID:            order.ID,
			UserID:        order.UserID,
			Total:         order.Total,
			Tax:           order.Tax,
			Shipping:      order.Shipping,
			Status:        order.Status,
			StatusText:    GetStatusText(order.Status),
			PaymentStatus: order.PaymentStatus,
			ItemCount:     int(itemCount),
			CreatedAt:     order.CreatedAt,
			UpdatedAt:     order.UpdatedAt,
		})
	}

	return response, nil
}

// Get order details for seller using GORM
func (os *OrderService) GetOrderDetails(orderID, sellerID uint) (*OrderDetailResponse, error) {
	var order models.Order

	// First check if seller has items in this order
	var count int64
	os.DB.Model(&models.OrderItem{}).
		Where("order_id = ? AND seller_id = ?", orderID, sellerID).
		Count(&count)

	if count == 0 {
		return nil, errors.New("order not found or unauthorized")
	}

	// Get order details using GORM
	if err := os.DB.First(&order, orderID).Error; err != nil {
		return nil, errors.New("order not found")
	}

	// Get order items for this seller
	items, err := os.GetOrderItems(orderID, sellerID)
	if err != nil {
		return nil, err
	}

	response := &OrderDetailResponse{
		ID:            order.ID,
		UserID:        order.UserID,
		Total:         order.Total,
		Tax:           order.Tax,
		Shipping:      order.Shipping,
		Status:        order.Status,
		StatusText:    GetStatusText(order.Status),
		PaymentStatus: order.PaymentStatus,
		AddressID:     order.AddressID,
		Items:         items,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}

	return response, nil
}

// Get order items using GORM with proper joins
func (os *OrderService) GetOrderItems(orderID, sellerID uint) ([]OrderItemResponse, error) {
	var items []models.OrderItem

	// Get order items with product and SKU info using GORM
	err := os.DB.
		Preload("SKU").
		Preload("SKU.Product").
		Where("order_id = ? AND seller_id = ?", orderID, sellerID).
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	// Convert to response format
	var response []OrderItemResponse
	for _, item := range items {
		response = append(response, OrderItemResponse{
			ID:           item.ID,
			SKUCode:      item.SKU.SKUCode,
			ProductTitle: item.SKU.Product.Title,
			Qty:          item.Qty,
			Price:        item.Price,
			Tax:          item.Tax,
		})
	}

	return response, nil
}

// Update order status using GORM
func (os *OrderService) UpdateOrderStatus(orderID, sellerID uint, newStatus int) error {
	// Verify seller has items in this order
	var count int64
	os.DB.Model(&models.OrderItem{}).
		Where("order_id = ? AND seller_id = ?", orderID, sellerID).
		Count(&count)

	if count == 0 {
		return errors.New("order not found or unauthorized")
	}

	// Update order status using GORM
	return os.DB.Model(&models.Order{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"status":     newStatus,
			"updated_at": time.Now(),
		}).Error
}

// Ship order using GORM
func (os *OrderService) ShipOrder(orderID, sellerID uint, req *ShipOrderRequest) error {
	// Verify seller owns this order
	var count int64
	os.DB.Model(&models.OrderItem{}).
		Where("order_id = ? AND seller_id = ?", orderID, sellerID).
		Count(&count)

	if count == 0 {
		return errors.New("order not found or unauthorized")
	}

	// Create shipment record using GORM
	shipment := &models.Shipment{
		OrderID:   orderID,
		Provider:  req.Provider,
		AWB:       req.AWB,
		Status:    1, // Shipped
		CreatedAt: time.Now(),
	}

	if err := os.DB.Create(shipment).Error; err != nil {
		return err
	}

	// Update order status to shipped using GORM
	return os.DB.Model(&models.Order{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"status":     2, // Shipped
			"updated_at": time.Now(),
		}).Error
}

// Get order statistics using GORM aggregation
func (os *OrderService) GetOrderStats(sellerID uint) (*OrderStatsResponse, error) {
	var stats OrderStatsResponse

	// Total orders count using GORM
	os.DB.Model(&models.OrderItem{}).
		Select("COUNT(DISTINCT order_id)").
		Where("seller_id = ?", sellerID).
		Scan(&stats.TotalOrders)

	// Orders by status using GORM with joins
	type StatusCount struct {
		Status int `json:"status"`
		Count  int `json:"count"`
	}

	var statusCounts []StatusCount
	os.DB.Model(&models.Order{}).
		Select("orders.status, COUNT(DISTINCT orders.id) as count").
		Joins("JOIN order_items ON orders.id = order_items.order_id").
		Where("order_items.seller_id = ?", sellerID).
		Group("orders.status").
		Scan(&statusCounts)

	// Map status counts
	stats.NewOrders = 0
	stats.ProcessingOrders = 0
	stats.ShippedOrders = 0

	for _, sc := range statusCounts {
		switch sc.Status {
		case 0:
			stats.NewOrders = sc.Count
		case 1:
			stats.ProcessingOrders = sc.Count
		case 2:
			stats.ShippedOrders = sc.Count
		}
	}

	// Total revenue using GORM aggregation
	var totalRevenue decimal.Decimal
	os.DB.Model(&models.OrderItem{}).
		Select("COALESCE(SUM(price * qty), 0)").
		Where("seller_id = ?", sellerID).
		Scan(&totalRevenue)
	stats.TotalRevenue = totalRevenue

	return &stats, nil
}

// DTOs (same as before)
type OrderResponse struct {
	ID            uint            `json:"id"`
	UserID        uint            `json:"user_id"`
	Total         decimal.Decimal `json:"total"`
	Tax           decimal.Decimal `json:"tax"`
	Shipping      decimal.Decimal `json:"shipping"`
	Status        int             `json:"status"`
	StatusText    string          `json:"status_text"`
	PaymentStatus int             `json:"payment_status"`
	ItemCount     int             `json:"item_count"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type OrderDetailResponse struct {
	ID            uint                  `json:"id"`
	UserID        uint                  `json:"user_id"`
	Total         decimal.Decimal       `json:"total"`
	Tax           decimal.Decimal       `json:"tax"`
	Shipping      decimal.Decimal       `json:"shipping"`
	Status        int                   `json:"status"`
	StatusText    string                `json:"status_text"`
	PaymentStatus int                   `json:"payment_status"`
	AddressID     uint                  `json:"address_id"`
	Items         []OrderItemResponse   `json:"items"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

type OrderItemResponse struct {
	ID           uint            `json:"id"`
	SKUCode      string          `json:"sku_code"`
	ProductTitle string          `json:"product_title"`
	Qty          int             `json:"quantity"`
	Price        decimal.Decimal `json:"price"`
	Tax          decimal.Decimal `json:"tax"`
}

type ShipOrderRequest struct {
	Provider string `json:"provider" binding:"required"`
	AWB      string `json:"awb" binding:"required"`
}

type UpdateStatusRequest struct {
	Status int `json:"status" binding:"required"`
}

type OrderStatsResponse struct {
	TotalOrders       int             `json:"total_orders"`
	NewOrders         int             `json:"new_orders"`
	ProcessingOrders  int             `json:"processing_orders"`
	ShippedOrders     int             `json:"shipped_orders"`
	TotalRevenue      decimal.Decimal `json:"total_revenue"`
}

func GetStatusText(status int) string {
	switch status {
	case 0:
		return "New"
	case 1:
		return "Processing"
	case 2:
		return "Shipped"
	case 3:
		return "Delivered"
	case 4:
		return "Cancelled"
	case 5:
		return "Returned"
	default:
		return "Unknown"
	}
}

