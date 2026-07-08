// Package deploymentexec is the deployment execution progress screen,
// opened from the deployment plan review screen. It depends only on
// *deployment.Engine and internal/operations — never on
// ApplicationService, ExecutionService, the Git Manager or the Nginx
// Manager directly. Execution is entirely sequential this iteration;
// there is no way to cancel a run in progress from this screen.
package deploymentexec

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/deployment"
	"github.com/JavierMunioz/IAMXFREE/internal/operations"
)

// BackMsg signals that the user asked to return to the deployment plan
// screen. This screen has no notion of "plan screen" beyond this —
// whatever hosts it decides what happens next.
type BackMsg struct{}

type buildFailedMsg struct {
	err error
}

// executionStartedMsg carries the compiled operations and a channel that
// will receive one OperationProgress per state transition until execution
// finishes, at which point the channel is closed.
type executionStartedMsg struct {
	ops []operations.Operation
	ch  <-chan operations.OperationProgress
}

type operationProgressMsg struct {
	progress operations.OperationProgress
}

type progressChannelClosedMsg struct{}

// Model is the deployment execution progress screen.
type Model struct {
	engine *deployment.Engine
	plan   deployment.DeploymentPlan

	buildErr error

	started   bool
	startedAt time.Time
	finished  bool
	results   []operations.OperationResult

	progressCh <-chan operations.OperationProgress

	width  int
	height int
}

// New builds the execution screen for plan, backed by engine.
func New(engine *deployment.Engine, plan deployment.DeploymentPlan) Model {
	return Model{engine: engine, plan: plan, width: 80, height: 24}
}

func (m Model) Init() tea.Cmd {
	return m.buildAndStartCmd()
}

// buildAndStartCmd compiles plan into operations and, if that succeeds,
// starts running them in a background goroutine — not a tea.Cmd itself,
// since a tea.Cmd only ever delivers one Msg when it returns, and this
// screen needs to keep receiving progress while the run continues. The
// goroutine closes ch once every operation has reached a terminal state.
func (m Model) buildAndStartCmd() tea.Cmd {
	engine := m.engine
	plan := m.plan
	return func() tea.Msg {
		ops, err := engine.BuildOperations(context.Background(), plan)
		if err != nil {
			return buildFailedMsg{err: err}
		}

		ch := make(chan operations.OperationProgress)
		go func() {
			defer close(ch)
			operations.NewExecutor().Execute(context.Background(), ops, func(p operations.OperationProgress) {
				ch <- p
			})
		}()

		return executionStartedMsg{ops: ops, ch: ch}
	}
}

// readProgressCmd blocks on ch and returns the next update (or
// progressChannelClosedMsg once it closes). Update re-issues this after
// every operationProgressMsg so the read loop keeps going.
func readProgressCmd(ch <-chan operations.OperationProgress) tea.Cmd {
	return func() tea.Msg {
		progress, ok := <-ch
		if !ok {
			return progressChannelClosedMsg{}
		}
		return operationProgressMsg{progress: progress}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case buildFailedMsg:
		m.buildErr = msg.err
		return m, nil

	case executionStartedMsg:
		m.started = true
		m.startedAt = time.Now().UTC()
		m.progressCh = msg.ch
		m.results = make([]operations.OperationResult, len(msg.ops))
		for i, op := range msg.ops {
			m.results[i] = operations.OperationResult{Name: op.Name, Component: op.Component, Method: op.Method, State: operations.StatePending}
		}
		return m, readProgressCmd(m.progressCh)

	case operationProgressMsg:
		m.results[msg.progress.Index] = msg.progress.Result
		return m, readProgressCmd(m.progressCh)

	case progressChannelClosedMsg:
		m.finished = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}
