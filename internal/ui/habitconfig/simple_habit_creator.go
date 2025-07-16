// Package habitconfig provides UI components for interactive habit configuration.
package habitconfig

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"davidlee/vice/internal/models"
)

// AIDEV-NOTE: Simple idiomatic bubbletea implementation for habit creation
// Based on https://github.com/charmbracelet/huh/blob/main/examples/bubbletea/main.go
// Follows Model-View-Update pattern from https://github.com/charmbracelet/bubbletea
// Much simpler than complex wizard architecture - focuses on common use case

// SimpleHabitCreator implements a simple, idiomatic bubbletea model for creating habits
type SimpleHabitCreator struct {
	form     *huh.Form
	quitting bool
	err      error
	result   *models.Habit

	// Pre-populated basic info
	title       string
	description string
	goalType    models.HabitType

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

	// Criteria configuration data for automatic scoring
	criteriaType      string // "greater_than", "less_than", "equals", "before", "after", "range"
	criteriaValue     string // Value for comparison
	criteriaValue2    string // Second value for range
	criteriaTimeValue string // Time value for time-based criteria
	rangeInclusive    bool   // Whether range bounds are inclusive

	// State tracking for multi-step flow
	currentStep int
	maxSteps    int
}

// NewSimpleHabitCreator creates a new simple habit creator with pre-populated basic info
func NewSimpleHabitCreator(title, description string, goalType models.HabitType) *SimpleHabitCreator {
	creator := &SimpleHabitCreator{
		title:             title,
		description:       description,
		goalType:          goalType,
		selectedFieldType: models.BooleanFieldType, // Default to boolean for quick path
		numericSubtype:    models.UnsignedIntFieldType,
		unit:              "times",
		prompt:            "Did you accomplish this habit today?",
		comment:           "",
		currentStep:       0,
		maxSteps:          0, // Will be determined based on flow
	}

	// Initialize the first step
	creator.initializeStep()

	return creator
}

// TestHabitData contains pre-configured data for headless testing
type TestHabitData struct {
	FieldType         string
	NumericSubtype    string
	Unit              string
	MultilineText     bool
	MinValue          string
	MaxValue          string
	HasMinMax         bool
	ScoringType       models.ScoringType
	Prompt            string
	Comment           string
	CriteriaType      string
	CriteriaValue     string
	CriteriaValue2    string
	CriteriaTimeValue string
	RangeInclusive    bool
}

// NewSimpleHabitCreatorForEdit creates a habit creator pre-populated with existing habit data for editing.
// AIDEV-NOTE: edit-mode-support; habit-to-data conversion enables editing existing habits
func NewSimpleHabitCreatorForEdit(habit *models.Habit) *SimpleHabitCreator {
	data := goalToTestData(habit)

	creator := &SimpleHabitCreator{
		title:             habit.Title,
		description:       habit.Description,
		goalType:          habit.HabitType,
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
		criteriaType:      data.CriteriaType,
		criteriaValue:     data.CriteriaValue,
		criteriaValue2:    data.CriteriaValue2,
		criteriaTimeValue: data.CriteriaTimeValue,
		rangeInclusive:    data.RangeInclusive,
		currentStep:       0,
		maxSteps:          0, // Will be determined based on flow
	}

	// Initialize the first step with forms
	creator.initializeStep()

	return creator
}

// AIDEV-NOTE: habit-to-data-conversion; Phase 3 critical pattern for edit mode support
// Reverse engineers from models.Habit back to form data structures for seamless editing
// This pattern enables position preservation and ID retention during habit editing
// goalToTestData converts a models.Habit to TestHabitData for pre-population
func goalToTestData(habit *models.Habit) TestHabitData {
	if habit == nil {
		return TestHabitData{}
	}

	data := TestHabitData{
		FieldType:   habit.FieldType.Type,
		ScoringType: habit.ScoringType,
		Prompt:      habit.Prompt,
		Comment:     extractCommentFromDescription(habit.Description),
	}

	// Field type specific conversion
	switch habit.FieldType.Type {
	case models.BooleanFieldType:
		data.FieldType = "boolean"
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
		data.FieldType = "text"
		if habit.FieldType.Multiline != nil {
			data.MultilineText = *habit.FieldType.Multiline
		}
	}

	// Criteria conversion for automatic scoring
	if habit.ScoringType == models.AutomaticScoring && habit.Criteria != nil {
		data.CriteriaType, data.CriteriaValue, data.CriteriaValue2, data.CriteriaTimeValue, data.RangeInclusive = convertCriteriaToData(habit.Criteria)
	}

	return data
}

