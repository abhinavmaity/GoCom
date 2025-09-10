package services

import (
	"errors"
	"gocom/main/internal/models"
	"gorm.io/gorm"
	//"time"
)

type InventoryService struct {
	db *gorm.DB
}

func NewInventoryService(db *gorm.DB) *InventoryService {
	return &InventoryService{db: db}
}

// ReserveInventory checks if enough inventory is available and reserves it
func (s *InventoryService) ReserveInventory(skuID uint, qty int) error {
	// Find the inventory for the SKU
	var inventory models.Inventory
	if err := s.db.Where("sku_id = ?", skuID).First(&inventory).Error; err != nil {
		return errors.New("inventory not found")
	}

	// Check if enough stock is available
	if inventory.OnHand-inventory.Reserved < qty {
		return errors.New("insufficient stock available")
	}

	// Reserve the inventory
	inventory.Reserved += qty
	if err := s.db.Save(&inventory).Error; err != nil {
		return err
	}

	return nil
}

// ReleaseInventory releases reserved inventory if payment fails or order is canceled
func (s *InventoryService) ReleaseInventory(skuID uint, qty int) error {
	// Find the inventory for the SKU
	var inventory models.Inventory
	if err := s.db.Where("sku_id = ?", skuID).First(&inventory).Error; err != nil {
		return errors.New("inventory not found")
	}

	// Release the reserved inventory
	inventory.Reserved -= qty
	if inventory.Reserved < 0 {
		inventory.Reserved = 0
	}

	if err := s.db.Save(&inventory).Error; err != nil {
		return err
	}

	return nil
}

// CheckInventory ensures there is enough stock available
func (s *InventoryService) CheckInventory(skuID uint, qty int) (bool, error) {
	// Find the inventory for the SKU
	var inventory models.Inventory
	if err := s.db.Where("sku_id = ?", skuID).First(&inventory).Error; err != nil {
		return false, errors.New("inventory not found")
	}

	// Check if there is enough inventory
	if inventory.OnHand-inventory.Reserved >= qty {
		return true, nil
	}
	return false, nil
}
