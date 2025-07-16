package wizard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"davidlee/vice/internal/models"
)

// AIDEV-NOTE: Step handler implementation pattern for habit flows
// Follow this pattern for elastic/informational habit step handlers:
// 1. Embed form state directly in handler struct (avoid form introspection)
// 2. Use pointer binding in huh forms (&h.fieldName)
// 3. Implement full StepHandler interface with proper validation
// 4. Handle conditional logic in CanNavigateTo/shouldSkip methods

// BasicInfoStepHandler handles basic habit information collection
type BasicInfoStepHandler struct {
	form         *huh.Form
	formActive   bool
	formComplete bool
	goalType     models.HabitType

	// Form data storage
	title       string
	description string
}

// NewBasicInfoStepHandler creates a new basic info step handler
func NewBasicInfoStepHandler(goalType models.HabitType) *BasicInfoStepHandler {
	return &BasicInfoStepHandler{
		goalType: goalType,
	}
}

// Render renders the basic information step
func (h *BasicInfoStepHandler) Render(state State) string {
	if h.form == nil {
		h.initializeForm(state)
	}

	// Render the form content
	if h.formActive {
		return h.form.View()
	}

	// Show completed state
	if h.formComplete {
		if stepData := state.GetStep(0); stepData != nil {
			if data, ok := stepData.(*BasicInfoStepData); ok {
				return fmt.Sprintf(`✅ Basic Information Completed

Title: %s
Description: %s
Habit Type: %s

Press 'n' to continue to next step.`,
					data.Title,
					data.Description,
					data.HabitType)
			}
		}
	}

	return "Loading basic information form..."
}

// Update handles messages for the basic info step
func (h *BasicInfoStepHandler) Update(msg tea.Msg, state State) (State, tea.Cmd) {
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

// Validate validates the basic info step data
func (h *BasicInfoStepHandler) Validate(state State) []ValidationError {
	var errors []ValidationError

	stepData := state.GetStep(0)
	if stepData == nil {
		errors = append(errors, ValidationError{
			Step:    0,
			Message: "Basic information is required",
		})
		return errors
	}

	if data, ok := stepData.(*BasicInfoStepData); ok {
		if strings.TrimSpace(data.Title) == "" {
			errors = append(errors, ValidationError{
				Step:    0,
				Field:   "title",
				Message: "Habit title is required",
			})
		}

		if len(data.Title) > 100 {
			errors = append(errors, ValidationError{
				Step:    0,
				Field:   "title",
				Message: "Habit title must be 100 characters or less",
			})
		}
	}

	return errors
}

// CanNavigateFrom checks if we can leave this step
func (h *BasicInfoStepHandler) CanNavigateFrom(state State) bool {
	return len(h.Validate(state)) == 0 && h.formComplete
}

// CanNavigateTo checks if we can enter this step
func (h *BasicInfoStepHandler) CanNavigateTo(_ State) bool {
	return true // Can always enter basic info step
}

// GetTitle returns the step title
func (h *BasicInfoStepHandler) GetTitle() string {
	return "Basic Information"
}

// GetDescription returns the step description
func (h *BasicInfoStepHandler) GetDescription() string {
	return "Enter the basic details for your habit"
}

// Private methods

func (h *BasicInfoStepHandler) initializeForm(state State) {
	// Get existing data if available
	if stepData := state.GetStep(0); stepData != nil {
		if data, ok := stepData.(*BasicInfoStepData); ok {
			h.title = data.Title
			h.description = data.Description
		}
	}

	// Create form
	h.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Habit Title").
				Description("Enter a descriptive name for your habit").
				Value(&h.title).
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
				Value(&h.description),
		),
	)

	h.formActive = true
	h.formComplete = false
}

func (h *BasicInfoStepHandler) extractFormData(state State) {
	// Create step data from stored form values
	stepData := &BasicInfoStepData{
		Title:       strings.TrimSpace(h.title),
		Description: strings.TrimSpace(h.description),
		HabitType:   h.goalType,
	}

	// Store in state
	state.SetStep(0, stepData)
}

// ScoringStepHandler handles scoring configuration for habits
type ScoringStepHandler struct {
	form         *huh.Form
	formActive   bool
	formComplete bool
	goalType     models.HabitType

	// Form data storage
	scoringType models.ScoringType
	direction   string
}

// NewScoringStepHandler creates a new scoring step handler
func NewScoringStepHandler(goalType models.HabitType) *ScoringStepHandler {
	return &ScoringStepHandler{
		goalType: goalType,
	}
}

// Render renders the scoring configuration step
func (h *ScoringStepHandler) Render(state State) string {
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
			if data, ok := stepData.(*ScoringStepData); ok {
				result := fmt.Sprintf("✅ Scoring Configuration Completed\n\nScoring Type: %s", data.ScoringType)

				if h.goalType == models.InformationalHabit && data.Direction != "" {
					result += fmt.Sprintf("\nDirection: %s", data.Direction)
				}

				result += "\n\nPress 'n' to continue to next step."
				return result
			}
		}
	}

	return "Loading scoring configuration form..."
}

