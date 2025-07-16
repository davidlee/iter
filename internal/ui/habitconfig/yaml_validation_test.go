package habitconfig

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
)

func TestYAMLFixtureValidation(t *testing.T) {
	goalParser := parser.NewHabitParser()

	t.Run("valid simple habit fixture", func(t *testing.T) {
		fixturePath := filepath.Join("..", "..", "..", "testdata", "habits", "valid_simple_habit.yml")

		schema, err := goalParser.LoadFromFile(fixturePath)
		require.NoError(t, err)
		assert.NotNil(t, schema)

		// Verify schema structure
		assert.Equal(t, "1.0.0", schema.Version)
		assert.Equal(t, "2024-01-01", schema.CreatedDate)
		assert.Len(t, schema.Habits, 2)

		// Verify first habit (manual scoring)
		goal1 := schema.Habits[0]
		assert.Equal(t, "Daily Exercise", goal1.Title)
		assert.Equal(t, "daily_exercise", goal1.ID)
		assert.Equal(t, models.SimpleHabit, goal1.HabitType)
		assert.Equal(t, models.BooleanFieldType, goal1.FieldType.Type)
		assert.Equal(t, models.ManualScoring, goal1.ScoringType)
		assert.Equal(t, "Did you exercise today?", goal1.Prompt)
		assert.Nil(t, goal1.Criteria)

		// Verify second habit (automatic scoring)
		goal2 := schema.Habits[1]
		assert.Equal(t, "Read for 30 Minutes", goal2.Title)
		assert.Equal(t, "daily_reading", goal2.ID)
		assert.Equal(t, models.AutomaticScoring, goal2.ScoringType)
		require.NotNil(t, goal2.Criteria)
		assert.Equal(t, "Reading completed", goal2.Criteria.Description)
		require.NotNil(t, goal2.Criteria.Condition.Equals)
		assert.Equal(t, true, *goal2.Criteria.Condition.Equals)
	})

	t.Run("valid informational habits fixture", func(t *testing.T) {
		fixturePath := filepath.Join("..", "..", "..", "testdata", "habits", "valid_informational_habits.yml")

		schema, err := goalParser.LoadFromFile(fixturePath)
		require.NoError(t, err)
		assert.NotNil(t, schema)

		// Should have 8 informational habits with different field types
		assert.Len(t, schema.Habits, 8)

		// Verify all habits are informational type
		for _, habit := range schema.Habits {
			assert.Equal(t, models.InformationalHabit, habit.HabitType)
			assert.Equal(t, models.ManualScoring, habit.ScoringType)
			assert.NotEmpty(t, habit.Direction)
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
				habit := schema.Habits[test.goalIndex]
				assert.Equal(t, test.expectedType, habit.FieldType.Type)
				assert.Equal(t, test.direction, habit.Direction)

				if test.hasUnit {
					assert.NotEmpty(t, habit.FieldType.Unit)
				}
				if test.hasMultiline {
					assert.NotNil(t, habit.FieldType.Multiline)
				}
				if test.hasMin {
					assert.NotNil(t, habit.FieldType.Min)
				}
				if test.hasMax {
					assert.NotNil(t, habit.FieldType.Max)
				}
			})
		}
	})

	t.Run("complex configuration fixture", func(t *testing.T) {
		fixturePath := filepath.Join("..", "..", "..", "testdata", "habits", "complex_configuration.yml")

		schema, err := goalParser.LoadFromFile(fixturePath)
		require.NoError(t, err)
		assert.NotNil(t, schema)

		// Should have 9 habits with mixed types
		assert.Len(t, schema.Habits, 9)

		// Count habit types
		simpleCount := 0
		elasticCount := 0
		informationalCount := 0

		for _, habit := range schema.Habits {
			switch habit.HabitType {
			case models.SimpleHabit:
				simpleCount++
			case models.ElasticHabit:
				elasticCount++
			case models.InformationalHabit:
				informationalCount++
			}
		}

		assert.Equal(t, 2, simpleCount, "Should have 2 simple habits")
		assert.Equal(t, 2, elasticCount, "Should have 2 elastic habits")
		assert.Equal(t, 5, informationalCount, "Should have 5 informational habits")

		// Verify elastic habits have all criteria
		for _, habit := range schema.Habits {
			if habit.HabitType == models.ElasticHabit && habit.ScoringType == models.AutomaticScoring {
				assert.NotNil(t, habit.MiniCriteria, "Elastic habit %s should have mini criteria", habit.Title)
				assert.NotNil(t, habit.MidiCriteria, "Elastic habit %s should have midi criteria", habit.Title)
				assert.NotNil(t, habit.MaxiCriteria, "Elastic habit %s should have maxi criteria", habit.Title)
			}
		}

		// Verify all habits have help text
		for _, habit := range schema.Habits {
			assert.NotEmpty(t, habit.HelpText, "Habit %s should have help text", habit.Title)
		}
	})
}

