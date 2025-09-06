package models

import "time"

type KYC struct {
	ID          uint   `gorm:"primaryKey"`
	SellerID    uint   `gorm:"not null"`
	Type        string // PAN, GSTIN, etc.
	DocumentURL string
	Status      int `gorm:"default:0"` // 0=pending, 1=approved, 2=rejected
	Remarks     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Seller      Seller `gorm:"foreignKey:SellerID"`
}
