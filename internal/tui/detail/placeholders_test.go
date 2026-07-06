package detail

import (
	"strings"
	"testing"
)

func TestPlaceholderKeysReportNotImplemented(t *testing.T) {
	app := newTestApp()

	tests := []struct {
		key  string
		want string
	}{
		{"r", "Restart"},
		{"l", "Logs"},
		{"e", "Editing configuration"},
	}

	for _, tc := range tests {
		m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
		m2, cmd := m.handleKey(keyMsg(tc.key))
		if cmd != nil {
			t.Errorf("key %q: expected no command, got one", tc.key)
		}
		if !strings.Contains(m2.status, tc.want) {
			t.Errorf("key %q: status = %q, want it to mention %q", tc.key, m2.status, tc.want)
		}
		if !strings.Contains(m2.status, "not implemented yet") {
			t.Errorf("key %q: status = %q, want it to say not implemented yet", tc.key, m2.status)
		}
	}
}

func TestActionBarListsEveryKeyNoDeadKeys(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m, _ = update(t, m, appLoadedMsg{app: app})

	bar := m.renderActionBar()
	for _, want := range []string{"s:", "x:", "r:", "l:", "e:", "f5:", "b/esc:", "q:"} {
		if !strings.Contains(bar, want) {
			t.Errorf("action bar missing %q:\n%s", want, bar)
		}
	}
}

func TestViewIncludesActionBar(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m, _ = update(t, m, appLoadedMsg{app: app})

	if !strings.Contains(m.View(), "s: start") {
		t.Fatalf("expected the view to include the action bar, got:\n%s", m.View())
	}
}
