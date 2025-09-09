package services

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gocom/main/internal/common/db"
	"gocom/main/internal/models"
)

type AddressService struct {
	DB *gorm.DB
}

func NewAddressService() *AddressService {
	return &AddressService{DB: db.GetDB()}
}


func (as *AddressService) AddAddress(sellerID uint, req *AddAddressRequest) (*models.Address, error) {
	var seller models.Seller
	if err := as.DB.First(&seller, sellerID).Error; err != nil {
		return nil, errors.New("seller not found")
	}

	address := &models.Address{
		SellerID:  &sellerID,
		Line1:     req.Line1,
		Line2:     req.Line2,
		City:      req.City,
		State:     req.State,
		Country:   req.Country,
		Pin:       req.Pin,
		CreatedAt: time.Now(),
	}

	if err := as.DB.Create(address).Error; err != nil {
		return nil, err
	}

	return address, nil
}

func (as *AddressService) GetSellerAddresses(sellerID uint) ([]AddressResponse, error) {
	var addresses []models.Address
	if err := as.DB.Where("seller_id = ?", sellerID).
		Order("created_at DESC").Find(&addresses).Error; err != nil {
		return nil, err
	}

	var response []AddressResponse
	for _, addr := range addresses {
		response = append(response, AddressResponse{
			ID:        addr.ID,
			Line1:     addr.Line1,
			Line2:     addr.Line2,
			City:      addr.City,
			State:     addr.State,
			Country:   addr.Country,
			Pin:       addr.Pin,
			CreatedAt: addr.CreatedAt,
		})
	}

	return response, nil
}


func (as *AddressService) UpdateAddress(addressID, sellerID uint, req *UpdateAddressRequest) (*models.Address, error) {
	var address models.Address
	if err := as.DB.Where("id = ? AND seller_id = ?", addressID, sellerID).
		First(&address).Error; err != nil {
		return nil, errors.New("address not found")
	}

	if req.Line1 != nil {
		address.Line1 = *req.Line1
	}
	if req.Line2 != nil {
		address.Line2 = *req.Line2
	}
	if req.City != nil {
		address.City = *req.City
	}
	if req.State != nil {
		address.State = *req.State
	}
	if req.Country != nil {
		address.Country = *req.Country
	}
	if req.Pin != nil {
		address.Pin = *req.Pin
	}

	if err := as.DB.Save(&address).Error; err != nil {
		return nil, err
	}

	return &address, nil
}

func (as *AddressService) DeleteAddress(addressID, sellerID uint) error {
	result := as.DB.Where("id = ? AND seller_id = ?", addressID, sellerID).
		Delete(&models.Address{})
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("address not found")
	}

	return nil
}


type AddAddressRequest struct {
	Line1   string `json:"line1" binding:"required"`
	Line2   string `json:"line2"`
	City    string `json:"city" binding:"required"`
	State   string `json:"state" binding:"required"`
	Country string `json:"country" binding:"required"`
	Pin     string `json:"pin" binding:"required,len=6"`
}

type UpdateAddressRequest struct {
	Line1   *string `json:"line1,omitempty"`
	Line2   *string `json:"line2,omitempty"`
	City    *string `json:"city,omitempty"`
	State   *string `json:"state,omitempty"`
	Country *string `json:"country,omitempty"`
	Pin     *string `json:"pin,omitempty"`
}

type AddressResponse struct {
	ID        uint      `json:"id"`
	Line1     string    `json:"line1"`
	Line2     string    `json:"line2"`
	City      string    `json:"city"`
	State     string    `json:"state"`
	Country   string    `json:"country"`
	Pin       string    `json:"pin"`
	CreatedAt time.Time `json:"created_at"`
}

