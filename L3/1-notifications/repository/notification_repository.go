package repository

import (
	"context"

	model "widlberries-go-course/L3-1/model"
)

// I regret adding interface, made an OOPs
type NotificationRepository interface {
	AddNotification(ctx context.Context, n model.Notification) (int64, error)
	GetNotification(ctx context.Context, id int64) (*model.Notification, error)
	UpdateNotification(ctx context.Context, n model.Notification) error
	UpdateNotificationStatus(ctx context.Context, id int64, status model.NotificationStatus) error
	DeleteNotification(ctx context.Context, id int64) error
}
