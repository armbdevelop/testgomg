package ports

import (
	"context"

	"github.com/armbdevelop/testgomg/internal/core/domain"
)

type PGRepository interface {
	// CreateCalculation сохраняет профиль и расчёт в одной транзакции, возвращает id расчёта.
	CreateCalculation(ctx context.Context, profile domain.MortgageProfile,
		calc domain.MortgageCalculation) (id int64, err error)
	GetCalculation(ctx context.Context, id int64) (calc domain.MortgageCalculation, err error)
	// UpdateCalculation дописывает результат воркера (суммы и график) в существующий расчёт.
	UpdateCalculation(ctx context.Context, calc domain.MortgageCalculation) (err error)
}

type CacheRepository interface {
	Get(ctx context.Context, key string) (id int64, err error)
	Set(ctx context.Context, key string, id int64) (err error)
}

type Service interface {
	CreateCalculation(ctx context.Context, profile domain.MortgageProfile) (id int64, err error)
	GetCalculation(ctx context.Context, id int64) (calc domain.MortgageCalculation, err error)
}
