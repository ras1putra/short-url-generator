package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"urlshortener/internal/repository"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// SetupTestDB connects
func SetupTestDB(t *testing.T) (*sql.DB, *repository.Queries) {
	t.Helper()

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	pass := getEnv("DB_PASSWORD", "secret")
	name := getEnv("DB_NAME", "urlshortener_test")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "Failed to connect to test database")

	err = db.Ping()
	require.NoError(t, err, "Database ping failed")

	_, err = db.Exec("TRUNCATE TABLE faucet_claims, transactions, wallets, clicks, urls, ads, users CASCADE")
	require.NoError(t, err, "Failed to truncate tables for test cleanup")

	queries := repository.New(db)
	return db, queries
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