// Update handles messages for the scoring step
func (h *ScoringStepHandler) Update(msg tea.Msg, state State) (State, tea.Cmd) {
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

// Validate validates the scoring step data
func (h *ScoringStepHandler) Validate(state State) []ValidationError {
	var errors []ValidationError

	stepData := state.GetStep(h.getStepIndex())
	if stepData == nil {
		// Only show validation error if form has been attempted (form is complete)
		// This prevents showing "required" errors when just starting the step
		if h.formComplete {
			errors = append(errors, ValidationError{
				Step:    h.getStepIndex(),
				Message: "Scoring configuration is required",
			})
		}
		return errors
	}

	// Scoring step is always valid once data is present
	return errors
}

// CanNavigateFrom checks if we can leave this step
func (h *ScoringStepHandler) CanNavigateFrom(state State) bool {
	return len(h.Validate(state)) == 0 && h.formComplete
}

// CanNavigateTo checks if we can enter this step
func (h *ScoringStepHandler) CanNavigateTo(state State) bool {
	// Can navigate to scoring if basic info is complete
	basicInfoErrors := h.validateBasicInfo(state)
	return len(basicInfoErrors) == 0
}

// GetTitle returns the step title
func (h *ScoringStepHandler) GetTitle() string {
	return "Scoring Configuration"
}

// GetDescription returns the step description
func (h *ScoringStepHandler) GetDescription() string {
	switch h.goalType {
	case models.SimpleHabit:
		return "Choose how this habit will be scored"
	case models.ElasticHabit:
		return "Choose how achievement levels will be determined"
	case models.InformationalHabit:
		return "Configure data collection preferences"
	default:
		return "Configure habit scoring"
	}
}

// Private methods for ScoringStepHandler

func (h *ScoringStepHandler) getStepIndex() int {
	switch h.goalType {
	case models.SimpleHabit:
		return 1 // basic_info(0) -> scoring(1)
	case models.ElasticHabit:
		return 2 // basic_info(0) -> field_config(1) -> scoring(2)
	case models.InformationalHabit:
		return 1 // basic_info(0) -> scoring(1) (but direction only)
	default:
		return 1
	}
}

func (h *ScoringStepHandler) initializeForm(state State) {
	// Get existing data if available
	if stepData := state.GetStep(h.getStepIndex()); stepData != nil {
		if data, ok := stepData.(*ScoringStepData); ok {
			h.scoringType = data.ScoringType
			h.direction = data.Direction
		}
	}

	var fields []huh.Field

	if h.goalType == models.InformationalHabit {
		// Informational habits only need direction
		fields = append(fields,
			huh.NewSelect[string]().
				Title("Value Direction").
				Description("Indicates if higher or lower values are generally better").
				Options(
					huh.NewOption("Higher is better", "higher_better"),
					huh.NewOption("Lower is better", "lower_better"),
					huh.NewOption("Neutral (no preference)", "neutral"),
				).
				Value(&h.direction),
		)
	} else {
		// Simple and elastic habits can choose scoring type
		fields = append(fields,
			huh.NewSelect[models.ScoringType]().
				Title("Scoring Type").
				Description("How should habit achievement be determined?").
				Options(
					huh.NewOption("Manual (I'll mark completion myself)", models.ManualScoring),
					huh.NewOption("Automatic (Based on criteria I define)", models.AutomaticScoring),
				).
				Value(&h.scoringType),
		)
	}

	// Create form
	h.form = huh.NewForm(huh.NewGroup(fields...))
	h.formActive = true
	h.formComplete = false
}

func (h *ScoringStepHandler) extractFormData(state State) {
	// Create step data from stored form values
	stepData := &ScoringStepData{
		ScoringType: h.scoringType,
		Direction:   h.direction,
	}

	// For informational habits, always set to manual scoring
	if h.goalType == models.InformationalHabit {
		stepData.ScoringType = models.ManualScoring
	}

	// Store in state
	state.SetStep(h.getStepIndex(), stepData)
}

func (h *ScoringStepHandler) validateBasicInfo(state State) []ValidationError {
	var errors []ValidationError

	stepData := state.GetStep(0)
	if stepData == nil {
		errors = append(errors, ValidationError{
			Step:    0,
			Message: "Basic information must be completed first",
		})
		return errors
	}

	if data, ok := stepData.(*BasicInfoStepData); ok {
		if data.Title == "" {
			errors = append(errors, ValidationError{
				Step:    0,
				Field:   "title",
				Message: "Habit title is required",
			})
		}
	}

	return errors
}

// ConfirmationStepHandler handles the final confirmation step
type ConfirmationStepHandler struct {
	form         *huh.Form
	formActive   bool
	formComplete bool
	goalType     models.HabitType
	confirmed    bool
}

// NewConfirmationStepHandler creates a new confirmation step handler
func NewConfirmationStepHandler(goalType models.HabitType) *ConfirmationStepHandler {
	return &ConfirmationStepHandler{
		goalType: goalType,
	}
}

// Render renders the confirmation step
func (h *ConfirmationStepHandler) Render(state State) string {
	if h.form == nil {
		h.initializeForm(state)
	}

	// Render the form content
	if h.formActive {
		return h.form.View()
	}

	// Show completed state
	if h.formComplete {
		if h.confirmed {
			return "✅ Habit Configuration Confirmed\n\nPress 'f' to finish and create the habit."
		}
		return "❌ Habit Creation Cancelled\n\nPress 'b' to go back or Ctrl+C to exit."
	}

	return "Loading confirmation form..."
}

// Update handles messages for the confirmation step
func (h *ConfirmationStepHandler) Update(msg tea.Msg, state State) (State, tea.Cmd) {
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

				// Extract confirmation choice
				h.extractFormData(state)
			}
		}
		return state, cmd
	}

	return state, nil
}

