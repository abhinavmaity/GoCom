// internal/marketplace/commerce/services/inventory_service.go
package services

import (
	"errors"
	"gocom/main/internal/models"
	"gorm.io/gorm"
)

type InventoryService struct {
	db *gorm.DB
}

func NewInventoryService(db *gorm.DB) *InventoryService {
	return &InventoryService{db: db}
}

func (s *InventoryService) CheckAvailability(skuID uint, qty int) error {
	var inventory models.Inventory
	err := s.db.Where("sku_id = ?", skuID).First(&inventory).Error
	if err != nil {
		return errors.New("inventory not found")
	}

	available := inventory.OnHand - inventory.Reserved
	if available < qty {
		return errors.New("insufficient stock")
	}

	return nil
}

func (s *InventoryService) ReserveStock(skuID uint, qty int) error {
	return s.db.Model(&models.Inventory{}).
		Where("sku_id = ? AND (on_hand - reserved) >= ?", skuID, qty).
		Update("reserved", gorm.Expr("reserved + ?", qty)).Error
}

func (s *InventoryService) ReleaseReservation(skuID uint, qty int) error {
	return s.db.Model(&models.Inventory{}).
		Where("sku_id = ?", skuID).
		Update("reserved", gorm.Expr("reserved - ?", qty)).Error
}
