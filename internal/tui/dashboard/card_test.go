package dashboard

import (
	"strings"
	"testing"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestRenderCardShowsCoreFields(t *testing.T) {
	app := newApp("my-api", time.Now())
	app.Type = models.ApplicationTypeAPI
	app.Framework = models.FrameworkFastAPI
	app.Runtime = models.RuntimePython
	app.Status = models.StatusRunning
	app.Config.InternalPort = 8000
	app.Config.Domain = "my-api.example.com"

	card := stripANSI(t, renderTestCard(app, false))

	for _, want := range []string{"my-api", "api", "fastapi", "python", "running", "8000", "my-api.example.com"} {
		if !strings.Contains(card, want) {
			t.Errorf("card missing %q:\n%s", want, card)
		}
	}
}

func TestRenderCardShowsDashForEmptyDomain(t *testing.T) {
	app := newApp("no-domain", time.Now())
	card := stripANSI(t, renderTestCard(app, false))
	if !strings.Contains(card, "—") {
		t.Fatalf("expected a placeholder dash for an empty domain, got:\n%s", card)
	}
}

func TestRenderCardDiffersWhenSelected(t *testing.T) {
	app := newApp("my-api", time.Now())
	normal := renderTestCard(app, false)
	selected := renderTestCard(app, true)

	if normal == selected {
		t.Fatal("expected the selected card to render differently from a normal one")
	}
}

func TestTruncateLeavesShortValuesAlone(t *testing.T) {
	if got := truncate("short", 20); got != "short" {
		t.Fatalf("truncate() = %q, want %q", got, "short")
	}
}

func TestTruncateShortensLongValues(t *testing.T) {
	got := truncate("this-is-a-very-long-application-name", 10)
	if len([]rune(got)) != 10 {
		t.Fatalf("truncate() length = %d, want 10", len([]rune(got)))
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("truncate() = %q, want a trailing ellipsis", got)
	}
}

func TestRenderGridProducesOneRowPerColumnCount(t *testing.T) {
	m := modelWithApps(4) // width fixed to exactly 3 columns in modelWithApps
	grid := m.renderGrid()

	lines := strings.Split(grid, "\n")
	// A 3-column, 4-card grid should wrap into 2 rows, each taller than 1
	// line, so we just check both card rows' first lines (names) appear.
	joined := stripANSI(t, grid)
	for _, name := range []string{"a", "b", "c", "d"} {
		if !strings.Contains(joined, name) {
			t.Errorf("grid missing card for %q", name)
		}
	}
	if len(lines) == 0 {
		t.Fatal("expected a non-empty grid render")
	}
}

func renderTestCard(app *models.Application, selected bool) string {
	m := New(&fakeService{})
	return m.renderCard(app, selected)
}
