// Package cmd provides the CLI commands for the vice application.
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"davidlee/vice/internal/config"
	"davidlee/vice/internal/debug"
	init_pkg "davidlee/vice/internal/init"
	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
	"davidlee/vice/internal/storage"
	"davidlee/vice/internal/ui"
	"davidlee/vice/internal/ui/entrymenu"
)

var (
	// Directory override flags
	configDir string // custom config directory path from CLI flag
	dataDir   string // custom data directory path from CLI flag
	cacheDir  string // custom cache directory path from CLI flag
	stateDir  string // custom state directory path from CLI flag

	// Context override flag
	contextFlag string // transient context override from CLI flag

	// Other flags
	debugMode bool // enables debug logging to file

	// viceEnv holds the resolved configuration environment
	viceEnv *config.ViceEnv
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vice",
	Short: "A CLI habit tracker",
	Long: `vice is a flexible command-line habit tracker supporting diverse habit types:
simple boolean habits, elastic habits with three achievement tiers (mini/midi/maxi), 
informational habits for data collection, and checklist-based habits. Features automatic
success criteria evaluation and stores data in local YAML files for portability.

Examples:
  vice                # Launch interactive entry menu (default)
  vice entry          # Record today's habit completion
  vice habit add      # Add new habits (simple/elastic/informational/checklist)
  vice todo           # View today's completion status dashboard
  
  # Directory overrides:
  vice --config-dir /path/to/config entry  # Use custom config directory
  vice --data-dir /path/to/data todo       # Use custom data directory
  
  # Context switching (transient):
  vice --context work entry               # Use work context for this command
  VICE_CONTEXT=work vice todo             # Use work context via environment variable
  
  # Context management (persistent):
  vice context list                       # Show all available contexts
  vice context switch work                # Switch to work context (persists)`,
	PersistentPreRunE: initializeViceEnv, // AIDEV-NOTE: Initializes interactive env - test cmd.Args() not cmd.Execute() to prevent hanging
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
		// Exit with error code after defer functions complete
		defer os.Exit(1)
	}
}

func init() {
	// Add persistent flags that apply to all commands

	// XDG directory overrides
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "",
		"custom config directory (default: XDG_CONFIG_HOME/vice or ~/.config/vice)")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "",
		"custom data directory (default: XDG_DATA_HOME/vice or ~/.local/share/vice)")
	rootCmd.PersistentFlags().StringVar(&cacheDir, "cache-dir", "",
		"custom cache directory (default: XDG_CACHE_HOME/vice or ~/.cache/vice)")
	rootCmd.PersistentFlags().StringVar(&stateDir, "state-dir", "",
		"custom state directory (default: XDG_STATE_HOME/vice or ~/.local/state/vice)")

	// Context override (always transient)
	rootCmd.PersistentFlags().StringVar(&contextFlag, "context", "",
		"use specific context for this command (transient, does not persist)")

	// Other flags
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false,
		"enable debug logging to file (creates vice-debug.log in config directory)")
}

// initializeViceEnv resolves the configuration environment based on CLI flags or defaults.
// This runs before any command execution to ensure environment is available.
// AIDEV-NOTE: T028-cmd-integration; replaced config.Paths with ViceEnv for context support
func initializeViceEnv(_ *cobra.Command, _ []string) error {
	var err error

	// Initialize ViceEnv with CLI flag overrides
	overrides := config.DirectoryOverrides{
		ConfigDir: configDir,
		DataDir:   dataDir,
		StateDir:  stateDir,
		CacheDir:  cacheDir,
		Context:   contextFlag,
	}
	viceEnv, err = config.GetViceEnvWithOverrides(overrides)
	if err != nil {
		return fmt.Errorf("failed to initialize ViceEnv: %w", err)
	}

	// Initialize debug logging if requested
	if debugMode {
		if err := debug.GetInstance().Initialize(viceEnv.ConfigDir); err != nil {
			return fmt.Errorf("failed to initialize debug logging: %w", err)
		}
		debug.General("Debug mode enabled via --debug flag")
	}

	return nil
}

