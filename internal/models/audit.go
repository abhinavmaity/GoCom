package models

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID        uint            `gorm:"primaryKey"`
	Actor     string          `gorm:"not null"`
	Action    string          `gorm:"not null"`
	Entity    string          `gorm:"not null"`
	EntityID  uint            `gorm:"not null"`
	Meta      json.RawMessage `gorm:"type:jsonb"`
	CreatedAt time.Time
}
