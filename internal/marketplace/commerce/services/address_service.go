package services

import (
	"errors"
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
func (s *AddressService) UpdateAddress(userID uint, addressID string, address models.Address) (*models.Address, error) {
	var existingAddress models.Address
	if err := s.db.Where("id = ? AND user_id = ?", addressID, userID).First(&existingAddress).Error; err != nil {
		return nil, errors.New("address not found")
	}

	// Update fields
	existingAddress.Line1 = address.Line1
	existingAddress.Line2 = address.Line2
	existingAddress.City = address.City
	existingAddress.State = address.State
	existingAddress.Country = address.Country
	existingAddress.Pin = address.Pin

	if err := s.db.Save(&existingAddress).Error; err != nil {
		return nil, err
	}

	return &existingAddress, nil
}

func (s *AddressService) DeleteAddress(userID uint, addressID string) error {
	if err := s.db.Where("id = ? AND user_id = ?", addressID, userID).Delete(&models.Address{}).Error; err != nil {
		return errors.New("failed to delete address")
	}

	return nil
}
