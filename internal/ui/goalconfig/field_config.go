package goalconfig

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"

	"davidlee/iter/internal/models"
)

// FieldConfig represents the complete configuration for a field type
type FieldConfig struct {
	Type      string
	Subtype   string   // For numeric fields: "unsigned_int", "unsigned_decimal", "decimal"
	Unit      string   // For numeric fields: "times", "reps", "kg", etc.
	Multiline bool     // For text fields
	Min       *float64 // For numeric fields (optional)
	Max       *float64 // For numeric fields (optional)
	Direction string   // For applicable fields: "higher_better", "lower_better", "neutral"
}

// FieldTypeSelector provides interactive field type selection and configuration
type FieldTypeSelector struct {
	selectedType   string
	numericSubtype string
	unit           string
	multilineText  bool
	minValue       string
	maxValue       string
	direction      string
	hasMinMax      bool
}

// NewFieldTypeSelector creates a new field type selector
func NewFieldTypeSelector() *FieldTypeSelector {
	return &FieldTypeSelector{
		numericSubtype: models.UnsignedIntFieldType, // Default numeric subtype
		unit:           "times",                     // Default unit
		direction:      "neutral",                   // Default direction
	}
}

// SelectFieldType prompts user to select a field type
func (fts *FieldTypeSelector) SelectFieldType() error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Field Type").
				Description("Choose what type of data this goal will track").
				Options(
					huh.NewOption("Boolean (True/False)", models.BooleanFieldType),
					huh.NewOption("Text (Written responses)", models.TextFieldType),
					huh.NewOption("Numeric (Numbers with units)", "numeric"),
					huh.NewOption("Time (Time of day)", models.TimeFieldType),
					huh.NewOption("Duration (Time periods)", models.DurationFieldType),
				).
				Value(&fts.selectedType),
		),
	)

	return form.Run()
}

// ConfigureField configures the selected field type with additional options
func (fts *FieldTypeSelector) ConfigureField() error {
	switch fts.selectedType {
	case "numeric":
		return fts.configureNumericField()
	case models.TextFieldType:
		return fts.configureTextField()
	case models.BooleanFieldType, models.TimeFieldType, models.DurationFieldType:
		// These field types need no additional configuration
		return nil
	default:
		return fmt.Errorf("unknown field type: %s", fts.selectedType)
	}
}

// ConfigureDirection prompts for direction preference for applicable field types
func (fts *FieldTypeSelector) ConfigureDirection() error {
	// Only certain field types support direction
	if !fts.supportsDirection() {
		fts.direction = "neutral"
		return nil
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Value Direction").
				Description("Indicates whether higher or lower values are generally better").
				Options(
					huh.NewOption("Higher is better (↑)", "higher_better"),
					huh.NewOption("Lower is better (↓)", "lower_better"),
					huh.NewOption("Neutral (no preference)", "neutral"),
				).
				Value(&fts.direction),
		),
	)

	return form.Run()
}

// GetFieldConfig returns the complete field configuration
func (fts *FieldTypeSelector) GetFieldConfig() *FieldConfig {
	config := &FieldConfig{
		Direction: fts.direction,
		Multiline: fts.multilineText,
	}

	// Set the actual field type based on selection
	switch fts.selectedType {
	case "numeric":
		config.Type = fts.numericSubtype
		config.Subtype = fts.numericSubtype
		config.Unit = fts.unit
		if fts.hasMinMax {
			if fts.minValue != "" {
				if minVal, err := strconv.ParseFloat(fts.minValue, 64); err == nil {
					config.Min = &minVal
				}
			}
			if fts.maxValue != "" {
				if maxVal, err := strconv.ParseFloat(fts.maxValue, 64); err == nil {
					config.Max = &maxVal
				}
			}
		}
	case models.BooleanFieldType, models.TextFieldType, models.TimeFieldType, models.DurationFieldType:
		config.Type = fts.selectedType
	}

	return config
}

// Private methods

func (fts *FieldTypeSelector) configureNumericField() error {
	// Step 1: Numeric subtype selection
	subtypeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Numeric Type").
				Description("Choose the type of numbers you'll be tracking").
				Options(
					huh.NewOption("Whole numbers (0, 1, 2, 3...)", models.UnsignedIntFieldType),
					huh.NewOption("Positive decimals (0.5, 1.25, 2.7...)", models.UnsignedDecimalFieldType),
					huh.NewOption("Any numbers (including negative)", models.DecimalFieldType),
				).
				Value(&fts.numericSubtype),
		),
	)

	if err := subtypeForm.Run(); err != nil {
		return err
	}

	// Step 2: Unit configuration
	unitForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Unit").
				Description("What unit will you measure in? (e.g., 'reps', 'kg', 'minutes', 'pages')").
				Placeholder("times").
				Value(&fts.unit).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						fts.unit = "times" // Default if empty
					} else {
						fts.unit = strings.TrimSpace(s)
					}
					return nil
				}),
		),
	)

	if err := unitForm.Run(); err != nil {
		return err
	}

	// Step 3: Optional min/max constraints
	constraintsForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Add Value Constraints?").
				Description("Do you want to set minimum or maximum value limits?").
				Value(&fts.hasMinMax),
		),
	)

	if err := constraintsForm.Run(); err != nil {
		return err
	}

	// Step 4: Configure min/max if requested
	if fts.hasMinMax {
		return fts.configureMinMax()
	}

	return nil
}

func (fts *FieldTypeSelector) configureTextField() error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Multiline Text").
				Description("Will you need multiple lines for text responses?").
				Value(&fts.multilineText),
		),
	)

	return form.Run()
}

func (fts *FieldTypeSelector) configureMinMax() error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Minimum Value (optional)").
				Description("Leave empty for no minimum limit").
				Value(&fts.minValue).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return nil // Empty is OK
					}
					if _, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err != nil {
						return fmt.Errorf("must be a valid number")
					}
					return nil
				}),

			huh.NewInput().
				Title("Maximum Value (optional)").
				Description("Leave empty for no maximum limit").
				Value(&fts.maxValue).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return nil // Empty is OK
					}
					if _, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err != nil {
						return fmt.Errorf("must be a valid number")
					}
					return nil
				}),
		),
	)

	return form.Run()
}

func (fts *FieldTypeSelector) supportsDirection() bool {
	// Only numeric, time, and duration fields support direction preference
	switch fts.selectedType {
	case "numeric", models.TimeFieldType, models.DurationFieldType:
		return true
	case models.BooleanFieldType, models.TextFieldType:
		return false
	default:
		return false
	}
}
