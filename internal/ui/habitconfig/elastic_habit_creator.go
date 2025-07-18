package habitconfig

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/davidlee/vice/internal/models"
)

// AIDEV-NOTE: Elastic habit creation following SimpleHabitCreator patterns with three-tier criteria
// Handles mini/midi/maxi achievement levels with automatic scoring and ordering validation
// Multi-step flow: Field Type → Field Config → Scoring → Criteria (3 levels) → Prompt

// ElasticHabitCreator implements a bubbletea model for creating elastic habits
type ElasticHabitCreator struct {
	form     *huh.Form
	quitting bool
	err      error
	result   *models.Habit

	// Pre-populated basic info
	title       string
	description string
	habitType   models.HabitType

	// Field configuration data - reuses SimpleHabitCreator patterns
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

	// Elastic-specific: Three-tier criteria configuration
	miniCriteriaType      string // "greater_than", "less_than", "equals", "before", "after", "range"
	miniCriteriaValue     string // Value for mini achievement
	miniCriteriaValue2    string // Second value for mini range
	miniCriteriaTimeValue string // Time value for mini time-based criteria
	miniRangeInclusive    bool   // Whether mini range bounds are inclusive

	midiCriteriaType      string // Midi achievement criteria
	midiCriteriaValue     string
	midiCriteriaValue2    string
	midiCriteriaTimeValue string
	midiRangeInclusive    bool

	maxiCriteriaType      string // Maxi achievement criteria
	maxiCriteriaValue     string
	maxiCriteriaValue2    string
	maxiCriteriaTimeValue string
	maxiRangeInclusive    bool

	// State tracking for multi-step flow
	currentStep int
	maxSteps    int
}

// NewElasticHabitCreator creates a new elastic habit creator with pre-populated basic info
func NewElasticHabitCreator(title, description string, habitType models.HabitType) *ElasticHabitCreator {
	creator := &ElasticHabitCreator{
		title:             title,
		description:       description,
		habitType:         habitType,
		selectedFieldType: models.BooleanFieldType, // Default, but elastic rarely uses boolean
		numericSubtype:    models.UnsignedIntFieldType,
		unit:              "times",
		prompt:            "How much did you achieve today?",
		comment:           "",
		currentStep:       0,
		maxSteps:          0, // Will be determined based on flow
	}

	// Initialize the first step
	creator.initializeStep()

	return creator
}

// TestElasticHabitData contains pre-configured data for headless testing
type TestElasticHabitData struct {
	FieldType      string
	NumericSubtype string
	Unit           string
	MultilineText  bool
	MinValue       string
	MaxValue       string
	HasMinMax      bool
	ScoringType    models.ScoringType
	Prompt         string
	Comment        string

	// Mini criteria
	MiniCriteriaType      string
	MiniCriteriaValue     string
	MiniCriteriaValue2    string
	MiniCriteriaTimeValue string
	MiniRangeInclusive    bool

	// Midi criteria
	MidiCriteriaType      string
	MidiCriteriaValue     string
	MidiCriteriaValue2    string
	MidiCriteriaTimeValue string
	MidiRangeInclusive    bool

	// Maxi criteria
	MaxiCriteriaType      string
	MaxiCriteriaValue     string
	MaxiCriteriaValue2    string
	MaxiCriteriaTimeValue string
	MaxiRangeInclusive    bool
}

// NewElasticHabitCreatorForEdit creates an elastic habit creator pre-populated with existing habit data for editing
func NewElasticHabitCreatorForEdit(habit *models.Habit) *ElasticHabitCreator {
	data := habitToTestElasticData(habit)
	return NewElasticHabitCreatorForTesting(habit.Title, habit.Description, habit.HabitType, data)
}

