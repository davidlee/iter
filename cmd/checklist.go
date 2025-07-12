package cmd

import (
	"github.com/spf13/cobra"
	// init_pkg "davidlee/iter/internal/init"
	"davidlee/iter/internal/ui"
)

// show a checklist
var checklistCmd = &cobra.Command{
	Use:   "checklist",
	Short: "Display a checklist",
	Long:  `Prototype checklist w bubbletea (UI only) `,
	RunE:  runChecklist,
}

func init() {
	rootCmd.AddCommand(checklistCmd)
}

func runChecklist(_ *cobra.Command, _ []string) error {
	ui.NewChecklistScreen()
	return nil
}
