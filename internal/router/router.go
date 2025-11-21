package router

import (
	"github.com/gin-gonic/gin"
	eventrest "github.com/ilam072/event-calendar/internal/event/rest"
	"github.com/ilam072/event-calendar/internal/middlewares"
	userrest "github.com/ilam072/event-calendar/internal/user/rest"
	"github.com/ilam072/event-calendar/pkg/jwt"
)

func New(userHandler *userrest.UserHandler, eventHandler *eventrest.EventHandler, manager *jwt.Manager) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	auth := engine.Group("/auth")
	// user
	auth.POST("sign-up", userHandler.SignUp)
	auth.POST("sign-in", userHandler.SignIn)

	api := engine.Group("/api/v1", middlewares.Auth(manager))
	// event
	api.POST("/events", eventHandler.CreateEvent)
	api.GET("/events", eventHandler.GetEvents) // query ?period=day&date=2025-08-30
	api.PUT("/events/:id", eventHandler.UpdateEvent)
	api.DELETE("/events/:id", eventHandler.DeleteEvent)

	return engine
}
