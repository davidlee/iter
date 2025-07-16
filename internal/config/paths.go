// Package config provides configuration management for the vice application.
package config

import (
	"os"
	"path/filepath"
)

// DefaultAppName is the application name used in XDG paths.
const DefaultAppName = "vice"

// Paths holds the resolved configuration paths for the application.
// AIDEV-NOTE: file-paths-central; all YAML file paths defined here for consistency
type Paths struct {
	ConfigDir            string
	HabitsFile           string
	EntriesFile          string
	ChecklistsFile       string
	ChecklistEntriesFile string // AIDEV-NOTE: separates templates from daily instances
}

// GetDefaultPaths returns the default XDG-compliant paths for the application.
// It follows the XDG Base Directory Specification:
// - Uses XDG_CONFIG_HOME if set, otherwise defaults to ~/.config
// - Creates application-specific subdirectory: vice/
// - Returns paths for habits.yml and entries.yml files
func GetDefaultPaths() (*Paths, error) {
	configDir, err := getXDGConfigDir()
	if err != nil {
		return nil, err
	}

	appConfigDir := filepath.Join(configDir, DefaultAppName)

	return &Paths{
		ConfigDir:            appConfigDir,
		HabitsFile:           filepath.Join(appConfigDir, "habits.yml"),
		EntriesFile:          filepath.Join(appConfigDir, "entries.yml"),
		ChecklistsFile:       filepath.Join(appConfigDir, "checklists.yml"),
		ChecklistEntriesFile: filepath.Join(appConfigDir, "checklist_entries.yml"),
	}, nil
}

// GetPathsWithConfigDir returns paths using the specified config directory.
// This is used when the user provides a custom config directory via CLI flag.
func GetPathsWithConfigDir(configDir string) *Paths {
	return &Paths{
		ConfigDir:            configDir,
		HabitsFile:           filepath.Join(configDir, "habits.yml"),
		EntriesFile:          filepath.Join(configDir, "entries.yml"),
		ChecklistsFile:       filepath.Join(configDir, "checklists.yml"),
		ChecklistEntriesFile: filepath.Join(configDir, "checklist_entries.yml"),
	}
}

// EnsureConfigDir creates the config directory if it doesn't exist.
// It creates the directory with appropriate permissions (0750).
func (p *Paths) EnsureConfigDir() error {
	return os.MkdirAll(p.ConfigDir, 0o750)
}

// getXDGConfigDir returns the XDG config directory following the spec:
// - If XDG_CONFIG_HOME is set and non-empty, use it
// - Otherwise, use ~/.config (where ~ is the user's home directory)
func getXDGConfigDir() (string, error) {
	// Check XDG_CONFIG_HOME first
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return xdgConfigHome, nil
	}

	// Fall back to ~/.config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config"), nil
}
