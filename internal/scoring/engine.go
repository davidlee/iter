// Package scoring provides functionality for evaluating habit achievements against criteria.
package scoring

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"davidlee/vice/internal/models"
)

// Engine handles scoring of habit entries against elastic habit criteria.
type Engine struct{}

// NewEngine creates a new scoring engine instance.
func NewEngine() *Engine {
	return &Engine{}
}

// ScoreResult represents the result of scoring a value against elastic criteria.
type ScoreResult struct {
	AchievementLevel models.AchievementLevel
	MetMini          bool
	MetMidi          bool
	MetMaxi          bool
}

// ScoreSimpleHabit evaluates a value against simple habit criteria and returns pass/fail.
// AIDEV-NOTE: habit-type-separation; dedicated scoring method prevents type masquerading anti-pattern
// Simple habits have a single criteria that determines pass (mini) or fail (none).
// IMPORTANT: This method was added to fix T016 - simple habits were incorrectly trying to masquerade as elastic habits.
func (e *Engine) ScoreSimpleHabit(habit *models.Habit, value interface{}) (*ScoreResult, error) {
	if habit == nil {
		return nil, fmt.Errorf("habit cannot be nil")
	}

	if !habit.IsSimple() {
		return nil, fmt.Errorf("habit %s is not a simple habit", habit.ID)
	}

	if !habit.RequiresAutomaticScoring() {
		return nil, fmt.Errorf("habit %s does not require automatic scoring", habit.ID)
	}

	if habit.Criteria == nil {
		return nil, fmt.Errorf("habit %s has no criteria for automatic scoring", habit.ID)
	}

	// Initialize result
	result := &ScoreResult{
		AchievementLevel: models.AchievementNone,
		MetMini:          false,
		MetMidi:          false,
		MetMaxi:          false,
	}

	// Convert value to appropriate type for evaluation
	evaluationValue, err := e.convertValueForEvaluation(value, habit.FieldType.Type)
	if err != nil {
		return nil, err
	}

	// Evaluate against the single criteria
	met, err := e.evaluateCriteria(evaluationValue, habit.Criteria, habit.FieldType.Type)
	if err != nil {
		return nil, err
	}

	// For simple habits, criteria met = mini achievement level (pass)
	if met {
		result.AchievementLevel = models.AchievementMini
		result.MetMini = true
	}

	return result, nil
}

// ScoreElasticHabit evaluates a value against elastic habit criteria and returns the achievement level.
// Returns the highest achievement level met (none, mini, midi, or maxi).
func (e *Engine) ScoreElasticHabit(habit *models.Habit, value interface{}) (*ScoreResult, error) {
	if habit == nil {
		return nil, fmt.Errorf("habit cannot be nil")
	}

	if !habit.IsElastic() {
		return nil, fmt.Errorf("habit %s is not an elastic habit", habit.ID)
	}

	if !habit.RequiresAutomaticScoring() {
		return nil, fmt.Errorf("habit %s does not require automatic scoring", habit.ID)
	}

	// Initialize result
	result := &ScoreResult{
		AchievementLevel: models.AchievementNone,
		MetMini:          false,
		MetMidi:          false,
		MetMaxi:          false,
	}

	// Convert value to appropriate type for evaluation
	evaluationValue, err := e.convertValueForEvaluation(value, habit.FieldType.Type)
	if err != nil {
		return nil, err
	}

	// Evaluate against each criteria level
	if habit.MiniCriteria != nil {
		met, err := e.evaluateCriteria(evaluationValue, habit.MiniCriteria, habit.FieldType.Type)
		if err != nil {
			return nil, err
		}
		result.MetMini = met
		if met {
			result.AchievementLevel = models.AchievementMini
		}
	}

	if habit.MidiCriteria != nil {
		met, err := e.evaluateCriteria(evaluationValue, habit.MidiCriteria, habit.FieldType.Type)
		if err != nil {
			return nil, err
		}
		result.MetMidi = met
		if met {
			result.AchievementLevel = models.AchievementMidi
		}
	}

	if habit.MaxiCriteria != nil {
		met, err := e.evaluateCriteria(evaluationValue, habit.MaxiCriteria, habit.FieldType.Type)
		if err != nil {
			return nil, err
		}
		result.MetMaxi = met
		if met {
			result.AchievementLevel = models.AchievementMaxi
		}
	}

	return result, nil
}

