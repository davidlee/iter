// Package config provides TOML configuration parsing for the vice application.
package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the structure of config.toml file.
// AIDEV-NOTE: toml-config-structure; defines app settings (not user data)
// AIDEV-NOTE: T028-toml-config; separation of concerns - config.toml for app settings, YAML for user data
type Config struct {
	Core CoreConfig `toml:"core"`
}

// CoreConfig represents the [core] section of config.toml.
type CoreConfig struct {
	Contexts []string `toml:"contexts"`
}

// DefaultConfig returns the default configuration values.
func DefaultConfig() *Config {
	return &Config{
		Core: CoreConfig{
			Contexts: []string{"personal", "work"},
		},
	}
}

// LoadConfig loads configuration from config.toml file.
// If the file doesn't exist, returns default configuration.
// If the file exists but has parsing errors, returns an error.
func LoadConfig(configPath string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File doesn't exist, return default config
		return DefaultConfig(), nil
	}

	// Read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse TOML
	config := &Config{}
	if err := toml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse TOML config %s: %w", configPath, err)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration in %s: %w", configPath, err)
	}

	return config, nil
}

// SaveConfig saves configuration to config.toml file.
func SaveConfig(configPath string, config *Config) error {
	// Validate before saving
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Marshal to TOML
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to TOML: %w", err)
	}

	// Write to file with appropriate permissions
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", configPath, err)
	}

	return nil
}

// validateConfig validates the configuration for consistency and correctness.
func validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate contexts
	if len(config.Core.Contexts) == 0 {
		return fmt.Errorf("at least one context must be defined in [core] contexts")
	}

	// Check for duplicate contexts
	seen := make(map[string]bool)
	for _, context := range config.Core.Contexts {
		if context == "" {
			return fmt.Errorf("context names cannot be empty")
		}
		if seen[context] {
			return fmt.Errorf("duplicate context name: %s", context)
		}
		seen[context] = true
	}

	return nil
}

// LoadViceEnvConfig loads ViceEnv with configuration from config.toml.
// This replaces the stub default loading in ViceEnv with actual TOML parsing.
// AIDEV-NOTE: T028-config-integration; bridges TOML configuration with ViceEnv runtime state
func LoadViceEnvConfig(env *ViceEnv) error {
	configPath := env.GetConfigTomlPath()

	config, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Update ViceEnv with loaded configuration
	env.Contexts = config.Core.Contexts

	// If current context is not in the loaded contexts, use first context as default
	contextValid := false
	for _, ctx := range env.Contexts {
		if env.Context == ctx {
			contextValid = true
			break
		}
	}

	// Reset to first context if current context is invalid and no override is set
	if !contextValid && env.ContextOverride == "" {
		env.Context = env.Contexts[0]
		env.ContextData = env.DataDir + "/" + env.Context
	}

	return nil
}

// EnsureConfigToml creates a config.toml file with default settings if it doesn't exist.
func EnsureConfigToml(configPath string) error {
	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		// File exists, nothing to do
		return nil
	}

	// Create default config and save it
	defaultConfig := DefaultConfig()
	if err := SaveConfig(configPath, defaultConfig); err != nil {
		return fmt.Errorf("failed to create default config.toml: %w", err)
	}

	return nil
}