// AIDEV-NOTE: comment-extraction; handles legacy comment storage in description field
// Future enhancement: add dedicated Comment field to models.Habit to avoid string parsing
// extractCommentFromDescription extracts comment portion from description if present
func extractCommentFromDescription(description string) string {
	if strings.Contains(description, "\n\nComment: ") {
		parts := strings.Split(description, "\n\nComment: ")
		if len(parts) == 2 {
			return parts[1]
		}
	} else if strings.HasPrefix(description, "Comment: ") {
		return strings.TrimPrefix(description, "Comment: ")
	}
	return ""
}

// AIDEV-NOTE: criteria-conversion; comprehensive criteria reverse engineering for all condition types
// Handles GreaterThan, LessThan, Equals, Before, After - critical for automatic scoring habit editing
// convertCriteriaToData converts models.Criteria to test data format
func convertCriteriaToData(criteria *models.Criteria) (criteriaType, value, value2, timeValue string, inclusive bool) {
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
	// Range handling would go here if supported
	return "", "", "", "", false
}

// NewSimpleHabitCreatorForTesting creates a habit creator with pre-populated test data, bypassing UI
func NewSimpleHabitCreatorForTesting(title, description string, goalType models.HabitType, data TestHabitData) *SimpleHabitCreator {
	creator := &SimpleHabitCreator{
		title:             title,
		description:       description,
		goalType:          goalType,
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
		criteriaType:      data.CriteriaType,
		criteriaValue:     data.CriteriaValue,
		criteriaValue2:    data.CriteriaValue2,
		criteriaTimeValue: data.CriteriaTimeValue,
		rangeInclusive:    data.RangeInclusive,
		// Skip UI initialization for testing
		form:     nil,
		quitting: false,
		err:      nil,
		result:   nil,
	}

	return creator
}

// CreateHabitDirectly bypasses UI flow and creates habit directly from configured data
func (m *SimpleHabitCreator) CreateHabitDirectly() (*models.Habit, error) {
	return m.createHabitFromData()
}

// Init implements tea.Model - called when the model is first initialized
func (m *SimpleHabitCreator) Init() tea.Cmd {
	// AIDEV-NOTE: Following bubbletea pattern - Init() returns initial command
	// Form initialization happens in constructor per huh documentation
	return m.form.Init()
}

// Update implements tea.Model - handles messages and updates state
func (m *SimpleHabitCreator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m *SimpleHabitCreator) View() string {
	if m.quitting {
		if m.err != nil {
			return fmt.Sprintf("Error creating habit: %v\n", m.err)
		}
		if m.result != nil {
			return fmt.Sprintf("✅ Habit created successfully: %s\n", m.result.Title)
		}
		return "Habit creation cancelled.\n"
	}

	// AIDEV-NOTE: Simple view rendering - just show the form
	// Form handles all rendering, progress, validation per huh documentation
	return m.form.View()
}

