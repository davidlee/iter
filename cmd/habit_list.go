package cmd

import (
	"github.com/spf13/cobra"

	initpkg "davidlee/vice/internal/init"
	"davidlee/vice/internal/ui/habitconfig"
)

// habitListCmd represents the habit list command
var habitListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all existing habits",
	Long: `Display all existing habits in a human-readable format.
Shows habit type, field type, scoring configuration, and other properties
for each defined habit.

Examples:
  vice habit list                     # List all habits
  vice --config-dir /tmp habit list   # Use custom config directory`,
	RunE: runHabitList,
}

func init() {
	habitCmd.AddCommand(habitListCmd)
}

func runHabitList(_ *cobra.Command, _ []string) error {
	// Get the resolved environment
	env := GetViceEnv()

	// Ensure context files exist, creating samples if missing
	initializer := initpkg.NewFileInitializer()
	if err := initializer.EnsureContextFiles(env); err != nil {
		return err
	}

	// Create habit configurator and run list display
	configurator := habitconfig.NewHabitConfigurator()
	return configurator.ListHabits(env.GetHabitsFile())
}
