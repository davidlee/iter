package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetDefaultViceEnv(t *testing.T) {
	// Save original environment
	originalEnvs := map[string]string{
		"VICE_CONFIG":     os.Getenv("VICE_CONFIG"),
		"VICE_DATA":       os.Getenv("VICE_DATA"),
		"VICE_STATE":      os.Getenv("VICE_STATE"),
		"VICE_CACHE":      os.Getenv("VICE_CACHE"),
		"VICE_CONTEXT":    os.Getenv("VICE_CONTEXT"),
		"XDG_CONFIG_HOME": os.Getenv("XDG_CONFIG_HOME"),
		"XDG_DATA_HOME":   os.Getenv("XDG_DATA_HOME"),
		"XDG_STATE_HOME":  os.Getenv("XDG_STATE_HOME"),
		"XDG_CACHE_HOME":  os.Getenv("XDG_CACHE_HOME"),
	}
	defer func() {
		for key, value := range originalEnvs {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Clear all environment variables
	for key := range originalEnvs {
		os.Unsetenv(key)
	}

	env, err := GetDefaultViceEnv()
	if err != nil {
		t.Fatalf("GetDefaultViceEnv() failed: %v", err)
	}

	// Test default XDG paths (should use ~/.config/vice, ~/.local/share/vice, etc.)
	homeDir, _ := os.UserHomeDir()
	expectedConfigDir := filepath.Join(homeDir, ".config", "vice")
	expectedDataDir := filepath.Join(homeDir, ".local", "share", "vice")
	expectedStateDir := filepath.Join(homeDir, ".local", "state", "vice")
	expectedCacheDir := filepath.Join(homeDir, ".cache", "vice")

	if env.ConfigDir != expectedConfigDir {
		t.Errorf("ConfigDir = %q, want %q", env.ConfigDir, expectedConfigDir)
	}
	if env.DataDir != expectedDataDir {
		t.Errorf("DataDir = %q, want %q", env.DataDir, expectedDataDir)
	}
	if env.StateDir != expectedStateDir {
		t.Errorf("StateDir = %q, want %q", env.StateDir, expectedStateDir)
	}
	if env.CacheDir != expectedCacheDir {
		t.Errorf("CacheDir = %q, want %q", env.CacheDir, expectedCacheDir)
	}

	// Test default contexts
	expectedContexts := []string{"personal", "work"}
	if len(env.Contexts) != len(expectedContexts) {
		t.Errorf("Contexts length = %d, want %d", len(env.Contexts), len(expectedContexts))
	}
	for i, ctx := range expectedContexts {
		if env.Contexts[i] != ctx {
			t.Errorf("Contexts[%d] = %q, want %q", i, env.Contexts[i], ctx)
		}
	}

	// Test default context (first in array)
	if env.Context != "personal" {
		t.Errorf("Context = %q, want %q", env.Context, "personal")
	}

	// Test computed context data path
	expectedContextData := filepath.Join(expectedDataDir, "personal")
	if env.ContextData != expectedContextData {
		t.Errorf("ContextData = %q, want %q", env.ContextData, expectedContextData)
	}
}

func TestViceEnvWithXDGOverrides(t *testing.T) {
	// Save original environment
	originalEnvs := map[string]string{
		"XDG_CONFIG_HOME": os.Getenv("XDG_CONFIG_HOME"),
		"XDG_DATA_HOME":   os.Getenv("XDG_DATA_HOME"),
		"XDG_STATE_HOME":  os.Getenv("XDG_STATE_HOME"),
		"XDG_CACHE_HOME":  os.Getenv("XDG_CACHE_HOME"),
	}
	defer func() {
		for key, value := range originalEnvs {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Set XDG environment variables
	os.Setenv("XDG_CONFIG_HOME", "/custom/config")
	os.Setenv("XDG_DATA_HOME", "/custom/data")
	os.Setenv("XDG_STATE_HOME", "/custom/state")
	os.Setenv("XDG_CACHE_HOME", "/custom/cache")

	env, err := GetDefaultViceEnv()
	if err != nil {
		t.Fatalf("GetDefaultViceEnv() failed: %v", err)
	}

	// Test XDG overrides
	if env.ConfigDir != "/custom/config/vice" {
		t.Errorf("ConfigDir = %q, want %q", env.ConfigDir, "/custom/config/vice")
	}
	if env.DataDir != "/custom/data/vice" {
		t.Errorf("DataDir = %q, want %q", env.DataDir, "/custom/data/vice")
	}
	if env.StateDir != "/custom/state/vice" {
		t.Errorf("StateDir = %q, want %q", env.StateDir, "/custom/state/vice")
	}
	if env.CacheDir != "/custom/cache/vice" {
		t.Errorf("CacheDir = %q, want %q", env.CacheDir, "/custom/cache/vice")
	}
}

func TestViceEnvWithViceOverrides(t *testing.T) {
	// Save original environment
	originalEnvs := map[string]string{
		"VICE_CONFIG":  os.Getenv("VICE_CONFIG"),
		"VICE_DATA":    os.Getenv("VICE_DATA"),
		"VICE_STATE":   os.Getenv("VICE_STATE"),
		"VICE_CACHE":   os.Getenv("VICE_CACHE"),
		"VICE_CONTEXT": os.Getenv("VICE_CONTEXT"),
	}
	defer func() {
		for key, value := range originalEnvs {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Set VICE environment variables
	os.Setenv("VICE_CONFIG", "/vice/config")
	os.Setenv("VICE_DATA", "/vice/data")
	os.Setenv("VICE_STATE", "/vice/state")
	os.Setenv("VICE_CACHE", "/vice/cache")
	os.Setenv("VICE_CONTEXT", "work")

	env, err := GetDefaultViceEnv()
	if err != nil {
		t.Fatalf("GetDefaultViceEnv() failed: %v", err)
	}

	// Test VICE overrides (highest priority)
	if env.ConfigDir != "/vice/config" {
		t.Errorf("ConfigDir = %q, want %q", env.ConfigDir, "/vice/config")
	}
	if env.DataDir != "/vice/data" {
		t.Errorf("DataDir = %q, want %q", env.DataDir, "/vice/data")
	}
	if env.StateDir != "/vice/state" {
		t.Errorf("StateDir = %q, want %q", env.StateDir, "/vice/state")
	}
	if env.CacheDir != "/vice/cache" {
		t.Errorf("CacheDir = %q, want %q", env.CacheDir, "/vice/cache")
	}

	// Test VICE_CONTEXT override
	if env.Context != "work" {
		t.Errorf("Context = %q, want %q", env.Context, "work")
	}
	if env.ContextOverride != "work" {
		t.Errorf("ContextOverride = %q, want %q", env.ContextOverride, "work")
	}
	if env.ContextData != "/vice/data/work" {
		t.Errorf("ContextData = %q, want %q", env.ContextData, "/vice/data/work")
	}
}

func TestGetViceEnvWithOverrides(t *testing.T) {
	// Clear environment
	os.Unsetenv("VICE_CONFIG")
	os.Unsetenv("VICE_CONTEXT")

	// Use temp directory for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")

	env, err := GetViceEnvWithOverrides(configDir, "testing")
	if err != nil {
		t.Fatalf("GetViceEnvWithOverrides() failed: %v", err)
	}

	// Test CLI flag overrides
	if env.ConfigDir != configDir {
		t.Errorf("ConfigDir = %q, want %q", env.ConfigDir, configDir)
	}
	if env.ConfigDirOverride != configDir {
		t.Errorf("ConfigDirOverride = %q, want %q", env.ConfigDirOverride, configDir)
	}
	if env.Context != "testing" {
		t.Errorf("Context = %q, want %q", env.Context, "testing")
	}
	if env.ContextOverride != "testing" {
		t.Errorf("ContextOverride = %q, want %q", env.ContextOverride, "testing")
	}
	if !strings.Contains(env.ContextData, "testing") {
		t.Errorf("ContextData = %q, should contain %q", env.ContextData, "testing")
	}
}

func TestViceEnvPaths(t *testing.T) {
	env := &ViceEnv{
		ConfigDir:   "/test/config",
		DataDir:     "/test/data",
		StateDir:    "/test/state",
		Context:     "personal",
		ContextData: "/test/data/personal",
	}

	// Test path methods
	tests := []struct {
		method   func() string
		expected string
	}{
		{env.GetConfigTomlPath, "/test/config/config.toml"},
		{env.GetStateFilePath, "/test/state/vice.yml"},
		{env.GetHabitsFile, "/test/data/personal/habits.yml"},
		{env.GetEntriesFile, "/test/data/personal/entries.yml"},
		{env.GetChecklistsFile, "/test/data/personal/checklists.yml"},
		{env.GetChecklistEntriesFile, "/test/data/personal/checklist_entries.yml"},
	}

	for _, test := range tests {
		if got := test.method(); got != test.expected {
			t.Errorf("Path method returned %q, want %q", got, test.expected)
		}
	}
}

func TestEnsureDirectories(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	env := &ViceEnv{
		ConfigDir:   filepath.Join(tempDir, "config"),
		DataDir:     filepath.Join(tempDir, "data"),
		StateDir:    filepath.Join(tempDir, "state"),
		CacheDir:    filepath.Join(tempDir, "cache"),
		ContextData: filepath.Join(tempDir, "data", "personal"),
	}

	if err := env.EnsureDirectories(); err != nil {
		t.Fatalf("EnsureDirectories() failed: %v", err)
	}

	// Check that all directories were created
	dirs := []string{env.ConfigDir, env.DataDir, env.StateDir, env.CacheDir, env.ContextData}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory %q was not created", dir)
		}
	}
}