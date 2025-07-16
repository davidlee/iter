// Package config provides environment and configuration management for the vice application.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// AIDEV-NOTE: DefaultAppName already defined in paths.go, reusing

// ViceEnv holds the complete runtime environment configuration for the application.
// AIDEV-NOTE: file-paths-central; replaces config.Paths with full XDG compliance and context support
type ViceEnv struct {
	// XDG Base Directory paths
	ConfigDir string // $VICE_CONFIG || $XDG_CONFIG_HOME/vice || ~/.config/vice
	DataDir   string // $VICE_DATA || $XDG_DATA_HOME/vice || ~/.local/share/vice
	StateDir  string // $VICE_STATE || $XDG_STATE_HOME/vice || ~/.local/state/vice
	CacheDir  string // $VICE_CACHE || $XDG_CACHE_HOME/vice || ~/.cache/vice

	// Context management
	Context     string // active context name (from state, ENV override, or CLI flag)
	ContextData string // computed path: $DataDir/$Context

	// Configuration settings (loaded from config.toml or defaults)
	Contexts []string // available contexts from config.toml [core] section

	// Override flags (for priority resolution)
	ConfigDirOverride string // from CLI --config-dir flag
	ContextOverride   string // from CLI --context flag (transient)
}

// GetDefaultViceEnv creates a ViceEnv with XDG-compliant defaults and environment variable overrides.
// Priority order: ENV vars → XDG defaults
func GetDefaultViceEnv() (*ViceEnv, error) {
	env := &ViceEnv{}

	// Resolve XDG directories with environment variable overrides
	var err error
	env.ConfigDir, err = resolveXDGDir("VICE_CONFIG", "XDG_CONFIG_HOME", ".config")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config directory: %w", err)
	}

	env.DataDir, err = resolveXDGDir("VICE_DATA", "XDG_DATA_HOME", ".local/share")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve data directory: %w", err)
	}

	env.StateDir, err = resolveXDGDir("VICE_STATE", "XDG_STATE_HOME", ".local/state")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve state directory: %w", err)
	}

	env.CacheDir, err = resolveXDGDir("VICE_CACHE", "XDG_CACHE_HOME", ".cache")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve cache directory: %w", err)
	}

	// Initialize default contexts (will be overridden by config.toml loading)
	env.Contexts = []string{"personal", "work"}
	env.Context = env.Contexts[0] // first context is default

	// Check for VICE_CONTEXT environment variable override
	if envContext := os.Getenv("VICE_CONTEXT"); envContext != "" {
		env.ContextOverride = envContext
		env.Context = envContext
	}

	// Compute context data path
	env.ContextData = filepath.Join(env.DataDir, env.Context)

	return env, nil
}

// GetViceEnvWithOverrides creates a ViceEnv with CLI flag overrides applied.
// Priority order: CLI flags → ENV vars → config.toml → XDG defaults
func GetViceEnvWithOverrides(configDirOverride, contextOverride string) (*ViceEnv, error) {
	env, err := GetDefaultViceEnv()
	if err != nil {
		return nil, err
	}

	// Apply CLI flag overrides
	if configDirOverride != "" {
		env.ConfigDirOverride = configDirOverride
		env.ConfigDir = configDirOverride
	}

	// Ensure directories exist before loading config
	if err := env.EnsureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to ensure directories: %w", err)
	}

	// Ensure config.toml exists with defaults
	configPath := env.GetConfigTomlPath()
	if err := EnsureConfigToml(configPath); err != nil {
		return nil, fmt.Errorf("failed to ensure config.toml: %w", err)
	}

	// Load configuration from config.toml (replaces stub defaults)
	if err := LoadViceEnvConfig(env); err != nil {
		return nil, fmt.Errorf("failed to load config.toml: %w", err)
	}

	// Apply context override after config loading
	if contextOverride != "" {
		env.ContextOverride = contextOverride
	}

	// Initialize context based on overrides and persisted state
	if err := InitializeContext(env); err != nil {
		return nil, fmt.Errorf("failed to initialize context: %w", err)
	}

	return env, nil
}

// EnsureDirectories creates all required directories if they don't exist.
// Creates directories with appropriate permissions (0750).
func (env *ViceEnv) EnsureDirectories() error {
	dirs := []string{env.ConfigDir, env.DataDir, env.StateDir, env.CacheDir, env.ContextData}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetConfigTomlPath returns the path to config.toml in the config directory.
func (env *ViceEnv) GetConfigTomlPath() string {
	return filepath.Join(env.ConfigDir, "config.toml")
}

// GetStateFilePath returns the path to vice.yml in the state directory.
func (env *ViceEnv) GetStateFilePath() string {
	return filepath.Join(env.StateDir, "vice.yml")
}

// Context-aware data file paths (equivalent to current config.Paths fields)

// GetHabitsFile returns the context-aware path to habits.yml.
func (env *ViceEnv) GetHabitsFile() string {
	return filepath.Join(env.ContextData, "habits.yml")
}

// GetEntriesFile returns the context-aware path to entries.yml.
func (env *ViceEnv) GetEntriesFile() string {
	return filepath.Join(env.ContextData, "entries.yml")
}

// GetChecklistsFile returns the context-aware path to checklists.yml.
func (env *ViceEnv) GetChecklistsFile() string {
	return filepath.Join(env.ContextData, "checklists.yml")
}

// GetChecklistEntriesFile returns the context-aware path to checklist_entries.yml.
func (env *ViceEnv) GetChecklistEntriesFile() string {
	return filepath.Join(env.ContextData, "checklist_entries.yml")
}

// resolveXDGDir resolves an XDG directory with the given priority:
// 1. VICE_* environment variable (if set)
// 2. XDG_* environment variable + app name (if set)
// 3. ~/.{fallback}/vice (fallback path)
func resolveXDGDir(viceEnvVar, xdgEnvVar, fallbackDir string) (string, error) {
	// Check VICE_* environment variable first
	if viceDir := os.Getenv(viceEnvVar); viceDir != "" {
		return viceDir, nil
	}

	// Check XDG_* environment variable
	if xdgDir := os.Getenv(xdgEnvVar); xdgDir != "" {
		return filepath.Join(xdgDir, DefaultAppName), nil
	}

	// Fall back to ~/.{fallback}/vice
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(homeDir, fallbackDir, DefaultAppName), nil
}