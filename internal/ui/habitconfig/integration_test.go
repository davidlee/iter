package habitconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	initpkg "github.com/davidlee/vice/internal/init"
	"github.com/davidlee/vice/internal/models"
	"github.com/davidlee/vice/internal/parser"
)

// TestHabitCreationWorkflows tests habit creation workflows without UI interaction
func TestSimpleHabitCreation(t *testing.T) {
	tempDir := t.TempDir()
	habitsFile := filepath.Join(tempDir, "habits.yml")
	entriesFile := filepath.Join(tempDir, "entries.yml")

	// Initialize empty files
	initializer := initpkg.NewFileInitializer()
	err := initializer.EnsureConfigFiles(habitsFile, entriesFile)
	require.NoError(t, err)

	// Create configurator
	_ = NewHabitConfigurator() // configurator for future use

	t.Run("simple habit with manual scoring", func(t *testing.T) {
		// Create a simple habit programmatically (simulating UI input)
		basicInfo := &BasicInfo{
			Title:       "Daily Exercise",
			Description: "Get some physical activity every day",
			HabitType:   models.SimpleHabit,
		}

		habit := &models.Habit{
			Title:       basicInfo.Title,
			Description: basicInfo.Description,
			HabitType:   basicInfo.HabitType,
			FieldType: models.FieldType{
				Type: models.BooleanFieldType,
			},
			ScoringType: models.ManualScoring,
			Prompt:      "Did you exercise today?",
		}

		// Save the habit using configurator's internal logic
		habitParser := parser.NewHabitParser()
		schema, err := habitParser.LoadFromFile(habitsFile)
		require.NoError(t, err)

		// Clear existing habits and add our test habit
		schema.Habits = []models.Habit{*habit}

		// Save updated schema
		err = habitParser.SaveToFile(schema, habitsFile)
		require.NoError(t, err)

		// Verify habit was saved correctly
		reloadedSchema, err := habitParser.LoadFromFile(habitsFile)
		require.NoError(t, err)
		require.Len(t, reloadedSchema.Habits, 1)

		savedHabit := reloadedSchema.Habits[0]
		assert.Equal(t, "Daily Exercise", savedHabit.Title)
		assert.Equal(t, "Get some physical activity every day", savedHabit.Description)
		assert.Equal(t, models.SimpleHabit, savedHabit.HabitType)
		assert.Equal(t, models.BooleanFieldType, savedHabit.FieldType.Type)
		assert.Equal(t, models.ManualScoring, savedHabit.ScoringType)
		assert.Equal(t, "Did you exercise today?", savedHabit.Prompt)
		assert.NotEmpty(t, savedHabit.ID) // ID should be generated
	})

	t.Run("simple habit with automatic scoring", func(t *testing.T) {
		habit := &models.Habit{
			Title:       "Read Daily",
			Description: "Read for at least 30 minutes",
			HabitType:   models.SimpleHabit,
			FieldType: models.FieldType{
				Type: models.BooleanFieldType,
			},
			ScoringType: models.AutomaticScoring,
			Prompt:      "Did you read today?",
			Criteria: &models.Criteria{
				Description: "Reading completed",
				Condition: &models.Condition{
					Equals: &[]bool{true}[0],
				},
			},
		}

		// Save the habit
		habitParser := parser.NewHabitParser()
		schema, err := habitParser.LoadFromFile(habitsFile)
		require.NoError(t, err)

		// Clear existing habits and add our test habit
		schema.Habits = []models.Habit{*habit}
		err = habitParser.SaveToFile(schema, habitsFile)
		require.NoError(t, err)

		// Verify habit with criteria was saved correctly
		reloadedSchema, err := habitParser.LoadFromFile(habitsFile)
		require.NoError(t, err)
		require.Len(t, reloadedSchema.Habits, 1)

		savedHabit := reloadedSchema.Habits[0]
		assert.Equal(t, models.AutomaticScoring, savedHabit.ScoringType)
		require.NotNil(t, savedHabit.Criteria)
		assert.Equal(t, "Reading completed", savedHabit.Criteria.Description)
		require.NotNil(t, savedHabit.Criteria.Condition.Equals)
		assert.Equal(t, true, *savedHabit.Criteria.Condition.Equals)
	})
}

