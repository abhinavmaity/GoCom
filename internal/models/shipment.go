package models

import "time"

type Shipment struct {
	ID        uint   `gorm:"primaryKey"`
	OrderID   uint   `gorm:"not null"`
	Provider  string // shiprocket, etc.
	AWB       string
	Status    int `gorm:"default:0"` // 0=created, 1=shipped, 2=delivered
	ETA       *time.Time
	CreatedAt time.Time
}
