package cmd

import (
	"github.com/spf13/cobra"

	initpkg "github.com/davidlee/vice/internal/init"
	"github.com/davidlee/vice/internal/ui/habitconfig"
)

// habitRemoveCmd represents the habit remove command
var habitRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an existing habit",
	Long: `Select and remove an existing habit from your configuration.
This command will present a list of existing habits to choose from,
show the habit details, and ask for confirmation before removal.

Examples:
  vice habit remove                     # Remove a habit
  vice --config-dir /tmp habit remove   # Use custom config directory`,
	RunE: runHabitRemove,
}

func init() {
	habitCmd.AddCommand(habitRemoveCmd)
}

func runHabitRemove(_ *cobra.Command, _ []string) error {
	// Get the resolved environment
	env := GetViceEnv()

	// Ensure context files exist, creating samples if missing
	initializer := initpkg.NewFileInitializer()
	if err := initializer.EnsureContextFiles(env); err != nil {
		return err
	}

	// Create habit configurator and run remove UI
	configurator := habitconfig.NewHabitConfigurator()
	return configurator.RemoveHabit(env.GetHabitsFile())
}
