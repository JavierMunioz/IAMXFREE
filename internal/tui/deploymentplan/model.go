// Package deploymentplan is the deployment plan review screen, opened from
// the application detail screen. It depends only on *deployment.Engine —
// never on ApplicationService, ExecutionService, the Git Manager or the
// Nginx Manager directly. It only ever calls Engine.Plan: this screen
// never executes anything, only displays what a deployment would need.
package deploymentplan

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/deployment"
)

// BackMsg signals that the user asked to return to the detail screen. This
// screen has no notion of "detail screen" beyond this — whatever hosts it
// decides what happens next.
type BackMsg struct{}

// ExecutePlanMsg signals that the user asked to actually run this plan.
// This screen has no notion of "execution screen" beyond this — whatever
// hosts it decides what happens next.
type ExecutePlanMsg struct {
	Plan deployment.DeploymentPlan
}

type planLoadedMsg struct {
	plan deployment.DeploymentPlan
}

type planLoadFailedMsg struct {
	err error
}

// Model is the deployment plan review screen.
type Model struct {
	engine *deployment.Engine
	appID  string

	plan    deployment.DeploymentPlan
	loaded  bool
	loadErr error

	width  int
	height int
}

// New builds the deployment plan screen for appID, backed by engine.
func New(engine *deployment.Engine, appID string) Model {
	return Model{engine: engine, appID: appID, width: 80, height: 24}
}

func (m Model) Init() tea.Cmd {
	return m.loadPlanCmd()
}

func (m Model) loadPlanCmd() tea.Cmd {
	engine := m.engine
	appID := m.appID
	return func() tea.Msg {
		plan, err := engine.Plan(context.Background(), appID)
		if err != nil {
			return planLoadFailedMsg{err: err}
		}
		return planLoadedMsg{plan: plan}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case planLoadedMsg:
		m.plan = msg.plan
		m.loaded = true
		m.loadErr = nil
		return m, nil

	case planLoadFailedMsg:
		m.loadErr = msg.err
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}