// habitToTestElasticData converts a models.Habit to TestElasticHabitData for pre-population
func habitToTestElasticData(habit *models.Habit) TestElasticHabitData {
	data := TestElasticHabitData{
		FieldType:   habit.FieldType.Type,
		ScoringType: habit.ScoringType,
		Prompt:      habit.Prompt,
		Comment:     extractCommentFromDescription(habit.Description),
	}

	// Field type specific conversion (reuse logic from simple creator)
	switch habit.FieldType.Type {
	case models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType:
		data.FieldType = "numeric"
		data.NumericSubtype = habit.FieldType.Type
		data.Unit = habit.FieldType.Unit
		if habit.FieldType.Min != nil {
			data.MinValue = fmt.Sprintf("%.2f", *habit.FieldType.Min)
			data.HasMinMax = true
		}
		if habit.FieldType.Max != nil {
			data.MaxValue = fmt.Sprintf("%.2f", *habit.FieldType.Max)
			data.HasMinMax = true
		}
	case models.TextFieldType:
		if habit.FieldType.Multiline != nil {
			data.MultilineText = *habit.FieldType.Multiline
		}
	}

	// Convert elastic-specific criteria
	if habit.MiniCriteria != nil {
		data.MiniCriteriaType, data.MiniCriteriaValue, data.MiniCriteriaValue2, data.MiniCriteriaTimeValue, data.MiniRangeInclusive = convertCriteriaToElasticData(habit.MiniCriteria)
	}
	if habit.MidiCriteria != nil {
		data.MidiCriteriaType, data.MidiCriteriaValue, data.MidiCriteriaValue2, data.MidiCriteriaTimeValue, data.MidiRangeInclusive = convertCriteriaToElasticData(habit.MidiCriteria)
	}
	if habit.MaxiCriteria != nil {
		data.MaxiCriteriaType, data.MaxiCriteriaValue, data.MaxiCriteriaValue2, data.MaxiCriteriaTimeValue, data.MaxiRangeInclusive = convertCriteriaToElasticData(habit.MaxiCriteria)
	}

	return data
}

// convertCriteriaToElasticData converts models.Criteria to elastic test data format
func convertCriteriaToElasticData(criteria *models.Criteria) (criteriaType, value, value2, timeValue string, inclusive bool) {
	if criteria.Condition == nil {
		return "", "", "", "", false
	}

	cond := criteria.Condition
	if cond.GreaterThan != nil {
		return "greater_than", fmt.Sprintf("%.2f", *cond.GreaterThan), "", "", false
	}
	if cond.GreaterThanOrEqual != nil {
		return "greater_than_or_equal", fmt.Sprintf("%.2f", *cond.GreaterThanOrEqual), "", "", false
	}
	if cond.LessThan != nil {
		return "less_than", fmt.Sprintf("%.2f", *cond.LessThan), "", "", false
	}
	if cond.LessThanOrEqual != nil {
		return "less_than_or_equal", fmt.Sprintf("%.2f", *cond.LessThanOrEqual), "", "", false
	}
	if cond.Equals != nil {
		return "equals", fmt.Sprintf("%t", *cond.Equals), "", "", false
	}
	if cond.Before != "" {
		return "before", "", "", cond.Before, false
	}
	if cond.After != "" {
		return "after", "", "", cond.After, false
	}
	return "", "", "", "", false
}

