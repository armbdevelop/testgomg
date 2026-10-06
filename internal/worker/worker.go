package worker

import (
	"context"
	"log"

	"github.com/armbdevelop/testgomg/internal/core/domain"
	"github.com/armbdevelop/testgomg/internal/core/ports"
)

type Worker struct {
	pgRepo ports.PGRepository
	tasks  <-chan domain.CalcTask // только читаем
}

func New(pgRepo ports.PGRepository, tasks <-chan domain.CalcTask) *Worker {
	return &Worker{
		pgRepo: pgRepo,
		tasks:  tasks,
	}
}

func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-w.tasks:
			calc := domain.Calculate(task.Profile)
			calc.ID = task.CalculationID

			if err := w.pgRepo.UpdateCalculation(ctx, calc); err != nil {
				log.Printf("worker: расчёт %d: %v", task.CalculationID, err)
			}
		}
	}
}
