package runtimehost_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

func TestToolAvailabilityFound(t *testing.T) {
	found := runtimehost.ToolAvailability{Status: runtimehost.AvailabilityFound}
	if !found.Found() {
		t.Error("expected Found() to be true for AvailabilityFound")
	}

	notFound := runtimehost.ToolAvailability{Status: runtimehost.AvailabilityNotFound}
	if notFound.Found() {
		t.Error("expected Found() to be false for AvailabilityNotFound")
	}
}

func TestExecutionErrorMessageIncludesStderr(t *testing.T) {
	err := &runtimehost.ExecutionError{
		Command:  "npm",
		Args:     []string{"install"},
		ExitCode: 1,
		Stderr:   "network error",
	}

	msg := err.Error()
	for _, want := range []string{"npm", "install", "1", "network error"} {
		if !strings.Contains(msg, want) {
			t.Errorf("Error() = %q, missing %q", msg, want)
		}
	}
}

func TestExecutionErrorMessageFallsBackToErr(t *testing.T) {
	cause := errors.New("executable not found")
	err := &runtimehost.ExecutionError{
		Command:  "missing-tool",
		ExitCode: -1,
		Err:      cause,
	}

	if !strings.Contains(err.Error(), "executable not found") {
		t.Errorf("Error() = %q, want it to mention the underlying cause", err.Error())
	}
}

func TestExecutionErrorUnwrap(t *testing.T) {
	cause := errors.New("boom")
	err := &runtimehost.ExecutionError{Err: cause}

	if !errors.Is(err, cause) {
		t.Error("expected errors.Is to find the wrapped cause")
	}
}
