package models

type Counter struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}