package goalconfig

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"davidlee/iter/internal/models"
)

// AIDEV-NOTE: Simple idiomatic bubbletea implementation for goal creation
// Based on https://github.com/charmbracelet/huh/blob/main/examples/bubbletea/main.go
// Follows Model-View-Update pattern from https://github.com/charmbracelet/bubbletea
// Much simpler than complex wizard architecture - focuses on common use case

// SimpleGoalCreator implements a simple, idiomatic bubbletea model for creating goals
type SimpleGoalCreator struct {
	form     *huh.Form
	quitting bool
	err      error
	result   *models.Goal

	// Pre-populated basic info
	title       string
	description string
	goalType    models.GoalType

	// Field configuration data - bound directly to form fields per huh documentation
	selectedFieldType string
	numericSubtype    string
	unit              string
	multilineText     bool
	minValue          string
	maxValue          string
	hasMinMax         bool
	scoringType       models.ScoringType
	prompt            string
	comment           string

	// State tracking for multi-step flow
	currentStep int
	maxSteps    int
}

// NewSimpleGoalCreator creates a new simple goal creator with pre-populated basic info
func NewSimpleGoalCreator(title, description string, goalType models.GoalType) *SimpleGoalCreator {
	creator := &SimpleGoalCreator{
		title:             title,
		description:       description,
		goalType:          goalType,
		selectedFieldType: models.BooleanFieldType, // Default to boolean for quick path
		numericSubtype:    models.UnsignedIntFieldType,
		unit:              "times",
		prompt:            "Did you accomplish this goal today?",
		comment:           "",
		currentStep:       0,
		maxSteps:          0, // Will be determined based on flow
	}

	// Initialize the first step
	creator.initializeStep()

	return creator
}

// Init implements tea.Model - called when the model is first initialized
func (m *SimpleGoalCreator) Init() tea.Cmd {
	// AIDEV-NOTE: Following bubbletea pattern - Init() returns initial command
	// Form initialization happens in constructor per huh documentation
	return m.form.Init()
}

// Update implements tea.Model - handles messages and updates state
func (m *SimpleGoalCreator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}

	// AIDEV-NOTE: Following huh/bubbletea integration pattern
	// Delegate form updates to huh, check for completion
	var cmd tea.Cmd
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// Check if current step is completed
	if m.form.State == huh.StateCompleted {
		if m.currentStep < m.maxSteps-1 {
			// Move to next step
			m.currentStep++
			m.initializeStep()
			return m, m.form.Init()
		}
		// All steps completed - create goal
		goal, err := m.createGoalFromData()
		if err != nil {
			m.err = err
		} else {
			m.result = goal
		}
		m.quitting = true
		return m, tea.Quit
	}

	return m, cmd
}

// View implements tea.Model - renders the current state
func (m *SimpleGoalCreator) View() string {
	if m.quitting {
		if m.err != nil {
			return fmt.Sprintf("Error creating goal: %v\n", m.err)
		}
		if m.result != nil {
			return fmt.Sprintf("✅ Goal created successfully: %s\n", m.result.Title)
		}
		return "Goal creation cancelled.\n"
	}

	// AIDEV-NOTE: Simple view rendering - just show the form
	// Form handles all rendering, progress, validation per huh documentation
	return m.form.View()
}

