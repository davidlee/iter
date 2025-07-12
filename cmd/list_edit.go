package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"davidlee/iter/internal/parser"
	"davidlee/iter/internal/ui/checklist"
)

// listEditCmd represents the list edit command
var listEditCmd = &cobra.Command{
	Use:   "edit [checklist_id]",
	Short: "Edit an existing checklist",
	Long: `Edit an existing checklist through the same multiline text interface
used for creating checklists. The current content will be pre-loaded
for modification.

Examples:
  iter list edit morning_routine       # Edit the "morning_routine" checklist
  iter list edit daily_review          # Edit the "daily_review" checklist`,
	Args: cobra.ExactArgs(1),
	RunE: runListEdit,
}

func init() {
	listCmd.AddCommand(listEditCmd)
}

func runListEdit(_ *cobra.Command, args []string) error {
	checklistID := args[0]

	// Validate the checklist ID
	if err := checklist.ValidateChecklistID(checklistID); err != nil {
		return fmt.Errorf("invalid checklist ID '%s': %w", checklistID, err)
	}

	// Get the resolved paths
	paths := GetPaths()

	// Initialize checklist parser
	checklistParser := parser.NewChecklistParser()

	// Load existing checklists
	schema, err := checklistParser.LoadFromFile(paths.ChecklistsFile)
	if err != nil {
		return fmt.Errorf("failed to load checklists: %w", err)
	}

	// Find the existing checklist
	existingChecklist, err := checklistParser.GetChecklistByID(schema, checklistID)
	if err != nil {
		return fmt.Errorf("checklist '%s' not found: %w", checklistID, err)
	}

	// Create editor configuration with existing data
	editorConfig := checklist.EditorConfig{
		ChecklistID:   checklistID,
		Title:         existingChecklist.Title,
		Description:   existingChecklist.Description,
		ExistingItems: existingChecklist.Items,
		IsEdit:        true,
	}

	// Run the editor
	editor := checklist.NewEditor(editorConfig)
	updatedChecklist, err := editor.Run()
	if err != nil {
		return fmt.Errorf("editor error: %w", err)
	}

	// Preserve the original creation date
	updatedChecklist.CreatedDate = existingChecklist.CreatedDate
	updatedChecklist.ModifiedDate = time.Now().Format("2006-01-02")

	// Update the checklist in the schema
	if err := checklistParser.UpdateChecklist(schema, updatedChecklist); err != nil {
		return fmt.Errorf("failed to update checklist: %w", err)
	}

	// Save the updated schema
	if err := checklistParser.SaveToFile(schema, paths.ChecklistsFile); err != nil {
		return fmt.Errorf("failed to save checklists: %w", err)
	}

	fmt.Printf("✓ Updated checklist '%s' with %d items\n", checklistID, len(updatedChecklist.Items))
	return nil
}
