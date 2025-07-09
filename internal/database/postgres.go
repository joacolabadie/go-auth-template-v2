package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joacolabadie/go-auth-template-v2/internal/config"
)

var db *pgxpool.Pool

func ConnectDatabase(cfg config.DatabaseConfig) error {
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString)
	if err != nil {
		return err
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime

	db, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return err
	}

	return nil
}

func GetPool() *pgxpool.Pool {
	return db
}

func CloseDatabase() {
	if db != nil {
		db.Close()
	}
}
