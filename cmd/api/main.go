package main

import (
	"auth-service/internal/app"
	"context"
	"log"
)

func main() {
	ctx := context.Background()
	app, err := app.NewApp(ctx)
	if err != nil {
		panic(err)
	}
	if err := app.Run(); err != nil {
		log.Fatalf("main.main, failed to run app %v", err)
	}
}
