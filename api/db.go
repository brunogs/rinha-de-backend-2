package api

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"strconv"
)

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "pass"
	dbName     = "rinha"
)

func NewPoolConnection(ctx context.Context) (*pgxpool.Pool, error) {
	host := dbHost
	if value, found := os.LookupEnv("DB_HOSTNAME"); found {
		host = value
	}
	connString := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", dbUser, dbPassword, host, dbPort, dbName)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	maxConns, minConns := 6, 6
	if value, found := os.LookupEnv("POOL_MAX"); found {
		maxConns, _ = strconv.Atoi(value)
	}
	if value, found := os.LookupEnv("POOL_MIN"); found {
		minConns, _ = strconv.Atoi(value)
	}
	config.MaxConns = int32(maxConns)
	config.MinConns = int32(minConns)
	return pgxpool.NewWithConfig(ctx, config)
}
