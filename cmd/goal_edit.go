package cmd

import (
	"fmt"

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
  iter goal edit --dry-run           # Preview edited YAML without saving
  iter goal edit --dry-run > goal.yml # Save preview to custom file
  iter --config-dir /tmp goal edit   # Use custom config directory`,
	RunE: runGoalEdit,
}

func init() {
	goalCmd.AddCommand(goalEditCmd)
	goalEditCmd.Flags().Bool("dry-run", false, "Preview edited YAML without saving to goals.yml")
}

func runGoalEdit(cmd *cobra.Command, _ []string) error {
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
		yamlOutput, err := configurator.EditGoalWithYAMLOutput(paths.GoalsFile)
		if err != nil {
			return err
		}
		// Output YAML to stdout
		fmt.Print(yamlOutput)
		return nil
	}

	// Normal mode: save to file
	return configurator.EditGoal(paths.GoalsFile)
}
