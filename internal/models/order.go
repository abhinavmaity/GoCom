package models

import "time"

type Order struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	AddressID uint      `gorm:"not null"`
	Total     float64   `gorm:"not null"`
	Status    string    `gorm:"default:'pending'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	PaymentID uint
}

type OrderItem struct {
	ID        uint    `gorm:"primaryKey"`
	OrderID   uint    `gorm:"not null"`
	SKUID     uint    `gorm:"not null"`
	Qty       int     `gorm:"not null"`
	Price     float64 `gorm:"not null"`
	Total     float64 `gorm:"not null"`
	Tax       float64 `gorm:"default:0.0"`
	SellerID  uint    `gorm:"not null"`
	CreatedAt time.Time
}
