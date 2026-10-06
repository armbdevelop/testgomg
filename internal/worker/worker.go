package worker

import (
	"context"
	"log"
	"time"

	"github.com/armbdevelop/testgomg/internal/core/domain"
	"github.com/armbdevelop/testgomg/internal/core/ports"
)

// taskTimeout ограничивает обработку одной задачи: зависшая БД или Redis
// не повесит единственного воркера навсегда.
const taskTimeout = 30 * time.Second

type Worker struct {
	pgRepo    ports.PGRepository
	cacheRepo ports.CacheRepository
	queue     ports.TaskQueue
}

func New(pgRepo ports.PGRepository, cacheRepo ports.CacheRepository, queue ports.TaskQueue) *Worker {
	return &Worker{pgRepo: pgRepo, cacheRepo: cacheRepo, queue: queue}
}

func (w *Worker) Run(ctx context.Context) {
	for {
		task, err := w.queue.Dequeue(ctx)
		if err != nil {
			return
		}

		taskCtx, cancel := context.WithTimeout(ctx, taskTimeout)
		w.process(taskCtx, task)
		cancel()
	}
}

func (w *Worker) process(ctx context.Context, task domain.CalcTask) {
	stored, err := w.pgRepo.GetCalculation(ctx, task.CalculationID, task.Profile.UserID)
	if err != nil {
		log.Printf("worker: расчёт %d, чтение: %v", task.CalculationID, err)

		return
	}

	if stored.PaymentSchedule == nil {
		calc := domain.Calculate(task.Profile)
		calc.ID = task.CalculationID

		if err = w.pgRepo.UpdateCalculation(ctx, calc); err != nil {
			log.Printf("worker: расчёт %d: %v", task.CalculationID, err)

			return
		}

		stored, err = w.pgRepo.GetCalculation(ctx, task.CalculationID, task.Profile.UserID)
		if err != nil {
			log.Printf("worker: расчёт %d, чтение результата: %v", task.CalculationID, err)

			return
		}
	}

	w.storeResult(ctx, task.Profile, stored)
}

func (w *Worker) storeResult(ctx context.Context, profile domain.MortgageProfile, calc domain.MortgageCalculation) {
	calc.Status = domain.StatusDone

	paramsKey, err := domain.UserCacheKey(profile)
	if err != nil {
		log.Printf("worker: расчёт %d, ключ кэша: %v", calc.ID, err)

		return
	}

	for _, key := range []string{paramsKey, domain.IDCacheKey(calc.ID)} {
		if err := w.cacheRepo.SetCalculation(ctx, key, calc); err != nil {
			log.Printf("worker: расчёт %d, запись в кэш: %v", calc.ID, err)
		}
	}
}
