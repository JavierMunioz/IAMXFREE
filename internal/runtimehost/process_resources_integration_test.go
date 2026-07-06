package runtimehost_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// These tests exercise LinuxHost.ProcessResources against the real /proc
// filesystem where available. IAMXFREE's dev machines are not necessarily
// Linux, so assertions branch on runtime.GOOS rather than assuming /proc
// exists — on a non-Linux platform, every metric must gracefully report
// itself as unavailable instead of erroring.

func TestLinuxHostProcessResourcesForCurrentProcess(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	resources, err := host.ProcessResources(os.Getpid())
	if err != nil {
		t.Fatalf("ProcessResources() error = %v", err)
	}

	if runtime.GOOS != "linux" {
		if resources.MemoryRSSKnown || resources.MemoryVSZKnown || resources.CPUPercentKnown {
			t.Fatalf("expected every metric to be unavailable on %s, got %+v", runtime.GOOS, resources)
		}
		return
	}

	if !resources.MemoryRSSKnown || resources.MemoryRSSBytes == 0 {
		t.Errorf("MemoryRSS = %d, known=%v, want a known nonzero value for the running test process", resources.MemoryRSSBytes, resources.MemoryRSSKnown)
	}
	if !resources.MemoryVSZKnown || resources.MemoryVSZBytes == 0 {
		t.Errorf("MemoryVSZ = %d, known=%v, want a known nonzero value for the running test process", resources.MemoryVSZBytes, resources.MemoryVSZKnown)
	}
}

func TestLinuxHostProcessResourcesUnknownPID(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	resources, err := host.ProcessResources(999999999)
	if err != nil {
		t.Fatalf("ProcessResources() error = %v, want nil (unavailable is not an error)", err)
	}
	if resources.CPUPercentKnown || resources.MemoryRSSKnown || resources.MemoryVSZKnown {
		t.Fatalf("expected every metric to be unavailable for a nonexistent pid, got %+v", resources)
	}
}
