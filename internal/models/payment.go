package models

import (
    "time"
    "github.com/shopspring/decimal"
)

type Payment struct {
    ID          uint            `gorm:"primaryKey" json:"id"`
    OrderID     uint            `gorm:"not null" json:"order_id"`
    Amount      decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"amount"`
    Method      string          `gorm:"size:50;not null" json:"method"` // card, upi, wallet, etc.
    Status      int             `gorm:"default:0" json:"status"` // 0=pending, 1=captured, 2=failed
    GatewayRef  string          `gorm:"size:100" json:"gateway_ref"`
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
    
    Order Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

const (
    PaymentStatusPending = iota
    PaymentStatusCaptured  
    PaymentStatusFailed
)
