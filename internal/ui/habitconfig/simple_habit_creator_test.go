package habitconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidlee/vice/internal/models"
)

func TestSimpleHabitCreator_NewCreator(t *testing.T) {
	creator := NewSimpleHabitCreator("Test Habit", "Test Description", models.SimpleHabit)

	assert.Equal(t, "Test Habit", creator.title)
	assert.Equal(t, "Test Description", creator.description)
	assert.Equal(t, models.SimpleHabit, creator.habitType)
	assert.Equal(t, models.BooleanFieldType, creator.selectedFieldType) // Default
	assert.Equal(t, 0, creator.currentStep)
	assert.NotNil(t, creator.form)
}

func TestSimpleHabitCreator_FieldTypeSupport(t *testing.T) {
	creator := NewSimpleHabitCreator("Test", "Test", models.SimpleHabit)

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

func TestSimpleHabitCreator_AutomaticScoringSupport(t *testing.T) {
	creator := NewSimpleHabitCreator("Test", "Test", models.SimpleHabit)

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

func TestSimpleHabitCreator_FlowAdjustment(t *testing.T) {
	creator := NewSimpleHabitCreator("Test", "Test", models.SimpleHabit)

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

func TestSimpleHabitCreator_ScoringTypeFlowAdjustment(t *testing.T) {
	creator := NewSimpleHabitCreator("Test", "Test", models.SimpleHabit)
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

func TestSimpleHabitCreator_FieldTypeResolution(t *testing.T) {
	creator := NewSimpleHabitCreator("Test", "Test", models.SimpleHabit)

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

func TestSimpleHabitCreator_HabitCreation_Boolean(t *testing.T) {
	creator := NewSimpleHabitCreator("Exercise", "Daily exercise habit", models.SimpleHabit)
	creator.selectedFieldType = models.BooleanFieldType
	creator.scoringType = models.ManualScoring
	creator.prompt = "Did you exercise today?"
	creator.comment = "Track daily exercise"

	habit, err := creator.createHabitFromData()
	require.NoError(t, err)
	require.NotNil(t, habit)

	assert.Equal(t, "Exercise", habit.Title)
	assert.Equal(t, "Daily exercise habit\n\nComment: Track daily exercise", habit.Description)
	assert.Equal(t, models.SimpleHabit, habit.HabitType)
	assert.Equal(t, models.BooleanFieldType, habit.FieldType.Type)
	assert.Equal(t, models.ManualScoring, habit.ScoringType)
	assert.Equal(t, "Did you exercise today?", habit.Prompt)
}

func TestSimpleHabitCreator_HabitCreation_Numeric(t *testing.T) {
	creator := NewSimpleHabitCreator("Push-ups", "Daily push-ups", models.SimpleHabit)
	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.UnsignedIntFieldType
	creator.unit = "reps"
	creator.hasMinMax = true
	creator.minValue = "10"
	creator.maxValue = "100"
	creator.scoringType = models.ManualScoring
	creator.prompt = "How many push-ups did you do?"

	habit, err := creator.createHabitFromData()
	require.NoError(t, err)
	require.NotNil(t, habit)

	assert.Equal(t, "Push-ups", habit.Title)
	assert.Equal(t, models.SimpleHabit, habit.HabitType)
	assert.Equal(t, models.UnsignedIntFieldType, habit.FieldType.Type)
	assert.Equal(t, "reps", habit.FieldType.Unit)
	assert.NotNil(t, habit.FieldType.Min)
	assert.Equal(t, 10.0, *habit.FieldType.Min)
	assert.NotNil(t, habit.FieldType.Max)
	assert.Equal(t, 100.0, *habit.FieldType.Max)
	assert.Equal(t, models.ManualScoring, habit.ScoringType)
	assert.Equal(t, "How many push-ups did you do?", habit.Prompt)
}

func TestSimpleHabitCreator_HabitCreation_Text(t *testing.T) {
	creator := NewSimpleHabitCreator("Journal", "Daily journaling", models.SimpleHabit)
	creator.selectedFieldType = models.TextFieldType
	creator.multilineText = true
	creator.scoringType = models.ManualScoring
	creator.prompt = "What did you write about today?"

	habit, err := creator.createHabitFromData()
	require.NoError(t, err)
	require.NotNil(t, habit)

	assert.Equal(t, "Journal", habit.Title)
	assert.Equal(t, models.SimpleHabit, habit.HabitType)
	assert.Equal(t, models.TextFieldType, habit.FieldType.Type)
	assert.NotNil(t, habit.FieldType.Multiline)
	assert.True(t, *habit.FieldType.Multiline)
	assert.Equal(t, models.ManualScoring, habit.ScoringType)
	assert.Equal(t, "What did you write about today?", habit.Prompt)
}

func TestSimpleHabitCreator_StateManagement(t *testing.T) {
	creator := NewSimpleHabitCreator("Test", "Test", models.SimpleHabit)

	// Test initial state
	assert.False(t, creator.IsCompleted())
	assert.False(t, creator.IsCancelled())

	// Test completion tracking
	assert.NotNil(t, creator.form)
	assert.Equal(t, 0, creator.currentStep)
}

func TestSimpleHabitCreator_AutomaticCriteria_Boolean(t *testing.T) {
	creator := NewSimpleHabitCreator("Exercise", "Daily exercise", models.SimpleHabit)
	creator.selectedFieldType = models.BooleanFieldType
	creator.scoringType = models.AutomaticScoring
	creator.criteriaType = "equals"
	creator.criteriaValue = "true"
	creator.prompt = "Did you exercise today?"

	habit, err := creator.createHabitFromData()
	require.NoError(t, err)
	require.NotNil(t, habit)

	assert.Equal(t, models.AutomaticScoring, habit.ScoringType)
	require.NotNil(t, habit.Criteria)
	assert.Equal(t, "Habit is complete when checked as true", habit.Criteria.Description)
	require.NotNil(t, habit.Criteria.Condition)
	require.NotNil(t, habit.Criteria.Condition.Equals)
	assert.True(t, *habit.Criteria.Condition.Equals)
}

func TestSimpleHabitCreator_AutomaticCriteria_NumericGreaterThan(t *testing.T) {
	creator := NewSimpleHabitCreator("Push-ups", "Daily push-ups", models.SimpleHabit)
	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.UnsignedIntFieldType
	creator.unit = "reps"
	creator.scoringType = models.AutomaticScoring
	creator.criteriaType = "greater_than_or_equal"
	creator.criteriaValue = "30"
	creator.prompt = "How many push-ups did you do?"

	habit, err := creator.createHabitFromData()
	require.NoError(t, err)
	require.NotNil(t, habit)

	assert.Equal(t, models.AutomaticScoring, habit.ScoringType)
	require.NotNil(t, habit.Criteria)
	assert.Equal(t, "Habit achieved when value >= 30.0 reps", habit.Criteria.Description)
	require.NotNil(t, habit.Criteria.Condition)
	require.NotNil(t, habit.Criteria.Condition.GreaterThanOrEqual)
	assert.Equal(t, 30.0, *habit.Criteria.Condition.GreaterThanOrEqual)
}

func TestSimpleHabitCreator_AutomaticCriteria_NumericRange(t *testing.T) {
	creator := NewSimpleHabitCreator("Sleep", "Sleep duration", models.SimpleHabit)
	creator.selectedFieldType = "numeric"
	creator.numericSubtype = models.UnsignedDecimalFieldType
	creator.unit = "hours"
	creator.scoringType = models.AutomaticScoring
	creator.criteriaType = "range"
	creator.criteriaValue = "7"
	creator.criteriaValue2 = "9"
	creator.rangeInclusive = true
	creator.prompt = "How many hours did you sleep?"

	habit, err := creator.createHabitFromData()
	require.NoError(t, err)
	require.NotNil(t, habit)

	assert.Equal(t, models.AutomaticScoring, habit.ScoringType)
	require.NotNil(t, habit.Criteria)
	assert.Equal(t, "Habit achieved when value is within 7.0 to 9.0 hours (inclusive)", habit.Criteria.Description)
	require.NotNil(t, habit.Criteria.Condition)
	require.NotNil(t, habit.Criteria.Condition.Range)
	assert.Equal(t, 7.0, habit.Criteria.Condition.Range.Min)
	assert.Equal(t, 9.0, habit.Criteria.Condition.Range.Max)
	require.NotNil(t, habit.Criteria.Condition.Range.MinInclusive)
	assert.True(t, *habit.Criteria.Condition.Range.MinInclusive)
}

func TestSimpleHabitCreator_AutomaticCriteria_Time(t *testing.T) {
	creator := NewSimpleHabitCreator("Wake Up", "Early wake up", models.SimpleHabit)
	creator.selectedFieldType = models.TimeFieldType
	creator.scoringType = models.AutomaticScoring
	creator.criteriaType = "before"
	creator.criteriaTimeValue = "07:00"
	creator.prompt = "What time did you wake up?"

	habit, err := creator.createHabitFromData()
	require.NoError(t, err)
	require.NotNil(t, habit)

	assert.Equal(t, models.AutomaticScoring, habit.ScoringType)
	require.NotNil(t, habit.Criteria)
	assert.Equal(t, "Habit achieved when time is before 07:00", habit.Criteria.Description)
	require.NotNil(t, habit.Criteria.Condition)
	assert.Equal(t, "07:00", habit.Criteria.Condition.Before)
}

func TestSimpleHabitCreator_AutomaticCriteria_Duration(t *testing.T) {
	creator := NewSimpleHabitCreator("Meditation", "Daily meditation", models.SimpleHabit)
	creator.selectedFieldType = models.DurationFieldType
	creator.scoringType = models.AutomaticScoring
	creator.criteriaType = "greater_than_or_equal"
	creator.criteriaValue = "20m"
	creator.prompt = "How long did you meditate?"

	habit, err := creator.createHabitFromData()
	require.NoError(t, err)
	require.NotNil(t, habit)

	assert.Equal(t, models.AutomaticScoring, habit.ScoringType)
	require.NotNil(t, habit.Criteria)
	assert.Equal(t, "Habit achieved when duration >= 20m", habit.Criteria.Description)
	require.NotNil(t, habit.Criteria.Condition)
	assert.Equal(t, "20m", habit.Criteria.Condition.After)
}

func TestSimpleHabitCreator_CriteriaBuilding_InvalidValues(t *testing.T) {
	creator := NewSimpleHabitCreator("Test", "Test", models.SimpleHabit)
	creator.selectedFieldType = "numeric"
	creator.scoringType = models.AutomaticScoring
	creator.criteriaType = "greater_than"
	creator.criteriaValue = "not_a_number"

	_, err := creator.buildCriteriaFromData()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid criteria value")
}

func TestSimpleHabitCreator_CriteriaBuilding_UnsupportedFieldType(t *testing.T) {
	creator := NewSimpleHabitCreator("Test", "Test", models.SimpleHabit)
	creator.selectedFieldType = "unsupported_type"
	creator.scoringType = models.AutomaticScoring

	_, err := creator.buildCriteriaFromData()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "automatic scoring not supported for field type")
}
