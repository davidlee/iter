package cmd

import (
	"github.com/spf13/cobra"

	initpkg "davidlee/iter/internal/init"
	"davidlee/iter/internal/ui/goalconfig"
)

// goalAddCmd represents the goal add command
var goalAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new goal with interactive prompts",
	Long: `Add a new goal through an interactive configuration interface.
This command will guide you through defining a goal's properties including
type, field type, scoring criteria, and other settings.

Examples:
  iter goal add                      # Add a new goal
  iter --config-dir /tmp goal add    # Use custom config directory`,
	RunE: runGoalAdd,
}

func init() {
	goalCmd.AddCommand(goalAddCmd)
}

func runGoalAdd(_ *cobra.Command, _ []string) error {
	// Get the resolved paths
	paths := GetPaths()

	// Ensure config files exist, creating samples if missing
	initializer := initpkg.NewFileInitializer()
	if err := initializer.EnsureConfigFiles(paths.GoalsFile, paths.EntriesFile); err != nil {
		return err
	}

	// Create goal configurator and run interactive UI
	configurator := goalconfig.NewGoalConfigurator()
	return configurator.AddGoal(paths.GoalsFile)
}
