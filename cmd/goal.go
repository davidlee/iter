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
  iter goal add     # Add a new goal with guided prompts
  iter goal list    # List all existing goals
  iter goal edit    # Edit an existing goal
  iter goal remove  # Remove a goal`,
}

func init() {
	rootCmd.AddCommand(goalCmd)
}
