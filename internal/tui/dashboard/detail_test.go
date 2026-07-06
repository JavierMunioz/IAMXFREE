package dashboard

import (
	"strings"
	"testing"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestRenderDetailShowsFullApplicationData(t *testing.T) {
	app := newApp("my-api", time.Now())
	app.Source.LocalPath = "/srv/apps/my-api"
	app.Source.RepositoryURL = "https://github.com/user/my-api.git"
	app.Source.MainBranch = "main"
	app.Config.InternalPort = 8000
	app.Config.Domain = "my-api.example.com"

	m := New(&fakeService{})
	m.apps = []*models.Application{app}
	m.selected = 0

	view := stripANSI(t, m.renderDetail())

	for _, want := range []string{
		"my-api", "/srv/apps/my-api", "https://github.com/user/my-api.git",
		"main", "8000", "my-api.example.com",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("detail view missing %q:\n%s", want, view)
		}
	}
}

func TestRenderFooterIsContextSensitive(t *testing.T) {
	m := modelWithApps(1)

	gridFooter := m.renderFooter()
	if !strings.Contains(gridFooter, "new application") {
		t.Errorf("expected grid footer to mention creating an application, got: %s", gridFooter)
	}

	m.mode = viewDetail
	detailFooter := m.renderFooter()
	if !strings.Contains(detailFooter, "back") {
		t.Errorf("expected detail footer to mention going back, got: %s", detailFooter)
	}
	if detailFooter == gridFooter {
		t.Error("expected the footer to change between grid and detail modes")
	}
}

func TestBodyRendersDetailWhenModeIsDetail(t *testing.T) {
	m := modelWithApps(1)
	m.mode = viewDetail

	body := stripANSI(t, m.renderBody())
	if !strings.Contains(body, "a") { // card name from modelWithApps
		t.Fatalf("expected detail body to render the selected app, got:\n%s", body)
	}
}
