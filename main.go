package main

import (
	"api/api"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func main() {
	pool, err := api.NewPoolConnection(context.Background())
	if err != nil {
		panic("Falha ao abrir pool de conexões" + err.Error())
	}

	queries := api.NewQueries(pool)
	//Fiber foi o primeiro teste
	/*
		app := fiber.New()
		handler := api.NewFiberHandler(queries, app)
		handler.SetupEndpoints()
		err = app.Listen(fmt.Sprintf(":%d", 3000))
		if err != nil {
			panic("Falhou ao iniciar Fiber")
		}
	*/

	// Std http package
	/*handler := api.NewStdHandler(queries)
	handler.SetupEndpoints()
	err = http.ListenAndServe(fmt.Sprintf(":%d", 3000), nil)
	if err != nil {
		panic("Falhou ao iniciar std http")
	}*/

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	handler := api.NewGinHandler(queries)
	handler.SetupEndpoints(r)
	if err := r.Run(":3000"); err != nil {
		panic("Failed to run gin server")
	}
}

func showPGXPoolData(pool *pgxpool.Pool) {
	for {
		time.Sleep(1 * time.Second)
		stat := pool.Stat()
		fmt.Println("max_conn=", stat.MaxConns(), " idle_conns=", stat.IdleConns(), " acquired=", stat.AcquiredConns(), " total=", stat.TotalConns())
	}
}
