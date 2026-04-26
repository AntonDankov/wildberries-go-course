package main

import (
	"errors"
	"log"
	"sync"
	"time"
)

type EventStorage struct {
	userEvents    map[uint][]Event
	nextEventID   uint
	rwMutext      sync.RWMutex
	maxCap        int
	threshHoldDif int
	resizeDif     int
}

func NewEventStorage() *EventStorage {
	return &EventStorage{
		userEvents:  make(map[uint][]Event),
		nextEventID: 1, // 0 is reserved as not valid
	}
}

func (eventStorage *EventStorage) CreateEvent(event Event) (Event, error) {
	eventStorage.rwMutext.Lock()
	defer eventStorage.rwMutext.Unlock()

	event.ID = eventStorage.nextEventID
	eventStorage.nextEventID++

	list := eventStorage.userEvents[event.UserID]
	eventStorage.userEvents[event.UserID] = append(list, event)

	return event, nil
}

func (eventStorage *EventStorage) DeleteEvent(userID uint, eventID uint) error {
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
	events[lastIndex] = Event{}

	changedEvents := events[:lastIndex]
	eventStorage.userEvents[userID] = changedEvents

	currentCap := cap(changedEvents)
	currentLen := len(changedEvents)
	if currentCap > eventStorage.maxCap && currentLen < currentCap/eventStorage.threshHoldDif {
		newEvents := make([]Event, currentLen, currentCap/eventStorage.resizeDif)
		copy(newEvents, changedEvents)
		eventStorage.userEvents[userID] = newEvents
	}

	return nil
}

func (eventStorage *EventStorage) UpdateEvent(event Event) (Event, error) {
	eventStorage.rwMutext.Lock()
	defer eventStorage.rwMutext.Unlock()

	events, exists := eventStorage.userEvents[event.UserID]
	if !exists {
		return Event{}, errors.New("user has no events")
	}

	targetIndex := -1
	for i, storedEvent := range events {
		if storedEvent.ID == event.ID {
			targetIndex = i
			break
		}
	}
	if targetIndex == -1 {
		return Event{}, errors.New("not found event")
	}

	events[targetIndex] = event

	return events[targetIndex], nil
}

func (eventStorage *EventStorage) GetEventByDay(userID uint, date time.Time) []Event {
	eventStorage.rwMutext.RLock()
	defer eventStorage.rwMutext.RUnlock()

	var eventList []Event
	events := eventStorage.userEvents[userID]
	for _, event := range events {
		if isEventOnSameDayDate(date, event) {
			eventList = append(eventList, event)
		}
	}

	return eventList
}

func (eventStorage *EventStorage) GetEventByWeek(userID uint, date time.Time) []Event {
	// We want include the provided date so -> [startDate, endDate)
	startDate := date.AddDate(0, 0, -1)
	endDate := date.AddDate(0, 0, 7)
	return eventStorage.GetEventByDateRange(userID, startDate, endDate)
}

func (eventStorage *EventStorage) GetEventByMonth(userID uint, date time.Time) []Event {
	// We want include the provided date so -> [startDate, endDate)
	startDate := date.AddDate(0, 0, -1)
	endDate := date.AddDate(0, 1, 0)
	return eventStorage.GetEventByDateRange(userID, startDate, endDate)
}

func (eventStorage *EventStorage) GetEventByDateRange(userID uint, startDate time.Time, endDate time.Time) []Event {
	eventStorage.rwMutext.RLock()
	defer eventStorage.rwMutext.RUnlock()

	var eventList []Event
	events := eventStorage.userEvents[userID]

	for _, event := range events {
		if startDate.Before(event.Date) && endDate.After(event.Date) {
			eventList = append(eventList, event)
		}
	}
	return eventList
}

func (eventStorage *EventStorage) RemoveEventsBeforeDate(date time.Time) {
	eventStorage.rwMutext.Lock()
	defer eventStorage.rwMutext.Unlock()

	for userID, events := range eventStorage.userEvents {
		filteredEvents := events[:0]
		for _, event := range events {
			if !date.After(event.Date) {
				filteredEvents = append(filteredEvents, event)
			} else {
				log.Printf("Removing event with id %d and date %v", event.ID, event.Date)
			}
		}

		eventStorage.userEvents[userID] = filteredEvents
		currentCap := cap(filteredEvents)
		currentLen := len(filteredEvents)
		if currentCap > eventStorage.maxCap && currentLen < currentCap/eventStorage.threshHoldDif {
			newEvents := make([]Event, currentLen, currentCap/eventStorage.resizeDif)
			copy(newEvents, filteredEvents)
			eventStorage.userEvents[userID] = newEvents
		}
	}
}

func (eventStorage *EventStorage) GetEventByUserIDAndID(userID uint, eventID uint) Event {
	eventStorage.rwMutext.RLock()
	defer eventStorage.rwMutext.RUnlock()

	foundEvent := Event{
		ID: 0,
	}
	for _, event := range eventStorage.userEvents[userID] {
		if event.ID == eventID {
			foundEvent = event
			break
		}
	}
	return foundEvent

}

func isEventOnSameDayDate(date time.Time, event Event) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := event.Date.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
