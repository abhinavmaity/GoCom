package models

import (
	"encoding/json"
	"time"

	"google.golang.org/genproto/googleapis/type/decimal"
)

type Coupon struct {
	ID         uint            `gorm:"primaryKey"`
	Code       string          `gorm:"unique;not null"`
	Type       int             // 1=percentage, 2=fixed
	Value      decimal.Decimal `gorm:"type:decimal(10,2)"`
	Conditions json.RawMessage `gorm:"type:jsonb"`
	StartAt    time.Time
	EndAt      time.Time
	UsageLimit int
	CreatedAt  time.Time
}
