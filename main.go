package main

import (
	"api/api"
	"context"
	"github.com/gin-gonic/gin"
	"os"
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

	listeners := api.NewListeners(queries)
	listeners.SetupListeners()

	port := os.Getenv("PORT")
	if err = r.Run(":" + port); err != nil {
		panic("Failed to run gin server")
	}
}
