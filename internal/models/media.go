package models

import (
    "time"
)

type Media struct {
    ID         uint   `gorm:"primaryKey" json:"id"`
    EntityType string `gorm:"not null" json:"entity_type"` // product, seller
    EntityID   uint   `gorm:"not null" json:"entity_id"`
    Type       string `gorm:"not null" json:"type"` // image, video
    URL        string `gorm:"not null" json:"url"`
    Alt        string `json:"alt"`
    SortOrder  int    `gorm:"default:0" json:"sort_order"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt       time.Time       `json:"updated_at"`
}
