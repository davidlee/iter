package wizard

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"davidlee/iter/internal/models"
)

// AIDEV-NOTE: Criteria step handler pattern for elastic goal mini/midi/maxi steps
// This handler demonstrates:
// - Conditional step skipping (shouldSkip method)  
// - Level-specific configuration (level field for "mini"/"midi"/"maxi")
// - Dynamic form field generation based on goal type
// Reuse this pattern for elastic goal criteria steps with different levels

// CriteriaStepHandler handles criteria configuration for automatic scoring
type CriteriaStepHandler struct {
	form         *huh.Form
	formActive   bool
	formComplete bool
	goalType     models.GoalType
	level        string // "simple", "mini", "midi", "maxi"
	
	// Form data storage
	description    string
	comparisonType string
	value          string
	booleanValue   bool
}

// NewCriteriaStepHandler creates a new criteria step handler
func NewCriteriaStepHandler(goalType models.GoalType, level string) *CriteriaStepHandler {
	return &CriteriaStepHandler{
		goalType: goalType,
		level:    level,
	}
}

// Render renders the criteria configuration step
func (h *CriteriaStepHandler) Render(state State) string {
	// Skip if manual scoring is selected
	if h.shouldSkip(state) {
		return h.renderSkipped()
	}
	
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
			if data, ok := stepData.(*CriteriaStepData); ok {
				result := fmt.Sprintf("✅ %s Criteria Completed\n\n", h.getLevelTitle())
				
				if data.Description != "" {
					result += fmt.Sprintf("Description: %s\n", data.Description)
				}
				
				// Show criteria details
				if h.goalType == models.SimpleGoal {
					result += fmt.Sprintf("Goal achieved when: %t\n", data.BooleanValue)
				} else {
					result += fmt.Sprintf("Comparison: %s\nValue: %s\n", data.ComparisonType, data.Value)
				}
				
				result += "\nPress 'n' to continue to next step."
				return result
			}
		}
	}
	
	return "Loading criteria configuration form..."
}

