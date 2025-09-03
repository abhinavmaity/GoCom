package models

import (
	"time"

	"google.golang.org/genproto/googleapis/type/decimal"
)

type Return struct {
	ID          uint `gorm:"primaryKey"`
	OrderID     uint `gorm:"not null"`
	OrderItemID uint `gorm:"not null"`
	Reason      string
	Status      int `gorm:"default:0"` // 0=requested, 1=approved, 2=rejected
	RefundID    *uint
	CreatedAt   time.Time
}

type Refund struct {
	ID          uint            `gorm:"primaryKey"`
	PaymentID   uint            `gorm:"not null"`
	Amount      decimal.Decimal `gorm:"type:decimal(10,2)"`
	Status      int             `gorm:"default:0"` // 0=pending, 1=processed
	ProcessedAt *time.Time
	CreatedAt   time.Time
}
