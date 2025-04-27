package stream

import (
	"context"
	"encoding/json"
	"strings"

	"wscollector/internal/bybit/memorystore"
	"wscollector/pkg/storage/postgres"

	"go.uber.org/zap"
)

// MakeMessageHandler returns a function that handles incoming WebSocket messages
// by parsing kline data and storing it in memory.
func MakeMessageHandler(logger *zap.Logger, store *memorystore.MemoryKlineStore,
	postgresClient *postgres.PostgresClient) func(msg []byte) {
	return func(msg []byte) {
		// Step 1: Extract topic string for early filtering
		var meta struct {
			Topic string `json:"topic"`
		}
		if err := json.Unmarshal(msg, &meta); err != nil {
			logger.Warn("failed to extract topic", zap.Error(err))
			return
		}
		if !isKlineTopic(meta.Topic) {
			return // Ignore non-kline messages (e.g., subscription responses)
		}

		// Step 2: Fully parse the kline message payload
		var parsed KlineMessage
		if err := json.Unmarshal(msg, &parsed); err != nil {
			logger.Warn("failed to parse kline payload", zap.Error(err))
			return
		}
		symbol := extractSymbolFromTopic(parsed.Topic) // e.g., "kline.1.BTCUSDT" → "BTCUSDT"

		// Step 3: Store parsed kline data
		for _, d := range parsed.Data {
			// Optional: store only confirmed klines
			if !d.Confirm {
				continue
			}

			kline := memorystore.Kline{
				Start:     d.Start,
				End:       d.End,
				Interval:  d.Interval,
				Open:      d.Open,
				Close:     d.Close,
				High:      d.High,
				Low:       d.Low,
				Volume:    d.Volume,
				Turnover:  d.Turnover,
				Confirm:   d.Confirm,
				Timestamp: d.Timestamp,
			}
			// Insert Kline data into Memory
			store.Add(memorystore.KlineMemory{
				Symbol: symbol,
				Kline:  kline,
			})

			ctx := context.Background()
			klineRecord, err := postgres.ToKlineRecord(symbol, kline)
			if err != nil {
				logger.Warn("failed to convert kline data to kline record", zap.Error(err))
			}
			// Insert Kline record into Postgres
			if err := postgresClient.InsertKline(ctx, klineRecord); err != nil {
				logger.Warn("failed to insert kline record", zap.Error(err))
			}
		}
	}
}

// isKlineTopic returns true if the topic string indicates a kline stream.
func isKlineTopic(topic string) bool {
	return strings.HasPrefix(topic, "kline.")
}

// extractSymbolFromTopic parses the symbol from a topic like "kline.1.BTCUSDT".
func extractSymbolFromTopic(topic string) string {
	parts := strings.Split(topic, ".")
	if len(parts) == 3 {
		return parts[2]
	}
	return ""
}
