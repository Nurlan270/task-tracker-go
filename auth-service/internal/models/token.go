package models

import "time"

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	TokenHash string    `gorm:"not null;type:char(64);uniqueIndex"`
	CreatedAt time.Time `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	RevokedAt *time.Time

	//	FK
	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}