// NewElasticHabitCreatorForTesting creates an elastic habit creator with pre-populated test data, bypassing UI
func NewElasticHabitCreatorForTesting(title, description string, habitType models.HabitType, data TestElasticHabitData) *ElasticHabitCreator {
	creator := &ElasticHabitCreator{
		title:             title,
		description:       description,
		habitType:         habitType,
		selectedFieldType: data.FieldType,
		numericSubtype:    data.NumericSubtype,
		unit:              data.Unit,
		multilineText:     data.MultilineText,
		minValue:          data.MinValue,
		maxValue:          data.MaxValue,
		hasMinMax:         data.HasMinMax,
		scoringType:       data.ScoringType,
		prompt:            data.Prompt,
		comment:           data.Comment,

		// Mini criteria
		miniCriteriaType:      data.MiniCriteriaType,
		miniCriteriaValue:     data.MiniCriteriaValue,
		miniCriteriaValue2:    data.MiniCriteriaValue2,
		miniCriteriaTimeValue: data.MiniCriteriaTimeValue,
		miniRangeInclusive:    data.MiniRangeInclusive,

		// Midi criteria
		midiCriteriaType:      data.MidiCriteriaType,
		midiCriteriaValue:     data.MidiCriteriaValue,
		midiCriteriaValue2:    data.MidiCriteriaValue2,
		midiCriteriaTimeValue: data.MidiCriteriaTimeValue,
		midiRangeInclusive:    data.MidiRangeInclusive,

		// Maxi criteria
		maxiCriteriaType:      data.MaxiCriteriaType,
		maxiCriteriaValue:     data.MaxiCriteriaValue,
		maxiCriteriaValue2:    data.MaxiCriteriaValue2,
		maxiCriteriaTimeValue: data.MaxiCriteriaTimeValue,
		maxiRangeInclusive:    data.MaxiRangeInclusive,

		// Skip UI initialization for testing
		form:     nil,
		quitting: false,
		err:      nil,
		result:   nil,
	}

	return creator
}

// CreateHabitDirectly bypasses UI flow and creates habit directly from configured data
func (m *ElasticHabitCreator) CreateHabitDirectly() (*models.Habit, error) {
	return m.createHabitFromData()
}

// Init implements tea.Model - called when the model is first initialized
func (m *ElasticHabitCreator) Init() tea.Cmd {
	// AIDEV-NOTE: Following bubbletea pattern - Init() returns initial command
	// Form initialization happens in constructor per huh documentation
	return m.form.Init()
}

// Update implements tea.Model - handles messages and updates state
func (m *ElasticHabitCreator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// Adjust flow if we just completed scoring type selection
		if m.isCurrentStepScoringType() {
			m.adjustFlowForScoringType()
		}

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
func (m *ElasticHabitCreator) View() string {
	if m.quitting {
		if m.err != nil {
			return fmt.Sprintf("Error creating elastic habit: %v\n", m.err)
		}
		if m.result != nil {
			return fmt.Sprintf("✅ Elastic habit created successfully: %s\n", m.result.Title)
		}
		return "Elastic habit creation cancelled.\n"
	}

	// AIDEV-NOTE: Simple view rendering - just show the form
	// Form handles all rendering, progress, validation per huh documentation
	return m.form.View()
}

// GetResult returns the created habit (after completion)
func (m *ElasticHabitCreator) GetResult() (*models.Habit, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

// IsCompleted returns true if the form was completed successfully
func (m *ElasticHabitCreator) IsCompleted() bool {
	return m.result != nil && m.err == nil
}

// IsCancelled returns true if the form was cancelled
func (m *ElasticHabitCreator) IsCancelled() bool {
	return m.quitting && m.result == nil && m.err == nil
}

// AIDEV-NOTE: multi-step-flow; elastic-specific flow with criteria for mini/midi/maxi
// initializeStep initializes the form for the current step
func (m *ElasticHabitCreator) initializeStep() {
	switch m.currentStep {
	case 0:
		// Start with field type selection (exclude boolean - not meaningful for elastic)
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
			// Skip this step, go to criteria or final form
			if m.scoringType == models.AutomaticScoring {
				m.form = m.createCriteriaDefinitionForm()
			} else {
				m.form = m.createPromptAndCommentForm()
			}
		}
	case 3:
		// Step depends on scoring type: criteria for automatic, prompt for manual
		if m.scoringType == models.AutomaticScoring {
			m.form = m.createCriteriaDefinitionForm()
		} else {
			m.form = m.createPromptAndCommentForm()
		}
	case 4:
		// Final step: prompt/comment (only reached with automatic scoring)
		m.form = m.createPromptAndCommentForm()
	default:
		m.err = fmt.Errorf("unknown step: %d", m.currentStep)
	}
}

// adjustFlowForFieldType adjusts the flow and max steps based on field type selection
func (m *ElasticHabitCreator) adjustFlowForFieldType() {
	// Determine actual number of steps needed based on field type
	steps := 1 // Field type selection (step 0)

	// Add field configuration step if needed
	if m.needsFieldConfiguration() {
		steps++ // Field configuration step
	}

	steps++ // Scoring type step

	// Add criteria step for automatic scoring (will be determined later)
	// For now, assume we might need it - will be adjusted in scoring step
	steps++ // Criteria step (conditional)
	steps++ // Prompt/comment step

	m.maxSteps = steps
}

// adjustFlowForScoringType adjusts the flow based on scoring type selection
func (m *ElasticHabitCreator) adjustFlowForScoringType() {
	if m.scoringType == models.ManualScoring {
		// Manual scoring doesn't need criteria step, reduce max steps by 1
		m.maxSteps--
	}
}

// isCurrentStepScoringType returns true if the current step is scoring type selection
func (m *ElasticHabitCreator) isCurrentStepScoringType() bool {
	// Scoring type is step 1 if no field config needed, step 2 if field config needed
	if m.needsFieldConfiguration() {
		return m.currentStep == 2
	}
	return m.currentStep == 1
}

// needsFieldConfiguration returns true if the selected field type needs configuration
func (m *ElasticHabitCreator) needsFieldConfiguration() bool {
	switch m.selectedFieldType {
	case "numeric":
		return true // Needs subtype, unit, constraints
	case models.TextFieldType:
		return true // Needs multiline option
	default:
		return false // Time, duration need no config (boolean excluded from elastic)
	}
}

// AIDEV-NOTE: field-type-selection; excludes boolean (not meaningful for mini/midi/maxi), focuses on measurable data
// createFieldTypeSelectionForm creates the field type selection form (excluding boolean)
func (m *ElasticHabitCreator) createFieldTypeSelectionForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Field Type").
				Description("Choose what type of data this elastic habit will collect\n(Elastic habits track achievement levels: mini/midi/maxi)").
				Options(
					// Note: Boolean excluded - not meaningful for three-tier achievement
					huh.NewOption("Text (Written notes/descriptions)", models.TextFieldType),
					huh.NewOption("Numeric (Numbers with units)", "numeric"),
					huh.NewOption("Time (Time of day)", models.TimeFieldType),
					huh.NewOption("Duration (Time periods)", models.DurationFieldType),
				).
				Value(&m.selectedFieldType),
		),
	)
}

