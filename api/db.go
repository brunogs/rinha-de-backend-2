package api

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "admin"
	dbPassword = "admin"
	dbName     = "rinha"
)

func NewPoolConnection(ctx context.Context) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	// set variables with min/max connections
	return pgxpool.New(ctx, connString)
}
