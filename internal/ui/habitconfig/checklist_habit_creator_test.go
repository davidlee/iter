package habitconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidlee/vice/internal/models"
)

func TestChecklistHabitCreator_LoadAvailableChecklists(t *testing.T) {
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
	creator := NewChecklistHabitCreator("Test Habit", "Test Description", models.ChecklistHabit, checklistsFile)

	// Verify no error occurred during initialization
	assert.NoError(t, creator.err)

	// Verify checklists were loaded
	assert.Len(t, creator.availableChecklists, 2)
	assert.Equal(t, "morning_routine", creator.availableChecklists[0].ID)
	assert.Equal(t, "Morning Routine", creator.availableChecklists[0].Title)
	assert.Equal(t, "evening_routine", creator.availableChecklists[1].ID)
	assert.Equal(t, "Evening Routine", creator.availableChecklists[1].Title)
}

func TestChecklistHabitCreator_NoChecklistsFile(t *testing.T) {
	// Test with non-existent checklists file
	creator := NewChecklistHabitCreator("Test Habit", "Test Description", models.ChecklistHabit, "/nonexistent/checklists.yml")

	// Should have error when no checklists are available (form creation sets this error)
	assert.Error(t, creator.err)
	assert.Contains(t, creator.err.Error(), "no checklists found")
	assert.Len(t, creator.availableChecklists, 0)
}

func TestChecklistHabitCreator_BuildHabitAutomaticScoring(t *testing.T) {
	// Create a test creator with mock data
	creator := &ChecklistHabitCreator{
		title:       "Morning Routine Habit",
		description: "Complete morning tasks",
		habitType:   models.ChecklistHabit,
		checklistID: "morning_routine",
		scoringType: models.AutomaticScoring,
		prompt:      "Did you complete your morning routine?",
	}

	// Build the result
	err := creator.buildResult()
	require.NoError(t, err)

	habit := creator.result
	require.NotNil(t, habit)

	// Verify habit properties
	assert.Equal(t, "Morning Routine Habit", habit.Title)
	assert.Equal(t, "Complete morning tasks", habit.Description)
	assert.Equal(t, models.ChecklistHabit, habit.HabitType)
	assert.Equal(t, models.ChecklistFieldType, habit.FieldType.Type)
	assert.Equal(t, "morning_routine", habit.FieldType.ChecklistID)
	assert.Equal(t, models.AutomaticScoring, habit.ScoringType)
	assert.Equal(t, "Did you complete your morning routine?", habit.Prompt)

	// Verify automatic scoring criteria
	require.NotNil(t, habit.Criteria)
	assert.Equal(t, "All checklist items completed", habit.Criteria.Description)
	require.NotNil(t, habit.Criteria.Condition)
	require.NotNil(t, habit.Criteria.Condition.ChecklistCompletion)
	assert.Equal(t, "all", habit.Criteria.Condition.ChecklistCompletion.RequiredItems)
}

func TestChecklistHabitCreator_BuildHabitManualScoring(t *testing.T) {
	// Create a test creator with manual scoring
	creator := &ChecklistHabitCreator{
		title:       "Evening Routine Habit",
		description: "Complete evening tasks",
		habitType:   models.ChecklistHabit,
		checklistID: "evening_routine",
		scoringType: models.ManualScoring,
		prompt:      "How well did you complete your evening routine?",
	}

	// Build the result
	err := creator.buildResult()
	require.NoError(t, err)

	habit := creator.result
	require.NotNil(t, habit)

	// Verify habit properties
	assert.Equal(t, "Evening Routine Habit", habit.Title)
	assert.Equal(t, "Complete evening tasks", habit.Description)
	assert.Equal(t, models.ChecklistHabit, habit.HabitType)
	assert.Equal(t, models.ChecklistFieldType, habit.FieldType.Type)
	assert.Equal(t, "evening_routine", habit.FieldType.ChecklistID)
	assert.Equal(t, models.ManualScoring, habit.ScoringType)
	assert.Equal(t, "How well did you complete your evening routine?", habit.Prompt)

	// Manual scoring should not have automatic criteria
	assert.Nil(t, habit.Criteria)
}

func TestChecklistHabitCreator_BuildHabitWithEmptyPrompt(t *testing.T) {
	// Create a test creator with empty prompt
	creator := &ChecklistHabitCreator{
		title:       "Test Habit",
		description: "Test Description",
		habitType:   models.ChecklistHabit,
		checklistID: "test_checklist",
		scoringType: models.ManualScoring,
		prompt:      "   ", // Whitespace only
	}

	// Build the result
	err := creator.buildResult()
	require.NoError(t, err)

	habit := creator.result
	require.NotNil(t, habit)

	// Should use default prompt
	assert.Equal(t, "Complete your checklist items today", habit.Prompt)
}

func TestChecklistHabitCreator_BuildHabitMissingChecklist(t *testing.T) {
	// Create a test creator without checklist ID
	creator := &ChecklistHabitCreator{
		title:       "Test Habit",
		description: "Test Description",
		habitType:   models.ChecklistHabit,
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
