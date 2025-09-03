package models

import "time"

type Inventory struct {
	ID         uint `gorm:"primaryKey"`
	SKUID      uint `gorm:"not null"`
	LocationID uint `gorm:"not null"`
	OnHand     int  `gorm:"default:0"`
	Reserved   int  `gorm:"default:0"`
	Threshold  int  `gorm:"default:0"`
	UpdatedAt  time.Time
}
