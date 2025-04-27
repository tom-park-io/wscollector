package storage

import "sync"

type MemoryStore struct {
	mu     sync.Mutex
	trades []Trade
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		trades: make([]Trade, 0),
	}
}

func (m *MemoryStore) SaveTrade(t Trade) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trades = append(m.trades, t)
	return nil
}

func (m *MemoryStore) GetTrades() []Trade {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Copy to avoid race
	copyTrades := make([]Trade, len(m.trades))
	copy(copyTrades, m.trades)
	return copyTrades
}
