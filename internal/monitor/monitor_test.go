package monitor_test

import (
	"testing"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/monitor"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func testSession() execution.Session {
	return execution.Session{
		PID:        4242,
		StartedAt:  time.Now().Add(-90 * time.Second).UTC(),
		Command:    "npm",
		Args:       []string{"start"},
		WorkingDir: "/srv/apps/my-api",
		Status:     execution.StatusRunning,
		Runtime:    models.RuntimeNode,
	}
}

func TestSnapshotReportsRunningProcessWithKnownResources(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunningPID(4242, true).
		WithProcessResources(4242, runtimehost.ProcessResources{
			CPUPercent:      12.5,
			CPUPercentKnown: true,
			MemoryRSSBytes:  100 * 1024 * 1024,
			MemoryRSSKnown:  true,
			MemoryVSZBytes:  500 * 1024 * 1024,
			MemoryVSZKnown:  true,
		})

	snap, err := monitor.New(host).Snapshot(testSession())
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if snap.Process.PID != 4242 {
		t.Errorf("PID = %d, want 4242", snap.Process.PID)
	}
	if snap.Process.State != monitor.ProcessStateRunning {
		t.Errorf("State = %q, want %q", snap.Process.State, monitor.ProcessStateRunning)
	}
	if snap.Process.Uptime < 89*time.Second {
		t.Errorf("Uptime = %v, want at least ~90s", snap.Process.Uptime)
	}

	if !snap.Resources.CPUPercent.Available || snap.Resources.CPUPercent.Value != 12.5 {
		t.Errorf("CPUPercent = %+v, want available 12.5", snap.Resources.CPUPercent)
	}
	if !snap.Resources.MemoryRSS.Available || snap.Resources.MemoryRSS.Value != 100*1024*1024 {
		t.Errorf("MemoryRSS = %+v, want available %d", snap.Resources.MemoryRSS, 100*1024*1024)
	}
	if !snap.Resources.MemoryVSZ.Available || snap.Resources.MemoryVSZ.Value != 500*1024*1024 {
		t.Errorf("MemoryVSZ = %+v, want available %d", snap.Resources.MemoryVSZ, 500*1024*1024)
	}

	if snap.Info.WorkingDir != "/srv/apps/my-api" || snap.Info.Command != "npm" || snap.Info.Runtime != models.RuntimeNode {
		t.Errorf("Info = %+v, unexpected", snap.Info)
	}
}

func TestSnapshotReportsStoppedProcess(t *testing.T) {
	host := runtimehosttest.NewFakeHost().WithRunningPID(4242, false)

	snap, err := monitor.New(host).Snapshot(testSession())
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}
	if snap.Process.State != monitor.ProcessStateStopped {
		t.Errorf("State = %q, want %q", snap.Process.State, monitor.ProcessStateStopped)
	}
}

func TestSnapshotNeverFabricatesUnavailableResources(t *testing.T) {
	host := runtimehosttest.NewFakeHost().WithRunningPID(4242, true)

	snap, err := monitor.New(host).Snapshot(testSession())
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}
	if snap.Resources.CPUPercent.Available {
		t.Error("expected CPUPercent to be unavailable when the Host never reported it")
	}
	if snap.Resources.MemoryRSS.Available {
		t.Error("expected MemoryRSS to be unavailable when the Host never reported it")
	}
	if snap.Resources.MemoryVSZ.Available {
		t.Error("expected MemoryVSZ to be unavailable when the Host never reported it")
	}
}
