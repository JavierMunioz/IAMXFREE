// Package inspection analyzes a directory on disk and reports what kind of
// project lives there — runtime, package manager, framework (when it can be
// inferred), and suggested commands — without ever running a command,
// installing a dependency, or modifying anything. It only reads files.
//
// It does not depend on internal/execution or internal/tui: it is meant to
// be reusable from anywhere in the project (the application-registration
// wizard is the first planned consumer, in a later iteration).
package inspection
