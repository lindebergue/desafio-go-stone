package database

import (
	"os"
	"testing"
)

const truncateQuery = `
	TRUNCATE TABLE accounts, transfers RESTART IDENTITY;
`

func TestPostgresDB(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL is not set; skipping postgres integration tests")
	}

	db, err := NewPostgresDB(url)
	if err != nil {
		t.Fatalf("error opening test database: %v", err)
	}

	if _, err := db.(*postgresDB).db.Exec(truncateQuery); err != nil {
		t.Fatalf("error preparing database for tests")
	}

	runDBTests(t, db)
}
