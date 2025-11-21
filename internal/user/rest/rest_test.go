package rest_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/ilam072/event-calendar/pkg/logger"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ilam072/event-calendar/internal/types/domain"
	"github.com/ilam072/event-calendar/internal/types/dto"
	"github.com/ilam072/event-calendar/internal/user/rest"

	"github.com/ilam072/event-calendar/internal/user/mocks"
)

func TestUserHandler_SignUp(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := mocks.NewMockUser(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)

	logStub := &logger.DummyLogger{}

	h := rest.NewUserHandler(mockUser, mockValidator, logStub)

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		body := `{"email":"test@mail.com","password":"123456"}`
		ctx.Request = httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(body))
		ctx.Request.Header.Set("Content-Type", "application/json")

		req := dto.RegisterUser{Email: "test@mail.com", Password: "123456"}

		mockValidator.EXPECT().Validate(req).Return(nil)
		mockUser.EXPECT().Register(context.Background(), req).Return("user-123", nil)

		h.SignUp(ctx)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", w.Code)
		}
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString("INVALID JSON"))
		ctx.Request.Header.Set("Content-Type", "application/json")

		h.SignUp(ctx)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		body := `{"email":"test@mail.com","password":""}`
		ctx.Request = httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(body))
		ctx.Request.Header.Set("Content-Type", "application/json")

		req := dto.RegisterUser{Email: "test@mail.com"}

		mockValidator.EXPECT().Validate(req).Return(errors.New("password required"))

		h.SignUp(ctx)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("user exists", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		body := `{"email":"exists@mail.com","password":"123456"}`
		ctx.Request = httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(body))
		ctx.Request.Header.Set("Content-Type", "application/json")

		req := dto.RegisterUser{Email: "exists@mail.com", Password: "123456"}

		mockValidator.EXPECT().Validate(req).Return(nil)
		mockUser.EXPECT().Register(context.Background(), req).Return("", domain.ErrUserExists)

		h.SignUp(ctx)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		body := `{"email":"error@mail.com","password":"123456"}`
		ctx.Request = httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(body))
		ctx.Request.Header.Set("Content-Type", "application/json")

		req := dto.RegisterUser{Email: "error@mail.com", Password: "123456"}

		mockValidator.EXPECT().Validate(req).Return(nil)
		mockUser.EXPECT().Register(context.Background(), req).Return("", errors.New("db error"))

		h.SignUp(ctx)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestUserHandler_SignIn(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := mocks.NewMockUser(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)

	logStub := &logger.DummyLogger{}

	h := rest.NewUserHandler(mockUser, mockValidator, logStub)

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		body := `{"email":"test@mail.com","password":"123456"}`
		ctx.Request = httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBufferString(body))
		ctx.Request.Header.Set("Content-Type", "application/json")

		req := dto.LoginUser{Email: "test@mail.com", Password: "123456"}

		mockValidator.EXPECT().Validate(req).Return(nil)
		mockUser.EXPECT().Login(context.Background(), req).Return("TOKEN_123", nil)

		h.SignIn(ctx)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBufferString("BAD"))
		ctx.Request.Header.Set("Content-Type", "application/json")

		h.SignIn(ctx)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		body := `{"email":"test@mail.com","password":""}`
		ctx.Request = httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBufferString(body))
		ctx.Request.Header.Set("Content-Type", "application/json")

		req := dto.LoginUser{Email: "test@mail.com"}

		mockValidator.EXPECT().Validate(req).Return(errors.New("password required"))

		h.SignIn(ctx)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		body := `{"email":"test@mail.com","password":"wrong"}`
		ctx.Request = httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBufferString(body))
		ctx.Request.Header.Set("Content-Type", "application/json")

		req := dto.LoginUser{Email: "test@mail.com", Password: "wrong"}

		mockValidator.EXPECT().Validate(req).Return(nil)
		mockUser.EXPECT().Login(context.Background(), req).Return("", domain.ErrInvalidCredentials)

		h.SignIn(ctx)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		body := `{"email":"test@mail.com","password":"123456"}`
		ctx.Request = httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBufferString(body))
		ctx.Request.Header.Set("Content-Type", "application/json")

		req := dto.LoginUser{Email: "test@mail.com", Password: "123456"}

		mockValidator.EXPECT().Validate(req).Return(nil)
		mockUser.EXPECT().Login(context.Background(), req).Return("", errors.New("db error"))

		h.SignIn(ctx)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}
