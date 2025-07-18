package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	initpkg "github.com/davidlee/vice/internal/init"
	"github.com/davidlee/vice/internal/ui/habitconfig"
)

// habitEditCmd represents the habit edit command
var habitEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit an existing habit",
	Long: `Select and modify an existing habit's configuration.
This command will present a list of existing habits to choose from,
then allow you to modify any of the habit's properties.

Examples:
  vice habit edit                     # Edit a habit
  vice habit edit --dry-run           # Preview edited YAML without saving
  vice habit edit --dry-run > habit.yml # Save preview to custom file
  vice --config-dir /tmp habit edit   # Use custom config directory`,
	RunE: runHabitEdit,
}

func init() {
	habitCmd.AddCommand(habitEditCmd)
	habitEditCmd.Flags().Bool("dry-run", false, "Preview edited YAML without saving to habits.yml")
}

func runHabitEdit(cmd *cobra.Command, _ []string) error {
	// Get the resolved environment
	env := GetViceEnv()

	// Check if dry-run flag is set
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return fmt.Errorf("failed to get dry-run flag: %w", err)
	}

	// Ensure context files exist, creating samples if missing
	initializer := initpkg.NewFileInitializer()
	if err := initializer.EnsureContextFiles(env); err != nil {
		return err
	}

	// Create habit configurator
	configurator := habitconfig.NewHabitConfigurator()

	if dryRun {
		// Dry-run mode: output YAML to stdout
		yamlOutput, err := configurator.EditHabitWithYAMLOutput(env.GetHabitsFile())
		if err != nil {
			return err
		}
		// Output YAML to stdout
		fmt.Print(yamlOutput)
		return nil
	}

	// Normal mode: save to file
	return configurator.EditHabit(env.GetHabitsFile())
}
