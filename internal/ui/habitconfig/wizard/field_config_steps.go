package wizard

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"davidlee/vice/internal/models"
)

// AIDEV-NOTE: Field configuration step handler for elastic and informational habits
// This handler demonstrates:
// - Dynamic form field generation based on selected field type
// - Field type-specific validation and configuration options
// - Unit, constraints, and direction configuration for complex field types
// Reuse this pattern for other configuration steps that need dynamic forms

// FieldConfigStepHandler handles field type configuration for elastic and informational habits
type FieldConfigStepHandler struct {
	form         *huh.Form
	formActive   bool
	formComplete bool
	habitType    models.HabitType

	// Form data storage
	fieldType string
	unit      string
	minValue  string
	maxValue  string
	multiline bool
	direction string // For informational habits
}

// NewFieldConfigStepHandler creates a new field configuration step handler
func NewFieldConfigStepHandler(habitType models.HabitType) *FieldConfigStepHandler {
	return &FieldConfigStepHandler{
		habitType: habitType,
	}
}

// Render renders the field configuration step
func (h *FieldConfigStepHandler) Render(state State) string {
	if h.form == nil {
		h.initializeForm(state)
	}

	// Render the form content
	if h.formActive {
		return h.form.View()
	}

	// Show completed state
	if h.formComplete {
		if stepData := state.GetStep(h.getStepIndex()); stepData != nil {
			if data, ok := stepData.(*FieldConfigStepData); ok {
				result := "✅ Field Configuration Completed\n\n"

				result += fmt.Sprintf("Field Type: %s\n", h.getFieldTypeDescription(data.FieldType))

				if data.Unit != "" {
					result += fmt.Sprintf("Unit: %s\n", data.Unit)
				}

				if data.Min != nil {
					result += fmt.Sprintf("Minimum: %.2f\n", *data.Min)
				}

				if data.Max != nil {
					result += fmt.Sprintf("Maximum: %.2f\n", *data.Max)
				}

				if data.Multiline {
					result += "Multi-line: Yes\n"
				}

				if h.habitType == models.InformationalHabit && h.direction != "" {
					result += fmt.Sprintf("Direction: %s\n", h.direction)
				}

				result += "\nPress 'n' to continue to next step."
				return result
			}
		}
	}

	return "Loading field configuration form..."
}

// Update handles messages for the field configuration step
func (h *FieldConfigStepHandler) Update(msg tea.Msg, state State) (State, tea.Cmd) {
	if h.form == nil {
		h.initializeForm(state)
	}

	if h.formActive && h.form != nil {
		form, cmd := h.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			h.form = f

			// Check if form is completed
			if h.form.State == huh.StateCompleted {
				h.formActive = false
				h.formComplete = true

				// Extract data and store in state
				h.extractFormData(state)
			}
		}
		return state, cmd
	}

	return state, nil
}

// Validate validates the field configuration step data
func (h *FieldConfigStepHandler) Validate(state State) []ValidationError {
	var errors []ValidationError

	stepData := state.GetStep(h.getStepIndex())
	if stepData == nil {
		// Only show validation error if form has been attempted (form is complete)
		// This prevents showing "required" errors when just starting the step
		if h.formComplete {
			errors = append(errors, ValidationError{
				Step:    h.getStepIndex(),
				Message: "Field configuration is required",
			})
		}
		return errors
	}

	if data, ok := stepData.(*FieldConfigStepData); ok {
		// Validate field type is selected
		if strings.TrimSpace(data.FieldType) == "" {
			errors = append(errors, ValidationError{
				Step:    h.getStepIndex(),
				Field:   "fieldType",
				Message: "Field type selection is required",
			})
		}

		// Validate min/max constraints
		if data.Min != nil && data.Max != nil {
			if *data.Min > *data.Max {
				errors = append(errors, ValidationError{
					Step:    h.getStepIndex(),
					Field:   "constraints",
					Message: "Minimum value cannot be greater than maximum value",
				})
			}
		}
	}

	return errors
}

// CanNavigateFrom checks if we can leave this step
func (h *FieldConfigStepHandler) CanNavigateFrom(state State) bool {
	return len(h.Validate(state)) == 0 && h.formComplete
}

// CanNavigateTo checks if we can enter this step
func (h *FieldConfigStepHandler) CanNavigateTo(state State) bool {
	// Can navigate to field config if basic info is complete
	basicInfoData := state.GetStep(0)
	if basicInfoData == nil {
		return false
	}

	if data, ok := basicInfoData.(*BasicInfoStepData); ok {
		return data.IsValid()
	}

	return false
}

// GetTitle returns the step title
func (h *FieldConfigStepHandler) GetTitle() string {
	return "Field Configuration"
}

// GetDescription returns the step description
func (h *FieldConfigStepHandler) GetDescription() string {
	return "Configure the data type and constraints for your habit"
}

// Private methods

func (h *FieldConfigStepHandler) getStepIndex() int {
	switch h.habitType {
	case models.ElasticHabit:
		return 1 // basic_info(0) -> field_config(1)
	case models.InformationalHabit:
		return 1 // basic_info(0) -> field_config(1)
	default:
		return 1
	}
}