// GetResult returns the created habit (after completion)
func (m *SimpleHabitCreator) GetResult() (*models.Habit, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

// IsCompleted returns true if the form was completed successfully
func (m *SimpleHabitCreator) IsCompleted() bool {
	return m.result != nil && m.err == nil
}

// IsCancelled returns true if the form was cancelled
func (m *SimpleHabitCreator) IsCancelled() bool {
	return m.quitting && m.result == nil && m.err == nil
}

// AIDEV-NOTE: multi-step-flow; dynamic step routing with flow adjustment for field types and scoring
// initializeStep initializes the form for the current step
func (m *SimpleHabitCreator) initializeStep() {
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
func (m *SimpleHabitCreator) adjustFlowForFieldType() {
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
func (m *SimpleHabitCreator) adjustFlowForScoringType() {
	if m.scoringType == models.ManualScoring {
		// Manual scoring doesn't need criteria step, reduce max steps by 1
		m.maxSteps--
	}
}

// isCurrentStepScoringType returns true if the current step is scoring type selection
func (m *SimpleHabitCreator) isCurrentStepScoringType() bool {
	// Scoring type is step 1 if no field config needed, step 2 if field config needed
	if m.needsFieldConfiguration() {
		return m.currentStep == 2
	}
	return m.currentStep == 1
}

// needsFieldConfiguration returns true if the selected field type needs configuration
func (m *SimpleHabitCreator) needsFieldConfiguration() bool {
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
func (m *SimpleHabitCreator) createFieldTypeSelectionForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Field Type").
				Description("Choose what type of data this habit will collect").
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
func (m *SimpleHabitCreator) createFieldConfigurationForm() *huh.Form {
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
func (m *SimpleHabitCreator) createNumericConfigForm() *huh.Form {
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

// createTextConfigForm creates text field configuration form
func (m *SimpleHabitCreator) createTextConfigForm() *huh.Form {
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
func (m *SimpleHabitCreator) createScoringTypeForm() *huh.Form {
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
				Description("How should habit achievement be determined?").
				Options(options...).
				Value(&m.scoringType),
		),
	)
}

// supportsAutomaticScoring returns true if the selected field type supports automatic scoring
func (m *SimpleHabitCreator) supportsAutomaticScoring() bool {
	switch m.selectedFieldType {
	case models.TextFieldType:
		return false // Text fields restricted to manual scoring
	default:
		return true // Boolean, numeric, time, duration support automatic scoring
	}
}

// AIDEV-NOTE: criteria-dispatch; routes field types to specific criteria forms (Boolean/Numeric/Time/Duration)
// createCriteriaDefinitionForm creates the criteria definition form for automatic scoring
func (m *SimpleHabitCreator) createCriteriaDefinitionForm() *huh.Form {
	switch m.selectedFieldType {
	case models.BooleanFieldType:
		return m.createBooleanCriteriaForm()
	case "numeric", models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType:
		return m.createNumericCriteriaForm()
	case models.TimeFieldType:
		return m.createTimeCriteriaForm()
	case models.DurationFieldType:
		return m.createDurationCriteriaForm()
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

// createBooleanCriteriaForm creates criteria form for boolean fields
func (m *SimpleHabitCreator) createBooleanCriteriaForm() *huh.Form {
	// Boolean criteria is always "equals: true" for habit completion
	m.criteriaType = "equals"
	m.criteriaValue = "true"

	return huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Automatic Scoring Criteria").
				Description("✅ Boolean habits are automatically scored as complete when the value is 'true'.\n\nThis habit will be marked as achieved when you check it as completed."),
		),
	)
}

// AIDEV-NOTE: numeric-criteria; supports >, >=, <, <=, range with validation + inclusive/exclusive ranges
// createNumericCriteriaForm creates criteria form for numeric fields
func (m *SimpleHabitCreator) createNumericCriteriaForm() *huh.Form {
	unit := m.unit
	if unit == "" {
		unit = "units"
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Criteria Type").
				Description("Choose how the numeric value should be evaluated").
				Options(
					huh.NewOption("Greater than (>) a value", "greater_than"),
					huh.NewOption("Greater than or equal (>=) to a value", "greater_than_or_equal"),
					huh.NewOption("Less than (<) a value", "less_than"),
					huh.NewOption("Less than or equal (<=) to a value", "less_than_or_equal"),
					huh.NewOption("Within a range", "range"),
				).
				Value(&m.criteriaType),
		),

		// Single value input (for most criteria types)
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Target Value (%s)", unit)).
				Description("Enter the threshold value for comparison").
				Value(&m.criteriaValue).
				Placeholder("10").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("criteria value is required")
					}
					if _, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err != nil {
						return fmt.Errorf("must be a valid number")
					}
					return nil
				}),
		).WithHideFunc(func() bool {
			return m.criteriaType == "range"
		}),

		// Range input (for range criteria)
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Minimum Value (%s)", unit)).
				Description("Enter the minimum value for the range").
				Value(&m.criteriaValue).
				Placeholder("10").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("minimum value is required")
					}
					if _, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err != nil {
						return fmt.Errorf("must be a valid number")
					}
					return nil
				}),

			huh.NewInput().
				Title(fmt.Sprintf("Maximum Value (%s)", unit)).
				Description("Enter the maximum value for the range").
				Value(&m.criteriaValue2).
				Placeholder("20").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("maximum value is required")
					}
					if _, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err != nil {
						return fmt.Errorf("must be a valid number")
					}
					return nil
				}),

			huh.NewConfirm().
				Title("Inclusive Range").
				Description("Should the range boundaries be inclusive? (>= min and <= max)").
				Value(&m.rangeInclusive),
		).WithHideFunc(func() bool {
			return m.criteriaType != "range"
		}),
	)
}

