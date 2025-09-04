// internal/marketplace/commerce/services/order_service.go
package services

import (
	//"github.com/shopspring/decimal"
	"gocom/main/internal/marketplace/commerce/dto"
	"gocom/main/internal/models"
	"gorm.io/gorm"
)

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) GetUserOrders(userID uint, page, limit int) ([]dto.OrderSummary, error) {
	var orders []models.Order
	offset := (page - 1) * limit

	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	var result []dto.OrderSummary
	for _, order := range orders {
		result = append(result, dto.OrderSummary{
			ID:            order.ID,
			Total:         order.Total.Add(order.Tax).Add(order.Shipping),
			Status:        s.getStatusText(order.Status),
			PaymentStatus: s.getPaymentStatusText(order.PaymentStatus),
			CreatedAt:     order.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return result, nil
}

func (s *OrderService) GetOrderDetails(orderID, userID uint) (*dto.OrderDetails, error) {
	var order models.Order
	err := s.db.Where("id = ? AND user_id = ?", orderID, userID).
		Preload("OrderItems").
		Preload("OrderItems.SKU").
		Preload("OrderItems.SKU.Product").
		Preload("Address").
		First(&order).Error

	if err != nil {
		return nil, err
	}

	var items []dto.OrderItemDto
	for _, item := range order.OrderItems {
		items = append(items, dto.OrderItemDto{
			SKUID:    item.SKUID,
			SKUCode:  item.SKU.SKUCode,
			Title:    item.SKU.Product.Title,
			Qty:      item.Qty,
			Price:    item.Price,
			Tax:      item.Tax,
			SellerID: item.SellerID,
		})
	}

	return &dto.OrderDetails{
		ID:            order.ID,
		Total:         order.Total,
		Tax:           order.Tax,
		Shipping:      order.Shipping,
		GrandTotal:    order.Total.Add(order.Tax).Add(order.Shipping),
		Status:        s.getStatusText(order.Status),
		PaymentStatus: s.getPaymentStatusText(order.PaymentStatus),
		Address:       s.mapAddressToDto(order.Address),
		Items:         items,
		CreatedAt:     order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (s *OrderService) getStatusText(status int) string {
	statusMap := map[int]string{
		0: "New",
		1: "Confirmed",
		2: "Shipped",
		3: "Delivered",
		4: "Cancelled",
		5: "Returned",
	}
	return statusMap[status]
}

func (s *OrderService) getPaymentStatusText(status int) string {
	statusMap := map[int]string{
		0: "Pending",
		1: "Captured",
		2: "Failed",
		3: "Refunded",
	}
	return statusMap[status]
}

func (s *OrderService) mapAddressToDto(address models.Address) dto.AddressDto {
	return dto.AddressDto{
		ID:      address.ID,
		Line1:   address.Line1,
		Line2:   address.Line2,
		City:    address.City,
		State:   address.State,
		Country: address.Country,
		Pin:     address.Pin,
	}
}
