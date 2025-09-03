package models

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"not null"`
	Email        string `gorm:"unique"`
	Phone        string `gorm:"unique"`
	PasswordHash string
	Status       int  `gorm:"default:1"` // 1=active, 0=inactive
	TwoFAEnabled bool `gorm:"default:false"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
