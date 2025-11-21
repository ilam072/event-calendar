package service_test

import (
	"context"
	"errors"
	"github.com/ilam072/event-calendar/internal/user/mocks"
	"go.uber.org/mock/gomock"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ilam072/event-calendar/internal/types/domain"
	"github.com/ilam072/event-calendar/internal/types/dto"
	"github.com/ilam072/event-calendar/internal/user/repo"
	"github.com/ilam072/event-calendar/internal/user/service"
	"golang.org/x/crypto/bcrypt"
)

func TestUser_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepo(ctrl)
	tokenManager := mocks.NewMockTokenManager(ctrl)

	s := service.NewUser(userRepo, tokenManager, time.Second*10)

	ctx := context.Background()

	req := dto.RegisterUser{
		Email:    "test@mail.com",
		Password: "123456",
	}

	t.Run("success", func(t *testing.T) {
		userRepo.
			EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(uuid.New(), nil)

		_, err := s.Register(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("user exists", func(t *testing.T) {
		userRepo.
			EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(uuid.Nil, repo.ErrUserExists)

		_, err := s.Register(ctx, req)
		if !errors.Is(err, domain.ErrUserExists) {
			t.Fatalf("expected ErrUserExists, got %v", err)
		}
	})
}

func TestUser_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepo(ctrl)
	tokenManager := mocks.NewMockTokenManager(ctrl)

	s := service.NewUser(userRepo, tokenManager, time.Second*10)

	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	dbUser := domain.User{
		ID:           uuid.New(),
		Email:        "test@mail.com",
		PasswordHash: string(hash),
	}

	req := dto.LoginUser{
		Email:    "test@mail.com",
		Password: "123456",
	}

	t.Run("success", func(t *testing.T) {
		userRepo.
			EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(dbUser, nil)

		tokenManager.
			EXPECT().
			NewToken(dbUser.ID.String(), gomock.Any()).
			Return("TOKEN_123", nil)

		token, err := s.Login(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token != "TOKEN_123" {
			t.Fatalf("unexpected token: %s", token)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		userRepo.
			EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(domain.User{}, repo.ErrUserNotFound)

		_, err := s.Login(ctx, req)
		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		invalidDBUser := dbUser
		invalidDBUser.PasswordHash = "WRONG"

		userRepo.
			EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(invalidDBUser, nil)

		_, err := s.Login(ctx, req)
		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
		}
	})

	t.Run("token gen failed", func(t *testing.T) {
		userRepo.
			EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(dbUser, nil)

		tokenManager.
			EXPECT().
			NewToken(dbUser.ID.String(), gomock.Any()).
			Return("", errors.New("token error"))

		_, err := s.Login(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
