package dto

import "time"

type CreateNotificationRequest struct {
	SendAt time.Time `json:"send_at" binding:"required"`
	Text   string    `json:"text" binding: "required"`
}
