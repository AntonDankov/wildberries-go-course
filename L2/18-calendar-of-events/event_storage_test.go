package main

import (
	"testing"
	"time"
)

func TestEventStorage_CreateEvent(t *testing.T) {
	// Given
	storage := NewEventStorage()
	date := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	// When
	event, err := storage.CreateEvent(1, date, "DummyEvent")
	// Then
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if event.ID != 0 {
		t.Errorf("expected ID 0, got %d", event.ID)
	}
	if event.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", event.UserID)
	}
	if len(storage.userEvents[1]) != 1 {
		t.Errorf("expected 1 event in storage, got %d", len(storage.userEvents[1]))
	}
}

func TestEventStorage_DeleteEvent(t *testing.T) {
	// Given
	storage := NewEventStorage()
	event, _ := storage.CreateEvent(1, time.Now(), "To Delete")

	// When
	err := storage.DeleteEvent(1, event.ID)
	// Then
	if err != nil {
		t.Errorf("delete failed: %v", err)
	}
	if len(storage.userEvents[1]) != 0 {
		t.Error("expected storage to be empty after delete")
	}
}

func TestEventStorage_DeleteEvent_NotFound(t *testing.T) {
	// Given
	storage := NewEventStorage()

	// When
	err := storage.DeleteEvent(1, 999)

	// Then
	if err == nil {
		t.Error("expected error when deleting non-existent event")
	}
}

func TestEventStorage_UpdateEvent(t *testing.T) {
	// Given
	storage := NewEventStorage()
	originalDate := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	event, _ := storage.CreateEvent(1, originalDate, "Original")
	newDate := originalDate.Add(24 * time.Hour)

	// When
	updated, err := storage.UpdateEvent(1, event.ID, newDate, "Updated")
	// Then
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Description != "Updated" {
		t.Errorf("expected description 'Updated', got '%s'", updated.Description)
	}
	if !updated.Date.Equal(newDate) {
		t.Error("date was not updated")
	}
}

func TestEventStorage_GetEventByDay(t *testing.T) {
	// Given
	storage := NewEventStorage()
	day1 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	storage.CreateEvent(1, day1, "Day 1 Event")
	storage.CreateEvent(1, day2, "Day 2 Event")

	// When
	events := storage.GetEventByDay(1, day1)

	// Then
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
	if events[0].Description != "Day 1 Event" {
		t.Errorf("wrong event returned")
	}
}

func TestEventStorage_GetEventByWeek(t *testing.T) {
	// Given
	storage := NewEventStorage()
	startOfWeek := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	storage.CreateEvent(1, startOfWeek, "Monday")
	storage.CreateEvent(1, startOfWeek.AddDate(0, 0, 3), "Thursday")
	storage.CreateEvent(1, startOfWeek.AddDate(0, 0, 7), "Next Week")

	// When
	events := storage.GetEventByWeek(1, startOfWeek)

	// Then
	if len(events) != 2 {
		t.Errorf("expected 2 events for the week, got %d", len(events))
	}
}
