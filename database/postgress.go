package database

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return err
	}

	DB = db
	return nil
}