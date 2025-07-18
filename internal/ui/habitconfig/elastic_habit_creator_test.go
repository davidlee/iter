// Package habitconfig provides UI components for interactive habit configuration.
package habitconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidlee/vice/internal/models"
)

func TestElasticHabitCreator_NewCreator(t *testing.T) {
	creator := NewElasticHabitCreator("Test Habit", "Test Description", models.ElasticHabit)

	assert.Equal(t, "Test Habit", creator.title)
	assert.Equal(t, "Test Description", creator.description)
	assert.Equal(t, models.ElasticHabit, creator.habitType)
	assert.Equal(t, models.TextFieldType, creator.selectedFieldType) // Default first option, boolean excluded for elastic
	assert.Equal(t, 0, creator.currentStep)
	assert.NotNil(t, creator.form)
}

func TestElasticHabitCreator_FieldTypeSupport(t *testing.T) {
	creator := NewElasticHabitCreator("Test", "Test", models.ElasticHabit)

	// Test field configuration detection (similar to SimpleHabitCreator)
	creator.selectedFieldType = models.TextFieldType
	assert.True(t, creator.needsFieldConfiguration())

	creator.selectedFieldType = "numeric"
	assert.True(t, creator.needsFieldConfiguration())

	creator.selectedFieldType = models.TimeFieldType
	assert.False(t, creator.needsFieldConfiguration())

	creator.selectedFieldType = models.DurationFieldType
	assert.False(t, creator.needsFieldConfiguration())
}

func TestElasticHabitCreator_AutomaticScoringSupport(t *testing.T) {
	creator := NewElasticHabitCreator("Test", "Test", models.ElasticHabit)

	// Test automatic scoring support by field type (text excluded for elastic habits)
	creator.selectedFieldType = models.TextFieldType
	assert.False(t, creator.supportsAutomaticScoring()) // Text restricted to manual

	creator.selectedFieldType = "numeric"
	assert.True(t, creator.supportsAutomaticScoring())

	creator.selectedFieldType = models.TimeFieldType
	assert.True(t, creator.supportsAutomaticScoring())

	creator.selectedFieldType = models.DurationFieldType
	assert.True(t, creator.supportsAutomaticScoring())
}

