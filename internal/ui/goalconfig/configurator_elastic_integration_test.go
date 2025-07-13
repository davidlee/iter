package goalconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/iter/internal/models"
)

// TestGoalConfigurator_ElasticGoalIntegration tests the integration between configurator and ElasticGoalCreator
func TestGoalConfigurator_ElasticGoalIntegration(t *testing.T) {
	// Test that configurator properly routes to ElasticGoalCreator
	configurator := NewGoalConfigurator()

	// Create basic info for an elastic goal
	basicInfo := &BasicInfo{
		Title:       "Test Elastic Goal",
		Description: "Test elastic goal integration",
		GoalType:    models.ElasticGoal,
	}

	// Mock empty goals list
	existingGoals := []models.Goal{}

	// Call runElasticGoalCreator with test data to verify it works
	// We can't test the full bubbletea interaction, but we can test that the method exists
	// and has the correct signature by attempting to call it
	creator := NewElasticGoalCreator(basicInfo.Title, basicInfo.Description, basicInfo.GoalType)
	assert.NotNil(t, creator, "ElasticGoalCreator should be created successfully")

	// Test that the creator is properly initialized
	assert.Equal(t, basicInfo.Title, creator.title)
	assert.Equal(t, basicInfo.Description, creator.description)
	assert.Equal(t, models.ElasticGoal, creator.goalType)

	// Verify routing method exists and has correct signature by checking it compiles
	// The actual UI interaction can't be tested without TTY, but the method should exist
	_, err := configurator.runElasticGoalCreator(basicInfo, existingGoals)
	// This will fail due to TTY requirement, but that's expected - we just want to verify the method exists
	assert.Error(t, err, "Should error due to TTY requirement, but method should exist")
	// In non-TTY environments, bubbletea programs fail to execute due to TTY requirement
	assert.Contains(t, err.Error(), "elastic goal creator execution failed", "Should be the expected error from bubbletea execution")
}

// TestGoalConfigurator_ElasticGoalCreatorCreation tests that NewElasticGoalCreator works correctly
func TestGoalConfigurator_ElasticGoalCreatorCreation(t *testing.T) {
	// Test basic creation
	creator := NewElasticGoalCreator("Test Goal", "Test Description", models.ElasticGoal)
	require.NotNil(t, creator)

	// Verify initial state
	assert.Equal(t, "Test Goal", creator.title)
	assert.Equal(t, "Test Description", creator.description)
	assert.Equal(t, models.ElasticGoal, creator.goalType)
	assert.Equal(t, 0, creator.currentStep)
	assert.NotNil(t, creator.form)

	// Verify state management methods exist
	assert.False(t, creator.IsCompleted())
	assert.False(t, creator.IsCancelled())
}

// TestGoalConfigurator_ElasticGoalHeadlessIntegration tests headless integration with configurator patterns
func TestGoalConfigurator_ElasticGoalHeadlessIntegration(t *testing.T) {
	// Test that ElasticGoalCreator can be used headlessly (like in integration tests)
	testData := TestElasticGoalData{
		FieldType:         models.TextFieldType,
		ScoringType:       models.ManualScoring,
		MultilineText:     true,
		Prompt:            "How was your exercise intensity today?",
		Comment:           "Track mini/midi/maxi achievement levels",
	}

	creator := NewElasticGoalCreatorForTesting("Exercise Intensity", "Track exercise achievement levels", models.ElasticGoal, testData)
	require.NotNil(t, creator)

	// Create goal directly (bypassing UI)
	goal, err := creator.CreateGoalDirectly()
	require.NoError(t, err)
	require.NotNil(t, goal)

	// Verify the goal is properly structured for elastic type
	assert.Equal(t, "Exercise Intensity", goal.Title)
	assert.Equal(t, models.ElasticGoal, goal.GoalType)
	assert.Equal(t, models.TextFieldType, goal.FieldType.Type)
	assert.Equal(t, models.ManualScoring, goal.ScoringType)

	// For manual scoring, elastic goals should not have criteria
	assert.Nil(t, goal.MiniCriteria)
	assert.Nil(t, goal.MidiCriteria)
	assert.Nil(t, goal.MaxiCriteria)

	// Validate goal passes schema validation
	err = goal.Validate()
	assert.NoError(t, err, "Generated elastic goal should pass validation")
}

// TestGoalConfigurator_ElasticGoalWithCriteria tests elastic goal with automatic scoring
func TestGoalConfigurator_ElasticGoalWithCriteria(t *testing.T) {
	// Test elastic goal with three-tier criteria
	testData := TestElasticGoalData{
		FieldType:         "numeric",
		NumericSubtype:    models.UnsignedIntFieldType,
		Unit:              "minutes",
		ScoringType:       models.AutomaticScoring,
		Prompt:            "How many minutes did you exercise?",
		MiniCriteriaValue: "15",
		MidiCriteriaValue: "30",
		MaxiCriteriaValue: "60",
	}

	creator := NewElasticGoalCreatorForTesting("Exercise Duration", "Track exercise minutes", models.ElasticGoal, testData)
	require.NotNil(t, creator)

	// Create goal directly
	goal, err := creator.CreateGoalDirectly()
	require.NoError(t, err)
	require.NotNil(t, goal)

	// Verify three-tier criteria are present
	require.NotNil(t, goal.MiniCriteria, "Elastic goal should have mini criteria")
	require.NotNil(t, goal.MidiCriteria, "Elastic goal should have midi criteria")
	require.NotNil(t, goal.MaxiCriteria, "Elastic goal should have maxi criteria")

	// Verify criteria values
	assert.Equal(t, 15.0, *goal.MiniCriteria.Condition.GreaterThanOrEqual)
	assert.Equal(t, 30.0, *goal.MidiCriteria.Condition.GreaterThanOrEqual)
	assert.Equal(t, 60.0, *goal.MaxiCriteria.Condition.GreaterThanOrEqual)

	// Validate goal passes schema validation
	err = goal.Validate()
	assert.NoError(t, err, "Generated elastic goal with criteria should pass validation")
}