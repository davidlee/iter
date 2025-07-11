package cmd

import (
	"github.com/spf13/cobra"

	initpkg "davidlee/iter/internal/init"
	"davidlee/iter/internal/ui/goalconfig"
)

// goalRemoveCmd represents the goal remove command
var goalRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an existing goal",
	Long: `Select and remove an existing goal from your configuration.
This command will present a list of existing goals to choose from,
show the goal details, and ask for confirmation before removal.

Examples:
  iter goal remove                     # Remove a goal
  iter --config-dir /tmp goal remove   # Use custom config directory`,
	RunE: runGoalRemove,
}

func init() {
	goalCmd.AddCommand(goalRemoveCmd)
}

func runGoalRemove(_ *cobra.Command, _ []string) error {
	// Get the resolved paths
	paths := GetPaths()

	// Ensure config files exist, creating samples if missing
	initializer := initpkg.NewFileInitializer()
	if err := initializer.EnsureConfigFiles(paths.GoalsFile, paths.EntriesFile); err != nil {
		return err
	}

	// Create goal configurator and run remove UI
	configurator := goalconfig.NewGoalConfigurator()
	return configurator.RemoveGoal(paths.GoalsFile)
}
