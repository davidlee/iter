package cmd

import (
	"github.com/spf13/cobra"
)

// goalCmd represents the goal command
var goalCmd = &cobra.Command{
	Use:   "goal",
	Short: "Manage goal definitions",
	Long: `Manage your habit tracking goals through interactive configuration. Supports
creating and editing diverse goal types: simple boolean goals, elastic goals with
achievement tiers (mini/midi/maxi), informational goals for data collection, and 
checklist-based goals. Configure automatic success criteria or manual evaluation.

Examples:
  vice goal add     # Add a new goal with guided prompts (all types supported)
  vice goal list    # List all existing goals with their configuration
  vice goal edit    # Edit an existing goal interactively
  vice goal remove  # Remove a goal`,
}

func init() {
	rootCmd.AddCommand(goalCmd)
}
