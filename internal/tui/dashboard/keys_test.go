package dashboard

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func modelWithApps(n int) Model {
	m := New(&fakeService{})
	now := time.Now()
	for i := 0; i < n; i++ {
		m.apps = append(m.apps, newApp(string(rune('a'+i)), now))
	}
	m.width = cardWidth*3 + cardGap*2 // exactly 3 columns
	return m
}

func sendKey(t *testing.T, m Model, s string) (Model, tea.Cmd) {
	t.Helper()
	var msg tea.KeyMsg
	switch s {
	case "up":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		msg = tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		msg = tea.KeyMsg{Type: tea.KeyRight}
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		msg = tea.KeyMsg{Type: tea.KeyShiftTab}
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
	updated, cmd := m.Update(msg)
	return updated.(Model), cmd
}

func TestGridQuits(t *testing.T) {
	m := modelWithApps(3)
	_, cmd := sendKey(t, m, "q")
	if cmd == nil {
		t.Fatal("expected a quit command")
	}
}

func TestGridOpensWizard(t *testing.T) {
	m := modelWithApps(3)
	_, cmd := sendKey(t, m, "a")
	if cmd == nil {
		t.Fatal("expected a command")
	}
	if _, ok := cmd().(OpenWizardMsg); !ok {
		t.Fatal("expected OpenWizardMsg")
	}
}

func TestArrowRightMovesSelectionAndClampsAtEdge(t *testing.T) {
	m := modelWithApps(3) // 3 columns, so this is a single row
	m, _ = sendKey(t, m, "right")
	if m.selected != 1 {
		t.Fatalf("selected = %d, want 1", m.selected)
	}
	m, _ = sendKey(t, m, "right")
	if m.selected != 2 {
		t.Fatalf("selected = %d, want 2", m.selected)
	}
	m, _ = sendKey(t, m, "right")
	if m.selected != 2 {
		t.Fatalf("selected = %d, want 2 (clamped)", m.selected)
	}
}

func TestArrowDownMovesByColumnCount(t *testing.T) {
	m := modelWithApps(6) // 3 columns -> 2 rows
	m, _ = sendKey(t, m, "down")
	if m.selected != 3 {
		t.Fatalf("selected = %d, want 3", m.selected)
	}
	m, _ = sendKey(t, m, "up")
	if m.selected != 0 {
		t.Fatalf("selected = %d, want 0", m.selected)
	}
}

func TestTabWrapsAround(t *testing.T) {
	m := modelWithApps(3)
	m.selected = 2
	m, _ = sendKey(t, m, "tab")
	if m.selected != 0 {
		t.Fatalf("selected = %d, want 0 (wrapped)", m.selected)
	}
	m, _ = sendKey(t, m, "shift+tab")
	if m.selected != 2 {
		t.Fatalf("selected = %d, want 2 (wrapped back)", m.selected)
	}
}

func TestEnterOpensDetailAndEscCloses(t *testing.T) {
	m := modelWithApps(3)
	m, _ = sendKey(t, m, "enter")
	if m.mode != viewDetail {
		t.Fatal("expected enter to open the detail view")
	}
	m, _ = sendKey(t, m, "esc")
	if m.mode != viewGrid {
		t.Fatal("expected esc to return to the grid view")
	}
}

func TestEnterDoesNothingWithNoApplications(t *testing.T) {
	m := New(&fakeService{})
	m, _ = sendKey(t, m, "enter")
	if m.mode != viewGrid {
		t.Fatal("expected enter to be a no-op with zero applications")
	}
}

func TestEditAndDeleteShowNotImplementedStatus(t *testing.T) {
	m := modelWithApps(1)

	m, _ = sendKey(t, m, "e")
	if m.status == "" {
		t.Fatal("expected a status message after pressing e")
	}

	m, _ = sendKey(t, m, "d")
	if m.status == "" {
		t.Fatal("expected a status message after pressing d")
	}
}

func TestRefreshTriggersReload(t *testing.T) {
	m := modelWithApps(1)
	m, cmd := sendKey(t, m, "r")
	if cmd == nil {
		t.Fatal("expected refresh to return a reload command")
	}
	if m.status == "" {
		t.Fatal("expected a status message while refreshing")
	}
	if _, ok := cmd().(appsLoadedMsg); !ok {
		t.Fatal("expected the reload command to eventually produce appsLoadedMsg")
	}
}

func TestQuitFromDetailView(t *testing.T) {
	m := modelWithApps(1)
	m, _ = sendKey(t, m, "enter")
	_, cmd := sendKey(t, m, "q")
	if cmd == nil {
		t.Fatal("expected quit to work from the detail view too")
	}
}
