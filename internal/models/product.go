package models

import "time"

type Product struct {
	ID          uint   `gorm:"primaryKey"`
	SellerID    uint   `gorm:"not null"`
	CategoryID  uint   `gorm:"not null"`
	Title       string `gorm:"not null"`
	Description string
	Brand       string
	Status      int `gorm:"default:0"` // 0=draft, 1=published
	Score       int `gorm:"default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
