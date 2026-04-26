package repository

import (
	"context"
	"database/sql"
	"fmt"
	"wildberries-go-course/L3-5/database"
	"wildberries-go-course/L3-5/model"
)

func CreateEvent(ctx context.Context, db database.DBTX, name string, seats int, bookSecondMaxTime int) (int64, error) {
	query := `
		INSERT INTO event (name, seats, book_second_max_time) 
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var eventID int64
	err := db.QueryRowContext(ctx, query, name, seats, bookSecondMaxTime).Scan(&eventID)
	if err != nil {
		return 0, fmt.Errorf("failed to create event: %v", err)
	}

	return eventID, nil
}

func GetEvent(ctx context.Context, db database.DBTX, eventID int64) (*model.Event, error) {
	query := `
		SELECT id, name, seats, book_second_max_time 
		FROM event 
		WHERE id = $1
	`

	var event model.Event
	err := db.QueryRowContext(ctx, query, eventID).Scan(
		&event.ID,
		&event.Name,
		&event.Seats,
		&event.BookSecondMaxTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found: %d", eventID)
		}
		return nil, fmt.Errorf("failed to get event: %v", err)
	}

	return &event, nil
}

func GetEvents(ctx context.Context, db database.DBTX) ([]model.Event, error) {
	query := `
		SELECT id, name, seats, book_second_max_time 
		FROM event
		ORDER by id DESC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.Event

	for rows.Next() {
		var event model.Event
		if err := rows.Scan(
			&event.ID,
			&event.Name,
			&event.Seats,
			&event.BookSecondMaxTime,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func GetEventBookedCount(ctx context.Context, db database.DBTX, eventID int64) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM book 
		WHERE event_id = $1 AND status = $2
	`

	var count int
	err := db.QueryRowContext(ctx, query, eventID, model.BookConfirmed).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get booked count: %v", err)
	}

	return count, nil
}
