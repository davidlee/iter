// Package habitconfig provides interactive UI components for habit configuration.
package habitconfig

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/davidlee/vice/internal/models"
)

// HabitBuilder orchestrates the complete habit creation process
type HabitBuilder struct {
	formBuilder     *HabitFormBuilder
	criteriaBuilder *CriteriaBuilder
}

// NewHabitBuilder creates a new habit builder
func NewHabitBuilder() *HabitBuilder {
	return &HabitBuilder{
		formBuilder:     NewHabitFormBuilder(),
		criteriaBuilder: NewCriteriaBuilder(),
	}
}

// BuildHabit runs the complete interactive flow to create a new habit
func (gb *HabitBuilder) BuildHabit(_ []models.Habit) (*models.Habit, error) {
	// Step 1: Basic information
	basicForm, basicInfo := gb.formBuilder.CreateBasicInfoForm()
	if err := basicForm.Run(); err != nil {
		return nil, fmt.Errorf("failed to collect basic information: %w", err)
	}

	// Step 2: Field type selection
	fieldForm, fieldInfo := gb.formBuilder.CreateFieldTypeForm(basicInfo.HabitType)
	if err := fieldForm.Run(); err != nil {
		return nil, fmt.Errorf("failed to collect field type: %w", err)
	}

	// Step 3: Field details (if needed)
	if needsFieldDetails(fieldInfo.Type) {
		detailsForm, detailsInfo := gb.formBuilder.CreateFieldDetailsForm(fieldInfo.Type)
		if err := detailsForm.Run(); err != nil {
			return nil, fmt.Errorf("failed to collect field details: %w", err)
		}
		// Merge details into fieldInfo
		fieldInfo.Unit = detailsInfo.Unit
		fieldInfo.Multiline = detailsInfo.Multiline
		fieldInfo.Min = detailsInfo.Min
		fieldInfo.Max = detailsInfo.Max
	}

	// Step 4: Scoring configuration
	scoringForm, scoringInfo := gb.formBuilder.CreateScoringForm(basicInfo.HabitType)
	if err := scoringForm.Run(); err != nil {
		return nil, fmt.Errorf("failed to collect scoring configuration: %w", err)
	}

	// Step 5: Criteria configuration (if automatic scoring)
	var criteria *models.Criteria
	var miniCriteria, midiCriteria, maxiCriteria *models.Criteria
	var err error

	if scoringInfo.ScoringType == models.AutomaticScoring {
		fieldType := models.FieldType{
			Type:      fieldInfo.Type,
			Unit:      fieldInfo.Unit,
			Multiline: &fieldInfo.Multiline,
			Min:       fieldInfo.Min,
			Max:       fieldInfo.Max,
		}

		switch basicInfo.HabitType {
		case models.SimpleHabit:
			criteria, err = gb.buildSimpleCriteria(fieldType)
			if err != nil {
				return nil, fmt.Errorf("failed to build criteria: %w", err)
			}

		case models.ElasticHabit:
			miniCriteria, midiCriteria, maxiCriteria, err = gb.buildElasticCriteria(fieldType)
			if err != nil {
				return nil, fmt.Errorf("failed to build elastic criteria: %w", err)
			}
		}
	}

	// Step 6: Build the complete habit
	habit := &models.Habit{
		Title:       strings.TrimSpace(basicInfo.Title),
		Description: strings.TrimSpace(basicInfo.Description),
		HabitType:   basicInfo.HabitType,
		// AIDEV-NOTE: Position is inferred and should not be set in habit creation
		// Position will be determined by the parser/schema based on order in habits.yml
		FieldType: models.FieldType{
			Type:      fieldInfo.Type,
			Unit:      fieldInfo.Unit,
			Multiline: &fieldInfo.Multiline,
			Min:       fieldInfo.Min,
			Max:       fieldInfo.Max,
		},
		ScoringType: scoringInfo.ScoringType,
		Direction:   scoringInfo.Direction,
	}

	// Add criteria based on habit type
	if criteria != nil {
		habit.Criteria = criteria
	}
	if miniCriteria != nil {
		habit.MiniCriteria = miniCriteria
	}
	if midiCriteria != nil {
		habit.MidiCriteria = midiCriteria
	}
	if maxiCriteria != nil {
		habit.MaxiCriteria = maxiCriteria
	}

	return habit, nil
}

