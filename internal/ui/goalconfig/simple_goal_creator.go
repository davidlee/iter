package goalconfig

import (
	"fmt"
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

	// Form data - bound directly to form fields per huh documentation
	title       string
	description string
	goalType    models.GoalType
	scoringType models.ScoringType
	prompt      string
}

// NewSimpleGoalCreator creates a new simple goal creator with pre-populated basic info
func NewSimpleGoalCreator(title, description string, goalType models.GoalType) *SimpleGoalCreator {
	creator := &SimpleGoalCreator{
		title:       title,
		description: description,
		goalType:    goalType,
		prompt:      "Did you accomplish this goal today?", // Default prompt
	}

	// Create sequential form following huh patterns
	creator.form = huh.NewForm(
		// Step 1: Scoring type selection (only step needed after basic info)
		huh.NewGroup(
			huh.NewSelect[models.ScoringType]().
				Key("scoring").
				Title("Scoring Type").
				Description("How should goal achievement be determined?").
				Options(
					huh.NewOption("Manual (I'll mark completion myself)", models.ManualScoring),
					huh.NewOption("Automatic (Based on criteria I define)", models.AutomaticScoring),
				).
				Value(&creator.scoringType),
		),

		// Step 2: Custom prompt for manual scoring (conditional)
		huh.NewGroup(
			huh.NewInput().
				Key("prompt").
				Title("Goal Prompt").
				Description("The question asked when tracking this goal").
				Value(&creator.prompt).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("prompt cannot be empty")
					}
					return nil
				}),
		).WithHideFunc(func() bool {
			// Hide prompt step if automatic scoring (will need criteria instead)
			return creator.scoringType == models.AutomaticScoring
		}),
	)

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

	// Check if form is completed
	if m.form.State == huh.StateCompleted {
		// Create goal from collected data
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

// createGoalFromData creates a models.Goal from the collected form data
func (m *SimpleGoalCreator) createGoalFromData() (*models.Goal, error) {
	// AIDEV-NOTE: Create goal structure matching expected YAML format from user testing
	// Expected structure:
	//   - title: Title
	//     id: title 
	//     goal_type: simple
	//     field_type:
	//       type: boolean
	//     scoring_type: manual
	//     prompt: Prompt text here?

	goal := &models.Goal{
		Title:       strings.TrimSpace(m.title),
		Description: strings.TrimSpace(m.description),
		GoalType:    m.goalType,
		FieldType: models.FieldType{
			Type: models.BooleanFieldType, // Simple goals are always boolean
		},
		ScoringType: m.scoringType,
	}

	// Add prompt for manual scoring
	if m.scoringType == models.ManualScoring {
		goal.Prompt = strings.TrimSpace(m.prompt)
	}

	// TODO: Add automatic criteria configuration for automatic scoring
	// For now, focus on manual scoring path

	return goal, nil
}