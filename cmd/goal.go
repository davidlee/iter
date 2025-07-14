package cmd

import (
	"github.com/spf13/cobra"
)

// goalCmd represents the goal command
var goalCmd = &cobra.Command{
	Use:   "goal",
	Short: "Manage goal definitions",
	Long: `Manage your habit tracking goals through interactive configuration.
This command provides subcommands to add, list, edit, and remove goals
without manually editing YAML files.

Examples:
  vice goal add     # Add a new goal with guided prompts
  vice goal list    # List all existing goals
  vice goal edit    # Edit an existing goal
  vice goal remove  # Remove a goal`,
}

func init() {
	rootCmd.AddCommand(goalCmd)
}
