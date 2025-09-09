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
	Status        int             `gorm:"default:0" json:"status"` // 0=new, 1=confirmed, 2=shipped, 3=delivered
	PaymentStatus int             `gorm:"default:0" json:"payment_status"` // 0=pending, 1=captured
	AddressID     uint            `gorm:"not null" json:"address_id"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	
	// Relations
	Items     []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	User      User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Address   Address     `gorm:"foreignKey:AddressID" json:"address,omitempty"`
	Payments  []Payment   `gorm:"foreignKey:OrderID" json:"payments,omitempty"`
	Shipments []Shipment  `gorm:"foreignKey:OrderID" json:"shipments,omitempty"`
}

type OrderItem struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	OrderID    uint            `gorm:"not null" json:"order_id"`
	SKUID      uint            `gorm:"not null" json:"sku_id"`
	Qty        int             `gorm:"not null" json:"quantity"`
	Price      decimal.Decimal `gorm:"type:decimal(10,2)" json:"price"`
	Tax        decimal.Decimal `gorm:"type:decimal(10,2)" json:"tax"`
	SellerID   uint            `gorm:"not null" json:"seller_id"`
	ShipmentID *uint           `json:"shipment_id,omitempty"`
	
	// Relations
	Order    Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	SKU      SKU       `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
	Seller   Seller    `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
	Shipment *Shipment `gorm:"foreignKey:ShipmentID" json:"shipment,omitempty"`
}
