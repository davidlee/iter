package wizard

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"davidlee/iter/internal/models"
)

// GoalWizardModel is the main bubbletea model for the goal creation wizard
type GoalWizardModel struct {
	state      WizardState
	navigation NavigationController
	renderer   FormRenderer
	steps      []StepHandler
	
	// Current form state
	currentForm *huh.Form
	formActive  bool
	
	// Result
	result *WizardResult
	done   bool
	
	// UI state
	width  int
	height int
}

// NewGoalWizardModel creates a new goal wizard model
func NewGoalWizardModel(goalType models.GoalType, _ []models.Goal) *GoalWizardModel {
	state := NewGoalWizardState(goalType)
	navigation := NewDefaultNavigationController()
	renderer := NewDefaultFormRenderer()
	
	// Create step handlers based on goal type
	steps := createStepHandlers(goalType)
	
	return &GoalWizardModel{
		state:      state,
		navigation: navigation,
		renderer:   renderer,
		steps:      steps,
		formActive: false,
		done:       false,
		width:      80,
		height:     24,
	}
}

// Init implements tea.Model
func (m *GoalWizardModel) Init() tea.Cmd {
	// Initialize the first step
	return m.initCurrentStep()
}

// Update implements tea.Model
func (m *GoalWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case NavigateBackMsg:
		return m.handleNavigateBack()
		
	case NavigateForwardMsg:
		return m.handleNavigateForward()
		
	case NavigateToStepMsg:
		return m.handleNavigateToStep(msg.Step)
		
	case CancelWizardMsg:
		return m.handleCancel()
		
	case FinishWizardMsg:
		return m.handleFinish()
		
	case StepCompletedMsg:
		return m.handleStepCompleted(msg.Step)
		
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
		
	default:
		// Delegate to current step handler if form is not active
		if !m.formActive && m.getCurrentStepHandler() != nil {
			newState, cmd := m.getCurrentStepHandler().Update(msg, m.state)
			m.state = newState
			return m, cmd
		}
		
		// If form is active, delegate to the form
		if m.currentForm != nil && m.formActive {
			form, cmd := m.currentForm.Update(msg)
			if f, ok := form.(*huh.Form); ok {
				m.currentForm = f
				
				// Check if form is completed
				if m.currentForm.State == huh.StateCompleted {
					return m.handleFormCompleted()
				}
			}
			return m, cmd
		}
	}
	
	return m, nil
}

// View implements tea.Model
func (m *GoalWizardModel) View() string {
	if m.done && m.result != nil {
		if m.result.Cancelled {
			return "Goal creation cancelled.\n"
		}
		if m.result.Error != nil {
			return fmt.Sprintf("Error: %v\n", m.result.Error)
		}
		return "Goal created successfully!\n"
	}
	
	// Get current step handler
	stepHandler := m.getCurrentStepHandler()
	if stepHandler == nil {
		return "Error: Invalid step\n"
	}
	
	// Render the step using the renderer
	stepView := stepHandler.Render(m.state)
	
	// Add navigation
	navView := m.renderer.RenderNavigation(m.navigation, m.state)
	
	// Add validation errors if any
	errors := stepHandler.Validate(m.state)
	errorView := m.renderer.RenderValidationErrors(errors)
	
	return stepView + "\n" + errorView + "\n" + navView
}

// GetResult returns the wizard result (call after wizard is done)
func (m *GoalWizardModel) GetResult() *WizardResult {
	return m.result
}

// IsDone returns true when the wizard is finished
func (m *GoalWizardModel) IsDone() bool {
	return m.done
}

// Helper methods

func (m *GoalWizardModel) getCurrentStepHandler() StepHandler {
	currentStep := m.state.GetCurrentStep()
	if currentStep < 0 || currentStep >= len(m.steps) {
		return nil
	}
	return m.steps[currentStep]
}

func (m *GoalWizardModel) initCurrentStep() tea.Cmd {
	// Initialize the current step (create forms, etc.)
	return nil
}

func (m *GoalWizardModel) handleNavigateBack() (tea.Model, tea.Cmd) {
	if m.navigation.CanGoBack(m.state) {
		currentStep := m.state.GetCurrentStep()
		m.state.SetCurrentStep(currentStep - 1)
		return m, m.initCurrentStep()
	}
	return m, nil
}

func (m *GoalWizardModel) handleNavigateForward() (tea.Model, tea.Cmd) {
	if m.navigation.CanGoForward(m.state) {
		// Validate current step before moving forward
		stepHandler := m.getCurrentStepHandler()
		if stepHandler != nil && !stepHandler.CanNavigateFrom(m.state) {
			return m, nil
		}
		
		// Mark current step as completed
		currentStep := m.state.GetCurrentStep()
		m.state.MarkStepCompleted(currentStep)
		
		// Move to next step
		m.state.SetCurrentStep(currentStep + 1)
		return m, m.initCurrentStep()
	}
	return m, nil
}

