package goalconfig

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
)

func TestYAMLFixtureValidation(t *testing.T) {
	goalParser := parser.NewGoalParser()

	t.Run("valid simple goal fixture", func(t *testing.T) {
		fixturePath := filepath.Join("..", "..", "..", "testdata", "goals", "valid_simple_goal.yml")

		schema, err := goalParser.LoadFromFile(fixturePath)
		require.NoError(t, err)
		assert.NotNil(t, schema)

		// Verify schema structure
		assert.Equal(t, "1.0.0", schema.Version)
		assert.Equal(t, "2024-01-01", schema.CreatedDate)
		assert.Len(t, schema.Goals, 2)

		// Verify first goal (manual scoring)
		goal1 := schema.Goals[0]
		assert.Equal(t, "Daily Exercise", goal1.Title)
		assert.Equal(t, "daily_exercise", goal1.ID)
		assert.Equal(t, models.SimpleGoal, goal1.GoalType)
		assert.Equal(t, models.BooleanFieldType, goal1.FieldType.Type)
		assert.Equal(t, models.ManualScoring, goal1.ScoringType)
		assert.Equal(t, "Did you exercise today?", goal1.Prompt)
		assert.Nil(t, goal1.Criteria)

		// Verify second goal (automatic scoring)
		goal2 := schema.Goals[1]
		assert.Equal(t, "Read for 30 Minutes", goal2.Title)
		assert.Equal(t, "daily_reading", goal2.ID)
		assert.Equal(t, models.AutomaticScoring, goal2.ScoringType)
		require.NotNil(t, goal2.Criteria)
		assert.Equal(t, "Reading completed", goal2.Criteria.Description)
		require.NotNil(t, goal2.Criteria.Condition.Equals)
		assert.Equal(t, true, *goal2.Criteria.Condition.Equals)
	})

	t.Run("valid informational goals fixture", func(t *testing.T) {
		fixturePath := filepath.Join("..", "..", "..", "testdata", "goals", "valid_informational_goals.yml")

		schema, err := goalParser.LoadFromFile(fixturePath)
		require.NoError(t, err)
		assert.NotNil(t, schema)

		// Should have 8 informational goals with different field types
		assert.Len(t, schema.Goals, 8)

		// Verify all goals are informational type
		for _, goal := range schema.Goals {
			assert.Equal(t, models.InformationalGoal, goal.GoalType)
			assert.Equal(t, models.ManualScoring, goal.ScoringType)
			assert.NotEmpty(t, goal.Direction)
		}

		// Test specific field type examples
		fieldTypeTests := map[string]struct {
			goalIndex    int
			expectedType string
			hasUnit      bool
			hasMultiline bool
			hasMin       bool
			hasMax       bool
			direction    string
		}{
			"boolean": {
				goalIndex:    0,
				expectedType: models.BooleanFieldType,
				direction:    "neutral",
			},
			"text_single": {
				goalIndex:    1,
				expectedType: models.TextFieldType,
				hasMultiline: true,
				direction:    "neutral",
			},
			"text_multi": {
				goalIndex:    2,
				expectedType: models.TextFieldType,
				hasMultiline: true,
				direction:    "neutral",
			},
			"unsigned_int": {
				goalIndex:    3,
				expectedType: models.UnsignedIntFieldType,
				hasUnit:      true,
				direction:    "higher_better",
			},
			"unsigned_decimal_constrained": {
				goalIndex:    4,
				expectedType: models.UnsignedDecimalFieldType,
				hasUnit:      true,
				hasMin:       true,
				hasMax:       true,
				direction:    "neutral",
			},
			"decimal_constrained": {
				goalIndex:    5,
				expectedType: models.DecimalFieldType,
				hasUnit:      true,
				hasMin:       true,
				hasMax:       true,
				direction:    "neutral",
			},
			"time": {
				goalIndex:    6,
				expectedType: models.TimeFieldType,
				direction:    "lower_better",
			},
			"duration": {
				goalIndex:    7,
				expectedType: models.DurationFieldType,
				direction:    "higher_better",
			},
		}

		for name, test := range fieldTypeTests {
			t.Run(name, func(t *testing.T) {
				goal := schema.Goals[test.goalIndex]
				assert.Equal(t, test.expectedType, goal.FieldType.Type)
				assert.Equal(t, test.direction, goal.Direction)

				if test.hasUnit {
					assert.NotEmpty(t, goal.FieldType.Unit)
				}
				if test.hasMultiline {
					assert.NotNil(t, goal.FieldType.Multiline)
				}
				if test.hasMin {
					assert.NotNil(t, goal.FieldType.Min)
				}
				if test.hasMax {
					assert.NotNil(t, goal.FieldType.Max)
				}
			})
		}
	})

	t.Run("complex configuration fixture", func(t *testing.T) {
		fixturePath := filepath.Join("..", "..", "..", "testdata", "goals", "complex_configuration.yml")

		schema, err := goalParser.LoadFromFile(fixturePath)
		require.NoError(t, err)
		assert.NotNil(t, schema)

		// Should have 9 goals with mixed types
		assert.Len(t, schema.Goals, 9)

		// Count goal types
		simpleCount := 0
		elasticCount := 0
		informationalCount := 0

		for _, goal := range schema.Goals {
			switch goal.GoalType {
			case models.SimpleGoal:
				simpleCount++
			case models.ElasticGoal:
				elasticCount++
			case models.InformationalGoal:
				informationalCount++
			}
		}

		assert.Equal(t, 2, simpleCount, "Should have 2 simple goals")
		assert.Equal(t, 2, elasticCount, "Should have 2 elastic goals")
		assert.Equal(t, 5, informationalCount, "Should have 5 informational goals")

		// Verify elastic goals have all criteria
		for _, goal := range schema.Goals {
			if goal.GoalType == models.ElasticGoal && goal.ScoringType == models.AutomaticScoring {
				assert.NotNil(t, goal.MiniCriteria, "Elastic goal %s should have mini criteria", goal.Title)
				assert.NotNil(t, goal.MidiCriteria, "Elastic goal %s should have midi criteria", goal.Title)
				assert.NotNil(t, goal.MaxiCriteria, "Elastic goal %s should have maxi criteria", goal.Title)
			}
		}

		// Verify all goals have help text
		for _, goal := range schema.Goals {
			assert.NotEmpty(t, goal.HelpText, "Goal %s should have help text", goal.Title)
		}
	})
}

