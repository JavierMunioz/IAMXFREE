package monitor

import (
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// Monitor builds a RuntimeSnapshot for a Session by asking runtimehost.Host
// about its live operating-system state. It is reused unchanged by every
// execution.Strategy — none of them build their own RuntimeSnapshot.
type Monitor struct {
	host runtimehost.Host
}

// New builds a Monitor backed by host.
func New(host runtimehost.Host) *Monitor {
	return &Monitor{host: host}
}

// Snapshot observes session's process right now: whether it is still
// running, its resource usage, and the context it was started with. It
// never starts, stops, or otherwise modifies the process — only reads.
func (m *Monitor) Snapshot(session execution.Session) (RuntimeSnapshot, error) {
	running, err := m.host.IsProcessRunning(session.PID)
	if err != nil {
		return RuntimeSnapshot{}, err
	}

	state := ProcessStateStopped
	if running {
		state = ProcessStateRunning
	}

	resources, err := m.host.ProcessResources(session.PID)
	if err != nil {
		return RuntimeSnapshot{}, err
	}

	return RuntimeSnapshot{
		Process: ProcessInfo{
			PID:       session.PID,
			State:     state,
			StartedAt: session.StartedAt,
			Uptime:    time.Since(session.StartedAt),
		},
		Resources: ResourceUsage{
			CPUPercent: metricFromResource(resources.CPUPercentKnown, resources.CPUPercent),
			MemoryRSS:  metricFromResource(resources.MemoryRSSKnown, float64(resources.MemoryRSSBytes)),
			MemoryVSZ:  metricFromResource(resources.MemoryVSZKnown, float64(resources.MemoryVSZBytes)),
		},
		Info: RuntimeInfo{
			WorkingDir: session.WorkingDir,
			Command:    session.Command,
			Args:       session.Args,
			Runtime:    session.Runtime,
		},
	}, nil
}

func metricFromResource(known bool, value float64) Metric {
	if !known {
		return unavailableMetric()
	}
	return availableMetric(value)
}
