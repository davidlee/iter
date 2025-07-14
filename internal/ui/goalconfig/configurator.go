package goalconfig

import (
	"fmt"
	"os"
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
	goalParser         *parser.GoalParser
	goalBuilder        *GoalBuilder
	legacyAdapter      *wizard.LegacyGoalAdapter
	preferLegacy       bool   // Configuration option for backwards compatibility
	checklistsFilePath string // Path to checklists.yml for checklist goal creation
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

// WithChecklistsFile sets the path to checklists.yml for checklist goal creation
func (gc *GoalConfigurator) WithChecklistsFile(checklistsFilePath string) *GoalConfigurator {
	gc.checklistsFilePath = checklistsFilePath
	return gc
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

	// AIDEV-NOTE: goal-type-routing; add new goal types here with corresponding creator methods
	// Route to appropriate goal creator based on goal type
	var newGoal *models.Goal

	switch basicInfo.GoalType {
	case models.InformationalGoal:
		newGoal, err = gc.runInformationalGoalCreator(basicInfo, schema.Goals)
		if err != nil {
			return fmt.Errorf("informational goal creation failed: %w", err)
		}
	case models.SimpleGoal:
		// Use simple goal creator for simple goals
		newGoal, err = gc.runSimpleGoalCreator(basicInfo, schema.Goals)
		if err != nil {
			return fmt.Errorf("simple goal creation failed: %w", err)
		}
	case models.ElasticGoal:
		// Use elastic goal creator for elastic goals
		newGoal, err = gc.runElasticGoalCreator(basicInfo, schema.Goals)
		if err != nil {
			return fmt.Errorf("elastic goal creation failed: %w", err)
		}
	case models.ChecklistGoal:
		// Use checklist goal creator for checklist goals
		newGoal, err = gc.runChecklistGoalCreator(basicInfo, schema.Goals)
		if err != nil {
			return fmt.Errorf("checklist goal creation failed: %w", err)
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

// ListGoals displays all existing goals in an interactive list view
func (gc *GoalConfigurator) ListGoals(goalsFilePath string) error {
	// Load existing goals
	schema, err := gc.goalParser.LoadFromFile(goalsFilePath)
	if err != nil {
		// Handle file not found gracefully
		if strings.Contains(err.Error(), "goals file not found") {
			fmt.Println("No goals file found. Use 'iter goal add' to create your first goal.")
			return nil
		}
		return fmt.Errorf("failed to load goals: %w", err)
	}

	// Handle empty goals list
	if len(schema.Goals) == 0 {
		fmt.Println("No goals configured yet. Use 'iter goal add' to create your first goal.")
		return nil
	}

	// Create and run the interactive list
	for {
		listModel := NewGoalListModel(schema.Goals)
		program := tea.NewProgram(listModel)

		// Run the program
		finalModel, err := program.Run()
		if err != nil {
			return fmt.Errorf("failed to run goal list interface: %w", err)
		}

		// Check if user selected a goal for editing or deletion
		if listModel, ok := finalModel.(*GoalListModel); ok {
			if editGoalID := listModel.GetSelectedGoalForEdit(); editGoalID != "" {
				// Edit the selected goal
				if err := gc.EditGoalByID(goalsFilePath, editGoalID); err != nil {
					return fmt.Errorf("failed to edit goal: %w", err)
				}
				// Reload goals after editing and continue the loop
				schema, err = gc.goalParser.LoadFromFile(goalsFilePath)
				if err != nil {
					return fmt.Errorf("failed to reload goals after edit: %w", err)
				}
				continue // Show list again with updated goals
			} else if deleteGoalID := listModel.GetSelectedGoalForDelete(); deleteGoalID != "" {
				// Delete the selected goal
				if err := gc.RemoveGoalByID(goalsFilePath, deleteGoalID); err != nil {
					return fmt.Errorf("failed to delete goal: %w", err)
				}
				// Reload goals after deletion and continue the loop
				schema, err = gc.goalParser.LoadFromFile(goalsFilePath)
				if err != nil {
					return fmt.Errorf("failed to reload goals after delete: %w", err)
				}
				// Check if any goals remain
				if len(schema.Goals) == 0 {
					fmt.Println("No goals remaining. Use 'iter goal add' to create your first goal.")
					break // Exit the loop
				}
				continue // Show list again with updated goals
			}
		}

		// No edit operation, exit normally
		break
	}

	return nil
}

// EditGoal presents an interactive UI to modify an existing goal.
// AIDEV-NOTE: goal-edit-flow; Phase 3 implementation - delegates to interactive list UI
// Public API maintains backward compatibility while ListGoals() handles selection+editing
// AIDEV-NOTE: goal-edit-integration; uses interactive list for goal selection and editing
func (gc *GoalConfigurator) EditGoal(goalsFilePath string) error {
	// Delegate to ListGoals which now handles edit operations
	return gc.ListGoals(goalsFilePath)
}

// EditGoalByID modifies a specific goal by ID (used internally by goal list UI).
// AIDEV-NOTE: position-preservation-architecture; maintains goal.Position and goal.ID during edits
// Critical for future reordering feature - goals stay in same list position after editing
func (gc *GoalConfigurator) EditGoalByID(goalsFilePath string, goalID string) error {
	// Load existing schema
	schema, err := gc.loadSchema(goalsFilePath)
	if err != nil {
		return fmt.Errorf("failed to load existing goals: %w", err)
	}

	// Find the goal to edit
	var goalToEdit *models.Goal
	var goalIndex int
	for i, goal := range schema.Goals {
		if goal.ID == goalID {
			goalToEdit = &goal
			goalIndex = i
			break
		}
	}

	if goalToEdit == nil {
		return fmt.Errorf("goal with ID %s not found", goalID)
	}

	// Display edit welcome message
	gc.displayEditGoalWelcome(goalToEdit)

	// AIDEV-NOTE: goal-edit-routing; preserve position and ID during edit operations
	// Route to appropriate goal creator based on goal type
	var editedGoal *models.Goal

	switch goalToEdit.GoalType {
	case models.InformationalGoal:
		editedGoal, err = gc.runInformationalGoalEditor(goalToEdit)
		if err != nil {
			return fmt.Errorf("informational goal editing failed: %w", err)
		}
	case models.SimpleGoal:
		editedGoal, err = gc.runSimpleGoalEditor(goalToEdit)
		if err != nil {
			return fmt.Errorf("simple goal editing failed: %w", err)
		}
	case models.ElasticGoal:
		editedGoal, err = gc.runElasticGoalEditor(goalToEdit)
		if err != nil {
			return fmt.Errorf("elastic goal editing failed: %w", err)
		}
	case models.ChecklistGoal:
		editedGoal, err = gc.runChecklistGoalEditor(goalToEdit)
		if err != nil {
			return fmt.Errorf("checklist goal editing failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported goal type for editing: %s", goalToEdit.GoalType)
	}

	// Preserve original ID and position
	editedGoal.ID = goalToEdit.ID
	editedGoal.Position = goalToEdit.Position

	// Validate the edited goal
	if err := editedGoal.Validate(); err != nil {
		return fmt.Errorf("goal validation failed: %w", err)
	}

	// Replace goal in schema at same position
	schema.Goals[goalIndex] = *editedGoal

	// Validate complete schema
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// Save updated schema
	if err := gc.saveSchema(schema, goalsFilePath); err != nil {
		return fmt.Errorf("failed to save goals: %w", err)
	}

	// Display success message
	gc.displayGoalEdited(editedGoal)

	return nil
}

// RemoveGoal presents an interactive UI to remove an existing goal.
// AIDEV-NOTE: goal-remove-flow; Phase 3 implementation - delegates to interactive list UI
// Public API maintains backward compatibility while ListGoals() handles selection+deletion
// AIDEV-NOTE: goal-remove-integration; uses interactive list for goal selection and deletion
func (gc *GoalConfigurator) RemoveGoal(goalsFilePath string) error {
	// Delegate to ListGoals which now handles delete operations
	return gc.ListGoals(goalsFilePath)
}

// RemoveGoalByID removes a specific goal by ID (used internally by goal list UI)
func (gc *GoalConfigurator) RemoveGoalByID(goalsFilePath string, goalID string) error {
	// Load existing schema
	schema, err := gc.loadSchema(goalsFilePath)
	if err != nil {
		return fmt.Errorf("failed to load existing goals: %w", err)
	}

	// Find the goal to delete
	var goalToDelete *models.Goal
	var goalIndex int
	for i, goal := range schema.Goals {
		if goal.ID == goalID {
			goalToDelete = &goal
			goalIndex = i
			break
		}
	}

	if goalToDelete == nil {
		return fmt.Errorf("goal with ID %s not found", goalID)
	}

	// Show confirmation dialog with backup option
	confirmed, createBackup, err := gc.confirmGoalDeletion(goalToDelete)
	if err != nil {
		return fmt.Errorf("failed to get deletion confirmation: %w", err)
	}

	if !confirmed {
		fmt.Println("Goal deletion cancelled.")
		return nil
	}

	// Create backup if requested
	if createBackup {
		if err := gc.createGoalsBackup(goalsFilePath); err != nil {
			// Warn but don't fail the deletion
			fmt.Printf("Warning: failed to create backup: %v\n", err)
			fmt.Println("Continuing with deletion...")
		}
	}

	// Remove goal from schema
	schema.Goals = append(schema.Goals[:goalIndex], schema.Goals[goalIndex+1:]...)

	// Validate complete schema
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("schema validation failed after removal: %w", err)
	}

	// Save updated schema
	if err := gc.saveSchema(schema, goalsFilePath); err != nil {
		return fmt.Errorf("failed to save goals after removal: %w", err)
	}

	// Display success message
	gc.displayGoalDeleted(goalToDelete)

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
					huh.NewOption("Checklist (Complete checklist items)", models.ChecklistGoal),
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

// AIDEV-NOTE: elastic-goal-creator-integration; follows same pattern as runSimpleGoalCreator for consistency
// runElasticGoalCreator runs the elastic goal creator with pre-populated basic info
func (gc *GoalConfigurator) runElasticGoalCreator(basicInfo *BasicInfo, _ []models.Goal) (*models.Goal, error) {
	// Create elastic goal creator with pre-populated basic info
	creator := NewElasticGoalCreator(basicInfo.Title, basicInfo.Description, basicInfo.GoalType)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("elastic goal creator execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*ElasticGoalCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("elastic goal creation was cancelled")
		}

		goal, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("elastic goal creation error: %w", err)
		}

		if goal == nil {
			return nil, fmt.Errorf("elastic goal creation completed without result")
		}

		// AIDEV-NOTE: Position is inferred and should not be set in goal creation
		// Position will be determined by the parser/schema based on order in goals.yml

		return goal, nil
	}

	return nil, fmt.Errorf("unexpected elastic creator model type")
}

// runInformationalGoalCreator runs the informational goal creator with pre-populated basic info
func (gc *GoalConfigurator) runInformationalGoalCreator(basicInfo *BasicInfo, _ []models.Goal) (*models.Goal, error) {
	// Create informational goal creator with pre-populated basic info
	creator := NewInformationalGoalCreator(basicInfo.Title, basicInfo.Description, basicInfo.GoalType)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("informational goal creator execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*InformationalGoalCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("informational goal creation was cancelled")
		}

		goal, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("informational goal creation error: %w", err)
		}

		if goal == nil {
			return nil, fmt.Errorf("informational goal creation completed without result")
		}

		// AIDEV-NOTE: Position is inferred and should not be set in goal creation
		// Position will be determined by the parser/schema based on order in goals.yml

		return goal, nil
	}

	return nil, fmt.Errorf("unexpected informational creator model type")
}

// AddGoalWithYAMLOutput creates a new goal and returns the resulting YAML without saving to file.
// This is used for dry-run operations where the user wants to preview the generated YAML.
func (gc *GoalConfigurator) AddGoalWithYAMLOutput(goalsFilePath string) (string, error) {
	// Load existing schema
	schema, err := gc.loadSchema(goalsFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to load existing goals: %w", err)
	}

	// Display welcome message
	gc.displayAddGoalWelcome()

	// Collect basic information first (title, description, goal type)
	basicInfo, err := gc.collectBasicInformation()
	if err != nil {
		return "", fmt.Errorf("basic information collection failed: %w", err)
	}

	// AIDEV-NOTE: goal-type-routing; add new goal types here with corresponding creator methods
	// Route to appropriate goal creator based on goal type
	var newGoal *models.Goal

	switch basicInfo.GoalType {
	case models.InformationalGoal:
		newGoal, err = gc.runInformationalGoalCreator(basicInfo, schema.Goals)
		if err != nil {
			return "", fmt.Errorf("informational goal creation failed: %w", err)
		}
	case models.SimpleGoal:
		// Use simple goal creator for simple goals
		newGoal, err = gc.runSimpleGoalCreator(basicInfo, schema.Goals)
		if err != nil {
			return "", fmt.Errorf("simple goal creation failed: %w", err)
		}
	case models.ElasticGoal:
		// Use elastic goal creator for elastic goals
		newGoal, err = gc.runElasticGoalCreator(basicInfo, schema.Goals)
		if err != nil {
			return "", fmt.Errorf("elastic goal creation failed: %w", err)
		}
	case models.ChecklistGoal:
		// Use checklist goal creator for checklist goals
		newGoal, err = gc.runChecklistGoalCreator(basicInfo, schema.Goals)
		if err != nil {
			return "", fmt.Errorf("checklist goal creation failed: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported goal type: %s", basicInfo.GoalType)
	}

	// Validate the new goal
	if err := newGoal.Validate(); err != nil {
		return "", fmt.Errorf("goal validation failed: %w", err)
	}

	// Add to schema (in memory only)
	schema.Goals = append(schema.Goals, *newGoal)

	// Validate complete schema
	if err := schema.Validate(); err != nil {
		return "", fmt.Errorf("schema validation failed: %w", err)
	}

	// Convert to YAML string
	yamlOutput, err := gc.goalParser.ToYAML(schema)
	if err != nil {
		return "", fmt.Errorf("failed to generate YAML output: %w", err)
	}

	// Display success message (to stderr to not interfere with YAML output)
	gc.displayGoalAddedDryRun(newGoal)

	return yamlOutput, nil
}

// EditGoalWithYAMLOutput edits a goal and returns the resulting YAML without saving to file.
// This is a placeholder implementation for T006 goal management features.
func (gc *GoalConfigurator) EditGoalWithYAMLOutput(_ string) (string, error) {
	return "", fmt.Errorf("goal editing not yet implemented - see T006 for goal management features")
}

// displayGoalAddedDryRun displays success message for dry-run mode (to stderr)
func (gc *GoalConfigurator) displayGoalAddedDryRun(goal *models.Goal) {
	// Note: Using fmt.Fprintf to stderr to not interfere with YAML output to stdout
	fmt.Fprintf(os.Stderr, "✅ Goal created successfully (dry-run mode): %s\n", goal.Title)
	fmt.Fprintf(os.Stderr, "Type: %s\n", goal.GoalType)
	fmt.Fprintf(os.Stderr, "Field: %s\n", goal.FieldType.Type)
	if goal.ScoringType != "" {
		fmt.Fprintf(os.Stderr, "Scoring: %s\n", goal.ScoringType)
	}
	fmt.Fprintf(os.Stderr, "\n")
}

// runChecklistGoalCreator runs the checklist goal creator with pre-populated basic info
func (gc *GoalConfigurator) runChecklistGoalCreator(basicInfo *BasicInfo, _ []models.Goal) (*models.Goal, error) {
	// Validate that checklists file path is configured
	if gc.checklistsFilePath == "" {
		return nil, fmt.Errorf("checklists file path not configured - use WithChecklistsFile()")
	}

	// Create checklist goal creator with pre-populated basic info
	creator := NewChecklistGoalCreator(basicInfo.Title, basicInfo.Description, basicInfo.GoalType, gc.checklistsFilePath)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("checklist goal creator execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*ChecklistGoalCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("checklist goal creation was cancelled")
		}

		goal, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("checklist goal creation error: %w", err)
		}

		if goal == nil {
			return nil, fmt.Errorf("checklist goal creation completed without result")
		}

		// AIDEV-NOTE: Position is inferred and should not be set in goal creation
		// Position will be determined by the parser/schema based on order in goals.yml

		return goal, nil
	}

	return nil, fmt.Errorf("unexpected checklist creator model type")
}

