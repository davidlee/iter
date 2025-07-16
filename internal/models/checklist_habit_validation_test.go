package models

import (
	"testing"
)

// AIDEV-NOTE: checklist-habit-validation-tests; comprehensive testing for checklist habit validation
// Tests criteria validation, cross-reference validation, and error handling for T007/4.4
// AIDEV-NOTE: testing-patterns; uses table-driven tests for comprehensive coverage

func TestHabitValidateChecklistCriteria(t *testing.T) {
	tests := []struct {
		name        string
		habit       Habit
		expectError bool
		errorSubstr string
	}{
		{
			name: "Valid checklist habit with automatic scoring and criteria",
			habit: Habit{
				Title:       "Test Checklist Habit",
				HabitType:   ChecklistHabit,
				ScoringType: AutomaticScoring,
				FieldType: FieldType{
					Type:        ChecklistFieldType,
					ChecklistID: "test_checklist",
				},
				Criteria: &Criteria{
					Description: "All items completed",
					Condition: &Condition{
						ChecklistCompletion: &ChecklistCompletionCondition{
							RequiredItems: "all",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Valid checklist habit with manual scoring",
			habit: Habit{
				Title:       "Test Checklist Habit",
				HabitType:   ChecklistHabit,
				ScoringType: ManualScoring,
				FieldType: FieldType{
					Type:        ChecklistFieldType,
					ChecklistID: "test_checklist",
				},
			},
			expectError: false,
		},
		{
			name: "Invalid checklist habit - missing scoring type",
			habit: Habit{
				Title:     "Test Checklist Habit",
				HabitType: ChecklistHabit,
				FieldType: FieldType{
					Type:        ChecklistFieldType,
					ChecklistID: "test_checklist",
				},
			},
			expectError: true,
			errorSubstr: "scoring_type is required for checklist habits",
		},
		{
			name: "Invalid checklist habit - wrong field type",
			habit: Habit{
				Title:       "Test Checklist Habit",
				HabitType:   ChecklistHabit,
				ScoringType: AutomaticScoring,
				FieldType: FieldType{
					Type: TextFieldType,
				},
			},
			expectError: true,
			errorSubstr: "checklist habits must use checklist field type",
		},
		{
			name: "Invalid checklist habit - missing checklist_id",
			habit: Habit{
				Title:       "Test Checklist Habit",
				HabitType:   ChecklistHabit,
				ScoringType: AutomaticScoring,
				FieldType: FieldType{
					Type: ChecklistFieldType,
				},
			},
			expectError: true,
			errorSubstr: "checklist_id is required for checklist field type",
		},
		{
			name: "Invalid checklist habit - automatic scoring without criteria",
			habit: Habit{
				Title:       "Test Checklist Habit",
				HabitType:   ChecklistHabit,
				ScoringType: AutomaticScoring,
				FieldType: FieldType{
					Type:        ChecklistFieldType,
					ChecklistID: "test_checklist",
				},
			},
			expectError: true,
			errorSubstr: "criteria is required for automatic scoring of checklist habits",
		},
		{
			name: "Invalid checklist habit - criteria without condition",
			habit: Habit{
				Title:       "Test Checklist Habit",
				HabitType:   ChecklistHabit,
				ScoringType: AutomaticScoring,
				FieldType: FieldType{
					Type:        ChecklistFieldType,
					ChecklistID: "test_checklist",
				},
				Criteria: &Criteria{
					Description: "All items completed",
				},
			},
			expectError: true,
			errorSubstr: "criteria condition is required",
		},
		{
			name: "Invalid checklist habit - invalid checklist completion condition",
			habit: Habit{
				Title:       "Test Checklist Habit",
				HabitType:   ChecklistHabit,
				ScoringType: AutomaticScoring,
				FieldType: FieldType{
					Type:        ChecklistFieldType,
					ChecklistID: "test_checklist",
				},
				Criteria: &Criteria{
					Description: "All items completed",
					Condition: &Condition{
						ChecklistCompletion: &ChecklistCompletionCondition{
							RequiredItems: "invalid",
						},
					},
				},
			},
			expectError: true,
			errorSubstr: "required_items must be 'all'",
		},
		{
			name: "Invalid checklist habit - empty required_items",
			habit: Habit{
				Title:       "Test Checklist Habit",
				HabitType:   ChecklistHabit,
				ScoringType: AutomaticScoring,
				FieldType: FieldType{
					Type:        ChecklistFieldType,
					ChecklistID: "test_checklist",
				},
				Criteria: &Criteria{
					Description: "All items completed",
					Condition: &Condition{
						ChecklistCompletion: &ChecklistCompletionCondition{
							RequiredItems: "",
						},
					},
				},
			},
			expectError: true,
			errorSubstr: "required_items field is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.habit.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Validate() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !containsSubstring(err.Error(), tt.errorSubstr) {
					t.Errorf("Validate() error = %q, expected to contain %q", err.Error(), tt.errorSubstr)
				}
			} else if err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestHabitValidateWithChecklistContext(t *testing.T) {
	tests := []struct {
		name            string
		habit           Habit
		checklistsExist func(string) bool
		expectError     bool
		errorSubstr     string
	}{
		{
			name: "Valid checklist habit with existing checklist reference",
			habit: Habit{
				Title:       "Test Checklist Habit",
				HabitType:   ChecklistHabit,
				ScoringType: ManualScoring,
				FieldType: FieldType{
					Type:        ChecklistFieldType,
					ChecklistID: "existing_checklist",
				},
			},
			checklistsExist: func(id string) bool {
				return id == "existing_checklist"
			},
			expectError: false,
		},
		{
			name: "Invalid checklist habit with non-existent checklist reference",
			habit: Habit{
				Title:       "Test Checklist Habit",
				HabitType:   ChecklistHabit,
				ScoringType: ManualScoring,
				FieldType: FieldType{
					Type:        ChecklistFieldType,
					ChecklistID: "non_existent_checklist",
				},
			},
			checklistsExist: func(id string) bool {
				return id == "existing_checklist"
			},
			expectError: true,
			errorSubstr: "references non-existent checklist 'non_existent_checklist'",
		},
		{
			name: "Valid non-checklist habit (should not be affected)",
			habit: Habit{
				Title:       "Test Simple Habit",
				HabitType:   SimpleHabit,
				ScoringType: ManualScoring,
				FieldType: FieldType{
					Type: BooleanFieldType,
				},
			},
			checklistsExist: func(_ string) bool {
				return false // No checklists exist
			},
			expectError: false,
		},
		{
			name: "Invalid checklist habit with basic validation errors",
			habit: Habit{
				Title:     "", // Invalid: empty title
				HabitType: ChecklistHabit,
				FieldType: FieldType{
					Type:        ChecklistFieldType,
					ChecklistID: "existing_checklist",
				},
			},
			checklistsExist: func(id string) bool {
				return id == "existing_checklist"
			},
			expectError: true,
			errorSubstr: "habit title is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.habit.ValidateWithChecklistContext(tt.checklistsExist)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateWithChecklistContext() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !containsSubstring(err.Error(), tt.errorSubstr) {
					t.Errorf("ValidateWithChecklistContext() error = %q, expected to contain %q", err.Error(), tt.errorSubstr)
				}
			} else if err != nil {
				t.Errorf("ValidateWithChecklistContext() unexpected error: %v", err)
			}
		})
	}
}

func TestHabitValidateChecklistCriteriaMethod(t *testing.T) {
	habit := Habit{
		Title:       "Test Habit",
		HabitType:   ChecklistHabit,
		ScoringType: AutomaticScoring,
		FieldType: FieldType{
			Type:        ChecklistFieldType,
			ChecklistID: "test_checklist",
		},
	}

	tests := []struct {
		name        string
		criteria    *Criteria
		expectError bool
		errorSubstr string
	}{
		{
			name:        "Nil criteria",
			criteria:    nil,
			expectError: false,
		},
		{
			name: "Valid criteria with checklist completion condition",
			criteria: &Criteria{
				Description: "All items completed",
				Condition: &Condition{
					ChecklistCompletion: &ChecklistCompletionCondition{
						RequiredItems: "all",
					},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid criteria - missing condition",
			criteria: &Criteria{
				Description: "All items completed",
			},
			expectError: true,
			errorSubstr: "criteria condition is required",
		},
		{
			name: "Invalid criteria - invalid checklist completion condition",
			criteria: &Criteria{
				Description: "All items completed",
				Condition: &Condition{
					ChecklistCompletion: &ChecklistCompletionCondition{
						RequiredItems: "invalid",
					},
				},
			},
			expectError: true,
			errorSubstr: "required_items must be 'all'",
		},
		{
			name: "Valid criteria - no checklist completion condition",
			criteria: &Criteria{
				Description: "Some other criteria",
				Condition: &Condition{
					Equals: boolPtr(true),
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := habit.validateChecklistCriteria(tt.criteria)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateChecklistCriteria() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !containsSubstring(err.Error(), tt.errorSubstr) {
					t.Errorf("validateChecklistCriteria() error = %q, expected to contain %q", err.Error(), tt.errorSubstr)
				}
			} else if err != nil {
				t.Errorf("validateChecklistCriteria() unexpected error: %v", err)
			}
		})
	}
}

func TestChecklistCompletionConditionValidate(t *testing.T) {
	tests := []struct {
		name        string
		condition   ChecklistCompletionCondition
		expectError bool
		errorSubstr string
	}{
		{
			name: "Valid condition with 'all' required items",
			condition: ChecklistCompletionCondition{
				RequiredItems: "all",
			},
			expectError: false,
		},
		{
			name: "Invalid condition - empty required items",
			condition: ChecklistCompletionCondition{
				RequiredItems: "",
			},
			expectError: true,
			errorSubstr: "required_items field is required",
		},
		{
			name: "Invalid condition - invalid required items value",
			condition: ChecklistCompletionCondition{
				RequiredItems: "some",
			},
			expectError: true,
			errorSubstr: "required_items must be 'all'",
		},
		{
			name: "Invalid condition - numeric value",
			condition: ChecklistCompletionCondition{
				RequiredItems: "50",
			},
			expectError: true,
			errorSubstr: "required_items must be 'all'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.condition.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Validate() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !containsSubstring(err.Error(), tt.errorSubstr) {
					t.Errorf("Validate() error = %q, expected to contain %q", err.Error(), tt.errorSubstr)
				}
			} else if err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

// Helper functions for tests
func containsSubstring(str, substr string) bool {
	return len(str) > 0 && len(substr) > 0 &&
		len(str) >= len(substr) &&
		findSubstring(str, substr)
}

func findSubstring(str, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(str) < len(substr) {
		return false
	}

	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
