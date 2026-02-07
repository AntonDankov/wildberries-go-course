package dto

import (
	"time"
	"wildberries-go-course/L3-7/model"
)

type ItemDTO struct {
	ID      int64   `json:"id"`
	OwnerID int64   `json:"owner_id"`
	Name    string  `json:"name" binding:"required,min=1"`
	Price   float64 `json:"price" binding:"required,min=0"`
	Amount  int     `json:"amount" binding:"required,min=0"`
}

type ItemHistoryDTO struct {
	ID              int64            `json:"id"`
	ItemID          int64            `json:"item_id"`
	Name            *string          `json:"name,omitempty"`
	Price           *float64         `json:"price,omitempty"`
	Amount          *int             `json:"amount,omitempty"`
	Action          model.ActionType `json:"action"`
	ChangedByUserID int64            `json:"changed_by_user_id"`
	ChangedAt       time.Time        `json:"changed_at"`
}
