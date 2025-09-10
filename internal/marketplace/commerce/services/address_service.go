package services

import (
	"gocom/main/internal/models"
	"gorm.io/gorm"
	//"errors"
)

type AddressService struct {
	db *gorm.DB
}

func NewAddressService(db *gorm.DB) *AddressService {
	return &AddressService{db: db}
}

// CreateAddress adds a new address for the user
func (s *AddressService) CreateAddress(userID uint, address models.Address) (*models.Address, error) {
	address.UserID = &userID
	if err := s.db.Create(&address).Error; err != nil {
		return nil, err
	}
	return &address, nil
}

// GetAddresses retrieves all addresses for a user
func (s *AddressService) GetAddresses(userID uint) ([]models.Address, error) {
	var addresses []models.Address
	if err := s.db.Where("user_id = ?", userID).Find(&addresses).Error; err != nil {
		return nil, err
	}
	return addresses, nil
}