func TestInvalidYAMLHandling(t *testing.T) {
	goalParser := parser.NewGoalParser()

	t.Run("invalid goals fixture fails validation", func(t *testing.T) {
		fixturePath := filepath.Join("..", "..", "..", "testdata", "goals", "invalid_goals.yml")

		// The file should parse (YAML is syntactically valid) but fail validation
		schema, err := goalParser.LoadFromFile(fixturePath)

		// Check if parsing fails due to malformed YAML or validation fails
		if err != nil {
			// YAML parsing failed (expected due to malformed syntax at end)
			assert.Contains(t, err.Error(), "YAML")
		} else {
			// YAML parsed but should fail validation
			require.NotNil(t, schema)
			err = schema.Validate()
			assert.Error(t, err, "Invalid schema should fail validation")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		fixturePath := filepath.Join("..", "..", "..", "testdata", "goals", "non_existent.yml")

		_, err := goalParser.LoadFromFile(fixturePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestYAMLRoundtripConsistency(t *testing.T) {
	goalParser := parser.NewGoalParser()

	fixtures := []string{
		"valid_simple_goal.yml",
		"valid_informational_goals.yml",
		"complex_configuration.yml",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			fixturePath := filepath.Join("..", "..", "..", "testdata", "goals", fixture)

			// Load original schema
			originalSchema, err := goalParser.LoadFromFile(fixturePath)
			require.NoError(t, err)

			// Convert to YAML
			yamlData, err := goalParser.ToYAML(originalSchema)
			require.NoError(t, err)
			assert.NotEmpty(t, yamlData)

			// Parse the generated YAML
			reparsedSchema, err := goalParser.ParseYAML([]byte(yamlData))
			require.NoError(t, err)

			// Verify consistency
			assert.Equal(t, originalSchema.Version, reparsedSchema.Version)
			assert.Equal(t, originalSchema.CreatedDate, reparsedSchema.CreatedDate)
			assert.Len(t, reparsedSchema.Goals, len(originalSchema.Goals))

			// Verify each goal's key properties
			for i, originalGoal := range originalSchema.Goals {
				reparsedGoal := reparsedSchema.Goals[i]

				assert.Equal(t, originalGoal.Title, reparsedGoal.Title)
				assert.Equal(t, originalGoal.ID, reparsedGoal.ID)
				assert.Equal(t, originalGoal.GoalType, reparsedGoal.GoalType)
				assert.Equal(t, originalGoal.FieldType.Type, reparsedGoal.FieldType.Type)
				assert.Equal(t, originalGoal.ScoringType, reparsedGoal.ScoringType)

				// Direction should be preserved for informational goals
				if originalGoal.GoalType == models.InformationalGoal {
					assert.Equal(t, originalGoal.Direction, reparsedGoal.Direction)
				}
			}
		})
	}
}

func TestFieldTypeValidationAcrossFixtures(t *testing.T) {
	goalParser := parser.NewGoalParser()

	t.Run("all field types represented", func(t *testing.T) {
		fixturePath := filepath.Join("..", "..", "..", "testdata", "goals", "valid_informational_goals.yml")

		schema, err := goalParser.LoadFromFile(fixturePath)
		require.NoError(t, err)

		// Collect all field types used
		fieldTypes := make(map[string]bool)
		for _, goal := range schema.Goals {
			fieldTypes[goal.FieldType.Type] = true
		}

		// Verify all major field types are represented
		expectedTypes := []string{
			models.BooleanFieldType,
			models.TextFieldType,
			models.UnsignedIntFieldType,
			models.UnsignedDecimalFieldType,
			models.DecimalFieldType,
			models.TimeFieldType,
			models.DurationFieldType,
		}

		for _, expectedType := range expectedTypes {
			assert.True(t, fieldTypes[expectedType], "Field type %s should be represented in fixtures", expectedType)
		}
	})

	t.Run("numeric constraints properly set", func(t *testing.T) {
		fixturePath := filepath.Join("..", "..", "..", "testdata", "goals", "valid_informational_goals.yml")

		schema, err := goalParser.LoadFromFile(fixturePath)
		require.NoError(t, err)

		// Find goals with constraints
		constrainedGoals := 0
		for _, goal := range schema.Goals {
			if goal.FieldType.Min != nil || goal.FieldType.Max != nil {
				constrainedGoals++

				// If both min and max are set, min should be less than max
				if goal.FieldType.Min != nil && goal.FieldType.Max != nil {
					assert.Less(t, *goal.FieldType.Min, *goal.FieldType.Max,
						"Goal %s: min should be less than max", goal.Title)
				}
			}
		}

		assert.Greater(t, constrainedGoals, 0, "Should have at least one goal with numeric constraints")
	})
}
