package domain

import (
	"github.com/google/uuid"
	"time"
)

type Event struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Date        time.Time
	Description string
	Sent        bool
	RemindAt    *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
