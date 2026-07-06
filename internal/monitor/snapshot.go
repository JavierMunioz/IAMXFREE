package monitor

import (
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// RuntimeSnapshot is a point-in-time observation of a Session's real
// operating-system state. Its three groups mirror how it will keep
// growing: more process detail, more resource metrics, more context —
// without reshaping what already exists.
type RuntimeSnapshot struct {
	Process   ProcessInfo
	Resources ResourceUsage
	Info      RuntimeInfo
}

// ProcessState is the observed lifecycle state of the process backing a
// Session, as seen from the operating system right now — independent of
// whatever Status the Session itself last recorded.
type ProcessState string

const (
	ProcessStateRunning ProcessState = "running"
	ProcessStateStopped ProcessState = "stopped"
)

// ProcessInfo is basic identity and timing for the observed process.
type ProcessInfo struct {
	PID       int
	State     ProcessState
	StartedAt time.Time
	Uptime    time.Duration
}

// ResourceUsage carries process resource metrics. Each is a Metric so
// "not available in this environment" can be represented explicitly
// instead of a fabricated zero.
type ResourceUsage struct {
	CPUPercent Metric
	MemoryRSS  Metric // bytes
	MemoryVSZ  Metric // bytes
}

// Metric is one observed numeric value that may not be available — e.g.
// CPU/memory sampling isn't supported on every platform this iteration.
// Value is meaningful only when Available is true; callers must check
// Available before displaying Value, never assume zero means "zero usage".
type Metric struct {
	Value     float64
	Available bool
}

// unavailableMetric is the zero value of Metric, spelled out for clarity at
// call sites that build one explicitly.
func unavailableMetric() Metric { return Metric{} }

func availableMetric(value float64) Metric { return Metric{Value: value, Available: true} }

// RuntimeInfo is the context the process was launched with — already known
// from the Session, so building it never requires an OS query.
type RuntimeInfo struct {
	WorkingDir string
	Command    string
	Args       []string
	Runtime    models.Runtime
}
