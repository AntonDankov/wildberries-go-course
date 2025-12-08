package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type EventHTTPHandler struct {
	eventStorage *EventStorage
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Result any `json:"result"`
}

func NewEventHTTPHandler(eventStorage *EventStorage) *EventHTTPHandler {
	return &EventHTTPHandler{eventStorage: eventStorage}
}

func (handler *EventHTTPHandler) CreateEvent(writer http.ResponseWriter, request *http.Request) {
	_, userID, date, description, err := parseEventDataFromJSONRequest(request)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
	}
	event, err := handler.eventStorage.CreateEvent(userID, date, description)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
	}
	eventDTO := convertEventToDTO(*event)
	handler.respondSuccess(writer, eventDTO)
}

func (handler *EventHTTPHandler) GetEventByDate(writer http.ResponseWriter, request *http.Request) {
	userID, date, err := handler.parseEventDataFromQueryParams(request)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
	}
	events := handler.eventStorage.GetEventByDay(userID, date)
	eventDTOs := convertEventsToDTO(events)
	handler.respondSuccess(writer, eventDTOs)
}

func (handler *EventHTTPHandler) GetEventByWeek(writer http.ResponseWriter, request *http.Request) {
	userID, date, err := handler.parseEventDataFromQueryParams(request)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
	}
	events := handler.eventStorage.GetEventByWeek(userID, date)
	eventDTOs := convertEventsToDTO(events)
	handler.respondSuccess(writer, eventDTOs)
}

func (handler *EventHTTPHandler) GetEventByMonth(writer http.ResponseWriter, request *http.Request) {
	userID, date, err := handler.parseEventDataFromQueryParams(request)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
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
	handler.respondSuccess(writer, "event was deleted")
}

func (handler *EventHTTPHandler) UpdateEvent(writer http.ResponseWriter, request *http.Request) {
	id, userID, date, description, err := parseEventDataFromJSONRequest(request)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusBadRequest)
	}
	event, err := handler.eventStorage.UpdateEvent(userID, id, date, description)
	if err != nil {
		handler.respondError(writer, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if event == nil || event.ID == -1 {
		handler.respondError(writer, "", http.StatusNotFound)
		return
	}
	eventDTO := convertEventToDTO(*event)
	handler.respondSuccess(writer, eventDTO)
}

func parseEventDataFromJSONRequest(request *http.Request) (int, int, time.Time, string, error) {
	if request.Header.Get("Content-Type") != "application/json" {
		return -1, -1, time.Time{}, "", ErrInvalidContentType
	}
	var dataDTO EventDTO
	if err := json.NewDecoder(request.Body).Decode(&dataDTO); err != nil {
		return -1, -1, time.Time{}, "", err
	}
	date, err := time.Parse("2006-01-02", dataDTO.Date)
	if err != nil {
		return -1, -1, time.Time{}, "", ErrInvalidDate
	}
	if err := ValidateEvent(dataDTO.ID, dataDTO.UserID, date, dataDTO.Description); err != nil {
		return -1, -1, time.Time{}, "", err
	}
	return dataDTO.ID, dataDTO.UserID, date, dataDTO.Description, nil
}

func parseDeleteEventDataFromJSONRequest(request *http.Request) (DeleteEventDTO, error) {
	if request.Header.Get("Content-Type") != "application/json" {
		return DeleteEventDTO{}, ErrInvalidContentType
	}
	var dataDTO DeleteEventDTO
	if err := json.NewDecoder(request.Body).Decode(&dataDTO); err != nil {
		return DeleteEventDTO{}, err
	}
	if dataDTO.ID < 0 {
		return DeleteEventDTO{}, ErrInvalidID
	}
	if dataDTO.UserID < 0 {
		return DeleteEventDTO{}, ErrInvalidUserID
	}
	return dataDTO, nil
}

func (handler *EventHTTPHandler) parseEventDataFromQueryParams(request *http.Request) (int, time.Time, error) {
	userID, err := parseUserIDFromQueryParams(request)
	if err != nil {
		return -1, time.Time{}, err
	}
	date, err := parseDateFromQueryParams(request)
	if err != nil {
		return -1, time.Time{}, err
	}
	return userID, date, nil
}

func parseUserIDFromQueryParams(request *http.Request) (int, error) {
	userIDStr := request.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return -1, ErrInvalidUserID
	}
	return userID, nil
}

func parseDateFromQueryParams(request *http.Request) (time.Time, error) {
	dateStr := request.URL.Query().Get("date")

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, ErrInvalidDate
	}
	return date, nil
}

func parseIDFromQueryParams(request *http.Request) (int, error) {
	idStr := request.URL.Query().Get("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return -1, ErrInvalidDate
	}
	return id, nil
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
