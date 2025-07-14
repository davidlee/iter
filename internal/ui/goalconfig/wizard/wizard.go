package wizard

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"davidlee/vice/internal/models"
)

// GoalWizardModel is the main bubbletea model for the goal creation wizard
type GoalWizardModel struct {
	state      State
	navigation NavigationController
	renderer   FormRenderer
	steps      []StepHandler

	// Current form state
	currentForm *huh.Form
	formActive  bool

	// Result
	result *Result
	done   bool

	// UI state
	width  int
	height int
}

// NewGoalWizardModel creates a new goal wizard model
func NewGoalWizardModel(goalType models.GoalType, _ []models.Goal) *GoalWizardModel {
	state := NewGoalState(goalType)
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

// NewGoalWizardModelWithBasicInfo creates a new goal wizard model with pre-populated basic info
func NewGoalWizardModelWithBasicInfo(goalType models.GoalType, _ []models.Goal, title, description string) *GoalWizardModel {
	state := NewGoalState(goalType)
	navigation := NewDefaultNavigationController()
	renderer := NewDefaultFormRenderer()

	// Pre-populate basic info in state
	basicInfo := &BasicInfoStepData{
		Title:       title,
		Description: description,
		GoalType:    goalType,
		valid:       true,
	}
	state.SetStep(0, basicInfo)
	state.MarkStepCompleted(0)

	// Create step handlers based on goal type
	steps := createStepHandlers(goalType)

	// Start from step 1 since basic info is pre-populated
	state.SetCurrentStep(1)

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
func (m *GoalWizardModel) GetResult() *Result {
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
	m.result = &Result{
		Cancelled: true,
	}
	m.done = true
	return m, tea.Quit
}

func (m *GoalWizardModel) handleFinish() (tea.Model, tea.Cmd) {
	// Check if on confirmation step and confirmed
	if m.state.GetCurrentStep() == m.state.GetTotalSteps()-1 {
		if confirmationHandler, ok := m.getCurrentStepHandler().(*ConfirmationStepHandler); ok {
			if !confirmationHandler.IsConfirmed() {
				// User chose not to confirm, treat as cancellation
				m.result = &Result{
					Cancelled: true,
				}
				m.done = true
				return m, tea.Quit
			}
		}
	}

	// Validate all steps
	errors := m.state.Validate()
	if len(errors) > 0 {
		m.result = &Result{
			Error: fmt.Errorf("validation failed: %d errors", len(errors)),
		}
		m.done = true
		return m, tea.Quit
	}

	// Convert state to goal
	goal, err := m.state.ToGoal()
	if err != nil {
		m.result = &Result{
			Error: fmt.Errorf("failed to create goal: %w", err),
		}
		m.done = true
		return m, tea.Quit
	}

	m.result = &Result{
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

// AIDEV-TODO: Add new goal types and step handlers here
// To implement elastic/informational flows:
// 1. Add case for new goal type
// 2. Create step handlers following simple_steps.go pattern
// 3. Use PlaceholderStepHandler for unimplemented steps
// 4. Update total step count in state.go calculateTotalSteps()

// createStepHandlers creates the appropriate step handlers for the goal type
func createStepHandlers(goalType models.GoalType) []StepHandler {
	var handlers []StepHandler

	// All goal types start with basic info
	handlers = append(handlers, NewBasicInfoStepHandler(goalType))

	switch goalType {
	case models.SimpleGoal:
		handlers = append(handlers,
			NewScoringStepHandler(goalType),
			NewCriteriaStepHandler(goalType, "simple"),
			NewConfirmationStepHandler(goalType),
		)
	case models.ElasticGoal:
		handlers = append(handlers,
			NewFieldConfigStepHandler(goalType),      // Step 1: Field type & config
			NewScoringStepHandler(goalType),          // Step 2: Scoring type
			NewCriteriaStepHandler(goalType, "mini"), // Step 3: Mini criteria
			NewCriteriaStepHandler(goalType, "midi"), // Step 4: Midi criteria
			NewCriteriaStepHandler(goalType, "maxi"), // Step 5: Maxi criteria
			NewValidationStepHandler(goalType),       // Step 6: Validation
			NewConfirmationStepHandler(goalType),     // Step 7: Confirmation
		)
	case models.InformationalGoal:
		handlers = append(handlers,
			NewFieldConfigStepHandler(goalType),  // Step 1: Field config & direction
			NewConfirmationStepHandler(goalType), // Step 2: Confirmation
		)
	}

	return handlers
}

// PlaceholderStepHandler is a minimal implementation for steps not yet implemented
type PlaceholderStepHandler struct {
	title       string
	description string
}

// Render renders a placeholder message
func (h *PlaceholderStepHandler) Render(_ State) string {
	return fmt.Sprintf("=== %s ===\n\n[Step implementation pending]\n\nPress 'n' to continue.", h.title)
}

// Update handles messages (placeholder)
func (h *PlaceholderStepHandler) Update(_ tea.Msg, state State) (State, tea.Cmd) {
	return state, nil
}

// Validate validates step data (placeholder)
func (h *PlaceholderStepHandler) Validate(_ State) []ValidationError {
	return nil
}

// CanNavigateFrom checks if we can leave this step (placeholder)
func (h *PlaceholderStepHandler) CanNavigateFrom(_ State) bool {
	return true
}

// CanNavigateTo checks if we can enter this step (placeholder)
func (h *PlaceholderStepHandler) CanNavigateTo(_ State) bool {
	return true
}

// GetTitle returns the step title
func (h *PlaceholderStepHandler) GetTitle() string {
	return h.title
}

// GetDescription returns the step description
func (h *PlaceholderStepHandler) GetDescription() string {
	return h.description
}
