package models

import (
    "time"
)

type KYC struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    SellerID    uint      `gorm:"not null" json:"seller_id"`
    Type        string    `gorm:"not null" json:"type"`
    DocumentURL string    `gorm:"not null" json:"document_url"`
    Status      int       `gorm:"default:0" json:"status"`
    Remarks     string    `json:"remarks"`
    CreatedAt   time.Time `json:"created_at"`
    
    // Relations
    Seller Seller `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
}