// GetResult returns the created goal (after completion)
func (m *SimpleGoalCreator) GetResult() (*models.Goal, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

// IsCompleted returns true if the form was completed successfully
func (m *SimpleGoalCreator) IsCompleted() bool {
	return m.result != nil && m.err == nil
}

// IsCancelled returns true if the form was cancelled
func (m *SimpleGoalCreator) IsCancelled() bool {
	return m.quitting && m.result == nil && m.err == nil
}

// initializeStep initializes the form for the current step
func (m *SimpleGoalCreator) initializeStep() {
	switch m.currentStep {
	case 0:
		// Start with field type selection, but default to Boolean for quick path
		m.form = m.createFieldTypeSelectionForm()
		m.maxSteps = 4 // Default to full flow, will be adjusted dynamically
	case 1:
		// After field type selection, determine if we need field configuration
		if m.needsFieldConfiguration() {
			m.form = m.createFieldConfigurationForm()
		} else {
			m.form = m.createScoringTypeForm()
		}
		m.adjustFlowForFieldType()
	case 2:
		if m.needsFieldConfiguration() {
			m.form = m.createScoringTypeForm()
		} else {
			// Skip this step, go to final form
			m.form = m.createPromptAndCommentForm()
		}
	case 3:
		// Final step: prompt/comment based on scoring type
		// TODO: Add criteria form for automatic scoring in subtask 1.3
		m.form = m.createPromptAndCommentForm()
	default:
		m.err = fmt.Errorf("unknown step: %d", m.currentStep)
	}
}

// adjustFlowForFieldType adjusts the flow and max steps based on field type selection
func (m *SimpleGoalCreator) adjustFlowForFieldType() {
	// Determine actual number of steps needed based on field type
	steps := 1 // Field type selection (step 0)
	
	// Add field configuration step if needed
	if m.needsFieldConfiguration() {
		steps++ // Field configuration step
	}
	
	steps++ // Scoring type step
	steps++ // Prompt/comment step
	
	m.maxSteps = steps
}

// needsFieldConfiguration returns true if the selected field type needs configuration
func (m *SimpleGoalCreator) needsFieldConfiguration() bool {
	switch m.selectedFieldType {
	case "numeric":
		return true // Needs subtype, unit, constraints
	case models.TextFieldType:
		return true // Needs multiline option
	default:
		return false // Boolean, time, duration need no config
	}
}

// createFieldTypeSelectionForm creates the field type selection form
func (m *SimpleGoalCreator) createFieldTypeSelectionForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Field Type").
				Description("Choose what type of data this goal will collect").
				Options(
					huh.NewOption("Boolean (Simple completion)", models.BooleanFieldType),
					huh.NewOption("Text (Written notes)", models.TextFieldType),
					huh.NewOption("Numeric (Numbers with units)", "numeric"),
					huh.NewOption("Time (Time of day)", models.TimeFieldType),
					huh.NewOption("Duration (Time periods)", models.DurationFieldType),
				).
				Value(&m.selectedFieldType),
		),
	)
}

// createFieldConfigurationForm creates the field configuration form (conditional)
func (m *SimpleGoalCreator) createFieldConfigurationForm() *huh.Form {
	switch m.selectedFieldType {
	case "numeric":
		return m.createNumericConfigForm()
	case models.TextFieldType:
		return m.createTextConfigForm()
	default:
		// Should not happen due to needsFieldConfiguration check
		return huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Field Configuration").
					Description("This field type requires no additional configuration."),
			),
		)
	}
}

// createNumericConfigForm creates numeric field configuration form
func (m *SimpleGoalCreator) createNumericConfigForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Numeric Type").
				Description("Choose the type of numbers this goal will track").
				Options(
					huh.NewOption("Whole numbers (0, 1, 2, ...)", models.UnsignedIntFieldType),
					huh.NewOption("Positive decimals (0.5, 1.2, ...)", models.UnsignedDecimalFieldType),
					huh.NewOption("Any numbers (-1, 0, 1.5, ...)", models.DecimalFieldType),
				).
				Value(&m.numericSubtype),

			huh.NewInput().
				Title("Unit").
				Description("The unit for this measurement (e.g., 'minutes', 'reps', 'pages')").
				Value(&m.unit).
				Placeholder("times"),

			huh.NewConfirm().
				Title("Add Min/Max Constraints").
				Description("Do you want to set minimum and maximum value limits?").
				Value(&m.hasMinMax),
		),

		// Conditional group for min/max values
		huh.NewGroup(
			huh.NewInput().
				Title("Minimum Value").
				Description("Minimum allowed value (optional)").
				Value(&m.minValue).
				Placeholder("0"),

			huh.NewInput().
				Title("Maximum Value").
				Description("Maximum allowed value (optional)").
				Value(&m.maxValue).
				Placeholder("100"),
		).WithHideFunc(func() bool {
			return !m.hasMinMax
		}),
	)
}

