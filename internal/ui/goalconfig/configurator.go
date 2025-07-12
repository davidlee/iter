package goalconfig

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
	"davidlee/iter/internal/parser"
	"davidlee/iter/internal/ui/goalconfig/wizard"
)

// GoalConfigurator provides UI for managing goal configurations
type GoalConfigurator struct {
	goalParser     *parser.GoalParser
	goalBuilder    *GoalBuilder
	legacyAdapter  *wizard.LegacyGoalAdapter
	preferLegacy   bool // Configuration option for backwards compatibility
}

// NewGoalConfigurator creates a new goal configurator instance
func NewGoalConfigurator() *GoalConfigurator {
	return &GoalConfigurator{
		goalParser:    parser.NewGoalParser(),
		goalBuilder:   NewGoalBuilder(),
		legacyAdapter: wizard.NewLegacyGoalAdapter(),
		preferLegacy:  false, // Default to enhanced interfaces
	}
}

// WithLegacyMode configures the configurator to prefer legacy forms
func (gc *GoalConfigurator) WithLegacyMode(prefer bool) *GoalConfigurator {
	gc.preferLegacy = prefer
	return gc
}

// AIDEV-NOTE: Simplified goal creation using idiomatic bubbletea patterns (Phase 2.8)
// Based on documentation review of https://github.com/charmbracelet/bubbletea and 
// https://github.com/charmbracelet/huh/blob/main/examples/bubbletea/main.go
// Replaced complex wizard architecture with simple Model-View-Update pattern
// Flow: Basic Info Collection → Simple Goal Creator (bubbletea) → Save to file
// Focus: Manual simple goals (most common use case) with custom prompts
// Future: Extend SimpleGoalCreator for other goal types as needed

// AddGoal presents an interactive UI to create a new goal
func (gc *GoalConfigurator) AddGoal(goalsFilePath string) error {
	// Load existing schema
	schema, err := gc.loadSchema(goalsFilePath)
	if err != nil {
		return fmt.Errorf("failed to load existing goals: %w", err)
	}

	// Display welcome message
	gc.displayAddGoalWelcome()

	// Collect basic information first (title, description, goal type)
	basicInfo, err := gc.collectBasicInformation()
	if err != nil {
		return fmt.Errorf("basic information collection failed: %w", err)
	}

	// Route to appropriate goal creator based on goal type
	var newGoal *models.Goal
	
	switch basicInfo.GoalType {
	case models.InformationalGoal:
		newGoal, err = gc.runInformationalGoalCreator(basicInfo, schema.Goals)
		if err != nil {
			return fmt.Errorf("informational goal creation failed: %w", err)
		}
	case models.SimpleGoal, models.ElasticGoal:
		// Use simplified goal creator for simple and elastic goals
		newGoal, err = gc.runSimpleGoalCreator(basicInfo, schema.Goals)
		if err != nil {
			return fmt.Errorf("goal creation failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported goal type: %s", basicInfo.GoalType)
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




// BasicInfo holds the pre-collected basic information for all goals
type BasicInfo struct {
	Title       string
	Description string
	GoalType    models.GoalType
}

// collectBasicInformation collects title, description, and goal type upfront
func (gc *GoalConfigurator) collectBasicInformation() (*BasicInfo, error) {
	var title, description string
	var goalType models.GoalType

	// Step 1: Collect Title and Description
	basicInfoForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Goal Title").
				Description("Enter a clear, descriptive title for your goal").
				Value(&title).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return fmt.Errorf("goal title is required")
					}
					if len(s) > 100 {
						return fmt.Errorf("goal title must be 100 characters or less")
					}
					return nil
				}),

			huh.NewText().
				Title("Description (optional)").
				Description("Provide additional context about this goal").
				Value(&description),
		),
	)

	if err := basicInfoForm.Run(); err != nil {
		return nil, err
	}

	// Step 2: Goal type selection
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
		return nil, err
	}

	basicInfo := &BasicInfo{
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		GoalType:    goalType,
	}

	return basicInfo, nil
}


// runSimpleGoalCreator runs the simplified goal creator with pre-populated basic info
func (gc *GoalConfigurator) runSimpleGoalCreator(basicInfo *BasicInfo, _ []models.Goal) (*models.Goal, error) {
	// Create simple goal creator with pre-populated basic info
	creator := NewSimpleGoalCreator(basicInfo.Title, basicInfo.Description, basicInfo.GoalType)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("goal creator execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*SimpleGoalCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("goal creation was cancelled")
		}

		goal, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("goal creation error: %w", err)
		}

		if goal == nil {
			return nil, fmt.Errorf("goal creation completed without result")
		}

		// AIDEV-NOTE: Position is inferred and should not be set in goal creation
		// Position will be determined by the parser/schema based on order in goals.yml

		return goal, nil
	}

	return nil, fmt.Errorf("unexpected creator model type")
}

// runInformationalGoalCreator runs the informational goal creator with pre-populated basic info
func (gc *GoalConfigurator) runInformationalGoalCreator(basicInfo *BasicInfo, _ []models.Goal) (*models.Goal, error) {
	// TODO: Implement InformationalGoalCreator for Phase 3.3
	// For now, create a placeholder informational goal
	
	goal := &models.Goal{
		Title:       basicInfo.Title,
		Description: basicInfo.Description,
		GoalType:    basicInfo.GoalType,
		FieldType: models.FieldType{
			Type: models.BooleanFieldType, // Default field type for now
		},
		ScoringType: models.ManualScoring, // Informational goals always use manual scoring
		Direction:   "neutral",            // Default direction
	}
	
	return goal, nil
}

