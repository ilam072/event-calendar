package reminder

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"time"

	"github.com/ilam072/event-calendar/internal/types/domain"
)

type EventRepo interface {
	GetEventByID(ctx context.Context, eventID uuid.UUID) (domain.Event, error)
	MarkReminderSent(ctx context.Context, eventID uuid.UUID) error
}

type UserRepo interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (domain.User, error)
}

type Sender interface {
	Send(subject string, message string, to string) error
}

type Task struct {
	EventID  uuid.UUID
	UserID   uuid.UUID
	RemindAt time.Time
}

type Worker struct {
	tasks     chan Task
	eventRepo EventRepo
	userRepo  UserRepo
	sender    Sender
	done      chan struct{}
}

func NewWorker(eventRepo EventRepo, userRepo UserRepo, sender Sender, buffer int) *Worker {
	return &Worker{
		tasks:     make(chan Task, buffer),
		eventRepo: eventRepo,
		userRepo:  userRepo,
		sender:    sender,
		done:      make(chan struct{}),
	}
}

func (w *Worker) TasksChan() chan<- Task {
	return w.tasks
}

func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case task, ok := <-w.tasks:
			if !ok {
				log.Info().Msg("Reminder worker stopped, channel closed")
				close(w.done)
				return
			}
			go w.handleTask(ctx, task)

		case <-ctx.Done():
			log.Info().Msg("Reminder worker stopped by context")
			close(w.done)
			return
		}
	}
}

func (w *Worker) handleTask(ctx context.Context, task Task) {
	delay := time.Until(task.RemindAt)

	log.Logger.Info().
		Str("event_id", task.EventID.String()).
		Str("user_id", task.UserID.String()).
		Str("remind_at", task.RemindAt.String()).
		Str("delay", time.Until(task.RemindAt).String()).
		Msg("Sending reminder...")

	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return
		}
	}

	log.Logger.Info().
		Str("event_id", task.EventID.String()).
		Str("user_id", task.UserID.String()).
		Msg("Sending reminder...")

	user, err := w.userRepo.GetUserByID(ctx, task.UserID)
	if err != nil {
		log.Error().Err(err).Str("op", "handleTask").Msg("failed to get user by id")
		return
	}

	event, err := w.eventRepo.GetEventByID(ctx, task.EventID)
	if err != nil {
		log.Error().Err(err).Str("op", "handleTask").Msg("failed to get event by id")
		return
	}

	message := fmt.Sprintf(`Event "%s" is coming up soon. 🔔`, event.Description)
	if err := w.sender.Send("Event reminder", message, user.Email); err != nil {
		log.Error().Err(err).Str("op", "handleTask").Msg("failed to send reminder")
		return
	}

	if err := w.eventRepo.MarkReminderSent(ctx, task.EventID); err != nil {
		log.Error().Err(err).Msg("failed to mark reminder sent")
	}
}

func (w *Worker) Stop() {
	close(w.tasks)
	<-w.done
	log.Info().Msg("Reminder worker fully stopped")
}
