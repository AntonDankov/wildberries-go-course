package repository

import (
	"context"
	"fmt"
	database "widlberries-go-course/L3-1/database"
	model "widlberries-go-course/L3-1/model"
)

type PostgresRepository struct {
	db *database.Database
}

func NewPostgresRepository(db *database.Database) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) AddNotification(ctx context.Context, notification model.Notification) (int64, error) {
	query := `
		INSERT INTO notifications (text, status, created_at, send_at, failed_attempts, next_attempt_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	var id int64
	err := r.db.Master.QueryRowContext(ctx, query,
		notification.Text,
		notification.Status,
		notification.CreatedAt,
		notification.SendAt,
		notification.FailedAttempts,
		notification.NextAttemptAt,
	).Scan(&id)
	if err != nil {
		fmt.Printf("FAILED with error %v\n", err)
		return -1, fmt.Errorf("failed to add notification: %v", err)
	}
	return id, nil
}

func (r *PostgresRepository) GetNotification(ctx context.Context, id int64) (*model.Notification, error) {
	query := `
		SELECT id, text, status, created_at, send_at, failed_attempts, next_attempt_at
		FROM notifications
		WHERE id = $1
	`

	var notification model.Notification
	err := r.db.Master.QueryRowContext(ctx, query, id).Scan(
		&notification.ID,
		&notification.Text,
		&notification.Status,
		&notification.CreatedAt,
		&notification.SendAt,
		&notification.FailedAttempts,
		&notification.NextAttemptAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return &notification, nil
}

func (r *PostgresRepository) UpdateNotification(ctx context.Context, notification model.Notification) error {
	query := `
		UPDATE notifications 
		SET text = $1,
				status = $2, 
		    send_at = $3, 
		    failed_attempts = $4, 
		    next_attempt_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		notification.Text,
		notification.Status,
		notification.SendAt,
		notification.FailedAttempts,
		notification.NextAttemptAt,
		notification.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification with id %d not found", notification.ID)
	}

	return nil
}

func (r *PostgresRepository) UpdateNotificationStatus(ctx context.Context, id int64, status model.NotificationStatus) error {
	query := `
		UPDATE notifications 
		SET status = $1 
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query,
		status,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to mark as deleted notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification with id %d not found", id)
	}
	return nil
}

func (r *PostgresRepository) DeleteNotification(ctx context.Context, id int64) error {
	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification with id %d not found", id)
	}
	return nil
}

// func addNotification(ctx context.Context, db *database.Database, notification model.Notification) (int64, error) {
// 	query := `
// 		INSERT INTO notifications (text, status, created_at, send_at, failed_attempts, next_attempt_at)
// 		VALUES ($1, $2, $3, $4, $5, $6)
// 		RETURNING id
// 	`
// 	var id int64
// 	err := db.Master.QueryRowContext(ctx, query,
// 		notification.Text,
// 		notification.Status,
// 		notification.CreatedAt,
// 		notification.SendAt,
// 		notification.FailedAttempts,
// 		notification.NextAttemptAt,
// 	).Scan(&id)
// 	if err != nil {
// 		fmt.Printf("FAILED with error %v\n", err)
// 		return -1, fmt.Errorf("failed to add notification: %v", err)
// 	}
// 	return id, nil
// }

// func DeleteMarkNotification(ctx context.Context, db *database.Database, id int64) error {
// 	return UpdateNotificationStatus(ctx, db, id, model.Deleted)
// }

// func UpdateNotificationStatus(ctx context.Context, db *database.Database, id int64, status model.NotificationStatus) error {
// 	query := `
// 		UPDATE notifications
// 		SET status = $1
// 		WHERE id = $2
// 	`
//
// 	result, err := db.ExecContext(ctx, query,
// 		status,
// 		id,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to mark as deleted notification: %w", err)
// 	}
//
// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		return fmt.Errorf("failed to get rows affected: %w", err)
// 	}
//
// 	if rowsAffected == 0 {
// 		return fmt.Errorf("notification with id %d not found", id)
// 	}
// 	return nil
// }

// func DeleteNotification(ctx context.Context, db *database.Database, id int) error {
// 	query := `DELETE FROM notifications WHERE id = $1`
//
// 	result, err := db.ExecContext(ctx, query, id)
// 	if err != nil {
// 		return fmt.Errorf("failed to delete notification: %w", err)
// 	}
//
// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		return fmt.Errorf("failed to get rows affected: %w", err)
// 	}
//
// 	if rowsAffected == 0 {
// 		return fmt.Errorf("notification with id %d not found", id)
// 	}
// 	return nil
// }
//
// func UpdateNotification(ctx context.Context, db *database.Database, notification model.Notification) error {
// 	query := `
// 		UPDATE notifications
// 		SET text = $1,
// 				status = $2,
// 		    send_at = $3,
// 		    failed_attempts = $4,
// 		    next_attempt_at = $5
// 		WHERE id = $6
// 	`
//
// 	result, err := db.ExecContext(ctx, query,
// 		notification.Text,
// 		notification.Status,
// 		notification.SendAt,
// 		notification.FailedAttempts,
// 		notification.NextAttemptAt,
// 		notification.ID,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to update notification: %w", err)
// 	}
//
// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		return fmt.Errorf("failed to get rows affected: %w", err)
// 	}
//
// 	if rowsAffected == 0 {
// 		return fmt.Errorf("notification with id %d not found", notification.ID)
// 	}
//
// 	return nil
// }
//
// func GetNotification(ctx context.Context, db *database.Database, id int64) (*model.Notification, error) {
// 	query := `
// 		SELECT id, text, status, created_at, send_at, failed_attempts, next_attempt_at
// 		FROM notifications
// 		WHERE id = $1
// 	`
//
// 	var notification model.Notification
// 	err := db.Master.QueryRowContext(ctx, query, id).Scan(
// 		&notification.ID,
// 		&notification.Text,
// 		&notification.Status,
// 		&notification.CreatedAt,
// 		&notification.SendAt,
// 		&notification.FailedAttempts,
// 		&notification.NextAttemptAt,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get notification: %w", err)
// 	}
//
// 	return &notification, nil
// }
