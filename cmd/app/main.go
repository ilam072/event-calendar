package main

import (
	"context"
	"github.com/ilam072/event-calendar/internal/config"
	eventrepo "github.com/ilam072/event-calendar/internal/event/repo"
	eventrest "github.com/ilam072/event-calendar/internal/event/rest"
	eventservice "github.com/ilam072/event-calendar/internal/event/service"
	"github.com/ilam072/event-calendar/internal/event/worker/janitor"
	"github.com/ilam072/event-calendar/internal/event/worker/reminder"
	"github.com/ilam072/event-calendar/internal/router"
	userrepo "github.com/ilam072/event-calendar/internal/user/repo"
	userrest "github.com/ilam072/event-calendar/internal/user/rest"
	userservice "github.com/ilam072/event-calendar/internal/user/service"
	"github.com/ilam072/event-calendar/internal/validator"
	"github.com/ilam072/event-calendar/pkg/db"
	"github.com/ilam072/event-calendar/pkg/email"
	"github.com/ilam072/event-calendar/pkg/jwt"
	"github.com/ilam072/event-calendar/pkg/logger"
	"github.com/rs/zerolog/log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Initialize config
	cfg := config.MustLoad()

	// Initialize logger
	asyncLog, err := logger.NewAsyncLogger(cfg.Logger.File, 1000)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to initialize async logger")
	}
	asyncLog.Start()
	defer asyncLog.Stop()

	// Context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Connect to DB
	DB, err := db.OpenDB(ctx, cfg.DB)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to connect to DB")
	}

	// Initialize token manager
	manager := jwt.NewManager([]byte(cfg.JWT.Secret))

	// Initialize validator
	v := validator.New()

	// Initialize email client
	emailClient := email.New(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.From)

	// Initialize user and event repositories
	userRepo := userrepo.NewUserRepo(DB)
	eventRepo := eventrepo.NewEventRepo(DB)

	// Initialize reminder worker
	reminderWorker := reminder.NewWorker(eventRepo, userRepo, emailClient, 100)
	go reminderWorker.Run(ctx)

	// Initialize janitor worker
	janitorWorker := janitor.NewWorker(eventRepo)
	go janitorWorker.Start()

	// Initialize user and event services
	user := userservice.NewUser(userRepo, manager, cfg.JWT.TokenTTL)
	event := eventservice.NewEvent(eventRepo, reminderWorker.TasksChan())

	// Initialize user and event handlers
	userHandler := userrest.NewUserHandler(user, v, asyncLog)
	eventHandler := eventrest.NewEventHandler(event, v, asyncLog)

	// Initialize Gin engine and set routes
	engine := router.New(userHandler, eventHandler, manager)

	// Initialize and start http server
	server := &http.Server{
		Addr:    cfg.Server.HTTPPort,
		Handler: engine,
	}

	go func() {
		if err = server.ListenAndServe(); err != nil {
			log.Logger.Fatal().Err(err).Msg("failed to listen start http server")
		}
	}()

	<-ctx.Done()

	// Graceful shutdown
	withTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err = server.Shutdown(withTimeout); err != nil {
		log.Logger.Error().Err(err).Msg("server shutdown failed")
	}

	DB.Close()

	janitorWorker.Stop()

	reminderWorker.Stop()
}
