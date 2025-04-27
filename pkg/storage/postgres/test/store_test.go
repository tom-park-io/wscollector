package storage

import (
	"testing"
)

// go test -v --run TestSaveAndRetrieveTrade
func TestSaveAndRetrieveTrade(t *testing.T) {
	store := NewMemoryStore()

	store.SaveTrade(Trade{
		Symbol: "BTCUSDT",
		Price:  45000.0,
		Volume: 0.123,
	})

	trades := store.GetTrades()
	t.Log("Stored trades: ", trades)

	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].Symbol != "BTCUSDT" {
		t.Errorf("unexpected symbol: %s", trades[0].Symbol)
	}
}
