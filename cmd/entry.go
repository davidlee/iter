package cmd

import (
	"github.com/spf13/cobra"

	init_pkg "davidlee/iter/internal/init"
	"davidlee/iter/internal/ui"
)

// entryCmd represents the entry command
var entryCmd = &cobra.Command{
	Use:   "entry",
	Short: "Record today's habit completion",
	Long: `Record today's habit completion by answering questions about your goals.
This command will present an interactive form where you can mark which habits
you completed today. Your entries are stored in entries.yml for tracking progress.

Examples:
  iter entry                    # Record today's habits
  iter --config-dir /tmp entry  # Use custom config directory`,
	RunE: runEntry,
}

func init() {
	rootCmd.AddCommand(entryCmd)
}

func runEntry(_ *cobra.Command, _ []string) error {
	// Get the resolved paths
	paths := GetPaths()

	// Ensure config files exist, creating samples if missing
	initializer := init_pkg.NewFileInitializer()
	if err := initializer.EnsureConfigFiles(paths.GoalsFile, paths.EntriesFile); err != nil {
		return err
	}

	// Create entry collector and run interactive UI
	collector := ui.NewEntryCollector(paths.ChecklistsFile)
	return collector.CollectTodayEntries(paths.GoalsFile, paths.EntriesFile)
}
