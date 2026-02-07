package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"wildberries-go-course/L3-5/database"
	"wildberries-go-course/L3-5/dto"
	"wildberries-go-course/L3-5/model"
	"wildberries-go-course/L3-5/repository"

	"github.com/wb-go/wbf/ginext"
)

func BookEvent(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		eventIDStr := c.Param("id")
		if eventIDStr == "" {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("missing event ID"))
			return
		}

		eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid event ID: %w", err))
			return
		}

		tx, err := db.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelSerializable,
		})
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to begin transaction: %w", err))
			return
		}
		defer tx.Rollback()

		event, bookedCount, err := repository.GetEventWithBookedCount(ctx, tx, eventID)
		if err != nil {
			addJSONWithError(c, http.StatusNotFound, fmt.Errorf("event not found: %w", err))
			return
		}

		if bookedCount >= event.Seats {
			addJSONWithError(c, http.StatusConflict, fmt.Errorf("no seats available"))
			return
		}

		bookID, err := repository.BookEvent(ctx, tx, eventID)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to book event: %w", err))
			return
		}

		if err := tx.Commit(); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %w", err))
			return
		}

		response := dto.BookDTO{
			ID:      bookID,
			EventID: eventID,
			Status:  model.BookPending,
		}

		c.JSON(http.StatusCreated, response)
	}
}

func ConfirmBook(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		bookIDStr := c.Param("id")
		if bookIDStr == "" {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("missing book ID"))
			return
		}

		bookID, err := strconv.ParseInt(bookIDStr, 10, 64)
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid book ID: %w", err))
			return
		}

		tx, err := db.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelSerializable,
		})
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to begin transaction: %w", err))
			return
		}
		defer tx.Rollback()

		book, err := repository.GetBook(ctx, db.Master, bookID)
		if err != nil {
			addJSONWithError(c, http.StatusNotFound, fmt.Errorf("book not found: %w", err))
			return
		}

		if book.Status == model.BookConfirmed {
			addJSONWithError(c, http.StatusConflict, fmt.Errorf("book already confirmed"))
			return
		}

		if book.Status == model.BookCancelled {
			addJSONWithError(c, http.StatusConflict, fmt.Errorf("book was cancelled"))
			return
		}

		if err := repository.ConfirmBook(ctx, tx, bookID); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to confirm book: %w", err))
			return
		}

		if err := tx.Commit(); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %w", err))
			return
		}

		response := dto.BookDTO{
			ID:      bookID,
			EventID: book.EventID,
			Status:  model.BookConfirmed,
		}

		c.JSON(http.StatusOK, response)
	}
}

func GetEventBooks(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		eventIDStr := c.Param("id")
		if eventIDStr == "" {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("missing event ID"))
			return
		}

		eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid event ID: %w", err))
			return
		}

		books, err := repository.GetEventBooks(ctx, db.Master, eventID)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to get event books: %w", err))
			return
		}

		response := make([]dto.BookDTO, len(books))
		for i, book := range books {
			response[i] = dto.BookDTO{
				ID:      book.ID,
				EventID: book.EventID,
				Status:  book.Status,
			}
		}

		c.JSON(http.StatusOK, ginext.H{
			"event_id": eventID,
			"books":    response,
			"count":    len(response),
		})
	}
}