// convertValueForEvaluation converts the input value to the appropriate type for evaluation.
func (e *Engine) convertValueForEvaluation(value interface{}, fieldType string) (interface{}, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}

	switch fieldType {
	case models.UnsignedIntFieldType:
		return e.convertToFloat64(value)
	case models.UnsignedDecimalFieldType, models.DecimalFieldType:
		return e.convertToFloat64(value)
	case models.DurationFieldType:
		return e.convertDurationToMinutes(value)
	case models.TimeFieldType:
		return e.convertTimeToMinutes(value)
	case models.BooleanFieldType:
		return e.convertToBool(value)
	case models.TextFieldType:
		return e.convertToString(value)
	default:
		return nil, fmt.Errorf("unsupported field type for scoring: %s", fieldType)
	}
}

// evaluateCriteria evaluates a value against specific criteria.
func (e *Engine) evaluateCriteria(value interface{}, criteria *models.Criteria, fieldType string) (bool, error) {
	if criteria == nil || criteria.Condition == nil {
		return false, fmt.Errorf("criteria or condition cannot be nil")
	}

	condition := criteria.Condition

	switch fieldType {
	case models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType, models.DurationFieldType:
		return e.evaluateNumericCondition(value, condition)
	case models.TimeFieldType:
		return e.evaluateTimeCondition(value, condition)
	case models.BooleanFieldType:
		return e.evaluateBooleanCondition(value, condition)
	case models.TextFieldType:
		return e.evaluateTextCondition(value, condition)
	default:
		return false, fmt.Errorf("unsupported field type for criteria evaluation: %s", fieldType)
	}
}

// evaluateNumericCondition evaluates numeric values against numeric conditions.
func (e *Engine) evaluateNumericCondition(value interface{}, condition *models.Condition) (bool, error) {
	numValue, ok := value.(float64)
	if !ok {
		return false, fmt.Errorf("expected numeric value, got %T", value)
	}

	// Check greater than
	if condition.GreaterThan != nil {
		return numValue > *condition.GreaterThan, nil
	}

	// Check greater than or equal
	if condition.GreaterThanOrEqual != nil {
		return numValue >= *condition.GreaterThanOrEqual, nil
	}

	// Check less than
	if condition.LessThan != nil {
		return numValue < *condition.LessThan, nil
	}

	// Check less than or equal
	if condition.LessThanOrEqual != nil {
		return numValue <= *condition.LessThanOrEqual, nil
	}

	// Check range
	if condition.Range != nil {
		minInclusive := true
		maxInclusive := true
		if condition.Range.MinInclusive != nil {
			minInclusive = *condition.Range.MinInclusive
		}
		if condition.Range.MaxInclusive != nil {
			maxInclusive = *condition.Range.MaxInclusive
		}

		var minMet, maxMet bool
		if minInclusive {
			minMet = numValue >= condition.Range.Min
		} else {
			minMet = numValue > condition.Range.Min
		}

		if maxInclusive {
			maxMet = numValue <= condition.Range.Max
		} else {
			maxMet = numValue < condition.Range.Max
		}

		return minMet && maxMet, nil
	}

	return false, fmt.Errorf("no valid numeric condition found")
}

// evaluateTimeCondition evaluates time values against time conditions.
func (e *Engine) evaluateTimeCondition(value interface{}, condition *models.Condition) (bool, error) {
	timeValue, ok := value.(float64) // Time converted to minutes since midnight
	if !ok {
		return false, fmt.Errorf("expected time value as minutes, got %T", value)
	}

	// Handle before/after time constraints
	if condition.Before != "" {
		beforeMinutes, err := e.parseTimeToMinutes(condition.Before)
		if err != nil {
			return false, fmt.Errorf("invalid before time: %w", err)
		}
		return timeValue < beforeMinutes, nil
	}

	if condition.After != "" {
		afterMinutes, err := e.parseTimeToMinutes(condition.After)
		if err != nil {
			return false, fmt.Errorf("invalid after time: %w", err)
		}
		return timeValue > afterMinutes, nil
	}

	// Fall back to numeric evaluation for other operators
	return e.evaluateNumericCondition(value, condition)
}