func (m *GoalWizardModel) handleNavigateToStep(targetStep int) (tea.Model, tea.Cmd) {
	if m.navigation.CanGoToStep(targetStep, m.state) {
		m.state.SetCurrentStep(targetStep)
		return m, m.initCurrentStep()
	}
	return m, nil
}

func (m *GoalWizardModel) handleCancel() (tea.Model, tea.Cmd) {
	m.result = &WizardResult{
		Cancelled: true,
	}
	m.done = true
	return m, tea.Quit
}

func (m *GoalWizardModel) handleFinish() (tea.Model, tea.Cmd) {
	// Validate all steps
	errors := m.state.Validate()
	if len(errors) > 0 {
		m.result = &WizardResult{
			Error: fmt.Errorf("validation failed: %d errors", len(errors)),
		}
		m.done = true
		return m, tea.Quit
	}
	
	// Convert state to goal
	goal, err := m.state.ToGoal()
	if err != nil {
		m.result = &WizardResult{
			Error: fmt.Errorf("failed to create goal: %w", err),
		}
		m.done = true
		return m, tea.Quit
	}
	
	m.result = &WizardResult{
		Goal: goal,
	}
	m.done = true
	return m, tea.Quit
}

func (m *GoalWizardModel) handleStepCompleted(step int) (tea.Model, tea.Cmd) {
	m.state.MarkStepCompleted(step)
	return m, nil
}

func (m *GoalWizardModel) handleFormCompleted() (tea.Model, tea.Cmd) {
	m.formActive = false
	// Extract data from completed form and store in state
	// This would be implemented based on the specific form type
	return m, nil
}

func (m *GoalWizardModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m.handleCancel()
	case "b":
		if !m.formActive {
			return m.handleNavigateBack()
		}
	case "n":
		if !m.formActive {
			return m.handleNavigateForward()
		}
	case "f":
		if !m.formActive && m.state.GetCurrentStep() == m.state.GetTotalSteps()-1 {
			return m.handleFinish()
		}
	}
	
	return m, nil
}

// createStepHandlers creates the appropriate step handlers for the goal type
func createStepHandlers(goalType models.GoalType) []StepHandler {
	var handlers []StepHandler
	
	// All goal types start with basic info
	handlers = append(handlers, NewBasicInfoStepHandler())
	
	switch goalType {
	case models.SimpleGoal:
		handlers = append(handlers, 
			NewScoringStepHandler(),
			NewCriteriaStepHandler("simple"),
			NewConfirmationStepHandler(),
		)
	case models.ElasticGoal:
		handlers = append(handlers,
			NewFieldConfigStepHandler(),
			NewScoringStepHandler(),
			NewCriteriaStepHandler("mini"),
			NewCriteriaStepHandler("midi"),
			NewCriteriaStepHandler("maxi"),
			NewValidationStepHandler(),
			NewConfirmationStepHandler(),
		)
	case models.InformationalGoal:
		handlers = append(handlers,
			NewFieldConfigStepHandler(),
			NewConfirmationStepHandler(),
		)
	}
	
	return handlers
}

// Placeholder step handler constructors - these would be implemented
func NewBasicInfoStepHandler() StepHandler {
	return &PlaceholderStepHandler{title: "Basic Information"}
}

func NewFieldConfigStepHandler() StepHandler {
	return &PlaceholderStepHandler{title: "Field Configuration"}
}

func NewScoringStepHandler() StepHandler {
	return &PlaceholderStepHandler{title: "Scoring Configuration"}
}

func NewCriteriaStepHandler(level string) StepHandler {
	return &PlaceholderStepHandler{title: fmt.Sprintf("Criteria - %s", level)}
}

func NewValidationStepHandler() StepHandler {
	return &PlaceholderStepHandler{title: "Validation"}
}

func NewConfirmationStepHandler() StepHandler {
	return &PlaceholderStepHandler{title: "Confirmation"}
}

// PlaceholderStepHandler is a minimal implementation for testing
type PlaceholderStepHandler struct {
	title       string
	description string
}

func (h *PlaceholderStepHandler) Render(state WizardState) string {
	return fmt.Sprintf("=== %s ===\n\n[Step content placeholder]", h.title)
}

func (h *PlaceholderStepHandler) Update(msg tea.Msg, state WizardState) (WizardState, tea.Cmd) {
	return state, nil
}

func (h *PlaceholderStepHandler) Validate(state WizardState) []ValidationError {
	return nil
}

func (h *PlaceholderStepHandler) CanNavigateFrom(state WizardState) bool {
	return true
}

func (h *PlaceholderStepHandler) CanNavigateTo(state WizardState) bool {
	return true
}

func (h *PlaceholderStepHandler) GetTitle() string {
	return h.title
}

func (h *PlaceholderStepHandler) GetDescription() string {
	return h.description
}