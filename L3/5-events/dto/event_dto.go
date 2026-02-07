package dto

import "wildberries-go-course/L3-5/model"

type EventDTO struct {
	ID                int64  `json:"id"`
	Name              string `json:"name" binding:"required"`
	Seats             int    `json:"seats" binding:"required,min=1"`
	BookSecondMaxTime int    `json:"book_second_max_time,min=60"`
}

func ConvertEventToDTO(event model.Event) EventDTO {
	return EventDTO{
		ID:                event.ID,
		Name:              event.Name,
		Seats:             event.Seats,
		BookSecondMaxTime: event.BookSecondMaxTime,
	}
}

type BookDTO struct {
	ID      int64            `json:"id"`
	EventID int64            `json:"event_id"`
	Status  model.BookStatus `json:"book_status"`
}
