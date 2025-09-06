// internal/models/user.go
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

	// Add these OTP fields
	OTPSecret   string `gorm:"column:otp_secret"`
	OTPEnabled  bool   `gorm:"default:false"`
	OTPVerified bool   `gorm:"default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Add RefreshToken model
type RefreshToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	Token     string    `gorm:"unique;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IsRevoked bool      `gorm:"default:false"`
	CreatedAt time.Time
	User      User `gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}
