package model

import "time"

type Item struct {
	ID        int64
	OwnerID   int64
	Name      string
	Price     float64
	Amount    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ActionType int

const (
	Insert ActionType = iota
	Update
	Delete
)

type ItemHistory struct {
	ID              int64      `json:"id"`
	ItemID          int64      `json:"item_id"`
	Name            *string    `json:"name,omitempty"`
	Price           *float64   `json:"price,omitempty"`
	Amount          *int       `json:"amount,omitempty"`
	Action          ActionType `json:"action"`
	ChangedByUserID int64      `json:"changed_by_user_id"`
	ChangedAt       time.Time  `json:"changed_at"`
}
