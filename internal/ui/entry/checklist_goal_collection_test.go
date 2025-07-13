package entry

import (
	"os"
	"path/filepath"
	"testing"

	"davidlee/iter/internal/models"
	"davidlee/iter/internal/scoring"
)

// AIDEV-NOTE: checklist-goal-tests; comprehensive testing for checklist goal collection flow
// Tests automatic/manual scoring, checklist data loading, and field type integration for T007/4.3
// AIDEV-NOTE: testing-patterns; uses NewChecklistGoalCollectionFlow() and CollectEntryDirectly() for headless testing
// All major scenarios covered: checklist completion, automatic scoring with criteria, manual scoring, error handling

func TestChecklistGoalCollectionFlow(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	checklistsPath := filepath.Join(tempDir, "checklists.yml")

	// Create test checklist data
	testChecklistData := `version: "1.0.0"
created_date: "2024-01-01"
checklists:
  - id: "test_checklist"
    title: "Test Checklist"
    description: "Test checklist for testing"
    items:
      - "# Section 1"
      - "Item 1"
      - "Item 2"
      - "# Section 2"
      - "Item 3"
      - "Item 4"
    created_date: "2024-01-01"
    modified_date: "2024-01-01"`

	err := os.WriteFile(checklistsPath, []byte(testChecklistData), 0600)
	if err != nil {
		t.Fatalf("Failed to create test checklist file: %v", err)
	}

	factory := NewEntryFieldInputFactory()
	scoringEngine := &scoring.Engine{} // Mock or real scoring engine

	flow := NewChecklistGoalCollectionFlow(factory, scoringEngine, checklistsPath)

	// Test flow type identification
	if flow.GetFlowType() != string(models.ChecklistGoal) {
		t.Errorf("GetFlowType() = %v, want %v", flow.GetFlowType(), string(models.ChecklistGoal))
	}

	// Test scoring requirement
	if !flow.RequiresScoring() {
		t.Errorf("RequiresScoring() expected true for checklist goals")
	}

	// Test supported field types
	expectedFieldTypes := []string{models.ChecklistFieldType}

	supportedTypes := flow.GetExpectedFieldTypes()
	if len(supportedTypes) != len(expectedFieldTypes) {
		t.Errorf("GetExpectedFieldTypes() length = %v, want %v", len(supportedTypes), len(expectedFieldTypes))
	}

	for _, expectedType := range expectedFieldTypes {
		found := false
		for _, supportedType := range supportedTypes {
			if supportedType == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetExpectedFieldTypes() missing expected type: %v", expectedType)
		}
	}
}

func TestChecklistGoalLoadChecklistData(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	checklistsPath := filepath.Join(tempDir, "checklists.yml")

	// Create test checklist data
	testChecklistData := `version: "1.0.0"
created_date: "2024-01-01"
checklists:
  - id: "test_checklist"
    title: "Test Checklist"
    description: "Test checklist for testing"
    items:
      - "# Section 1"
      - "Item 1"
      - "Item 2"
      - "Item 3"
    created_date: "2024-01-01"
    modified_date: "2024-01-01"`

	err := os.WriteFile(checklistsPath, []byte(testChecklistData), 0600)
	if err != nil {
		t.Fatalf("Failed to create test checklist file: %v", err)
	}

	factory := NewEntryFieldInputFactory()
	scoringEngine := &scoring.Engine{}
	flow := NewChecklistGoalCollectionFlow(factory, scoringEngine, checklistsPath)

	// Test valid checklist loading
	goal := models.Goal{
		Title:    "Test Goal",
		GoalType: models.ChecklistGoal,
		FieldType: models.FieldType{
			Type:        models.ChecklistFieldType,
			ChecklistID: "test_checklist",
		},
	}

	checklist, err := flow.loadChecklistData(goal)
	if err != nil {
		t.Errorf("loadChecklistData() failed: %v", err)
	}

	if checklist.ID != "test_checklist" {
		t.Errorf("loadChecklistData() ID = %v, want %v", checklist.ID, "test_checklist")
	}

	if checklist.Title != "Test Checklist" {
		t.Errorf("loadChecklistData() Title = %v, want %v", checklist.Title, "Test Checklist")
	}

	if len(checklist.Items) != 4 {
		t.Errorf("loadChecklistData() Items length = %v, want %v", len(checklist.Items), 4)
	}

	// Test item count calculation
	totalItems := checklist.GetTotalItemCount()
	if totalItems != 3 { // Should exclude heading items starting with "# "
		t.Errorf("GetTotalItemCount() = %v, want %v", totalItems, 3)
	}
}

