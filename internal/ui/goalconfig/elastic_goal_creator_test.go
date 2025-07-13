// Package goalconfig provides UI components for interactive goal configuration.
package goalconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/iter/internal/models"
)

func TestElasticGoalCreator_NewCreator(t *testing.T) {
	creator := NewElasticGoalCreator("Test Goal", "Test Description", models.ElasticGoal)

	assert.Equal(t, "Test Goal", creator.title)
	assert.Equal(t, "Test Description", creator.description)
	assert.Equal(t, models.ElasticGoal, creator.goalType)
	assert.Equal(t, models.TextFieldType, creator.selectedFieldType) // Default first option, boolean excluded for elastic
	assert.Equal(t, 0, creator.currentStep)
	assert.NotNil(t, creator.form)
}

func TestElasticGoalCreator_FieldTypeSupport(t *testing.T) {
	creator := NewElasticGoalCreator("Test", "Test", models.ElasticGoal)

	// Test field configuration detection (similar to SimpleGoalCreator)
	creator.selectedFieldType = models.TextFieldType
	assert.True(t, creator.needsFieldConfiguration())

	creator.selectedFieldType = "numeric"
	assert.True(t, creator.needsFieldConfiguration())

	creator.selectedFieldType = models.TimeFieldType
	assert.False(t, creator.needsFieldConfiguration())

	creator.selectedFieldType = models.DurationFieldType
	assert.False(t, creator.needsFieldConfiguration())
}

func TestElasticGoalCreator_AutomaticScoringSupport(t *testing.T) {
	creator := NewElasticGoalCreator("Test", "Test", models.ElasticGoal)

	// Test automatic scoring support by field type (text excluded for elastic goals)
	creator.selectedFieldType = models.TextFieldType
	assert.False(t, creator.supportsAutomaticScoring()) // Text restricted to manual

	creator.selectedFieldType = "numeric"
	assert.True(t, creator.supportsAutomaticScoring())

	creator.selectedFieldType = models.TimeFieldType
	assert.True(t, creator.supportsAutomaticScoring())

	creator.selectedFieldType = models.DurationFieldType
	assert.True(t, creator.supportsAutomaticScoring())
}

func TestElasticGoalCreator_FlowAdjustment(t *testing.T) {
	creator := NewElasticGoalCreator("Test", "Test", models.ElasticGoal)

	// Test flow adjustment for different field types (includes potential criteria step)
	creator.selectedFieldType = models.TextFieldType
	creator.adjustFlowForFieldType()
	assert.Equal(t, 5, creator.maxSteps) // Field type, field config, scoring, criteria, prompt

	creator.selectedFieldType = "numeric"
	creator.adjustFlowForFieldType()
	assert.Equal(t, 5, creator.maxSteps) // Field type, field config, scoring, criteria, prompt

	creator.selectedFieldType = models.TimeFieldType
	creator.adjustFlowForFieldType()
	assert.Equal(t, 4, creator.maxSteps) // Field type, scoring, criteria, prompt

	creator.selectedFieldType = models.DurationFieldType
	creator.adjustFlowForFieldType()
	assert.Equal(t, 4, creator.maxSteps) // Field type, scoring, criteria, prompt
}

