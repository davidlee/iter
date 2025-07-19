// Package config provides environment and configuration management for the vice application.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidlee/vice/internal/zk"
)

// AIDEV-NOTE: DefaultAppName already defined in paths.go, reusing

// ViceEnv holds the complete runtime environment configuration for the application.
// AIDEV-NOTE: file-paths-central; replaces config.Paths with full XDG compliance and context support
// AIDEV-NOTE: T028-xdg-compliance; implements full XDG Base Directory Specification with context awareness
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

	// Tool integrations
	ZK *zk.ZKExecutable // ZK tool integration (nil if unavailable)

	// Override flags (for priority resolution)
	ConfigDirOverride string // from CLI --config-dir flag
	DataDirOverride   string // from CLI --data-dir flag
	StateDirOverride  string // from CLI --state-dir flag
	CacheDirOverride  string // from CLI --cache-dir flag
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

	// Initialize ZK tool integration
	env.ZK = zk.NewZKExecutable()

	return env, nil
}

// DirectoryOverrides holds CLI flag overrides for all XDG directories.
// AIDEV-NOTE: T028/3.1-directory-flags; consolidates CLI flag overrides for clean function signature
type DirectoryOverrides struct {
	ConfigDir string
	DataDir   string
	StateDir  string
	CacheDir  string
	Context   string
}

// GetViceEnvWithOverrides creates a ViceEnv with CLI flag overrides applied.
// Priority order: CLI flags → ENV vars → config.toml → XDG defaults
// AIDEV-NOTE: T028-initialization-flow; complete ViceEnv setup with context initialization and directory creation
func GetViceEnvWithOverrides(overrides DirectoryOverrides) (*ViceEnv, error) {
	env, err := GetDefaultViceEnv()
	if err != nil {
		return nil, err
	}

	// Apply CLI flag overrides
	if overrides.ConfigDir != "" {
		env.ConfigDirOverride = overrides.ConfigDir
		env.ConfigDir = overrides.ConfigDir
	}
	if overrides.DataDir != "" {
		env.DataDirOverride = overrides.DataDir
		env.DataDir = overrides.DataDir
	}
	if overrides.StateDir != "" {
		env.StateDirOverride = overrides.StateDir
		env.StateDir = overrides.StateDir
	}
	if overrides.CacheDir != "" {
		env.CacheDirOverride = overrides.CacheDir
		env.CacheDir = overrides.CacheDir
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
	if overrides.Context != "" {
		env.ContextOverride = overrides.Context
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

// GetFlotsamDir returns the context-aware path to the flotsam directory.
// AIDEV-NOTE: T027/3.2-flotsam-paths; context-aware flotsam directory for markdown note storage
// AIDEV-NOTE: path-pattern-flotsam; follows same pattern as GetHabitsFile/GetEntriesFile for context isolation
func (env *ViceEnv) GetFlotsamDir() string {
	return filepath.Join(env.ContextData, "flotsam")
}

// GetFlotsamCacheDB returns the context-aware path to the flotsam SQLite cache database.
// AIDEV-NOTE: T027/3.2-flotsam-cache; ADR-004 SQLite cache strategy for performance
// AIDEV-NOTE: cache-db-future; will be used for SRS performance cache when Phase 4 (Core Operations) is implemented
func (env *ViceEnv) GetFlotsamCacheDB() string {
	return filepath.Join(env.ContextData, "flotsam.db")
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

// ZK Integration Methods
// AIDEV-NOTE: T041/4.1-zk-integration; ViceEnv methods for graceful ZK tool integration

// IsZKAvailable returns true if zk tool is available for use.
func (env *ViceEnv) IsZKAvailable() bool {
	return env.ZK != nil && env.ZK.Available()
}

// WarnZKUnavailable prints a warning if zk is unavailable (once per session).
// Use this for interactive sessions where user should know about missing functionality.
func (env *ViceEnv) WarnZKUnavailable() {
	if env.ZK != nil {
		env.ZK.WarnIfUnavailable()
	}
}

// ZKList delegates to zk list command with graceful degradation.
// Returns note paths or error with installation guidance.
func (env *ViceEnv) ZKList(filters ...string) ([]string, error) {
	if !env.IsZKAvailable() {
		return nil, fmt.Errorf("zk not available - install from https://github.com/zk-org/zk")
	}

	return env.ZK.List(filters...)
}

// ZKEdit delegates to zk edit command with graceful degradation.
// Returns error with installation guidance if zk unavailable.
func (env *ViceEnv) ZKEdit(paths ...string) error {
	if !env.IsZKAvailable() {
		return fmt.Errorf("zk not available - install from https://github.com/zk-org/zk")
	}

	return env.ZK.Edit(paths...)
}

// GetZKNotebookDir returns the path to the ZK notebook within current context.
// This is where .zk directory should be located.
func (env *ViceEnv) GetZKNotebookDir() string {
	return env.GetFlotsamDir() // ZK notebook is co-located with flotsam directory
}

// ValidateZKNotebook validates ZK notebook configuration for compatibility.
// Currently a NOOP placeholder for future enhancement (T046).
func (env *ViceEnv) ValidateZKNotebook() error {
	notebookDir := env.GetZKNotebookDir()
	configPath := filepath.Join(notebookDir, ".zk", "config.toml")

	return zk.ValidateZKConfig(configPath)
}
