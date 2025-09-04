// internal/marketplace/commerce/services/address_service.go
package services

import (
	"gocom/main/internal/marketplace/commerce/dto"
	"gocom/main/internal/models"
	"gorm.io/gorm"
)

type AddressService struct {
	db *gorm.DB
}

func NewAddressService(db *gorm.DB) *AddressService {
	return &AddressService{db: db}
}

func (s *AddressService) CreateAddress(userID uint, req dto.CreateAddressRequest) (*models.Address, error) {
	address := &models.Address{
		UserID:  &userID,
		Line1:   req.Line1,
		Line2:   req.Line2,
		City:    req.City,
		State:   req.State,
		Country: req.Country,
		Pin:     req.Pin,
	}

	err := s.db.Create(address).Error
	return address, err
}

func (s *AddressService) GetUserAddresses(userID uint) ([]models.Address, error) {
	var addresses []models.Address
	err := s.db.Where("user_id = ?", userID).Find(&addresses).Error
	return addresses, err
}
