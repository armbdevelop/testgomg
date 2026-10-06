package initx

import (
	"context"
	"fmt"
	"time"

	"github.com/armbdevelop/testgomg/config"
	"github.com/armbdevelop/testgomg/internal/adapters/driven/postgres"
	"github.com/armbdevelop/testgomg/internal/adapters/driven/redis"
	drivinghttp "github.com/armbdevelop/testgomg/internal/adapters/driving/http"
	"github.com/armbdevelop/testgomg/internal/core/service"
	"github.com/armbdevelop/testgomg/internal/worker"
	"github.com/gofiber/fiber/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	goredis "github.com/redis/go-redis/v9"
)

const tasksBufferSize = 100

type App struct {
	http *fiber.App
	host string
}

func NewApp(cfg *config.Config) (*App, error) {
	db, err := sqlx.Open("pgx", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.Username,
		cfg.Postgres.Password, cfg.Postgres.Database, cfg.Postgres.SSLMode))
	if err != nil {
		return nil, fmt.Errorf("initx.NewApp.Postgres: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("initx.NewApp.PingPostgres: %w", err)
	}

	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err = redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("initx.NewApp.PingRedis: %w", err)
	}

	pgRepo := postgres.NewPGRepository(db)
	cacheRepo := redis.NewCacheRepository(redisClient,
		time.Duration(cfg.Redis.CacheTTL)*time.Second)

	queue := worker.NewChannelQueue(tasksBufferSize)

	calcWorker := worker.New(pgRepo, cacheRepo, queue)
	go calcWorker.Run(context.Background())

	svc := service.NewService(pgRepo, cacheRepo, queue)
	handlers := drivinghttp.NewHandlers(svc)

	app := fiber.New(fiber.Config{ErrorHandler: drivinghttp.ErrorHandler})
	drivinghttp.MapRoutes(app, handlers, drivinghttp.NewAuthMiddleware(cfg.Auth.JWTSecret))

	return &App{http: app, host: cfg.Server.Host}, nil
}

func (a *App) Run() error {
	if err := a.http.Listen(a.host); err != nil {
		return fmt.Errorf("initx.Run: %w", err)
	}

	return nil
}
