package execution_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

func drainLogEvents(t *testing.T, stream execution.LogStream, deadline time.Duration) []execution.LogEvent {
	t.Helper()
	var events []execution.LogEvent
	timeout := time.After(deadline)
	for {
		select {
		case e, ok := <-stream.Events():
			if !ok {
				return events
			}
			events = append(events, e)
		case <-timeout:
			t.Fatalf("timed out waiting for the log stream to close; got %d events so far", len(events))
		}
	}
}

func TestNodeStrategyLogsAdaptsChunksAndAppendsEOF(t *testing.T) {
	chunks := []runtimehost.OutputChunk{
		{Stream: runtimehost.OutputStdout, Line: "listening on :3000", Time: time.Now()},
		{Stream: runtimehost.OutputStderr, Line: "deprecation warning", Time: time.Now()},
	}
	host := healthyFakeHost().WithOutputStream(4242, chunks, nil)
	strategy := execution.NewNodeStrategy(host)

	stream, err := strategy.Logs(context.Background(), healthyNodeApp(), execution.Session{PID: 4242})
	if err != nil {
		t.Fatalf("Logs() error = %v", err)
	}

	events := drainLogEvents(t, stream, 2*time.Second)
	if len(events) != 3 {
		t.Fatalf("len(events) = %d, want 3: %+v", len(events), events)
	}
	if events[0].Type != execution.LogEventStdout || events[0].Content != "listening on :3000" {
		t.Errorf("events[0] = %+v, want stdout %q", events[0], "listening on :3000")
	}
	if events[1].Type != execution.LogEventStderr || events[1].Content != "deprecation warning" {
		t.Errorf("events[1] = %+v, want stderr %q", events[1], "deprecation warning")
	}
	if events[2].Type != execution.LogEventEOF {
		t.Errorf("events[2].Type = %q, want %q", events[2].Type, execution.LogEventEOF)
	}
	if stream.Err() != nil {
		t.Fatalf("Err() = %v, want nil", stream.Err())
	}
}

func TestNodeStrategyLogsAppendsErrorEventOnAbnormalEnd(t *testing.T) {
	wantErr := errors.New("pipe broke")
	host := healthyFakeHost().WithOutputStream(4242, nil, wantErr)
	strategy := execution.NewNodeStrategy(host)

	stream, err := strategy.Logs(context.Background(), healthyNodeApp(), execution.Session{PID: 4242})
	if err != nil {
		t.Fatalf("Logs() error = %v", err)
	}

	events := drainLogEvents(t, stream, 2*time.Second)
	if len(events) != 1 || events[0].Type != execution.LogEventError {
		t.Fatalf("events = %+v, want a single error event", events)
	}
	if !errors.Is(stream.Err(), wantErr) {
		t.Fatalf("Err() = %v, want %v", stream.Err(), wantErr)
	}
}

func TestNodeStrategyLogsPropagatesStreamOutputError(t *testing.T) {
	host := healthyFakeHost() // no output stream configured for pid 4242
	strategy := execution.NewNodeStrategy(host)

	if _, err := strategy.Logs(context.Background(), healthyNodeApp(), execution.Session{PID: 4242}); err == nil {
		t.Fatal("expected an error when no output was ever captured for this pid")
	}
}

func TestNodeStrategyLogsCloseStopsDelivery(t *testing.T) {
	chunks := []runtimehost.OutputChunk{
		{Stream: runtimehost.OutputStdout, Line: "line1", Time: time.Now()},
	}
	host := healthyFakeHost().WithOutputStream(4242, chunks, nil)
	strategy := execution.NewNodeStrategy(host)

	stream, err := strategy.Logs(context.Background(), healthyNodeApp(), execution.Session{PID: 4242})
	if err != nil {
		t.Fatalf("Logs() error = %v", err)
	}

	if err := stream.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Draining after Close should terminate promptly rather than hang,
	// regardless of how many events were already in flight.
	drainLogEvents(t, stream, 2*time.Second)
}
