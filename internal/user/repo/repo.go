package repo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/ilam072/event-calendar/internal/types/domain"
	"github.com/ilam072/event-calendar/pkg/errutils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

var (
	ErrUserExists   = errors.New("user exists")
	ErrUserNotFound = errors.New("user not found")
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateUser(ctx context.Context, user domain.User) (uuid.UUID, error) {
	query := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id;
	`

	var ID uuid.UUID
	if err := r.db.QueryRow(ctx, query, user.Email, user.PasswordHash).Scan(&ID); err != nil {
		if isUniqueViolation(err) {
			return uuid.Nil, errutils.Wrap("failed to create user", ErrUserExists)
		}
		return uuid.Nil, errutils.Wrap("failed to create user", err)
	}

	return ID, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1;
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, userID).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errutils.Wrap("failed to get user", ErrUserNotFound)
		}
		return domain.User{}, errutils.Wrap("failed to get user", err)
	}

	return user, nil
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1;
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errutils.Wrap("failed to get user", ErrUserNotFound)
		}
		return domain.User{}, errutils.Wrap("failed to get user", err)
	}

	return user, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
