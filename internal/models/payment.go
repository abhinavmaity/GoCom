package models

import (
	"time"

	"google.golang.org/genproto/googleapis/type/decimal"
)

type Payment struct {
	ID        uint `gorm:"primaryKey"`
	OrderID   uint `gorm:"not null"`
	IntentID  string
	Provider  string          // razorpay, payu, etc.
	Amount    decimal.Decimal `gorm:"type:decimal(10,2)"`
	Status    int             `gorm:"default:0"` // 0=pending, 1=captured, 2=failed
	TxnRef    string
	CreatedAt time.Time
}
