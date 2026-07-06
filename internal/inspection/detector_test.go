package inspection

import "testing"

// fakeDetector is used only to exercise the Detector contract and
// DetectionInput's helpers in isolation from any real technology.
type fakeDetector struct {
	markerFile string
}

func (d *fakeDetector) Name() string { return "fake" }

func (d *fakeDetector) Detect(in DetectionInput) (Detection, bool) {
	if !in.Has(d.markerFile) {
		return Detection{}, false
	}
	data, err := in.ReadFile(d.markerFile)
	notes := []string(nil)
	if err != nil {
		notes = append(notes, "marker file could not be read")
	}
	return Detection{
		Type:         ProjectType("fake"),
		MatchedFiles: []string{d.markerFile},
		Name:         string(data),
		Confidence:   ConfidenceHigh,
		Notes:        notes,
	}, true
}

var _ Detector = (*fakeDetector)(nil)

func TestDetectionInputHas(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "marker.txt", "hello")

	input := buildInput(t, dir)
	if !input.Has("marker.txt") {
		t.Fatal("expected Has(marker.txt) to be true")
	}
	if input.Has("does-not-exist") {
		t.Fatal("expected Has(does-not-exist) to be false")
	}
}

func TestDetectionInputReadFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "marker.txt", "hello")

	input := buildInput(t, dir)
	data, err := input.ReadFile("marker.txt")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("ReadFile() = %q, want %q", data, "hello")
	}
}

func TestFakeDetectorNotFound(t *testing.T) {
	dir := t.TempDir()
	input := buildInput(t, dir)

	d := &fakeDetector{markerFile: "marker.txt"}
	if _, ok := d.Detect(input); ok {
		t.Fatal("expected Detect() to report false when the marker file is absent")
	}
}

func TestFakeDetectorFound(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "marker.txt", "hello")
	input := buildInput(t, dir)

	d := &fakeDetector{markerFile: "marker.txt"}
	detection, ok := d.Detect(input)
	if !ok {
		t.Fatal("expected Detect() to report true when the marker file is present")
	}
	if detection.Name != "hello" {
		t.Fatalf("Name = %q, want %q", detection.Name, "hello")
	}
	if len(detection.Notes) != 0 {
		t.Fatalf("Notes = %v, want empty", detection.Notes)
	}
}
