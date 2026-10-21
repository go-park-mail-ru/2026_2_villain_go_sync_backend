package models

import "time"

type Resume struct {
	ID          int       `json:"id"`
	SeekerID    int       `json:"seeker_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Experience  string    `json:"experience"`
	Education   string    `json:"education"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
