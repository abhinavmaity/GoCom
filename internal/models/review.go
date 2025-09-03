package models

import (
	"encoding/json"
	"time"
)

type Review struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"not null"`
	ProductID uint `gorm:"not null"`
	Rating    int  `gorm:"not null"`
	Text      string
	Media     json.RawMessage `gorm:"type:jsonb"`
	Status    int             `gorm:"default:1"` // 1=approved, 0=pending
	CreatedAt time.Time
}
