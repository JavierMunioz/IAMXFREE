package dashboard

import (
	"strings"
	"testing"
)

func TestRenderEmptyStateSuggestsPressingA(t *testing.T) {
	m := New(&fakeService{}, &fakeExecutionService{})
	view := stripANSI(t, m.renderEmptyState())

	if !strings.Contains(view, "No applications registered yet") {
		t.Errorf("expected a friendly empty-state title, got:\n%s", view)
	}
	if !strings.Contains(view, "a") {
		t.Errorf("expected a hint mentioning the 'a' key, got:\n%s", view)
	}
}

func TestBodyRendersEmptyStateWithNoApplications(t *testing.T) {
	m := New(&fakeService{}, &fakeExecutionService{})
	body := stripANSI(t, m.renderBody())

	if !strings.Contains(body, "No applications registered yet") {
		t.Fatalf("expected renderBody() to show the empty state, got:\n%s", body)
	}
}

func TestBodyRendersGridWithApplications(t *testing.T) {
	m := modelWithApps(1)
	body := stripANSI(t, m.renderBody())

	if strings.Contains(body, "No applications registered yet") {
		t.Fatal("expected renderBody() to show the grid, not the empty state")
	}
}
