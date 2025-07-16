package habitconfig

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"davidlee/vice/internal/models"
)

// AIDEV-NOTE: Informational habit creator using idiomatic bubbletea patterns
// Based on SimpleHabitCreator implementation following https://github.com/charmbracelet/huh/blob/main/examples/bubbletea/main.go
// Flow: Field Type Selection → Field Configuration → Direction Preference → Save
// Focuses on data collection without scoring or criteria

// InformationalHabitCreator implements a bubbletea model for creating informational habits
type InformationalHabitCreator struct {
	form     *huh.Form
	quitting bool
	err      error
	result   *models.Habit

	// Pre-populated basic info
	title       string
	description string
	habitType   models.HabitType

	// Field configuration data - bound directly to form fields per huh documentation
	selectedFieldType string
	numericSubtype    string
	unit              string
	multilineText     bool
	minValue          string
	maxValue          string
	direction         string
	hasMinMax         bool
	prompt            string

	// State tracking
	currentStep int
	maxSteps    int
}

// NewInformationalHabitCreatorForEdit creates an informational habit creator pre-populated with existing habit data for editing
func NewInformationalHabitCreatorForEdit(habit *models.Habit) *InformationalHabitCreator {
	creator := &InformationalHabitCreator{
		title:             habit.Title,
		description:       habit.Description,
		habitType:         habit.HabitType,
		selectedFieldType: habit.FieldType.Type,
		numericSubtype:    habit.FieldType.Type,
		unit:              habit.FieldType.Unit,
		direction:         habit.Direction,
		prompt:            habit.Prompt,
		currentStep:       0,
		maxSteps:          4,
	}

	// Handle field type specific configuration
	switch habit.FieldType.Type {
	case models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType:
		creator.selectedFieldType = "numeric"
		creator.numericSubtype = habit.FieldType.Type
		if habit.FieldType.Min != nil {
			creator.minValue = fmt.Sprintf("%.2f", *habit.FieldType.Min)
			creator.hasMinMax = true
		}
		if habit.FieldType.Max != nil {
			creator.maxValue = fmt.Sprintf("%.2f", *habit.FieldType.Max)
			creator.hasMinMax = true
		}
	case models.TextFieldType:
		if habit.FieldType.Multiline != nil {
			creator.multilineText = *habit.FieldType.Multiline
		}
	}

	// Initialize the first form
	creator.initializeStep()

	return creator
}

// NewInformationalHabitCreator creates a new informational habit creator with pre-populated basic info
func NewInformationalHabitCreator(title, description string, habitType models.HabitType) *InformationalHabitCreator {
	creator := &InformationalHabitCreator{
		title:          title,
		description:    description,
		habitType:      habitType,
		numericSubtype: models.UnsignedIntFieldType, // Default numeric subtype
		unit:           "times",                     // Default unit
		direction:      "neutral",                   // Default direction
		prompt:         "",                          // Will be set based on field type
		currentStep:    0,
		maxSteps:       4, // Field Type → Field Config → Direction → Prompt
	}

	// Initialize the first form
	creator.initializeStep()

	return creator
}

// Init implements tea.Model - called when the model is first initialized
func (m *InformationalHabitCreator) Init() tea.Cmd {
	// AIDEV-NOTE: Following bubbletea pattern - Init() returns initial command
	// Form initialization happens in constructor per huh documentation
	return m.form.Init()
}

// Update implements tea.Model - handles messages and updates state
func (m *InformationalHabitCreator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// All steps completed - create habit
		habit, err := m.createHabitFromData()
		if err != nil {
			m.err = err
		} else {
			m.result = habit
		}
		m.quitting = true
		return m, tea.Quit
	}

	return m, cmd
}

// View implements tea.Model - renders the current state
func (m *InformationalHabitCreator) View() string {
	if m.quitting {
		if m.err != nil {
			return fmt.Sprintf("Error creating informational habit: %v\n", m.err)
		}
		if m.result != nil {
			return fmt.Sprintf("✅ Informational habit created successfully: %s\n", m.result.Title)
		}
		return "Informational habit creation cancelled.\n"
	}

	// AIDEV-NOTE: Simple view rendering with step indicator
	// Form handles all rendering, progress, validation per huh documentation
	stepHeader := fmt.Sprintf("Step %d of %d: %s\n\n", m.currentStep+1, m.maxSteps, m.getStepTitle())
	return stepHeader + m.form.View()
}