// createFieldConfigurationForm creates the field configuration form (reuses SimpleHabitCreator patterns)
func (m *ElasticHabitCreator) createFieldConfigurationForm() *huh.Form {
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

// createNumericConfigForm creates numeric field configuration form (reused from SimpleHabitCreator)
func (m *ElasticHabitCreator) createNumericConfigForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Numeric Type").
				Description("Choose the type of numbers this habit will track").
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

// createTextConfigForm creates text field configuration form (reused from SimpleHabitCreator)
func (m *ElasticHabitCreator) createTextConfigForm() *huh.Form {
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
func (m *ElasticHabitCreator) createScoringTypeForm() *huh.Form {
	options := []huh.Option[models.ScoringType]{
		huh.NewOption("Manual (I'll rate mini/midi/maxi myself)", models.ManualScoring),
	}

	// Only allow automatic scoring for field types that support criteria
	if m.supportsAutomaticScoring() {
		options = append(options, huh.NewOption("Automatic (Based on three-tier criteria I define)", models.AutomaticScoring))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[models.ScoringType]().
				Title("Scoring Type").
				Description("How should elastic habit achievement levels be determined?").
				Options(options...).
				Value(&m.scoringType),
		),
	)
}

// supportsAutomaticScoring returns true if the selected field type supports automatic scoring
func (m *ElasticHabitCreator) supportsAutomaticScoring() bool {
	switch m.selectedFieldType {
	case models.TextFieldType:
		return false // Text fields restricted to manual scoring
	default:
		return true // Numeric, time, duration support automatic scoring
	}
}

// AIDEV-NOTE: elastic-criteria-dispatch; creates three-tier criteria forms for mini/midi/maxi
// createCriteriaDefinitionForm creates the three-tier criteria definition form for automatic scoring
func (m *ElasticHabitCreator) createCriteriaDefinitionForm() *huh.Form {
	// For elastic habits, we need to collect criteria for all three tiers
	// This is more complex than simple habits and requires careful UX design
	switch m.selectedFieldType {
	case "numeric", models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType:
		return m.createNumericElasticCriteriaForm()
	case models.TimeFieldType:
		return m.createTimeElasticCriteriaForm()
	case models.DurationFieldType:
		return m.createDurationElasticCriteriaForm()
	default:
		// This shouldn't happen due to supportsAutomaticScoring check
		return huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Criteria Definition").
					Description("This field type does not support automatic scoring."),
			),
		)
	}
}