func (h *FieldConfigStepHandler) initializeForm(state State) {
	// Get existing data if available
	if stepData := state.GetStep(h.getStepIndex()); stepData != nil {
		if data, ok := stepData.(*FieldConfigStepData); ok {
			h.fieldType = data.FieldType
			h.unit = data.Unit
			if data.Min != nil {
				h.minValue = fmt.Sprintf("%.2f", *data.Min)
			}
			if data.Max != nil {
				h.maxValue = fmt.Sprintf("%.2f", *data.Max)
			}
			h.multiline = data.Multiline
			h.direction = data.Direction
		}
	}

	var fields []huh.Field

	// Field type selection
	fields = append(fields,
		huh.NewSelect[string]().
			Title("Field Type").
			Description("What type of data will you track for this habit?").
			Options(h.getFieldTypeOptions()...).
			Value(&h.fieldType),
	)

	// Unit input (for numeric types)
	fields = append(fields,
		huh.NewInput().
			Title("Unit (optional)").
			Description("e.g., 'minutes', 'pages', 'km', 'calories'").
			Value(&h.unit),
	)

	// Min/Max constraints for numeric types
	fields = append(fields,
		huh.NewInput().
			Title("Minimum Value (optional)").
			Description("Minimum allowed value for validation").
			Value(&h.minValue).
			Validate(h.createNumericValidator("minimum")),
	)

	fields = append(fields,
		huh.NewInput().
			Title("Maximum Value (optional)").
			Description("Maximum allowed value for validation").
			Value(&h.maxValue).
			Validate(h.createNumericValidator("maximum")),
	)

	// Multi-line option for text fields
	if h.habitType == models.InformationalHabit {
		fields = append(fields,
			huh.NewConfirm().
				Title("Multi-line Text?").
				Description("Allow multiple lines of text input").
				Value(&h.multiline),
		)

		// Direction for informational habits
		fields = append(fields,
			huh.NewSelect[string]().
				Title("Direction").
				Description("How should this data be interpreted?").
				Options(
					huh.NewOption("Higher is better", "higher"),
					huh.NewOption("Lower is better", "lower"),
					huh.NewOption("Neutral/tracking only", "neutral"),
				).
				Value(&h.direction),
		)
	}

	// Create form
	h.form = huh.NewForm(huh.NewGroup(fields...))
	h.formActive = true
	h.formComplete = false
}

func (h *FieldConfigStepHandler) extractFormData(state State) {
	// Parse min/max values
	var minPtr, maxPtr *float64

	if strings.TrimSpace(h.minValue) != "" {
		if val, err := strconv.ParseFloat(h.minValue, 64); err == nil {
			minPtr = &val
		}
	}

	if strings.TrimSpace(h.maxValue) != "" {
		if val, err := strconv.ParseFloat(h.maxValue, 64); err == nil {
			maxPtr = &val
		}
	}

	// Create step data from stored form values
	stepData := &FieldConfigStepData{
		FieldType: h.fieldType,
		Unit:      strings.TrimSpace(h.unit),
		Min:       minPtr,
		Max:       maxPtr,
		Multiline: h.multiline,
		Direction: h.direction, // Store direction for informational habits
		valid:     true,
	}

	// Store in state
	state.SetStep(h.getStepIndex(), stepData)
}

func (h *FieldConfigStepHandler) getFieldTypeOptions() []huh.Option[string] {
	if h.habitType == models.ElasticHabit {
		return []huh.Option[string]{
			huh.NewOption("Whole numbers (0, 1, 2, ...)", models.UnsignedIntFieldType),
			huh.NewOption("Decimal numbers (0.5, 1.25, ...)", models.DecimalFieldType),
			huh.NewOption("Duration (minutes, hours)", models.DurationFieldType),
			huh.NewOption("Time (9:30 AM)", models.TimeFieldType),
		}
	}

	// Informational habits support all field types
	return []huh.Option[string]{
		huh.NewOption("Text/Comment", models.TextFieldType),
		huh.NewOption("Yes/No", models.BooleanFieldType),
		huh.NewOption("Whole numbers (0, 1, 2, ...)", models.UnsignedIntFieldType),
		huh.NewOption("Decimal numbers (0.5, 1.25, ...)", models.DecimalFieldType),
		huh.NewOption("Duration (minutes, hours)", models.DurationFieldType),
		huh.NewOption("Time (9:30 AM)", models.TimeFieldType),
	}
}

func (h *FieldConfigStepHandler) getFieldTypeDescription(fieldType string) string {
	switch fieldType {
	case models.TextFieldType:
		return "Text/Comment"
	case models.BooleanFieldType:
		return "Yes/No"
	case models.UnsignedIntFieldType:
		return "Whole numbers"
	case models.DecimalFieldType:
		return "Decimal numbers"
	case models.DurationFieldType:
		return "Duration"
	case models.TimeFieldType:
		return "Time"
	default:
		return fieldType
	}
}

func (h *FieldConfigStepHandler) createNumericValidator(fieldName string) func(string) error {
	return func(s string) error {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil // Optional field
		}

		_, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("%s must be a valid number", fieldName)
		}

		return nil
	}
}
