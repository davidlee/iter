// Package wizard provides bubbletea-based wizard components for goal configuration.
package wizard

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"davidlee/iter/internal/models"
)

// WizardState represents the complete state of the goal creation wizard
type WizardState interface {
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
	Render(state WizardState) string
	
	// Update handles tea messages and returns updated state and commands
	Update(msg tea.Msg, state WizardState) (WizardState, tea.Cmd)
	
	// Validate checks if the current step data is valid
	Validate(state WizardState) []ValidationError
	
	// CanNavigateFrom returns true if user can leave this step
	CanNavigateFrom(state WizardState) bool
	
	// CanNavigateTo returns true if user can enter this step
	CanNavigateTo(state WizardState) bool
	
	// GetTitle returns the title for this step
	GetTitle() string
	
	// GetDescription returns the description for this step
	GetDescription() string
}

// NavigationController manages wizard navigation
type NavigationController interface {
	CanGoBack(state WizardState) bool
	CanGoForward(state WizardState) bool
	CanGoToStep(index int, state WizardState) bool
	GoBack() tea.Cmd
	GoForward() tea.Cmd
	GoToStep(index int) tea.Cmd
	Cancel() tea.Cmd
	Finish() tea.Cmd
}

// HuhFormStep wraps a huh form for use in bubbletea wizard
type HuhFormStep struct {
	form        *huh.Form
	title       string
	description string
	validator   func(interface{}) []ValidationError
	onComplete  func(result interface{}) StepData
	isActive    bool
	isCompleted bool
}

// FormRenderer handles rendering of forms and wizard chrome
type FormRenderer interface {
	RenderForm(step HuhFormStep, state WizardState) string
	RenderProgress(current, total int, completedSteps []int) string
	RenderNavigation(nav NavigationController, state WizardState) string
	RenderSummary(state WizardState) string
	RenderValidationErrors(errors []ValidationError) string
}

// ValidationCollector manages validation across wizard steps
type ValidationCollector interface {
	CollectErrors(state WizardState) []ValidationError
	ValidateStep(stepIndex int, data StepData) []ValidationError
	ValidateCrossStep(state WizardState) []ValidationError
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
	StepPending StepStatus = iota
	StepInProgress
	StepCompleted
	StepError
)

// Tea messages for wizard navigation
type (
	NavigateBackMsg    struct{}
	NavigateForwardMsg struct{}
	NavigateToStepMsg  struct{ Step int }
	CancelWizardMsg    struct{}
	FinishWizardMsg    struct{}
	StepCompletedMsg   struct{ Step int }
	ValidationErrorMsg struct{ Errors []ValidationError }
)

// WizardResult represents the final result of the wizard
type WizardResult struct {
	Goal      *models.Goal
	Cancelled bool
	Error     error
}