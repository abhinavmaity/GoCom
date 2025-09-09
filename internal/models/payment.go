package models

import (
	"time"
	"github.com/shopspring/decimal"
)

type Payment struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	OrderID   uint            `gorm:"not null" json:"order_id"`
	IntentID  string          `json:"intent_id"`
	Provider  string          `json:"provider"` // razorpay, etc.
	Amount    decimal.Decimal `gorm:"type:decimal(10,2)" json:"amount"`
	Status    int             `gorm:"default:0" json:"status"` // 0=pending, 1=captured
	TxnRef    string          `json:"transaction_reference"`
	CreatedAt time.Time       `json:"created_at"`
	
	// Relations
	Order Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}