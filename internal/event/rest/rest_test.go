package rest_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ilam072/event-calendar/internal/event/mocks"
	"github.com/ilam072/event-calendar/internal/event/rest"
	"github.com/ilam072/event-calendar/internal/types/domain"
	"github.com/ilam072/event-calendar/internal/types/dto"
	"github.com/ilam072/event-calendar/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

var log = &logger.DummyLogger{}

//
// --------------------------------------------------------------------------------------------
// CreateEvent
// --------------------------------------------------------------------------------------------

func TestCreateEvent_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := rest.NewEventHandler(
		mocks.NewMockEvent(ctrl),
		mocks.NewMockValidator(ctrl),
		log,
	)
	r := routerWithHandler(h)

	req := httptest.NewRequest("POST", "/event", bytes.NewBufferString("bad"))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateEvent_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockValidator := mocks.NewMockValidator(ctrl)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(errors.New("validation failed"))

	h := rest.NewEventHandler(
		mocks.NewMockEvent(ctrl),
		mockValidator,
		log,
	)
	r := routerWithHandler(h)

	body := `{"date":"2025-01-01T00:00:00Z","description":"x"}`
	req := httptest.NewRequest("POST", "/event", bytes.NewBufferString(body))
	req = addUserID(req)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateEvent_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockValidator := mocks.NewMockValidator(ctrl)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	h := rest.NewEventHandler(
		mocks.NewMockEvent(ctrl),
		mockValidator,
		log,
	)
	r := routerWithHandler(h)

	body := `{"date":"2025-01-01T00:00:00Z","description":"x"}`
	req := httptest.NewRequest("POST", "/event", bytes.NewBufferString(body))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCreateEvent_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvent := mocks.NewMockEvent(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)

	userID := uuid.New()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockEvent.EXPECT().
		CreateEvent(gomock.Any(), gomock.Any(), userID).
		Return(uuid.New(), nil)

	h := rest.NewEventHandler(mockEvent, mockValidator, log)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID.String())
	})
	r.POST("/event", h.CreateEvent)

	body := `{"date":"2025-01-01T00:00:00Z","description":"test"}`
	req := httptest.NewRequest("POST", "/event", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateEvent_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvent := mocks.NewMockEvent(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)

	userID := uuid.New()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockEvent.EXPECT().
		CreateEvent(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(uuid.Nil, errors.New("boom"))

	h := rest.NewEventHandler(mockEvent, mockValidator, log)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID.String())
	})
	r.POST("/event", h.CreateEvent)

	body := `{"date":"2025-01-01T00:00:00Z","description":"test"}`
	req := httptest.NewRequest("POST", "/event", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

//
// --------------------------------------------------------------------------------------------
// GetEvents
// --------------------------------------------------------------------------------------------

func TestGetEvents_EmptyPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := rest.NewEventHandler(
		mocks.NewMockEvent(ctrl),
		mocks.NewMockValidator(ctrl),
		log,
	)
	r := routerWithHandler(h)

	req := httptest.NewRequest("GET", "/event", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetEvents_InvalidDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := rest.NewEventHandler(
		mocks.NewMockEvent(ctrl),
		mocks.NewMockValidator(ctrl),
		log,
	)
	r := routerWithHandler(h)

	req := httptest.NewRequest("GET", "/event?period=day&date=bad", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetEvents_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := rest.NewEventHandler(
		mocks.NewMockEvent(ctrl),
		mocks.NewMockValidator(ctrl),
		log,
	)
	r := routerWithHandler(h)

	req := httptest.NewRequest("GET", "/event?period=day&date=2025-01-01", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetEvents_InvalidPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvent := mocks.NewMockEvent(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)

	h := rest.NewEventHandler(mockEvent, mockValidator, log)

	r := gin.New()
	r.GET("/event", func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		h.GetEvents(c)
	})

	req := httptest.NewRequest("GET", "/event?period=year&date=2025-01-01", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetEvents_Success_Day(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvent := mocks.NewMockEvent(ctrl)
	mockEvent.EXPECT().
		GetEventsForDay(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(dto.GetEventsResponse{}, nil)

	h := rest.NewEventHandler(mockEvent, mocks.NewMockValidator(ctrl), log)
	r := gin.New()
	r.GET("/event", func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		h.GetEvents(c)
	})

	req := httptest.NewRequest("GET", "/event?period=day&date=2025-01-01", nil)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetEvents_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvent := mocks.NewMockEvent(ctrl)
	mockEvent.EXPECT().
		GetEventsForDay(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(dto.GetEventsResponse{}, errors.New("boom"))

	h := rest.NewEventHandler(mockEvent, mocks.NewMockValidator(ctrl), log)

	r := gin.New()
	r.GET("/event", func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		h.GetEvents(c)
	})

	req := httptest.NewRequest("GET", "/event?period=day&date=2025-01-01", nil)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

//
// --------------------------------------------------------------------------------------------
// UpdateEvent
// --------------------------------------------------------------------------------------------

func TestUpdateEvent_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := rest.NewEventHandler(
		mocks.NewMockEvent(ctrl),
		mocks.NewMockValidator(ctrl),
		log,
	)
	r := routerWithHandler(h)

	req := httptest.NewRequest("PUT", "/event/invalid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateEvent_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := rest.NewEventHandler(
		mocks.NewMockEvent(ctrl),
		mocks.NewMockValidator(ctrl),
		log,
	)
	r := routerWithHandler(h)

	id := uuid.New().String()
	req := httptest.NewRequest("PUT", "/event/"+id, bytes.NewBufferString("bad"))
	req = addUserID(req)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateEvent_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockValidator := mocks.NewMockValidator(ctrl)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(errors.New("bad"))

	h := rest.NewEventHandler(
		mocks.NewMockEvent(ctrl),
		mockValidator,
		log,
	)
	r := routerWithHandler(h)

	id := uuid.New().String()
	body := `{"date":"2025-01-01T00:00:00Z","description":"x","remind_at":"2025-01-02T00:00:00Z"}`

	req := httptest.NewRequest("PUT", "/event/"+id, bytes.NewBufferString(body))
	req = addUserID(req)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateEvent_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvent := mocks.NewMockEvent(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)

	eventID := uuid.New()
	userID := uuid.New()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockEvent.EXPECT().
		UpdateEvent(gomock.Any(), gomock.Any(), eventID, userID).
		Return(domain.ErrEventNotFound)

	h := rest.NewEventHandler(mockEvent, mockValidator, log)

	r := gin.New()
	r.PUT("/event/:id", func(c *gin.Context) {
		c.Set("user_id", userID.String())
		h.UpdateEvent(c)
	})

	body := `{"date":"2025-01-01T00:00:00Z","description":"x","remind_at":"2025-01-02T00:00:00Z"}`
	req := httptest.NewRequest("PUT", "/event/"+eventID.String(), bytes.NewBufferString(body))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateEvent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvent := mocks.NewMockEvent(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockEvent.EXPECT().
		UpdateEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	h := rest.NewEventHandler(mockEvent, mockValidator, log)
	r := gin.New()
	r.PUT("/event/:id", func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		h.UpdateEvent(c)
	})

	body := `{"date":"2025-01-01T00:00:00Z","description":"x","remind_at":"2025-01-02T00:00:00Z"}`
	req := httptest.NewRequest("PUT", "/event/"+uuid.New().String(), bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdateEvent_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvent := mocks.NewMockEvent(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockEvent.EXPECT().
		UpdateEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("boom"))

	h := rest.NewEventHandler(mockEvent, mockValidator, log)
	r := gin.New()
	r.PUT("/event/:id", func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		h.UpdateEvent(c)
	})

	body := `{"date":"2025-01-01T00:00:00Z","description":"x","remind_at":"2025-01-02T00:00:00Z"}`
	req := httptest.NewRequest("PUT", "/event/"+uuid.New().String(), bytes.NewBufferString(body))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

//
// --------------------------------------------------------------------------------------------
// DeleteEvent
// --------------------------------------------------------------------------------------------

func TestDeleteEvent_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := rest.NewEventHandler(
		mocks.NewMockEvent(ctrl),
		mocks.NewMockValidator(ctrl),
		log,
	)
	r := routerWithHandler(h)

	req := httptest.NewRequest("DELETE", "/event/invalid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleteEvent_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvent := mocks.NewMockEvent(ctrl)
	mockEvent.EXPECT().
		DeleteEvent(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.ErrEventNotFound)

	h := rest.NewEventHandler(mockEvent, mocks.NewMockValidator(ctrl), log)
	r := gin.New()
	r.DELETE("/event/:id", func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		h.DeleteEvent(c)
	})

	req := httptest.NewRequest("DELETE", "/event/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDeleteEvent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvent := mocks.NewMockEvent(ctrl)
	mockEvent.EXPECT().
		DeleteEvent(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	h := rest.NewEventHandler(mockEvent, mocks.NewMockValidator(ctrl), log)
	r := gin.New()
	r.DELETE("/event/:id", func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		h.DeleteEvent(c)
	})

	req := httptest.NewRequest("DELETE", "/event/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDeleteEvent_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvent := mocks.NewMockEvent(ctrl)
	mockEvent.EXPECT().
		DeleteEvent(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("boom"))

	h := rest.NewEventHandler(mockEvent, mocks.NewMockValidator(ctrl), log)
	r := gin.New()
	r.DELETE("/event/:id", func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		h.DeleteEvent(c)
	})

	req := httptest.NewRequest("DELETE", "/event/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func addUserID(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), "user_id", uuid.New().String())
	return req.WithContext(ctx)
}

func routerWithHandler(h *rest.EventHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/event", h.CreateEvent)
	r.GET("/event", h.GetEvents)
	r.PUT("/event/:id", h.UpdateEvent)
	r.DELETE("/event/:id", h.DeleteEvent)

	return r
}
