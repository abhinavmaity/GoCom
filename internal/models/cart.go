package models

import (
	"github.com/shopspring/decimal"
	"time"
	//"google.golang.org/genproto/googleapis/type/decimal"
)

type Cart struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null"`
	Currency  string `gorm:"default:INR"`
	CreatedAt time.Time
}

type CartItem struct {
	ID        uint            `gorm:"primaryKey"`
	CartID    uint            `gorm:"not null"`
	SKUID     uint            `gorm:"not null"`
	Qty       int             `gorm:"not null"`
	Price     decimal.Decimal `gorm:"type:decimal(10,2)"`
	CreatedAt time.Time
}
