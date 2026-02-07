package model

import "time"

type Event struct {
	ID                int64
	Name              string
	Seats             int
	BookSecondMaxTime int
}

type BookStatus int

const (
	BookPending BookStatus = iota
	BookConfirmed
	BookCancelled
)

type Book struct {
	ID        int64
	EventID   int64
	Status    BookStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