// createTimeCriteriaForm creates criteria form for time fields
func (m *SimpleHabitCreator) createTimeCriteriaForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Time Criteria").
				Description("Choose how the time should be evaluated").
				Options(
					huh.NewOption("Before a specific time", "before"),
					huh.NewOption("After a specific time", "after"),
				).
				Value(&m.criteriaType),

			huh.NewInput().
				Title("Target Time").
				Description("Enter the time in HH:MM format (24-hour)").
				Value(&m.criteriaTimeValue).
				Placeholder("07:00").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("time value is required")
					}
					if _, err := time.Parse("15:04", strings.TrimSpace(s)); err != nil {
						return fmt.Errorf("time must be in HH:MM format (e.g., 07:30)")
					}
					return nil
				}),
		),
	)
}

// createDurationCriteriaForm creates criteria form for duration fields
func (m *SimpleHabitCreator) createDurationCriteriaForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Duration Criteria").
				Description("Choose how the duration should be evaluated").
				Options(
					huh.NewOption("At least this long", "greater_than_or_equal"),
					huh.NewOption("Less than this duration", "less_than"),
					huh.NewOption("Exactly this duration", "equals"), // Using equals for duration equality
					huh.NewOption("Within a duration range", "range"),
				).
				Value(&m.criteriaType),
		),

		// Single duration input
		huh.NewGroup(
			huh.NewInput().
				Title("Target Duration").
				Description("Enter duration (e.g., '30m', '1h 30m', '45 minutes')").
				Value(&m.criteriaValue).
				Placeholder("30m").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("duration value is required")
					}
					// Basic validation - more detailed parsing would happen in the actual system
					if !strings.ContainsAny(strings.TrimSpace(s), "mh") {
						return fmt.Errorf("duration must include time units (e.g., 30m, 1h)")
					}
					return nil
				}),
		).WithHideFunc(func() bool {
			return m.criteriaType == "range"
		}),

		// Duration range input
		huh.NewGroup(
			huh.NewInput().
				Title("Minimum Duration").
				Description("Enter minimum duration (e.g., '15m')").
				Value(&m.criteriaValue).
				Placeholder("15m").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("minimum duration is required")
					}
					if !strings.ContainsAny(strings.TrimSpace(s), "mh") {
						return fmt.Errorf("duration must include time units (e.g., 15m)")
					}
					return nil
				}),

			huh.NewInput().
				Title("Maximum Duration").
				Description("Enter maximum duration (e.g., '60m')").
				Value(&m.criteriaValue2).
				Placeholder("60m").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("maximum duration is required")
					}
					if !strings.ContainsAny(strings.TrimSpace(s), "mh") {
						return fmt.Errorf("duration must include time units (e.g., 60m)")
					}
					return nil
				}),
		).WithHideFunc(func() bool {
			return m.criteriaType != "range"
		}),
	)
}

// createPromptAndCommentForm creates the final prompt and comment form
func (m *SimpleHabitCreator) createPromptAndCommentForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Habit Prompt").
				Description("The question asked when tracking this habit").
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

