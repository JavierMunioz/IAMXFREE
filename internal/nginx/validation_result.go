package nginx

// ValidationResult is the outcome of checking a config: either a
// VirtualHost's in-memory representation before it is ever written to
// disk, or Nginx's own judgment of the config currently on disk (`nginx
// -t`). Errors is never a guess — it is left empty whenever the check
// passed or its output could not be broken into individual problems, in
// which case Output still has the full detail.
type ValidationResult struct {
	Valid  bool
	Errors []string
	Output string
}
