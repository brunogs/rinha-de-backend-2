package main

import (
	"api/api"
	"context"
	"github.com/gin-gonic/gin"
)

func main() {
	pool, err := api.NewPoolConnection(context.Background())
	if err != nil {
		panic("Falha ao abrir pool de conexões" + err.Error())
	}
	queries := api.NewQueries(pool)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	handler := api.NewGinHandler(queries)
	handler.SetupEndpoints(r)
	if err = r.Run(":3000"); err != nil {
		panic("Failed to run gin server")
	}
}
