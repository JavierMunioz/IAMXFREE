package execution

import "sync"

// Registry holds the Strategy implementations available to a Resolver.
// Adding support for a new technology means constructing its Strategy and
// calling Register — nothing else in this package changes.
type Registry struct {
	mu         sync.RWMutex
	strategies []Strategy
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds strategy to the registry. Registration order matters only
// in that a Resolver tries strategies in that order and stops at the first
// match — if two strategies could both handle the same application, the
// one registered first wins.
func (r *Registry) Register(strategy Strategy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.strategies = append(r.strategies, strategy)
}

// Strategies returns a copy of every registered strategy, in registration
// order. Callers are free to mutate the returned slice without affecting
// the registry.
func (r *Registry) Strategies() []Strategy {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Strategy, len(r.strategies))
	copy(out, r.strategies)
	return out
}
