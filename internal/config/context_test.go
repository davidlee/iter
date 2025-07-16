package config

import (
	"os"
	"path/filepath"
	"testing"
)

func createTestEnv(t *testing.T) *ViceEnv {
	tempDir := t.TempDir()
	
	env := &ViceEnv{
		ConfigDir:   filepath.Join(tempDir, "config"),
		DataDir:     filepath.Join(tempDir, "data"),
		StateDir:    filepath.Join(tempDir, "state"),
		CacheDir:    filepath.Join(tempDir, "cache"),
		Context:     "personal",
		ContextData: filepath.Join(tempDir, "data", "personal"),
		Contexts:    []string{"personal", "work", "travel"},
	}
	
	if err := env.EnsureDirectories(); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}
	
	return env
}

func TestLoadContextStateNoFile(t *testing.T) {
	env := createTestEnv(t)
	
	// No state file exists, should return default (first context)
	context, err := LoadContextState(env)
	if err != nil {
		t.Fatalf("LoadContextState() failed: %v", err)
	}
	
	if context != "personal" {
		t.Errorf("LoadContextState() = %q, want %q", context, "personal")
	}
}

func TestSaveAndLoadContextState(t *testing.T) {
	env := createTestEnv(t)
	
	// Save state
	err := SaveContextState(env, "work")
	if err != nil {
		t.Fatalf("SaveContextState() failed: %v", err)
	}
	
	// Verify state file was created
	stateFile := env.GetStateFilePath()
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("State file was not created")
	}
	
	// Load state back
	context, err := LoadContextState(env)
	if err != nil {
		t.Fatalf("LoadContextState() failed: %v", err)
	}
	
	if context != "work" {
		t.Errorf("LoadContextState() = %q, want %q", context, "work")
	}
}

func TestLoadContextStateInvalidContext(t *testing.T) {
	env := createTestEnv(t)
	
	// Save an invalid context
	err := SaveContextState(env, "invalid")
	if err != nil {
		t.Fatalf("SaveContextState() failed: %v", err)
	}
	
	// Loading should return default context instead of invalid one
	context, err := LoadContextState(env)
	if err != nil {
		t.Fatalf("LoadContextState() failed: %v", err)
	}
	
	if context != "personal" {
		t.Errorf("LoadContextState() with invalid stored context = %q, want %q", context, "personal")
	}
}

func TestSwitchContext(t *testing.T) {
	env := createTestEnv(t)
	
	// Switch to valid context
	err := SwitchContext(env, "work")
	if err != nil {
		t.Fatalf("SwitchContext() failed: %v", err)
	}
	
	// Verify context was updated
	if env.Context != "work" {
		t.Errorf("Context not updated, got %q, want %q", env.Context, "work")
	}
	
	// Verify ContextData was updated
	expectedContextData := filepath.Join(env.DataDir, "work")
	if env.ContextData != expectedContextData {
		t.Errorf("ContextData not updated, got %q, want %q", env.ContextData, expectedContextData)
	}
	
	// Verify directory was created
	if _, err := os.Stat(env.ContextData); os.IsNotExist(err) {
		t.Error("Work context directory was not created")
	}
	
	// Verify state was persisted
	loadedContext, err := LoadContextState(env)
	if err != nil {
		t.Fatalf("Failed to load saved context: %v", err)
	}
	if loadedContext != "work" {
		t.Errorf("Persisted context = %q, want %q", loadedContext, "work")
	}
}

func TestSwitchContextInvalid(t *testing.T) {
	env := createTestEnv(t)
	originalContext := env.Context
	
	// Try to switch to invalid context
	err := SwitchContext(env, "invalid")
	if err == nil {
		t.Error("SwitchContext() should fail for invalid context")
	}
	
	// Verify context didn't change
	if env.Context != originalContext {
		t.Errorf("Context changed on invalid switch, got %q, want %q", env.Context, originalContext)
	}
}

func TestSwitchContextWithOverride(t *testing.T) {
	env := createTestEnv(t)
	env.ContextOverride = "work" // Simulate CLI override
	
	// Switch context - should not persist when override is set
	err := SwitchContext(env, "travel")
	if err != nil {
		t.Fatalf("SwitchContext() failed: %v", err)
	}
	
	// Context should be updated in env
	if env.Context != "travel" {
		t.Errorf("Context not updated, got %q, want %q", env.Context, "travel")
	}
	
	// Create a new env to test persistence (simulating restart)
	newEnv := createTestEnv(t)
	loadedContext, err := LoadContextState(newEnv)
	if err != nil {
		t.Fatalf("Failed to load context: %v", err)
	}
	
	// Should still be default since override was active
	if loadedContext != "personal" {
		t.Errorf("Context was persisted despite override, got %q, want %q", loadedContext, "personal")
	}
}

func TestInitializeContextDefault(t *testing.T) {
	env := createTestEnv(t)
	
	// No override, no saved state - should use default
	err := InitializeContext(env)
	if err != nil {
		t.Fatalf("InitializeContext() failed: %v", err)
	}
	
	if env.Context != "personal" {
		t.Errorf("Context = %q, want %q", env.Context, "personal")
	}
	
	expectedContextData := filepath.Join(env.DataDir, "personal")
	if env.ContextData != expectedContextData {
		t.Errorf("ContextData = %q, want %q", env.ContextData, expectedContextData)
	}
}

func TestInitializeContextWithOverride(t *testing.T) {
	env := createTestEnv(t)
	env.ContextOverride = "work"
	
	// Save a different context to state
	err := SaveContextState(env, "travel")
	if err != nil {
		t.Fatalf("SaveContextState() failed: %v", err)
	}
	
	// Initialize - should use override, not saved state
	err = InitializeContext(env)
	if err != nil {
		t.Fatalf("InitializeContext() failed: %v", err)
	}
	
	if env.Context != "work" {
		t.Errorf("Context = %q, want %q (override should take precedence)", env.Context, "work")
	}
}

func TestInitializeContextWithSavedState(t *testing.T) {
	env := createTestEnv(t)
	
	// Save context to state
	err := SaveContextState(env, "travel")
	if err != nil {
		t.Fatalf("SaveContextState() failed: %v", err)
	}
	
	// Initialize - should use saved state
	err = InitializeContext(env)
	if err != nil {
		t.Fatalf("InitializeContext() failed: %v", err)
	}
	
	if env.Context != "travel" {
		t.Errorf("Context = %q, want %q (should load from saved state)", env.Context, "travel")
	}
	
	expectedContextData := filepath.Join(env.DataDir, "travel")
	if env.ContextData != expectedContextData {
		t.Errorf("ContextData = %q, want %q", env.ContextData, expectedContextData)
	}
}