func TestInvalidYAMLHandling(t *testing.T) {
	goalParser := parser.NewHabitParser()

	t.Run("invalid habits fixture fails validation", func(t *testing.T) {
		fixturePath := filepath.Join("..", "..", "..", "testdata", "habits", "invalid_habits.yml")

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
		fixturePath := filepath.Join("..", "..", "..", "testdata", "habits", "non_existent.yml")

		_, err := goalParser.LoadFromFile(fixturePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestYAMLRoundtripConsistency(t *testing.T) {
	goalParser := parser.NewHabitParser()

	fixtures := []string{
		"valid_simple_habit.yml",
		"valid_informational_habits.yml",
		"complex_configuration.yml",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			fixturePath := filepath.Join("..", "..", "..", "testdata", "habits", fixture)

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
			assert.Len(t, reparsedSchema.Habits, len(originalSchema.Habits))

			// Verify each habit's key properties
			for i, originalHabit := range originalSchema.Habits {
				reparsedHabit := reparsedSchema.Habits[i]

				assert.Equal(t, originalHabit.Title, reparsedHabit.Title)
				assert.Equal(t, originalHabit.ID, reparsedHabit.ID)
				assert.Equal(t, originalHabit.HabitType, reparsedHabit.HabitType)
				assert.Equal(t, originalHabit.FieldType.Type, reparsedHabit.FieldType.Type)
				assert.Equal(t, originalHabit.ScoringType, reparsedHabit.ScoringType)

				// Direction should be preserved for informational habits
				if originalHabit.HabitType == models.InformationalHabit {
					assert.Equal(t, originalHabit.Direction, reparsedHabit.Direction)
				}
			}
		})
	}
}

func TestFieldTypeValidationAcrossFixtures(t *testing.T) {
	goalParser := parser.NewHabitParser()

	t.Run("all field types represented", func(t *testing.T) {
		fixturePath := filepath.Join("..", "..", "..", "testdata", "habits", "valid_informational_habits.yml")

		schema, err := goalParser.LoadFromFile(fixturePath)
		require.NoError(t, err)

		// Collect all field types used
		fieldTypes := make(map[string]bool)
		for _, habit := range schema.Habits {
			fieldTypes[habit.FieldType.Type] = true
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
		fixturePath := filepath.Join("..", "..", "..", "testdata", "habits", "valid_informational_habits.yml")

		schema, err := goalParser.LoadFromFile(fixturePath)
		require.NoError(t, err)

		// Find habits with constraints
		constrainedHabits := 0
		for _, habit := range schema.Habits {
			if habit.FieldType.Min != nil || habit.FieldType.Max != nil {
				constrainedHabits++

				// If both min and max are set, min should be less than max
				if habit.FieldType.Min != nil && habit.FieldType.Max != nil {
					assert.Less(t, *habit.FieldType.Min, *habit.FieldType.Max,
						"Habit %s: min should be less than max", habit.Title)
				}
			}
		}

		assert.Greater(t, constrainedHabits, 0, "Should have at least one habit with numeric constraints")
	})
}
