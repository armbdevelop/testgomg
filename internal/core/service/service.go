package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/armbdevelop/testgomg/internal/core/domain"
	"github.com/armbdevelop/testgomg/internal/core/ports"
)

type service struct {
	pgRepo    ports.PGRepository
	cacheRepo ports.CacheRepository
	tasks     chan domain.CalcTask // очередь задач воркеру: id расчётов + снапшот профиля
}

func (s *service) CreateCalculation(ctx context.Context, profile domain.MortgageProfile) (id int64, err error) {
	if err = validateProfile(profile); err != nil {
		return 0, err
	}

	key, err := cacheKey(profile)
	if err != nil {
		return 0, fmt.Errorf("service.CreateCalculation.CacheKey: %w", err)
	}

	if id, err = s.cacheRepo.Get(ctx, key); err == nil {
		return id, nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return 0, fmt.Errorf("service.CreateCalculation.CacheGet: %w", err)
	}

	if id, err = s.pgRepo.CreateCalculation(ctx, profile,
		domain.MortgageCalculation{}); err != nil {
		return 0, err
	}

	// Задачу воркеру на расчёт графика
	s.tasks <- domain.CalcTask{CalculationID: id, Profile: profile}

	if err = s.cacheRepo.Set(ctx, key, id); err != nil {
		return 0, fmt.Errorf("service.CreateCalculation.CacheSet: %w", err)
	}

	return id, nil
}

func (s *service) GetCalculation(ctx context.Context, id int64) (calc domain.MortgageCalculation, err error) {
	calc, err = s.pgRepo.GetCalculation(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return calc, fmt.Errorf("%w: расчёт с id %d не найден", domain.ErrNotFound, id)
		}

		return calc, fmt.Errorf("service.GetCalculation.PGGet: %w", err)
	}

	calc.Status = "done"
	if calc.PaymentSchedule == nil {
		calc.Status = "pending"
	}

	return calc, nil
}

func NewService(pgRepo ports.PGRepository, cacheRepo ports.CacheRepository, tasks chan domain.CalcTask) ports.Service {
	return &service{
		pgRepo:    pgRepo,
		cacheRepo: cacheRepo,
		tasks:     tasks,
	}
}

func validateProfile(p domain.MortgageProfile) error {
	if !p.PropertyType.Valid() {
		return fmt.Errorf("%w: некорректный тип недвижимости",
			domain.ErrValidation)
	}

	if p.MatCapitalIncluded && p.MatCapitalAmount == nil {
		return fmt.Errorf("%w: укажите сумму маткапитала",
			domain.ErrValidation)
	}
	// маткапитал учитывается только при MatCapitalIncluded
	matCapital := 0.0
	if p.MatCapitalIncluded {
		matCapital = *p.MatCapitalAmount
	}

	if p.DownPaymentAmount+matCapital >= p.PropertyPrice {
		return fmt.Errorf("%w: взнос и маткапитал покрывают стоимость, кредит не нужен", domain.ErrValidation)
	}

	return nil
}

func cacheKey(p domain.MortgageProfile) (string, error) {
	data, err := json.Marshal(p) // только входные поля влияют на ключ
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(data)

	return "mortgage:" + hex.EncodeToString(sum[:]), nil
}
