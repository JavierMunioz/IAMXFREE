package execution_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
)

func TestNodeStrategyStopCallsHostStopProcess(t *testing.T) {
	host := healthyFakeHost().WithRunningPID(4242, true)
	strategy := execution.NewNodeStrategy(host)

	session := execution.Session{PID: 4242}
	if err := strategy.Stop(context.Background(), healthyNodeApp(), session); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if !host.Stopped(4242) {
		t.Fatal("expected the session's PID to have been stopped")
	}
}

func TestNodeStrategyStopPropagatesHostError(t *testing.T) {
	wantErr := errors.New("no such process")
	host := healthyFakeHost().WithRunningPID(4242, true).WithStopError(4242, wantErr)
	strategy := execution.NewNodeStrategy(host)

	session := execution.Session{PID: 4242}
	if err := strategy.Stop(context.Background(), healthyNodeApp(), session); !errors.Is(err, wantErr) {
		t.Fatalf("Stop() error = %v, want %v", err, wantErr)
	}
}