func TestElasticGoalCreator_ScoringTypeFlowAdjustment(t *testing.T) {
	creator := NewElasticGoalCreator("Test", "Test", models.ElasticGoal)
	creator.selectedFieldType = models.TimeFieldType
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

func TestElasticGoalCreator_FieldTypeResolution(t *testing.T) {
	creator := NewElasticGoalCreator("Test", "Test", models.ElasticGoal)

	// Test field type resolution
	creator.selectedFieldType = models.TextFieldType
	assert.Equal(t, models.TextFieldType, creator.getResolvedFieldType())

	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.UnsignedIntFieldType
	assert.Equal(t, models.UnsignedIntFieldType, creator.getResolvedFieldType())

	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.DecimalFieldType
	assert.Equal(t, models.DecimalFieldType, creator.getResolvedFieldType())
}

func TestElasticGoalCreator_StateManagement(t *testing.T) {
	creator := NewElasticGoalCreator("Test", "Test", models.ElasticGoal)

	// Test initial state
	assert.False(t, creator.IsCompleted())
	assert.False(t, creator.IsCancelled())

	// Test state tracking
	assert.NotNil(t, creator.form)
	assert.Equal(t, 0, creator.currentStep)
}

func TestElasticGoalCreator_GoalCreation_Text_Manual(t *testing.T) {
	creator := NewElasticGoalCreator("Exercise Log", "Daily exercise notes", models.ElasticGoal)
	creator.selectedFieldType = models.TextFieldType
	creator.multilineText = true
	creator.scoringType = models.ManualScoring
	creator.prompt = "How was your exercise intensity today?"
	creator.comment = "Track mini/midi/maxi subjectively"

	goal, err := creator.createGoalFromData()
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.Equal(t, "Exercise Log", goal.Title)
	assert.Equal(t, "Daily exercise notes\n\nComment: Track mini/midi/maxi subjectively", goal.Description)
	assert.Equal(t, models.ElasticGoal, goal.GoalType)
	assert.Equal(t, models.TextFieldType, goal.FieldType.Type)
	assert.NotNil(t, goal.FieldType.Multiline)
	assert.True(t, *goal.FieldType.Multiline)
	assert.Equal(t, models.ManualScoring, goal.ScoringType)
	assert.Equal(t, "How was your exercise intensity today?", goal.Prompt)
	assert.Nil(t, goal.Criteria) // Manual scoring, no single criteria
	assert.Nil(t, goal.MiniCriteria)
	assert.Nil(t, goal.MidiCriteria)
	assert.Nil(t, goal.MaxiCriteria)
}

func TestElasticGoalCreator_GoalCreation_Numeric_Manual(t *testing.T) {
	creator := NewElasticGoalCreator("Exercise Minutes", "Exercise duration tracking", models.ElasticGoal)
	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.UnsignedIntFieldType
	creator.unit = "minutes"
	creator.hasMinMax = true
	creator.minValue = "0"
	creator.maxValue = "180"
	creator.scoringType = models.ManualScoring
	creator.prompt = "How many minutes did you exercise?"

	goal, err := creator.createGoalFromData()
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.Equal(t, "Exercise Minutes", goal.Title)
	assert.Equal(t, models.ElasticGoal, goal.GoalType)
	assert.Equal(t, models.UnsignedIntFieldType, goal.FieldType.Type)
	assert.Equal(t, "minutes", goal.FieldType.Unit)
	assert.NotNil(t, goal.FieldType.Min)
	assert.Equal(t, 0.0, *goal.FieldType.Min)
	assert.NotNil(t, goal.FieldType.Max)
	assert.Equal(t, 180.0, *goal.FieldType.Max)
	assert.Equal(t, models.ManualScoring, goal.ScoringType)
	assert.Equal(t, "How many minutes did you exercise?", goal.Prompt)
}

func TestElasticGoalCreator_ThreeTierCriteria_Numeric(t *testing.T) {
	creator := NewElasticGoalCreator("Exercise", "Daily exercise goal", models.ElasticGoal)
	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.UnsignedIntFieldType
	creator.unit = "minutes"
	creator.scoringType = models.AutomaticScoring
	creator.prompt = "How many minutes did you exercise?"

	// Set three-tier criteria (mini: 15min, midi: 30min, maxi: 60min)
	creator.miniCriteriaValue = "15"
	creator.midiCriteriaValue = "30"
	creator.maxiCriteriaValue = "60"

	goal, err := creator.createGoalFromData()
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.Equal(t, models.AutomaticScoring, goal.ScoringType)

	// Validate mini criteria
	require.NotNil(t, goal.MiniCriteria)
	assert.Equal(t, "Mini achievement when value >= 15.0 minutes", goal.MiniCriteria.Description)
	require.NotNil(t, goal.MiniCriteria.Condition)
	require.NotNil(t, goal.MiniCriteria.Condition.GreaterThanOrEqual)
	assert.Equal(t, 15.0, *goal.MiniCriteria.Condition.GreaterThanOrEqual)

	// Validate midi criteria
	require.NotNil(t, goal.MidiCriteria)
	assert.Equal(t, "Midi achievement when value >= 30.0 minutes", goal.MidiCriteria.Description)
	require.NotNil(t, goal.MidiCriteria.Condition)
	require.NotNil(t, goal.MidiCriteria.Condition.GreaterThanOrEqual)
	assert.Equal(t, 30.0, *goal.MidiCriteria.Condition.GreaterThanOrEqual)

	// Validate maxi criteria
	require.NotNil(t, goal.MaxiCriteria)
	assert.Equal(t, "Maxi achievement when value >= 60.0 minutes", goal.MaxiCriteria.Description)
	require.NotNil(t, goal.MaxiCriteria.Condition)
	require.NotNil(t, goal.MaxiCriteria.Condition.GreaterThanOrEqual)
	assert.Equal(t, 60.0, *goal.MaxiCriteria.Condition.GreaterThanOrEqual)
}

func TestElasticGoalCreator_ThreeTierCriteria_Time(t *testing.T) {
	creator := NewElasticGoalCreator("Wake Up", "Early wake up goal", models.ElasticGoal)
	creator.selectedFieldType = models.TimeFieldType
	creator.scoringType = models.AutomaticScoring
	creator.prompt = "What time did you wake up?"

	// Set three-tier time criteria (earlier = better achievement)
	creator.miniCriteriaTimeValue = "08:00" // Mini: before 8am
	creator.midiCriteriaTimeValue = "07:00" // Midi: before 7am
	creator.maxiCriteriaTimeValue = "06:00" // Maxi: before 6am

	goal, err := creator.createGoalFromData()
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.Equal(t, models.AutomaticScoring, goal.ScoringType)

	// Validate mini criteria
	require.NotNil(t, goal.MiniCriteria)
	assert.Equal(t, "Mini achievement when time is before 08:00", goal.MiniCriteria.Description)
	require.NotNil(t, goal.MiniCriteria.Condition)
	assert.Equal(t, "08:00", goal.MiniCriteria.Condition.Before)

	// Validate midi criteria
	require.NotNil(t, goal.MidiCriteria)
	assert.Equal(t, "Midi achievement when time is before 07:00", goal.MidiCriteria.Description)
	require.NotNil(t, goal.MidiCriteria.Condition)
	assert.Equal(t, "07:00", goal.MidiCriteria.Condition.Before)

	// Validate maxi criteria
	require.NotNil(t, goal.MaxiCriteria)
	assert.Equal(t, "Maxi achievement when time is before 06:00", goal.MaxiCriteria.Description)
	require.NotNil(t, goal.MaxiCriteria.Condition)
	assert.Equal(t, "06:00", goal.MaxiCriteria.Condition.Before)
}

func TestElasticGoalCreator_ThreeTierCriteria_Duration(t *testing.T) {
	creator := NewElasticGoalCreator("Meditation", "Daily meditation goal", models.ElasticGoal)
	creator.selectedFieldType = models.DurationFieldType
	creator.scoringType = models.AutomaticScoring
	creator.prompt = "How long did you meditate?"

	// Set three-tier duration criteria (longer = better achievement)
	creator.miniCriteriaValue = "10m" // Mini: 10+ minutes
	creator.midiCriteriaValue = "20m" // Midi: 20+ minutes
	creator.maxiCriteriaValue = "30m" // Maxi: 30+ minutes

	goal, err := creator.createGoalFromData()
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.Equal(t, models.AutomaticScoring, goal.ScoringType)

	// Validate mini criteria
	require.NotNil(t, goal.MiniCriteria)
	assert.Equal(t, "Mini achievement when duration >= 10m", goal.MiniCriteria.Description)
	require.NotNil(t, goal.MiniCriteria.Condition)
	assert.Equal(t, "10m", goal.MiniCriteria.Condition.After)

	// Validate midi criteria
	require.NotNil(t, goal.MidiCriteria)
	assert.Equal(t, "Midi achievement when duration >= 20m", goal.MidiCriteria.Description)
	require.NotNil(t, goal.MidiCriteria.Condition)
	assert.Equal(t, "20m", goal.MidiCriteria.Condition.After)

	// Validate maxi criteria
	require.NotNil(t, goal.MaxiCriteria)
	assert.Equal(t, "Maxi achievement when duration >= 30m", goal.MaxiCriteria.Description)
	require.NotNil(t, goal.MaxiCriteria.Condition)
	assert.Equal(t, "30m", goal.MaxiCriteria.Condition.After)
}

func TestElasticGoalCreator_CriteriaBuilding_InvalidValues(t *testing.T) {
	creator := NewElasticGoalCreator("Test", "Test", models.ElasticGoal)
	creator.selectedFieldType = "numeric"
	creator.scoringType = models.AutomaticScoring
	creator.miniCriteriaValue = "not_a_number"

	_, err := creator.buildCriteriaFromData("mini")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mini criteria value")
}

func TestElasticGoalCreator_CriteriaBuilding_UnsupportedFieldType(t *testing.T) {
	creator := NewElasticGoalCreator("Test", "Test", models.ElasticGoal)
	creator.selectedFieldType = "unsupported_type"
	creator.scoringType = models.AutomaticScoring

	_, err := creator.buildCriteriaFromData("mini")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "automatic scoring not supported for field type")
}

