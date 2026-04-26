package main

import (
	"context"
	"log"
	"time"
)

func runCleanScheduler(ctx context.Context, interval time.Duration, deleteAfterLate time.Duration, eventStorage *EventStorage) {

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	if deleteAfterLate > 0 {
		deleteAfterLate = -1 * deleteAfterLate
	}

	for {
		select {
		case <-ticker.C:
			dateAfterDelete := time.Now().Add(deleteAfterLate)
			eventStorage.RemoveEventsBeforeDate(dateAfterDelete)
		case <-ctx.Done():
			return
		}
	}

}

func runNotificationScheduler(ctx context.Context, notificationHeap *NotificationEventMinHeap, messageChan chan NotificationEventMessage) {
	waitingTimer := time.NewTimer(0)
	for {
		select {
		case <-waitingTimer.C:
			durationToNextEvent := processNotificationHeap(notificationHeap)
			waitingTimer.Reset(durationToNextEvent)
		case message, ok := <-messageChan:
			if !ok {
				log.Printf("messageChan closed, exiting processor")
				return
			}
			log.Printf("We got message in notification processor")
			if message.event.ShouldNotify {
				switch message.Type {
				case Add:
					notificationHeap.Push(message.event)
				case Delete:
					notificationHeap.RemoveEvent(message.event.ID)
				case Update:
					notificationHeap.UpdateEvent(message.event)
				}
				waitingTimer.Reset(0)
			}

		case <-ctx.Done():
			log.Printf("We are done with context in notification")
			return
		}
	}

}

func processNotificationHeap(notificationHeap *NotificationEventMinHeap) time.Duration {
	durationToNextEvent := 16 * time.Minute
	for len(notificationHeap.heap) > 0 {
		timeNow := time.Now()
		timeOfEvent := notificationHeap.heap[0].Date
		log.Printf("time of next event %v and cur time is %v", timeOfEvent, timeNow)
		if !timeNow.Before(timeOfEvent) {
			event := notificationHeap.Pop()

			notify(event)
		} else {
			durationToNextEvent = timeOfEvent.Sub(timeNow)
			break
		}

	}
	return durationToNextEvent
}

type NotificationEventMessageType uint8

const (
	Add NotificationEventMessageType = 1 << iota
	Delete
	Update
)

type NotificationEventMessage struct {
	event Event
	Type  NotificationEventMessageType
}

func runNotificationEvents(ctx context.Context, messageChan chan NotificationEventMessage) {

	notificationHeap := NewNotificationEventMinHeap()

	go func() {
		runNotificationScheduler(ctx, notificationHeap, messageChan)
	}()

}

type NotificationEventMinHeap struct {
	heap             []Event
	mapEventIDInHeap map[uint]uint
	maxCap           int
	threshHoldDif    int
	resizeDif        int
}

func NewNotificationEventMinHeap() *NotificationEventMinHeap {
	return &NotificationEventMinHeap{
		heap:             make([]Event, 0, 100),
		mapEventIDInHeap: make(map[uint]uint),
		maxCap:           128,
		// Should > 2
		threshHoldDif: 4,
		// Should be at least 2 and less than threshHoldDIf
		resizeDif: 2,
	}
}

func (eventHeap *NotificationEventMinHeap) Push(event Event) {

	eventHeap.heap = append(eventHeap.heap, event)
	heap := eventHeap.heap
	insertedIndex := len(heap) - 1
	eventHeap.mapEventIDInHeap[event.ID] = uint(insertedIndex)
	eventHeap.bubleUp(insertedIndex)
}

