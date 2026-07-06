package planner

// NewDefaultRegistry returns a Registry with every built-in Planner already
// registered. Only Node is implemented so far; adding the next technology
// (Python, Go, Docker, ...) never means changing this function's callers —
// only adding one more Register call here.
func NewDefaultRegistry() *Registry {
	registry := NewRegistry()
	registry.Register(NewNodePlanner())
	return registry
}
