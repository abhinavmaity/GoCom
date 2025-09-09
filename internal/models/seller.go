package models

import (
	"time"
)

type Seller struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	LegalName   string    `gorm:"not null" json:"legal_name"`
	DisplayName string    `json:"display_name"`
	GSTIN       string    `json:"gstin"`
	PAN         string    `gorm:"not null" json:"pan"`
	BankRef     string    `json:"bank_ref"`
	Status      int       `gorm:"default:0" json:"status"` // 0=pending, 1=approved, 2=rejected
	RiskScore   int       `gorm:"default:0" json:"risk_score"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Users    []SellerUser `gorm:"foreignKey:SellerID" json:"users,omitempty"`
	KYC      []KYC        `gorm:"foreignKey:SellerID" json:"kyc,omitempty"`
	Products []Product    `gorm:"foreignKey:SellerID" json:"products,omitempty"`
}

type SellerUser struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	SellerID uint `gorm:"not null" json:"seller_id"`
	UserID   uint `gorm:"not null" json:"user_id"`
	Role     string `json:"role"` // owner, manager, staff
	Status   int `gorm:"default:1" json:"status"` // 1=active, 0=inactive

	Seller Seller `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}


const (
	SellerStatusPending = iota
	SellerStatusApproved
	SellerStatusRejected
)

