package cmd

import (
	"github.com/spf13/cobra"

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

	// Create entry collector and run interactive UI
	collector := ui.NewEntryCollector()
	return collector.CollectTodayEntries(paths.GoalsFile, paths.EntriesFile)
}
