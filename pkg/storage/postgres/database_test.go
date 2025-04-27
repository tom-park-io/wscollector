package postgres_test

import (
	"testing"

	"wscollector/config"
	"wscollector/pkg/storage/postgres"
)

// go test -v --run TestCreateDatabase
func TestCreateDatabase(t *testing.T) {
	cfg := config.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "yourpw",
		DBName:   "test_kline_db",
		SSLMode:  "disable",
	}

	err := postgres.CreateDatabase(cfg)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
}
