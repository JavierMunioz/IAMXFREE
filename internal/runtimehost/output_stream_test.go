package runtimehost_test

import (
	"context"
	"testing"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

func drainChunks(t *testing.T, stream runtimehost.OutputStream, deadline time.Duration) []runtimehost.OutputChunk {
	t.Helper()
	var chunks []runtimehost.OutputChunk
	timeout := time.After(deadline)
	for {
		select {
		case c, ok := <-stream.Chunks():
			if !ok {
				return chunks
			}
			chunks = append(chunks, c)
		case <-timeout:
			t.Fatalf("timed out waiting for the stream to close; got %d chunks so far", len(chunks))
		}
	}
}

func TestLinuxHostStreamOutputCapturesStdoutAndStderr(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	pid, err := host.StartProcess(context.Background(), runtimehost.Command{
		Name: "sh",
		Args: []string{"-c", "echo out1; echo err1 1>&2; echo out2"},
	})
	if err != nil {
		t.Fatalf("StartProcess() error = %v", err)
	}

	stream, err := host.StreamOutput(pid)
	if err != nil {
		t.Fatalf("StreamOutput() error = %v", err)
	}

	chunks := drainChunks(t, stream, 2*time.Second)
	if stream.Err() != nil {
		t.Fatalf("Err() = %v, want nil", stream.Err())
	}

	if len(chunks) != 3 {
		t.Fatalf("len(chunks) = %d, want 3: %+v", len(chunks), chunks)
	}
	for _, c := range chunks {
		if c.Time.IsZero() {
			t.Errorf("chunk %+v has a zero Time", c)
		}
	}

	// stdout and stderr are independent pipes: the shell writes to both in
	// order, but there's no guarantee about how their reads interleave
	// relative to each other, only within each stream individually.
	var stdoutLines, stderrLines []string
	for _, c := range chunks {
		switch c.Stream {
		case runtimehost.OutputStdout:
			stdoutLines = append(stdoutLines, c.Line)
		case runtimehost.OutputStderr:
			stderrLines = append(stderrLines, c.Line)
		default:
			t.Errorf("unexpected stream kind %q", c.Stream)
		}
	}
	if len(stdoutLines) != 2 || stdoutLines[0] != "out1" || stdoutLines[1] != "out2" {
		t.Errorf("stdout lines = %v, want [out1 out2]", stdoutLines)
	}
	if len(stderrLines) != 1 || stderrLines[0] != "err1" {
		t.Errorf("stderr lines = %v, want [err1]", stderrLines)
	}
}

func TestLinuxHostStreamOutputUnknownPID(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	if _, err := host.StreamOutput(999999); err == nil {
		t.Fatal("expected an error for a pid this Host never started")
	}
}

func TestLinuxHostStreamOutputReplaysBacklogAfterProcessExits(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	pid, err := host.StartProcess(context.Background(), runtimehost.Command{
		Name: "sh",
		Args: []string{"-c", "echo done"},
	})
	if err != nil {
		t.Fatalf("StartProcess() error = %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		running, err := host.IsProcessRunning(pid)
		if err != nil {
			t.Fatalf("IsProcessRunning() error = %v", err)
		}
		if !running {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	stream, err := host.StreamOutput(pid)
	if err != nil {
		t.Fatalf("StreamOutput() error = %v", err)
	}
	chunks := drainChunks(t, stream, 2*time.Second)
	if len(chunks) != 1 || chunks[0].Line != "done" {
		t.Fatalf("chunks = %+v, want a single %q chunk", chunks, "done")
	}
}

func TestLinuxHostStreamOutputSupportsMultipleSubscribers(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	pid, err := host.StartProcess(context.Background(), runtimehost.Command{
		Name: "sh",
		Args: []string{"-c", "echo hello"},
	})
	if err != nil {
		t.Fatalf("StartProcess() error = %v", err)
	}

	first, err := host.StreamOutput(pid)
	if err != nil {
		t.Fatalf("StreamOutput() error = %v", err)
	}
	second, err := host.StreamOutput(pid)
	if err != nil {
		t.Fatalf("StreamOutput() error = %v", err)
	}

	firstChunks := drainChunks(t, first, 2*time.Second)
	secondChunks := drainChunks(t, second, 2*time.Second)

	if len(firstChunks) != 1 || len(secondChunks) != 1 {
		t.Fatalf("expected both subscribers to see one chunk, got %d and %d", len(firstChunks), len(secondChunks))
	}
}

func TestLinuxHostStreamOutputCloseStopsDelivery(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	pid, err := host.StartProcess(context.Background(), runtimehost.Command{
		Name: "sh",
		Args: []string{"-c", "i=0; while [ $i -lt 40 ]; do echo line$i; i=$((i+1)); sleep 0.05; done"},
	})
	if err != nil {
		t.Fatalf("StartProcess() error = %v", err)
	}
	defer host.StopProcess(pid)

	stream, err := host.StreamOutput(pid)
	if err != nil {
		t.Fatalf("StreamOutput() error = %v", err)
	}

	select {
	case <-stream.Chunks():
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the first chunk")
	}

	if err := stream.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	deadline := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-stream.Chunks():
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("timed out waiting for the stream to close after Close()")
		}
	}
}
