package cmd

import (
	"github.com/spf13/cobra"

	initpkg "davidlee/iter/internal/init"
	"davidlee/iter/internal/ui/goalconfig"
)

// goalEditCmd represents the goal edit command
var goalEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit an existing goal",
	Long: `Select and modify an existing goal's configuration.
This command will present a list of existing goals to choose from,
then allow you to modify any of the goal's properties.

Examples:
  iter goal edit                     # Edit a goal
  iter --config-dir /tmp goal edit   # Use custom config directory`,
	RunE: runGoalEdit,
}

func init() {
	goalCmd.AddCommand(goalEditCmd)
}

func runGoalEdit(_ *cobra.Command, _ []string) error {
	// Get the resolved paths
	paths := GetPaths()

	// Ensure config files exist, creating samples if missing
	initializer := initpkg.NewFileInitializer()
	if err := initializer.EnsureConfigFiles(paths.GoalsFile, paths.EntriesFile); err != nil {
		return err
	}

	// Create goal configurator and run edit UI
	configurator := goalconfig.NewGoalConfigurator()
	return configurator.EditGoal(paths.GoalsFile)
}
