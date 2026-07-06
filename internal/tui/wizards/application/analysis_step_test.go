package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizards/application"
)

type fakeSetupService struct {
	proposal services.ApplicationSetupProposal
	err      error
	calls    []string
}

func (f *fakeSetupService) Inspect(_ context.Context, path string) (services.ApplicationSetupProposal, error) {
	f.calls = append(f.calls, path)
	return f.proposal, f.err
}

func TestAnalysisStepInspectsOnFirstFocus(t *testing.T) {
	setup := &fakeSetupService{proposal: services.ApplicationSetupProposal{ProjectType: "node"}}
	path := "/srv/apps/my-api"
	step := application.NewAnalysisStep(setup, func() string { return path })

	step.Focus()

	if len(setup.calls) != 1 || setup.calls[0] != path {
		t.Fatalf("calls = %v, want a single call with %q", setup.calls, path)
	}
	if step.Proposal().ProjectType != "node" {
		t.Fatalf("Proposal().ProjectType = %q, want %q", step.Proposal().ProjectType, "node")
	}
}

func TestAnalysisStepDoesNotReinspectSamePath(t *testing.T) {
	setup := &fakeSetupService{}
	path := "/srv/apps/my-api"
	step := application.NewAnalysisStep(setup, func() string { return path })

	step.Focus()
	step.Focus()
	step.Focus()

	if len(setup.calls) != 1 {
		t.Fatalf("calls = %v, want exactly 1 (same path should not re-inspect)", setup.calls)
	}
}

func TestAnalysisStepReinspectsWhenPathChanges(t *testing.T) {
	setup := &fakeSetupService{}
	path := "/srv/apps/my-api"
	step := application.NewAnalysisStep(setup, func() string { return path })

	step.Focus()
	path = "/srv/apps/other-api"
	step.Focus()

	if len(setup.calls) != 2 {
		t.Fatalf("calls = %v, want 2 (path changed)", setup.calls)
	}
	if setup.calls[1] != "/srv/apps/other-api" {
		t.Fatalf("second call path = %q, want %q", setup.calls[1], "/srv/apps/other-api")
	}
}

func TestAnalysisStepValidatePassesWhenInspectSucceeds(t *testing.T) {
	setup := &fakeSetupService{}
	step := application.NewAnalysisStep(setup, func() string { return "/srv/apps/my-api" })
	step.Focus()

	if err := step.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestAnalysisStepValidateFailsWhenInspectFails(t *testing.T) {
	wantErr := errors.New("no such directory")
	setup := &fakeSetupService{err: wantErr}
	step := application.NewAnalysisStep(setup, func() string { return "/does/not/exist" })
	step.Focus()

	if err := step.Validate(); !errors.Is(err, wantErr) {
		t.Fatalf("Validate() error = %v, want %v", err, wantErr)
	}
	if !strings.Contains(step.View(), wantErr.Error()) {
		t.Fatalf("expected the view to show the error, got:\n%s", step.View())
	}
}

func TestAnalysisStepViewShowsDetectedInformation(t *testing.T) {
	setup := &fakeSetupService{proposal: services.ApplicationSetupProposal{
		ProjectType:    "node",
		PackageManager: "npm",
		MatchedFiles:   []string{"package.json", "package-lock.json"},
		Confidence:     "high",
		Warnings:       []string{`no "start" script found`},
		Notes:          []string{"a Dockerfile was also found"},
	}}
	step := application.NewAnalysisStep(setup, func() string { return "/srv/apps/my-api" })
	step.Focus()

	view := step.View()
	for _, want := range []string{"node", "npm", "package.json", "high", "start", "Dockerfile"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q:\n%s", want, view)
		}
	}
}

func TestAnalysisStepIsNeverModal(t *testing.T) {
	step := application.NewAnalysisStep(&fakeSetupService{}, func() string { return "" })
	if step.Modal() {
		t.Fatal("AnalysisStep should never be modal")
	}
}
