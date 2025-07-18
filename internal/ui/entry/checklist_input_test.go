package entry

import (
	"testing"

	"github.com/davidlee/vice/internal/models"
)

// AIDEV-NOTE: checklist-input-tests; comprehensive testing for checklist input component integration
// Tests dynamic checklist loading, item selection, progress tracking, and scoring integration

func TestChecklistEntryInput(t *testing.T) {
	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title:  "Daily Tasks",
			Prompt: "Which tasks did you complete today?",
		},
		FieldType: models.FieldType{
			Type:        models.ChecklistFieldType,
			ChecklistID: "daily_routine", // This will fallback to placeholder since file doesn't exist
		},
		ExistingEntry: nil,
		ShowScoring:   false,
	}

	input := NewChecklistEntryInput(config)

	// Test field type
	if input.GetFieldType() != models.ChecklistFieldType {
		t.Errorf("GetFieldType() = %v, want %v", input.GetFieldType(), models.ChecklistFieldType)
	}

	// Test that items are loaded (should fallback to placeholder due to missing file)
	if len(input.availableItems) == 0 {
		t.Errorf("availableItems should not be empty")
	}

	// Test that checklist ID is stored
	if input.checklistID != "daily_routine" {
		t.Errorf("checklistID = %v, want %v", input.checklistID, "daily_routine")
	}
}

func TestChecklistEntryInputWithoutChecklistID(t *testing.T) {
	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title: "Test Checklist Habit",
		},
		FieldType: models.FieldType{
			Type: models.ChecklistFieldType,
			// No ChecklistID provided
		},
		ExistingEntry: nil,
		ShowScoring:   false,
	}

	input := NewChecklistEntryInput(config)

	// Should use placeholder items
	expectedItems := []string{"Item 1", "Item 2", "Item 3"}
	if len(input.availableItems) != len(expectedItems) {
		t.Errorf("availableItems length = %v, want %v", len(input.availableItems), len(expectedItems))
	}

	for i, expected := range expectedItems {
		if input.availableItems[i] != expected {
			t.Errorf("availableItems[%d] = %v, want %v", i, input.availableItems[i], expected)
		}
	}
}

func TestChecklistEntryInputSelection(t *testing.T) {
	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title: "Test Selection",
		},
		FieldType: models.FieldType{
			Type: models.ChecklistFieldType,
		},
		ShowScoring: false,
	}

	input := NewChecklistEntryInput(config)

	// Test setting selected items
	testSelection := []string{"Item 1", "Item 3"}
	if err := input.SetExistingValue(testSelection); err != nil {
		t.Errorf("SetExistingValue() unexpected error: %v", err)
	}

	// Test getting selected items
	value := input.GetValue()
	selectedItems, ok := value.([]string)
	if !ok {
		t.Errorf("GetValue() returned %T, want []string", value)
	}

	if len(selectedItems) != len(testSelection) {
		t.Errorf("selected items length = %v, want %v", len(selectedItems), len(testSelection))
	}

	// Test string representation
	stringValue := input.GetStringValue()
	expected := "Item 1, Item 3"
	if stringValue != expected {
		t.Errorf("GetStringValue() = %v, want %v", stringValue, expected)
	}
}

func TestChecklistEntryInputValidation(t *testing.T) {
	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title: "Validation Test",
		},
		FieldType: models.FieldType{
			Type: models.ChecklistFieldType,
		},
		ShowScoring: false,
	}

	input := NewChecklistEntryInput(config)

	// Test valid selection
	validSelection := []string{"Item 1", "Item 2"}
	if err := input.SetExistingValue(validSelection); err != nil {
		t.Errorf("SetExistingValue() unexpected error: %v", err)
	}

	if err := input.Validate(); err != nil {
		t.Errorf("Validate() expected no error for valid selection, got: %v", err)
	}

	// Test invalid selection (item not in available items)
	invalidSelection := []string{"Item 1", "Invalid Item"}
	if err := input.SetExistingValue(invalidSelection); err != nil {
		t.Errorf("SetExistingValue() unexpected error: %v", err)
	}

	if err := input.Validate(); err == nil {
		t.Errorf("Validate() expected error for invalid selection, got nil")
	}
}

