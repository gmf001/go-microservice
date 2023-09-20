package app

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Router http.Handler
	rdb *redis.Client
}

func New() *App {
	godotenv.Load()
	redisURL := os.Getenv("REDIS_URL")
	opt, _ := redis.ParseURL(redisURL)

	app := &App{
		Router: loadRoutes(),
		rdb: redis.NewClient(opt),
	}

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr: ":3000",
		Handler: a.Router,
	}

	err := a.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	fmt.Println("starting server on port 3000")

	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}