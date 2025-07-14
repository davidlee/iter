package cmd

import (
	"github.com/spf13/cobra"

	"davidlee/vice/internal/ui"
)

// todoCmd represents the todo command
var todoCmd = &cobra.Command{
	Use:   "todo",
	Short: "Display today's habit status dashboard",
	Long: `Display a table showing today's habit completion status.

Shows each goal with its current status:
  ✓ Completed
  ○ Pending  
  ⤫ Skipped

Examples:
  vice todo                    # Show today's status table (bubbles)
  vice todo --ascii            # Show plain ASCII table
  vice todo -m                 # Output markdown todo list
  vice --config-dir /tmp todo  # Use custom config directory`,
	RunE: runTodo,
}

var (
	markdownOutput bool
	asciiOutput    bool
)

func init() {
	todoCmd.Flags().BoolVarP(&markdownOutput, "markdown", "m", false, "Output as markdown todo list")
	todoCmd.Flags().BoolVar(&asciiOutput, "ascii", false, "Output as plain ASCII table")
	rootCmd.AddCommand(todoCmd)
}

func runTodo(_ *cobra.Command, _ []string) error {
	// Get the resolved paths
	paths := GetPaths()

	// Create todo dashboard
	dashboard := ui.NewTodoDashboard(paths)

	// Display in requested format
	if markdownOutput {
		return dashboard.DisplayMarkdown()
	}
	if asciiOutput {
		return dashboard.DisplayASCII()
	}
	return dashboard.Display()
}