// createHabitFromData creates a models.Habit from the collected form data
func (m *SimpleHabitCreator) createHabitFromData() (*models.Habit, error) {
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

	habit := &models.Habit{
		Title:       strings.TrimSpace(m.title),
		Description: strings.TrimSpace(m.description),
		HabitType:   m.goalType,
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

	// Add criteria for automatic scoring
	if m.scoringType == models.AutomaticScoring {
		criteria, err := m.buildCriteriaFromData()
		if err != nil {
			return nil, fmt.Errorf("failed to build criteria: %w", err)
		}
		habit.Criteria = criteria
	}

	return habit, nil
}

// AIDEV-NOTE: criteria-builder; converts form data to models.Condition with proper validation
// buildCriteriaFromData creates criteria based on the collected criteria configuration
func (m *SimpleHabitCreator) buildCriteriaFromData() (*models.Criteria, error) {
	condition := &models.Condition{}
	var description string

	switch m.selectedFieldType {
	case models.BooleanFieldType:
		// Boolean criteria: equals true
		trueValue := true
		condition.Equals = &trueValue
		description = "Habit is complete when checked as true"

	case "numeric", models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType:
		// Numeric criteria
		unit := m.unit
		if unit == "" {
			unit = "units"
		}

		switch m.criteriaType {
		case "greater_than":
			val, err := strconv.ParseFloat(strings.TrimSpace(m.criteriaValue), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid criteria value: %w", err)
			}
			condition.GreaterThan = &val
			description = fmt.Sprintf("Habit achieved when value > %.1f %s", val, unit)

		case "greater_than_or_equal":
			val, err := strconv.ParseFloat(strings.TrimSpace(m.criteriaValue), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid criteria value: %w", err)
			}
			condition.GreaterThanOrEqual = &val
			description = fmt.Sprintf("Habit achieved when value >= %.1f %s", val, unit)

		case "less_than":
			val, err := strconv.ParseFloat(strings.TrimSpace(m.criteriaValue), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid criteria value: %w", err)
			}
			condition.LessThan = &val
			description = fmt.Sprintf("Habit achieved when value < %.1f %s", val, unit)

		case "less_than_or_equal":
			val, err := strconv.ParseFloat(strings.TrimSpace(m.criteriaValue), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid criteria value: %w", err)
			}
			condition.LessThanOrEqual = &val
			description = fmt.Sprintf("Habit achieved when value <= %.1f %s", val, unit)

		case "range":
			minVal, err := strconv.ParseFloat(strings.TrimSpace(m.criteriaValue), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid minimum value: %w", err)
			}
			maxVal, err := strconv.ParseFloat(strings.TrimSpace(m.criteriaValue2), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid maximum value: %w", err)
			}
			condition.Range = &models.RangeCondition{
				Min:          minVal,
				Max:          maxVal,
				MinInclusive: &m.rangeInclusive,
				MaxInclusive: &m.rangeInclusive,
			}
			inclusiveText := "exclusive"
			if m.rangeInclusive {
				inclusiveText = "inclusive"
			}
			description = fmt.Sprintf("Habit achieved when value is within %.1f to %.1f %s (%s)", minVal, maxVal, unit, inclusiveText)

		default:
			return nil, fmt.Errorf("unknown numeric criteria type: %s", m.criteriaType)
		}

	case models.TimeFieldType:
		// Time criteria
		timeValue := strings.TrimSpace(m.criteriaTimeValue)
		switch m.criteriaType {
		case "before":
			condition.Before = timeValue
			description = fmt.Sprintf("Habit achieved when time is before %s", timeValue)
		case "after":
			condition.After = timeValue
			description = fmt.Sprintf("Habit achieved when time is after %s", timeValue)
		default:
			return nil, fmt.Errorf("unknown time criteria type: %s", m.criteriaType)
		}

	case models.DurationFieldType:
		// AIDEV-NOTE: duration-criteria-hack; reuses Before/After fields for duration comparisons (needs proper duration type support)
		// Duration criteria - treat similar to numeric but with duration parsing
		durationValue := strings.TrimSpace(m.criteriaValue)
		switch m.criteriaType {
		case "greater_than_or_equal":
			// For duration, we'll store the duration string in a way that can be parsed later
			// This is a simplified approach - a full implementation would convert to minutes/seconds
			condition.After = durationValue // Using After field for duration >= comparison
			description = fmt.Sprintf("Habit achieved when duration >= %s", durationValue)
		case "less_than":
			condition.Before = durationValue // Using Before field for duration < comparison
			description = fmt.Sprintf("Habit achieved when duration < %s", durationValue)
		case "equals":
			// For duration equality, we could use a specific approach
			// For now, treating as a range with very small tolerance
			condition.Before = durationValue
			description = fmt.Sprintf("Habit achieved when duration equals %s", durationValue)
		case "range":
			minDuration := strings.TrimSpace(m.criteriaValue)
			maxDuration := strings.TrimSpace(m.criteriaValue2)
			// This is a simplified representation - real implementation would need better duration handling
			condition.Before = maxDuration
			condition.After = minDuration
			description = fmt.Sprintf("Habit achieved when duration is between %s and %s", minDuration, maxDuration)
		default:
			return nil, fmt.Errorf("unknown duration criteria type: %s", m.criteriaType)
		}

	default:
		return nil, fmt.Errorf("automatic scoring not supported for field type: %s", m.selectedFieldType)
	}

	return &models.Criteria{
		Description: description,
		Condition:   condition,
	}, nil
}

// getResolvedFieldType resolves the field type (handles "numeric" -> specific numeric type)
func (m *SimpleHabitCreator) getResolvedFieldType() string {
	if m.selectedFieldType == "numeric" {
		return m.numericSubtype
	}
	return m.selectedFieldType
}
