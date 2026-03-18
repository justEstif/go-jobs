package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justestif/go-jobs/internal/adapters/postgres/queries"
)

var (
	DB   *queries.Queries
	Pool *pgxpool.Pool
)

func InitDB() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable not set")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	Pool = pool
	DB = queries.New(pool)
	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
