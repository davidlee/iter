package cmd

import (
	"github.com/spf13/cobra"
)

// habitCmd represents the habit command
// AIDEV-NOTE: T018/5.3-help-update; enhanced help documents all habit types and automatic criteria features
var habitCmd = &cobra.Command{
	Use:   "habit",
	Short: "Manage habit definitions",
	Long: `Manage your habit tracking habits through interactive configuration. Supports
creating and editing diverse habit types: simple boolean habits, elastic habits with
achievement tiers (mini/midi/maxi), informational habits for data collection, and 
checklist-based habits. Configure automatic success criteria or manual evaluation.

Examples:
  vice habit add     # Add a new habit with guided prompts (all types supported)
  vice habit list    # List all existing habits with their configuration
  vice habit edit    # Edit an existing habit interactively
  vice habit remove  # Remove a habit`,
}

func init() {
	rootCmd.AddCommand(habitCmd)
}
