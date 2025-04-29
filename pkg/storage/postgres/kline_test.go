package postgres_test

import (
	"context"
	"testing"
	"time"

	"wscollector/config"
	"wscollector/pkg/storage/postgres"
)

// go test -v --run TestKlineCRUD
func TestKlineCRUD(t *testing.T) {
	cfg := config.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "yourpw",
		DBName:   "wscollector",
		SSLMode:  "disable",
		TimeZone: "UTC",
	}

	client, err := postgres.NewClient(cfg.DSN("dev"))
	if err != nil {
		t.Fatalf("failed to connect to DB: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	if err := client.AutoMigrateKlineRecord(); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// Create
	now := time.Now().Truncate(time.Minute)
	record := &postgres.KlineRecord{
		Symbol:    "BTCUSDT",
		Interval:  "1h",
		Start:     now,
		End:       now.Add(time.Minute),
		Open:      31400.0,
		Close:     31500.0,
		High:      31600.0,
		Low:       31300.0,
		Volume:    123.45,
		Turnover:  3890000.0,
		Confirm:   true,
		Timestamp: time.Now(),
	}

	if err := client.InsertKline(ctx, record); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	// Read
	got, err := client.GetKline(ctx, "BTCUSDT", "1h", now)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Symbol != "BTCUSDT" || got.Open != 31400.0 {
		t.Errorf("unexpected kline values: %+v", got)
	}

	// Update
	if err := client.UpdateKlineConfirm(ctx, got.ID, true); err != nil {
		t.Errorf("update confirm failed: %v", err)
	}

	// Re-fetch and check
	updated, err := client.GetKline(ctx, "BTCUSDT", "1h", now)
	if err != nil {
		t.Fatalf("get after update failed: %v", err)
	}
	if !updated.Confirm {
		t.Errorf("confirm was not updated")
	}

	// Delete
	if err := client.DeleteOldKlines(ctx, time.Now().Add(1*time.Hour)); err != nil {
		t.Errorf("delete failed: %v", err)
	}

	// Check deletion
	_, err = client.GetKline(ctx, "BTCUSDT", "1h", now)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
