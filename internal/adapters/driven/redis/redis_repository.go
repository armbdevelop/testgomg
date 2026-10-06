package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/armbdevelop/testgomg/internal/core/domain"
	"github.com/armbdevelop/testgomg/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

type cacheRepo struct {
	client *redis.Client
	ttl    time.Duration
}

func (c *cacheRepo) Get(ctx context.Context, key string) (id int64, err error) {
	id, err = c.client.Get(ctx, key).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, domain.ErrNotFound
		}

		return 0, fmt.Errorf("cacheRepo.Get: %w", err)
	}

	return id, nil
}

func (c *cacheRepo) Set(ctx context.Context, key string, id int64) (err error) {
	if err = c.client.Set(ctx, key, id, c.ttl).Err(); err != nil {
		return fmt.Errorf("cacheRepo.Set: %w", err)
	}

	return nil
}

func NewCacheRepository(client *redis.Client, ttl time.Duration) ports.CacheRepository {
	return &cacheRepo{
		client: client,
		ttl:    ttl,
	}
}
