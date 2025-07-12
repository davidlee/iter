package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"davidlee/iter/internal/models"
	"davidlee/iter/internal/parser"
	"davidlee/iter/internal/ui/checklist"
)

// listEntryCmd represents the list entry command (with optional ID)
var listEntryCmd = &cobra.Command{
	Use:   "entry [checklist_id]",
	Short: "Complete a checklist",
	Long: `Enter checklist completion mode. If no checklist ID is provided,
you'll be shown a menu to select from available checklists.
Navigate with arrow keys or a/e, toggle items with space or enter, quit with q.

Examples:
  iter list entry                     # Select from a menu of checklists
  iter list entry morning_routine     # Complete the "morning_routine" checklist
  iter list entry daily_review        # Complete the "daily_review" checklist`,
	Args: cobra.MaximumNArgs(1),
	RunE: runListEntry,
}

func init() {
	listCmd.AddCommand(listEntryCmd)
}

func runListEntry(_ *cobra.Command, args []string) error {
	// Get the resolved paths
	paths := GetPaths()

	// Initialize parsers
	checklistParser := parser.NewChecklistParser()
	entriesParser := parser.NewChecklistEntriesParser()

	// Load existing checklists
	schema, err := checklistParser.LoadFromFile(paths.ChecklistsFile)
	if err != nil {
		return fmt.Errorf("failed to load checklists: %w", err)
	}

	var targetChecklist *models.Checklist

	if len(args) == 0 {
		// No ID provided - show selection menu
		if len(schema.Checklists) == 0 {
			return fmt.Errorf("no checklists available. Create one with 'iter list add <id>'")
		}

		selector := checklist.NewSelector(schema.Checklists)
		targetChecklist, err = selector.SelectChecklist()
		if err != nil {
			return fmt.Errorf("checklist selection error: %w", err)
		}
	} else {
		// ID provided - find the specific checklist
		checklistID := args[0]
		targetChecklist, err = checklistParser.GetChecklistByID(schema, checklistID)
		if err != nil {
			return fmt.Errorf("checklist '%s' not found: %w", checklistID, err)
		}
	}

	// Load or create checklist entries schema
	entriesSchema, err := entriesParser.EnsureSchemaExists(paths.ChecklistEntriesFile)
	if err != nil {
		return fmt.Errorf("failed to load checklist entries: %w", err)
	}

	// Get today's date and check for existing entry
	today := entriesParser.GetTodaysDate()
	existingEntry := entriesParser.GetChecklistEntryForDate(entriesSchema, today, targetChecklist.ID)

	var completion *models.ChecklistCompletion

	if existingEntry != nil {
		// Restore previous state for today's checklist
		// AIDEV-NOTE: state-restoration; same-day resume prevents data loss
		fmt.Printf("📋 Resuming checklist '%s' for %s (previous completion restored)\n", targetChecklist.ID, today)

		// Convert ChecklistEntry to ChecklistCompletion for UI compatibility
		previousCompletion := &models.ChecklistCompletion{
			ChecklistID:     existingEntry.ChecklistID,
			CompletedItems:  existingEntry.CompletedItems,
			CompletionTime:  existingEntry.CompletionTime,
			PartialComplete: existingEntry.PartialComplete,
		}

		completion, err = checklist.RunChecklistCompletionWithState(targetChecklist, previousCompletion)
		if err != nil {
			return fmt.Errorf("checklist completion error: %w", err)
		}
	} else {
		// Fresh checklist for today
		fmt.Printf("📋 Starting checklist '%s' for %s\n", targetChecklist.ID, today)

		completion, err = checklist.RunChecklistCompletion(targetChecklist)
		if err != nil {
			return fmt.Errorf("checklist completion error: %w", err)
		}
	}

	// Set completion time
	completion.CompletionTime = time.Now().Format(time.RFC3339)

	// Save completion state to checklist_entries.yml
	// AIDEV-NOTE: persistent-state; daily entries preserved across sessions
	entry := models.ChecklistEntry{
		ChecklistID:     completion.ChecklistID,
		CompletedItems:  completion.CompletedItems,
		CompletionTime:  completion.CompletionTime,
		PartialComplete: completion.PartialComplete,
	}

	if err := entriesParser.SaveChecklistEntryForDate(entriesSchema, today, targetChecklist.ID, entry); err != nil {
		return fmt.Errorf("failed to save checklist entry: %w", err)
	}

	if err := entriesParser.SaveToFile(entriesSchema, paths.ChecklistEntriesFile); err != nil {
		return fmt.Errorf("failed to save checklist entries file: %w", err)
	}

	// Display completion summary
	completedCount := 0
	for _, completed := range completion.CompletedItems {
		if completed {
			completedCount++
		}
	}

	totalItems := targetChecklist.GetTotalItemCount()

	fmt.Printf("\n✓ Checklist '%s' completed: %d/%d items\n", targetChecklist.ID, completedCount, totalItems)

	if completedCount == totalItems {
		fmt.Println("🎉 All items completed!")
	} else if completedCount > 0 {
		fmt.Printf("📝 Partial completion (%d%% done)\n", (completedCount*100)/totalItems)
	} else {
		fmt.Println("📋 No items completed")
	}

	fmt.Printf("💾 Completion state saved for %s\n", today)

	return nil
}
