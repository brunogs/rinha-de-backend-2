package api

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "admin"
	dbPassword = "admin"
	dbName     = "rinha"
	dbMinConns = 60
)

func NewPoolConnection(ctx context.Context) (*pgxpool.Pool, error) {
	host, found := os.LookupEnv("DB_HOSTNAME")
	if !found {
		host = dbHost
	}
	connString := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?pool_min_conns=%d", dbUser, dbPassword, host, dbPort, dbName, dbMinConns)
	return pgxpool.New(ctx, connString)
}
