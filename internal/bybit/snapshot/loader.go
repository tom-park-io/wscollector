package snapshot

import (
	"context"

	"wscollector/config"
	"wscollector/pkg/bybit"

	"go.uber.org/zap"
)

type SymbolLoader struct {
	Cfg        config.Config
	RestClient *bybit.RESTClient
	Logger     *zap.Logger
}

// LoadSymbols fetches USDT-margined altcoin trading pairs from Bybit
// and streams them into the provided channel.
// The function applies a 10-second timeout to the REST request.
func (l *SymbolLoader) LoadSymbols(ch chan<- string) error {
	defer close(ch) // Ensure downstream consumers can exit cleanly

	ctx, cancel := context.WithTimeout(context.Background(), l.Cfg.Bybit.REST.Timeout)
	defer cancel()

	symbols, err := l.RestClient.GetUSDTAltcoinSymbols(ctx)
	if err != nil {
		l.Logger.Error("failed to load USDT altcoin symbols", zap.Error(err))
		return err
	}
	l.Logger.Info("loaded symbols", zap.Int("count", len(symbols)))

	for _, symbol := range symbols {
		select {
		case ch <- symbol:
		case <-ctx.Done():
			l.Logger.Warn("symbol streaming interrupted", zap.Error(ctx.Err()))
			return ctx.Err()
		}
	}

	return nil
}
