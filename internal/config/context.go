// Package config provides context management for the vice application.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ContextState represents the persisted state in vice.yml.
// AIDEV-NOTE: T028/2.1-state-persistence; tracks active context between invocations
// AIDEV-NOTE: T028-state-yaml-structure; simple version + active_context for future extensibility
type ContextState struct {
	Version       string `yaml:"version"`
	ActiveContext string `yaml:"active_context"`
}

// LoadContextState loads the active context from the state file.
// Returns default context (first in contexts array) if state file doesn't exist.
func LoadContextState(env *ViceEnv) (string, error) {
	stateFile := env.GetStateFilePath()
	
	// Check if state file exists
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		// No state file, return default context (first in array)
		if len(env.Contexts) > 0 {
			return env.Contexts[0], nil
		}
		return "personal", nil // fallback default
	}

	// Read state file
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return "", fmt.Errorf("failed to read state file %s: %w", stateFile, err)
	}

	// Parse YAML
	var state ContextState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return "", fmt.Errorf("failed to parse state file %s: %w", stateFile, err)
	}

	// Validate that the stored context exists in available contexts
	for _, ctx := range env.Contexts {
		if ctx == state.ActiveContext {
			return state.ActiveContext, nil
		}
	}

	// Stored context no longer valid, return default
	if len(env.Contexts) > 0 {
		return env.Contexts[0], nil
	}
	return "personal", nil
}

// SaveContextState saves the active context to the state file.
func SaveContextState(env *ViceEnv, activeContext string) error {
	state := ContextState{
		Version:       "1.0",
		ActiveContext: activeContext,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Ensure state directory exists
	stateFile := env.GetStateFilePath()
	if err := os.MkdirAll(filepath.Dir(stateFile), 0o750); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(stateFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write state file %s: %w", stateFile, err)
	}

	return nil
}

// SwitchContext updates the ViceEnv to use a new context and persists the change.
// This handles the complete context switching process including state persistence.
func SwitchContext(env *ViceEnv, newContext string) error {
	// Validate context exists
	contextValid := false
	for _, ctx := range env.Contexts {
		if ctx == newContext {
			contextValid = true
			break
		}
	}
	if !contextValid {
		return fmt.Errorf("context '%s' not found in available contexts %v", newContext, env.Contexts)
	}

	// Update ViceEnv
	env.Context = newContext
	env.ContextData = filepath.Join(env.DataDir, newContext)

	// Ensure new context directory exists
	if err := env.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create context directories: %w", err)
	}

	// Save state (unless this is a transient override)
	if env.ContextOverride == "" {
		if err := SaveContextState(env, newContext); err != nil {
			return fmt.Errorf("failed to save context state: %w", err)
		}
	}

	return nil
}

// InitializeContext sets up the context in ViceEnv based on overrides and persisted state.
// Priority: ContextOverride (CLI/ENV) → Persisted State → Default (first context)
// AIDEV-NOTE: T028-priority-resolution; implements ENV vars → CLI flags → config.toml → XDG defaults hierarchy
func InitializeContext(env *ViceEnv) error {
	var activeContext string
	var err error

	// Check for override (CLI flag or ENV variable)
	if env.ContextOverride != "" {
		activeContext = env.ContextOverride
	} else {
		// Load from persisted state
		activeContext, err = LoadContextState(env)
		if err != nil {
			return fmt.Errorf("failed to load context state: %w", err)
		}
	}

	// Set the context in ViceEnv
	env.Context = activeContext
	env.ContextData = filepath.Join(env.DataDir, activeContext)

	// Ensure context directory exists
	if err := env.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to ensure context directories: %w", err)
	}

	return nil
}