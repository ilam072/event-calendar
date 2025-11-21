package service

import (
	"github.com/ilam072/event-calendar/internal/types/domain"
	"github.com/ilam072/event-calendar/internal/types/dto"
)

func domainToGetEventsResponse(domainEvents []domain.Event) dto.GetEventsResponse {
	events := make([]dto.Event, 0, len(domainEvents))
	for _, e := range domainEvents {
		events = append(events, dto.Event{
			ID:          e.ID,
			UserID:      e.UserID,
			Date:        e.Date,
			Description: e.Description,
		})
	}

	return dto.GetEventsResponse{
		Events: events,
	}
}
