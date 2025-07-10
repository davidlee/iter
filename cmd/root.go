// Package cmd provides the CLI commands for the iter application.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"davidlee/iter/internal/config"
)

var (
	// configDir holds the custom config directory path from CLI flag
	configDir string
	// paths holds the resolved configuration paths
	paths *config.Paths
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iter",
	Short: "A CLI habit tracker",
	Long: `iter is a command-line habit tracker. It supports simple boolean goals,
storing your data in local YAML files for easy version control and portability.

Examples:
  iter entry          # Record today's habit completion
  iter goals          # Display current goals
  iter --config-dir /path/to/config entry  # Use custom config directory`,
	PersistentPreRunE: initializePaths,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add persistent flags that apply to all commands
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "",
		"custom config directory (default: XDG_CONFIG_HOME/iter or ~/.config/iter)")
}

// initializePaths resolves the configuration paths based on CLI flags or defaults.
// This runs before any command execution to ensure paths are available.
func initializePaths(_ *cobra.Command, _ []string) error {
	var err error

	if configDir != "" {
		// Use custom config directory from CLI flag
		paths = config.GetPathsWithConfigDir(configDir)
	} else {
		// Use default XDG-compliant paths
		paths, err = config.GetDefaultPaths()
		if err != nil {
			return fmt.Errorf("failed to resolve default config paths: %w", err)
		}
	}

	// Ensure the config directory exists
	if err := paths.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", paths.ConfigDir, err)
	}

	return nil
}

// GetPaths returns the resolved configuration paths.
// This should be called after cobra command execution has started.
func GetPaths() *config.Paths {
	return paths
}
