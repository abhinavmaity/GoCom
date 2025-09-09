package models

import (
    "time"
)

type Inventory struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    SKUID      uint      `gorm:"not null" json:"sku_id"`
    LocationID uint      `gorm:"default:1" json:"location_id"`
    OnHand     int       `gorm:"default:0" json:"on_hand"`
    Reserved   int       `gorm:"default:0" json:"reserved"`
    Threshold  int       `gorm:"default:10" json:"threshold"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
    
    // Relations
    SKU SKU `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
}