func (eventHeap *NotificationEventMinHeap) Pop() Event {

	heap := eventHeap.heap
	if len(heap) == 0 {
		return Event{}
	}
	resultEvent := heap[0]
	lastIndex := len(heap) - 1
	heap[0] = heap[lastIndex]
	delete(eventHeap.mapEventIDInHeap, resultEvent.ID)

	//Shrinking if needed
	eventHeap.heap[lastIndex] = Event{}
	eventHeap.heap = eventHeap.heap[:lastIndex]
	currentCap := cap(eventHeap.heap)
	currentLen := len(eventHeap.heap)
	if currentCap > eventHeap.maxCap && currentLen < currentCap/eventHeap.threshHoldDif {
		newHeap := make([]Event, currentLen, currentCap/eventHeap.resizeDif)
		copy(newHeap, eventHeap.heap)
		eventHeap.heap = newHeap
	}

	eventHeap.bubleDown(0)

	return resultEvent

}

func (eventHeap *NotificationEventMinHeap) RemoveEvent(eventID uint) {
	index, exists := eventHeap.mapEventIDInHeap[eventID]
	if !exists {
		return

	}

	lastIndex := len(eventHeap.heap) - 1
	lastEvent := eventHeap.heap[lastIndex]
	eventHeap.heap[index] = lastEvent
	eventHeap.heap = eventHeap.heap[:lastIndex]
	eventHeap.mapEventIDInHeap[lastEvent.ID] = index
	delete(eventHeap.mapEventIDInHeap, eventID)
	eventHeap.fix(int(index))

}

func (eventHeap *NotificationEventMinHeap) UpdateEvent(event Event) {

	index, exists := eventHeap.mapEventIDInHeap[event.ID]
	if !exists {
		eventHeap.Push(event)
		return
	}

	eventHeap.heap[index] = event

	eventHeap.fix(int(index))

}

func (eventHeap *NotificationEventMinHeap) fix(index int) {
	if index >= len(eventHeap.heap) {
		return
	}
	shouldBubleUp := false
	event := eventHeap.heap[index]
	if index > 0 {
		topIndex := (index - 1) / 2
		topEvent := eventHeap.heap[topIndex]
		if event.Date.Compare(topEvent.Date) == -1 {
			shouldBubleUp = true
		}
	}
	if shouldBubleUp {
		eventHeap.bubleUp(int(index))
	} else {
		eventHeap.bubleDown(int(index))
	}
}

func (eventHeap *NotificationEventMinHeap) bubleUp(index int) {
	heap := eventHeap.heap
	if len(heap) == 0 {
		return
	}
	event := eventHeap.heap[index]
	for {
		if index == 0 {
			break
		}
		parentIndex := (index - 1) / 2
		parentNode := heap[parentIndex]
		if event.Date.Compare(parentNode.Date) == -1 {
			heap[index] = parentNode
			eventHeap.mapEventIDInHeap[parentNode.ID] = uint(index)

			heap[parentIndex] = event
			index = parentIndex
			eventHeap.mapEventIDInHeap[event.ID] = uint(index)
		} else {
			break
		}
	}

}

func (eventHeap *NotificationEventMinHeap) bubleDown(index int) {
	heap := eventHeap.heap
	if len(heap) == 0 {
		return
	}
	event := heap[index]
	eventHeap.mapEventIDInHeap[event.ID] = uint(index)
	for {
		if index == len(heap)-1 {
			break
		}
		//first check left
		leftIndex := index*2 + 1
		if leftIndex >= len(heap) {
			break
		}
		nextIndex := leftIndex
		nextEvent := heap[leftIndex]
		rightIndex := index*2 + 2
		if rightIndex < len(heap) {
			leftEvent := heap[leftIndex]
			rightEvent := heap[rightIndex]
			if rightEvent.Date.Compare(leftEvent.Date) == -1 {
				nextIndex = rightIndex
				nextEvent = rightEvent
			}
		}
		if event.Date.Compare(nextEvent.Date) == 1 {
			heap[index] = nextEvent
			eventHeap.mapEventIDInHeap[nextEvent.ID] = uint(index)
			index = nextIndex
			heap[index] = event
			eventHeap.mapEventIDInHeap[event.ID] = uint(index)
		} else {
			break
		}

	}
}

func notify(event Event) {
	// no proper implementation just logging instead
	log.Printf("Notification: eventID: %d, userID: %d, Date: %s, Description: %s", event.ID, event.UserID, event.Date, event.Description)
}
