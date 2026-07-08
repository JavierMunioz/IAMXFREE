// Package tui contains the presentation layer built with Bubble Tea and
// Lipgloss. It renders application state produced by internal/core and
// translates keyboard input into commands, but holds no business logic.
package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/deployment"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/dashboard"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/deploymentexec"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/deploymentplan"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/detail"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/logs"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizards/application"
)

type screen int

const (
	screenDashboard screen = iota
	screenWizard
	screenDetail
	screenLogs
	screenDeploymentPlan
	screenDeploymentExec
)

// applicationRegisteredMsg is emitted once ApplicationService.Register
// succeeds for an application submitted through the create-application
// wizard.
type applicationRegisteredMsg struct {
	app *models.Application
}

// applicationRegistrationFailedMsg is emitted when converting the wizard's
// result into a draft, or persisting it, fails.
type applicationRegistrationFailedMsg struct {
	err error
}

// RootModel is the top-level Bubble Tea model. It owns which screen is
// active — the dashboard or the create-application wizard — and delegates
// the actual persistence decision to services.ApplicationService and the
// project-analysis decision to services.ApplicationSetupService.
type RootModel struct {
	service          services.ApplicationService
	execution        services.ExecutionService
	setup            services.ApplicationSetupService
	deploymentEngine *deployment.Engine

	screen         screen
	dashboard      dashboard.Model
	wizard         wizard.Model
	detail         detail.Model
	logs           logs.Model
	deploymentPlan deploymentplan.Model
	deploymentExec deploymentexec.Model
}

// NewRootModel builds the initial application model.
func NewRootModel(service services.ApplicationService, execution services.ExecutionService, setup services.ApplicationSetupService, deploymentEngine *deployment.Engine) RootModel {
	return RootModel{
		service:          service,
		execution:        execution,
		setup:            setup,
		deploymentEngine: deploymentEngine,
		screen:           screenDashboard,
		dashboard:        dashboard.New(service, execution),
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.dashboard.Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case wizard.CompletedMsg:
		m.screen = screenDashboard

		draft, err := application.DraftFromResult(msg.Result)
		if err != nil {
			m.dashboard = m.dashboard.SetError(err)
			return m, nil
		}
		return m, m.registerCmd(draft.ToApplication())

	case wizard.CancelledMsg:
		m.screen = screenDashboard
		m.dashboard = m.dashboard.SetStatus("Registration cancelled.")
		return m, nil

	case dashboard.OpenWizardMsg:
		m.screen = screenWizard
		m.wizard = wizard.New("New application", application.Steps(m.setup))
		return m, m.wizard.Init()

	case dashboard.OpenDetailMsg:
		m.screen = screenDetail
		m.detail = detail.New(m.service, m.execution, msg.AppID)
		return m, m.detail.Init()

	case detail.BackMsg:
		m.screen = screenDashboard
		return m, m.dashboard.Reload()

	case detail.OpenLogsMsg:
		m.screen = screenLogs
		m.logs = logs.New(m.execution, msg.AppID, msg.Session)
		return m, m.logs.Init()

	case logs.BackMsg:
		m.screen = screenDetail
		return m, nil

	case detail.OpenDeploymentPlanMsg:
		m.screen = screenDeploymentPlan
		m.deploymentPlan = deploymentplan.New(m.deploymentEngine, msg.AppID)
		return m, m.deploymentPlan.Init()

	case deploymentplan.BackMsg:
		m.screen = screenDetail
		return m, nil

	case deploymentplan.ExecutePlanMsg:
		m.screen = screenDeploymentExec
		m.deploymentExec = deploymentexec.New(m.deploymentEngine, msg.Plan)
		return m, m.deploymentExec.Init()

	case deploymentexec.BackMsg:
		m.screen = screenDeploymentPlan
		return m, nil

	case applicationRegisteredMsg:
		m.dashboard = m.dashboard.SetStatus(fmt.Sprintf("Application %q registered.", msg.app.Name))
		return m, m.dashboard.Reload()

	case applicationRegistrationFailedMsg:
		m.dashboard = m.dashboard.SetError(msg.err)
		return m, nil

	case tea.WindowSizeMsg:
		updated, cmd := m.dashboard.Update(msg)
		m.dashboard = updated.(dashboard.Model)
		return m, cmd
	}

	switch m.screen {
	case screenWizard:
		updated, cmd := m.wizard.Update(msg)
		m.wizard = updated.(wizard.Model)
		return m, cmd

	case screenDetail:
		updated, cmd := m.detail.Update(msg)
		m.detail = updated.(detail.Model)
		return m, cmd

	case screenLogs:
		updated, cmd := m.logs.Update(msg)
		m.logs = updated.(logs.Model)
		return m, cmd

	case screenDeploymentPlan:
		updated, cmd := m.deploymentPlan.Update(msg)
		m.deploymentPlan = updated.(deploymentplan.Model)
		return m, cmd

	case screenDeploymentExec:
		updated, cmd := m.deploymentExec.Update(msg)
		m.deploymentExec = updated.(deploymentexec.Model)
		return m, cmd
	}

	updated, cmd := m.dashboard.Update(msg)
	m.dashboard = updated.(dashboard.Model)
	return m, cmd
}

func (m RootModel) registerCmd(app *models.Application) tea.Cmd {
	service := m.service
	return func() tea.Msg {
		if err := service.Register(context.Background(), app); err != nil {
			return applicationRegistrationFailedMsg{err: err}
		}
		return applicationRegisteredMsg{app: app}
	}
}

func (m RootModel) View() string {
	switch m.screen {
	case screenWizard:
		return m.wizard.View()
	case screenDetail:
		return m.detail.View()
	case screenLogs:
		return m.logs.View()
	case screenDeploymentPlan:
		return m.deploymentPlan.View()
	case screenDeploymentExec:
		return m.deploymentExec.View()
	}
	return m.dashboard.View()
}

// Run starts the Bubble Tea program using the terminal's real stdin/stdout.
func Run(service services.ApplicationService, execution services.ExecutionService, setup services.ApplicationSetupService, deploymentEngine *deployment.Engine) error {
	program := tea.NewProgram(NewRootModel(service, execution, setup, deploymentEngine), tea.WithAltScreen())
	_, err := program.Run()
	return err
}
