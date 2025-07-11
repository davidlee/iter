package wizard

import (
	tea "github.com/charmbracelet/bubbletea"
)

// DefaultNavigationController implements NavigationController
type DefaultNavigationController struct{}

// NewDefaultNavigationController creates a new navigation controller
func NewDefaultNavigationController() *DefaultNavigationController {
	return &DefaultNavigationController{}
}

func (n *DefaultNavigationController) CanGoBack(state WizardState) bool {
	return state.GetCurrentStep() > 0
}

func (n *DefaultNavigationController) CanGoForward(state WizardState) bool {
	currentStep := state.GetCurrentStep()
	
	// Can go forward if:
	// 1. Not on the last step
	// 2. Current step is completed or can be skipped
	if currentStep >= state.GetTotalSteps()-1 {
		return false
	}
	
	// Check if current step is valid for navigation
	return state.IsStepCompleted(currentStep) || n.canSkipStep(currentStep, state)
}

func (n *DefaultNavigationController) CanGoToStep(index int, state WizardState) bool {
	// Can only jump to steps that are:
	// 1. Within valid range
	// 2. Already completed OR the next logical step
	if index < 0 || index >= state.GetTotalSteps() {
		return false
	}
	
	// Can always go to completed steps
	if state.IsStepCompleted(index) {
		return true
	}
	
	// Can go to the immediate next step if all previous steps are completed
	currentStep := state.GetCurrentStep()
	if index == currentStep+1 {
		return n.allPreviousStepsCompleted(index, state)
	}
	
	return false
}

func (n *DefaultNavigationController) GoBack() tea.Cmd {
	return func() tea.Msg {
		return NavigateBackMsg{}
	}
}

func (n *DefaultNavigationController) GoForward() tea.Cmd {
	return func() tea.Msg {
		return NavigateForwardMsg{}
	}
}

func (n *DefaultNavigationController) GoToStep(index int) tea.Cmd {
	return func() tea.Msg {
		return NavigateToStepMsg{Step: index}
	}
}

func (n *DefaultNavigationController) Cancel() tea.Cmd {
	return func() tea.Msg {
		return CancelWizardMsg{}
	}
}

func (n *DefaultNavigationController) Finish() tea.Cmd {
	return func() tea.Msg {
		return FinishWizardMsg{}
	}
}

// Helper methods

func (n *DefaultNavigationController) canSkipStep(_ int, state WizardState) bool {
	// Some steps can be skipped based on previous choices
	// For example, criteria steps can be skipped if manual scoring is selected
	
	// This logic would be implemented based on the specific step dependencies
	// For now, we don't allow skipping incomplete steps
	return false
}

func (n *DefaultNavigationController) allPreviousStepsCompleted(stepIndex int, state WizardState) bool {
	for i := 0; i < stepIndex; i++ {
		if !state.IsStepCompleted(i) && !n.canSkipStep(i, state) {
			return false
		}
	}
	return true
}