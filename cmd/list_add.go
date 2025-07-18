package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/davidlee/vice/internal/models"
	"github.com/davidlee/vice/internal/parser"
	"github.com/davidlee/vice/internal/ui/checklist"
)

// listAddCmd represents the list add command
var listAddCmd = &cobra.Command{
	Use:   "add [checklist_id]",
	Short: "Add a new checklist with interactive prompts",
	Long: `Add a new checklist through a simple multiline text interface.
You'll enter checklist items one per line, with headings prefixed by "# ".
If no checklist ID is provided, one will be generated from the title.

Examples:
  vice list add                       # Add a checklist (ID generated from title)
  vice list add morning_routine       # Add a checklist called "morning_routine"
  vice list add daily_review          # Add a checklist called "daily_review"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runListAdd,
}

func init() {
	listCmd.AddCommand(listAddCmd)
}

func runListAdd(_ *cobra.Command, args []string) error {
	var checklistID string
	var generateID bool

	if len(args) > 0 {
		// ID provided - validate it
		checklistID = args[0]
		if err := checklist.ValidateChecklistID(checklistID); err != nil {
			return fmt.Errorf("invalid checklist ID '%s': %w", checklistID, err)
		}
	} else {
		// No ID provided - will generate from title
		// AIDEV-NOTE: id-generation; consistent with habit ID logic (T007 Phase 3.1)
		generateID = true
	}

	// Get the resolved environment
	env := GetViceEnv()

	// Initialize checklist parser
	checklistParser := parser.NewChecklistParser()

	// Load existing checklists or create empty schema
	var schema *models.ChecklistSchema
	var err error

	if _, statErr := os.Stat(env.GetChecklistsFile()); os.IsNotExist(statErr) {
		// Create new schema if file doesn't exist
		schema = &models.ChecklistSchema{
			Version:     "1.0.0",
			CreatedDate: time.Now().Format("2006-01-02"),
			Checklists:  []models.Checklist{},
		}
	} else {
		// Load existing schema
		schema, err = checklistParser.LoadFromFile(env.GetChecklistsFile())
		if err != nil {
			return fmt.Errorf("failed to load checklists: %w", err)
		}
	}

	// Create editor configuration
	editorConfig := checklist.EditorConfig{
		ChecklistID: checklistID,
		IsEdit:      false,
		GenerateID:  generateID,
	}

	// Run the editor
	editor := checklist.NewEditor(editorConfig)
	newChecklist, err := editor.Run()
	if err != nil {
		return fmt.Errorf("editor error: %w", err)
	}

	// Check for ID conflicts (both provided and generated IDs)
	if _, err := checklistParser.GetChecklistByID(schema, newChecklist.ID); err == nil {
		return fmt.Errorf("checklist with ID '%s' already exists", newChecklist.ID)
	}

	// Add the checklist to the schema
	if err := checklistParser.AddChecklist(schema, newChecklist); err != nil {
		return fmt.Errorf("failed to add checklist: %w", err)
	}

	// Save the updated schema
	if err := checklistParser.SaveToFile(schema, env.GetChecklistsFile()); err != nil {
		return fmt.Errorf("failed to save checklists: %w", err)
	}

	fmt.Printf("✓ Created checklist '%s' with %d items\n", newChecklist.ID, len(newChecklist.Items))
	return nil
}
