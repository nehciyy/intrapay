package db_test

import (
	"os"
	"testing"

	"github.com/nehciyy/intrapay/internal/db"
)

func TestInitDB(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("Skipping DB test: DATABASE_URL env var not set")
	}

	dbConn, err := db.InitDB()
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		t.Fatalf("DB ping failed: %v", err)
	}
}