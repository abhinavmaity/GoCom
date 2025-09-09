package models

import (
	"time"

	"google.golang.org/genproto/googleapis/type/decimal"
)

type Return struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	OrderID     uint   `gorm:"not null" json:"order_id"`
	OrderItemID uint   `gorm:"not null" json:"order_item_id"`
	Reason      string `json:"reason"`
	Status      int    `gorm:"default:0" json:"status"` // 0=requested, 1=approved
	RefundID    *uint  `json:"refund_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	
	Order     Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	OrderItem OrderItem `gorm:"foreignKey:OrderItemID" json:"order_item,omitempty"`
	Refund    *Refund   `gorm:"foreignKey:RefundID" json:"refund,omitempty"`
}

type Refund struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	PaymentID   uint            `gorm:"not null" json:"payment_id"`
	Amount      decimal.Decimal `gorm:"type:decimal(10,2)" json:"amount"`
	Status      int             `gorm:"default:0" json:"status"` // 0=pending, 1=processed
	ProcessedAt *time.Time      `json:"processed_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	
	Payment Payment `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`
}