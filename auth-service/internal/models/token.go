package models

import "time"

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	Hash      string    `gorm:"not null;type:char(64);uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null"`

	//	FK
	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}