func TestChecklistGoalLoadChecklistDataErrors(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	scoringEngine := &scoring.Engine{}

	// Test with non-existent file
	flow := NewChecklistGoalCollectionFlow(factory, scoringEngine, "/nonexistent/path.yml")

	goal := models.Goal{
		Title:    "Test Goal",
		GoalType: models.ChecklistGoal,
		FieldType: models.FieldType{
			Type:        models.ChecklistFieldType,
			ChecklistID: "test_checklist",
		},
	}

	_, err := flow.loadChecklistData(goal)
	if err == nil {
		t.Error("loadChecklistData() expected error for non-existent file")
	}

	// Test with invalid checklist ID
	tempDir := t.TempDir()
	checklistsPath := filepath.Join(tempDir, "checklists.yml")

	testChecklistData := `version: "1.0.0"
created_date: "2024-01-01"
checklists:
  - id: "different_checklist"
    title: "Different Checklist"
    items: []`

	err = os.WriteFile(checklistsPath, []byte(testChecklistData), 0600)
	if err != nil {
		t.Fatalf("Failed to create test checklist file: %v", err)
	}

	flow = NewChecklistGoalCollectionFlow(factory, scoringEngine, checklistsPath)
	_, err = flow.loadChecklistData(goal)
	if err == nil {
		t.Error("loadChecklistData() expected error for invalid checklist ID")
	}
}

