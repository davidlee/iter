package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestViceEnvIntegration tests the complete ViceEnv setup with config.toml creation.
func TestViceEnvIntegration(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	// Clear environment variables
	originalEnvs := map[string]string{
		"VICE_CONFIG": os.Getenv("VICE_CONFIG"),
		"VICE_DATA":   os.Getenv("VICE_DATA"),
		"VICE_STATE":  os.Getenv("VICE_STATE"),
		"VICE_CACHE":  os.Getenv("VICE_CACHE"),
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

	// Set test environment
	os.Setenv("VICE_CONFIG", filepath.Join(tempDir, "config"))
	os.Setenv("VICE_DATA", filepath.Join(tempDir, "data"))
	os.Setenv("VICE_STATE", filepath.Join(tempDir, "state"))
	os.Setenv("VICE_CACHE", filepath.Join(tempDir, "cache"))

	// Test complete ViceEnv setup
	env, err := GetViceEnvWithOverrides("", "")
	if err != nil {
		t.Fatalf("GetViceEnvWithOverrides() failed: %v", err)
	}

	// Verify directories were created
	expectedDirs := []string{
		env.ConfigDir,
		env.DataDir,
		env.StateDir,
		env.CacheDir,
		env.ContextData,
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory %q was not created", dir)
		}
	}

	// Verify config.toml was created
	configPath := env.GetConfigTomlPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config.toml was not created")
	}

	// Verify config.toml content
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load created config.toml: %v", err)
	}

	expectedContexts := []string{"personal", "work"}
	if len(config.Core.Contexts) != len(expectedContexts) {
		t.Errorf("Config contexts length = %d, want %d", len(config.Core.Contexts), len(expectedContexts))
	}

	// Verify ViceEnv loaded config correctly
	if len(env.Contexts) != len(expectedContexts) {
		t.Errorf("ViceEnv contexts length = %d, want %d", len(env.Contexts), len(expectedContexts))
	}

	// Verify default context
	if env.Context != "personal" {
		t.Errorf("ViceEnv context = %q, want %q", env.Context, "personal")
	}

	// Test context data paths
	expectedHabitsFile := filepath.Join(tempDir, "data", "personal", "habits.yml")
	if env.GetHabitsFile() != expectedHabitsFile {
		t.Errorf("GetHabitsFile() = %q, want %q", env.GetHabitsFile(), expectedHabitsFile)
	}
}

// TestViceEnvWithCustomConfig tests ViceEnv with existing custom config.toml.
func TestViceEnvWithCustomConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Create custom config.toml
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	customTOML := `[core]
contexts = ["home", "office", "travel"]
`
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(customTOML), 0o644); err != nil {
		t.Fatalf("Failed to write custom config: %v", err)
	}

	// Set environment to use temp directory
	os.Setenv("VICE_CONFIG", configDir)
	os.Setenv("VICE_DATA", filepath.Join(tempDir, "data"))
	os.Setenv("VICE_STATE", filepath.Join(tempDir, "state"))
	os.Setenv("VICE_CACHE", filepath.Join(tempDir, "cache"))
	defer func() {
		os.Unsetenv("VICE_CONFIG")
		os.Unsetenv("VICE_DATA")
		os.Unsetenv("VICE_STATE")
		os.Unsetenv("VICE_CACHE")
	}()

	// Test ViceEnv setup with existing config
	env, err := GetViceEnvWithOverrides("", "")
	if err != nil {
		t.Fatalf("GetViceEnvWithOverrides() with custom config failed: %v", err)
	}

	// Verify custom contexts were loaded
	expectedContexts := []string{"home", "office", "travel"}
	if len(env.Contexts) != len(expectedContexts) {
		t.Errorf("ViceEnv contexts length = %d, want %d", len(env.Contexts), len(expectedContexts))
	}
	for i, ctx := range expectedContexts {
		if env.Contexts[i] != ctx {
			t.Errorf("ViceEnv contexts[%d] = %q, want %q", i, env.Contexts[i], ctx)
		}
	}

	// Verify default context is first in custom list
	if env.Context != "home" {
		t.Errorf("ViceEnv context = %q, want %q", env.Context, "home")
	}

	// Verify context data path
	expectedContextData := filepath.Join(tempDir, "data", "home")
	if env.ContextData != expectedContextData {
		t.Errorf("ViceEnv ContextData = %q, want %q", env.ContextData, expectedContextData)
	}
}

// TestViceEnvWithContextOverride tests context override functionality.
func TestViceEnvWithContextOverride(t *testing.T) {
	tempDir := t.TempDir()

	// Set environment
	os.Setenv("VICE_CONFIG", filepath.Join(tempDir, "config"))
	os.Setenv("VICE_DATA", filepath.Join(tempDir, "data"))
	os.Setenv("VICE_STATE", filepath.Join(tempDir, "state"))
	os.Setenv("VICE_CACHE", filepath.Join(tempDir, "cache"))
	defer func() {
		os.Unsetenv("VICE_CONFIG")
		os.Unsetenv("VICE_DATA")
		os.Unsetenv("VICE_STATE")
		os.Unsetenv("VICE_CACHE")
	}()

	// Test with context override
	env, err := GetViceEnvWithOverrides("", "work")
	if err != nil {
		t.Fatalf("GetViceEnvWithOverrides() with context override failed: %v", err)
	}

	// Verify context override
	if env.Context != "work" {
		t.Errorf("ViceEnv context = %q, want %q", env.Context, "work")
	}
	if env.ContextOverride != "work" {
		t.Errorf("ViceEnv ContextOverride = %q, want %q", env.ContextOverride, "work")
	}

	// Verify context data path reflects override
	expectedContextData := filepath.Join(tempDir, "data", "work")
	if env.ContextData != expectedContextData {
		t.Errorf("ViceEnv ContextData = %q, want %q", env.ContextData, expectedContextData)
	}

	// Verify work context directory was created
	if _, err := os.Stat(expectedContextData); os.IsNotExist(err) {
		t.Error("Work context directory was not created")
	}
}