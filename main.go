package main

import (
	"context"
	"fmt"

	"github.com/gmf001/go-microservice/app"
)

func main() {
	app := app.New()
	err := app.Start(context.TODO())
	if err != nil {
		fmt.Printf("failed to start app: %s\n", err)
	}
}