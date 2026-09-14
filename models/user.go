package models

import "time"

type User struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string     `gorm:"size:255;not null" json:"name"`
	Email           string     `gorm:"size:255;not null;unique" json:"email"`
	Role            string     `gorm:"size:255;default:'user'" json:"role"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	Password        string     `gorm:"size:255;not null" json:"-"`
	RememberToken   string     `gorm:"size:100" json:"-"`
	CreatedAt       *time.Time `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}
