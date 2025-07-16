package cmd

import (
	"github.com/spf13/cobra"

	init_pkg "davidlee/vice/internal/init"
	"davidlee/vice/internal/ui"
)

// menuFlag indicates whether to launch the interactive menu interface
var menuFlag bool

// entryCmd represents the entry command
// AIDEV-NOTE: T018/5.3-help-update; comprehensive help text reflects full feature set (simple/elastic/informational/checklist habits)
var entryCmd = &cobra.Command{
	Use:   "entry",
	Short: "Record today's habit completion",
	Long: `Record today's habit data through interactive collection forms. Supports all habit types:
simple boolean tracking, elastic habits with achievement tiers, informational data collection,
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
	// Get the resolved environment
	env := GetViceEnv()

	// Ensure context files exist, creating samples if missing
	initializer := init_pkg.NewFileInitializer()
	if err := initializer.EnsureContextFiles(env); err != nil {
		return err
	}

	if menuFlag {
		return runEntryMenu(env)
	}

	// Create entry collector and run interactive UI
	collector := ui.NewEntryCollector(env.GetChecklistsFile())
	return collector.CollectTodayEntries(env.GetHabitsFile(), env.GetEntriesFile())
}
