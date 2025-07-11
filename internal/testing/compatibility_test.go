// Package testing provides backwards compatibility and integration testing utilities.
package testing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/iter/internal/parser"
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
			description: "User goals without explicit position fields (T002 failure case)",
		},
		{
			name:        "minimal_goals",
			file:        "minimal_goals.yml",
			description: "Minimal user configuration with single goal",
		},
		{
			name:        "mixed_goal_types",
			file:        "mixed_goal_types.yml",
			description: "Mix of simple, elastic, and informational goals",
		},
	}

	for _, pattern := range userPatterns {
		t.Run(pattern.name, func(t *testing.T) {
			// Load real user data pattern
			patternPath := filepath.Join("../../testdata/user_patterns", pattern.file)
			data, err := os.ReadFile(filepath.Clean(patternPath)) //nolint:gosec // Test files are controlled
			require.NoError(t, err, "Should be able to read user pattern %s", pattern.file)

			// Test that parser can load it
			goalParser := parser.NewGoalParser()
			schema, err := goalParser.ParseYAML(data)
			require.NoError(t, err, "Parser should handle user pattern %s: %s", pattern.file, pattern.description)

			// Test that validation passes
			err = schema.Validate()
			require.NoError(t, err, "Validation should pass for user pattern %s: %s", pattern.file, pattern.description)

			// Verify goals were loaded correctly
			assert.Greater(t, len(schema.Goals), 0, "Should have at least one goal")

			// Verify positions were auto-assigned correctly
			for i, goal := range schema.Goals {
				expectedPosition := i + 1
				assert.Equal(t, expectedPosition, goal.Position,
					"Goal %d should have position %d (auto-assigned)", i, expectedPosition)
			}

			// Verify all goals have valid IDs (auto-generated if needed)
			for _, goal := range schema.Goals {
				assert.NotEmpty(t, goal.ID, "Goal should have an ID")
				assert.NotEmpty(t, goal.Title, "Goal should have a title")
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

	baseGoal := `
goals:
  - title: "Test Goal"
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
			yamlContent += baseGoal

			goalParser := parser.NewGoalParser()
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
// based on the order of goals in the YAML file, addressing the T002 issue.
func TestPositionInferenceFromFileOrder(t *testing.T) {
	yamlContent := `
version: "1.0.0"
goals:
  - title: "First Goal"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "First prompt"
  - title: "Second Goal"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Second prompt"
  - title: "Third Goal"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Third prompt"`

	goalParser := parser.NewGoalParser()
	schema, err := goalParser.ParseYAML([]byte(yamlContent))
	require.NoError(t, err)

	err = schema.Validate()
	require.NoError(t, err)

	// Verify positions were assigned correctly based on order
	require.Len(t, schema.Goals, 3)
	assert.Equal(t, 1, schema.Goals[0].Position)
	assert.Equal(t, 2, schema.Goals[1].Position)
	assert.Equal(t, 3, schema.Goals[2].Position)

	// Verify titles match expected order
	assert.Equal(t, "First Goal", schema.Goals[0].Title)
	assert.Equal(t, "Second Goal", schema.Goals[1].Title)
	assert.Equal(t, "Third Goal", schema.Goals[2].Title)
}

// TestMissingFieldsHandling ensures that goals with missing optional fields still work.
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
goals:
  - title: "Test Goal"
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
goals:
  - title: "Test Goal"
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
goals:
  - title: "Test Goal"
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
			goalParser := parser.NewGoalParser()
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
