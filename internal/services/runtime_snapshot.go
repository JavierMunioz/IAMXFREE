package services

import (
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/monitor"
)

// RuntimeSnapshot is what the TUI sees about a session's real
// operating-system state — a flat projection of monitor.RuntimeSnapshot, so
// internal/tui never needs to import internal/monitor or
// internal/runtimehost directly.
type RuntimeSnapshot struct {
	PID       int
	State     string
	StartedAt time.Time
	Uptime    time.Duration

	CPUPercent     Metric
	MemoryRSSBytes Metric
	MemoryVSZBytes Metric

	WorkingDir string
	Command    string
	Args       []string
	Runtime    models.Runtime
}

// Metric is one observed numeric value that may not be available on this
// platform or for this process. Callers must check Available before using
// Value — it is never a fabricated zero.
type Metric struct {
	Value     float64
	Available bool
}

func toRuntimeSnapshot(snap monitor.RuntimeSnapshot) RuntimeSnapshot {
	return RuntimeSnapshot{
		PID:       snap.Process.PID,
		State:     string(snap.Process.State),
		StartedAt: snap.Process.StartedAt,
		Uptime:    snap.Process.Uptime,

		CPUPercent:     toMetric(snap.Resources.CPUPercent),
		MemoryRSSBytes: toMetric(snap.Resources.MemoryRSS),
		MemoryVSZBytes: toMetric(snap.Resources.MemoryVSZ),

		WorkingDir: snap.Info.WorkingDir,
		Command:    snap.Info.Command,
		Args:       snap.Info.Args,
		Runtime:    snap.Info.Runtime,
	}
}

func toMetric(m monitor.Metric) Metric {
	return Metric{Value: m.Value, Available: m.Available}
}
