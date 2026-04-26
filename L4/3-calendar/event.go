package main

import (
	"time"
)

type Event struct {
	ID           uint
	UserID       uint
	Date         time.Time
	Description  string
	ShouldNotify bool
}

type EventDTO struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`
	Date         string `json:"date"`
	Description  string `json:"description"`
	ShouldNotify bool   `json:"should_notify"`
}

type DeleteEventDTO struct {
	ID     uint `json:"id"`
	UserID uint `json:"user_id"`
}

func ValidateEvent(event Event, shouldValidateID bool) error {
	if shouldValidateID && event.ID == 0 {
		return ErrInvalidID
	}
	if event.UserID == 0 {
		return ErrInvalidUserID
	}
	if event.Date.IsZero() {
		return ErrInvalidDate
	}
	if event.Description == "" {
		return ErrInvalidDescription
	}
	return nil
}

const DateFormatEventDTO = "2006-01-02 15:04:05"

func convertEventToDTO(event Event) EventDTO {
	return EventDTO{
		ID:           event.ID,
		UserID:       event.UserID,
		Date:         event.Date.Format(DateFormatEventDTO),
		Description:  event.Description,
		ShouldNotify: event.ShouldNotify,
	}
}

func convertEventsToDTO(events []Event) []EventDTO {
	dtos := make([]EventDTO, 0, len(events))
	for _, event := range events {
		dtos = append(dtos, convertEventToDTO(event))
	}
	return dtos
}
