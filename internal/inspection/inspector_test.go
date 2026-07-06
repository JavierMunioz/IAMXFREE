package inspection

import (
	"testing"
)

func TestInspectReturnsNoDetectionsForEmptyDir(t *testing.T) {
	dir := t.TempDir()
	registry := NewRegistry()
	registry.Register(&fakeDetector{markerFile: "marker.txt"})

	result, err := NewInspector(registry).Inspect(dir)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if len(result.Detections) != 0 {
		t.Fatalf("Detections = %v, want empty", result.Detections)
	}
	if _, ok := result.Primary(); ok {
		t.Fatal("expected Primary() to report false with no detections")
	}
}

func TestInspectRunsEveryRegisteredDetector(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "a")
	writeFile(t, dir, "b.txt", "b")

	registry := NewRegistry()
	registry.Register(&fakeDetector{markerFile: "a.txt"})
	registry.Register(&fakeDetector{markerFile: "b.txt"})
	registry.Register(&fakeDetector{markerFile: "does-not-exist.txt"})

	result, err := NewInspector(registry).Inspect(dir)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if len(result.Detections) != 2 {
		t.Fatalf("len(Detections) = %d, want 2", len(result.Detections))
	}
}

func TestInspectReturnsErrorForMissingRoot(t *testing.T) {
	registry := NewRegistry()
	if _, err := NewInspector(registry).Inspect("/does/not/exist/at/all"); err == nil {
		t.Fatal("expected an error for a nonexistent root")
	}
}

// confidenceDetector lets tests control exactly which Confidence a fake
// detection reports, to exercise Result.Primary()'s tie-break rules.
type confidenceDetector struct {
	name       string
	confidence Confidence
}

func (d *confidenceDetector) Name() string { return d.name }
func (d *confidenceDetector) Detect(DetectionInput) (Detection, bool) {
	return Detection{Type: ProjectType(d.name), Confidence: d.confidence}, true
}

func TestResultPrimaryPicksHighestConfidence(t *testing.T) {
	dir := t.TempDir()
	registry := NewRegistry()
	registry.Register(&confidenceDetector{name: "low", confidence: ConfidenceLow})
	registry.Register(&confidenceDetector{name: "high", confidence: ConfidenceHigh})
	registry.Register(&confidenceDetector{name: "medium", confidence: ConfidenceMedium})

	result, err := NewInspector(registry).Inspect(dir)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}

	primary, ok := result.Primary()
	if !ok {
		t.Fatal("expected a primary detection")
	}
	if primary.Type != ProjectType("high") {
		t.Fatalf("Primary().Type = %q, want %q", primary.Type, "high")
	}
}

func TestResultPrimaryBreaksTiesByRegistrationOrder(t *testing.T) {
	dir := t.TempDir()
	registry := NewRegistry()
	registry.Register(&confidenceDetector{name: "first", confidence: ConfidenceHigh})
	registry.Register(&confidenceDetector{name: "second", confidence: ConfidenceHigh})

	result, err := NewInspector(registry).Inspect(dir)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}

	primary, ok := result.Primary()
	if !ok {
		t.Fatal("expected a primary detection")
	}
	if primary.Type != ProjectType("first") {
		t.Fatalf("Primary().Type = %q, want %q (first registered)", primary.Type, "first")
	}
}