func TestChecklistGoalCollectEntryDirectly(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	checklistsPath := filepath.Join(tempDir, "checklists.yml")

	// Create test checklist data
	testChecklistData := `version: "1.0.0"
created_date: "2024-01-01"
checklists:
  - id: "test_checklist"
    title: "Test Checklist"
    description: "Test checklist for testing"
    items:
      - "# Section 1"
      - "Item 1"
      - "Item 2"
      - "Item 3"
    created_date: "2024-01-01"
    modified_date: "2024-01-01"`

	err := os.WriteFile(checklistsPath, []byte(testChecklistData), 0600)
	if err != nil {
		t.Fatalf("Failed to create test checklist file: %v", err)
	}

	factory := NewEntryFieldInputFactory()
	scoringEngine := &scoring.Engine{}
	flow := NewChecklistGoalCollectionFlow(factory, scoringEngine, checklistsPath)

	tests := []struct {
		name                string
		scoringType         models.ScoringType
		value               interface{}
		expectedAchievement models.AchievementLevel
		expectError         bool
	}{
		{
			name:                "Manual scoring with full completion",
			scoringType:         models.ManualScoring,
			value:               []string{"Item 1", "Item 2", "Item 3"},
			expectedAchievement: models.AchievementMaxi,
			expectError:         false,
		},
		{
			name:                "Manual scoring with partial completion",
			scoringType:         models.ManualScoring,
			value:               []string{"Item 1", "Item 2"},
			expectedAchievement: models.AchievementMini, // 2/3 = 66.7%, which is >= 50% but < 75%
			expectError:         false,
		},
		{
			name:                "Manual scoring with minimal completion",
			scoringType:         models.ManualScoring,
			value:               []string{"Item 1"},
			expectedAchievement: models.AchievementNone, // 1/3 = 33.3%, which is < 50%
			expectError:         false,
		},
		{
			name:                "Manual scoring with no completion",
			scoringType:         models.ManualScoring,
			value:               []string{},
			expectedAchievement: models.AchievementNone,
			expectError:         false,
		},
		{
			name:                "Automatic scoring with full completion",
			scoringType:         models.AutomaticScoring,
			value:               []string{"Item 1", "Item 2", "Item 3"},
			expectedAchievement: models.AchievementMaxi,
			expectError:         false,
		},
		{
			name:                "Automatic scoring with partial completion",
			scoringType:         models.AutomaticScoring,
			value:               []string{"Item 1", "Item 2"},
			expectedAchievement: models.AchievementNone, // With "all" criteria, partial completion = None
			expectError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goal := models.Goal{
				Title:       "Test Goal",
				GoalType:    models.ChecklistGoal,
				ScoringType: tt.scoringType,
				FieldType: models.FieldType{
					Type:        models.ChecklistFieldType,
					ChecklistID: "test_checklist",
				},
			}

			// Add criteria for automatic scoring tests
			if tt.scoringType == models.AutomaticScoring {
				goal.Criteria = &models.Criteria{
					Condition: &models.Condition{
						ChecklistCompletion: &models.ChecklistCompletionCondition{
							RequiredItems: "all",
						},
					},
				}
			}

			result, err := flow.CollectEntryDirectly(goal, tt.value, "test notes", nil)

			if tt.expectError {
				if err == nil {
					t.Errorf("CollectEntryDirectly() expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("CollectEntryDirectly() unexpected error: %v", err)
				return
			}

			if result.AchievementLevel == nil {
				t.Errorf("CollectEntryDirectly() AchievementLevel is nil")
				return
			}

			if *result.AchievementLevel != tt.expectedAchievement {
				t.Errorf("CollectEntryDirectly() AchievementLevel = %v, want %v", *result.AchievementLevel, tt.expectedAchievement)
			}

			if result.Notes != "test notes" {
				t.Errorf("CollectEntryDirectly() Notes = %v, want %v", result.Notes, "test notes")
			}

			if result.Status != models.EntryCompleted {
				t.Errorf("CollectEntryDirectly() Status = %v, want %v", result.Status, models.EntryCompleted)
			}
		})
	}
}

func TestChecklistGoalDetermineTestingAchievementLevel(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	checklistsPath := filepath.Join(tempDir, "checklists.yml")

	// Create test checklist data
	testChecklistData := `version: "1.0.0"
created_date: "2024-01-01"
checklists:
  - id: "test_checklist"
    title: "Test Checklist"
    items:
      - "Item 1"
      - "Item 2"
      - "Item 3"
      - "Item 4"
    created_date: "2024-01-01"
    modified_date: "2024-01-01"`

	err := os.WriteFile(checklistsPath, []byte(testChecklistData), 0600)
	if err != nil {
		t.Fatalf("Failed to create test checklist file: %v", err)
	}

	factory := NewEntryFieldInputFactory()
	scoringEngine := &scoring.Engine{}
	flow := NewChecklistGoalCollectionFlow(factory, scoringEngine, checklistsPath)

	goal := models.Goal{
		Title:    "Test Goal",
		GoalType: models.ChecklistGoal,
		FieldType: models.FieldType{
			Type:        models.ChecklistFieldType,
			ChecklistID: "test_checklist",
		},
	}

	tests := []struct {
		name                string
		value               interface{}
		expectedAchievement models.AchievementLevel
	}{
		{
			name:                "Full completion (100%)",
			value:               []string{"Item 1", "Item 2", "Item 3", "Item 4"},
			expectedAchievement: models.AchievementMaxi,
		},
		{
			name:                "High completion (75%)",
			value:               []string{"Item 1", "Item 2", "Item 3"},
			expectedAchievement: models.AchievementMidi,
		},
		{
			name:                "Medium completion (50%)",
			value:               []string{"Item 1", "Item 2"},
			expectedAchievement: models.AchievementMini,
		},
		{
			name:                "Low completion (25%)",
			value:               []string{"Item 1"},
			expectedAchievement: models.AchievementNone,
		},
		{
			name:                "No completion (0%)",
			value:               []string{},
			expectedAchievement: models.AchievementNone,
		},
		{
			name:                "Invalid value type",
			value:               "invalid",
			expectedAchievement: models.AchievementNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := flow.determineTestingAchievementLevel(goal, tt.value)

			if level == nil {
				t.Errorf("determineTestingAchievementLevel() returned nil")
				return
			}

			if *level != tt.expectedAchievement {
				t.Errorf("determineTestingAchievementLevel() = %v, want %v", *level, tt.expectedAchievement)
			}
		})
	}
}