// evaluateBooleanCondition evaluates boolean values against boolean conditions.
func (e *Engine) evaluateBooleanCondition(value interface{}, condition *models.Condition) (bool, error) {
	boolValue, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("expected boolean value, got %T", value)
	}

	if condition.Equals != nil {
		return boolValue == *condition.Equals, nil
	}

	return false, fmt.Errorf("no valid boolean condition found")
}

// evaluateTextCondition evaluates text values against text conditions.
func (e *Engine) evaluateTextCondition(value interface{}, condition *models.Condition) (bool, error) {
	// For text fields, we primarily support logical operators and length-based comparisons
	textValue, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("expected string value, got %T", value)
	}

	// For text, we can evaluate length as a numeric comparison
	textLength := float64(len(textValue))

	// Check if any numeric operators are defined (treating them as length comparisons)
	if condition.GreaterThan != nil {
		return textLength > *condition.GreaterThan, nil
	}
	if condition.GreaterThanOrEqual != nil {
		return textLength >= *condition.GreaterThanOrEqual, nil
	}
	if condition.LessThan != nil {
		return textLength < *condition.LessThan, nil
	}
	if condition.LessThanOrEqual != nil {
		return textLength <= *condition.LessThanOrEqual, nil
	}

	// For text fields without specific criteria, assume any non-empty text meets the criteria
	return len(strings.TrimSpace(textValue)) > 0, nil
}

// Helper conversion functions

func (e *Engine) convertToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func (e *Engine) convertDurationToMinutes(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64, float32, int, int64, uint, uint64:
		// Assume numeric values are already in minutes
		return e.convertToFloat64(v)
	case string:
		// Parse duration string (e.g., "1h30m", "90", "90m")
		return e.parseDurationToMinutes(v)
	default:
		return 0, fmt.Errorf("cannot convert %T to duration in minutes", value)
	}
}

func (e *Engine) convertTimeToMinutes(value interface{}) (float64, error) {
	switch v := value.(type) {
	case string:
		// Parse time string (e.g., "14:30", "2:30 PM")
		return e.parseTimeToMinutes(v)
	case time.Time:
		// Convert time.Time to minutes since midnight
		return float64(v.Hour()*60 + v.Minute()), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to time in minutes", value)
	}
}

func (e *Engine) convertToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("cannot convert %T to boolean", value)
	}
}

func (e *Engine) convertToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// parseDurationToMinutes parses duration strings to minutes.
func (e *Engine) parseDurationToMinutes(duration string) (float64, error) {
	duration = strings.TrimSpace(duration)

	// Try parsing as Go duration first (e.g., "1h30m", "90m")
	if strings.ContainsAny(duration, "hms") {
		d, err := time.ParseDuration(duration)
		if err == nil {
			return d.Minutes(), nil
		}
	}

	// Try parsing as plain number (assume minutes)
	if minutes, err := strconv.ParseFloat(duration, 64); err == nil {
		return minutes, nil
	}

	// Try parsing "HH:MM:SS" format
	if strings.Count(duration, ":") == 2 {
		parts := strings.Split(duration, ":")
		if len(parts) == 3 {
			hours, err1 := strconv.ParseFloat(parts[0], 64)
			minutes, err2 := strconv.ParseFloat(parts[1], 64)
			seconds, err3 := strconv.ParseFloat(parts[2], 64)
			if err1 == nil && err2 == nil && err3 == nil {
				return hours*60 + minutes + seconds/60, nil
			}
		}
	}

	return 0, fmt.Errorf("cannot parse duration: %s", duration)
}

// parseTimeToMinutes parses time strings to minutes since midnight.
func (e *Engine) parseTimeToMinutes(timeStr string) (float64, error) {
	timeStr = strings.TrimSpace(timeStr)

	// Try parsing "HH:MM" format
	if strings.Count(timeStr, ":") == 1 {
		parts := strings.Split(timeStr, ":")
		if len(parts) == 2 {
			hours, err1 := strconv.ParseFloat(parts[0], 64)
			minutes, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil && hours >= 0 && hours < 24 && minutes >= 0 && minutes < 60 {
				return hours*60 + minutes, nil
			}
		}
	}

	return 0, fmt.Errorf("cannot parse time: %s (expected HH:MM format)", timeStr)
}
