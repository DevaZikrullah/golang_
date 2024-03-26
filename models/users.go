package models

import "time"

type Users struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	Username  string    `json:"Username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Token     string    `json:"token"`
	Quests    []Quest   `json:"quests" gorm:"foreignkey:UserID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