// Update handles messages for the criteria step
func (h *CriteriaStepHandler) Update(msg tea.Msg, state State) (State, tea.Cmd) {
	// Skip if manual scoring is selected
	if h.shouldSkip(state) {
		return state, nil
	}
	
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

// Validate validates the criteria step data
func (h *CriteriaStepHandler) Validate(state State) []ValidationError {
	// Skip validation if manual scoring
	if h.shouldSkip(state) {
		return nil
	}
	
	var errors []ValidationError
	
	stepData := state.GetStep(h.getStepIndex())
	if stepData == nil {
		errors = append(errors, ValidationError{
			Step:    h.getStepIndex(),
			Message: fmt.Sprintf("%s criteria configuration is required for automatic scoring", h.getLevelTitle()),
		})
		return errors
	}
	
	if data, ok := stepData.(*CriteriaStepData); ok {
		// Validate criteria data
		if h.goalType != models.SimpleGoal {
			if strings.TrimSpace(data.Value) == "" {
				errors = append(errors, ValidationError{
					Step:    h.getStepIndex(),
					Field:   "value",
					Message: "Criteria value is required",
				})
			}
			
			// Validate numeric values
			if h.isNumericComparison(data.ComparisonType) {
				if _, err := strconv.ParseFloat(data.Value, 64); err != nil {
					errors = append(errors, ValidationError{
						Step:    h.getStepIndex(),
						Field:   "value",
						Message: "Value must be a valid number",
					})
				}
			}
		}
	}
	
	return errors
}

// CanNavigateFrom checks if we can leave this step
func (h *CriteriaStepHandler) CanNavigateFrom(state State) bool {
	if h.shouldSkip(state) {
		return true
	}
	return len(h.Validate(state)) == 0 && h.formComplete
}

// CanNavigateTo checks if we can enter this step
func (h *CriteriaStepHandler) CanNavigateTo(state State) bool {
	// Can navigate to criteria if previous steps are complete and automatic scoring is selected
	if h.shouldSkip(state) {
		return true
	}
	
	// Check if scoring step is complete and automatic scoring is selected
	scoringStepIndex := h.getScoringStepIndex()
	scoringData := state.GetStep(scoringStepIndex)
	if scoringData == nil {
		return false
	}
	
	if data, ok := scoringData.(*ScoringStepData); ok {
		return data.ScoringType == models.AutomaticScoring
	}
	
	return false
}

// GetTitle returns the step title
func (h *CriteriaStepHandler) GetTitle() string {
	return fmt.Sprintf("%s Criteria", h.getLevelTitle())
}

// GetDescription returns the step description
func (h *CriteriaStepHandler) GetDescription() string {
	return fmt.Sprintf("Define the criteria for %s achievement", strings.ToLower(h.getLevelTitle()))
}

// Private methods

func (h *CriteriaStepHandler) shouldSkip(state State) bool {
	scoringStepIndex := h.getScoringStepIndex()
	scoringData := state.GetStep(scoringStepIndex)
	if scoringData == nil {
		return false
	}
	
	if data, ok := scoringData.(*ScoringStepData); ok {
		return data.ScoringType == models.ManualScoring
	}
	
	return false
}

func (h *CriteriaStepHandler) renderSkipped() string {
	return fmt.Sprintf("⏭️ %s Criteria Skipped\n\nManual scoring selected - criteria not needed.\n\nPress 'n' to continue.", h.getLevelTitle())
}

func (h *CriteriaStepHandler) getStepIndex() int {
	switch h.goalType {
	case models.SimpleGoal:
		return 2 // basic_info(0) -> scoring(1) -> criteria(2)
	case models.ElasticGoal:
		switch h.level {
		case "mini":
			return 3 // basic_info(0) -> field_config(1) -> scoring(2) -> mini(3)
		case "midi":
			return 4 // basic_info(0) -> field_config(1) -> scoring(2) -> mini(3) -> midi(4)
		case "maxi":
			return 5 // basic_info(0) -> field_config(1) -> scoring(2) -> mini(3) -> midi(4) -> maxi(5)
		}
	}
	return 2
}

func (h *CriteriaStepHandler) getScoringStepIndex() int {
	switch h.goalType {
	case models.SimpleGoal:
		return 1
	case models.ElasticGoal:
		return 2
	default:
		return 1
	}
}

func (h *CriteriaStepHandler) getLevelTitle() string {
	switch h.level {
	case "simple":
		return "Simple"
	case "mini":
		return "Mini"
	case "midi":
		return "Midi"
	case "maxi":
		return "Maxi"
	default:
		if len(h.level) > 0 {
			return strings.ToUpper(string(h.level[0])) + h.level[1:]
		}
		return h.level
	}
}

func (h *CriteriaStepHandler) initializeForm(state State) {
	// Get existing data if available
	if stepData := state.GetStep(h.getStepIndex()); stepData != nil {
		if data, ok := stepData.(*CriteriaStepData); ok {
			h.description = data.Description
			h.comparisonType = data.ComparisonType
			h.value = data.Value
			h.booleanValue = data.BooleanValue
		}
	} else {
		// Set defaults
		h.booleanValue = true
	}
	
	var fields []huh.Field
	
	// Description field
	fields = append(fields,
		huh.NewInput().
			Title(fmt.Sprintf("%s Level Description (optional)", h.getLevelTitle())).
			Description(fmt.Sprintf("Describe what %s achievement means", strings.ToLower(h.getLevelTitle()))).
			Value(&h.description),
	)
	
	if h.goalType == models.SimpleGoal {
		// Simple boolean criteria
		fields = append(fields,
			huh.NewConfirm().
				Title("Achievement Criteria").
				Description("Goal is achieved when the answer is 'Yes'").
				Value(&h.booleanValue),
		)
	} else {
		// Get field type - for now default to unsigned int for simple goals
		fieldType := models.UnsignedIntFieldType
		
		// Comparison type selection
		fields = append(fields,
			huh.NewSelect[string]().
				Title("Comparison Type").
				Description("How should the value be compared?").
				Options(h.getComparisonOptions(fieldType)...).
				Value(&h.comparisonType),
		)
		
		// Value input
		fields = append(fields,
			huh.NewInput().
				Title("Value").
				Description(h.getValueDescription(fieldType)).
				Value(&h.value).
				Validate(h.createValueValidator(fieldType)),
		)
	}
	
	// Create form
	h.form = huh.NewForm(huh.NewGroup(fields...))
	h.formActive = true
	h.formComplete = false
}

func (h *CriteriaStepHandler) extractFormData(state State) {
	// Create step data from stored form values
	stepData := &CriteriaStepData{
		Level:          h.level,
		Description:    strings.TrimSpace(h.description),
		ComparisonType: h.comparisonType,
		Value:          strings.TrimSpace(h.value),
		BooleanValue:   h.booleanValue,
	}
	
	// Store in state
	state.SetStep(h.getStepIndex(), stepData)
}

func (h *CriteriaStepHandler) getComparisonOptions(_ string) []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("Greater than or equal to", "gte"),
		huh.NewOption("Greater than", "gt"),
		huh.NewOption("Less than or equal to", "lte"),
		huh.NewOption("Less than", "lt"),
	}
}

func (h *CriteriaStepHandler) getValueDescription(_ string) string {
	return "Enter a positive whole number"
}

func (h *CriteriaStepHandler) createValueValidator(_ string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("value cannot be empty")
		}

		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("must be a whole number")
		}
		if val < 0 {
			return fmt.Errorf("must be a positive number")
		}

		return nil
	}
}

func (h *CriteriaStepHandler) isNumericComparison(comparisonType string) bool {
	return comparisonType == "gt" || comparisonType == "gte" || 
		   comparisonType == "lt" || comparisonType == "lte"
}