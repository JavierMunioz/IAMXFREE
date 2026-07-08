package nginx

import (
	"context"
	"errors"
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// ValidateConfig runs `nginx -t` against the config currently on disk and
// reports what it found. Unlike ValidateVirtualHost, this checks the real
// installation through the Manager's Host — it is Nginx's own judgment,
// not this package's.
func (m *Manager) ValidateConfig(ctx context.Context) (ValidationResult, error) {
	result, runErr := m.host.RunCaptured(ctx, runtimehost.Command{Name: "nginx", Args: []string{"-t"}})
	output := strings.TrimSpace(result.Stdout + result.Stderr)

	if runErr != nil {
		var execErr *runtimehost.ExecutionError
		// ExitCode == -1 means nginx never actually ran (e.g. not on
		// PATH) — a real failure to report, not a config verdict.
		if !errors.As(runErr, &execErr) || execErr.ExitCode < 0 {
			return ValidationResult{}, runErr
		}
	}

	return ValidationResult{
		Valid:  runErr == nil,
		Errors: extractNginxErrorLines(output),
		Output: output,
	}, nil
}

// extractNginxErrorLines pulls out the lines of output that name an actual
// problem, so a caller doesn't have to grep nginx's raw text itself.
func extractNginxErrorLines(output string) []string {
	var errs []string
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "[emerg]") || strings.Contains(line, "[error]") {
			errs = append(errs, strings.TrimSpace(line))
		}
	}
	return errs
}
