package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test default contexts
	expectedContexts := []string{"personal", "work"}
	if len(config.Core.Contexts) != len(expectedContexts) {
		t.Errorf("Default contexts length = %d, want %d", len(config.Core.Contexts), len(expectedContexts))
	}
	for i, ctx := range expectedContexts {
		if config.Core.Contexts[i] != ctx {
			t.Errorf("Default contexts[%d] = %q, want %q", i, config.Core.Contexts[i], ctx)
		}
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	// Test with non-existent file
	config, err := LoadConfig("/nonexistent/config.toml")
	if err != nil {
		t.Fatalf("LoadConfig() with missing file failed: %v", err)
	}

	// Should return default config
	expected := DefaultConfig()
	if len(config.Core.Contexts) != len(expected.Core.Contexts) {
		t.Errorf("Missing file config contexts length = %d, want %d", len(config.Core.Contexts), len(expected.Core.Contexts))
	}
}

func TestLoadConfigValidFile(t *testing.T) {
	// Create temp file with valid TOML
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	validTOML := `[core]
contexts = ["home", "office", "travel"]
`
	if err := os.WriteFile(configPath, []byte(validTOML), 0o644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() with valid file failed: %v", err)
	}

	expectedContexts := []string{"home", "office", "travel"}
	if len(config.Core.Contexts) != len(expectedContexts) {
		t.Errorf("Loaded contexts length = %d, want %d", len(config.Core.Contexts), len(expectedContexts))
	}
	for i, ctx := range expectedContexts {
		if config.Core.Contexts[i] != ctx {
			t.Errorf("Loaded contexts[%d] = %q, want %q", i, config.Core.Contexts[i], ctx)
		}
	}
}

func TestLoadConfigInvalidTOML(t *testing.T) {
	// Create temp file with invalid TOML
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	invalidTOML := `[core
contexts = ["invalid"
`
	if err := os.WriteFile(configPath, []byte(invalidTOML), 0o644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() with invalid TOML should have failed")
	}
	if !strings.Contains(err.Error(), "failed to parse TOML") {
		t.Errorf("Expected TOML parsing error, got: %v", err)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name: "empty contexts",
			config: &Config{
				Core: CoreConfig{Contexts: []string{}},
			},
			wantErr: true,
			errMsg:  "at least one context must be defined",
		},
		{
			name: "empty context name",
			config: &Config{
				Core: CoreConfig{Contexts: []string{"valid", ""}},
			},
			wantErr: true,
			errMsg:  "context names cannot be empty",
		},
		{
			name: "duplicate contexts",
			config: &Config{
				Core: CoreConfig{Contexts: []string{"personal", "work", "personal"}},
			},
			wantErr: true,
			errMsg:  "duplicate context name",
		},
		{
			name: "valid config",
			config: &Config{
				Core: CoreConfig{Contexts: []string{"personal", "work", "travel"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateConfig() should have failed for %s", tt.name)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateConfig() failed for valid config: %v", err)
				}
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	config := &Config{
		Core: CoreConfig{
			Contexts: []string{"test1", "test2"},
		},
	}

	// Save config
	if err := SaveConfig(configPath, config); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "[core]") {
		t.Error("Saved config should contain [core] section")
	}
	if !strings.Contains(content, "contexts") {
		t.Error("Saved config should contain contexts")
	}
	if !strings.Contains(content, "test1") || !strings.Contains(content, "test2") {
		t.Error("Saved config should contain test contexts")
	}
}

func TestLoadViceEnvConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Create ViceEnv with temp directories
	env := &ViceEnv{
		ConfigDir:   tempDir,
		DataDir:     filepath.Join(tempDir, "data"),
		Context:     "initial",
		ContextData: filepath.Join(tempDir, "data", "initial"),
		Contexts:    []string{"initial"},
	}

	// Create config.toml
	configPath := env.GetConfigTomlPath()
	customTOML := `[core]
contexts = ["custom1", "custom2", "custom3"]
`
	if err := os.WriteFile(configPath, []byte(customTOML), 0o644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config into ViceEnv
	if err := LoadViceEnvConfig(env); err != nil {
		t.Fatalf("LoadViceEnvConfig() failed: %v", err)
	}

	// Verify contexts were updated
	expectedContexts := []string{"custom1", "custom2", "custom3"}
	if len(env.Contexts) != len(expectedContexts) {
		t.Errorf("ViceEnv contexts length = %d, want %d", len(env.Contexts), len(expectedContexts))
	}
	for i, ctx := range expectedContexts {
		if env.Contexts[i] != ctx {
			t.Errorf("ViceEnv contexts[%d] = %q, want %q", i, env.Contexts[i], ctx)
		}
	}

	// Verify context was reset to first valid context
	if env.Context != "custom1" {
		t.Errorf("ViceEnv context = %q, want %q", env.Context, "custom1")
	}
}

func TestEnsureConfigToml(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	// Ensure config file is created
	if err := EnsureConfigToml(configPath); err != nil {
		t.Fatalf("EnsureConfigToml() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config.toml was not created")
	}

	// Verify file content
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load created config: %v", err)
	}

	expected := DefaultConfig()
	if len(config.Core.Contexts) != len(expected.Core.Contexts) {
		t.Errorf("Created config contexts length = %d, want %d", len(config.Core.Contexts), len(expected.Core.Contexts))
	}

	// Test that existing file is not overwritten
	customTOML := `[core]
contexts = ["existing"]
`
	if err := os.WriteFile(configPath, []byte(customTOML), 0o644); err != nil {
		t.Fatalf("Failed to write custom config: %v", err)
	}

	// Ensure should not overwrite
	if err := EnsureConfigToml(configPath); err != nil {
		t.Fatalf("EnsureConfigToml() failed on existing file: %v", err)
	}

	// Verify file was not overwritten
	config, err = LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load existing config: %v", err)
	}

	if len(config.Core.Contexts) != 1 || config.Core.Contexts[0] != "existing" {
		t.Error("EnsureConfigToml() overwrote existing config file")
	}
}