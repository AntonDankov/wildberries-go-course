package model

import "time"

type Item struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"owner_id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Amount    int       `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ActionType int

const (
	Insert ActionType = iota
	Update
	Delete
)

type ItemHistory struct {
	ID        int64      `json:"id"`
	ItemID    int64      `json:"item_id"`
	Name      *string    `json:"name,omitempty"`
	Price     *float64   `json:"price,omitempty"`
	Amount    *int       `json:"amount,omitempty"`
	Action    ActionType `json:"action"`
	UserID    int64      `json:"user_id"`
	Username  string     `json:"username"`
	ChangedAt time.Time  `json:"changed_at"`
}
