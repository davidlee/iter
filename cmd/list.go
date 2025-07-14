package cmd

import (
	"github.com/spf13/cobra"
)

// listCmd represents the list command for checklist management
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Manage checklist definitions",
	Long: `Manage your checklists through interactive configuration.
This command provides subcommands to add, edit, and complete checklists
stored in checklists.yml.

Examples:
  vice list add morning_routine    # Add a new checklist with guided prompts
  vice list edit morning_routine   # Edit an existing checklist
  vice list entry                  # Select and complete a checklist
  vice list entry morning_routine  # Complete a specific checklist`,
}

func init() {
	rootCmd.AddCommand(listCmd)
}
