package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/armbdevelop/testgomg/internal/core/domain"
	"github.com/armbdevelop/testgomg/internal/core/ports"
)

const (
	opTimeout = 10 * time.Second
	// maxMortgageTermYears — разумный потолок срока ипотеки.
	maxMortgageTermYears = 50
)

type service struct {
	pgRepo    ports.PGRepository
	cacheRepo ports.CacheRepository
	queue     ports.TaskQueue
}

func NewService(pgRepo ports.PGRepository, cacheRepo ports.CacheRepository, queue ports.TaskQueue) ports.Service {
	return &service{pgRepo: pgRepo, cacheRepo: cacheRepo, queue: queue}
}

func (s *service) CreateCalculation(ctx context.Context, profile domain.MortgageProfile) (id int64, err error) {
	if err = validateProfile(profile); err != nil {
		return 0, err
	}

	key, err := domain.UserCacheKey(profile)
	if err != nil {
		return 0, fmt.Errorf("service.CreateCalculation.CacheKey: %w", err)
	}

	if calc, getErr := s.cacheRepo.GetCalculation(ctx, key); getErr == nil &&
		calc.ID > 0 && calc.UserID == profile.UserID && calc.PaymentSchedule != nil {
		return calc.ID, nil
	} else if getErr != nil && !errors.Is(getErr, domain.ErrNotFound) {
		log.Printf("service.CreateCalculation.CacheGet: %v", getErr)
	}

	calc, err := s.pgRepo.GetOrCreateCalculation(ctx, profile)
	if err != nil {
		return 0, err
	}

	opCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), opTimeout)
	defer cancel()

	if calc.PaymentSchedule != nil {
		calc.Status = domain.StatusDone
		for _, cacheKey := range []string{key, domain.IDCacheKey(calc.ID)} {
			if cacheErr := s.cacheRepo.SetCalculation(opCtx, cacheKey, calc); cacheErr != nil {
				log.Printf("service.CreateCalculation.CacheSet: %v", cacheErr)
			}
		}
	} else if err = s.queue.Enqueue(opCtx, domain.CalcTask{CalculationID: calc.ID, Profile: profile}); err != nil {
		return 0, fmt.Errorf("service.CreateCalculation.Enqueue: %w", err)
	}

	return calc.ID, nil
}

func (s *service) GetCalculation(ctx context.Context, id int64,
	userID string) (calc domain.MortgageCalculation, err error) {
	if cached, cacheErr := s.cacheRepo.GetCalculation(ctx, domain.IDCacheKey(id)); cacheErr == nil {
		if cached.UserID != userID {
			return calc, fmt.Errorf("%w: расчёт с id %d не найден", domain.ErrNotFound, id)
		}

		return cached, nil
	}

	calc, err = s.pgRepo.GetCalculation(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return calc, fmt.Errorf("%w: расчёт с id %d не найден", domain.ErrNotFound, id)
		}

		return calc, fmt.Errorf("service.GetCalculation.PGGet: %w", err)
	}

	calc.Status = domain.StatusDone
	if calc.PaymentSchedule == nil {
		calc.Status = domain.StatusPending
	}

	return calc, nil
}

func validateProfile(p domain.MortgageProfile) error {
	if p.PropertyPrice <= 0 || p.DownPaymentAmount < 0 ||
		p.InterestRate <= 0 || p.MortgageTermYears <= 0 {
		return fmt.Errorf("%w: цена, ставка и срок должны быть положительными, взнос — неотрицательным",
			domain.ErrValidation)
	}

	if p.MatCapitalAmount != nil && *p.MatCapitalAmount < 0 {
		return fmt.Errorf("%w: маткапитал не может быть отрицательным", domain.ErrValidation)
	}

	// Без потолка срока один запрос со сроком в миллион лет уронит воркер (OOM).
	if p.MortgageTermYears > maxMortgageTermYears {
		return fmt.Errorf("%w: срок ипотеки не может превышать %d лет", domain.ErrValidation, maxMortgageTermYears)
	}

	if !p.PropertyType.Valid() {
		return fmt.Errorf("%w: некорректный тип недвижимости", domain.ErrValidation)
	}

	if p.MatCapitalIncluded && p.MatCapitalAmount == nil {
		return fmt.Errorf("%w: укажите сумму маткапитала", domain.ErrValidation)
	}

	matCapital := 0.0
	if p.MatCapitalIncluded {
		matCapital = *p.MatCapitalAmount
	}

	if p.DownPaymentAmount+matCapital >= p.PropertyPrice {
		return fmt.Errorf("%w: взнос и маткапитал покрывают стоимость, кредит не нужен", domain.ErrValidation)
	}

	return nil
}
