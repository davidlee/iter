package cmd

import (
	"github.com/spf13/cobra"

	initpkg "davidlee/vice/internal/init"
	"davidlee/vice/internal/ui/goalconfig"
)

// goalListCmd represents the goal list command
var goalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all existing goals",
	Long: `Display all existing goals in a human-readable format.
Shows goal type, field type, scoring configuration, and other properties
for each defined goal.

Examples:
  vice goal list                     # List all goals
  vice --config-dir /tmp goal list   # Use custom config directory`,
	RunE: runGoalList,
}

func init() {
	goalCmd.AddCommand(goalListCmd)
}

func runGoalList(_ *cobra.Command, _ []string) error {
	// Get the resolved paths
	paths := GetPaths()

	// Ensure config files exist, creating samples if missing
	initializer := initpkg.NewFileInitializer()
	if err := initializer.EnsureConfigFiles(paths.GoalsFile, paths.EntriesFile); err != nil {
		return err
	}

	// Create goal configurator and run list display
	configurator := goalconfig.NewGoalConfigurator()
	return configurator.ListGoals(paths.GoalsFile)
}
