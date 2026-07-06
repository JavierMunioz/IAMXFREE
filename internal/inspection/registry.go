package inspection

import "sync"

// Registry holds the Detector implementations available to an Inspector.
// Adding support for a new technology means constructing its Detector and
// calling Register — nothing else in this package changes.
type Registry struct {
	mu        sync.RWMutex
	detectors []Detector
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds detector to the registry, in the order Inspector will run
// them.
func (r *Registry) Register(detector Detector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.detectors = append(r.detectors, detector)
}

// Detectors returns a copy of every registered detector, in registration
// order. Callers are free to mutate the returned slice without affecting
// the registry.
func (r *Registry) Detectors() []Detector {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Detector, len(r.detectors))
	copy(out, r.detectors)
	return out
}
