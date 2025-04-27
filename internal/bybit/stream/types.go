package stream

import "wscollector/internal/bybit/memorystore"

// KlineMessage represents a WebSocket message from Bybit containing kline (candlestick) data.
type KlineMessage struct {
	Topic string              `json:"topic"` // Topic string indicating the subscription stream, e.g., "kline.1.BTCUSDT"
	Data  []memorystore.Kline `json:"data"`  // Array of kline (candlestick) data entries
	Ts    int64               `json:"ts"`    // Timestamp (in milliseconds) when the message was received
	Type  string              `json:"type"`  // Message type, e.g., "snapshot" or "delta"
}
