package models

import "time"

type Address struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    *uint  // for marketplace users
	SellerID  *uint  // for sellers
	Line1     string `gorm:"not null"`
	Line2     string
	City      string `gorm:"not null"`
	State     string `gorm:"not null"`
	Country   string `gorm:"not null"`
	Pin       string `gorm:"not null"`
	CreatedAt time.Time
}