// AIDEV-NOTE: numeric-elastic-criteria; complex three-tier form with real-time validation for mini ≤ midi ≤ maxi
// createNumericElasticCriteriaForm creates three-tier criteria form for numeric fields
func (m *ElasticHabitCreator) createNumericElasticCriteriaForm() *huh.Form {
	unit := m.unit
	if unit == "" {
		unit = "units"
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Three-Tier Achievement Criteria").
				Description(fmt.Sprintf("Define thresholds for mini, midi, and maxi achievement levels.\nValues should increase: mini ≤ midi ≤ maxi\n\nUnit: %s", unit)),
		),

		// Mini criteria (lowest achievement)
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Mini Achievement (%s)", unit)).
				Description("Minimum threshold for basic achievement").
				Value(&m.miniCriteriaValue).
				Placeholder("10").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("mini criteria value is required")
					}
					if _, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err != nil {
						return fmt.Errorf("must be a valid number")
					}
					return nil
				}),
		),

		// Midi criteria (medium achievement)
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Midi Achievement (%s)", unit)).
				Description("Threshold for good achievement (must be ≥ mini)").
				Value(&m.midiCriteriaValue).
				Placeholder("20").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("midi criteria value is required")
					}
					midiVal, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
					if err != nil {
						return fmt.Errorf("must be a valid number")
					}
					// AIDEV-NOTE: real-time-validation; prevents invalid orderings during form entry (not just at habit creation)
					// Validate midi ≥ mini if mini is set
					if miniStr := strings.TrimSpace(m.miniCriteriaValue); miniStr != "" {
						if miniVal, err := strconv.ParseFloat(miniStr, 64); err == nil {
							if midiVal < miniVal {
								return fmt.Errorf("midi value (%.1f) must be ≥ mini value (%.1f)", midiVal, miniVal)
							}
						}
					}
					return nil
				}),
		),

		// Maxi criteria (highest achievement)
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Maxi Achievement (%s)", unit)).
				Description("Threshold for excellent achievement (must be ≥ midi)").
				Value(&m.maxiCriteriaValue).
				Placeholder("30").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("maxi criteria value is required")
					}
					maxiVal, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
					if err != nil {
						return fmt.Errorf("must be a valid number")
					}
					// Validate maxi ≥ midi if midi is set
					if midiStr := strings.TrimSpace(m.midiCriteriaValue); midiStr != "" {
						if midiVal, err := strconv.ParseFloat(midiStr, 64); err == nil {
							if maxiVal < midiVal {
								return fmt.Errorf("maxi value (%.1f) must be ≥ midi value (%.1f)", maxiVal, midiVal)
							}
						}
					}
					return nil
				}),
		),
	)
}

