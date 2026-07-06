package planner

import "sync"

// Registry holds the Planner implementations available to a
// DeploymentPlanner. Adding support for a new technology means constructing
// its Planner and calling Register — nothing else in this package changes.
type Registry struct {
	mu       sync.RWMutex
	planners []Planner
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds planner to the registry, in the order a DeploymentPlanner
// will try them.
func (r *Registry) Register(planner Planner) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.planners = append(r.planners, planner)
}

// Planners returns a copy of every registered planner, in registration
// order. Callers are free to mutate the returned slice without affecting
// the registry.
func (r *Registry) Planners() []Planner {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Planner, len(r.planners))
	copy(out, r.planners)
	return out
}
