package models

import "time"

type Product struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	Name      string    `json:"name"`
	UserID    uint      `json:"user_id"`
	Qty       int       `json:"qty"`
	UomID     uint      `json:"uom_id"`
	Uom       Uom       `json:"uom" gorm:"foreignkey:UomID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
