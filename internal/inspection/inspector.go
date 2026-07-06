package inspection

import (
	"fmt"
	"os"
)

// Result is everything the Inspector discovered about one directory. It may
// hold more than one Detection — e.g. a Node project that also ships a
// Dockerfile — the Inspector never has to pick a single "winner."
type Result struct {
	Root       string
	Detections []Detection
}

// Primary returns the Detection with the highest Confidence, or false if
// nothing was detected at all. Ties break in favor of whichever detector
// was registered first, the same rule execution.Resolver uses.
func (r Result) Primary() (Detection, bool) {
	if len(r.Detections) == 0 {
		return Detection{}, false
	}

	best := r.Detections[0]
	for _, d := range r.Detections[1:] {
		if confidenceRank(d.Confidence) > confidenceRank(best.Confidence) {
			best = d
		}
	}
	return best, true
}

func confidenceRank(c Confidence) int {
	switch c {
	case ConfidenceHigh:
		return 2
	case ConfidenceMedium:
		return 1
	default:
		return 0
	}
}

// Inspector analyzes a filesystem path and reports every technology its
// registered Detectors recognize. It performs no writes and runs no
// commands — only reads.
type Inspector struct {
	registry *Registry
}

// NewInspector builds an Inspector backed by registry.
func NewInspector(registry *Registry) *Inspector {
	return &Inspector{registry: registry}
}

// Inspect analyzes root and returns everything the registered detectors
// found there. It lists root exactly once and shares that listing with
// every detector.
func (i *Inspector) Inspect(root string) (Result, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return Result{}, fmt.Errorf("inspection: read %s: %w", root, err)
	}

	input := DetectionInput{Root: root, Entries: make(map[string]bool, len(entries))}
	for _, entry := range entries {
		input.Entries[entry.Name()] = true
	}

	var detections []Detection
	for _, detector := range i.registry.Detectors() {
		if detection, ok := detector.Detect(input); ok {
			detections = append(detections, detection)
		}
	}

	return Result{Root: root, Detections: detections}, nil
}
