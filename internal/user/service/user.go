package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ilam072/event-calendar/internal/types/domain"
	"github.com/ilam072/event-calendar/internal/types/dto"
	"github.com/ilam072/event-calendar/internal/user/repo"
	"github.com/ilam072/event-calendar/pkg/errutils"
	"golang.org/x/crypto/bcrypt"
	"time"
)

//go:generate mockgen -source=user.go -destination=../mocks/service_mocks.go -package=mocks
type UserRepo interface {
	CreateUser(ctx context.Context, user domain.User) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
}

type TokenManager interface {
	NewToken(userID string, ttl time.Duration) (string, error)
}

type User struct {
	repo     UserRepo
	manager  TokenManager
	tokenTTL time.Duration
}

func NewUser(repo UserRepo, manager TokenManager, tokenTTL time.Duration) *User {
	return &User{
		repo:     repo,
		manager:  manager,
		tokenTTL: tokenTTL,
	}
}

func (u *User) Register(ctx context.Context, user dto.RegisterUser) (string, error) {
	const op = "service.user.Register"

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errutils.Wrap(op, err)
	}

	domainUser := domain.User{
		Email:        user.Email,
		PasswordHash: string(passwordHash),
	}

	ID, err := u.repo.CreateUser(ctx, domainUser)
	if err != nil {
		if errors.Is(err, repo.ErrUserExists) {
			return "", errutils.Wrap(op, domain.ErrUserExists)
		}
		return "", errutils.Wrap(op, err)
	}

	return ID.String(), nil
}

func (u *User) Login(ctx context.Context, creds dto.LoginUser) (string, error) {
	const op = "service.user.Login"

	user, err := u.repo.GetUserByEmail(ctx, creds.Email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			return "", errutils.Wrap(op, domain.ErrInvalidCredentials)
		}
		return "", errutils.Wrap(op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		return "", errutils.Wrap(op, domain.ErrInvalidCredentials)
	}

	token, err := u.manager.NewToken(user.ID.String(), u.tokenTTL)
	if err != nil {
		return "", errutils.Wrap(op, err)
	}

	return token, nil
}
