package memorystore

import (
	"sync"
)

type MemoryKlineStore struct {
	globalMu sync.RWMutex
	data     map[string]*symbolKlineStore
}

type symbolKlineStore struct {
	mu     sync.Mutex
	klines []Kline
}

func NewKlineStore() *MemoryKlineStore {
	return &MemoryKlineStore{
		data: make(map[string]*symbolKlineStore),
	}
}

func (s *MemoryKlineStore) Add(k KlineMemory) {
	// Fast path: lock per-symbol store only
	s.globalMu.RLock()
	store, ok := s.data[k.Symbol]
	s.globalMu.RUnlock()

	if !ok {
		// Need to initialize new symbol store (exclusive lock)
		s.globalMu.Lock()
		if store, ok = s.data[k.Symbol]; !ok {
			store = &symbolKlineStore{}
			s.data[k.Symbol] = store
		}
		s.globalMu.Unlock()
	}

	// Per-symbol locking
	store.mu.Lock()
	store.klines = append(store.klines, k.Kline)
	store.mu.Unlock()
}

func (s *MemoryKlineStore) GetBySymbol(symbol string) []Kline {
	s.globalMu.RLock()
	store, ok := s.data[symbol]
	s.globalMu.RUnlock()
	if !ok {
		return nil
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	cp := make([]Kline, len(store.klines))
	copy(cp, store.klines)
	return cp
}

func (s *MemoryKlineStore) GetAll() map[string][]Kline {
	s.globalMu.RLock()
	defer s.globalMu.RUnlock()

	result := make(map[string][]Kline, len(s.data))
	for sym, store := range s.data {
		store.mu.Lock()
		cp := make([]Kline, len(store.klines))
		copy(cp, store.klines)
		store.mu.Unlock()
		result[sym] = cp
	}
	return result
}

// CountAll returns the total number of Klines stored across all symbols.
func (s *MemoryKlineStore) CountAll() int {
	s.globalMu.RLock()
	defer s.globalMu.RUnlock()

	total := 0
	for _, store := range s.data {
		store.mu.Lock()
		total += len(store.klines)
		store.mu.Unlock()
	}
	return total
}