func TestChecklistGoalDetermineTestingAchievementLevelErrors(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	scoringEngine := &scoring.Engine{}

	// Test with non-existent checklist file
	flow := NewChecklistGoalCollectionFlow(factory, scoringEngine, "/nonexistent/path.yml")

	goal := models.Goal{
		Title:    "Test Goal",
		GoalType: models.ChecklistGoal,
		FieldType: models.FieldType{
			Type:        models.ChecklistFieldType,
			ChecklistID: "test_checklist",
		},
	}

	level := flow.determineTestingAchievementLevel(goal, []string{"Item 1"})

	if level == nil {
		t.Errorf("determineTestingAchievementLevel() returned nil")
		return
	}

	if *level != models.AchievementNone {
		t.Errorf("determineTestingAchievementLevel() = %v, want %v for error case", *level, models.AchievementNone)
	}
}

func TestChecklistGoalAutomaticScoringWithCriteria(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	checklistsPath := filepath.Join(tempDir, "checklists.yml")

	// Create test checklist data
	testChecklistData := `version: "1.0.0"
created_date: "2024-01-01"
checklists:
  - id: "test_checklist"
    title: "Test Checklist"
    items:
      - "Item 1"
      - "Item 2"
      - "Item 3"
    created_date: "2024-01-01"
    modified_date: "2024-01-01"`

	err := os.WriteFile(checklistsPath, []byte(testChecklistData), 0600)
	if err != nil {
		t.Fatalf("Failed to create test checklist file: %v", err)
	}

	factory := NewEntryFieldInputFactory()
	scoringEngine := &scoring.Engine{}
	flow := NewChecklistGoalCollectionFlow(factory, scoringEngine, checklistsPath)

	goal := models.Goal{
		Title:       "Test Goal",
		GoalType:    models.ChecklistGoal,
		ScoringType: models.AutomaticScoring,
		FieldType: models.FieldType{
			Type:        models.ChecklistFieldType,
			ChecklistID: "test_checklist",
		},
		Criteria: &models.Criteria{
			Condition: &models.Condition{
				ChecklistCompletion: &models.ChecklistCompletionCondition{
					RequiredItems: "all",
				},
			},
		},
	}

	tests := []struct {
		name                string
		value               interface{}
		expectedAchievement models.AchievementLevel
		expectError         bool
	}{
		{
			name:                "Full completion with 'all' criteria",
			value:               []string{"Item 1", "Item 2", "Item 3"},
			expectedAchievement: models.AchievementMaxi,
			expectError:         false,
		},
		{
			name:                "Partial completion with 'all' criteria",
			value:               []string{"Item 1", "Item 2"},
			expectedAchievement: models.AchievementNone,
			expectError:         false,
		},
		{
			name:                "No completion with 'all' criteria",
			value:               []string{},
			expectedAchievement: models.AchievementNone,
			expectError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := flow.CollectEntryDirectly(goal, tt.value, "test notes", nil)

			if tt.expectError {
				if err == nil {
					t.Errorf("CollectEntryDirectly() expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("CollectEntryDirectly() unexpected error: %v", err)
				return
			}

			if result.AchievementLevel == nil {
				t.Errorf("CollectEntryDirectly() AchievementLevel is nil")
				return
			}

			if *result.AchievementLevel != tt.expectedAchievement {
				t.Errorf("CollectEntryDirectly() AchievementLevel = %v, want %v", *result.AchievementLevel, tt.expectedAchievement)
			}
		})
	}
}
