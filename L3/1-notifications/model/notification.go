package model

import "time"

type NotificationStatus int

const (
	AwaitsSending NotificationStatus = iota
	FailedAndAwaits
	Failed
	Sended
	Deleted
)

type Notification struct {
	ID             int64
	Status         NotificationStatus
	Text           string
	CreatedAt      time.Time
	SendAt         time.Time
	FailedAttempts int
	NextAttemptAt  time.Time
}
