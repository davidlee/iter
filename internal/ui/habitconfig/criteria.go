package habitconfig

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"

	"davidlee/vice/internal/models"
)

// CriteriaBuilder provides methods to build criteria configuration forms
type CriteriaBuilder struct {
	formBuilder *HabitFormBuilder
}

// NewCriteriaBuilder creates a new criteria builder
func NewCriteriaBuilder() *CriteriaBuilder {
	return &CriteriaBuilder{
		formBuilder: NewHabitFormBuilder(),
	}
}

// CriteriaConfig holds criteria configuration
type CriteriaConfig struct {
	Description      string
	ComparisonType   string
	Value            string
	RangeMin         string
	RangeMax         string
	TimeAfter        string
	TimeBefore       string
	BooleanValue     bool
	TextContains     string
	TextMatchesRegex string
}

// CreateSimpleCriteriaForm creates a form for simple habit criteria
func (cb *CriteriaBuilder) CreateSimpleCriteriaForm(fieldType models.FieldType) (*huh.Form, *CriteriaConfig) {
	config := &CriteriaConfig{}

	if fieldType.Type == models.BooleanFieldType {
		// Boolean fields are straightforward - just true/false
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Completion Criteria").
					Description("Habit is achieved when the answer is 'Yes'").
					Value(&config.BooleanValue),

				huh.NewInput().
					Title("Description (optional)").
					Description("Explain what achieving this habit means").
					Value(&config.Description),
			),
		)
		config.BooleanValue = true // Default for boolean is true
		return form, config
	}

	// For other field types, provide comparison options
	return cb.createComparisonCriteriaForm(fieldType, config)
}

// CreateElasticCriteriaForm creates forms for elastic habit criteria (mini/midi/maxi)
func (cb *CriteriaBuilder) CreateElasticCriteriaForm(fieldType models.FieldType, level string) (*huh.Form, *CriteriaConfig) {
	config := &CriteriaConfig{}

	if fieldType.Type == models.BooleanFieldType {
		// Boolean elastic habits don't make much sense, but handle gracefully
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("%s Level Description", strings.ToUpper(string(level[0]))+level[1:])).
					Description(fmt.Sprintf("Describe what %s achievement means", level)).
					Value(&config.Description).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return fmt.Errorf("description is required for %s level", level)
						}
						return nil
					}),
			),
		)
		config.BooleanValue = true
		return form, config
	}

	// For other field types, provide comparison options with level context
	return cb.createElasticComparisonForm(fieldType, level, config)
}

func (cb *CriteriaBuilder) createComparisonCriteriaForm(fieldType models.FieldType, config *CriteriaConfig) (*huh.Form, *CriteriaConfig) {
	var fields []huh.Field

	fields = append(fields,
		huh.NewInput().
			Title("Description (optional)").
			Description("Explain what achieving this habit means").
			Value(&config.Description),
	)

	switch fieldType.Type {
	case models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType, models.DurationFieldType:
		fields = append(fields,
			huh.NewSelect[string]().
				Title("Comparison Type").
				Description("How should the value be compared?").
				Options(
					huh.NewOption("Greater than or equal to", "gte").Selected(true),
					huh.NewOption("Greater than", "gt"),
					huh.NewOption("Less than or equal to", "lte"),
					huh.NewOption("Less than", "lt"),
					huh.NewOption("Within range", "range"),
				).
				Value(&config.ComparisonType),

			huh.NewInput().
				Title("Value").
				Description(cb.getValueDescription(fieldType)).
				Value(&config.Value).
				Validate(cb.createValueValidator(fieldType)),
		)

	case models.TimeFieldType:
		fields = append(fields,
			huh.NewSelect[string]().
				Title("Comparison Type").
				Description("How should the time be compared?").
				Options(
					huh.NewOption("After (or at) time", "after").Selected(true),
					huh.NewOption("Before (or at) time", "before"),
					huh.NewOption("Within time range", "range"),
				).
				Value(&config.ComparisonType),

			huh.NewInput().
				Title("Time").
				Description("Enter time in HH:MM format (e.g., 14:30)").
				Value(&config.Value).
				Validate(cb.createTimeValidator()),
		)

	case models.TextFieldType:
		fields = append(fields,
			huh.NewSelect[string]().
				Title("Text Matching").
				Description("How should the text be evaluated?").
				Options(
					huh.NewOption("Contains text", "contains").Selected(true),
					huh.NewOption("Matches regex pattern", "regex"),
				).
				Value(&config.ComparisonType),

			huh.NewInput().
				Title("Pattern").
				Description("Text to search for or regex pattern").
				Value(&config.Value).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("pattern cannot be empty")
					}
					return nil
				}),
		)
	}

	return huh.NewForm(huh.NewGroup(fields...)), config
}

