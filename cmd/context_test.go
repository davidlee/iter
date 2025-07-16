package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"davidlee/vice/internal/config"
)

func TestContextCommands(t *testing.T) {
	// Save original environment
	originalEnvs := map[string]string{
		"VICE_CONFIG": os.Getenv("VICE_CONFIG"),
		"VICE_DATA":   os.Getenv("VICE_DATA"),
		"VICE_STATE":  os.Getenv("VICE_STATE"),
		"VICE_CACHE":  os.Getenv("VICE_CACHE"),
	}
	defer func() {
		for key, value := range originalEnvs {
			if value == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, value)
			}
		}
	}()

	// Setup test environment
	tempDir := t.TempDir()
	_ = os.Setenv("VICE_CONFIG", filepath.Join(tempDir, "config"))
	_ = os.Setenv("VICE_DATA", filepath.Join(tempDir, "data"))
	_ = os.Setenv("VICE_STATE", filepath.Join(tempDir, "state"))
	_ = os.Setenv("VICE_CACHE", filepath.Join(tempDir, "cache"))

	// Create test config.toml with custom contexts
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	testTOML := `[core]
contexts = ["home", "office", "travel"]
`
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(testTOML), 0o600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Initialize global viceEnv
	overrides := config.DirectoryOverrides{}
	var err error
	viceEnv, err = config.GetViceEnvWithOverrides(overrides)
	if err != nil {
		t.Fatalf("Failed to initialize ViceEnv: %v", err)
	}

	// Test context list
	err = runContextList(nil, nil)
	if err != nil {
		t.Errorf("runContextList() failed: %v", err)
	}

	// Test context show
	err = runContextShow(nil, nil)
	if err != nil {
		t.Errorf("runContextShow() failed: %v", err)
	}

	// Test context switch to valid context
	err = runContextSwitch(nil, []string{"office"})
	if err != nil {
		t.Errorf("runContextSwitch() to valid context failed: %v", err)
	}

	// Verify switch worked
	if viceEnv.Context != "office" {
		t.Errorf("Context switch failed: expected 'office', got '%s'", viceEnv.Context)
	}

	// Test context switch to invalid context
	err = runContextSwitch(nil, []string{"nonexistent"})
	if err == nil {
		t.Error("runContextSwitch() to invalid context should have failed")
	}

	// Test switch to already active context
	err = runContextSwitch(nil, []string{"office"})
	if err != nil {
		t.Errorf("runContextSwitch() to current context failed: %v", err)
	}
}

func TestContextCommandValidation(t *testing.T) {
	// Test that context switch requires exactly one argument
	cmd := contextSwitchCmd
	
	// Test no args
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("contextSwitchCmd should require exactly one argument")
	}

	// Test multiple args
	cmd.SetArgs([]string{"context1", "context2"})
	err = cmd.Execute()
	if err == nil {
		t.Error("contextSwitchCmd should require exactly one argument")
	}
}