// createTimeElasticCriteriaForm creates three-tier criteria form for time fields
func (m *ElasticHabitCreator) createTimeElasticCriteriaForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Three-Tier Time Achievement Criteria").
				Description("Define time thresholds for mini, midi, and maxi achievement levels.\nUse HH:MM format (24-hour). Example: wake-up times 07:00 (mini), 06:30 (midi), 06:00 (maxi)"),
		),

		// Mini criteria
		huh.NewGroup(
			huh.NewInput().
				Title("Mini Achievement Time").
				Description("Time threshold for basic achievement").
				Value(&m.miniCriteriaTimeValue).
				Placeholder("07:00").
				Validate(m.validateTimeInput),
		),

		// Midi criteria
		huh.NewGroup(
			huh.NewInput().
				Title("Midi Achievement Time").
				Description("Time threshold for good achievement").
				Value(&m.midiCriteriaTimeValue).
				Placeholder("06:30").
				Validate(m.validateTimeInput),
		),

		// Maxi criteria
		huh.NewGroup(
			huh.NewInput().
				Title("Maxi Achievement Time").
				Description("Time threshold for excellent achievement").
				Value(&m.maxiCriteriaTimeValue).
				Placeholder("06:00").
				Validate(m.validateTimeInput),
		),
	)
}

// createDurationElasticCriteriaForm creates three-tier criteria form for duration fields
func (m *ElasticHabitCreator) createDurationElasticCriteriaForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Three-Tier Duration Achievement Criteria").
				Description("Define duration thresholds for mini, midi, and maxi achievement levels.\nUse formats like '15m', '1h', '1h 30m'. Values should increase: mini ≤ midi ≤ maxi"),
		),

		// Mini criteria
		huh.NewGroup(
			huh.NewInput().
				Title("Mini Achievement Duration").
				Description("Duration threshold for basic achievement").
				Value(&m.miniCriteriaValue).
				Placeholder("15m").
				Validate(m.validateDurationInput),
		),

		// Midi criteria
		huh.NewGroup(
			huh.NewInput().
				Title("Midi Achievement Duration").
				Description("Duration threshold for good achievement").
				Value(&m.midiCriteriaValue).
				Placeholder("30m").
				Validate(m.validateDurationInput),
		),

		// Maxi criteria
		huh.NewGroup(
			huh.NewInput().
				Title("Maxi Achievement Duration").
				Description("Duration threshold for excellent achievement").
				Value(&m.maxiCriteriaValue).
				Placeholder("60m").
				Validate(m.validateDurationInput),
		),
	)
}

// validateTimeInput validates time input format
func (m *ElasticHabitCreator) validateTimeInput(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("time value is required")
	}
	// Basic time format validation (same as SimpleHabitCreator)
	if !strings.Contains(s, ":") || len(strings.Split(s, ":")) != 2 {
		return fmt.Errorf("time must be in HH:MM format (e.g., 07:30)")
	}
	return nil
}

// validateDurationInput validates duration input format
func (m *ElasticHabitCreator) validateDurationInput(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("duration value is required")
	}
	// Basic validation - more detailed parsing would happen in the actual system
	if !strings.ContainsAny(strings.TrimSpace(s), "mh") {
		return fmt.Errorf("duration must include time units (e.g., 30m, 1h)")
	}
	return nil
}

// createPromptAndCommentForm creates the final prompt and comment form (reused from SimpleHabitCreator)
func (m *ElasticHabitCreator) createPromptAndCommentForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Habit Prompt").
				Description("The question asked when tracking this elastic habit").
				Value(&m.prompt).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("prompt cannot be empty")
					}
					return nil
				}),

			huh.NewInput().
				Title("Additional Comment (optional)").
				Description("Optional comment or context for this habit").
				Value(&m.comment).
				Placeholder("Any additional notes about this habit..."),
		),
	)
}