func TestChecklistEntryInputExistingValue(t *testing.T) {
	existingSelection := []string{"Item 2", "Item 3"}
	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title: "Existing Value Test",
		},
		FieldType: models.FieldType{
			Type: models.ChecklistFieldType,
		},
		ExistingEntry: &ExistingEntry{
			Value: existingSelection,
		},
		ShowScoring: false,
	}

	input := NewChecklistEntryInput(config)

	// Check that existing value was set
	currentSelection := input.selectedItems
	if len(currentSelection) != len(existingSelection) {
		t.Errorf("selectedItems length = %v, want %v", len(currentSelection), len(existingSelection))
	}

	for i, expected := range existingSelection {
		if currentSelection[i] != expected {
			t.Errorf("selectedItems[%d] = %v, want %v", i, currentSelection[i], expected)
		}
	}
}

func TestChecklistEntryInputScoringAwareness(t *testing.T) {
	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title:       "Scoring Test",
			ScoringType: models.AutomaticScoring,
		},
		FieldType: models.FieldType{
			Type: models.ChecklistFieldType,
		},
		ShowScoring: true,
	}

	input := NewChecklistEntryInput(config)

	if !input.CanShowScoring() {
		t.Errorf("CanShowScoring() expected true for automatic scoring habit")
	}

	// Test scoring display update
	level := models.AchievementMidi
	if err := input.UpdateScoringDisplay(&level); err != nil {
		t.Errorf("UpdateScoringDisplay() unexpected error: %v", err)
	}
}

func TestChecklistEntryInputProgress(t *testing.T) {
	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title: "Progress Test",
		},
		FieldType: models.FieldType{
			Type: models.ChecklistFieldType,
		},
		ShowScoring: false,
	}

	input := NewChecklistEntryInput(config)

	// Initially no items completed
	completed, total := input.GetCompletionProgress()
	if completed != 0 {
		t.Errorf("initial completed = %v, want 0", completed)
	}
	if total != 3 { // Should have 3 placeholder items
		t.Errorf("total = %v, want 3", total)
	}

	// Select some items
	selection := []string{"Item 1", "Item 2"}
	if err := input.SetExistingValue(selection); err != nil {
		t.Errorf("SetExistingValue() unexpected error: %v", err)
	}

	completed, total = input.GetCompletionProgress()
	if completed != 2 {
		t.Errorf("completed after selection = %v, want 2", completed)
	}
	if total != 3 {
		t.Errorf("total after selection = %v, want 3", total)
	}
}

func TestChecklistEntryInputInvalidValueType(t *testing.T) {
	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title: "Invalid Type Test",
		},
		FieldType: models.FieldType{
			Type: models.ChecklistFieldType,
		},
		ShowScoring: false,
	}

	input := NewChecklistEntryInput(config)

	// Test setting invalid value type
	if err := input.SetExistingValue("not a slice"); err == nil {
		t.Errorf("SetExistingValue() expected error for invalid type, got nil")
	}

	// Test setting invalid value type (wrong slice type)
	if err := input.SetExistingValue([]int{1, 2, 3}); err == nil {
		t.Errorf("SetExistingValue() expected error for wrong slice type, got nil")
	}
}

func TestChecklistEntryInputCustomChecklistsPath(t *testing.T) {
	customPath := "/custom/path/checklists.yml"
	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title: "Custom Path Test",
		},
		FieldType: models.FieldType{
			Type:        models.ChecklistFieldType,
			ChecklistID: "test_checklist",
		},
		ChecklistsPath: customPath,
		ShowScoring:    false,
	}

	input := NewChecklistEntryInput(config)

	// Verify custom path is used
	if input.checklistsPath != customPath {
		t.Errorf("checklistsPath = %v, want %v", input.checklistsPath, customPath)
	}
}

func TestFactoryChecklistCreation(t *testing.T) {
	factory := NewEntryFieldInputFactory()

	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title: "Factory Test",
		},
		FieldType: models.FieldType{
			Type: models.ChecklistFieldType,
		},
	}

	input, err := factory.CreateInput(config)
	if err != nil {
		t.Errorf("CreateInput(checklist) unexpected error: %v", err)
	}

	if input.GetFieldType() != models.ChecklistFieldType {
		t.Errorf("Checklist input GetFieldType() = %v, want %v", input.GetFieldType(), models.ChecklistFieldType)
	}

	// Test scoring-aware creation
	scoringInput, err := factory.CreateScoringAwareInput(config)
	if err != nil {
		t.Errorf("CreateScoringAwareInput(checklist) unexpected error: %v", err)
	}

	if scoringInput.GetFieldType() != models.ChecklistFieldType {
		t.Errorf("Scoring-aware checklist input GetFieldType() = %v, want %v", scoringInput.GetFieldType(), models.ChecklistFieldType)
	}
}
