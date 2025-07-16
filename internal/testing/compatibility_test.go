// Package testing provides backwards compatibility and integration testing utilities.
package testing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/vice/internal/parser"
)

// TestUserDataBackwardsCompatibility ensures that real user data patterns continue to work
// after schema or validation changes. This prevents T002-style failures where documentation
// changes weren't synchronized with validation code.
func TestUserDataBackwardsCompatibility(t *testing.T) {
	userPatterns := []struct {
		name        string
		file        string
		description string
	}{
		{
			name:        "no_position_fields",
			file:        "no_position_fields.yml",
			description: "User habits without explicit position fields (T002 failure case)",
		},
		{
			name:        "minimal_habits",
			file:        "minimal_habits.yml",
			description: "Minimal user configuration with single habit",
		},
		{
			name:        "mixed_goal_types",
			file:        "mixed_goal_types.yml",
			description: "Mix of simple, elastic, and informational habits",
		},
	}

	for _, pattern := range userPatterns {
		t.Run(pattern.name, func(t *testing.T) {
			// Load real user data pattern
			patternPath := filepath.Join("../../testdata/user_patterns", pattern.file)
			data, err := os.ReadFile(filepath.Clean(patternPath)) //nolint:gosec // Test files are controlled
			require.NoError(t, err, "Should be able to read user pattern %s", pattern.file)

			// Test that parser can load it
			goalParser := parser.NewHabitParser()
			schema, err := goalParser.ParseYAML(data)
			require.NoError(t, err, "Parser should handle user pattern %s: %s", pattern.file, pattern.description)

			// Test that validation passes
			err = schema.Validate()
			require.NoError(t, err, "Validation should pass for user pattern %s: %s", pattern.file, pattern.description)

			// Verify habits were loaded correctly
			assert.Greater(t, len(schema.Habits), 0, "Should have at least one habit")

			// Verify positions were auto-assigned correctly
			for i, habit := range schema.Habits {
				expectedPosition := i + 1
				assert.Equal(t, expectedPosition, habit.Position,
					"Habit %d should have position %d (auto-assigned)", i, expectedPosition)
			}

			// Verify all habits have valid IDs (auto-generated if needed)
			for _, habit := range schema.Habits {
				assert.NotEmpty(t, habit.ID, "Habit should have an ID")
				assert.NotEmpty(t, habit.Title, "Habit should have a title")
			}
		})
	}
}

// TestSchemaVersionCompatibility ensures that different schema versions continue to work.
func TestSchemaVersionCompatibility(t *testing.T) {
	versions := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"1.1.0", true}, // Should work with future versions
		{"", false},     // Empty version should fail
	}

	baseHabit := `
habits:
  - title: "Test Habit"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Test prompt"`

	for _, tc := range versions {
		t.Run(tc.version, func(t *testing.T) {
			yamlContent := ""
			if tc.version != "" {
				yamlContent = "version: \"" + tc.version + "\"\n"
			}
			yamlContent += baseHabit

			goalParser := parser.NewHabitParser()
			schema, err := goalParser.ParseYAML([]byte(yamlContent))

			if tc.valid {
				require.NoError(t, err, "Version %s should be parseable", tc.version)
				err = schema.Validate()
				require.NoError(t, err, "Version %s should validate", tc.version)
			} else {
				// Empty version should be caught during validation
				if err == nil {
					err = schema.Validate()
				}
				assert.Error(t, err, "Version %s should be invalid", tc.version)
			}
		})
	}
}

// TestPositionInferenceFromFileOrder verifies that positions are correctly assigned
// based on the order of habits in the YAML file, addressing the T002 issue.
func TestPositionInferenceFromFileOrder(t *testing.T) {
	yamlContent := `
version: "1.0.0"
habits:
  - title: "First Habit"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "First prompt"
  - title: "Second Habit"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Second prompt"
  - title: "Third Habit"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Third prompt"`

	goalParser := parser.NewHabitParser()
	schema, err := goalParser.ParseYAML([]byte(yamlContent))
	require.NoError(t, err)

	err = schema.Validate()
	require.NoError(t, err)

	// Verify positions were assigned correctly based on order
	require.Len(t, schema.Habits, 3)
	assert.Equal(t, 1, schema.Habits[0].Position)
	assert.Equal(t, 2, schema.Habits[1].Position)
	assert.Equal(t, 3, schema.Habits[2].Position)

	// Verify titles match expected order
	assert.Equal(t, "First Habit", schema.Habits[0].Title)
	assert.Equal(t, "Second Habit", schema.Habits[1].Title)
	assert.Equal(t, "Third Habit", schema.Habits[2].Title)
}

// TestMissingFieldsHandling ensures that habits with missing optional fields still work.
func TestMissingFieldsHandling(t *testing.T) {
	testCases := []struct {
		name        string
		yamlContent string
		shouldWork  bool
	}{
		{
			name: "missing_created_date",
			yamlContent: `
version: "1.0.0"
habits:
  - title: "Test Habit"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Test prompt"`,
			shouldWork: true,
		},
		{
			name: "missing_description",
			yamlContent: `
version: "1.0.0"
habits:
  - title: "Test Habit"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Test prompt"`,
			shouldWork: true,
		},
		{
			name: "missing_help_text",
			yamlContent: `
version: "1.0.0"
habits:
  - title: "Test Habit"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Test prompt"`,
			shouldWork: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			goalParser := parser.NewHabitParser()
			schema, err := goalParser.ParseYAML([]byte(tc.yamlContent))

			if tc.shouldWork {
				require.NoError(t, err, "Should parse successfully")
				err = schema.Validate()
				require.NoError(t, err, "Should validate successfully")
			} else {
				if err == nil {
					err = schema.Validate()
				}
				assert.Error(t, err, "Should fail validation")
			}
		})
	}
}
