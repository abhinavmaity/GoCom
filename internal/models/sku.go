package models

import (
    "time"
    "encoding/json"
    "github.com/shopspring/decimal"
)

type SKU struct {
    ID         uint            `gorm:"primaryKey" json:"id"`
    ProductID  uint            `gorm:"not null" json:"product_id"`
    SKUCode    string          `gorm:"unique;not null" json:"sku_code"`
    Attributes json.RawMessage `gorm:"type:json" json:"attributes"` // {color: "red", size: "L"}
    PriceMRP   decimal.Decimal `gorm:"type:decimal(10,2)" json:"price_mrp"`
    PriceSell  decimal.Decimal `gorm:"type:decimal(10,2)" json:"price_sell"`
    TaxPct     decimal.Decimal `gorm:"type:decimal(5,2)" json:"tax_pct"`
    Barcode    string          `json:"barcode"`
    IsActive   bool            `gorm:"default:true" json:"is_active"`
    CreatedAt  time.Time       `json:"created_at"`
    
    // Relations  
    Product    Product         `gorm:"foreignKey:ProductID" json:"product,omitempty"`
    Inventory  []Inventory     `gorm:"foreignKey:SKUID" json:"inventory,omitempty"`
}

// SKU attribute helper
type SKUAttributes struct {
    Color    string `json:"color,omitempty"`
    Size     string `json:"size,omitempty"`
    Material string `json:"material,omitempty"`
    Model    string `json:"model,omitempty"`
}

func (s *SKU) GetAttributes() (*SKUAttributes, error) {
    var attrs SKUAttributes
    if len(s.Attributes) > 0 {
        err := json.Unmarshal(s.Attributes, &attrs)
        return &attrs, err
    }
    return &attrs, nil
}

func (s *SKU) SetAttributes(attrs SKUAttributes) error {
    data, err := json.Marshal(attrs)
    if err != nil {
        return err
    }
    s.Attributes = data
    return nil
}