func (gb *HabitBuilder) buildSimpleCriteria(fieldType models.FieldType) (*models.Criteria, error) {
	criteriaForm, criteriaConfig := gb.criteriaBuilder.CreateSimpleCriteriaForm(fieldType)
	if err := criteriaForm.Run(); err != nil {
		return nil, err
	}

	return gb.configToCriteria(criteriaConfig, fieldType)
}

func (gb *HabitBuilder) buildElasticCriteria(fieldType models.FieldType) (*models.Criteria, *models.Criteria, *models.Criteria, error) {
	// Build criteria for each level
	levels := []string{"mini", "midi", "maxi"}
	criteriaList := make([]*models.Criteria, 3)

	for i, level := range levels {
		fmt.Printf("\n=== Configuring %s level ===\n", strings.ToUpper(string(level[0]))+level[1:])

		criteriaForm, criteriaConfig := gb.criteriaBuilder.CreateElasticCriteriaForm(fieldType, level)
		if err := criteriaForm.Run(); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to configure %s level: %w", level, err)
		}

		criteria, err := gb.configToCriteria(criteriaConfig, fieldType)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to build %s criteria: %w", level, err)
		}

		criteriaList[i] = criteria
	}

	return criteriaList[0], criteriaList[1], criteriaList[2], nil
}

func (gb *HabitBuilder) configToCriteria(config *CriteriaConfig, fieldType models.FieldType) (*models.Criteria, error) {
	criteria := &models.Criteria{
		Description: config.Description,
		Condition:   &models.Condition{},
	}

	switch fieldType.Type {
	case models.BooleanFieldType:
		criteria.Condition.Equals = &config.BooleanValue

	case models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType, models.DurationFieldType:
		value, err := strconv.ParseFloat(config.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid numeric value: %w", err)
		}

		switch config.ComparisonType {
		case "gt":
			criteria.Condition.GreaterThan = &value
		case "gte":
			criteria.Condition.GreaterThanOrEqual = &value
		case "lt":
			criteria.Condition.LessThan = &value
		case "lte":
			criteria.Condition.LessThanOrEqual = &value
		case "range":
			minVal, err := strconv.ParseFloat(config.RangeMin, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid range minimum: %w", err)
			}
			maxVal, err := strconv.ParseFloat(config.RangeMax, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid range maximum: %w", err)
			}
			criteria.Condition.Range = &models.RangeCondition{
				Min: minVal,
				Max: maxVal,
			}
		}

	case models.TimeFieldType:
		switch config.ComparisonType {
		case "after":
			criteria.Condition.After = config.Value
		case "before":
			criteria.Condition.Before = config.Value
		case "range":
			// For now, just use After field - TimeRange may need to be added to model later
			criteria.Condition.After = config.TimeAfter
			criteria.Condition.Before = config.TimeBefore
		}

	case models.TextFieldType:
		// Text condition fields don't exist in current model
		// For now, we'll need to extend the model or use a different approach
		// This is a placeholder that will need model updates
		return nil, fmt.Errorf("text field criteria not yet implemented in the model")
	}

	return criteria, nil
}

// BuildHabitWithBasicInfo runs the habit creation flow with pre-populated basic info
func (gb *HabitBuilder) BuildHabitWithBasicInfo(_ interface{}, existingHabits []models.Habit) (*models.Habit, error) {
	// For now, delegate to the original BuildHabit since it already collects basic info
	// This maintains backwards compatibility while providing the interface needed
	// AIDEV-TODO: Optimize to skip basic info collection when it's pre-populated
	return gb.BuildHabit(existingHabits)
}

// Helper functions

func needsFieldDetails(fieldType string) bool {
	// Only certain field types need additional configuration
	return isNumericField(fieldType) || fieldType == models.TextFieldType
}
