package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"wildberries-go-course/L3-5/database"
	"wildberries-go-course/L3-5/dto"
	"wildberries-go-course/L3-5/repository"

	"github.com/wb-go/wbf/ginext"
)

func CreateEvent(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var req dto.EventDTO

		if err := c.ShouldBindJSON(&req); err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
			return
		}

		eventID, err := repository.CreateEvent(ctx, db.Master, req.Name, req.Seats, req.BookSecondMaxTime)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to create event: %w", err))
			return
		}

		req.ID = eventID
		c.JSON(http.StatusCreated, req)
	}
}

func GetEvent(ctx context.Context, db *database.Database) ginext.HandlerFunc {
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

		event, err := repository.GetEvent(ctx, db.Master, eventID)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to get event: %w", err))
			return
		}

		c.JSON(http.StatusOK, ginext.H{
			"event": dto.ConvertEventToDTO(*event),
		})
	}
}

func addJSONWithError(c *ginext.Context, httpCode int, err error) {
	c.JSON(httpCode, ginext.H{
		"error": err.Error(),
	})
}
