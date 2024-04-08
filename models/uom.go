package models

import "time"

type Uom struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	Name      string    `json:"name"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