// GetResult returns the created habit (after completion)
func (m *InformationalHabitCreator) GetResult() (*models.Habit, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

// IsCompleted returns true if the form was completed successfully
func (m *InformationalHabitCreator) IsCompleted() bool {
	return m.result != nil && m.err == nil
}

// IsCancelled returns true if the form was cancelled
func (m *InformationalHabitCreator) IsCancelled() bool {
	return m.quitting && m.result == nil && m.err == nil
}

// Private methods

func (m *InformationalHabitCreator) initializeStep() {
	switch m.currentStep {
	case 0:
		m.form = m.createFieldTypeSelectionForm()
	case 1:
		m.form = m.createFieldConfigurationForm()
	case 2:
		m.form = m.createDirectionSelectionForm()
	case 3:
		m.form = m.createPromptForm()
	default:
		m.err = fmt.Errorf("unknown step: %d", m.currentStep)
	}
}

func (m *InformationalHabitCreator) getStepTitle() string {
	switch m.currentStep {
	case 0:
		return "Field Type Selection"
	case 1:
		return "Field Configuration"
	case 2:
		return "Direction Preference"
	case 3:
		return "Habit Prompt"
	default:
		return "Unknown Step"
	}
}

func (m *InformationalHabitCreator) createFieldTypeSelectionForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Field Type").
				Description("Choose what type of data this habit will track").
				Options(
					huh.NewOption("Boolean (True/False)", models.BooleanFieldType),
					huh.NewOption("Text (Written responses)", models.TextFieldType),
					huh.NewOption("Numeric (Numbers with units)", "numeric"),
					huh.NewOption("Time (Time of day)", models.TimeFieldType),
					huh.NewOption("Duration (Time periods)", models.DurationFieldType),
				).
				Value(&m.selectedFieldType),
		),
	)
}

func (m *InformationalHabitCreator) createFieldConfigurationForm() *huh.Form {
	// Create conditional groups based on field type
	var groups []*huh.Group

	switch m.selectedFieldType {
	case "numeric":
		groups = append(groups, m.createNumericConfigGroups()...)
	case models.TextFieldType:
		groups = append(groups, m.createTextConfigGroup())
	case models.BooleanFieldType, models.TimeFieldType, models.DurationFieldType:
		// These field types need no additional configuration
		// Create a simple confirmation group
		groups = append(groups, huh.NewGroup(
			huh.NewNote().
				Title("Field Configuration").
				Description(fmt.Sprintf("✅ %s fields require no additional configuration.",
					m.getFieldTypeDisplayName(m.selectedFieldType))),
		))
	}

	if len(groups) == 0 {
		// Fallback group if no configuration is needed
		groups = append(groups, huh.NewGroup(
			huh.NewNote().
				Title("Configuration Complete").
				Description("This field type is ready to use."),
		))
	}

	return huh.NewForm(groups...)
}

func (m *InformationalHabitCreator) createDirectionSelectionForm() *huh.Form {
	// Only show direction selection for field types that support it
	if !m.supportsDirection() {
		// Skip direction selection with a note
		return huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Direction Preference").
					Description(fmt.Sprintf("✅ %s fields use neutral direction (no preference).",
						m.getFieldTypeDisplayName(m.selectedFieldType))),
			),
		)
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Value Direction").
				Description("Indicates whether higher or lower values are generally better").
				Options(
					huh.NewOption("Higher is better (↑)", "higher_better"),
					huh.NewOption("Lower is better (↓)", "lower_better"),
					huh.NewOption("Neutral (no preference)", "neutral"),
				).
				Value(&m.direction),
		),
	)
}

func (m *InformationalHabitCreator) createPromptForm() *huh.Form {
	// Generate a default prompt based on field type
	defaultPrompt := m.generateDefaultPrompt()
	if m.prompt == "" {
		m.prompt = defaultPrompt
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Habit Prompt").
				Description("The question asked when tracking this habit").
				Placeholder(defaultPrompt).
				Value(&m.prompt).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("prompt cannot be empty")
					}
					return nil
				}),
		),
	)
}

func (m *InformationalHabitCreator) createNumericConfigGroups() []*huh.Group {
	groups := []*huh.Group{
		// Numeric subtype selection
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Numeric Type").
				Description("Choose the type of numbers you'll be tracking").
				Options(
					huh.NewOption("Whole numbers (0, 1, 2, 3...)", models.UnsignedIntFieldType),
					huh.NewOption("Positive decimals (0.5, 1.25, 2.7...)", models.UnsignedDecimalFieldType),
					huh.NewOption("Any numbers (including negative)", models.DecimalFieldType),
				).
				Value(&m.numericSubtype),
		),

		// Unit configuration
		huh.NewGroup(
			huh.NewInput().
				Title("Unit").
				Description("What unit will you measure in? (e.g., 'reps', 'kg', 'minutes', 'pages')").
				Placeholder("times").
				Value(&m.unit).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						m.unit = "times" // Default if empty
					} else {
						m.unit = strings.TrimSpace(s)
					}
					return nil
				}),
		),

		// Optional constraints
		huh.NewGroup(
			huh.NewConfirm().
				Title("Add Value Constraints?").
				Description("Do you want to set minimum or maximum value limits?").
				Value(&m.hasMinMax),
		),
	}

	// Add min/max configuration group if requested
	groups = append(groups, huh.NewGroup(
		huh.NewInput().
			Title("Minimum Value (optional)").
			Description("Leave empty for no minimum limit").
			Value(&m.minValue).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return nil // Empty is OK
				}
				if _, err := fmt.Sscanf(strings.TrimSpace(s), "%f", new(float64)); err != nil {
					return fmt.Errorf("must be a valid number")
				}
				return nil
			}),

		huh.NewInput().
			Title("Maximum Value (optional)").
			Description("Leave empty for no maximum limit").
			Value(&m.maxValue).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return nil // Empty is OK
				}
				if _, err := fmt.Sscanf(strings.TrimSpace(s), "%f", new(float64)); err != nil {
					return fmt.Errorf("must be a valid number")
				}
				return nil
			}),
	).WithHideFunc(func() bool {
		// Hide min/max inputs if user doesn't want constraints
		return !m.hasMinMax
	}))

	return groups
}