func TestElasticGoalCreator_CriteriaBuilding_UnknownTier(t *testing.T) {
	creator := NewElasticGoalCreator("Test", "Test", models.ElasticGoal)
	creator.selectedFieldType = models.TimeFieldType
	creator.scoringType = models.AutomaticScoring

	_, err := creator.buildCriteriaFromData("unknown_tier")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tier: unknown_tier")
}

func TestElasticGoalCreator_ValidationTimeInput(t *testing.T) {
	creator := NewElasticGoalCreator("Test", "Test", models.ElasticGoal)

	// Test valid time input
	err := creator.validateTimeInput("07:30")
	assert.NoError(t, err)

	// Test empty input
	err = creator.validateTimeInput("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "time value is required")

	// Test invalid format (no colon)
	err = creator.validateTimeInput("730")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HH:MM format")

	// Test invalid format (wrong parts)
	err = creator.validateTimeInput("7:30:45")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HH:MM format")
}

func TestElasticGoalCreator_ValidationDurationInput(t *testing.T) {
	creator := NewElasticGoalCreator("Test", "Test", models.ElasticGoal)

	// Test valid duration inputs
	err := creator.validateDurationInput("30m")
	assert.NoError(t, err)

	err = creator.validateDurationInput("1h")
	assert.NoError(t, err)

	err = creator.validateDurationInput("1h 30m")
	assert.NoError(t, err)

	// Test empty input
	err = creator.validateDurationInput("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duration value is required")

	// Test invalid format (no time units)
	err = creator.validateDurationInput("30")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duration must include time units")
}
