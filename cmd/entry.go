package cmd

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"davidlee/vice/internal/config"
	init_pkg "davidlee/vice/internal/init"
	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
	"davidlee/vice/internal/storage"
	"davidlee/vice/internal/ui"
	"davidlee/vice/internal/ui/entrymenu"
)

// menuFlag indicates whether to launch the interactive menu interface
var menuFlag bool

// entryCmd represents the entry command
// AIDEV-NOTE: T018/5.3-help-update; comprehensive help text reflects full feature set (simple/elastic/informational/checklist goals)
var entryCmd = &cobra.Command{
	Use:   "entry",
	Short: "Record today's habit completion",
	Long: `Record today's habit data through interactive collection forms. Supports all goal types:
simple boolean tracking, elastic goals with achievement tiers, informational data collection,
and checklist completion. Features automatic success evaluation based on configured criteria.
Your entries are stored in entries.yml for progress tracking and analysis.

Examples:
  vice entry                    # Record today's habits (sequential form)
  vice entry --menu             # Launch interactive menu interface (recommended)
  vice --config-dir /tmp entry  # Use custom config directory`,
	RunE: runEntry,
}

func init() {
	rootCmd.AddCommand(entryCmd)
	entryCmd.Flags().BoolVar(&menuFlag, "menu", false, "Launch interactive menu interface")
}

func runEntry(_ *cobra.Command, _ []string) error {
	// Get the resolved paths
	paths := GetPaths()

	// Ensure config files exist, creating samples if missing
	initializer := init_pkg.NewFileInitializer()
	if err := initializer.EnsureConfigFiles(paths.GoalsFile, paths.EntriesFile); err != nil {
		return err
	}

	if menuFlag {
		return runEntryMenu(paths)
	}

	// Create entry collector and run interactive UI
	collector := ui.NewEntryCollector(paths.ChecklistsFile)
	return collector.CollectTodayEntries(paths.GoalsFile, paths.EntriesFile)
}

// runEntryMenu launches the interactive entry menu interface.
// AIDEV-NOTE: entry-menu-integration; T018 command integration for --menu flag
func runEntryMenu(paths *config.Paths) error {
	// Load goals
	goalParser := parser.NewGoalParser()
	schema, err := goalParser.LoadFromFile(paths.GoalsFile)
	if err != nil {
		return fmt.Errorf("failed to load goals: %w", err)
	}

	if len(schema.Goals) == 0 {
		return fmt.Errorf("no goals found in %s", paths.GoalsFile)
	}

	// Load existing entries for today
	entryStorage := storage.NewEntryStorage()
	entries, err := loadTodayEntries(entryStorage, paths.EntriesFile)
	if err != nil {
		return fmt.Errorf("failed to load existing entries: %w", err)
	}

	// AIDEV-NOTE: T018/3.1-menu-launch; EntryCollector setup for menu integration
	// Create and initialize entry collector for menu usage
	collector := ui.NewEntryCollector(paths.ChecklistsFile)
	// CRITICAL: InitializeForMenu() must be called to convert GoalEntry format to collector format
	collector.InitializeForMenu(schema.Goals, entries)

	// AIDEV-NOTE: T018/3.2-auto-save; pass entriesFile path for automatic persistence
	// Create and run entry menu with complete integration: collector + auto-save + return behavior
	model := entrymenu.NewEntryMenuModel(schema.Goals, entries, collector, paths.EntriesFile)

	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()

	return err
}

// loadTodayEntries loads existing entries for today's date.
func loadTodayEntries(entryStorage *storage.EntryStorage, entriesFile string) (map[string]models.GoalEntry, error) {
	// Load entry log
	entryLog, err := entryStorage.LoadFromFile(entriesFile)
	if err != nil {
		// If file doesn't exist, return empty entries
		if os.IsNotExist(err) {
			return make(map[string]models.GoalEntry), nil
		}
		return nil, err
	}

	// Find today's entries
	today := time.Now().Format("2006-01-02")
	for _, dayEntry := range entryLog.Entries {
		if dayEntry.Date == today {
			// Convert to map for easy lookup
			entriesMap := make(map[string]models.GoalEntry)
			for _, goalEntry := range dayEntry.Goals {
				entriesMap[goalEntry.GoalID] = goalEntry
			}
			return entriesMap, nil
		}
	}

	// No entries for today
	return make(map[string]models.GoalEntry), nil
}
