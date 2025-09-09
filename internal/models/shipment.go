package models

import "time"

type Shipment struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	OrderID   uint       `gorm:"not null" json:"order_id"`
	Provider  string     `json:"provider"` // internal, delhivery, etc.
	AWB       string     `json:"awb"`
	Status    int        `gorm:"default:0" json:"status"` // 0=created, 1=shipped, 2=delivered
	ETA       *time.Time `json:"eta,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	
	// Relations
	Order Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}