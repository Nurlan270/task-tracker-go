package entity

import "time"

type Task struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CompletedAt *time.Time `json:"completed_at"`
}
