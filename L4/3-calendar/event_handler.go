package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

type EventHTTPHandler struct {
	eventStorage            *EventStorage
	notificationMessageChan chan NotificationEventMessage
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Result any `json:"result"`
}

func NewEventHTTPHandler(eventStorage *EventStorage, notificationMessage chan NotificationEventMessage) *EventHTTPHandler {
	return &EventHTTPHandler{eventStorage: eventStorage, notificationMessageChan: notificationMessage}
}

func (handler *EventHTTPHandler) CreateEvent(writer http.ResponseWriter, request *http.Request) {
	event, err := parseEventDataFromJSONRequest(request, false)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
		return
	}
	createdEvent, err := handler.eventStorage.CreateEvent(event)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
		return
	}
	eventDTO := convertEventToDTO(createdEvent)
	notificationMessage := NotificationEventMessage{
		event: createdEvent,
		Type:  Add,
	}
	handler.notificationMessageChan <- notificationMessage
	handler.respondSuccess(writer, eventDTO)
}

func (handler *EventHTTPHandler) GetEventByDate(writer http.ResponseWriter, request *http.Request) {
	userID, date, err := handler.parseEventDataFromQueryParams(request)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
		return
	}
	events := handler.eventStorage.GetEventByDay(userID, date)
	eventDTOs := convertEventsToDTO(events)
	handler.respondSuccess(writer, eventDTOs)
}

func (handler *EventHTTPHandler) GetEventByWeek(writer http.ResponseWriter, request *http.Request) {
	userID, date, err := handler.parseEventDataFromQueryParams(request)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
		return
	}
	events := handler.eventStorage.GetEventByWeek(userID, date)
	eventDTOs := convertEventsToDTO(events)
	handler.respondSuccess(writer, eventDTOs)
}

func (handler *EventHTTPHandler) GetEventByMonth(writer http.ResponseWriter, request *http.Request) {
	userID, date, err := handler.parseEventDataFromQueryParams(request)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
		return
	}
	events := handler.eventStorage.GetEventByMonth(userID, date)
	eventDTOs := convertEventsToDTO(events)
	handler.respondSuccess(writer, eventDTOs)
}

func (handler *EventHTTPHandler) DeleteEvent(writer http.ResponseWriter, request *http.Request) {
	deleveEventDTO, err := parseDeleteEventDataFromJSONRequest(request)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
		return
	}
	err = handler.eventStorage.DeleteEvent(deleveEventDTO.UserID, deleveEventDTO.ID)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusServiceUnavailable)
		return
	}
	notificationMessage := NotificationEventMessage{
		event: Event{ID: deleveEventDTO.ID},
		Type:  Delete,
	}
	handler.notificationMessageChan <- notificationMessage
	handler.respondSuccess(writer, "event was deleted")
}

func (handler *EventHTTPHandler) UpdateEvent(writer http.ResponseWriter, request *http.Request) {
	event, err := parseEventDataFromJSONRequest(request, true)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
		return
	}
	updatedEvent, err := handler.eventStorage.UpdateEvent(event)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if updatedEvent.ID == 0 {
		handler.respondError(writer, "", http.StatusNotFound)
		return
	}
	eventDTO := convertEventToDTO(updatedEvent)
	notificationMessage := NotificationEventMessage{
		event: event,
		Type:  Update,
	}
	handler.notificationMessageChan <- notificationMessage
	handler.respondSuccess(writer, eventDTO)
}

func parseEventDataFromJSONRequest(request *http.Request, shouldValidateID bool) (Event, error) {
	event := Event{}
	if request.Header.Get("Content-Type") != "application/json" {
		return event, ErrInvalidContentType
	}
	var dataDTO EventDTO
	if err := json.NewDecoder(request.Body).Decode(&dataDTO); err != nil {
		return event, err
	}
	date, err := time.ParseInLocation(DateFormatEventDTO, dataDTO.Date, time.Local)
	if err != nil {
		log.Printf("Failed with date, %v", err)
		return event, ErrInvalidDate
	}
	event = Event{
		ID:           dataDTO.ID,
		UserID:       dataDTO.UserID,
		Date:         date,
		Description:  dataDTO.Description,
		ShouldNotify: dataDTO.ShouldNotify,
	}
	if err := ValidateEvent(event, shouldValidateID); err != nil {
		return event, err
	}
	return event, nil
}

func parseDeleteEventDataFromJSONRequest(request *http.Request) (DeleteEventDTO, error) {
	if request.Header.Get("Content-Type") != "application/json" {
		return DeleteEventDTO{}, ErrInvalidContentType
	}
	var dataDTO DeleteEventDTO
	if err := json.NewDecoder(request.Body).Decode(&dataDTO); err != nil {
		return DeleteEventDTO{}, err
	}
	if dataDTO.ID == 0 {
		return DeleteEventDTO{}, ErrInvalidID
	}
	if dataDTO.UserID == 0 {
		return DeleteEventDTO{}, ErrInvalidUserID
	}
	return dataDTO, nil
}

// This only used in handlers which returns multiple entities (by date,week,month)
func (handler *EventHTTPHandler) parseEventDataFromQueryParams(request *http.Request) (uint, time.Time, error) {
	userID, err := parseUserIDFromQueryParams(request)
	if err != nil {
		return 0, time.Time{}, err
	}
	date, err := parseDateFromQueryParams(request)
	if err != nil {
		return 0, time.Time{}, err
	}
	return userID, date, nil
}

func parseUserIDFromQueryParams(request *http.Request) (uint, error) {
	userIDStr := request.URL.Query().Get("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return 0, ErrInvalidUserID
	}
	return uint(userID), nil
}

func parseDateFromQueryParams(request *http.Request) (time.Time, error) {
	dateStr := request.URL.Query().Get("date")

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, ErrInvalidDate
	}
	return date, nil
}

func parseIDFromQueryParams(request *http.Request) (uint, error) {
	idStr := request.URL.Query().Get("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, ErrInvalidDate
	}
	return uint(id), nil
}

func (handler *EventHTTPHandler) respondSuccess(writer http.ResponseWriter, data any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(SuccessResponse{Result: data})
}

func (handler *EventHTTPHandler) respondError(writer http.ResponseWriter, message string, statusCode int) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	json.NewEncoder(writer).Encode(ErrorResponse{Error: message})
}
