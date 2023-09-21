package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/gmf001/go-microservice/app"
)

func main() {
	app := app.New()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt) // graceful shutdown
	defer cancel()

	err := app.Start(ctx)
	if err != nil {
		fmt.Printf("failed to start app: %s\n", err)
	}

}