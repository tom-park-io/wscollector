package memorystore

// KlineMemory represents a kline with an attached trading symbol.
// Typically constructed by combining topic metadata (e.g., "kline.1.BTCUSDT") with the kline payload.
type KlineMemory struct {
	Symbol string `json:"symbol"` // Trading symbol (e.g., "BTCUSDT")
	Kline
}

// Kline represents a single candlestick (1m, 5m, etc.) received from the Bybit WebSocket stream.
type Kline struct {
	Start     int64  `json:"start"`     // Start time of the kline (in milliseconds since epoch)
	End       int64  `json:"end"`       // End time of the kline (in milliseconds since epoch)
	Interval  string `json:"interval"`  // Interval of the kline (e.g., "1", "5", "15") — in minutes
	Open      string `json:"open"`      // Opening price
	Close     string `json:"close"`     // Closing price
	High      string `json:"high"`      // Highest price during the interval
	Low       string `json:"low"`       // Lowest price during the interval
	Volume    string `json:"volume"`    // Trade volume (number of units traded)
	Turnover  string `json:"turnover"`  // Total traded value (usually in USD)
	Confirm   bool   `json:"confirm"`   // Whether the kline is finalized (true when the interval closes)
	Timestamp int64  `json:"timestamp"` // Time when the event was generated (in milliseconds since epoch)
}
