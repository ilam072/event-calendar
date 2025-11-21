package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ilam072/event-calendar/internal/response"
	"github.com/ilam072/event-calendar/internal/types/domain"
	"github.com/ilam072/event-calendar/internal/types/dto"
	"github.com/ilam072/event-calendar/pkg/logger"
	"net/http"
	"strings"
	"time"
)

//go:generate mockgen -source=rest.go -destination=../mocks/rest_mocks.go -package=mocks
type Event interface {
	CreateEvent(ctx context.Context, event dto.CreateEventRequest, userID uuid.UUID) (uuid.UUID, error)
	UpdateEvent(ctx context.Context, event dto.UpdateEventRequest, eventID uuid.UUID, userID uuid.UUID) error
	DeleteEvent(ctx context.Context, eventID uuid.UUID, userID uuid.UUID) error
	GetEventsForDay(ctx context.Context, userID uuid.UUID, date time.Time) (dto.GetEventsResponse, error)
	GetEventsForWeek(ctx context.Context, userID uuid.UUID, date time.Time) (dto.GetEventsResponse, error)
	GetEventsForMonth(ctx context.Context, userID uuid.UUID, date time.Time) (dto.GetEventsResponse, error)
}

type Validator interface {
	Validate(i interface{}) error
}

type EventHandler struct {
	event     Event
	validator Validator
	logger    logger.Logger
}

func NewEventHandler(event Event, validator Validator, logger logger.Logger) *EventHandler {
	return &EventHandler{event: event, validator: validator, logger: logger}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var event dto.CreateEventRequest
	if err := c.BindJSON(&event); err != nil {
		h.logger.Warn().Err(err).Msg("failed to bind create event json")
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validator.Validate(event); err != nil {
		response.BadRequest(c, fmt.Sprintf("validation error: %s", err.Error()))
		return
	}

	userID, ok := h.getUserData(c)
	if !ok {
		return
	}

	eventID, err := h.event.CreateEvent(c.Request.Context(), event, userID)
	if err != nil {
		h.logger.Error().Err(err).Any("event", event).Msg("failed to create event")
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"event_id": eventID})
}

func (h *EventHandler) GetEvents(c *gin.Context) {
	period := c.Query("period")
	if period == "" {
		response.BadRequest(c, "empty query param 'period': must be 'day', 'week' or 'month'")
		return
	}

	date, err := time.Parse(time.DateOnly, c.Query("date"))
	if err != nil {
		response.BadRequest(c, "invalid query param 'date' format, must be YYYY-MM-DD")
		return
	}

	userID, ok := h.getUserData(c)
	if !ok {
		return
	}

	var events dto.GetEventsResponse
	switch strings.ToLower(period) {
	case "day":
		events, err = h.event.GetEventsForDay(c.Request.Context(), userID, date)
	case "week":
		events, err = h.event.GetEventsForWeek(c.Request.Context(), userID, date)
	case "month":
		events, err = h.event.GetEventsForMonth(c.Request.Context(), userID, date)
	default:
		response.BadRequest(c, "unexpected query param 'period': must be 'day', 'week' or 'month'")
		return
	}
	if err != nil {
		h.logger.Error().
			Err(err).
			Any("period", period).
			Str("date", date.String()).
			Str("user_id", userID.String()).
			Msg("failed to update event")
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, events)
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Warn().Err(err).Msg("failed to parse event id into uuid")
		response.BadRequest(c, "event id must be UUID format")
		return
	}

	var event dto.UpdateEventRequest
	if err = c.BindJSON(&event); err != nil {
		h.logger.Warn().Err(err).Msg("failed to bind update event json")
		response.BadRequest(c, "invalid request body")
		return
	}

	if err = h.validator.Validate(event); err != nil {
		response.BadRequest(c, fmt.Sprintf("validation error: %s", err.Error()))
		return
	}

	userID, ok := h.getUserData(c)
	if !ok {
		return
	}

	if err = h.event.UpdateEvent(c.Request.Context(), event, eventID, userID); err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			response.NotFound(c)
			return
		}
		h.logger.Error().Err(err).Any("event", event).Str("event_id", eventID.String()).Msg("failed to update event")
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusOK)
}

func (h *EventHandler) DeleteEvent(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Warn().Err(err).Msg("failed to parse event id into uuid")
		response.BadRequest(c, "event id must be UUID format")
		return
	}

	userID, ok := h.getUserData(c)
	if !ok {
		return
	}

	if err = h.event.DeleteEvent(c.Request.Context(), eventID, userID); err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			response.NotFound(c)
			return
		}
		h.logger.Error().Err(err).Str("event_id", eventID.String()).Msg("failed to delete event")
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusOK)
}

func (h *EventHandler) getUserData(c *gin.Context) (uuid.UUID, bool) {
	userID, ok := c.Get("user_id")
	if !ok {
		h.logger.Error().Msg("failed to get user_id from ctx")
		response.Unauthorized(c, "missing user_id in token")
		return uuid.Nil, false
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Warn().Any("user_id", userID).Msg("failed to parse user id into uuid")
		response.Unauthorized(c, "user_id must be uuid format")
		return uuid.Nil, false
	}

	return userUUID, true
}
