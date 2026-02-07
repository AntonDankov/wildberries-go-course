package repository

import (
	"context"
	"database/sql"
	"fmt"
	"wildberries-go-course/L3-5/database"
	"wildberries-go-course/L3-5/model"

	"github.com/wb-go/wbf/zlog"
)

func BookEvent(ctx context.Context, db database.DBTX, eventID int64) (int64, error) {
	query := `
		INSERT INTO book (event_id, status) 
		VALUES ($1, $2)
		RETURNING id
	`

	var bookID int64
	err := db.QueryRowContext(ctx, query, eventID, model.BookPending).Scan(&bookID)
	if err != nil {
		return 0, fmt.Errorf("failed to book event: %v", err)
	}

	return bookID, nil
}

func ConfirmBook(ctx context.Context, db database.DBTX, bookID int64) error {
	query := `
		UPDATE book 
		SET status = $1, updated_at = NOW() 
		WHERE id = $2
	`

	result, err := db.ExecContext(ctx, query, model.BookConfirmed, bookID)
	if err != nil {
		return fmt.Errorf("failed to confirm book: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("book not found: %d", bookID)
	}

	return nil
}

func GetBook(ctx context.Context, db database.DBTX, bookID int64) (*model.Book, error) {
	query := `
		SELECT id, event_id, status, created_at, updated_at 
		FROM book 
		WHERE id = $1
	`

	var book model.Book
	err := db.QueryRowContext(ctx, query, bookID).Scan(
		&book.ID,
		&book.EventID,
		&book.Status,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("book not found: %d", bookID)
		}
		return nil, fmt.Errorf("failed to get book: %v", err)
	}

	return &book, nil
}

func GetEventBooks(ctx context.Context, db database.DBTX, eventID int64) ([]*model.Book, error) {
	query := `
		SELECT id, event_id, status, created_at, updated_at 
		FROM book 
		WHERE event_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event books: %v", err)
	}
	defer rows.Close()

	var books []*model.Book
	for rows.Next() {
		var book model.Book
		err := rows.Scan(
			&book.ID,
			&book.EventID,
			&book.Status,
			&book.CreatedAt,
			&book.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan book: %v", err)
		}
		books = append(books, &book)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating books: %v", err)
	}

	return books, nil
}

func GetEventWithBookedCount(ctx context.Context, db database.DBTX, eventID int64) (*model.Event, int, error) {
	query := `
		SELECT e.id, e.name, e.seats, e.book_second_max_time,
		       (SELECT COUNT(*) FROM book WHERE event_id = e.id AND status != $2) as booked_count
		FROM event e
		WHERE e.id = $1
	`

	var event model.Event
	var bookedCount int
	err := db.QueryRowContext(ctx, query, eventID, model.BookCancelled).Scan(
		&event.ID,
		&event.Name,
		&event.Seats,
		&event.BookSecondMaxTime,
		&bookedCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, fmt.Errorf("event not found: %d", eventID)
		}
		return nil, 0, fmt.Errorf("failed to get event with booked count: %v", err)
	}

	return &event, bookedCount, nil
}

func CancelTimeoutedBooks(ctx context.Context, db database.DBTX) error {
	query := `
		UPDATE book b SET
		status = $1, updated_at = NOW()
		FROM event e
		WHERE b.event_id = e.id AND b.status = $2
		AND NOW() > b.created_at + make_interval(secs => e.book_second_max_time) 
	`

	result, err := db.ExecContext(ctx, query, model.BookCancelled, model.BookPending)
	if err != nil {
		return fmt.Errorf("failed to cancel timed out books : %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get amount of affected rows: %v", err)
	}

	if rowsAffected > 0 {
		zlog.Logger.Info().Msgf("Amount of books canceled by timeout are: %v", rowsAffected)
	}

	return nil
}