// AIDEV-NOTE: elastic-habit-builder; constructs habit with mini/midi/maxi criteria validation
// createHabitFromData creates a models.Habit from the collected form data
func (m *ElasticHabitCreator) createHabitFromData() (*models.Habit, error) {
	// Build field type configuration (reuses SimpleHabitCreator patterns)
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

	habit := &models.Habit{
		Title:       strings.TrimSpace(m.title),
		Description: strings.TrimSpace(m.description),
		HabitType:   m.habitType,
		FieldType:   fieldType,
		ScoringType: m.scoringType,
		Prompt:      strings.TrimSpace(m.prompt),
	}

	// Add comment if provided
	if comment := strings.TrimSpace(m.comment); comment != "" {
		// Note: Comment field doesn't exist in models.Habit yet - this is a design decision point
		// For now, we could append it to description or add it as HelpText
		if habit.Description != "" {
			habit.Description = habit.Description + "\n\nComment: " + comment
		} else {
			habit.Description = "Comment: " + comment
		}
	}

	// Add three-tier criteria for automatic scoring
	if m.scoringType == models.AutomaticScoring {
		miniCriteria, err := m.buildCriteriaFromData("mini")
		if err != nil {
			return nil, fmt.Errorf("failed to build mini criteria: %w", err)
		}
		habit.MiniCriteria = miniCriteria

		midiCriteria, err := m.buildCriteriaFromData("midi")
		if err != nil {
			return nil, fmt.Errorf("failed to build midi criteria: %w", err)
		}
		habit.MidiCriteria = midiCriteria

		maxiCriteria, err := m.buildCriteriaFromData("maxi")
		if err != nil {
			return nil, fmt.Errorf("failed to build maxi criteria: %w", err)
		}
		habit.MaxiCriteria = maxiCriteria
	}

	return habit, nil
}

// AIDEV-NOTE: elastic-criteria-builder; converts three-tier form data to models.Condition structures
// buildCriteriaFromData creates criteria based on the collected criteria configuration for specific tier
func (m *ElasticHabitCreator) buildCriteriaFromData(tier string) (*models.Criteria, error) {
	condition := &models.Condition{}
	var description string
	var criteriaValue, criteriaTimeValue string

	// Get the appropriate values based on tier
	switch tier {
	case "mini":
		criteriaValue = m.miniCriteriaValue
		criteriaTimeValue = m.miniCriteriaTimeValue
	case "midi":
		criteriaValue = m.midiCriteriaValue
		criteriaTimeValue = m.midiCriteriaTimeValue
	case "maxi":
		criteriaValue = m.maxiCriteriaValue
		criteriaTimeValue = m.maxiCriteriaTimeValue
	default:
		return nil, fmt.Errorf("unknown tier: %s", tier)
	}

	switch m.selectedFieldType {
	case "numeric", models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType:
		// Numeric criteria - all tiers use greater_than_or_equal for elastic habits
		unit := m.unit
		if unit == "" {
			unit = "units"
		}

		val, err := strconv.ParseFloat(strings.TrimSpace(criteriaValue), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid %s criteria value: %w", tier, err)
		}
		condition.GreaterThanOrEqual = &val
		description = fmt.Sprintf("%s achievement when value >= %.1f %s", strings.ToUpper(tier[:1])+tier[1:], val, unit)

	case models.TimeFieldType:
		// Time criteria - for elastic habits, we typically use "before" (e.g., wake up before X)
		// But this depends on the habit direction - this is simplified
		timeValue := strings.TrimSpace(criteriaTimeValue)
		condition.Before = timeValue
		description = fmt.Sprintf("%s achievement when time is before %s", strings.ToUpper(tier[:1])+tier[1:], timeValue)

	case models.DurationFieldType:
		// Duration criteria - use greater_than_or_equal approach (similar to numeric)
		durationValue := strings.TrimSpace(criteriaValue)
		condition.After = durationValue // Using After field for duration >= comparison
		description = fmt.Sprintf("%s achievement when duration >= %s", strings.ToUpper(tier[:1])+tier[1:], durationValue)

	default:
		return nil, fmt.Errorf("automatic scoring not supported for field type: %s", m.selectedFieldType)
	}

	return &models.Criteria{
		Description: description,
		Condition:   condition,
	}, nil
}

// getResolvedFieldType resolves the field type (handles "numeric" -> specific numeric type)
func (m *ElasticHabitCreator) getResolvedFieldType() string {
	if m.selectedFieldType == "numeric" {
		return m.numericSubtype
	}
	return m.selectedFieldType
}
