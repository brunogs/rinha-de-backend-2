package main

import (
	"api/api"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func main() {
	pool, err := api.NewPoolConnection(context.Background())
	if err != nil {
		panic("Falha ao abrir pool de conexões" + err.Error())
	}

	queries := api.NewQueries(pool)
	app := fiber.New()
	//app.Use(pprof.New())
	handler := api.NewHandler(queries, app)
	handler.SetupEndpoints()

	go showPGXPoolData(pool)

	err = app.Listen(fmt.Sprintf(":%d", 3000))
	if err != nil {
		panic("Falhou ao iniciar Fiber")
	}
}

func showPGXPoolData(pool *pgxpool.Pool) {
	for {
		time.Sleep(1 * time.Second)
		stat := pool.Stat()
		fmt.Println("max_conn=", stat.MaxConns(), " idle_conns=", stat.IdleConns(), " acquired=", stat.AcquiredConns(), " total=", stat.TotalConns())
	}
}
