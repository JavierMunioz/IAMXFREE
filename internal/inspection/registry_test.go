package inspection

import "testing"

func TestRegistryStartsEmpty(t *testing.T) {
	r := NewRegistry()
	if got := r.Detectors(); len(got) != 0 {
		t.Fatalf("Detectors() = %v, want empty", got)
	}
}

func TestRegistryRegisterPreservesOrder(t *testing.T) {
	r := NewRegistry()
	first := &fakeDetector{markerFile: "a.txt"}
	second := &fakeDetector{markerFile: "b.txt"}

	r.Register(first)
	r.Register(second)

	got := r.Detectors()
	if len(got) != 2 {
		t.Fatalf("len(Detectors()) = %d, want 2", len(got))
	}
	if got[0] != Detector(first) || got[1] != Detector(second) {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestRegistryDetectorsReturnsACopy(t *testing.T) {
	r := NewRegistry()
	r.Register(&fakeDetector{markerFile: "a.txt"})

	got := r.Detectors()
	got[0] = &fakeDetector{markerFile: "mutated.txt"}

	again := r.Detectors()
	if again[0].(*fakeDetector).markerFile != "a.txt" {
		t.Fatalf("mutating the returned slice affected the registry: %+v", again)
	}
}
