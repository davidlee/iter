package goalconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	initpkg "davidlee/vice/internal/init"
	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
)

// TestGoalCreationWorkflows tests goal creation workflows without UI interaction
func TestSimpleGoalCreation(t *testing.T) {
	tempDir := t.TempDir()
	goalsFile := filepath.Join(tempDir, "goals.yml")
	entriesFile := filepath.Join(tempDir, "entries.yml")

	// Initialize empty files
	initializer := initpkg.NewFileInitializer()
	err := initializer.EnsureConfigFiles(goalsFile, entriesFile)
	require.NoError(t, err)

	// Create configurator
	_ = NewGoalConfigurator() // configurator for future use

	t.Run("simple goal with manual scoring", func(t *testing.T) {
		// Create a simple goal programmatically (simulating UI input)
		basicInfo := &BasicInfo{
			Title:       "Daily Exercise",
			Description: "Get some physical activity every day",
			GoalType:    models.SimpleGoal,
		}

		goal := &models.Goal{
			Title:       basicInfo.Title,
			Description: basicInfo.Description,
			GoalType:    basicInfo.GoalType,
			FieldType: models.FieldType{
				Type: models.BooleanFieldType,
			},
			ScoringType: models.ManualScoring,
			Prompt:      "Did you exercise today?",
		}

		// Save the goal using configurator's internal logic
		goalParser := parser.NewGoalParser()
		schema, err := goalParser.LoadFromFile(goalsFile)
		require.NoError(t, err)

		// Clear existing goals and add our test goal
		schema.Goals = []models.Goal{*goal}

		// Save updated schema
		err = goalParser.SaveToFile(schema, goalsFile)
		require.NoError(t, err)

		// Verify goal was saved correctly
		reloadedSchema, err := goalParser.LoadFromFile(goalsFile)
		require.NoError(t, err)
		require.Len(t, reloadedSchema.Goals, 1)

		savedGoal := reloadedSchema.Goals[0]
		assert.Equal(t, "Daily Exercise", savedGoal.Title)
		assert.Equal(t, "Get some physical activity every day", savedGoal.Description)
		assert.Equal(t, models.SimpleGoal, savedGoal.GoalType)
		assert.Equal(t, models.BooleanFieldType, savedGoal.FieldType.Type)
		assert.Equal(t, models.ManualScoring, savedGoal.ScoringType)
		assert.Equal(t, "Did you exercise today?", savedGoal.Prompt)
		assert.NotEmpty(t, savedGoal.ID) // ID should be generated
	})

	t.Run("simple goal with automatic scoring", func(t *testing.T) {
		goal := &models.Goal{
			Title:       "Read Daily",
			Description: "Read for at least 30 minutes",
			GoalType:    models.SimpleGoal,
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

		// Save the goal
		goalParser := parser.NewGoalParser()
		schema, err := goalParser.LoadFromFile(goalsFile)
		require.NoError(t, err)

		// Clear existing goals and add our test goal
		schema.Goals = []models.Goal{*goal}
		err = goalParser.SaveToFile(schema, goalsFile)
		require.NoError(t, err)

		// Verify goal with criteria was saved correctly
		reloadedSchema, err := goalParser.LoadFromFile(goalsFile)
		require.NoError(t, err)
		require.Len(t, reloadedSchema.Goals, 1)

		savedGoal := reloadedSchema.Goals[0]
		assert.Equal(t, models.AutomaticScoring, savedGoal.ScoringType)
		require.NotNil(t, savedGoal.Criteria)
		assert.Equal(t, "Reading completed", savedGoal.Criteria.Description)
		require.NotNil(t, savedGoal.Criteria.Condition.Equals)
		assert.Equal(t, true, *savedGoal.Criteria.Condition.Equals)
	})
}

func TestInformationalGoalCreation(t *testing.T) {
	tempDir := t.TempDir()
	goalsFile := filepath.Join(tempDir, "goals.yml")
	entriesFile := filepath.Join(tempDir, "entries.yml")

	// Initialize empty files
	initializer := initpkg.NewFileInitializer()
	err := initializer.EnsureConfigFiles(goalsFile, entriesFile)
	require.NoError(t, err)

	goalParser := parser.NewGoalParser()

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
			// Create informational goal
			goal := &models.Goal{
				Title:       tc.expectedTitle,
				Description: "Test goal for " + tc.fieldType,
				GoalType:    models.InformationalGoal,
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

			// Save the goal
			schema, err := goalParser.LoadFromFile(goalsFile)
			require.NoError(t, err)

			// Clear existing goals and add our test goal
			schema.Goals = []models.Goal{*goal}
			err = goalParser.SaveToFile(schema, goalsFile)
			require.NoError(t, err)

			// Verify goal was saved correctly
			reloadedSchema, err := goalParser.LoadFromFile(goalsFile)
			require.NoError(t, err)
			require.Len(t, reloadedSchema.Goals, 1)

			savedGoal := reloadedSchema.Goals[0]
			assert.Equal(t, tc.expectedTitle, savedGoal.Title)
			assert.Equal(t, models.InformationalGoal, savedGoal.GoalType)
			assert.Equal(t, tc.fieldType, savedGoal.FieldType.Type)
			assert.Equal(t, models.ManualScoring, savedGoal.ScoringType)
			assert.Equal(t, tc.direction, savedGoal.Direction)

			// Check field-specific configurations
			if tc.unit != "" {
				assert.Equal(t, tc.unit, savedGoal.FieldType.Unit)
			}
			if tc.min != nil {
				require.NotNil(t, savedGoal.FieldType.Min)
				assert.Equal(t, *tc.min, *savedGoal.FieldType.Min)
			}
			if tc.max != nil {
				require.NotNil(t, savedGoal.FieldType.Max)
				assert.Equal(t, *tc.max, *savedGoal.FieldType.Max)
			}
			if tc.multiline != nil {
				require.NotNil(t, savedGoal.FieldType.Multiline)
				assert.Equal(t, *tc.multiline, *savedGoal.FieldType.Multiline)
			}

			assert.NotEmpty(t, savedGoal.ID) // ID should be generated
		})
	}
}

func TestGoalValidationWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	goalsFile := filepath.Join(tempDir, "goals.yml")
	entriesFile := filepath.Join(tempDir, "entries.yml")

	// Initialize empty files
	initializer := initpkg.NewFileInitializer()
	err := initializer.EnsureConfigFiles(goalsFile, entriesFile)
	require.NoError(t, err)

	goalParser := parser.NewGoalParser()

	t.Run("valid goal passes validation", func(t *testing.T) {
		goal := &models.Goal{
			Title:       "Valid Goal",
			Description: "This is a valid goal",
			GoalType:    models.SimpleGoal,
			FieldType: models.FieldType{
				Type: models.BooleanFieldType,
			},
			ScoringType: models.ManualScoring,
			Prompt:      "Did you complete this goal?",
		}

		// Test validation directly
		err := goal.Validate()
		assert.NoError(t, err)

		// Test validation through schema
		schema, err := goalParser.LoadFromFile(goalsFile)
		require.NoError(t, err)

		// Clear existing goals and add our test goal
		schema.Goals = []models.Goal{*goal}
		err = schema.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid goal fails validation", func(t *testing.T) {
		// Missing title
		invalidGoal := &models.Goal{
			Description: "Goal without title",
			GoalType:    models.SimpleGoal,
			FieldType: models.FieldType{
				Type: models.BooleanFieldType,
			},
			ScoringType: models.ManualScoring,
		}

		err := invalidGoal.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title")
	})

	t.Run("numeric constraints validation", func(t *testing.T) {
		// Invalid: min > max
		minVal := 10.0
		maxVal := 5.0
		invalidGoal := &models.Goal{
			Title:       "Invalid Numeric Goal",
			Description: "Min greater than max",
			GoalType:    models.InformationalGoal,
			FieldType: models.FieldType{
				Type: models.UnsignedDecimalFieldType,
				Min:  &minVal,
				Max:  &maxVal,
			},
			ScoringType: models.ManualScoring,
		}

		err := invalidGoal.Validate()
		assert.Error(t, err)
	})
}

func TestYAMLGenerationWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	goalsFile := filepath.Join(tempDir, "goals.yml")
	entriesFile := filepath.Join(tempDir, "entries.yml")

	// Initialize empty files
	initializer := initpkg.NewFileInitializer()
	err := initializer.EnsureConfigFiles(goalsFile, entriesFile)
	require.NoError(t, err)

	goalParser := parser.NewGoalParser()
	_ = NewGoalConfigurator() // configurator for future use

	t.Run("dry-run mode generates valid YAML", func(t *testing.T) {
		// Create a test goal
		goal := &models.Goal{
			Title:       "Test Goal",
			Description: "Goal for YAML generation testing",
			GoalType:    models.SimpleGoal,
			FieldType: models.FieldType{
				Type: models.BooleanFieldType,
			},
			ScoringType: models.ManualScoring,
			Prompt:      "Did you complete this test goal?",
		}

		// Add goal to schema
		schema, err := goalParser.LoadFromFile(goalsFile)
		require.NoError(t, err)

		// Clear existing goals and add our test goal
		schema.Goals = []models.Goal{*goal}

		// Test YAML generation
		yamlOutput, err := goalParser.ToYAML(schema)
		require.NoError(t, err)
		assert.NotEmpty(t, yamlOutput)

		// Verify generated YAML is parseable
		parsedSchema, err := goalParser.ParseYAML([]byte(yamlOutput))
		require.NoError(t, err)
		require.Len(t, parsedSchema.Goals, 1)

		// Verify content matches original
		parsedGoal := parsedSchema.Goals[0]
		assert.Equal(t, goal.Title, parsedGoal.Title)
		assert.Equal(t, goal.Description, parsedGoal.Description)
		assert.Equal(t, goal.GoalType, parsedGoal.GoalType)
		assert.Equal(t, goal.FieldType.Type, parsedGoal.FieldType.Type)
		assert.Equal(t, goal.ScoringType, parsedGoal.ScoringType)
		assert.Equal(t, goal.Prompt, parsedGoal.Prompt)
	})

	t.Run("YAML output doesn't modify files", func(t *testing.T) {
		// Get initial file modification time
		initialStat, err := os.Stat(goalsFile)
		require.NoError(t, err)
		initialModTime := initialStat.ModTime()

		// Load schema
		schema, err := goalParser.LoadFromFile(goalsFile)
		require.NoError(t, err)

		// Generate YAML (simulating dry-run)
		_, err = goalParser.ToYAML(schema)
		require.NoError(t, err)

		// Verify file wasn't modified
		finalStat, err := os.Stat(goalsFile)
		require.NoError(t, err)
		finalModTime := finalStat.ModTime()

		assert.Equal(t, initialModTime, finalModTime, "File should not be modified during YAML generation")
	})
}