// displayEditGoalWelcome displays welcome message for goal editing
func (gc *GoalConfigurator) displayEditGoalWelcome(goal *models.Goal) {
	welcomeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("11")). // Bright yellow for edit
		Margin(1, 0)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")). // Gray
		Margin(0, 0, 1, 0)

	fmt.Println(welcomeStyle.Render("✏️ Edit Goal"))
	fmt.Printf("Editing: %s\n", goal.Title)
	fmt.Println(descStyle.Render("Update goal configuration through guided prompts."))
}

// displayGoalEdited displays success message for goal editing
func (gc *GoalConfigurator) displayGoalEdited(goal *models.Goal) {
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10")). // Bright green
		Margin(1, 0)

	goalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // Bright cyan
		Bold(true)

	fmt.Println(successStyle.Render("✅ Goal Updated Successfully!"))
	fmt.Printf("Goal: %s\n", goalStyle.Render(goal.Title))
	fmt.Printf("Type: %s\n", goal.GoalType)
	fmt.Printf("Field: %s\n", goal.FieldType.Type)
	if goal.ScoringType != "" {
		fmt.Printf("Scoring: %s\n", goal.ScoringType)
	}
	fmt.Println()
}

// runSimpleGoalEditor runs the simple goal editor with pre-populated data
func (gc *GoalConfigurator) runSimpleGoalEditor(goal *models.Goal) (*models.Goal, error) {
	// Create simple goal creator with pre-populated data
	creator := NewSimpleGoalCreatorForEdit(goal)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("goal editor execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*SimpleGoalCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("goal editing was cancelled")
		}

		editedGoal, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("goal editing error: %w", err)
		}

		if editedGoal == nil {
			return nil, fmt.Errorf("goal editing completed without result")
		}

		return editedGoal, nil
	}

	return nil, fmt.Errorf("unexpected editor model type")
}