// GetViceEnv returns the resolved configuration environment.
// This should be called after cobra command execution has started.
// AIDEV-NOTE: T028-cmd-integration; replaced GetPaths() with GetViceEnv() for context support
func GetViceEnv() *config.ViceEnv {
	return viceEnv
}

// GetPaths returns legacy config.Paths for backward compatibility during transition.
// AIDEV-NOTE: T028-transition; temporary compatibility function for remaining cmd files
func GetPaths() *config.Paths {
	if viceEnv == nil {
		return nil
	}
	return &config.Paths{
		ConfigDir:            viceEnv.ConfigDir,
		HabitsFile:           viceEnv.GetHabitsFile(),
		EntriesFile:          viceEnv.GetEntriesFile(),
		ChecklistsFile:       viceEnv.GetChecklistsFile(),
		ChecklistEntriesFile: viceEnv.GetChecklistEntriesFile(),
	}
}

// runDefaultCommand handles the default behavior when 'vice' is called without arguments.
// AIDEV-NOTE: T018/4.2-default-command; launches entry menu as default behavior for streamlined UX
// AIDEV-NOTE: fang-integration; uses Charmbracelet Fang for enhanced CLI styling (automatic --version, styled help)
func runDefaultCommand(_ *cobra.Command, _ []string) error {
	// Get the resolved environment
	env := GetViceEnv()

	// Ensure context files exist, creating samples if missing
	initializer := init_pkg.NewFileInitializer()
	if err := initializer.EnsureContextFiles(env); err != nil {
		return err
	}

	// Launch entry menu as default behavior
	return runEntryMenu(env)
}

// runEntryMenu launches the interactive entry menu interface.
// AIDEV-NOTE: entry-menu-integration; T018 command integration for --menu flag
func runEntryMenu(env *config.ViceEnv) error {
	// Load habits
	habitParser := parser.NewHabitParser()
	schema, err := habitParser.LoadFromFile(env.GetHabitsFile())
	if err != nil {
		return fmt.Errorf("failed to load habits: %w", err)
	}

	if len(schema.Habits) == 0 {
		return fmt.Errorf("no habits found in %s", env.GetHabitsFile())
	}

	// Load existing entries for today
	entryStorage := storage.NewEntryStorage()
	entries, err := loadTodayEntries(entryStorage, env.GetEntriesFile())
	if err != nil {
		return fmt.Errorf("failed to load existing entries: %w", err)
	}

	// AIDEV-NOTE: T018/3.1-menu-launch; EntryCollector setup for menu integration
	// Create and initialize entry collector for menu usage
	collector := ui.NewEntryCollector(env.GetChecklistsFile())
	// CRITICAL: InitializeForMenu() must be called to convert HabitEntry format to collector format
	collector.InitializeForMenu(schema.Habits, entries)

	// AIDEV-NOTE: T018/3.2-auto-save; pass entriesFile path for automatic persistence
	// Create and run entry menu with complete integration: collector + auto-save + return behavior
	model := entrymenu.NewEntryMenuModel(schema.Habits, entries, collector, env.GetEntriesFile())

	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()

	return err
}

// loadTodayEntries loads existing entries for today's date.
func loadTodayEntries(entryStorage *storage.EntryStorage, entriesFile string) (map[string]models.HabitEntry, error) {
	// Load entry log
	entryLog, err := entryStorage.LoadFromFile(entriesFile)
	if err != nil {
		// If file doesn't exist, return empty entries
		if os.IsNotExist(err) {
			return make(map[string]models.HabitEntry), nil
		}
		return nil, err
	}

	// Find today's entries
	today := time.Now().Format("2006-01-02")
	for _, dayEntry := range entryLog.Entries {
		if dayEntry.Date == today {
			// Convert to map for easy lookup
			entriesMap := make(map[string]models.HabitEntry)
			for _, habitEntry := range dayEntry.Habits {
				entriesMap[habitEntry.HabitID] = habitEntry
			}
			return entriesMap, nil
		}
	}

	// No entries for today
	return make(map[string]models.HabitEntry), nil
}
