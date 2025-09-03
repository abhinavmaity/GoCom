package models

import "time"

type Seller struct {
	ID          uint   `gorm:"primaryKey"`
	LegalName   string `gorm:"not null"`
	DisplayName string
	GSTIN       string
	PAN         string
	BankRef     string
	Status      int `gorm:"default:0"` // 0=pending, 1=approved, 2=rejected
	RiskScore   int `gorm:"default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SellerUser struct {
	ID       uint `gorm:"primaryKey"`
	SellerID uint `gorm:"not null"`
	UserID   uint `gorm:"not null"`
	Role     string
	Status   int `gorm:"default:1"`
}
