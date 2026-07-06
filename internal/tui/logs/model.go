// Package logs is the real-time log viewer screen, opened from the
// application detail screen. It depends only on services.ExecutionService
// — never on internal/execution or internal/runtimehost directly.
package logs

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// DefaultBufferCapacity is how many log lines are retained for scrollback
// unless NewWithCapacity is used to override it.
const DefaultBufferCapacity = 2000

// BackMsg signals that the user asked to return to the detail screen. The
// logs view has no notion of "detail screen" beyond this — whatever hosts
// it decides what happens next.
type BackMsg struct{}

type logsOpenedMsg struct {
	stream services.LogStream
}

type logsOpenFailedMsg struct {
	err error
}

type logEventMsg struct {
	event services.LogEvent
}

type logStreamEndedMsg struct{}

// Model is the real-time log viewer.
type Model struct {
	executionService services.ExecutionService
	appID            string
	session          services.RunSession

	stream      services.LogStream
	buffer      *ringBuffer
	streamEnded bool
	openErr     error

	// followLive, when true, always shows the newest retained lines —
	// scrolling up disables it; scrolling back down to the bottom (or
	// pressing "end") re-enables it. topIndex is only meaningful while
	// followLive is false.
	followLive bool
	topIndex   int

	width  int
	height int
}

// New builds the logs view for appID's currently tracked session, using
// DefaultBufferCapacity for scrollback.
func New(executionService services.ExecutionService, appID string, session services.RunSession) Model {
	return NewWithCapacity(executionService, appID, session, DefaultBufferCapacity)
}

// NewWithCapacity is New with an explicit scrollback buffer capacity.
func NewWithCapacity(executionService services.ExecutionService, appID string, session services.RunSession, capacity int) Model {
	return Model{
		executionService: executionService,
		appID:            appID,
		session:          session,
		buffer:           newRingBuffer(capacity),
		followLive:       true,
		width:            80,
		height:           24,
	}
}

func (m Model) Init() tea.Cmd {
	return m.openLogsCmd()
}

func (m Model) openLogsCmd() tea.Cmd {
	service := m.executionService
	appID := m.appID
	session := m.session
	return func() tea.Msg {
		stream, err := service.OpenLogs(context.Background(), appID, session)
		if err != nil {
			return logsOpenFailedMsg{err: err}
		}
		return logsOpenedMsg{stream: stream}
	}
}

// readNextEventCmd blocks on stream's Events channel and returns the next
// event (or logStreamEndedMsg once the channel closes). Update re-issues
// this command after every event so the read loop keeps going.
func readNextEventCmd(stream services.LogStream) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-stream.Events()
		if !ok {
			return logStreamEndedMsg{}
		}
		return logEventMsg{event: event}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case logsOpenedMsg:
		m.stream = msg.stream
		return m, readNextEventCmd(m.stream)

	case logsOpenFailedMsg:
		m.openErr = msg.err
		return m, nil

	case logEventMsg:
		m.buffer.append(msg.event)
		return m, readNextEventCmd(m.stream)

	case logStreamEndedMsg:
		m.streamEnded = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}
