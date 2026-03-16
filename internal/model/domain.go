package model

import "time"

type Domain struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Subdomain string    `gorm:"uniqueIndex;size:64;not null" json:"subdomain"`
	Token     string    `gorm:"uniqueIndex;size:128;not null" json:"token"`
	Status    string    `gorm:"size:16;not null;default:active" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
