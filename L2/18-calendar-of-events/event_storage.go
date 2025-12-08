package main

import (
	"errors"
	"sync"
	"time"
)

type EventStorage struct {
	userEvents  map[int][]*Event
	nextEventID int
	rwMutext    sync.RWMutex
}

func NewEventStorage() *EventStorage {
	return &EventStorage{
		userEvents:  make(map[int][]*Event),
		nextEventID: 0,
	}
}

func (eventStorage *EventStorage) CreateEvent(userID int, date time.Time, description string) (*Event, error) {
	if err := ValidateEvent(0, userID, date, description); err != nil {
		return nil, err
	}

	eventStorage.rwMutext.Lock()
	defer eventStorage.rwMutext.Unlock()

	event := &Event{
		ID:          eventStorage.nextEventID,
		UserID:      userID,
		Date:        date,
		Description: description,
	}
	eventStorage.nextEventID++

	list := eventStorage.userEvents[userID]
	eventStorage.userEvents[userID] = append(list, event)

	return event, nil
}

func (eventStorage *EventStorage) DeleteEvent(userID int, eventID int) error {
	eventStorage.rwMutext.Lock()
	defer eventStorage.rwMutext.Unlock()

	events, exists := eventStorage.userEvents[userID]
	if !exists {
		return errors.New("user has no events")
	}

	targetIndex := -1
	for i, event := range events {
		if event.ID == eventID {
			targetIndex = i
			break
		}
	}
	if targetIndex == -1 {
		return errors.New("not found event")
	}

	lastIndex := len(events) - 1
	events[targetIndex] = events[lastIndex]
	events[lastIndex] = nil

	eventStorage.userEvents[userID] = events[:lastIndex]

	return nil
}

func (eventStorage *EventStorage) UpdateEvent(userID int, eventID int, date time.Time, description string) (*Event, error) {
	eventStorage.rwMutext.Lock()
	defer eventStorage.rwMutext.Unlock()

	events, exists := eventStorage.userEvents[userID]
	if !exists {
		return nil, errors.New("user has no events")
	}

	targetIndex := -1
	for i, event := range events {
		if event.ID == eventID {
			targetIndex = i
			break
		}
	}
	if targetIndex == -1 {
		return nil, errors.New("not found event")
	}

	events[targetIndex].Description = description
	events[targetIndex].Date = date

	return events[targetIndex], nil
}

func (eventStorage *EventStorage) GetEventByDay(userID int, date time.Time) []*Event {
	eventStorage.rwMutext.RLock()
	defer eventStorage.rwMutext.RUnlock()

	var eventList []*Event
	events := eventStorage.userEvents[userID]
	for _, event := range events {
		if isEventOnDate(date, event) {
			eventList = append(eventList, event)
		}
	}

	return eventList
}

func (eventStorage *EventStorage) GetEventByWeek(userID int, date time.Time) []*Event {
	// We want include the provided date so -> [startDate, endDate)
	startDate := date.AddDate(0, 0, -1)
	endDate := date.AddDate(0, 0, 7)
	return eventStorage.GetEventByDateRange(userID, startDate, endDate)
}

func (eventStorage *EventStorage) GetEventByMonth(userID int, date time.Time) []*Event {
	// We want include the provided date so -> [startDate, endDate)
	startDate := date.AddDate(0, 0, -1)
	endDate := date.AddDate(0, 1, 0)
	return eventStorage.GetEventByDateRange(userID, startDate, endDate)
}

func (eventStorage *EventStorage) GetEventByDateRange(userID int, startDate time.Time, endDate time.Time) []*Event {
	eventStorage.rwMutext.RLock()
	defer eventStorage.rwMutext.RUnlock()

	var eventList []*Event
	events := eventStorage.userEvents[userID]

	for _, event := range events {
		if startDate.Before(event.Date) && endDate.After(event.Date) {
			eventList = append(eventList, event)
		}
	}
	return eventList
}

func isEventOnDate(date time.Time, event *Event) bool {
	return date.Equal(event.Date)
}
