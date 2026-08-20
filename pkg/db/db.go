package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

//go:embed migrations/001_gm_schema.sql
var schemaSQL string

type Config struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Open(ctx context.Context, cfg Config) (*sql.DB, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("database URL is required")
	}
	if cfg.MaxOpenConns <= 0 {
		cfg.MaxOpenConns = 10
	}
	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 5
	}
	if cfg.ConnMaxLifetime <= 0 {
		cfg.ConnMaxLifetime = 30 * time.Minute
	}
	database, err := sql.Open("sqlserver", cfg.URL)
	if err != nil {
		return nil, err
	}
	database.SetMaxOpenConns(cfg.MaxOpenConns)
	database.SetMaxIdleConns(cfg.MaxIdleConns)
	database.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	if err = database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	return database, nil
}

func Migrate(ctx context.Context, database *sql.DB) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, schemaSQL); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
