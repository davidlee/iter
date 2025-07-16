package habitconfig

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/models"
)

// HabitFormBuilder provides methods to build interactive forms for habit configuration
type HabitFormBuilder struct {
	titleStyle       lipgloss.Style
	descriptionStyle lipgloss.Style
	helpStyle        lipgloss.Style
	errorStyle       lipgloss.Style
}

// NewHabitFormBuilder creates a new form builder with consistent styling
func NewHabitFormBuilder() *HabitFormBuilder {
	return &HabitFormBuilder{
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")). // Bright blue
			Margin(1, 0),
		descriptionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Italic(true),
		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")). // Bright green
			Faint(true),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")). // Bright red
			Bold(true),
	}
}

// HabitBasicInfo holds basic habit information
type HabitBasicInfo struct {
	Title       string
	Description string
	HabitType   models.HabitType
}

// FieldTypeInfo holds field type configuration
type FieldTypeInfo struct {
	Type      string
	Unit      string
	Multiline bool
	Min       *float64
	Max       *float64
}

// ScoringInfo holds scoring configuration
type ScoringInfo struct {
	ScoringType models.ScoringType
	Direction   string // For informational habits
}

// CreateBasicInfoForm creates a form for collecting basic habit information
func (fb *HabitFormBuilder) CreateBasicInfoForm() (*huh.Form, *HabitBasicInfo) {
	info := &HabitBasicInfo{}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Habit Title").
				Description("Enter a descriptive name for your habit").
				Value(&info.Title).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("title cannot be empty")
					}
					if len(s) > 100 {
						return fmt.Errorf("title must be 100 characters or less")
					}
					return nil
				}),

			huh.NewText().
				Title("Description (optional)").
				Description("Provide additional context about this habit").
				Value(&info.Description),

			huh.NewSelect[models.HabitType]().
				Title("Habit Type").
				Description("Choose how this habit will be tracked and scored").
				Options(
					huh.NewOption("Simple (Pass/Fail)", models.SimpleHabit).
						Selected(true),
					huh.NewOption("Elastic (Mini/Midi/Maxi levels)", models.ElasticHabit),
					huh.NewOption("Informational (Data tracking only)", models.InformationalHabit),
				).
				Value(&info.HabitType),
		),
	)

	return form, info
}

// CreateFieldTypeForm creates a form for selecting field type based on habit type
func (fb *HabitFormBuilder) CreateFieldTypeForm(goalType models.HabitType) (*huh.Form, *FieldTypeInfo) {
	info := &FieldTypeInfo{}

	var fieldTypeOptions []huh.Option[string]
	var defaultSelection string

	switch goalType {
	case models.SimpleHabit:
		// Simple habits are always boolean
		fieldTypeOptions = []huh.Option[string]{
			huh.NewOption("Boolean (Yes/No)", models.BooleanFieldType).Selected(true),
		}
		defaultSelection = models.BooleanFieldType

	case models.ElasticHabit, models.InformationalHabit:
		fieldTypeOptions = []huh.Option[string]{
			huh.NewOption("Number (unsigned integer)", models.UnsignedIntFieldType).Selected(true),
			huh.NewOption("Number (unsigned decimal)", models.UnsignedDecimalFieldType),
			huh.NewOption("Number (decimal)", models.DecimalFieldType),
			huh.NewOption("Duration (e.g., 30m, 1h30m)", models.DurationFieldType),
			huh.NewOption("Time (e.g., 14:30)", models.TimeFieldType),
			huh.NewOption("Text", models.TextFieldType),
		}
		defaultSelection = models.UnsignedIntFieldType
	}

	info.Type = defaultSelection

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Field Type").
				Description("Choose the data type for this habit").
				Options(fieldTypeOptions...).
				Value(&info.Type),
		),
	)

	return form, info
}

// CreateFieldDetailsForm creates a form for configuring field-specific details
func (fb *HabitFormBuilder) CreateFieldDetailsForm(fieldType string) (*huh.Form, *FieldTypeInfo) {
	info := &FieldTypeInfo{Type: fieldType}

	var fields []huh.Field

	// Add unit configuration for numeric fields
	if isNumericField(fieldType) {
		fields = append(fields,
			huh.NewInput().
				Title("Unit (optional)").
				Description("e.g., 'steps', 'minutes', 'cups', '$'").
				Value(&info.Unit),
		)

		// Add min/max for numeric fields
		var minStr, maxStr string
		fields = append(fields,
			huh.NewInput().
				Title("Minimum value (optional)").
				Description("Lowest valid value for this field").
				Value(&minStr).
				Validate(func(s string) error {
					if s == "" {
						return nil
					}
					val, err := strconv.ParseFloat(s, 64)
					if err != nil {
						return fmt.Errorf("must be a valid number")
					}
					info.Min = &val
					return nil
				}),

			huh.NewInput().
				Title("Maximum value (optional)").
				Description("Highest valid value for this field").
				Value(&maxStr).
				Validate(func(s string) error {
					if s == "" {
						return nil
					}
					val, err := strconv.ParseFloat(s, 64)
					if err != nil {
						return fmt.Errorf("must be a valid number")
					}
					info.Max = &val
					return nil
				}),
		)
	}

	// Add multiline option for text fields
	if fieldType == models.TextFieldType {
		fields = append(fields,
			huh.NewConfirm().
				Title("Multiline text?").
				Description("Allow multiple lines of text input").
				Value(&info.Multiline),
		)
	}

	form := huh.NewForm(huh.NewGroup(fields...))
	return form, info
}

// CreateScoringForm creates a form for configuring scoring behavior
func (fb *HabitFormBuilder) CreateScoringForm(goalType models.HabitType) (*huh.Form, *ScoringInfo) {
	info := &ScoringInfo{}

	var fields []huh.Field

	if goalType == models.InformationalHabit {
		// Informational habits only need direction
		fields = append(fields,
			huh.NewSelect[string]().
				Title("Value Direction").
				Description("Indicates if higher or lower values are generally better").
				Options(
					huh.NewOption("Higher is better", "higher_better").Selected(true),
					huh.NewOption("Lower is better", "lower_better"),
					huh.NewOption("Neutral (no preference)", "neutral"),
				).
				Value(&info.Direction),
		)
		info.ScoringType = models.ManualScoring // Informational habits don't have scoring
	} else {
		// Simple and elastic habits can choose scoring type
		fields = append(fields,
			huh.NewSelect[models.ScoringType]().
				Title("Scoring Type").
				Description("How should habit achievement be determined?").
				Options(
					huh.NewOption("Manual (I'll mark completion myself)", models.ManualScoring).Selected(true),
					huh.NewOption("Automatic (Based on criteria I define)", models.AutomaticScoring),
				).
				Value(&info.ScoringType),
		)
	}

	form := huh.NewForm(huh.NewGroup(fields...))
	return form, info
}

// Helper functions

func isNumericField(fieldType string) bool {
	return fieldType == models.UnsignedIntFieldType ||
		fieldType == models.UnsignedDecimalFieldType ||
		fieldType == models.DecimalFieldType ||
		fieldType == models.DurationFieldType
}
