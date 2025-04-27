package memorystore

import "sync"

type MemorySymbolStore struct {
	mu      sync.Mutex
	symbols []string
}

func NewSymbolStore() *MemorySymbolStore {
	return &MemorySymbolStore{
		symbols: make([]string, 0),
	}
}

func (s *MemorySymbolStore) Add(symbol string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.symbols = append(s.symbols, symbol)
}

func (s *MemorySymbolStore) StartWorker(ch <-chan string) {
	go func() {
		for symbol := range ch {
			s.Add(symbol)
		}
	}()
}

func (s *MemorySymbolStore) GetAll() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.symbols))
	copy(out, s.symbols)
	return out
}
