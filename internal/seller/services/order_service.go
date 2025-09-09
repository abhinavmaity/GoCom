package services

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


// Add wrapper method for handler compatibility
func (os *OrderService) GetSellerOrders(sellerID uint) ([]OrderResponse, error) {
    // Default pagination values
    return os.GetSellerOrdersPaginated(sellerID, 1, 50)
}

// Rename existing method
func (os *OrderService) GetSellerOrdersPaginated(sellerID uint, page, limit int) ([]OrderResponse, error) {
    // Your existing implementation...
    var orders []models.Order
    offset := (page - 1) * limit
    
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

    // Convert to response format...
    var response []OrderResponse
    for _, order := range orders {
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

// Fix ShipOrder method signature
func (os *OrderService) ShipOrder(orderID, sellerID uint, provider, awb string) error {
    // Verify seller owns this order
    var count int64
    os.DB.Model(&models.OrderItem{}).
        Where("order_id = ? AND seller_id = ?", orderID, sellerID).
        Count(&count)
        
    if count == 0 {
        return errors.New("order not found or unauthorized")
    }

    // Create shipment record
    shipment := &models.Shipment{
        OrderID:   orderID,
        Provider:  provider,
        AWB:       awb,
        Status:    1, // Shipped
        CreatedAt: time.Now(),
    }

    if err := os.DB.Create(shipment).Error; err != nil {
        return err
    }

    // Update order status to shipped
    return os.DB.Model(&models.Order{}).
        Where("id = ?", orderID).
        Updates(map[string]interface{}{
            "status":     2, // Shipped
            "updated_at": time.Now(),
        }).Error
}

// ✅ FIXED: Correct method signature
func (os *OrderService) GetOrderDetails(orderID, sellerID uint) (*OrderDetailResponse, error) {
	var order models.Order
	
	// Check if seller has items in this order
	var count int64
	os.DB.Model(&models.OrderItem{}).
		Where("order_id = ? AND seller_id = ?", orderID, sellerID).
		Count(&count)
	
	if count == 0 {
		return nil, errors.New("order not found or unauthorized")
	}

	if err := os.DB.First(&order, orderID).Error; err != nil {
		return nil, errors.New("order not found")
	}

	// Get order items for this seller
	var items []models.OrderItem
	err := os.DB.Preload("SKU").Preload("SKU.Product").
		Where("order_id = ? AND seller_id = ?", orderID, sellerID).
		Find(&items).Error
	
	if err != nil {
		return nil, err
	}

	var itemResponses []OrderItemResponse
	for _, item := range items {
		itemResponses = append(itemResponses, OrderItemResponse{
			ID:           item.ID,
			SKUCode:      item.SKU.SKUCode,
			ProductTitle: item.SKU.Product.Title,
			Qty:          item.Qty,
			Price:        item.Price,
			Tax:          item.Tax,
		})
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
		Items:         itemResponses,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}

	return response, nil
}



func (os *OrderService) UpdateOrderStatus(orderID, sellerID uint, newStatus int) error {
	var count int64
	os.DB.Model(&models.OrderItem{}).
		Where("order_id = ? AND seller_id = ?", orderID, sellerID).
		Count(&count)
	
	if count == 0 {
		return errors.New("order not found or unauthorized")
	}

	return os.DB.Model(&models.Order{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"status":     newStatus,
			"updated_at": time.Now(),
		}).Error
}

// Response DTOs
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
	ID            uint                `json:"id"`
	UserID        uint                `json:"user_id"`
	Total         decimal.Decimal     `json:"total"`
	Tax           decimal.Decimal     `json:"tax"`
	Shipping      decimal.Decimal     `json:"shipping"`
	Status        int                 `json:"status"`
	StatusText    string              `json:"status_text"`
	PaymentStatus int                 `json:"payment_status"`
	AddressID     uint                `json:"address_id"`
	Items         []OrderItemResponse `json:"items"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

type OrderItemResponse struct {
	ID           uint            `json:"id"`
	SKUCode      string          `json:"sku_code"`
	ProductTitle string          `json:"product_title"`
	Qty          int             `json:"quantity"`
	Price        decimal.Decimal `json:"price"`
	Tax          decimal.Decimal `json:"tax"`
}

func GetStatusText(status int) string {
	switch status {
	case 0:
		return "New"
	case 1:
		return "Confirmed"
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
