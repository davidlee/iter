package goalconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/iter/internal/models"
)

func TestSimpleGoalCreator_NewCreator(t *testing.T) {
	creator := NewSimpleGoalCreator("Test Goal", "Test Description", models.SimpleGoal)

	assert.Equal(t, "Test Goal", creator.title)
	assert.Equal(t, "Test Description", creator.description)
	assert.Equal(t, models.SimpleGoal, creator.goalType)
	assert.Equal(t, models.BooleanFieldType, creator.selectedFieldType) // Default
	assert.Equal(t, 0, creator.currentStep)
	assert.NotNil(t, creator.form)
}

func TestSimpleGoalCreator_FieldTypeSupport(t *testing.T) {
	creator := NewSimpleGoalCreator("Test", "Test", models.SimpleGoal)

	// Test field configuration detection
	creator.selectedFieldType = models.BooleanFieldType
	assert.False(t, creator.needsFieldConfiguration())

	creator.selectedFieldType = models.TextFieldType
	assert.True(t, creator.needsFieldConfiguration())

	creator.selectedFieldType = "numeric"
	assert.True(t, creator.needsFieldConfiguration())

	creator.selectedFieldType = models.TimeFieldType
	assert.False(t, creator.needsFieldConfiguration())

	creator.selectedFieldType = models.DurationFieldType
	assert.False(t, creator.needsFieldConfiguration())
}

func TestSimpleGoalCreator_AutomaticScoringSupport(t *testing.T) {
	creator := NewSimpleGoalCreator("Test", "Test", models.SimpleGoal)

	// Test automatic scoring support by field type
	creator.selectedFieldType = models.BooleanFieldType
	assert.True(t, creator.supportsAutomaticScoring())

	creator.selectedFieldType = models.TextFieldType
	assert.False(t, creator.supportsAutomaticScoring()) // Text restricted to manual

	creator.selectedFieldType = "numeric"
	assert.True(t, creator.supportsAutomaticScoring())

	creator.selectedFieldType = models.TimeFieldType
	assert.True(t, creator.supportsAutomaticScoring())

	creator.selectedFieldType = models.DurationFieldType
	assert.True(t, creator.supportsAutomaticScoring())
}

func TestSimpleGoalCreator_FlowAdjustment(t *testing.T) {
	creator := NewSimpleGoalCreator("Test", "Test", models.SimpleGoal)

	// Test flow adjustment for different field types
	creator.selectedFieldType = models.BooleanFieldType
	creator.adjustFlowForFieldType()
	assert.Equal(t, 3, creator.maxSteps) // Field type, scoring, prompt

	creator.selectedFieldType = models.TextFieldType
	creator.adjustFlowForFieldType()
	assert.Equal(t, 4, creator.maxSteps) // Field type, field config, scoring, prompt

	creator.selectedFieldType = "numeric"
	creator.adjustFlowForFieldType()
	assert.Equal(t, 4, creator.maxSteps) // Field type, field config, scoring, prompt
}

func TestSimpleGoalCreator_FieldTypeResolution(t *testing.T) {
	creator := NewSimpleGoalCreator("Test", "Test", models.SimpleGoal)

	// Test field type resolution
	creator.selectedFieldType = models.BooleanFieldType
	assert.Equal(t, models.BooleanFieldType, creator.getResolvedFieldType())

	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.UnsignedIntFieldType
	assert.Equal(t, models.UnsignedIntFieldType, creator.getResolvedFieldType())

	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.DecimalFieldType
	assert.Equal(t, models.DecimalFieldType, creator.getResolvedFieldType())
}

func TestSimpleGoalCreator_GoalCreation_Boolean(t *testing.T) {
	creator := NewSimpleGoalCreator("Exercise", "Daily exercise goal", models.SimpleGoal)
	creator.selectedFieldType = models.BooleanFieldType
	creator.scoringType = models.ManualScoring
	creator.prompt = "Did you exercise today?"
	creator.comment = "Track daily exercise"

	goal, err := creator.createGoalFromData()
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.Equal(t, "Exercise", goal.Title)
	assert.Equal(t, "Daily exercise goal\n\nComment: Track daily exercise", goal.Description)
	assert.Equal(t, models.SimpleGoal, goal.GoalType)
	assert.Equal(t, models.BooleanFieldType, goal.FieldType.Type)
	assert.Equal(t, models.ManualScoring, goal.ScoringType)
	assert.Equal(t, "Did you exercise today?", goal.Prompt)
}

func TestSimpleGoalCreator_GoalCreation_Numeric(t *testing.T) {
	creator := NewSimpleGoalCreator("Push-ups", "Daily push-ups", models.SimpleGoal)
	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.UnsignedIntFieldType
	creator.unit = "reps"
	creator.hasMinMax = true
	creator.minValue = "10"
	creator.maxValue = "100"
	creator.scoringType = models.ManualScoring
	creator.prompt = "How many push-ups did you do?"

	goal, err := creator.createGoalFromData()
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.Equal(t, "Push-ups", goal.Title)
	assert.Equal(t, models.SimpleGoal, goal.GoalType)
	assert.Equal(t, models.UnsignedIntFieldType, goal.FieldType.Type)
	assert.Equal(t, "reps", goal.FieldType.Unit)
	assert.NotNil(t, goal.FieldType.Min)
	assert.Equal(t, 10.0, *goal.FieldType.Min)
	assert.NotNil(t, goal.FieldType.Max)
	assert.Equal(t, 100.0, *goal.FieldType.Max)
	assert.Equal(t, models.ManualScoring, goal.ScoringType)
	assert.Equal(t, "How many push-ups did you do?", goal.Prompt)
}

func TestSimpleGoalCreator_GoalCreation_Text(t *testing.T) {
	creator := NewSimpleGoalCreator("Journal", "Daily journaling", models.SimpleGoal)
	creator.selectedFieldType = models.TextFieldType
	creator.multilineText = true
	creator.scoringType = models.ManualScoring
	creator.prompt = "What did you write about today?"

	goal, err := creator.createGoalFromData()
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.Equal(t, "Journal", goal.Title)
	assert.Equal(t, models.SimpleGoal, goal.GoalType)
	assert.Equal(t, models.TextFieldType, goal.FieldType.Type)
	assert.NotNil(t, goal.FieldType.Multiline)
	assert.True(t, *goal.FieldType.Multiline)
	assert.Equal(t, models.ManualScoring, goal.ScoringType)
	assert.Equal(t, "What did you write about today?", goal.Prompt)
}

func TestSimpleGoalCreator_StateManagement(t *testing.T) {
	creator := NewSimpleGoalCreator("Test", "Test", models.SimpleGoal)

	// Test initial state
	assert.False(t, creator.IsCompleted())
	assert.False(t, creator.IsCancelled())

	// Test completion tracking
	assert.NotNil(t, creator.form)
	assert.Equal(t, 0, creator.currentStep)
}