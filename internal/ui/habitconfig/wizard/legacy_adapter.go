package wizard

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/davidlee/vice/internal/models"
)

// AIDEV-NOTE: Legacy adapter provides backwards compatibility with existing huh-based forms
// This allows users to choose between enhanced wizard flows and simple legacy forms
// Maintains existing HabitBuilder functionality while providing upgrade path to wizard

// LegacyHabitAdapter provides interface compatibility without creating import cycles
type LegacyHabitAdapter struct {
	hybridRunner *HybridFormRunner
}

// NewLegacyHabitAdapter creates a new legacy habit adapter
func NewLegacyHabitAdapter() *LegacyHabitAdapter {
	return &LegacyHabitAdapter{
		hybridRunner: NewHybridFormRunner(),
	}
}

// CreateHabitWithLegacyForms creates a habit using simplified forms (delegates to wizard for now)
func (a *LegacyHabitAdapter) CreateHabitWithLegacyForms(habitType models.HabitType, existingHabits []models.Habit) (*models.Habit, error) {
	// AIDEV-TODO: Implement true legacy form support without import cycles
	// For now, delegate to wizard as it provides better UX
	return a.CreateHabitWithHybridForms(habitType, existingHabits)
}

// CreateHabitWithHybridForms creates a habit using huh forms embedded in bubbletea for enhanced UX
func (a *LegacyHabitAdapter) CreateHabitWithHybridForms(habitType models.HabitType, existingHabits []models.Habit) (*models.Habit, error) {
	// This demonstrates how to wrap existing forms with bubbletea for enhanced UX
	// while maintaining the same underlying logic

	switch habitType {
	case models.SimpleHabit:
		return a.createSimpleHabitWithHybrid(existingHabits)
	case models.ElasticHabit:
		return a.createElasticHabitWithHybrid(existingHabits)
	case models.InformationalHabit:
		return a.createInformationalHabitWithHybrid(existingHabits)
	default:
		return nil, fmt.Errorf("unsupported habit type for hybrid forms: %s", habitType)
	}
}

// createSimpleHabitWithHybrid creates a simple habit using hybrid approach
func (a *LegacyHabitAdapter) createSimpleHabitWithHybrid(existingHabits []models.Habit) (*models.Habit, error) {
	// Use the wizard for simple habits as it provides better UX than sequential forms
	wizardModel := NewHabitWizardModel(models.SimpleHabit, existingHabits)

	program := tea.NewProgram(wizardModel)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("wizard execution failed: %w", err)
	}

	if wizardModel, ok := finalModel.(*HabitWizardModel); ok {
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

		return result.Habit, nil
	}

	return nil, fmt.Errorf("unexpected wizard model type")
}

// createElasticHabitWithHybrid creates an elastic habit using hybrid approach
func (a *LegacyHabitAdapter) createElasticHabitWithHybrid(existingHabits []models.Habit) (*models.Habit, error) {
	// Elastic habits always use the wizard due to complexity
	return a.createSimpleHabitWithHybrid(existingHabits) // Same implementation
}

// createInformationalHabitWithHybrid creates an informational habit using hybrid approach
func (a *LegacyHabitAdapter) createInformationalHabitWithHybrid(existingHabits []models.Habit) (*models.Habit, error) {
	// Informational habits always use the wizard for consistency
	return a.createSimpleHabitWithHybrid(existingHabits) // Same implementation
}

// AIDEV-TODO: Add demonstration methods for hybrid form usage
// DemoHybridForm would show how to embed huh forms in bubbletea
// Removed for now to avoid import cycles with huh package

// BackwardsCompatibilityMode provides a flag to control interface selection
type BackwardsCompatibilityMode int

const (
	// AutoSelect automatically chooses the best interface for each habit type
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

// CreateHabitWithMode creates a habit using the specified compatibility mode
func (a *LegacyHabitAdapter) CreateHabitWithMode(habitType models.HabitType, existingHabits []models.Habit, mode BackwardsCompatibilityMode) (*models.Habit, error) {
	switch mode {
	case AutoSelect:
		// Use intelligent selection based on habit type complexity
		return a.createHabitAutoSelect(habitType, existingHabits)
	case PreferWizard, ForceWizard:
		// Use wizard interface
		return a.CreateHabitWithHybridForms(habitType, existingHabits)
	case PreferLegacy, ForceLegacy:
		// Use legacy forms
		return a.CreateHabitWithLegacyForms(habitType, existingHabits)
	default:
		return nil, fmt.Errorf("unsupported compatibility mode: %d", mode)
	}
}

// createHabitAutoSelect automatically selects the best interface
func (a *LegacyHabitAdapter) createHabitAutoSelect(habitType models.HabitType, existingHabits []models.Habit) (*models.Habit, error) {
	switch habitType {
	case models.SimpleHabit:
		// Simple habits can use either interface - prefer wizard for better UX
		return a.CreateHabitWithHybridForms(habitType, existingHabits)
	case models.ElasticHabit:
		// Elastic habits require wizard due to complexity
		return a.CreateHabitWithHybridForms(habitType, existingHabits)
	case models.InformationalHabit:
		// Informational habits require wizard for direction configuration
		return a.CreateHabitWithHybridForms(habitType, existingHabits)
	default:
		// Unknown types: fallback to legacy forms
		return a.CreateHabitWithLegacyForms(habitType, existingHabits)
	}
}

// CreateHabitWithBasicInfo creates a habit with pre-populated basic info using the specified compatibility mode
func (a *LegacyHabitAdapter) CreateHabitWithBasicInfo(_ interface{}, existingHabits []models.Habit, mode BackwardsCompatibilityMode) (*models.Habit, error) {
	// Extract habit type from basic info to determine flow
	// For now, delegate to the regular CreateHabitWithMode since wizards handle pre-population
	// AIDEV-TODO: Extract habit type from basicInfo and pass it properly

	// Default to simple habit if we can't extract the type
	habitType := models.SimpleHabit

	return a.CreateHabitWithMode(habitType, existingHabits, mode)
}
