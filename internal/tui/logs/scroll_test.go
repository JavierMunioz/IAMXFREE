package logs

import (
	"strings"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func modelWithLines(n int, visibleHeight int) Model {
	m := New(&fakeExecutionService{}, "app-1", services.RunSession{})
	m.height = visibleHeight + 4 // renderVisibleLines reserves 4 lines of chrome
	for i := 0; i < n; i++ {
		m.buffer.append(services.LogEvent{Type: "stdout", Content: "line" + string(rune('a'+i))})
	}
	return m
}

func TestScrollUpDisablesFollowLive(t *testing.T) {
	m := modelWithLines(20, 5)

	m = m.scrollBy(-1)
	if m.followLive {
		t.Fatal("expected scrolling up to disable followLive")
	}
}

func TestScrollUpThenDownReturnsToFollowLive(t *testing.T) {
	m := modelWithLines(20, 5)

	m = m.scrollBy(-3)
	if m.followLive {
		t.Fatal("expected followLive to be disabled after scrolling up")
	}
	m = m.scrollBy(3)
	if !m.followLive {
		t.Fatal("expected scrolling back down to the bottom to re-enable followLive")
	}
}

func TestScrollUpClampsAtOldestLine(t *testing.T) {
	m := modelWithLines(10, 5)

	for i := 0; i < 20; i++ {
		m = m.scrollBy(-1)
	}
	top := m.currentTop(m.buffer.len())
	if top != 0 {
		t.Fatalf("top = %d, want 0 (clamped at the oldest line)", top)
	}
}

func TestScrollToStartShowsOldestLines(t *testing.T) {
	m := modelWithLines(20, 5)
	m = m.scrollToStart()

	if m.followLive {
		t.Fatal("expected scrollToStart to disable followLive")
	}
	view := stripANSI(t, m.View())
	if !strings.Contains(view, "linea") {
		t.Fatalf("expected the oldest line to be visible, got:\n%s", view)
	}
}

func TestScrollToEndReenablesFollowLive(t *testing.T) {
	m := modelWithLines(20, 5)
	m = m.scrollToStart()
	m = m.scrollToEnd()

	if !m.followLive {
		t.Fatal("expected scrollToEnd to re-enable followLive")
	}
}

func TestPageUpAndPageDownScrollByVisibleHeight(t *testing.T) {
	m := modelWithLines(20, 5)
	m = m.scrollBy(-m.visibleHeight())

	total := m.buffer.len()
	top := m.currentTop(total)
	wantTop := total - m.visibleHeight() - m.visibleHeight()
	if wantTop < 0 {
		wantTop = 0
	}
	if top != wantTop {
		t.Fatalf("top = %d, want %d", top, wantTop)
	}
}

func TestScrollKeysUpdateModel(t *testing.T) {
	m := modelWithLines(20, 5)

	m, _ = update(t, m, keyMsg("up"))
	if m.followLive {
		t.Fatal("expected the up key to disable followLive")
	}

	m, _ = update(t, m, keyMsg("end"))
	if !m.followLive {
		t.Fatal("expected the end key to re-enable followLive")
	}
}

func TestHomeAndEndKeys(t *testing.T) {
	m := modelWithLines(20, 5)

	m, _ = update(t, m, keyMsg("home"))
	if m.followLive || m.topIndex != 0 {
		t.Fatalf("expected home to jump to the start, got followLive=%v topIndex=%d", m.followLive, m.topIndex)
	}

	m, _ = update(t, m, keyMsg("end"))
	if !m.followLive {
		t.Fatal("expected end to re-enable followLive")
	}
}
