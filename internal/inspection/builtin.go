package inspection

// NewDefaultRegistry returns a Registry with every built-in Detector already
// registered: Node, Python, Go, PHP, Rust, Java and Docker. Adding a new
// technology never means changing this function's callers — only adding one
// more Register call here.
func NewDefaultRegistry() *Registry {
	registry := NewRegistry()
	registry.Register(NewNodeDetector())
	registry.Register(NewPythonDetector())
	registry.Register(NewGoDetector())
	registry.Register(NewPHPDetector())
	registry.Register(NewRustDetector())
	registry.Register(NewJavaDetector())
	registry.Register(NewDockerDetector())
	return registry
}
