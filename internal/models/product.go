package models

import (
    "time"
)

type Product struct {
    ID          uint            `gorm:"primaryKey" json:"id"`
    SellerID    uint            `gorm:"not null" json:"seller_id"`
    CategoryID  uint            `gorm:"not null" json:"category_id"`
    Title       string          `gorm:"not null" json:"title"`
    Description string          `gorm:"type:text" json:"description"`
    Brand       string          `json:"brand"`
    Status      int             `gorm:"default:0" json:"status"` // 0=draft, 1=published, 2=rejected
    Score       int             `gorm:"default:0" json:"score"`  // Content quality score
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
    
    Category    Category        `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
    Seller      Seller          `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
    SKUs        []SKU           `gorm:"foreignKey:ProductID" json:"skus,omitempty"`
    Media       []Media         `gorm:"polymorphic:Entity" json:"media,omitempty"`
}

const (
    ProductStatusDraft = iota
    ProductStatusPublished
    ProductStatusRejected
)