// createTextConfigForm creates text field configuration form
func (m *SimpleGoalCreator) createTextConfigForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Multiline Text").
				Description("Allow multiple lines of text for longer responses?").
				Value(&m.multilineText),
		),
	)
}

// createScoringTypeForm creates the scoring type selection form
func (m *SimpleGoalCreator) createScoringTypeForm() *huh.Form {
	options := []huh.Option[models.ScoringType]{
		huh.NewOption("Manual (I'll mark completion myself)", models.ManualScoring),
	}

	// Only allow automatic scoring for field types that support criteria
	if m.supportsAutomaticScoring() {
		options = append(options, huh.NewOption("Automatic (Based on criteria I define)", models.AutomaticScoring))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[models.ScoringType]().
				Title("Scoring Type").
				Description("How should goal achievement be determined?").
				Options(options...).
				Value(&m.scoringType),
		),
	)
}

// supportsAutomaticScoring returns true if the selected field type supports automatic scoring
func (m *SimpleGoalCreator) supportsAutomaticScoring() bool {
	switch m.selectedFieldType {
	case models.TextFieldType:
		return false // Text fields restricted to manual scoring
	default:
		return true // Boolean, numeric, time, duration support automatic scoring
	}
}

// createPromptAndCommentForm creates the final prompt and comment form
func (m *SimpleGoalCreator) createPromptAndCommentForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Goal Prompt").
				Description("The question asked when tracking this goal").
				Value(&m.prompt).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("prompt cannot be empty")
					}
					return nil
				}),

			huh.NewInput().
				Title("Additional Comment (optional)").
				Description("Optional comment or context for this goal").
				Value(&m.comment).
				Placeholder("Any additional notes about this goal..."),
		),
	)
}

// createGoalFromData creates a models.Goal from the collected form data
func (m *SimpleGoalCreator) createGoalFromData() (*models.Goal, error) {
	// Build field type configuration
	fieldType := models.FieldType{
		Type: m.getResolvedFieldType(),
	}

	// Add field type specific configuration
	switch m.selectedFieldType {
	case "numeric":
		fieldType.Type = m.numericSubtype
		fieldType.Unit = strings.TrimSpace(m.unit)
		if m.hasMinMax {
			if minVal := strings.TrimSpace(m.minValue); minVal != "" {
				if val, err := strconv.ParseFloat(minVal, 64); err == nil {
					fieldType.Min = &val
				}
			}
			if maxVal := strings.TrimSpace(m.maxValue); maxVal != "" {
				if val, err := strconv.ParseFloat(maxVal, 64); err == nil {
					fieldType.Max = &val
				}
			}
		}
	case models.TextFieldType:
		fieldType.Multiline = &m.multilineText
	}

	goal := &models.Goal{
		Title:       strings.TrimSpace(m.title),
		Description: strings.TrimSpace(m.description),
		GoalType:    m.goalType,
		FieldType:   fieldType,
		ScoringType: m.scoringType,
		Prompt:      strings.TrimSpace(m.prompt),
	}

	// Add comment if provided
	if comment := strings.TrimSpace(m.comment); comment != "" {
		// Note: Comment field doesn't exist in models.Goal yet - this is a design decision point
		// For now, we could append it to description or add it as HelpText
		if goal.Description != "" {
			goal.Description = goal.Description + "\n\nComment: " + comment
		} else {
			goal.Description = "Comment: " + comment
		}
	}

	// TODO: Add automatic criteria configuration for automatic scoring
	// For now, focus on manual scoring path

	return goal, nil
}

// getResolvedFieldType resolves the field type (handles "numeric" -> specific numeric type)
func (m *SimpleGoalCreator) getResolvedFieldType() string {
	if m.selectedFieldType == "numeric" {
		return m.numericSubtype
	}
	return m.selectedFieldType
}
