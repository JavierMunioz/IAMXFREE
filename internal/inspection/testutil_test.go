package inspection

import (
	"os"
	"path/filepath"
	"testing"
)

// buildInput mirrors what Inspector.Inspect does before calling a Detector:
// list root once and hand detectors the resulting set of entries.
func buildInput(t *testing.T, root string) DetectionInput {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir(%s) error = %v", root, err)
	}

	input := DetectionInput{Root: root, Entries: make(map[string]bool, len(entries))}
	for _, entry := range entries {
		input.Entries[entry.Name()] = true
	}
	return input
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", name, err)
	}
}
