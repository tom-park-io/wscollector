package snapshot

import (
	"context"

	"wscollector/config"
	"wscollector/pkg/bybit"

	"go.uber.org/zap"
)

// LoadSymbols fetches USDT-margined altcoin trading pairs from Bybit
// and streams them into the provided channel.
// The function applies a 5-second timeout to the REST request.
func LoadSymbols(ch chan<- string, cfg *config.Config, client *bybit.RESTClient, log *zap.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Bybit.REST.Timeout)
	defer cancel()

	symbols, err := client.GetUSDTAltcoinSymbols(ctx)
	if err != nil {
		log.Error("failed to load USDT altcoin symbols", zap.Error(err))
		return err
	}
	log.Info("loaded symbols", zap.Int("count", len(symbols)))

	for _, symbol := range symbols {
		ch <- symbol
	}
	return nil
}
