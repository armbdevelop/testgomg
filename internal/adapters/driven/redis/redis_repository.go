package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/armbdevelop/testgomg/internal/core/domain"
	"github.com/armbdevelop/testgomg/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

type cacheRepo struct {
	client    *redis.Client
	resultTTL time.Duration
}

func NewCacheRepository(client *redis.Client, resultTTL time.Duration) ports.CacheRepository {
	return &cacheRepo{client: client, resultTTL: resultTTL}
}

func (c *cacheRepo) GetCalculation(ctx context.Context, key string) (calc domain.MortgageCalculation, err error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return calc, domain.ErrNotFound
		}

		return calc, fmt.Errorf("cacheRepo.GetCalculation: %w", err)
	}

	if err = json.Unmarshal(data, &calc); err != nil {
		return calc, fmt.Errorf("cacheRepo.GetCalculation.Unmarshal: %w", err)
	}

	return calc, nil
}

func (c *cacheRepo) SetCalculation(ctx context.Context, key string, calc domain.MortgageCalculation) (err error) {
	data, err := json.Marshal(calc)
	if err != nil {
		return fmt.Errorf("cacheRepo.SetCalculation.Marshal: %w", err)
	}

	if err = c.client.Set(ctx, key, data, c.resultTTL).Err(); err != nil {
		return fmt.Errorf("cacheRepo.SetCalculation: %w", err)
	}

	return nil
}
