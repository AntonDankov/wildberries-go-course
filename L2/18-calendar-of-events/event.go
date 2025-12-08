package main

import (
	"time"
)

type Event struct {
	ID          int
	UserID      int
	Date        time.Time
	Description string
}

type EventDTO struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	Date        string `json:"date"`
	Description string `json:"description"`
}

type DeleteEventDTO struct {
	ID     int `json:"id"`
	UserID int `json:"user_id"`
}

func ValidateEvent(id int, userID int, date time.Time, description string) error {
	if id < 0 {
		return ErrInvalidID
	}
	if userID < 0 {
		return ErrInvalidUserID
	}
	if date.IsZero() {
		return ErrInvalidDate
	}
	if description == "" {
		return ErrInvalidDescription
	}
	return nil
}

func convertEventToDTO(event Event) EventDTO {
	return EventDTO{
		ID:          event.ID,
		UserID:      event.UserID,
		Date:        event.Date.Format("2006-01-02"),
		Description: event.Description,
	}
}

func convertEventsToDTO(events []*Event) []EventDTO {
	dtos := make([]EventDTO, 0, len(events))
	for _, event := range events {
		dtos = append(dtos, convertEventToDTO(*event))
	}
	return dtos
}
