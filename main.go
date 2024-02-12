package main

import (
	"api/api"
	"context"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func main() {
	pool, err := api.NewPoolConnection(context.Background())
	if err != nil {
		panic("Falha ao abrir pool de conexões" + err.Error())
	}
	queries := api.NewQueries(pool)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	logger := zap.NewNop()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	handler := api.NewGinHandler(queries)
	handler.SetupEndpoints(r)

	if err = r.Run(":3000"); err != nil {
		panic("Failed to run gin server")
	}
}
