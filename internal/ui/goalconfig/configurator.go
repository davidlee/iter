package goalconfig

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
	"davidlee/iter/internal/parser"
	"davidlee/iter/internal/ui/goalconfig/wizard"
)

// GoalConfigurator provides UI for managing goal configurations
type GoalConfigurator struct {
	goalParser  *parser.GoalParser
	goalBuilder *GoalBuilder
}

// NewGoalConfigurator creates a new goal configurator instance
func NewGoalConfigurator() *GoalConfigurator {
	return &GoalConfigurator{
		goalParser:  parser.NewGoalParser(),
		goalBuilder: NewGoalBuilder(),
	}
}

// AddGoal presents an interactive UI to create a new goal
func (gc *GoalConfigurator) AddGoal(goalsFilePath string) error {
	// Load existing schema
	schema, err := gc.loadSchema(goalsFilePath)
	if err != nil {
		return fmt.Errorf("failed to load existing goals: %w", err)
	}

	// Display welcome message
	gc.displayAddGoalWelcome()

	// Prompt for goal type first to determine which flow to use
	goalType, useWizard, err := gc.promptForGoalTypeAndMode()
	if err != nil {
		return fmt.Errorf("goal type selection failed: %w", err)
	}

	var newGoal *models.Goal

	if useWizard {
		// Use enhanced bubbletea wizard for complex flows
		newGoal, err = gc.runGoalWizard(goalType, schema.Goals)
		if err != nil {
			return fmt.Errorf("goal creation wizard failed: %w", err)
		}
	} else {
		// Use existing huh forms for simple flows
		newGoal, err = gc.goalBuilder.BuildGoal(schema.Goals)
		if err != nil {
			return fmt.Errorf("goal creation cancelled or failed: %w", err)
		}
	}

	// Validate the new goal
	if err := newGoal.Validate(); err != nil {
		return fmt.Errorf("goal validation failed: %w", err)
	}

	// Add to schema
	schema.Goals = append(schema.Goals, *newGoal)

	// Validate complete schema
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// Save updated schema
	if err := gc.saveSchema(schema, goalsFilePath); err != nil {
		return fmt.Errorf("failed to save goals: %w", err)
	}

	// Display success message
	gc.displayGoalAdded(newGoal)

	return nil
}

func (gc *GoalConfigurator) displayAddGoalWelcome() {
	welcomeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")). // Bright blue
		Margin(1, 0)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")). // Gray
		Margin(0, 0, 1, 0)

	fmt.Println(welcomeStyle.Render("🎯 Add New Goal"))
	fmt.Println(descStyle.Render("Let's create a new goal through guided prompts."))
}

func (gc *GoalConfigurator) displayGoalAdded(goal *models.Goal) {
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10")). // Bright green
		Margin(1, 0)

	goalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // Bright cyan
		Bold(true)

	fmt.Println(successStyle.Render("✅ Goal Added Successfully!"))
	fmt.Printf("Goal: %s\n", goalStyle.Render(goal.Title))
	fmt.Printf("Type: %s\n", goal.GoalType)
	fmt.Printf("Field: %s\n", goal.FieldType.Type)
	if goal.ScoringType != "" {
		fmt.Printf("Scoring: %s\n", goal.ScoringType)
	}
	fmt.Println()
}

// ListGoals displays all existing goals in a formatted view
func (gc *GoalConfigurator) ListGoals(_ string) error {
	// TODO: Phase 3 - Implement goal listing UI
	return nil
}

// EditGoal presents an interactive UI to modify an existing goal
func (gc *GoalConfigurator) EditGoal(_ string) error {
	// TODO: Phase 4 - Implement goal editing UI
	return nil
}

// RemoveGoal presents an interactive UI to remove an existing goal
func (gc *GoalConfigurator) RemoveGoal(_ string) error {
	// TODO: Phase 5 - Implement goal removal UI
	return nil
}

// loadSchema loads and parses the goals schema from file
func (gc *GoalConfigurator) loadSchema(goalsFilePath string) (*models.Schema, error) {
	return gc.goalParser.LoadFromFileWithIDPersistence(goalsFilePath, true)
}

// saveSchema saves the goals schema back to file
func (gc *GoalConfigurator) saveSchema(schema *models.Schema, goalsFilePath string) error {
	return gc.goalParser.SaveToFile(schema, goalsFilePath)
}

// promptForGoalTypeAndMode prompts for goal type and determines whether to use wizard
func (gc *GoalConfigurator) promptForGoalTypeAndMode() (models.GoalType, bool, error) {
	var goalType models.GoalType
	var useEnhanced bool

	// Goal type selection
	goalTypeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[models.GoalType]().
				Title("Goal Type").
				Description("Choose how this goal will be tracked and scored").
				Options(
					huh.NewOption("Simple (Pass/Fail)", models.SimpleGoal),
					huh.NewOption("Elastic (Mini/Midi/Maxi levels)", models.ElasticGoal),
					huh.NewOption("Informational (Data tracking only)", models.InformationalGoal),
				).
				Value(&goalType),
		),
	)

	if err := goalTypeForm.Run(); err != nil {
		return "", false, err
	}

	// Mode selection - offer enhanced wizard for complex goal types
	if goalType == models.ElasticGoal {
		modeForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Use Enhanced Wizard?").
					Description("Enhanced wizard provides progress tracking, navigation, and better error recovery for complex elastic goals.").
					Affirmative("Enhanced Wizard").
					Negative("Simple Forms").
					Value(&useEnhanced),
			),
		)

		if err := modeForm.Run(); err != nil {
			return goalType, false, err
		}
	}

	return goalType, useEnhanced, nil
}

// runGoalWizard runs the bubbletea-based goal creation wizard
func (gc *GoalConfigurator) runGoalWizard(goalType models.GoalType, existingGoals []models.Goal) (*models.Goal, error) {
	// Create and run the wizard
	wizardModel := wizard.NewGoalWizardModel(goalType, existingGoals)
	
	program := tea.NewProgram(wizardModel)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("wizard execution failed: %w", err)
	}

	// Extract result from final model
	if wizardModel, ok := finalModel.(*wizard.GoalWizardModel); ok {
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