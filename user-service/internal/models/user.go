package models

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"not null;size:100"`
	Email    string `gorm:"not null;uniqueIndex"`
	Password string `gorm:"not null"`
}
