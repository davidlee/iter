package goalconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/iter/internal/models"
)

func TestChecklistGoalCreator_LoadAvailableChecklists(t *testing.T) {
	// Create a temporary checklists.yml file
	tempDir := t.TempDir()
	checklistsFile := filepath.Join(tempDir, "checklists.yml")

	checklistsContent := `version: "1.0.0"
created_date: "2025-07-12"
checklists:
  - id: morning_routine
    title: Morning Routine
    description: Daily morning checklist
    items:
      - "# health"
      - "take vitamins"
      - "drink water"
    created_date: "2025-07-12"
    modified_date: "2025-07-12"
  - id: evening_routine
    title: Evening Routine
    description: End of day checklist
    items:
      - "review day"
      - "plan tomorrow"
    created_date: "2025-07-12"
    modified_date: "2025-07-12"
`

	err := os.WriteFile(checklistsFile, []byte(checklistsContent), 0o600)
	require.NoError(t, err)

	// Test loading available checklists
	creator := NewChecklistGoalCreator("Test Goal", "Test Description", models.ChecklistGoal, checklistsFile)

	// Verify no error occurred during initialization
	assert.NoError(t, creator.err)

	// Verify checklists were loaded
	assert.Len(t, creator.availableChecklists, 2)
	assert.Equal(t, "morning_routine", creator.availableChecklists[0].ID)
	assert.Equal(t, "Morning Routine", creator.availableChecklists[0].Title)
	assert.Equal(t, "evening_routine", creator.availableChecklists[1].ID)
	assert.Equal(t, "Evening Routine", creator.availableChecklists[1].Title)
}

func TestChecklistGoalCreator_NoChecklistsFile(t *testing.T) {
	// Test with non-existent checklists file
	creator := NewChecklistGoalCreator("Test Goal", "Test Description", models.ChecklistGoal, "/nonexistent/checklists.yml")

	// Should have error when no checklists are available (form creation sets this error)
	assert.Error(t, creator.err)
	assert.Contains(t, creator.err.Error(), "no checklists found")
	assert.Len(t, creator.availableChecklists, 0)
}

func TestChecklistGoalCreator_BuildGoalAutomaticScoring(t *testing.T) {
	// Create a test creator with mock data
	creator := &ChecklistGoalCreator{
		title:       "Morning Routine Goal",
		description: "Complete morning tasks",
		goalType:    models.ChecklistGoal,
		checklistID: "morning_routine",
		scoringType: models.AutomaticScoring,
		prompt:      "Did you complete your morning routine?",
	}

	// Build the result
	err := creator.buildResult()
	require.NoError(t, err)

	goal := creator.result
	require.NotNil(t, goal)

	// Verify goal properties
	assert.Equal(t, "Morning Routine Goal", goal.Title)
	assert.Equal(t, "Complete morning tasks", goal.Description)
	assert.Equal(t, models.ChecklistGoal, goal.GoalType)
	assert.Equal(t, models.ChecklistFieldType, goal.FieldType.Type)
	assert.Equal(t, "morning_routine", goal.FieldType.ChecklistID)
	assert.Equal(t, models.AutomaticScoring, goal.ScoringType)
	assert.Equal(t, "Did you complete your morning routine?", goal.Prompt)

	// Verify automatic scoring criteria
	require.NotNil(t, goal.Criteria)
	assert.Equal(t, "All checklist items completed", goal.Criteria.Description)
	require.NotNil(t, goal.Criteria.Condition)
	require.NotNil(t, goal.Criteria.Condition.ChecklistCompletion)
	assert.Equal(t, "all", goal.Criteria.Condition.ChecklistCompletion.RequiredItems)
}

func TestChecklistGoalCreator_BuildGoalManualScoring(t *testing.T) {
	// Create a test creator with manual scoring
	creator := &ChecklistGoalCreator{
		title:       "Evening Routine Goal",
		description: "Complete evening tasks",
		goalType:    models.ChecklistGoal,
		checklistID: "evening_routine",
		scoringType: models.ManualScoring,
		prompt:      "How well did you complete your evening routine?",
	}

	// Build the result
	err := creator.buildResult()
	require.NoError(t, err)

	goal := creator.result
	require.NotNil(t, goal)

	// Verify goal properties
	assert.Equal(t, "Evening Routine Goal", goal.Title)
	assert.Equal(t, "Complete evening tasks", goal.Description)
	assert.Equal(t, models.ChecklistGoal, goal.GoalType)
	assert.Equal(t, models.ChecklistFieldType, goal.FieldType.Type)
	assert.Equal(t, "evening_routine", goal.FieldType.ChecklistID)
	assert.Equal(t, models.ManualScoring, goal.ScoringType)
	assert.Equal(t, "How well did you complete your evening routine?", goal.Prompt)

	// Manual scoring should not have automatic criteria
	assert.Nil(t, goal.Criteria)
}

func TestChecklistGoalCreator_BuildGoalWithEmptyPrompt(t *testing.T) {
	// Create a test creator with empty prompt
	creator := &ChecklistGoalCreator{
		title:       "Test Goal",
		description: "Test Description",
		goalType:    models.ChecklistGoal,
		checklistID: "test_checklist",
		scoringType: models.ManualScoring,
		prompt:      "   ", // Whitespace only
	}

	// Build the result
	err := creator.buildResult()
	require.NoError(t, err)

	goal := creator.result
	require.NotNil(t, goal)

	// Should use default prompt
	assert.Equal(t, "Complete your checklist items today", goal.Prompt)
}

func TestChecklistGoalCreator_BuildGoalMissingChecklist(t *testing.T) {
	// Create a test creator without checklist ID
	creator := &ChecklistGoalCreator{
		title:       "Test Goal",
		description: "Test Description",
		goalType:    models.ChecklistGoal,
		checklistID: "", // Missing checklist ID
		scoringType: models.ManualScoring,
		prompt:      "Test prompt",
	}

	// Build the result
	err := creator.buildResult()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checklist selection is required")
	assert.Nil(t, creator.result)
}
