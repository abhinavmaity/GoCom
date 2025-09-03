package models

import (
	"time"

	"google.golang.org/genproto/googleapis/type/decimal"
)

type Order struct {
	ID            uint            `gorm:"primaryKey"`
	UserID        uint            `gorm:"not null"`
	Total         decimal.Decimal `gorm:"type:decimal(10,2)"`
	Tax           decimal.Decimal `gorm:"type:decimal(10,2)"`
	Shipping      decimal.Decimal `gorm:"type:decimal(10,2)"`
	Status        int             `gorm:"default:0"` // 0=new, 1=confirmed, 2=shipped, et
	PaymentStatus int             `gorm:"default:0"` // 0=pending, 1=captured, 2=failed
	AddressID     uint            `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type OrderItem struct {
	ID         uint            `gorm:"primaryKey"`
	OrderID    uint            `gorm:"not null"`
	SKUID      uint            `gorm:"not null"`
	Qty        int             `gorm:"not null"`
	Price      decimal.Decimal `gorm:"type:decimal(10,2)"`
	Tax        decimal.Decimal `gorm:"type:decimal(10,2)"`
	SellerID   uint            `gorm:"not null"`
	ShipmentID *uint
}
