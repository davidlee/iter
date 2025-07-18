package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	initpkg "github.com/davidlee/vice/internal/init"
	"github.com/davidlee/vice/internal/ui/habitconfig"
)

// habitAddCmd represents the habit add command
// AIDEV-NOTE: T018/5.3-help-update; detailed help explains all habit types and automatic criteria configuration options
var habitAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new habit with interactive prompts",
	Long: `Add a new habit through an interactive configuration interface. Supports all habit types:
simple boolean habits, elastic habits with mini/midi/maxi achievement tiers, informational 
habits for data collection, and checklist-based habits. Configure automatic success criteria
with numeric/time/boolean conditions or choose manual evaluation.

Examples:
  vice habit add                      # Add a new habit (all types supported)
  vice habit add --dry-run            # Preview YAML without saving
  vice habit add --dry-run > habit.yml # Save preview to custom file
  vice --config-dir /tmp habit add    # Use custom config directory`,
	RunE: runHabitAdd,
}

func init() {
	habitCmd.AddCommand(habitAddCmd)
	habitAddCmd.Flags().Bool("dry-run", false, "Preview generated YAML without saving to habits.yml")
}

func runHabitAdd(cmd *cobra.Command, _ []string) error {
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
	configurator := habitconfig.NewHabitConfigurator().WithChecklistsFile(env.GetChecklistsFile())

	if dryRun {
		// Dry-run mode: output YAML to stdout
		yamlOutput, err := configurator.AddHabitWithYAMLOutput(env.GetHabitsFile())
		if err != nil {
			return err
		}
		// Output YAML to stdout
		fmt.Print(yamlOutput)
		return nil
	}

	// Normal mode: save to file
	return configurator.AddHabit(env.GetHabitsFile())
}
