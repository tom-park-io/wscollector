package memorystore

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"sync"

	"go.uber.org/zap"
)

type MemorySymbolStore struct {
	mu          sync.Mutex
	symbols     map[string]struct{}
	WsInterval  string
	klineTopics []string
	lastHash    uint64
	logger      *zap.Logger
}

// Constructor: Initializes a new symbol store
func NewSymbolStore(interval string, logger *zap.Logger) *MemorySymbolStore {
	return &MemorySymbolStore{
		symbols:    make(map[string]struct{}),
		WsInterval: interval,
		logger:     logger,
	}
}

// Add inserts a new symbol into the store
func (s *MemorySymbolStore) Add(symbol string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.symbols[symbol] = struct{}{}
}

// Remove deletes a symbol from the store
func (s *MemorySymbolStore) Remove(symbol string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.symbols, symbol)
}

// Contains checks if a symbol exists in the store
func (s *MemorySymbolStore) Contains(symbol string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.symbols[symbol]
	return exists
}

// Count returns the number of unique symbols currently stored
func (s *MemorySymbolStore) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.symbols)
}

// GetAll returns a slice of all stored symbols
func (s *MemorySymbolStore) GetAll() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, len(s.symbols))
	for sym := range s.symbols {
		out = append(out, sym)
	}
	return out
}

// GetKlineTopics returns the cached list of kline stream topics.
// It only regenerates the list if the symbol set has changed.
func (s *MemorySymbolStore) GetKlineTopics(interval string) []string {
	return s.buildKlineTopics(interval, true)
}

// RefreshKlineTopics forces a regeneration of the kline topic list,
// regardless of whether the symbol set has changed.
func (s *MemorySymbolStore) RefreshKlineTopics(interval string) []string {
	return s.buildKlineTopics(interval, false)
}

// buildKlineTopics computes the kline topics based on the current symbol set.
// It acquires a mutex lock before accessing or updating internal state.
// If useCache is true, it returns the cached topic list when the symbol set hasn't changed.
func (s *MemorySymbolStore) buildKlineTopics(interval string, useCache bool) []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.buildKlineTopicsUnlocked(interval, useCache)
}

// buildKlineTopicsUnlocked is the non-locking version of buildKlineTopics.
// It assumes that the caller has already acquired the necessary mutex lock.
func (s *MemorySymbolStore) buildKlineTopicsUnlocked(interval string, useCache bool) []string {

	symbols := make([]string, 0, len(s.symbols))
	for sym := range s.symbols {
		symbols = append(symbols, sym)
	}
	sort.Strings(symbols)

	source := strings.Join(symbols, ",")
	hash := fnvHash(source)

	if useCache && s.klineTopics != nil && hash == s.lastHash {
		s.logger.Debug("returning cached kline topics")
		return s.klineTopics
	}

	topics := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		topics = append(topics, fmt.Sprintf("kline.%s.%s", interval, symbol))
	}

	s.klineTopics = topics
	s.lastHash = hash

	s.logger.Info("kline topics regenerated",
		zap.Int("total_symbols", len(topics)),
		zap.Bool("forced", !useCache),
	)

	return topics
}

func fnvHash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// StartSymbolSyncWorker listens to a symbol stream from the channel,
// compares it with the current in-memory set, and updates the store
// only if the contents have changed. It logs any differences and
// regenerates the kline topic list accordingly.
func (s *MemorySymbolStore) StartSymbolSyncWorker(ch <-chan string) {
	go func() {
		// Collect incoming symbols into a new temporary set
		newSymbols := make(map[string]struct{})
		for symbol := range ch {
			newSymbols[symbol] = struct{}{}
		}

		newHash := computeSymbolHash(newSymbols)

		s.mu.Lock()
		defer s.mu.Unlock()

		// Skip update if no symbol changes detected
		if newHash == s.lastHash {
			s.logger.Info("symbol set unchanged; skipping update and topic refresh")
			return
		}

		s.logger.Info("symbol set changed; applying update")

		added, removed := 0, 0

		// Log and count symbols that were removed
		for sym := range s.symbols {
			if _, stillExists := newSymbols[sym]; !stillExists {
				s.logger.Info("symbol removed", zap.String("symbol", sym))
				removed++
			}
		}

		// Log and count symbols that were newly added
		for sym := range newSymbols {
			if _, alreadyExists := s.symbols[sym]; !alreadyExists {
				s.logger.Info("symbol added", zap.String("symbol", sym))
				added++
			}
		}

		// Replace the current store with the updated set
		s.symbols = newSymbols
		s.lastHash = newHash
		s.logger.Info("symbol store synchronized",
			zap.Int("total", len(s.symbols)),
			zap.Int("added", added),
			zap.Int("removed", removed),
		)

		// Rebuild kline topics (mutex must be held)
		s.buildKlineTopicsUnlocked(s.WsInterval, false)
	}()
}

func computeSymbolHash(symbols map[string]struct{}) uint64 {
	syms := make([]string, 0, len(symbols))
	for sym := range symbols {
		syms = append(syms, sym)
	}
	sort.Strings(syms)

	src := strings.Join(syms, ",")
	return fnvHash(src)
}
