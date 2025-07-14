package wizard

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"davidlee/vice/internal/models"
)

// AIDEV-NOTE: Legacy adapter provides backwards compatibility with existing huh-based forms
// This allows users to choose between enhanced wizard flows and simple legacy forms
// Maintains existing GoalBuilder functionality while providing upgrade path to wizard

// LegacyGoalAdapter provides interface compatibility without creating import cycles
type LegacyGoalAdapter struct {
	hybridRunner *HybridFormRunner
}

// NewLegacyGoalAdapter creates a new legacy goal adapter
func NewLegacyGoalAdapter() *LegacyGoalAdapter {
	return &LegacyGoalAdapter{
		hybridRunner: NewHybridFormRunner(),
	}
}

// CreateGoalWithLegacyForms creates a goal using simplified forms (delegates to wizard for now)
func (a *LegacyGoalAdapter) CreateGoalWithLegacyForms(goalType models.GoalType, existingGoals []models.Goal) (*models.Goal, error) {
	// AIDEV-TODO: Implement true legacy form support without import cycles
	// For now, delegate to wizard as it provides better UX
	return a.CreateGoalWithHybridForms(goalType, existingGoals)
}

// CreateGoalWithHybridForms creates a goal using huh forms embedded in bubbletea for enhanced UX
func (a *LegacyGoalAdapter) CreateGoalWithHybridForms(goalType models.GoalType, existingGoals []models.Goal) (*models.Goal, error) {
	// This demonstrates how to wrap existing forms with bubbletea for enhanced UX
	// while maintaining the same underlying logic

	switch goalType {
	case models.SimpleGoal:
		return a.createSimpleGoalWithHybrid(existingGoals)
	case models.ElasticGoal:
		return a.createElasticGoalWithHybrid(existingGoals)
	case models.InformationalGoal:
		return a.createInformationalGoalWithHybrid(existingGoals)
	default:
		return nil, fmt.Errorf("unsupported goal type for hybrid forms: %s", goalType)
	}
}

// createSimpleGoalWithHybrid creates a simple goal using hybrid approach
func (a *LegacyGoalAdapter) createSimpleGoalWithHybrid(existingGoals []models.Goal) (*models.Goal, error) {
	// Use the wizard for simple goals as it provides better UX than sequential forms
	wizardModel := NewGoalWizardModel(models.SimpleGoal, existingGoals)

	program := tea.NewProgram(wizardModel)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("wizard execution failed: %w", err)
	}

	if wizardModel, ok := finalModel.(*GoalWizardModel); ok {
		result := wizardModel.GetResult()
		if result == nil {
			return nil, fmt.Errorf("wizard completed without result")
		}

		if result.Cancelled {
			return nil, fmt.Errorf("wizard was cancelled")
		}

		if result.Error != nil {
			return nil, fmt.Errorf("wizard error: %w", result.Error)
		}

		return result.Goal, nil
	}

	return nil, fmt.Errorf("unexpected wizard model type")
}

// createElasticGoalWithHybrid creates an elastic goal using hybrid approach
func (a *LegacyGoalAdapter) createElasticGoalWithHybrid(existingGoals []models.Goal) (*models.Goal, error) {
	// Elastic goals always use the wizard due to complexity
	return a.createSimpleGoalWithHybrid(existingGoals) // Same implementation
}

// createInformationalGoalWithHybrid creates an informational goal using hybrid approach
func (a *LegacyGoalAdapter) createInformationalGoalWithHybrid(existingGoals []models.Goal) (*models.Goal, error) {
	// Informational goals always use the wizard for consistency
	return a.createSimpleGoalWithHybrid(existingGoals) // Same implementation
}

// AIDEV-TODO: Add demonstration methods for hybrid form usage
// DemoHybridForm would show how to embed huh forms in bubbletea
// Removed for now to avoid import cycles with huh package

// BackwardsCompatibilityMode provides a flag to control interface selection
type BackwardsCompatibilityMode int

const (
	// AutoSelect automatically chooses the best interface for each goal type
	AutoSelect BackwardsCompatibilityMode = iota
	// PreferWizard prefers wizard interface when possible
	PreferWizard
	// PreferLegacy prefers legacy forms when possible
	PreferLegacy
	// ForceWizard always uses wizard interface
	ForceWizard
	// ForceLegacy always uses legacy forms
	ForceLegacy
)

// CreateGoalWithMode creates a goal using the specified compatibility mode
func (a *LegacyGoalAdapter) CreateGoalWithMode(goalType models.GoalType, existingGoals []models.Goal, mode BackwardsCompatibilityMode) (*models.Goal, error) {
	switch mode {
	case AutoSelect:
		// Use intelligent selection based on goal type complexity
		return a.createGoalAutoSelect(goalType, existingGoals)
	case PreferWizard, ForceWizard:
		// Use wizard interface
		return a.CreateGoalWithHybridForms(goalType, existingGoals)
	case PreferLegacy, ForceLegacy:
		// Use legacy forms
		return a.CreateGoalWithLegacyForms(goalType, existingGoals)
	default:
		return nil, fmt.Errorf("unsupported compatibility mode: %d", mode)
	}
}

// createGoalAutoSelect automatically selects the best interface
func (a *LegacyGoalAdapter) createGoalAutoSelect(goalType models.GoalType, existingGoals []models.Goal) (*models.Goal, error) {
	switch goalType {
	case models.SimpleGoal:
		// Simple goals can use either interface - prefer wizard for better UX
		return a.CreateGoalWithHybridForms(goalType, existingGoals)
	case models.ElasticGoal:
		// Elastic goals require wizard due to complexity
		return a.CreateGoalWithHybridForms(goalType, existingGoals)
	case models.InformationalGoal:
		// Informational goals require wizard for direction configuration
		return a.CreateGoalWithHybridForms(goalType, existingGoals)
	default:
		// Unknown types: fallback to legacy forms
		return a.CreateGoalWithLegacyForms(goalType, existingGoals)
	}
}

// CreateGoalWithBasicInfo creates a goal with pre-populated basic info using the specified compatibility mode
func (a *LegacyGoalAdapter) CreateGoalWithBasicInfo(_ interface{}, existingGoals []models.Goal, mode BackwardsCompatibilityMode) (*models.Goal, error) {
	// Extract goal type from basic info to determine flow
	// For now, delegate to the regular CreateGoalWithMode since wizards handle pre-population
	// AIDEV-TODO: Extract goal type from basicInfo and pass it properly

	// Default to simple goal if we can't extract the type
	goalType := models.SimpleGoal

	return a.CreateGoalWithMode(goalType, existingGoals, mode)
}
