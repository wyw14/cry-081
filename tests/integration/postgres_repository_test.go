package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/wyw14/cry-081/internal/repository/postgres"
)

func TestPostgresConnection(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	database, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := database.Ping(ctx); err != nil {
		t.Fatal(err)
	}
}
