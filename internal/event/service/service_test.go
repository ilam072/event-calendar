package service_test

import (
	"context"
	"errors"
	"go.uber.org/mock/gomock"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ilam072/event-calendar/internal/event/mocks"
	"github.com/ilam072/event-calendar/internal/event/repo"
	"github.com/ilam072/event-calendar/internal/event/service"
	"github.com/ilam072/event-calendar/internal/event/worker/reminder"
	"github.com/ilam072/event-calendar/internal/types/domain"
	"github.com/ilam072/event-calendar/internal/types/dto"
)

func TestCreateEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)

	reminderChan := make(chan reminder.Task, 1)

	svc := service.NewEvent(mockRepo, reminderChan)

	userID := uuid.New()
	eventID := uuid.New()
	date := time.Now()
	remindAt := time.Now().Add(10 * time.Minute)

	req := dto.CreateEventRequest{
		Date:        date,
		Description: "Test",
		RemindAt:    &remindAt,
	}

	mockRepo.
		EXPECT().
		CreateEvent(gomock.Any(), domain.Event{
			UserID:      userID,
			Date:        date,
			Description: "Test",
			RemindAt:    &remindAt,
		}).
		Return(eventID, nil)

	id, err := svc.CreateEvent(context.Background(), req, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != eventID {
		t.Fatalf("expected %s, got %s", eventID, id)
	}

	select {
	case task := <-reminderChan:
		if task.EventID != eventID || task.UserID != userID {
			t.Fatalf("wrong reminder task sent")
		}
	default:
		t.Fatalf("reminder task was not sent")
	}
}

func TestCreateEvent_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)
	reminderChan := make(chan reminder.Task, 1)
	svc := service.NewEvent(mockRepo, reminderChan)

	userID := uuid.New()
	req := dto.CreateEventRequest{
		Date:        time.Now(),
		Description: "Test",
	}

	mockRepo.
		EXPECT().
		CreateEvent(gomock.Any(), gomock.Any()).
		Return(uuid.Nil, errors.New("db failure"))

	_, err := svc.CreateEvent(context.Background(), req, userID)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestUpdateEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)
	reminderChan := make(chan reminder.Task)
	svc := service.NewEvent(mockRepo, reminderChan)

	req := dto.UpdateEventRequest{
		Date:        time.Now(),
		Description: "Updated",
	}

	eventID := uuid.New()
	userID := uuid.New()

	expected := domain.Event{
		ID:          eventID,
		UserID:      userID,
		Date:        req.Date,
		Description: req.Description,
		RemindAt:    req.RemindAt,
	}

	mockRepo.
		EXPECT().
		UpdateEvent(gomock.Any(), expected).
		Return(nil)

	err := svc.UpdateEvent(context.Background(), req, eventID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateEvent_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)
	reminderChan := make(chan reminder.Task)
	svc := service.NewEvent(mockRepo, reminderChan)

	mockRepo.
		EXPECT().
		UpdateEvent(gomock.Any(), gomock.Any()).
		Return(repo.ErrEventNotFound)

	err := svc.UpdateEvent(context.Background(), dto.UpdateEventRequest{}, uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrEventNotFound) {
		t.Fatalf("expected domain.ErrEventNotFound, got %v", err)
	}
}

func TestDeleteEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)
	reminderChan := make(chan reminder.Task)
	svc := service.NewEvent(mockRepo, reminderChan)

	eventID := uuid.New()
	userID := uuid.New()

	mockRepo.
		EXPECT().
		DeleteEvent(gomock.Any(), eventID, userID).
		Return(nil)

	err := svc.DeleteEvent(context.Background(), eventID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteEvent_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)
	reminderChan := make(chan reminder.Task)
	svc := service.NewEvent(mockRepo, reminderChan)

	mockRepo.
		EXPECT().
		DeleteEvent(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(repo.ErrEventNotFound)

	err := svc.DeleteEvent(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestGetEventsForDay(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)
	reminderChan := make(chan reminder.Task)
	svc := service.NewEvent(mockRepo, reminderChan)

	userID := uuid.New()
	date := time.Now()

	events := []domain.Event{
		{ID: uuid.New(), UserID: userID, Description: "A"},
	}

	mockRepo.
		EXPECT().
		GetEventsForDay(gomock.Any(), userID, date).
		Return(events, nil)

	resp, err := svc.GetEventsForDay(context.Background(), userID, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(resp.Events))
	}
}

func TestGetEventsForWeek(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)
	svc := service.NewEvent(mockRepo, make(chan reminder.Task))

	userID := uuid.New()
	date := time.Now()

	mockRepo.
		EXPECT().
		GetEventsForWeek(gomock.Any(), userID, date).
		Return([]domain.Event{}, nil)

	_, err := svc.GetEventsForWeek(context.Background(), userID, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetEventsForMonth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)
	svc := service.NewEvent(mockRepo, make(chan reminder.Task))

	userID := uuid.New()
	date := time.Now()

	mockRepo.
		EXPECT().
		GetEventsForMonth(gomock.Any(), userID, date).
		Return([]domain.Event{}, nil)

	_, err := svc.GetEventsForMonth(context.Background(), userID, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
