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

	// Test flow adjustment for different field types (includes potential criteria step)
	creator.selectedFieldType = models.BooleanFieldType
	creator.adjustFlowForFieldType()
	assert.Equal(t, 4, creator.maxSteps) // Field type, scoring, criteria, prompt

	creator.selectedFieldType = models.TextFieldType
	creator.adjustFlowForFieldType()
	assert.Equal(t, 5, creator.maxSteps) // Field type, field config, scoring, criteria, prompt

	creator.selectedFieldType = "numeric"
	creator.adjustFlowForFieldType()
	assert.Equal(t, 5, creator.maxSteps) // Field type, field config, scoring, criteria, prompt
}

func TestSimpleGoalCreator_ScoringTypeFlowAdjustment(t *testing.T) {
	creator := NewSimpleGoalCreator("Test", "Test", models.SimpleGoal)
	creator.selectedFieldType = models.BooleanFieldType
	creator.adjustFlowForFieldType()
	initialSteps := creator.maxSteps

	// Manual scoring should reduce steps by 1 (no criteria step needed)
	creator.scoringType = models.ManualScoring
	creator.adjustFlowForScoringType()
	assert.Equal(t, initialSteps-1, creator.maxSteps)

	// Reset and test automatic scoring (no change)
	creator.adjustFlowForFieldType() // Reset to initial
	creator.scoringType = models.AutomaticScoring
	creator.adjustFlowForScoringType()
	assert.Equal(t, initialSteps, creator.maxSteps) // No change for automatic
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

func TestSimpleGoalCreator_AutomaticCriteria_Boolean(t *testing.T) {
	creator := NewSimpleGoalCreator("Exercise", "Daily exercise", models.SimpleGoal)
	creator.selectedFieldType = models.BooleanFieldType
	creator.scoringType = models.AutomaticScoring
	creator.criteriaType = "equals"
	creator.criteriaValue = "true"
	creator.prompt = "Did you exercise today?"

	goal, err := creator.createGoalFromData()
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.Equal(t, models.AutomaticScoring, goal.ScoringType)
	require.NotNil(t, goal.Criteria)
	assert.Equal(t, "Goal is complete when checked as true", goal.Criteria.Description)
	require.NotNil(t, goal.Criteria.Condition)
	require.NotNil(t, goal.Criteria.Condition.Equals)
	assert.True(t, *goal.Criteria.Condition.Equals)
}

func TestSimpleGoalCreator_AutomaticCriteria_NumericGreaterThan(t *testing.T) {
	creator := NewSimpleGoalCreator("Push-ups", "Daily push-ups", models.SimpleGoal)
	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.UnsignedIntFieldType
	creator.unit = "reps"
	creator.scoringType = models.AutomaticScoring
	creator.criteriaType = "greater_than_or_equal"
	creator.criteriaValue = "30"
	creator.prompt = "How many push-ups did you do?"

	goal, err := creator.createGoalFromData()
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.Equal(t, models.AutomaticScoring, goal.ScoringType)
	require.NotNil(t, goal.Criteria)
	assert.Equal(t, "Goal achieved when value >= 30.0 reps", goal.Criteria.Description)
	require.NotNil(t, goal.Criteria.Condition)
	require.NotNil(t, goal.Criteria.Condition.GreaterThanOrEqual)
	assert.Equal(t, 30.0, *goal.Criteria.Condition.GreaterThanOrEqual)
}

func TestSimpleGoalCreator_AutomaticCriteria_NumericRange(t *testing.T) {
	creator := NewSimpleGoalCreator("Sleep", "Sleep duration", models.SimpleGoal)
	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.UnsignedDecimalFieldType
	creator.unit = "hours"
	creator.scoringType = models.AutomaticScoring
	creator.criteriaType = "range"
	creator.criteriaValue = "7"
	creator.criteriaValue2 = "9"
	creator.rangeInclusive = true
	creator.prompt = "How many hours did you sleep?"

	goal, err := creator.createGoalFromData()
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.Equal(t, models.AutomaticScoring, goal.ScoringType)
	require.NotNil(t, goal.Criteria)
	assert.Equal(t, "Goal achieved when value is within 7.0 to 9.0 hours (inclusive)", goal.Criteria.Description)
	require.NotNil(t, goal.Criteria.Condition)
	require.NotNil(t, goal.Criteria.Condition.Range)
	assert.Equal(t, 7.0, goal.Criteria.Condition.Range.Min)
	assert.Equal(t, 9.0, goal.Criteria.Condition.Range.Max)
	require.NotNil(t, goal.Criteria.Condition.Range.MinInclusive)
	assert.True(t, *goal.Criteria.Condition.Range.MinInclusive)
}

func TestSimpleGoalCreator_AutomaticCriteria_Time(t *testing.T) {
	creator := NewSimpleGoalCreator("Wake Up", "Early wake up", models.SimpleGoal)
	creator.selectedFieldType = models.TimeFieldType
	creator.scoringType = models.AutomaticScoring
	creator.criteriaType = "before"
	creator.criteriaTimeValue = "07:00"
	creator.prompt = "What time did you wake up?"

	goal, err := creator.createGoalFromData()
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.Equal(t, models.AutomaticScoring, goal.ScoringType)
	require.NotNil(t, goal.Criteria)
	assert.Equal(t, "Goal achieved when time is before 07:00", goal.Criteria.Description)
	require.NotNil(t, goal.Criteria.Condition)
	assert.Equal(t, "07:00", goal.Criteria.Condition.Before)
}

func TestSimpleGoalCreator_AutomaticCriteria_Duration(t *testing.T) {
	creator := NewSimpleGoalCreator("Meditation", "Daily meditation", models.SimpleGoal)
	creator.selectedFieldType = models.DurationFieldType
	creator.scoringType = models.AutomaticScoring
	creator.criteriaType = "greater_than_or_equal"
	creator.criteriaValue = "20m"
	creator.prompt = "How long did you meditate?"

	goal, err := creator.createGoalFromData()
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.Equal(t, models.AutomaticScoring, goal.ScoringType)
	require.NotNil(t, goal.Criteria)
	assert.Equal(t, "Goal achieved when duration >= 20m", goal.Criteria.Description)
	require.NotNil(t, goal.Criteria.Condition)
	assert.Equal(t, "20m", goal.Criteria.Condition.After)
}

func TestSimpleGoalCreator_CriteriaBuilding_InvalidValues(t *testing.T) {
	creator := NewSimpleGoalCreator("Test", "Test", models.SimpleGoal)
	creator.selectedFieldType = "numeric"
	creator.scoringType = models.AutomaticScoring
	creator.criteriaType = "greater_than"
	creator.criteriaValue = "not_a_number"

	_, err := creator.buildCriteriaFromData()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid criteria value")
}

func TestSimpleGoalCreator_CriteriaBuilding_UnsupportedFieldType(t *testing.T) {
	creator := NewSimpleGoalCreator("Test", "Test", models.SimpleGoal)
	creator.selectedFieldType = "unsupported_type"
	creator.scoringType = models.AutomaticScoring

	_, err := creator.buildCriteriaFromData()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "automatic scoring not supported for field type")
}