package models

import (
    "time"
    "encoding/json"
)

type Category struct {
    ID               uint            `gorm:"primaryKey" json:"id"`
    ParentID         *uint           `json:"parent_id"`
    Name             string          `gorm:"not null" json:"name"`
    AttributesSchema json.RawMessage `gorm:"type:json" json:"attributes_schema"`
    SEOSlug          string          `json:"seo_slug"`
    IsActive         bool            `gorm:"default:true" json:"is_active"`
    CreatedAt        time.Time       `json:"created_at"`
    
    // Relations
    Parent           *Category       `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
    Children         []Category      `gorm:"foreignKey:ParentID" json:"children,omitempty"`
    Products         []Product       `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

// Category attribute schema helper
type AttributeDefinition struct {
    Name        string   `json:"name"`
    Type        string   `json:"type"`        // text, select, number, boolean
    Required    bool     `json:"required"`
    Options     []string `json:"options,omitempty"` // For select type
    Validation  string   `json:"validation,omitempty"`
}

type CategorySchema struct {
    Attributes []AttributeDefinition `json:"attributes"`
}

