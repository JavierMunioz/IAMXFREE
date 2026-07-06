package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultApplicationsDir returns the directory where the JSON application
// store keeps its files by default: "<user config dir>/iamxfree/applications".
func DefaultApplicationsDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config: resolve user config directory: %w", err)
	}
	return filepath.Join(base, "iamxfree", "applications"), nil
}
