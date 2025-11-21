package repo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/ilam072/event-calendar/internal/types/domain"
	"github.com/ilam072/event-calendar/pkg/errutils"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

var (
	ErrEventNotFound = errors.New("event not found")
)

type EventRepo struct {
	db *pgxpool.Pool
}

func NewEventRepo(db *pgxpool.Pool) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) CreateEvent(ctx context.Context, event domain.Event) (uuid.UUID, error) {
	query := `
		INSERT INTO events (user_id, event_date, description, remind_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	var ID uuid.UUID
	if err := r.db.QueryRow(ctx, query, event.UserID, event.Date, event.Description, event.RemindAt).Scan(&ID); err != nil {
		return uuid.Nil, errutils.Wrap("failed to create user", err)
	}

	return ID, nil

}

func (r *EventRepo) GetEventByID(ctx context.Context, eventID uuid.UUID) (domain.Event, error) {
	query := `
		SELECT id, user_id, event_date, description, remind_at, sent, created_at, updated_at
		FROM events
		WHERE id = $1;
	`

	var event domain.Event
	err := r.db.QueryRow(ctx, query, eventID).
		Scan(&event.ID, &event.UserID, &event.Date, &event.Description, &event.RemindAt, &event.Sent, &event.CreatedAt, &event.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Event{}, errutils.Wrap("failed to get event", ErrEventNotFound)
		}
		return domain.Event{}, errutils.Wrap("failed to get event", err)
	}

	return event, nil

}

func (r *EventRepo) UpdateEvent(ctx context.Context, event domain.Event) error {
	query := `
        UPDATE events
        SET event_date = $1,
        	description = $2,
        	remind_at = $3,
        	updated_at = now()
        WHERE id = $4 AND user_id = $5;
    `

	res, err := r.db.Exec(
		ctx,
		query,
		event.Date,
		event.Description,
		event.RemindAt,
		event.ID,
		event.UserID,
	)
	if err != nil {
		return errutils.Wrap("failed to update event", err)
	}

	if res.RowsAffected() == 0 {
		return ErrEventNotFound
	}

	return nil
}

func (r *EventRepo) DeleteEvent(ctx context.Context, eventID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1 AND user_id = $2;`

	res, err := r.db.Exec(ctx, query, eventID, userID)
	if err != nil {
		return errutils.Wrap("failed to delete event", err)
	}

	if res.RowsAffected() == 0 {
		return ErrEventNotFound
	}

	return nil
}

func (r *EventRepo) GetEventsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]domain.Event, error) {
	query := `
		SELECT 
		    id, 
		    user_id, 
		    event_date, 
		    description, 
		    remind_at, 
		    sent,
		    created_at,
		    updated_at
		FROM events
		WHERE user_id = $1 AND event_date = $2
	`

	rows, err := r.db.Query(ctx, query, userID, date)
	if err != nil {
		return nil, errutils.Wrap("failed to get events for day", err)
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(
			&event.ID,
			&event.UserID,
			&event.Date,
			&event.Description,
			&event.RemindAt,
			&event.Sent,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, errutils.Wrap("failed to scan", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventRepo) GetEventsForWeek(ctx context.Context, userID uuid.UUID, start time.Time) ([]domain.Event, error) {
	end := start.AddDate(0, 0, 7)

	query := `
		SELECT 
		    id, 
		    user_id, 
		    event_date, 
		    description, 
		    remind_at, 
		    sent,
		    created_at,
		    updated_at
		FROM events
		WHERE user_id = $1 AND event_date >= $2 AND event_date < $3
	`

	rows, err := r.db.Query(ctx, query, userID, start, end)
	if err != nil {
		return nil, errutils.Wrap("failed to get events for week", err)
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(
			&event.ID,
			&event.UserID,
			&event.Date,
			&event.Description,
			&event.RemindAt,
			&event.Sent,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, errutils.Wrap("failed to scan", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventRepo) GetEventsForMonth(ctx context.Context, userID uuid.UUID, start time.Time) ([]domain.Event, error) {
	end := start.AddDate(0, 1, 0)

	query := `
		SELECT 
		    id, 
		    user_id, 
		    event_date, 
		    description, 
		    remind_at, 
		    sent,
		    created_at,
		    updated_at
		FROM events
		WHERE user_id = $1 AND event_date >= $2 AND event_date < $3
	`

	rows, err := r.db.Query(ctx, query, userID, start, end)
	if err != nil {
		return nil, errutils.Wrap("failed to get events for month", err)
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(
			&event.ID,
			&event.UserID,
			&event.Date,
			&event.Description,
			&event.RemindAt,
			&event.Sent,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, errutils.Wrap("failed to scan", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventRepo) MarkReminderSent(ctx context.Context, eventID uuid.UUID) error {
	query := `UPDATE events SET sent = true, updated_at = now() WHERE id = $1;`
	if _, err := r.db.Exec(ctx, query, eventID); err != nil {
		return errutils.Wrap("failed to set sent to 'true'", err)
	}
	return nil
}

func (r *EventRepo) ArchiveOldEvents(ctx context.Context) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return errutils.Wrap("failed to begin tx", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	query := `
        INSERT INTO events_archive (id, user_id, event_date, description, archived_at, original_created_at, original_updated_at)
        SELECT id, user_id, event_date, description, NOW(), created_at, updated_at
        FROM events
        WHERE event_date < CURRENT_DATE;
    `
	if _, err = tx.Exec(ctx, query); err != nil {
		return errutils.Wrap("failed to archive events", err)
	}

	query = `
        DELETE FROM events 
        WHERE event_date < CURRENT_DATE;
    `
	if _, err = tx.Exec(ctx, query); err != nil {
		return errutils.Wrap("failed to delete old events", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return errutils.Wrap("failed to commit tx", err)
	}

	return nil
}
