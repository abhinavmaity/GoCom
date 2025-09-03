package models

import (
	"encoding/json"
	"time"
)

type Category struct {
	ID               uint `gorm:"primaryKey"`
	ParentID         *uint
	Name             string          `gorm:"not null"`
	AttributesSchema json.RawMessage `gorm:"type:jsonb"`
	SEOSlug          string
	CreatedAt        time.Time
}
