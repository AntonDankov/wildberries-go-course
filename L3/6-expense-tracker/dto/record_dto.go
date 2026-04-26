package dto

import (
	"fmt"
	"time"
	"wildberries-go-course/L3-6/model"
)

type RecordDTO struct {
	ID       int64   `json:"id"`
	Type     int     `json:"type"`
	Category int     `json:"category"`
	Amount   float64 `json:"amount" binding:"required,min=0"`
	Date     string  `json:"date" binding:"required"`
}

func ConvertRecordFromDTO(recordDTO RecordDTO) (model.Record, error) {
	var record model.Record
	if recordDTO.Type < 0 || recordDTO.Type >= int(model.RecordTypeEnd) {
		return record, fmt.Errorf("ivalid type %d", recordDTO.Type)
	}
	if recordDTO.Category < 0 || recordDTO.Category >= int(model.RecordCategoryEnd) {
		return record, fmt.Errorf("ivalid category %d", recordDTO.Category)
	}
	if recordDTO.Amount < 0 {
		return record, fmt.Errorf("amount must be non-negative, got %.2f", recordDTO.Amount)
	}
	date, err := time.Parse("2006-01-02", recordDTO.Date)
	if err != nil {
		return record, fmt.Errorf("date format should be YYYY-MM-DD: %w", err)
	}

	record = model.Record{
		Type:     model.RecordType(recordDTO.Type),
		Category: model.RecordCategory(recordDTO.Category),
		Amount:   recordDTO.Amount,
		Date:     date,
	}

	return record, nil
}