// runElasticGoalEditor runs the elastic goal editor with pre-populated data
func (gc *GoalConfigurator) runElasticGoalEditor(goal *models.Goal) (*models.Goal, error) {
	// Create elastic goal creator with pre-populated data
	creator := NewElasticGoalCreatorForEdit(goal)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("elastic goal editor execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*ElasticGoalCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("elastic goal editing was cancelled")
		}

		editedGoal, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("elastic goal editing error: %w", err)
		}

		if editedGoal == nil {
			return nil, fmt.Errorf("elastic goal editing completed without result")
		}

		return editedGoal, nil
	}

	return nil, fmt.Errorf("unexpected elastic editor model type")
}

// runInformationalGoalEditor runs the informational goal editor with pre-populated data
func (gc *GoalConfigurator) runInformationalGoalEditor(goal *models.Goal) (*models.Goal, error) {
	// Create informational goal creator with pre-populated data
	creator := NewInformationalGoalCreatorForEdit(goal)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("informational goal editor execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*InformationalGoalCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("informational goal editing was cancelled")
		}

		editedGoal, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("informational goal editing error: %w", err)
		}

		if editedGoal == nil {
			return nil, fmt.Errorf("informational goal editing completed without result")
		}

		return editedGoal, nil
	}

	return nil, fmt.Errorf("unexpected informational editor model type")
}

