package models

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Email        string    `gorm:"size:255;uniqueIndex" json:"email"`         // Fixed: Added size:255
	Phone        string    `gorm:"size:20;uniqueIndex" json:"phone"`          // Fixed: Added size:20
	PasswordHash string    `json:"-"`
	Status       int       `gorm:"default:1" json:"status"` // 1=active, 0=inactive
	TwoFAEnabled bool      `gorm:"default:false" json:"two_fa_enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

const (
	UserStatusInactive = 0
	UserStatusActive   = 1
)
