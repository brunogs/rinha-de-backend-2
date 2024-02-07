package main

import (
	"api/api"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	pool, err := api.NewPoolConnection(context.Background())
	if err != nil {
		log.Fatal("Falha ao abrir pool de conexões", err)
	}

	queries := api.NewQueries(pool)
	app := fiber.New()

	handler := api.NewHandler(queries, app)
	handler.SetupEndpoints()

	// replace port from env var
	err = app.Listen(fmt.Sprintf(":%d", 3000))
	if err != nil {
		log.Fatal("Falhou ao iniciar Fiber")
	}
}