func TestInformationalHabitCreation(t *testing.T) {
	tempDir := t.TempDir()
	habitsFile := filepath.Join(tempDir, "habits.yml")
	entriesFile := filepath.Join(tempDir, "entries.yml")

	// Initialize empty files
	initializer := initpkg.NewFileInitializer()
	err := initializer.EnsureConfigFiles(habitsFile, entriesFile)
	require.NoError(t, err)

	habitParser := parser.NewHabitParser()

	testCases := []struct {
		name          string
		fieldType     string
		unit          string
		min           *float64
		max           *float64
		multiline     *bool
		direction     string
		expectedTitle string
	}{
		{
			name:          "boolean field",
			fieldType:     models.BooleanFieldType,
			direction:     "neutral",
			expectedTitle: "Mood Tracking",
		},
		{
			name:          "text field single line",
			fieldType:     models.TextFieldType,
			multiline:     &[]bool{false}[0],
			direction:     "neutral",
			expectedTitle: "Daily Reflection",
		},
		{
			name:          "text field multiline",
			fieldType:     models.TextFieldType,
			multiline:     &[]bool{true}[0],
			direction:     "neutral",
			expectedTitle: "Journal Entry",
		},
		{
			name:          "unsigned int with unit",
			fieldType:     models.UnsignedIntFieldType,
			unit:          "reps",
			direction:     "higher_better",
			expectedTitle: "Push-ups",
		},
		{
			name:          "unsigned decimal with constraints",
			fieldType:     models.UnsignedDecimalFieldType,
			unit:          "hours",
			min:           &[]float64{4.0}[0],
			max:           &[]float64{12.0}[0],
			direction:     "neutral",
			expectedTitle: "Sleep Hours",
		},
		{
			name:          "decimal with custom unit",
			fieldType:     models.DecimalFieldType,
			unit:          "liters",
			direction:     "higher_better",
			expectedTitle: "Water Intake",
		},
		{
			name:          "time field",
			fieldType:     models.TimeFieldType,
			direction:     "lower_better",
			expectedTitle: "Wake Up Time",
		},
		{
			name:          "duration field",
			fieldType:     models.DurationFieldType,
			direction:     "higher_better",
			expectedTitle: "Meditation Session",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create informational habit
			habit := &models.Habit{
				Title:       tc.expectedTitle,
				Description: "Test habit for " + tc.fieldType,
				HabitType:   models.InformationalHabit,
				FieldType: models.FieldType{
					Type:      tc.fieldType,
					Unit:      tc.unit,
					Min:       tc.min,
					Max:       tc.max,
					Multiline: tc.multiline,
				},
				ScoringType: models.ManualScoring,
				Direction:   tc.direction,
				Prompt:      "Test prompt for " + tc.fieldType,
			}

			// Save the habit
			schema, err := habitParser.LoadFromFile(habitsFile)
			require.NoError(t, err)

			// Clear existing habits and add our test habit
			schema.Habits = []models.Habit{*habit}
			err = habitParser.SaveToFile(schema, habitsFile)
			require.NoError(t, err)

			// Verify habit was saved correctly
			reloadedSchema, err := habitParser.LoadFromFile(habitsFile)
			require.NoError(t, err)
			require.Len(t, reloadedSchema.Habits, 1)

			savedHabit := reloadedSchema.Habits[0]
			assert.Equal(t, tc.expectedTitle, savedHabit.Title)
			assert.Equal(t, models.InformationalHabit, savedHabit.HabitType)
			assert.Equal(t, tc.fieldType, savedHabit.FieldType.Type)
			assert.Equal(t, models.ManualScoring, savedHabit.ScoringType)
			assert.Equal(t, tc.direction, savedHabit.Direction)

			// Check field-specific configurations
			if tc.unit != "" {
				assert.Equal(t, tc.unit, savedHabit.FieldType.Unit)
			}
			if tc.min != nil {
				require.NotNil(t, savedHabit.FieldType.Min)
				assert.Equal(t, *tc.min, *savedHabit.FieldType.Min)
			}
			if tc.max != nil {
				require.NotNil(t, savedHabit.FieldType.Max)
				assert.Equal(t, *tc.max, *savedHabit.FieldType.Max)
			}
			if tc.multiline != nil {
				require.NotNil(t, savedHabit.FieldType.Multiline)
				assert.Equal(t, *tc.multiline, *savedHabit.FieldType.Multiline)
			}

			assert.NotEmpty(t, savedHabit.ID) // ID should be generated
		})
	}
}

func TestHabitValidationWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	habitsFile := filepath.Join(tempDir, "habits.yml")
	entriesFile := filepath.Join(tempDir, "entries.yml")

	// Initialize empty files
	initializer := initpkg.NewFileInitializer()
	err := initializer.EnsureConfigFiles(habitsFile, entriesFile)
	require.NoError(t, err)

	habitParser := parser.NewHabitParser()

	t.Run("valid habit passes validation", func(t *testing.T) {
		habit := &models.Habit{
			Title:       "Valid Habit",
			Description: "This is a valid habit",
			HabitType:   models.SimpleHabit,
			FieldType: models.FieldType{
				Type: models.BooleanFieldType,
			},
			ScoringType: models.ManualScoring,
			Prompt:      "Did you complete this habit?",
		}

		// Test validation directly
		err := habit.Validate()
		assert.NoError(t, err)

		// Test validation through schema
		schema, err := habitParser.LoadFromFile(habitsFile)
		require.NoError(t, err)

		// Clear existing habits and add our test habit
		schema.Habits = []models.Habit{*habit}
		err = schema.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid habit fails validation", func(t *testing.T) {
		// Missing title
		invalidHabit := &models.Habit{
			Description: "Habit without title",
			HabitType:   models.SimpleHabit,
			FieldType: models.FieldType{
				Type: models.BooleanFieldType,
			},
			ScoringType: models.ManualScoring,
		}

		err := invalidHabit.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title")
	})

	t.Run("numeric constraints validation", func(t *testing.T) {
		// Invalid: min > max
		minVal := 10.0
		maxVal := 5.0
		invalidHabit := &models.Habit{
			Title:       "Invalid Numeric Habit",
			Description: "Min greater than max",
			HabitType:   models.InformationalHabit,
			FieldType: models.FieldType{
				Type: models.UnsignedDecimalFieldType,
				Min:  &minVal,
				Max:  &maxVal,
			},
			ScoringType: models.ManualScoring,
		}

		err := invalidHabit.Validate()
		assert.Error(t, err)
	})
}

func TestYAMLGenerationWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	habitsFile := filepath.Join(tempDir, "habits.yml")
	entriesFile := filepath.Join(tempDir, "entries.yml")

	// Initialize empty files
	initializer := initpkg.NewFileInitializer()
	err := initializer.EnsureConfigFiles(habitsFile, entriesFile)
	require.NoError(t, err)

	habitParser := parser.NewHabitParser()
	_ = NewHabitConfigurator() // configurator for future use

	t.Run("dry-run mode generates valid YAML", func(t *testing.T) {
		// Create a test habit
		habit := &models.Habit{
			Title:       "Test Habit",
			Description: "Habit for YAML generation testing",
			HabitType:   models.SimpleHabit,
			FieldType: models.FieldType{
				Type: models.BooleanFieldType,
			},
			ScoringType: models.ManualScoring,
			Prompt:      "Did you complete this test habit?",
		}

		// Add habit to schema
		schema, err := habitParser.LoadFromFile(habitsFile)
		require.NoError(t, err)

		// Clear existing habits and add our test habit
		schema.Habits = []models.Habit{*habit}

		// Test YAML generation
		yamlOutput, err := habitParser.ToYAML(schema)
		require.NoError(t, err)
		assert.NotEmpty(t, yamlOutput)

		// Verify generated YAML is parseable
		parsedSchema, err := habitParser.ParseYAML([]byte(yamlOutput))
		require.NoError(t, err)
		require.Len(t, parsedSchema.Habits, 1)

		// Verify content matches original
		parsedHabit := parsedSchema.Habits[0]
		assert.Equal(t, habit.Title, parsedHabit.Title)
		assert.Equal(t, habit.Description, parsedHabit.Description)
		assert.Equal(t, habit.HabitType, parsedHabit.HabitType)
		assert.Equal(t, habit.FieldType.Type, parsedHabit.FieldType.Type)
		assert.Equal(t, habit.ScoringType, parsedHabit.ScoringType)
		assert.Equal(t, habit.Prompt, parsedHabit.Prompt)
	})

	t.Run("YAML output doesn't modify files", func(t *testing.T) {
		// Get initial file modification time
		initialStat, err := os.Stat(habitsFile)
		require.NoError(t, err)
		initialModTime := initialStat.ModTime()

		// Load schema
		schema, err := habitParser.LoadFromFile(habitsFile)
		require.NoError(t, err)

		// Generate YAML (simulating dry-run)
		_, err = habitParser.ToYAML(schema)
		require.NoError(t, err)

		// Verify file wasn't modified
		finalStat, err := os.Stat(habitsFile)
		require.NoError(t, err)
		finalModTime := finalStat.ModTime()

		assert.Equal(t, initialModTime, finalModTime, "File should not be modified during YAML generation")
	})
}