func (m *InformationalHabitCreator) createTextConfigGroup() *huh.Group {
	return huh.NewGroup(
		huh.NewConfirm().
			Title("Multiline Text").
			Description("Will you need multiple lines for text responses?").
			Value(&m.multilineText),
	)
}

func (m *InformationalHabitCreator) supportsDirection() bool {
	// Only numeric, time, and duration fields support direction preference
	switch m.selectedFieldType {
	case "numeric", models.TimeFieldType, models.DurationFieldType:
		return true
	case models.BooleanFieldType, models.TextFieldType:
		return false
	default:
		return false
	}
}

func (m *InformationalHabitCreator) getFieldTypeDisplayName(fieldType string) string {
	switch fieldType {
	case models.BooleanFieldType:
		return "Boolean"
	case models.TextFieldType:
		return "Text"
	case "numeric":
		return "Numeric"
	case models.TimeFieldType:
		return "Time"
	case models.DurationFieldType:
		return "Duration"
	default:
		return "Unknown"
	}
}

func (m *InformationalHabitCreator) generateDefaultPrompt() string {
	switch m.selectedFieldType {
	case models.BooleanFieldType:
		return fmt.Sprintf("Did you %s today?", strings.ToLower(m.title))
	case models.TextFieldType:
		if m.multilineText {
			return fmt.Sprintf("What details do you want to record about %s?", strings.ToLower(m.title))
		}
		return fmt.Sprintf("How would you describe your %s today?", strings.ToLower(m.title))
	case "numeric":
		if m.unit != "" && m.unit != "times" {
			return fmt.Sprintf("How many %s did you record for %s?", m.unit, strings.ToLower(m.title))
		}
		return fmt.Sprintf("What number do you want to record for %s?", strings.ToLower(m.title))
	case models.TimeFieldType:
		return fmt.Sprintf("What time did you %s?", strings.ToLower(m.title))
	case models.DurationFieldType:
		return fmt.Sprintf("How long did you spend on %s?", strings.ToLower(m.title))
	default:
		return fmt.Sprintf("What value do you want to record for %s?", strings.ToLower(m.title))
	}
}

// createHabitFromData creates a models.Habit from the collected form data
func (m *InformationalHabitCreator) createHabitFromData() (*models.Habit, error) {
	// AIDEV-NOTE: Create informational habit structure matching expected YAML format
	// Informational habits have no scoring and no criteria - pure data collection
	// Expected structure:
	//   - title: Title
	//     id: title
	//     habit_type: informational
	//     field_type:
	//       type: [boolean|text|unsigned_int|unsigned_decimal|decimal|time|duration]
	//       unit: [for numeric fields]
	//       multiline: [for text fields]
	//       min: [optional, for numeric fields]
	//       max: [optional, for numeric fields]
	//     scoring_type: manual  # Always manual for informational
	//     direction: [higher_better|lower_better|neutral]
	//     prompt: "Question asked during entry recording"

	habit := &models.Habit{
		Title:       strings.TrimSpace(m.title),
		Description: strings.TrimSpace(m.description),
		HabitType:   m.habitType,
		FieldType:   m.createFieldType(),
		ScoringType: models.ManualScoring, // Informational habits always use manual scoring
		Direction:   m.direction,
		Prompt:      strings.TrimSpace(m.prompt),
	}

	return habit, nil
}

func (m *InformationalHabitCreator) createFieldType() models.FieldType {
	fieldType := models.FieldType{}

	// Set the actual field type based on selection
	switch m.selectedFieldType {
	case "numeric":
		fieldType.Type = m.numericSubtype
		fieldType.Unit = m.unit
		if m.hasMinMax {
			if m.minValue != "" {
				var minVal float64
				if n, err := fmt.Sscanf(m.minValue, "%f", &minVal); err == nil && n == 1 {
					fieldType.Min = &minVal
				}
			}
			if m.maxValue != "" {
				var maxVal float64
				if n, err := fmt.Sscanf(m.maxValue, "%f", &maxVal); err == nil && n == 1 {
					fieldType.Max = &maxVal
				}
			}
		}
	case models.TextFieldType:
		fieldType.Type = models.TextFieldType
		fieldType.Multiline = &m.multilineText
	case models.BooleanFieldType, models.TimeFieldType, models.DurationFieldType:
		fieldType.Type = m.selectedFieldType
	}

	return fieldType
}
