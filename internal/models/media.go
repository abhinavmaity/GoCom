package models

import "time"

type Media struct {
	ID         uint   `gorm:"primaryKey"`
	EntityType string `gorm:"not null"` // product, review, etc.
	EntityID   uint   `gorm:"not null"`
	URL        string `gorm:"not null"`
	Type       string // image, video
	AltText    string
	Sort       int `gorm:"default:0"`
	CreatedAt  time.Time
}
