package models

import "time"

type RefreshToken struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null"`
	Hash      string `gorm:"not null;uniqueIndex"`
	ExpiresAt time.Time

	//	FK
	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}
