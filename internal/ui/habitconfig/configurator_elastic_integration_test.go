package habitconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidlee/vice/internal/models"
)

// TestHabitConfigurator_ElasticHabitIntegration tests the integration between configurator and ElasticHabitCreator
func TestHabitConfigurator_ElasticHabitIntegration(t *testing.T) {
	// Test that configurator properly routes to ElasticHabitCreator
	configurator := NewHabitConfigurator()
	assert.NotNil(t, configurator)

	// Test that ElasticHabitCreator can be created with proper basic info
	basicInfo := &BasicInfo{
		Title:       "Test Elastic Habit",
		Description: "Test elastic habit integration",
		HabitType:   models.ElasticHabit,
	}

	creator := NewElasticHabitCreator(basicInfo.Title, basicInfo.Description, basicInfo.HabitType)
	assert.NotNil(t, creator, "ElasticHabitCreator should be created successfully")

	// Test that the creator is properly initialized
	assert.Equal(t, basicInfo.Title, creator.title)
	assert.Equal(t, basicInfo.Description, creator.description)
	assert.Equal(t, models.ElasticHabit, creator.habitType)

	// Test headless habit creation (the proper way to test without TTY)
	testData := TestElasticHabitData{
		FieldType:     models.TextFieldType,
		ScoringType:   models.ManualScoring,
		MultilineText: true,
		Prompt:        "How was your test habit today?",
		Comment:       "Test elastic habit creation",
	}

	headlessCreator := NewElasticHabitCreatorForTesting(basicInfo.Title, basicInfo.Description, basicInfo.HabitType, testData)
	require.NotNil(t, headlessCreator)

	// Create habit directly (bypassing UI)
	habit, err := headlessCreator.CreateHabitDirectly()
	require.NoError(t, err)
	require.NotNil(t, habit)

	// Verify the habit is properly structured
	assert.Equal(t, basicInfo.Title, habit.Title)
	assert.Equal(t, basicInfo.Description+"\n\nComment: Test elastic habit creation", habit.Description)
	assert.Equal(t, models.ElasticHabit, habit.HabitType)
	assert.Equal(t, models.TextFieldType, habit.FieldType.Type)
	assert.Equal(t, models.ManualScoring, habit.ScoringType)

	// For manual scoring, elastic habits should not have criteria
	assert.Nil(t, habit.MiniCriteria)
	assert.Nil(t, habit.MidiCriteria)
	assert.Nil(t, habit.MaxiCriteria)

	// Validate habit passes schema validation
	err = habit.Validate()
	assert.NoError(t, err, "Generated elastic habit should pass validation")
}

// TestHabitConfigurator_ElasticHabitCreatorCreation tests that NewElasticHabitCreator works correctly
func TestHabitConfigurator_ElasticHabitCreatorCreation(t *testing.T) {
	// Test basic creation
	creator := NewElasticHabitCreator("Test Habit", "Test Description", models.ElasticHabit)
	require.NotNil(t, creator)

	// Verify initial state
	assert.Equal(t, "Test Habit", creator.title)
	assert.Equal(t, "Test Description", creator.description)
	assert.Equal(t, models.ElasticHabit, creator.habitType)
	assert.Equal(t, 0, creator.currentStep)
	assert.NotNil(t, creator.form)

	// Verify state management methods exist
	assert.False(t, creator.IsCompleted())
	assert.False(t, creator.IsCancelled())
}

// TestHabitConfigurator_ElasticHabitHeadlessIntegration tests headless integration with configurator patterns
func TestHabitConfigurator_ElasticHabitHeadlessIntegration(t *testing.T) {
	// Test that ElasticHabitCreator can be used headlessly (like in integration tests)
	testData := TestElasticHabitData{
		FieldType:     models.TextFieldType,
		ScoringType:   models.ManualScoring,
		MultilineText: true,
		Prompt:        "How was your exercise intensity today?",
		Comment:       "Track mini/midi/maxi achievement levels",
	}

	creator := NewElasticHabitCreatorForTesting("Exercise Intensity", "Track exercise achievement levels", models.ElasticHabit, testData)
	require.NotNil(t, creator)

	// Create habit directly (bypassing UI)
	habit, err := creator.CreateHabitDirectly()
	require.NoError(t, err)
	require.NotNil(t, habit)

	// Verify the habit is properly structured for elastic type
	assert.Equal(t, "Exercise Intensity", habit.Title)
	assert.Equal(t, models.ElasticHabit, habit.HabitType)
	assert.Equal(t, models.TextFieldType, habit.FieldType.Type)
	assert.Equal(t, models.ManualScoring, habit.ScoringType)

	// For manual scoring, elastic habits should not have criteria
	assert.Nil(t, habit.MiniCriteria)
	assert.Nil(t, habit.MidiCriteria)
	assert.Nil(t, habit.MaxiCriteria)

	// Validate habit passes schema validation
	err = habit.Validate()
	assert.NoError(t, err, "Generated elastic habit should pass validation")
}

// TestHabitConfigurator_ElasticHabitWithCriteria tests elastic habit with automatic scoring
func TestHabitConfigurator_ElasticHabitWithCriteria(t *testing.T) {
	// Test elastic habit with three-tier criteria
	testData := TestElasticHabitData{
		FieldType:         "numeric",
		NumericSubtype:    models.UnsignedIntFieldType,
		Unit:              "minutes",
		ScoringType:       models.AutomaticScoring,
		Prompt:            "How many minutes did you exercise?",
		MiniCriteriaValue: "15",
		MidiCriteriaValue: "30",
		MaxiCriteriaValue: "60",
	}

	creator := NewElasticHabitCreatorForTesting("Exercise Duration", "Track exercise minutes", models.ElasticHabit, testData)
	require.NotNil(t, creator)

	// Create habit directly
	habit, err := creator.CreateHabitDirectly()
	require.NoError(t, err)
	require.NotNil(t, habit)

	// Verify three-tier criteria are present
	require.NotNil(t, habit.MiniCriteria, "Elastic habit should have mini criteria")
	require.NotNil(t, habit.MidiCriteria, "Elastic habit should have midi criteria")
	require.NotNil(t, habit.MaxiCriteria, "Elastic habit should have maxi criteria")

	// Verify criteria values
	assert.Equal(t, 15.0, *habit.MiniCriteria.Condition.GreaterThanOrEqual)
	assert.Equal(t, 30.0, *habit.MidiCriteria.Condition.GreaterThanOrEqual)
	assert.Equal(t, 60.0, *habit.MaxiCriteria.Condition.GreaterThanOrEqual)

	// Validate habit passes schema validation
	err = habit.Validate()
	assert.NoError(t, err, "Generated elastic habit with criteria should pass validation")
}
