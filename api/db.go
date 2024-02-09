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
	dbUser     = "postgres"
	dbPassword = "pass"
	dbName     = "rinha"
)

func NewPoolConnection(ctx context.Context) (*pgxpool.Pool, error) {
	host, found := os.LookupEnv("DB_HOSTNAME")
	if !found {
		host = dbHost
	}
	connString := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", dbUser, dbPassword, host, dbPort, dbName)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	config.MaxConns = 20
	config.MinConns = 20
	return pgxpool.NewWithConfig(ctx, config)
}
