package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
	dataSource := os.Getenv("DATABASE_URL")
    if dataSource == "" {
        return nil, fmt.Errorf("DATABASE_URL is not set")
    }

    db, err := sql.Open("postgres", dataSource)
    if err != nil {
        return nil, fmt.Errorf("failed to open DB: %w", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to connect to DB: %w", err)
    }

    fmt.Println("Connected to PostgreSQL successfully")
	return db, nil
}

var ErrMissingDSN = sql.ErrConnDone
