package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"wildberries-go-course/L3-6/database"
	"wildberries-go-course/L3-6/dto"
	"wildberries-go-course/L3-6/repository"

	"github.com/wb-go/wbf/ginext"
)

func CreateRecord(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var recordDTO dto.RecordDTO

		if err := c.ShouldBindJSON(&recordDTO); err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
			return
		}

		record, err := dto.ConvertRecordFromDTO(recordDTO)
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, err)
			return
		}

		recordID, err := repository.CreateRecord(ctx, db.Master, record)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to create record: %w", err))
			return
		}

		record.ID = recordID
		c.JSON(http.StatusCreated, record)
	}
}

func UpdateRecord(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		recordIDStr := c.Param("id")
		if recordIDStr == "" {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("missing record ID"))
			return
		}

		recordID, err := strconv.ParseInt(recordIDStr, 10, 64)
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid record ID: %w", err))
			return
		}

		var recordDTO dto.RecordDTO

		if err := c.ShouldBindJSON(&recordDTO); err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
			return
		}

		record, err := dto.ConvertRecordFromDTO(recordDTO)
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, err)
			return
		}

		if err := repository.UpdateRecord(ctx, db.Master, recordID, record); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to update record: %w", err))
			return
		}

		c.JSON(http.StatusOK, record)
	}
}

func DeleteRecord(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		recordIDStr := c.Param("id")
		if recordIDStr == "" {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("missing record ID"))
			return
		}

		recordID, err := strconv.ParseInt(recordIDStr, 10, 64)
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid record ID: %w", err))
			return
		}

		if err := repository.DeleteRecord(ctx, db.Master, recordID); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to delete record: %w", err))
			return
		}

		c.JSON(http.StatusOK, ginext.H{
			"message": "record deleted successfully",
		})
	}
}

func GetRecords(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		filter := dto.ParseFilter(c)

		records, err := repository.GetRecords(ctx, db.Master, filter)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to get records: %w", err))
			return
		}

		recordDTOs := make([]dto.RecordDTO, len(records))
		for i, record := range records {
			recordDTOs[i] = dto.RecordDTO{
				ID:       record.ID,
				Type:     int(record.Type),
				Category: int(record.Category),
				Amount:   record.Amount,
				Date:     record.Date.Format("2006-01-02"),
			}
		}

		c.JSON(http.StatusOK, ginext.H{
			"records": recordDTOs,
			"count":   len(recordDTOs),
		})
	}
}

func addJSONWithError(c *ginext.Context, httpCode int, err error) {
	c.JSON(httpCode, ginext.H{
		"error": err.Error(),
	})
}