// runChecklistGoalEditor runs the checklist goal editor with pre-populated data
func (gc *GoalConfigurator) runChecklistGoalEditor(goal *models.Goal) (*models.Goal, error) {
	// Validate that checklists file path is configured
	if gc.checklistsFilePath == "" {
		return nil, fmt.Errorf("checklists file path not configured - use WithChecklistsFile()")
	}

	// Create checklist goal creator with pre-populated data
	creator := NewChecklistGoalCreatorForEdit(goal, gc.checklistsFilePath)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("checklist goal editor execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*ChecklistGoalCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("checklist goal editing was cancelled")
		}

		editedGoal, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("checklist goal editing error: %w", err)
		}

		if editedGoal == nil {
			return nil, fmt.Errorf("checklist goal editing completed without result")
		}

		return editedGoal, nil
	}

	return nil, fmt.Errorf("unexpected checklist editor model type")
}

// AIDEV-NOTE: backup-protection-strategy; dual confirmation with overwrite protection prevents data loss
// Default yes for backup creation aligns with user safety expectations
// confirmGoalDeletion shows confirmation dialog for goal deletion
func (gc *GoalConfigurator) confirmGoalDeletion(goal *models.Goal) (confirmed bool, createBackup bool, err error) {
	// Display goal details
	gc.displayDeleteGoalWelcome(goal)

	// Confirmation form with backup option
	var confirmDelete bool
	backupOption := true // Default to yes

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Delete Goal?").
				Description(fmt.Sprintf("Are you sure you want to delete '%s'?", goal.Title)).
				Value(&confirmDelete),

			huh.NewConfirm().
				Title("Create Backup?").
				Description("Create backup file before deleting? (Recommended)").
				Value(&backupOption),
		),
	)

	if err := form.Run(); err != nil {
		return false, false, fmt.Errorf("confirmation form failed: %w", err)
	}

	return confirmDelete, backupOption, nil
}

