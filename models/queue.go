package models

import "time"
//coba tesy
type Queue struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	CounterID     uint       `json:"counter_id"`
	QueueNumber   string     `json:"queue_number"`
	CustomerName  string     `json:"customer_name"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	CalledAt      *time.Time `json:"called_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}