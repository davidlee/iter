package cmd

import (
	"fmt"

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
  iter goal add --dry-run            # Preview YAML without saving
  iter goal add --dry-run > goal.yml # Save preview to custom file
  iter --config-dir /tmp goal add    # Use custom config directory`,
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
	configurator := goalconfig.NewGoalConfigurator()

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
