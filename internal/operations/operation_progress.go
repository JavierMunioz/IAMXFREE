package operations

// OperationProgress is one update the Executor emits while running: one
// Operation's current OperationResult, and its position among the whole
// run. A caller streaming these (e.g. a TUI progress screen) can render a
// live view without polling.
type OperationProgress struct {
	Index  int
	Total  int
	Result OperationResult
}
