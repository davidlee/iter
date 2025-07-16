package wizard

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"davidlee/vice/internal/models"
)

// AIDEV-NOTE: Validation step handler for elastic habits - validates mini ≤ midi ≤ maxi constraints
// This handler demonstrates:
// - Cross-step validation logic (checking multiple previous steps)
// - Complex business rule validation (ordering constraints)
// - Interactive validation with user feedback and correction options
// - Preview of complete habit configuration before final confirmation

// ValidationStepHandler handles validation of elastic habit criteria constraints
type ValidationStepHandler struct {
	form             *huh.Form
	formActive       bool
	formComplete     bool
	goalType         models.HabitType
	validationOk     bool
	validationErrors []string

	// User choice for handling validation errors
	action string // "fix", "proceed", "cancel"
}

// NewValidationStepHandler creates a new validation step handler
func NewValidationStepHandler(goalType models.HabitType) *ValidationStepHandler {
	return &ValidationStepHandler{
		goalType: goalType,
	}
}

// Render renders the validation step
func (h *ValidationStepHandler) Render(state State) string {
	if h.form == nil {
		h.initializeForm(state)
	}

	// Render the form content
	if h.formActive {
		return h.form.View()
	}

	// Show completed state
	if h.formComplete {
		if h.validationOk {
			return "✅ Validation Passed\n\nAll criteria constraints are valid.\nMini ≤ Midi ≤ Maxi relationship confirmed.\n\nPress 'n' to continue to final confirmation."
		}
		result := "⚠️ Validation Issues Found\n\n"
		for _, err := range h.validationErrors {
			result += fmt.Sprintf("• %s\n", err)
		}
		result += "\nValidation completed with issues.\nPress 'n' to continue to final confirmation."
		return result
	}

	return "Running validation..."
}

// Update handles messages for the validation step
func (h *ValidationStepHandler) Update(msg tea.Msg, state State) (State, tea.Cmd) {
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

				// Handle user's choice
				h.handleValidationChoice(state)
			}
		}
		return state, cmd
	}

	return state, nil
}

// Validate validates the validation step (meta!)
func (h *ValidationStepHandler) Validate(_ State) []ValidationError {
	// The validation step itself doesn't need validation
	// It validates other steps
	return nil
}

// CanNavigateFrom checks if we can leave this step
func (h *ValidationStepHandler) CanNavigateFrom(_ State) bool {
	return h.formComplete
}

// CanNavigateTo checks if we can enter this step
func (h *ValidationStepHandler) CanNavigateTo(state State) bool {
	// Can navigate to validation if all criteria steps are complete
	if h.goalType != models.ElasticHabit {
		return true // Skip validation for non-elastic habits
	}

	// Check that all three criteria steps (mini, midi, maxi) are complete
	miniData := state.GetStep(3) // mini criteria step
	midiData := state.GetStep(4) // midi criteria step
	maxiData := state.GetStep(5) // maxi criteria step

	return miniData != nil && midiData != nil && maxiData != nil
}

// GetTitle returns the step title
func (h *ValidationStepHandler) GetTitle() string {
	return "Validation"
}

// GetDescription returns the step description
func (h *ValidationStepHandler) GetDescription() string {
	return "Validate habit configuration and criteria constraints"
}

// Private methods

func (h *ValidationStepHandler) initializeForm(state State) {
	// Run validation logic
	h.runValidation(state)

	var fields []huh.Field

	if h.validationOk {
		// Validation passed - just show confirmation
		fields = append(fields,
			huh.NewConfirm().
				Title("Validation Passed").
				Description("All criteria constraints are valid. Continue to final confirmation?").
				Affirmative("Continue").
				Negative("Review"),
		)
	} else {
		// Validation failed - show options
		validationSummary := "Validation found issues:\n"
		for _, err := range h.validationErrors {
			validationSummary += fmt.Sprintf("• %s\n", err)
		}

		fields = append(fields,
			huh.NewSelect[string]().
				Title("Validation Issues Found").
				Description(validationSummary+"\nHow would you like to proceed?").
				Options(
					huh.NewOption("Continue anyway (manual review)", "proceed"),
					huh.NewOption("Go back to fix issues", "fix"),
					huh.NewOption("Cancel habit creation", "cancel"),
				).
				Value(&h.action),
		)
	}

	// Create form
	h.form = huh.NewForm(huh.NewGroup(fields...))
	h.formActive = true
	h.formComplete = false
}

