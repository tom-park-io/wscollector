package postgres_test

import (
	"context"
	"testing"
	"time"

	"wscollector/config"
	"wscollector/pkg/storage/postgres"
)

// go test -v --run ^TestPostgresInvalidDSN$
func TestPostgresInvalidDSN(t *testing.T) {
	invalidDSN := "host=invalid port=5432 user=fail password=fail dbname=fail sslmode=disable"

	_, err := postgres.NewClient(invalidDSN)
	if err == nil {
		t.Fatal("expected error for invalid DSN, got nil")
	}
}

// go test -v --run ^TestPostgresClientWithConfig$
func TestPostgresClientWithConfig(t *testing.T) {
	cfg := config.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "yourpw",
		DBName:   "wscollector",
		SSLMode:  "disable",

		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 1 * time.Hour,
	}

	dsn := cfg.DSN("dev")

	client, err := postgres.NewClient(dsn)
	if err != nil {
		t.Fatalf("failed to create Postgres client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if !client.IsHealthy(ctx) {
		t.Fatal("expected healthy DB connection")
	}

	if err := client.AutoMigrateKlineRecord(); err != nil {
		t.Fatalf("auto migration failed: %v", err)
	}
}
