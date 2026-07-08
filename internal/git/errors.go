package git

import "errors"

// ErrEmptyPath is returned by every Manager method when given an empty
// path. It is a deliberate, explicit failure rather than silently falling
// back to the Host's own working directory — which could easily land on a
// git repository that is not the one the caller meant to inspect.
var ErrEmptyPath = errors.New("git: path is empty")