func (h *ValidationStepHandler) runValidation(state State) {
	h.validationErrors = []string{}
	h.validationOk = true

	if h.goalType != models.ElasticHabit {
		return // No validation needed for non-elastic habits
	}

	// Get criteria data
	miniData := state.GetStep(3)
	midiData := state.GetStep(4)
	maxiData := state.GetStep(5)

	if miniData == nil || midiData == nil || maxiData == nil {
		h.validationErrors = append(h.validationErrors, "Missing criteria configuration")
		h.validationOk = false
		return
	}

	// Cast to criteria data
	mini, miniOk := miniData.(*CriteriaStepData)
	midi, midiOk := midiData.(*CriteriaStepData)
	maxi, maxiOk := maxiData.(*CriteriaStepData)

	if !miniOk || !midiOk || !maxiOk {
		h.validationErrors = append(h.validationErrors, "Invalid criteria data format")
		h.validationOk = false
		return
	}

	// Check if automatic scoring is being used
	scoringData := state.GetStep(2)
	if scoringData != nil {
		if scoring, ok := scoringData.(*ScoringStepData); ok {
			if scoring.ScoringType == models.ManualScoring {
				return // No validation needed for manual scoring
			}
		}
	}

	// Validate mini ≤ midi ≤ maxi constraints
	errors := h.validateCriteriaOrder(mini, midi, maxi)
	if len(errors) > 0 {
		h.validationErrors = append(h.validationErrors, errors...)
		h.validationOk = false
	}

	// Validate individual criteria
	errors = h.validateIndividualCriteria(mini, midi, maxi)
	if len(errors) > 0 {
		h.validationErrors = append(h.validationErrors, errors...)
		h.validationOk = false
	}
}

func (h *ValidationStepHandler) validateCriteriaOrder(mini, midi, maxi *CriteriaStepData) []string {
	var errors []string

	// Convert values to numbers for comparison
	miniVal, miniErr := h.parseNumericValue(mini.Value)
	midiVal, midiErr := h.parseNumericValue(midi.Value)
	maxiVal, maxiErr := h.parseNumericValue(maxi.Value)

	if miniErr != nil {
		errors = append(errors, fmt.Sprintf("Mini criteria value is not numeric: %s", mini.Value))
	}
	if midiErr != nil {
		errors = append(errors, fmt.Sprintf("Midi criteria value is not numeric: %s", midi.Value))
	}
	if maxiErr != nil {
		errors = append(errors, fmt.Sprintf("Maxi criteria value is not numeric: %s", maxi.Value))
	}

	// If we have valid numbers, check ordering
	if miniErr == nil && midiErr == nil && maxiErr == nil {
		if miniVal > midiVal {
			errors = append(errors, fmt.Sprintf("Mini value (%.2f) cannot be greater than Midi value (%.2f)", miniVal, midiVal))
		}
		if midiVal > maxiVal {
			errors = append(errors, fmt.Sprintf("Midi value (%.2f) cannot be greater than Maxi value (%.2f)", midiVal, maxiVal))
		}
		if miniVal > maxiVal {
			errors = append(errors, fmt.Sprintf("Mini value (%.2f) cannot be greater than Maxi value (%.2f)", miniVal, maxiVal))
		}
	}

	return errors
}

func (h *ValidationStepHandler) validateIndividualCriteria(mini, midi, maxi *CriteriaStepData) []string {
	var errors []string

	// Check that comparison types are consistent
	if mini.ComparisonType != midi.ComparisonType || midi.ComparisonType != maxi.ComparisonType {
		errors = append(errors, "All criteria levels should use the same comparison type")
	}

	// Check for empty descriptions (warning)
	if strings.TrimSpace(mini.Description) == "" {
		errors = append(errors, "Mini level description is empty")
	}
	if strings.TrimSpace(midi.Description) == "" {
		errors = append(errors, "Midi level description is empty")
	}
	if strings.TrimSpace(maxi.Description) == "" {
		errors = append(errors, "Maxi level description is empty")
	}

	return errors
}

func (h *ValidationStepHandler) parseNumericValue(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}
	return strconv.ParseFloat(value, 64)
}

func (h *ValidationStepHandler) handleValidationChoice(_ State) {
	// Store the user's choice for navigation decisions
	// The navigation controller will use this to determine next steps

	// Action values: "fix" (go back), "cancel" (abort), "proceed" (continue)
	// Navigation will handle the actual routing based on h.action
	_ = h.action // Action is stored for navigation controller to use
}