// Validate validates the confirmation step data
func (h *ConfirmationStepHandler) Validate(_ State) []ValidationError {
	// Confirmation step doesn't need validation beyond completion
	return nil
}

// CanNavigateFrom checks if we can leave this step
func (h *ConfirmationStepHandler) CanNavigateFrom(_ State) bool {
	return h.formComplete
}

// CanNavigateTo checks if we can enter this step
func (h *ConfirmationStepHandler) CanNavigateTo(state State) bool {
	// Can navigate to confirmation if all previous steps are valid
	for i := 0; i < h.getStepIndex(); i++ {
		stepData := state.GetStep(i)
		if stepData == nil {
			return false
		}
	}
	return true
}

// GetTitle returns the step title
func (h *ConfirmationStepHandler) GetTitle() string {
	return "Confirmation"
}

// GetDescription returns the step description
func (h *ConfirmationStepHandler) GetDescription() string {
	return "Review and confirm your habit configuration"
}

// IsConfirmed returns whether the user confirmed habit creation
func (h *ConfirmationStepHandler) IsConfirmed() bool {
	return h.confirmed
}

// Private methods for ConfirmationStepHandler

func (h *ConfirmationStepHandler) getStepIndex() int {
	switch h.goalType {
	case models.SimpleHabit:
		return 3 // basic_info(0) -> scoring(1) -> criteria(2) -> confirmation(3)
	case models.ElasticHabit:
		return 7 // basic_info(0) -> field_config(1) -> scoring(2) -> mini(3) -> midi(4) -> maxi(5) -> validation(6) -> confirmation(7)
	case models.InformationalHabit:
		return 2 // basic_info(0) -> field_config(1) -> confirmation(2)
	default:
		return 3
	}
}

func (h *ConfirmationStepHandler) initializeForm(state State) {
	// Generate habit preview
	preview := h.generateHabitPreview(state)

	var confirmed bool

	// Create confirmation form
	h.form = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Habit Preview").
				Description(preview),

			huh.NewConfirm().
				Title("Create this habit?").
				Description("Confirm to create the habit with the above configuration").
				Value(&confirmed),
		),
	)

	h.formActive = true
	h.formComplete = false
}

func (h *ConfirmationStepHandler) extractFormData(_ State) {
	// The confirmed value is already stored directly in h.confirmed
	// due to the pointer binding in the form
}

func (h *ConfirmationStepHandler) generateHabitPreview(state State) string {
	var preview strings.Builder

	// Basic info
	if basicData := state.GetStep(0); basicData != nil {
		if data, ok := basicData.(*BasicInfoStepData); ok {
			preview.WriteString(fmt.Sprintf("Title: %s\n", data.Title))
			if data.Description != "" {
				preview.WriteString(fmt.Sprintf("Description: %s\n", data.Description))
			}
			preview.WriteString(fmt.Sprintf("Type: %s\n\n", data.HabitType))
		}
	}

	// Scoring configuration
	var scoringStepIndex int
	switch h.goalType {
	case models.SimpleHabit:
		scoringStepIndex = 1
	case models.ElasticHabit:
		scoringStepIndex = 2
	case models.InformationalHabit:
		scoringStepIndex = 1
	}

	if scoringData := state.GetStep(scoringStepIndex); scoringData != nil {
		if data, ok := scoringData.(*ScoringStepData); ok {
			if h.goalType == models.InformationalHabit {
				preview.WriteString(fmt.Sprintf("Direction: %s\n", data.Direction))
			} else {
				preview.WriteString(fmt.Sprintf("Scoring: %s\n", data.ScoringType))
			}
		}
	}

	// Criteria (if automatic scoring)
	if h.goalType == models.SimpleHabit {
		if criteriaData := state.GetStep(2); criteriaData != nil {
			if data, ok := criteriaData.(*CriteriaStepData); ok {
				if data.Description != "" {
					preview.WriteString(fmt.Sprintf("Criteria: %s\n", data.Description))
				}
				preview.WriteString(fmt.Sprintf("Achievement: %t\n", data.BooleanValue))
			}
		}
	}

	return preview.String()
}
