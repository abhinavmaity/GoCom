package models

import (
    "encoding/json"
    "time"
)

type Category struct {
    ID              uint            `gorm:"primaryKey" json:"id"`
    ParentID        *uint           `json:"parent_id"`
    Name            string          `gorm:"not null" json:"name"`
    AttributesSchema json.RawMessage `gorm:"type:json" json:"attributes_schema"`
    SEOSlug         string          `gorm:"unique" json:"seo_slug"`
    IsActive        bool            `gorm:"default:true" json:"is_active"`
    CreatedAt       time.Time       `json:"created_at"`
    UpdatedAt       time.Time       `json:"updated_at"`
    

    Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
    Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
    Products []Product  `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}



type AttributeDefinition struct {
    Name        string   `json:"name"`
    Type        string   `json:"type"`       
    Required    bool     `json:"required"`
    Options     []string `json:"options,omitempty"` 
    Validation  string   `json:"validation,omitempty"`
}

type CategorySchema struct {
    Attributes []AttributeDefinition `json:"attributes"`
}

