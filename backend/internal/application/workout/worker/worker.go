package worker

import (
	"context"
	"log/slog"
	"time"
)

type Job interface {
	Execute(ctx context.Context) error
}

type Worker struct {
	job      Job
	interval time.Duration
	log      *slog.Logger
}

func New(job Job, interval time.Duration, log *slog.Logger) *Worker {
	if interval <= 0 {
		interval = time.Minute
	}
	return &Worker{job: job, interval: interval, log: log}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.job.Execute(ctx); err != nil {
				w.log.Error("workout worker job failed", slog.String("err", err.Error()))
			}
		}
	}
}
