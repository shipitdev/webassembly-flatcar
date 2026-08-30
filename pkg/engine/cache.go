package engine

import (
	"sync"
	"sync/atomic"
)

// StateStore manages in-memory transactions with thread-safe RWMutex synchronization.
// Note: For WASI single-threaded runtimes, lock contention is zero, but keeping
// RWMutex ensures parity when compiled natively for host Linux.
type StateStore struct {
	mu           sync.RWMutex
	orders       map[string]float64
	totalRevenue uint64 // atomic cent accumulator
}

func NewStateStore() *StateStore {
	return &StateStore{
		orders: make(map[string]float64, 1024), // pre-allocate initial bucket capacity
	}
}

// Record persists the order total and increments global revenue counters.
func (s *StateStore) Record(orderID string, total float64) {
	s.mu.Lock()
	s.orders[orderID] = total
	s.mu.Unlock()

	atomic.AddUint64(&s.totalRevenue, uint64(total*100))
}

// Count returns the active in-memory cache size.
func (s *StateStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.orders)
}

// RevenueUSD returns the cumulative revenue in USD.
func (s *StateStore) RevenueUSD() float64 {
	cents := atomic.LoadUint64(&s.totalRevenue)
	return float64(cents) / 100.0
}
