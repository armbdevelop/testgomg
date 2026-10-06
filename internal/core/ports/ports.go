package ports

import (
	"context"

	"github.com/armbdevelop/testgomg/internal/core/domain"
)

type PGRepository interface {
	// GetOrCreateCalculation атомарно находит или создаёт расчёт с теми же параметрами и владельцем.
	GetOrCreateCalculation(
		ctx context.Context,
		profile domain.MortgageProfile,
	) (calc domain.MortgageCalculation, err error)
	GetCalculation(ctx context.Context, id int64, userID string) (calc domain.MortgageCalculation, err error)
	UpdateCalculation(ctx context.Context, calc domain.MortgageCalculation) (err error)
}

type CacheRepository interface {
	GetCalculation(ctx context.Context, key string) (calc domain.MortgageCalculation, err error)
	SetCalculation(ctx context.Context, key string, calc domain.MortgageCalculation) (err error)
}

type TaskQueue interface {
	Enqueue(ctx context.Context, task domain.CalcTask) (err error)
	Dequeue(ctx context.Context) (task domain.CalcTask, err error)
}

type Service interface {
	CreateCalculation(ctx context.Context, profile domain.MortgageProfile) (id int64, err error)
	GetCalculation(ctx context.Context, id int64, userID string) (calc domain.MortgageCalculation, err error)
}
