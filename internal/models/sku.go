package models

import (
	"encoding/json"
	"time"

	"google.golang.org/genproto/googleapis/type/decimal"
)

type SKU struct {
	ID         uint            `gorm:"primaryKey"`
	ProductID  uint            `gorm:"not null"`
	SKUCode    string          `gorm:"unique;not null"`
	Attributes json.RawMessage `gorm:"type:jsonb"`
	PriceMRP   decimal.Decimal `gorm:"type:decimal(10,2)"`
	PriceSell  decimal.Decimal `gorm:"type:decimal(10,2)"`
	TaxPct     decimal.Decimal `gorm:"type:decimal(5,2)"`
	Barcode    string
	CreatedAt  time.Time
}
