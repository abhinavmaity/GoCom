// internal/marketplace/commerce/services/checkout_service.go
package services

import (
	"errors"
	"github.com/shopspring/decimal"
	"gocom/main/internal/marketplace/commerce/dto"
	"gocom/main/internal/models"
	"gorm.io/gorm"
)

type CheckoutService struct {
	db               *gorm.DB
	inventoryService *InventoryService
}

func NewCheckoutService(db *gorm.DB) *CheckoutService {
	return &CheckoutService{
		db:               db,
		inventoryService: NewInventoryService(db),
	}
}

func (s *CheckoutService) ValidateAndReserve(cartID, userID, addressID uint) (*dto.CheckoutResponse, error) {
	// Get cart with items
	var cart models.Cart
	err := s.db.Where("id = ? AND user_id = ?", cartID, userID).
		Preload("Items").
		Preload("Items.SKU").
		Preload("Items.SKU.Product").
		First(&cart).Error
	if err != nil {
		return nil, errors.New("cart not found")
	}

	if len(cart.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	// Validate address belongs to user
	var address models.Address
	err = s.db.Where("id = ? AND user_id = ?", addressID, userID).First(&address).Error
	if err != nil {
		return nil, errors.New("address not found")
	}

	// Check inventory availability
	for _, item := range cart.Items {
		err = s.inventoryService.CheckAvailability(item.SKUID, item.Qty)
		if err != nil {
			return nil, err
		}
	}

	// Calculate totals
	var total, tax decimal.Decimal
	var orderItems []dto.OrderItemDto

	for _, item := range cart.Items {
		itemTotal := item.Price.Mul(decimal.NewFromInt(int64(item.Qty)))
		itemTax := itemTotal.Mul(item.SKU.TaxPct).Div(decimal.NewFromInt(100))

		total = total.Add(itemTotal)
		tax = tax.Add(itemTax)

		orderItems = append(orderItems, dto.OrderItemDto{
			SKUID:    item.SKUID,
			SKUCode:  item.SKU.SKUCode,
			Title:    item.SKU.Product.Title,
			Qty:      item.Qty,
			Price:    item.Price,
			Tax:      itemTax,
			SellerID: item.SKU.Product.SellerID,
		})
	}

	shipping := decimal.Zero // Free shipping for PoC
	grandTotal := total.Add(tax).Add(shipping)

	// Create order
	order := &models.Order{
		UserID:        userID,
		Total:         total,
		Tax:           tax,
		Shipping:      shipping,
		Status:        0, // New
		PaymentStatus: 0, // Pending
		AddressID:     addressID,
	}

	tx := s.db.Begin()
	err = tx.Create(order).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create order items and reserve inventory
	for _, cartItem := range cart.Items {
		orderItem := &models.OrderItem{
			OrderID:  order.ID,
			SKUID:    cartItem.SKUID,
			Qty:      cartItem.Qty,
			Price:    cartItem.Price,
			Tax:      cartItem.Price.Mul(cartItem.SKU.TaxPct).Div(decimal.NewFromInt(100)),
			SellerID: cartItem.SKU.Product.SellerID,
		}

		err = tx.Create(orderItem).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		// Reserve inventory
		err = tx.Model(&models.Inventory{}).
			Where("sku_id = ?", cartItem.SKUID).
			Update("reserved", gorm.Expr("reserved + ?", cartItem.Qty)).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &dto.CheckoutResponse{
		OrderID:    order.ID,
		Total:      total,
		Tax:        tax,
		Shipping:   shipping,
		GrandTotal: grandTotal,
		Items:      orderItems,
	}, nil
}
