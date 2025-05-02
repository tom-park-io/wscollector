package symbolmeta

import (
	"time"
	"wscollector/internal/bybit/snapshot"

	"go.uber.org/zap"
)

type MidnightLoader struct {
	Load func() <-chan string
}

func DefaultLoadFn(loader *snapshot.SymbolLoader) func() <-chan string {
	return func() <-chan string {
		symbolCh := make(chan string, 100)

		go func() {
			if err := loader.LoadSymbols(symbolCh); err != nil {
				loader.Logger.Fatal("failed to load symbols", zap.Error(err))
			}
		}()

		return symbolCh
	}
}

// StartMidnightLoader schedules the given load function to run once at UTC midnight and then every 24 hours.
func (m *MidnightLoader) Start(proc func(<-chan string)) {
	go func() {
		// Run immediately once at startup
		m.runOnce(proc)

		// Wait until next UTC midnight
		now := time.Now().UTC()
		nextMidnight := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
		time.Sleep(time.Until(nextMidnight))

		// Then run once every 24 hours
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			m.runOnce(proc)
			<-ticker.C
		}
	}()
}

func (m *MidnightLoader) runOnce(proc func(<-chan string)) {
	symbolCh := m.Load()
	proc(symbolCh)
}