func TestElasticHabitCreator_FlowAdjustment(t *testing.T) {
	creator := NewElasticHabitCreator("Test", "Test", models.ElasticHabit)

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

func TestElasticHabitCreator_ScoringTypeFlowAdjustment(t *testing.T) {
	creator := NewElasticHabitCreator("Test", "Test", models.ElasticHabit)
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

func TestElasticHabitCreator_FieldTypeResolution(t *testing.T) {
	creator := NewElasticHabitCreator("Test", "Test", models.ElasticHabit)

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

func TestElasticHabitCreator_StateManagement(t *testing.T) {
	creator := NewElasticHabitCreator("Test", "Test", models.ElasticHabit)

	// Test initial state
	assert.False(t, creator.IsCompleted())
	assert.False(t, creator.IsCancelled())

	// Test state tracking
	assert.NotNil(t, creator.form)
	assert.Equal(t, 0, creator.currentStep)
}

func TestElasticHabitCreator_HabitCreation_Text_Manual(t *testing.T) {
	creator := NewElasticHabitCreator("Exercise Log", "Daily exercise notes", models.ElasticHabit)
	creator.selectedFieldType = models.TextFieldType
	creator.multilineText = true
	creator.scoringType = models.ManualScoring
	creator.prompt = "How was your exercise intensity today?"
	creator.comment = "Track mini/midi/maxi subjectively"

	habit, err := creator.createHabitFromData()
	require.NoError(t, err)
	require.NotNil(t, habit)

	assert.Equal(t, "Exercise Log", habit.Title)
	assert.Equal(t, "Daily exercise notes\n\nComment: Track mini/midi/maxi subjectively", habit.Description)
	assert.Equal(t, models.ElasticHabit, habit.HabitType)
	assert.Equal(t, models.TextFieldType, habit.FieldType.Type)
	assert.NotNil(t, habit.FieldType.Multiline)
	assert.True(t, *habit.FieldType.Multiline)
	assert.Equal(t, models.ManualScoring, habit.ScoringType)
	assert.Equal(t, "How was your exercise intensity today?", habit.Prompt)
	assert.Nil(t, habit.Criteria) // Manual scoring, no single criteria
	assert.Nil(t, habit.MiniCriteria)
	assert.Nil(t, habit.MidiCriteria)
	assert.Nil(t, habit.MaxiCriteria)
}

func TestElasticHabitCreator_HabitCreation_Numeric_Manual(t *testing.T) {
	creator := NewElasticHabitCreator("Exercise Minutes", "Exercise duration tracking", models.ElasticHabit)
	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.UnsignedIntFieldType
	creator.unit = "minutes"
	creator.hasMinMax = true
	creator.minValue = "0"
	creator.maxValue = "180"
	creator.scoringType = models.ManualScoring
	creator.prompt = "How many minutes did you exercise?"

	habit, err := creator.createHabitFromData()
	require.NoError(t, err)
	require.NotNil(t, habit)

	assert.Equal(t, "Exercise Minutes", habit.Title)
	assert.Equal(t, models.ElasticHabit, habit.HabitType)
	assert.Equal(t, models.UnsignedIntFieldType, habit.FieldType.Type)
	assert.Equal(t, "minutes", habit.FieldType.Unit)
	assert.NotNil(t, habit.FieldType.Min)
	assert.Equal(t, 0.0, *habit.FieldType.Min)
	assert.NotNil(t, habit.FieldType.Max)
	assert.Equal(t, 180.0, *habit.FieldType.Max)
	assert.Equal(t, models.ManualScoring, habit.ScoringType)
	assert.Equal(t, "How many minutes did you exercise?", habit.Prompt)
}

func TestElasticHabitCreator_ThreeTierCriteria_Numeric(t *testing.T) {
	creator := NewElasticHabitCreator("Exercise", "Daily exercise habit", models.ElasticHabit)
	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.UnsignedIntFieldType
	creator.unit = "minutes"
	creator.scoringType = models.AutomaticScoring
	creator.prompt = "How many minutes did you exercise?"

	// Set three-tier criteria (mini: 15min, midi: 30min, maxi: 60min)
	creator.miniCriteriaValue = "15"
	creator.midiCriteriaValue = "30"
	creator.maxiCriteriaValue = "60"

	habit, err := creator.createHabitFromData()
	require.NoError(t, err)
	require.NotNil(t, habit)

	assert.Equal(t, models.AutomaticScoring, habit.ScoringType)

	// Validate mini criteria
	require.NotNil(t, habit.MiniCriteria)
	assert.Equal(t, "Mini achievement when value >= 15.0 minutes", habit.MiniCriteria.Description)
	require.NotNil(t, habit.MiniCriteria.Condition)
	require.NotNil(t, habit.MiniCriteria.Condition.GreaterThanOrEqual)
	assert.Equal(t, 15.0, *habit.MiniCriteria.Condition.GreaterThanOrEqual)

	// Validate midi criteria
	require.NotNil(t, habit.MidiCriteria)
	assert.Equal(t, "Midi achievement when value >= 30.0 minutes", habit.MidiCriteria.Description)
	require.NotNil(t, habit.MidiCriteria.Condition)
	require.NotNil(t, habit.MidiCriteria.Condition.GreaterThanOrEqual)
	assert.Equal(t, 30.0, *habit.MidiCriteria.Condition.GreaterThanOrEqual)

	// Validate maxi criteria
	require.NotNil(t, habit.MaxiCriteria)
	assert.Equal(t, "Maxi achievement when value >= 60.0 minutes", habit.MaxiCriteria.Description)
	require.NotNil(t, habit.MaxiCriteria.Condition)
	require.NotNil(t, habit.MaxiCriteria.Condition.GreaterThanOrEqual)
	assert.Equal(t, 60.0, *habit.MaxiCriteria.Condition.GreaterThanOrEqual)
}

func TestElasticHabitCreator_ThreeTierCriteria_Time(t *testing.T) {
	creator := NewElasticHabitCreator("Wake Up", "Early wake up habit", models.ElasticHabit)
	creator.selectedFieldType = models.TimeFieldType
	creator.scoringType = models.AutomaticScoring
	creator.prompt = "What time did you wake up?"

	// Set three-tier time criteria (earlier = better achievement)
	creator.miniCriteriaTimeValue = "08:00" // Mini: before 8am
	creator.midiCriteriaTimeValue = "07:00" // Midi: before 7am
	creator.maxiCriteriaTimeValue = "06:00" // Maxi: before 6am

	habit, err := creator.createHabitFromData()
	require.NoError(t, err)
	require.NotNil(t, habit)

	assert.Equal(t, models.AutomaticScoring, habit.ScoringType)

	// Validate mini criteria
	require.NotNil(t, habit.MiniCriteria)
	assert.Equal(t, "Mini achievement when time is before 08:00", habit.MiniCriteria.Description)
	require.NotNil(t, habit.MiniCriteria.Condition)
	assert.Equal(t, "08:00", habit.MiniCriteria.Condition.Before)

	// Validate midi criteria
	require.NotNil(t, habit.MidiCriteria)
	assert.Equal(t, "Midi achievement when time is before 07:00", habit.MidiCriteria.Description)
	require.NotNil(t, habit.MidiCriteria.Condition)
	assert.Equal(t, "07:00", habit.MidiCriteria.Condition.Before)

	// Validate maxi criteria
	require.NotNil(t, habit.MaxiCriteria)
	assert.Equal(t, "Maxi achievement when time is before 06:00", habit.MaxiCriteria.Description)
	require.NotNil(t, habit.MaxiCriteria.Condition)
	assert.Equal(t, "06:00", habit.MaxiCriteria.Condition.Before)
}

func TestElasticHabitCreator_ThreeTierCriteria_Duration(t *testing.T) {
	creator := NewElasticHabitCreator("Meditation", "Daily meditation habit", models.ElasticHabit)
	creator.selectedFieldType = models.DurationFieldType
	creator.scoringType = models.AutomaticScoring
	creator.prompt = "How long did you meditate?"

	// Set three-tier duration criteria (longer = better achievement)
	creator.miniCriteriaValue = "10m" // Mini: 10+ minutes
	creator.midiCriteriaValue = "20m" // Midi: 20+ minutes
	creator.maxiCriteriaValue = "30m" // Maxi: 30+ minutes

	habit, err := creator.createHabitFromData()
	require.NoError(t, err)
	require.NotNil(t, habit)

	assert.Equal(t, models.AutomaticScoring, habit.ScoringType)

	// Validate mini criteria
	require.NotNil(t, habit.MiniCriteria)
	assert.Equal(t, "Mini achievement when duration >= 10m", habit.MiniCriteria.Description)
	require.NotNil(t, habit.MiniCriteria.Condition)
	assert.Equal(t, "10m", habit.MiniCriteria.Condition.After)

	// Validate midi criteria
	require.NotNil(t, habit.MidiCriteria)
	assert.Equal(t, "Midi achievement when duration >= 20m", habit.MidiCriteria.Description)
	require.NotNil(t, habit.MidiCriteria.Condition)
	assert.Equal(t, "20m", habit.MidiCriteria.Condition.After)

	// Validate maxi criteria
	require.NotNil(t, habit.MaxiCriteria)
	assert.Equal(t, "Maxi achievement when duration >= 30m", habit.MaxiCriteria.Description)
	require.NotNil(t, habit.MaxiCriteria.Condition)
	assert.Equal(t, "30m", habit.MaxiCriteria.Condition.After)
}

func TestElasticHabitCreator_CriteriaBuilding_InvalidValues(t *testing.T) {
	creator := NewElasticHabitCreator("Test", "Test", models.ElasticHabit)
	creator.selectedFieldType = "numeric"
	creator.scoringType = models.AutomaticScoring
	creator.miniCriteriaValue = "not_a_number"

	_, err := creator.buildCriteriaFromData("mini")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mini criteria value")
}

func TestElasticHabitCreator_CriteriaBuilding_UnsupportedFieldType(t *testing.T) {
	creator := NewElasticHabitCreator("Test", "Test", models.ElasticHabit)
	creator.selectedFieldType = "unsupported_type"
	creator.scoringType = models.AutomaticScoring

	_, err := creator.buildCriteriaFromData("mini")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "automatic scoring not supported for field type")
}

func TestElasticHabitCreator_CriteriaBuilding_UnknownTier(t *testing.T) {
	creator := NewElasticHabitCreator("Test", "Test", models.ElasticHabit)
	creator.selectedFieldType = models.TimeFieldType
	creator.scoringType = models.AutomaticScoring

	_, err := creator.buildCriteriaFromData("unknown_tier")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tier: unknown_tier")
}

func TestElasticHabitCreator_ValidationTimeInput(t *testing.T) {
	creator := NewElasticHabitCreator("Test", "Test", models.ElasticHabit)

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

func TestElasticHabitCreator_ValidationDurationInput(t *testing.T) {
	creator := NewElasticHabitCreator("Test", "Test", models.ElasticHabit)

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
