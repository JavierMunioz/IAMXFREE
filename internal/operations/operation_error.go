package operations

// OperationError describes why an Operation failed: a message meant for
// display, and the underlying error for anything that needs to inspect or
// wrap it further.
type OperationError struct {
	Message string
	Cause   error
}

func (e *OperationError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *OperationError) Unwrap() error { return e.Cause }
