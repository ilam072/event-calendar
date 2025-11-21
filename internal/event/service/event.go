package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ilam072/event-calendar/internal/event/repo"
	"github.com/ilam072/event-calendar/internal/event/worker/reminder"
	"github.com/ilam072/event-calendar/internal/types/domain"
	"github.com/ilam072/event-calendar/internal/types/dto"
	"github.com/ilam072/event-calendar/pkg/errutils"
	"time"
)

//go:generate mockgen -source=event.go -destination=../mocks/service_mocks.go -package=mocks
type EventRepo interface {
	CreateEvent(ctx context.Context, event domain.Event) (uuid.UUID, error)
	UpdateEvent(ctx context.Context, event domain.Event) error
	DeleteEvent(ctx context.Context, eventID uuid.UUID, userID uuid.UUID) error
	GetEventsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]domain.Event, error)
	GetEventsForWeek(ctx context.Context, userID uuid.UUID, start time.Time) ([]domain.Event, error)
	GetEventsForMonth(ctx context.Context, userID uuid.UUID, start time.Time) ([]domain.Event, error)
}

type Event struct {
	eventRepo EventRepo
	reminders chan<- reminder.Task
}

func NewEvent(repo EventRepo, reminderChan chan<- reminder.Task) *Event {
	return &Event{
		eventRepo: repo,
		reminders: reminderChan,
	}
}

func (e *Event) CreateEvent(ctx context.Context, event dto.CreateEventRequest, userID uuid.UUID) (uuid.UUID, error) {
	const op = "service.event.Create"

	domainEvent := domain.Event{
		UserID:      userID,
		Date:        event.Date,
		Description: event.Description,
		RemindAt:    event.RemindAt,
	}

	id, err := e.eventRepo.CreateEvent(ctx, domainEvent)
	if err != nil {
		return uuid.Nil, errutils.Wrap(op, err)
	}

	if event.RemindAt != nil && !event.RemindAt.IsZero() {
		e.reminders <- reminder.Task{
			EventID:  id,
			UserID:   userID,
			RemindAt: *event.RemindAt,
		}
	}

	return id, nil
}

func (e *Event) UpdateEvent(ctx context.Context, event dto.UpdateEventRequest, eventID uuid.UUID, userID uuid.UUID) error {
	const op = "service.event.Update"

	domainEvent := domain.Event{
		ID:          eventID,
		UserID:      userID,
		Date:        event.Date,
		Description: event.Description,
		RemindAt:    event.RemindAt,
	}

	if err := e.eventRepo.UpdateEvent(ctx, domainEvent); err != nil {
		if errors.Is(err, repo.ErrEventNotFound) {
			return errutils.Wrap(op, domain.ErrEventNotFound)
		}
		return errutils.Wrap(op, err)
	}

	return nil
}

func (e *Event) DeleteEvent(ctx context.Context, eventID uuid.UUID, userID uuid.UUID) error {
	const op = "service.event.Delete"

	if err := e.eventRepo.DeleteEvent(ctx, eventID, userID); err != nil {
		if errors.Is(err, repo.ErrEventNotFound) {
			return errutils.Wrap(op, domain.ErrEventNotFound)
		}
		return errutils.Wrap(op, err)
	}

	return nil
}

func (e *Event) GetEventsForDay(ctx context.Context, userID uuid.UUID, date time.Time) (dto.GetEventsResponse, error) {
	const op = "service.event.GetForDay"

	domainEvents, err := e.eventRepo.GetEventsForDay(ctx, userID, date)
	if err != nil {
		return dto.GetEventsResponse{}, errutils.Wrap(op, err)
	}

	return domainToGetEventsResponse(domainEvents), nil
}

func (e *Event) GetEventsForWeek(ctx context.Context, userID uuid.UUID, date time.Time) (dto.GetEventsResponse, error) {
	const op = "service.event.GetForWeek"

	domainEvents, err := e.eventRepo.GetEventsForWeek(ctx, userID, date)
	if err != nil {
		return dto.GetEventsResponse{}, errutils.Wrap(op, err)
	}

	return domainToGetEventsResponse(domainEvents), nil
}

func (e *Event) GetEventsForMonth(ctx context.Context, userID uuid.UUID, date time.Time) (dto.GetEventsResponse, error) {
	const op = "service.event.GetForMonth"

	domainEvents, err := e.eventRepo.GetEventsForMonth(ctx, userID, date)
	if err != nil {
		return dto.GetEventsResponse{}, errutils.Wrap(op, err)
	}

	return domainToGetEventsResponse(domainEvents), nil
}
