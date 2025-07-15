package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	initpkg "davidlee/vice/internal/init"
	"davidlee/vice/internal/ui/goalconfig"
)

// goalAddCmd represents the goal add command
var goalAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new goal with interactive prompts",
	Long: `Add a new goal through an interactive configuration interface. Supports all goal types:
simple boolean goals, elastic goals with mini/midi/maxi achievement tiers, informational 
goals for data collection, and checklist-based goals. Configure automatic success criteria
with numeric/time/boolean conditions or choose manual evaluation.

Examples:
  vice goal add                      # Add a new goal (all types supported)
  vice goal add --dry-run            # Preview YAML without saving
  vice goal add --dry-run > goal.yml # Save preview to custom file
  vice --config-dir /tmp goal add    # Use custom config directory`,
	RunE: runGoalAdd,
}

func init() {
	goalCmd.AddCommand(goalAddCmd)
	goalAddCmd.Flags().Bool("dry-run", false, "Preview generated YAML without saving to goals.yml")
}

func runGoalAdd(cmd *cobra.Command, _ []string) error {
	// Get the resolved paths
	paths := GetPaths()

	// Check if dry-run flag is set
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return fmt.Errorf("failed to get dry-run flag: %w", err)
	}

	// Ensure config files exist, creating samples if missing
	initializer := initpkg.NewFileInitializer()
	if err := initializer.EnsureConfigFiles(paths.GoalsFile, paths.EntriesFile); err != nil {
		return err
	}

	// Create goal configurator
	configurator := goalconfig.NewGoalConfigurator().WithChecklistsFile(paths.ChecklistsFile)

	if dryRun {
		// Dry-run mode: output YAML to stdout
		yamlOutput, err := configurator.AddGoalWithYAMLOutput(paths.GoalsFile)
		if err != nil {
			return err
		}
		// Output YAML to stdout
		fmt.Print(yamlOutput)
		return nil
	}

	// Normal mode: save to file
	return configurator.AddGoal(paths.GoalsFile)
}
