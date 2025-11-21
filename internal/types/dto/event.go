package dto

import (
	"github.com/google/uuid"
	"time"
)

type CreateEventRequest struct {
	Date        time.Time  `json:"date" validate:"required"`
	Description string     `json:"description" validate:"required,min=1,max=500"`
	RemindAt    *time.Time `json:"remind_at,omitempty"`
}

type UpdateEventRequest struct {
	Date        time.Time  `json:"date" validate:"required"`
	Description string     `json:"description" validate:"required,min=1,max=500"`
	RemindAt    *time.Time `json:"remind_at" validate:"required"`
}

type Event struct {
	ID          uuid.UUID `json:"event_id"`
	UserID      uuid.UUID `json:"user_id"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

type GetEventsResponse struct {
	Events []Event `json:"events"`
}