// createGoalsBackup creates a backup of the goals file
func (gc *GoalConfigurator) createGoalsBackup(goalsFilePath string) error {
	// Check if goals file exists
	if _, err := os.Stat(goalsFilePath); os.IsNotExist(err) {
		return fmt.Errorf("goals file not found: %s", goalsFilePath)
	}

	// Create backup filename
	backupPath := goalsFilePath + ".backup"

	// Check if backup already exists and ask for confirmation
	if _, err := os.Stat(backupPath); err == nil {
		var overwrite bool
		overwriteForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Backup Exists").
					Description(fmt.Sprintf("Backup file %s already exists. Overwrite?", backupPath)).
					Value(&overwrite),
			),
		)

		if err := overwriteForm.Run(); err != nil {
			return fmt.Errorf("backup overwrite confirmation failed: %w", err)
		}

		if !overwrite {
			return fmt.Errorf("backup creation cancelled - existing backup preserved")
		}
	}

	// Read original file
	data, err := os.ReadFile(goalsFilePath)
	if err != nil {
		return fmt.Errorf("failed to read goals file for backup: %w", err)
	}

	// Write backup file
	if err := os.WriteFile(backupPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	fmt.Printf("✅ Backup created: %s\n", backupPath)
	return nil
}

// displayDeleteGoalWelcome displays information about the goal to be deleted
func (gc *GoalConfigurator) displayDeleteGoalWelcome(goal *models.Goal) {
	welcomeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("9")). // Bright red for delete
		Margin(1, 0)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")). // Gray
		Margin(0, 0, 1, 0)

	goalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")). // White
		Bold(true)

	fmt.Println(welcomeStyle.Render("🗑️ Delete Goal"))
	fmt.Printf("Goal: %s\n", goalStyle.Render(goal.Title))
	if goal.Description != "" {
		fmt.Printf("Description: %s\n", goal.Description)
	}
	fmt.Printf("Type: %s\n", goal.GoalType)
	fmt.Println(descStyle.Render("This action cannot be undone without a backup."))
}

// displayGoalDeleted displays success message for goal deletion
func (gc *GoalConfigurator) displayGoalDeleted(goal *models.Goal) {
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10")). // Bright green
		Margin(1, 0)

	goalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // Bright cyan
		Bold(true)

	fmt.Println(successStyle.Render("✅ Goal Deleted Successfully!"))
	fmt.Printf("Deleted: %s\n", goalStyle.Render(goal.Title))
	fmt.Println()
}
