package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ilam072/event-calendar/internal/response"
	"github.com/ilam072/event-calendar/internal/types/domain"
	"github.com/ilam072/event-calendar/internal/types/dto"
	"github.com/ilam072/event-calendar/pkg/logger"
	"net/http"
)

//go:generate mockgen -source=rest.go -destination=../mocks/rest_mocks.go -package=mocks
type User interface {
	Register(ctx context.Context, user dto.RegisterUser) (string, error)
	Login(ctx context.Context, creds dto.LoginUser) (string, error)
}

type Validator interface {
	Validate(i interface{}) error
}

type UserHandler struct {
	user      User
	validator Validator
	logger    logger.Logger
}

func NewUserHandler(user User, validator Validator, logger logger.Logger) *UserHandler {
	return &UserHandler{user: user, validator: validator, logger: logger}
}

func (h *UserHandler) SignUp(c *gin.Context) {
	var user dto.RegisterUser
	if err := c.BindJSON(&user); err != nil {
		h.logger.Warn().Err(err).Msg("failed to bind register user json")
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validator.Validate(user); err != nil {
		response.BadRequest(c, fmt.Sprintf("validation error: %s", err.Error()))
		return
	}

	ID, err := h.user.Register(c.Request.Context(), user)
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			response.Conflict(c, "USER_EXISTS", "user with such email already exists")
			return
		}
		h.logger.Error().Err(err).Any("user", user).Msg("failed to register user")
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user_id": ID})
}

func (h *UserHandler) SignIn(c *gin.Context) {
	var user dto.LoginUser
	if err := c.BindJSON(&user); err != nil {
		h.logger.Warn().Err(err).Msg("failed to bind register user json")
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validator.Validate(user); err != nil {
		response.BadRequest(c, fmt.Sprintf("validation error: %s", err.Error()))
		return
	}

	token, err := h.user.Login(c.Request.Context(), user)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			response.Unauthorized(c, "invalid credentials")
			return
		}
		h.logger.Error().Err(err).Any("user", user).Msg("failed to login user")
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
