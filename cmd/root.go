// Package cmd provides the CLI commands for the vice application.
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"davidlee/vice/internal/config"
	"davidlee/vice/internal/debug"
	init_pkg "davidlee/vice/internal/init"
)

var (
	// configDir holds the custom config directory path from CLI flag
	configDir string
	// debugMode enables debug logging to file
	debugMode bool
	// paths holds the resolved configuration paths
	paths *config.Paths
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vice",
	Short: "A CLI habit tracker",
	Long: `vice is a flexible command-line habit tracker supporting diverse goal types:
simple boolean goals, elastic goals with three achievement tiers (mini/midi/maxi), 
informational goals for data collection, and checklist-based goals. Features automatic
success criteria evaluation and stores data in local YAML files for portability.

Examples:
  vice                # Launch interactive entry menu (default)
  vice entry          # Record today's habit completion
  vice goal add       # Add new goals (simple/elastic/informational/checklist)
  vice todo           # View today's completion status dashboard
  vice --config-dir /path/to/config entry  # Use custom config directory`,
	PersistentPreRunE: initializePaths,
	RunE:              runDefaultCommand,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// AIDEV-NOTE: T018/5.3-fang-integration; replaced cobra.Execute() with fang.Execute() for enhanced CLI styling
func Execute() {
	// Ensure debug logger is closed on exit
	defer func() {
		if err := debug.GetInstance().Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close debug logger: %v\n", err)
		}
	}()

	err := fang.Execute(context.Background(), rootCmd)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add persistent flags that apply to all commands
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "",
		"custom config directory (default: XDG_CONFIG_HOME/vice or ~/.config/vice)")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false,
		"enable debug logging to file (creates vice-debug.log in config directory)")
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

	// Initialize debug logging if requested
	if debugMode {
		if err := debug.GetInstance().Initialize(paths.ConfigDir); err != nil {
			return fmt.Errorf("failed to initialize debug logging: %w", err)
		}
		debug.General("Debug mode enabled via --debug flag")
	}

	return nil
}

// GetPaths returns the resolved configuration paths.
// This should be called after cobra command execution has started.
func GetPaths() *config.Paths {
	return paths
}

// runDefaultCommand handles the default behavior when 'vice' is called without arguments.
// AIDEV-NOTE: T018/4.2-default-command; launches entry menu as default behavior for streamlined UX
// AIDEV-NOTE: fang-integration; uses Charmbracelet Fang for enhanced CLI styling (automatic --version, styled help)
func runDefaultCommand(_ *cobra.Command, _ []string) error {
	// Get the resolved paths
	paths := GetPaths()

	// Ensure config files exist, creating samples if missing
	initializer := init_pkg.NewFileInitializer()
	if err := initializer.EnsureConfigFiles(paths.GoalsFile, paths.EntriesFile); err != nil {
		return err
	}

	// Launch entry menu as default behavior
	return runEntryMenu(paths)
}
