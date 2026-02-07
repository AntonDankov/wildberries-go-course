package model

import "time"

type Comment struct {
	ID        int64     `json:"id"`
	Text      string    `json:"text"`
	Depth     int       `json:"depth,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}