func (cb *CriteriaBuilder) createElasticComparisonForm(fieldType models.FieldType, level string, config *CriteriaConfig) (*huh.Form, *CriteriaConfig) {
	levelTitle := strings.ToUpper(string(level[0])) + level[1:]

	var fields []huh.Field

	fields = append(fields,
		huh.NewInput().
			Title(fmt.Sprintf("%s Level Description", levelTitle)).
			Description(fmt.Sprintf("Describe what %s achievement means", level)).
			Value(&config.Description),
	)

	switch fieldType.Type {
	case models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType, models.DurationFieldType:
		fields = append(fields,
			huh.NewSelect[string]().
				Title("Comparison Type").
				Description(fmt.Sprintf("How should values be compared for %s level?", level)).
				Options(
					huh.NewOption("Greater than or equal to", "gte").Selected(true),
					huh.NewOption("Greater than", "gt"),
					huh.NewOption("Less than or equal to", "lte"),
					huh.NewOption("Less than", "lt"),
					huh.NewOption("Within range", "range"),
				).
				Value(&config.ComparisonType),

			huh.NewInput().
				Title(fmt.Sprintf("%s Level Value", levelTitle)).
				Description(cb.getValueDescription(fieldType)).
				Value(&config.Value).
				Validate(cb.createValueValidator(fieldType)),
		)

	case models.TimeFieldType:
		fields = append(fields,
			huh.NewSelect[string]().
				Title("Comparison Type").
				Description(fmt.Sprintf("How should time be compared for %s level?", level)).
				Options(
					huh.NewOption("After (or at) time", "after").Selected(true),
					huh.NewOption("Before (or at) time", "before"),
					huh.NewOption("Within time range", "range"),
				).
				Value(&config.ComparisonType),

			huh.NewInput().
				Title(fmt.Sprintf("%s Level Time", levelTitle)).
				Description("Enter time in HH:MM format (e.g., 14:30)").
				Value(&config.Value).
				Validate(cb.createTimeValidator()),
		)

	case models.TextFieldType:
		fields = append(fields,
			huh.NewSelect[string]().
				Title("Text Matching").
				Description(fmt.Sprintf("How should text be evaluated for %s level?", level)).
				Options(
					huh.NewOption("Contains text", "contains").Selected(true),
					huh.NewOption("Matches regex pattern", "regex"),
				).
				Value(&config.ComparisonType),

			huh.NewInput().
				Title(fmt.Sprintf("%s Level Pattern", levelTitle)).
				Description("Text to search for or regex pattern").
				Value(&config.Value).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("pattern cannot be empty")
					}
					return nil
				}),
		)
	}

	return huh.NewForm(huh.NewGroup(fields...)), config
}

// Helper methods

func (cb *CriteriaBuilder) getValueDescription(fieldType models.FieldType) string {
	switch fieldType.Type {
	case models.UnsignedIntFieldType:
		return "Enter a positive whole number"
	case models.UnsignedDecimalFieldType:
		return "Enter a positive decimal number"
	case models.DecimalFieldType:
		return "Enter a decimal number"
	case models.DurationFieldType:
		return "Enter duration (e.g., 30m, 1h30m, 2h)"
	default:
		return "Enter the comparison value"
	}
}

func (cb *CriteriaBuilder) createValueValidator(fieldType models.FieldType) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("value cannot be empty")
		}

		switch fieldType.Type {
		case models.UnsignedIntFieldType:
			val, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("must be a whole number")
			}
			if val < 0 {
				return fmt.Errorf("must be a positive number")
			}

		case models.UnsignedDecimalFieldType:
			val, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a valid number")
			}
			if val < 0 {
				return fmt.Errorf("must be a positive number")
			}

		case models.DecimalFieldType:
			_, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a valid number")
			}

		case models.DurationFieldType:
			if !isDurationFormat(s) {
				return fmt.Errorf("must be in duration format (e.g., 30m, 1h30m)")
			}
		}

		return nil
	}
}

func (cb *CriteriaBuilder) createTimeValidator() func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("time cannot be empty")
		}

		if !isTimeFormat(s) {
			return fmt.Errorf("must be in HH:MM format (e.g., 14:30)")
		}

		return nil
	}
}

// Helper functions for validation

func isDurationFormat(s string) bool {
	// Simple validation for duration format like "30m", "1h30m", "2h"
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	// Check for basic duration patterns
	for _, suffix := range []string{"m", "h", "s"} {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}

	// Check for combined patterns like "1h30m"
	if strings.Contains(s, "h") && strings.Contains(s, "m") {
		return true
	}

	return false
}

func isTimeFormat(s string) bool {
	// Simple validation for HH:MM format
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) != 2 {
		return false
	}

	hour, err1 := strconv.Atoi(parts[0])
	minute, err2 := strconv.Atoi(parts[1])

	return err1 == nil && err2 == nil &&
		hour >= 0 && hour <= 23 &&
		minute >= 0 && minute <= 59
}
