package janitor

import (
	"context"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"time"
)

type EventRepo interface {
	ArchiveOldEvents(ctx context.Context) error
}

type Worker struct {
	cron      *cron.Cron
	eventRepo EventRepo
}

func NewWorker(eventRepo EventRepo) *Worker {
	c := cron.New(cron.WithSeconds())
	return &Worker{cron: c, eventRepo: eventRepo}
}

func (w *Worker) RegisterJobs() {
	schedule := "0 */5 * * * *"

	if _, err := w.cron.AddFunc(schedule, func() {
		log.Logger.Print("[JOB] Archiving old events...\n")

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if err := w.eventRepo.ArchiveOldEvents(ctx); err != nil {
			log.Logger.Error().Err(err).Msg("[JOB] ArchiveOldEvents failed")
		}
	}); err != nil {
		log.Logger.Error().Err(err).Msg("[CRON] Failed to register ArchiveOldEvents job")
	} else {
		log.Logger.Print("[CRON] ArchiveOldEvents job registered successfully")
	}
}

func (w *Worker) Start() {
	w.RegisterJobs()
	w.cron.Start()
	log.Logger.Println("[CRON] Worker started")
}

func (w *Worker) Stop() {
	log.Logger.Println("[CRON] Stopping scheduler...")
	ctx := w.cron.Stop()
	<-ctx.Done()
}
