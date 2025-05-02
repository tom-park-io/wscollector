package collector

import (
	"context"
	"fmt"
	"time"

	"wscollector/config"
	"wscollector/internal/bybit/memorystore"
	"wscollector/internal/bybit/snapshot"
	"wscollector/internal/bybit/stream"
	"wscollector/internal/bybit/symbolmeta"
	"wscollector/pkg/bybit"
	"wscollector/pkg/storage/postgres"

	"go.uber.org/zap"
)

// StartCollector initializes the data pipeline for Bybit linear market data.
// It loads symbol metadata via REST, sets up a WebSocket stream for klines,
// and stores them in-memory (and optionally to DB).
func StartCollector(cfg config.Config, logger *zap.Logger) error {

	// Initialize PostgreSQL Client
	postgresClient, err := postgres.InitializeAndMigrateKlineRecord(cfg.App.Env, cfg.Postgres, true)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}

	// Create REST client and channel for symbol metadata
	restClient := bybit.NewRESTClient(cfg.Bybit.REST.BaseURL, cfg.Bybit.REST.Timeout)

	// Initialize the symbol loader with required dependencies
	loader := &snapshot.SymbolLoader{
		Cfg:        cfg,
		RestClient: restClient,
		Logger:     logger,
	}

	// Construct the midnight loader with a strategy that fetches symbols asynchronously
	midnight := &symbolmeta.MidnightLoader{
		Load: symbolmeta.DefaultLoadFn(loader),
	}

	// Parse the interval string into a KlineIntervalMeta type
	klineMeta, err := bybit.ParseKlineInterval(cfg.Bybit.WS.Interval)
	if err != nil {
		return fmt.Errorf("failed to parse interval: %w", err)
	}

	// Initialize in-memory symbol store and start worker to consume incoming symbols
	symbolStore := memorystore.NewSymbolStore(klineMeta.APIValue, logger)
	midnight.Start(symbolStore.StartSymbolSyncWorker)

	logger.Info("waiting 5 seconds before starting symbol sync", zap.String("reason", "initialization delay"))
	time.Sleep(5 * time.Second)

	// TODO: Concurrent tasks
	sem := make(chan struct{}, 10) // max 10 concurrent tasks
	// Prepare kline subscription topics
	end := time.Now()
	start := end.Add(-4 * time.Hour)

	symbols := symbolStore.GetAll()
	for _, symbol := range symbols {
		symbol := symbol // capture
		sem <- struct{}{}

		go func() {
			defer func() { <-sem }()

			var failed bool

			// Context with timeout for safety
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Bybit.REST.Timeout)
			// fetch
			restData, err := restClient.GetKlines(ctx, "linear", symbol,
				cfg.Bybit.WS.Interval, start, end)
			cancel()
			if err != nil {
				logger.Warn("failed to fetch kline from REST", zap.String("symbol", symbol), zap.Error(err))
				failed = true
				goto LOG_DONE
			}

			for _, kline := range restData {
				// Convert to DB record
				klineRecord, err := postgres.ToKlineRecord(symbol, kline)
				if err != nil {
					logger.Warn("failed to convert kline data to kline record", zap.String("symbol", symbol), zap.Error(err))
					failed = true
					continue
				}

				// Insert Kline record into Postgres
				// context for DB insert (short timeout)
				dbCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				err = postgresClient.InsertKline(dbCtx, klineRecord)
				cancel()
				if err != nil {
					logger.Warn("failed to insert kline into DB", zap.String("symbol", symbol), zap.Error(err))
					failed = true
					continue
				}
			}

		LOG_DONE:
			if failed {
				logger.Warn("finished with errors for symbol", zap.String("symbol", symbol))
			} else {
				logger.Info("completed successfully for symbol", zap.String("symbol", symbol))
			}
		}()
	}

	// Initialize WebSocket client
	wsClient := bybit.NewWSClient(cfg.Bybit.WS.URL, symbolStore, logger)
	klineStore := memorystore.NewKlineStore()

	// Register WebSocket message handler
	wsClient.SetMessageHandler(stream.MakeMessageHandler(logger, klineStore, postgresClient))

	// Periodically print stored Kline count for visibility
	go func() {
		for {
			count := klineStore.CountAll()
			logger.Info("current saved klines", zap.Int("count", count))

			time.Sleep(5 * time.Second)
		}
	}()

	// Connect to WebSocket with the list of symbols
	if err := wsClient.Connect(); err != nil {
		return err
	}
	go wsClient.Listen() // explicitly start listener

	return nil
}
