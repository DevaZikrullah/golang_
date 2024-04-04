package models

import "time"

type Quest struct {
	ID          uint      `json:"id" gorm:"primary_key"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Reward      int       `json:"reward"`
	UserID      uint      `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CompletedQuest struct {
	ID          uint `gorm:"primary_key"`
	UserID      uint
	QuestID     uint
	CompletedAt time.Time
}
