package models

import (
    "time"
    "github.com/shopspring/decimal"
)

type Order struct {
    ID            uint            `gorm:"primaryKey" json:"id"`
    UserID        uint            `gorm:"not null" json:"user_id"`
    Total         decimal.Decimal `gorm:"type:decimal(10,2)" json:"total"`
    Tax           decimal.Decimal `gorm:"type:decimal(10,2)" json:"tax"`
    Shipping      decimal.Decimal `gorm:"type:decimal(10,2)" json:"shipping"`
    Status        int             `gorm:"default:0" json:"status"`
    PaymentStatus int             `gorm:"default:0" json:"payment_status"`
    AddressID     uint            `json:"address_id"`
    CreatedAt     time.Time       `json:"created_at"`
    UpdatedAt     time.Time       `json:"updated_at"`
    Items []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

type OrderItem struct {
    ID       uint            `gorm:"primaryKey" json:"id"`
    OrderID  uint            `gorm:"not null" json:"order_id"`
    SKUID    uint            `gorm:"not null" json:"sku_id"`
    SellerID uint            `gorm:"not null" json:"seller_id"`
    Qty      int             `gorm:"not null" json:"qty"`
    Price    decimal.Decimal `gorm:"type:decimal(10,2)" json:"price"`
    Tax      decimal.Decimal `gorm:"type:decimal(10,2)" json:"tax"`
    SKU SKU `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
}