package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rdb *redis.Client
}

func New() *App {
	godotenv.Load()
	redisURL := os.Getenv("REDIS_URL")
	opt, _ := redis.ParseURL(redisURL)

	app := &App{
		rdb: redis.NewClient(opt),
	}

	app.loadRoutes()

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr: ":3000",
		Handler: a.router,
	}

	err := a.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	defer func() {
		if err := a.rdb.Close(); err != nil {
			fmt.Printf("failed to close redis connection: %s\n", err)
		}
	}()

	fmt.Println("starting server on port 3000")

	ch := make(chan error, 1)

	go func() {
		err = server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
		case err := <-ch:
			return err
		case <-ctx.Done():
			fmt.Println("shutting down server")
			timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			return server.Shutdown(timeout)
	}
}