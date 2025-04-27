package bybit

import (
	"context"
	"testing"
	"time"
)

// go test -v --run TestGetUSDTAltcoinSymbols
func TestGetUSDTAltcoinSymbols(t *testing.T) {
	// Create the REST client with real base URL
	client := NewRESTClient("https://api.bybit.com", 10*time.Second)

	// Context with timeout for safety
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	symbols, err := client.GetUSDTAltcoinSymbols(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(symbols) == 0 {
		t.Fatal("expected non-empty symbol list, got 0")
	}

	// Optional: print first few symbols for visual confirmation
	t.Logf("got %d USDT altcoin symbols (example: %v)", len(symbols), symbols[:min(len(symbols), 5)])
}

// go test -v --run TestGetKlines
func TestGetKlines(t *testing.T) {
	client := NewRESTClient("https://api.bybit.com", 10*time.Second)

	// Context with timeout for safety
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	end := time.Now()
	start := end.Add(-4 * time.Hour)

	// Call the GetKlines method with known-good parameters.
	resp, err := client.GetKlines(ctx,
		"linear",  // market category (e.g., "linear", "spot", "inverse")
		"BTCUSDT", // symbol
		"1",       // interval in minutes as string
		start,
		end,
	)
	if err != nil {
		t.Fatalf("GetKlines returned error: %v", err)
	}

	if len(resp) == 0 {
		t.Error("Expected non-empty response body")
	}
	t.Logf("Received response: %v", resp)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
