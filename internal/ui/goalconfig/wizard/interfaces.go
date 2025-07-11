// Package wizard provides bubbletea-based wizard components for goal configuration.
//
// AIDEV-NOTE: Core wizard interfaces - extend here for new step types or wizard patterns
// Key extension points:
// - StepHandler interface: implement for new step types (elastic field config, validation steps)
// - State interface: extend for additional wizard types beyond goal configuration
// - FormRenderer interface: customize for different visual themes or layouts
package wizard

import (
	tea "github.com/charmbracelet/bubbletea"

	"davidlee/iter/internal/models"
)

// State represents the complete state of the goal creation wizard
type State interface {
	GetStep(index int) StepData
	SetStep(index int, data StepData)
	Validate() []ValidationError
	ToGoal() (*models.Goal, error)
	Serialize() ([]byte, error)
	Deserialize([]byte) error
	GetCurrentStep() int
	SetCurrentStep(int)
	GetTotalSteps() int
	IsStepCompleted(index int) bool
	MarkStepCompleted(index int)
}

// StepHandler defines the interface for individual wizard steps
type StepHandler interface {
	// Render returns the string representation of this step
	Render(state State) string
	
	// Update handles tea messages and returns updated state and commands
	Update(msg tea.Msg, state State) (State, tea.Cmd)
	
	// Validate checks if the current step data is valid
	Validate(state State) []ValidationError
	
	// CanNavigateFrom returns true if user can leave this step
	CanNavigateFrom(state State) bool
	
	// CanNavigateTo returns true if user can enter this step
	CanNavigateTo(state State) bool
	
	// GetTitle returns the title for this step
	GetTitle() string
	
	// GetDescription returns the description for this step
	GetDescription() string
}

// NavigationController manages wizard navigation
type NavigationController interface {
	CanGoBack(state State) bool
	CanGoForward(state State) bool
	CanGoToStep(index int, state State) bool
	GoBack() tea.Cmd
	GoForward() tea.Cmd
	GoToStep(index int) tea.Cmd
	Cancel() tea.Cmd
	Finish() tea.Cmd
}

// FormRenderer handles rendering of forms and wizard chrome
type FormRenderer interface {
	RenderProgress(current, total int, completedSteps []int) string
	RenderNavigation(nav NavigationController, state State) string
	RenderSummary(state State) string
	RenderValidationErrors(errors []ValidationError) string
}

// ValidationCollector manages validation across wizard steps
type ValidationCollector interface {
	CollectErrors(state State) []ValidationError
	ValidateStep(stepIndex int, data StepData) []ValidationError
	ValidateCrossStep(state State) []ValidationError
}

// ProgressTracker manages step completion and progress
type ProgressTracker interface {
	GetCurrentStep() int
	GetTotalSteps() int
	GetCompletedSteps() []int
	GetStepStatus(index int) StepStatus
	MarkStepCompleted(index int)
	MarkStepInProgress(index int)
	MarkStepPending(index int)
}

// Supporting types

// StepData represents the data for a single wizard step
type StepData interface {
	IsValid() bool
	GetData() interface{}
	SetData(interface{}) error
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string
	Message string
	Step    int
}

// StepStatus represents the status of a wizard step
type StepStatus int

const (
	// StepPending indicates the step hasn't been started
	StepPending StepStatus = iota
	// StepInProgress indicates the step is currently active
	StepInProgress
	// StepCompleted indicates the step has been finished
	StepCompleted
	// StepError indicates the step has an error
	StepError
)

// NavigateBackMsg requests navigation to the previous step
type NavigateBackMsg struct{}

// NavigateForwardMsg requests navigation to the next step
type NavigateForwardMsg struct{}

// NavigateToStepMsg requests navigation to a specific step
type NavigateToStepMsg struct{ Step int }

// CancelWizardMsg requests wizard cancellation
type CancelWizardMsg struct{}

// FinishWizardMsg requests wizard completion
type FinishWizardMsg struct{}

// StepCompletedMsg indicates a step has been completed
type StepCompletedMsg struct{ Step int }

// ValidationErrorMsg contains validation errors
type ValidationErrorMsg struct{ Errors []ValidationError }

// Result represents the final result of the wizard
type Result struct {
	Goal      *models.Goal
	Cancelled bool
	Error     error
}