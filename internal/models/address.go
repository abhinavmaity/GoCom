package models

import (
    "time"
)

type Address struct {
    ID       uint   `gorm:"primaryKey" json:"id"`
    UserID   *uint  `json:"user_id,omitempty"`
    SellerID *uint  `json:"seller_id,omitempty"`
    Line1    string `gorm:"not null" json:"line1"`
    Line2    string `json:"line2"`
    City     string `gorm:"not null" json:"city"`
    State    string `gorm:"not null" json:"state"`
    Country  string `gorm:"not null" json:"country"`
    Pin      string `gorm:"not null" json:"pin"`
    CreatedAt time.Time `json:"created_at"`
    
    // Relations
    User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Seller Seller `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
}
