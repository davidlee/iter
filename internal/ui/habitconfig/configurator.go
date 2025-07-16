package habitconfig

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
	"davidlee/vice/internal/ui/habitconfig/wizard"
)

// HabitConfigurator provides UI for managing habit configurations
type HabitConfigurator struct {
	goalParser         *parser.HabitParser
	goalBuilder        *HabitBuilder
	legacyAdapter      *wizard.LegacyHabitAdapter
	preferLegacy       bool   // Configuration option for backwards compatibility
	checklistsFilePath string // Path to checklists.yml for checklist habit creation
}

// NewHabitConfigurator creates a new habit configurator instance
func NewHabitConfigurator() *HabitConfigurator {
	return &HabitConfigurator{
		goalParser:    parser.NewHabitParser(),
		goalBuilder:   NewHabitBuilder(),
		legacyAdapter: wizard.NewLegacyHabitAdapter(),
		preferLegacy:  false, // Default to enhanced interfaces
	}
}

// WithChecklistsFile sets the path to checklists.yml for checklist habit creation
func (gc *HabitConfigurator) WithChecklistsFile(checklistsFilePath string) *HabitConfigurator {
	gc.checklistsFilePath = checklistsFilePath
	return gc
}

// WithLegacyMode configures the configurator to prefer legacy forms
func (gc *HabitConfigurator) WithLegacyMode(prefer bool) *HabitConfigurator {
	gc.preferLegacy = prefer
	return gc
}

// AIDEV-NOTE: Simplified habit creation using idiomatic bubbletea patterns (Phase 2.8)
// Based on documentation review of https://github.com/charmbracelet/bubbletea and
// https://github.com/charmbracelet/huh/blob/main/examples/bubbletea/main.go
// Replaced complex wizard architecture with simple Model-View-Update pattern
// Flow: Basic Info Collection → Simple Habit Creator (bubbletea) → Save to file
// Focus: Manual simple habits (most common use case) with custom prompts
// Future: Extend SimpleHabitCreator for other habit types as needed

// AddHabit presents an interactive UI to create a new habit
func (gc *HabitConfigurator) AddHabit(habitsFilePath string) error {
	// Load existing schema
	schema, err := gc.loadSchema(habitsFilePath)
	if err != nil {
		return fmt.Errorf("failed to load existing habits: %w", err)
	}

	// Display welcome message
	gc.displayAddHabitWelcome()

	// Collect basic information first (title, description, habit type)
	basicInfo, err := gc.collectBasicInformation()
	if err != nil {
		return fmt.Errorf("basic information collection failed: %w", err)
	}

	// AIDEV-NOTE: habit-type-routing; add new habit types here with corresponding creator methods
	// Route to appropriate habit creator based on habit type
	var newHabit *models.Habit

	switch basicInfo.HabitType {
	case models.InformationalHabit:
		newHabit, err = gc.runInformationalHabitCreator(basicInfo, schema.Habits)
		if err != nil {
			return fmt.Errorf("informational habit creation failed: %w", err)
		}
	case models.SimpleHabit:
		// Use simple habit creator for simple habits
		newHabit, err = gc.runSimpleHabitCreator(basicInfo, schema.Habits)
		if err != nil {
			return fmt.Errorf("simple habit creation failed: %w", err)
		}
	case models.ElasticHabit:
		// Use elastic habit creator for elastic habits
		newHabit, err = gc.runElasticHabitCreator(basicInfo, schema.Habits)
		if err != nil {
			return fmt.Errorf("elastic habit creation failed: %w", err)
		}
	case models.ChecklistHabit:
		// Use checklist habit creator for checklist habits
		newHabit, err = gc.runChecklistHabitCreator(basicInfo, schema.Habits)
		if err != nil {
			return fmt.Errorf("checklist habit creation failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported habit type: %s", basicInfo.HabitType)
	}

	// Validate the new habit
	if err := newHabit.Validate(); err != nil {
		return fmt.Errorf("habit validation failed: %w", err)
	}

	// Add to schema
	schema.Habits = append(schema.Habits, *newHabit)

	// Validate complete schema
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// Save updated schema
	if err := gc.saveSchema(schema, habitsFilePath); err != nil {
		return fmt.Errorf("failed to save habits: %w", err)
	}

	// Display success message
	gc.displayHabitAdded(newHabit)

	return nil
}

func (gc *HabitConfigurator) displayAddHabitWelcome() {
	welcomeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")). // Bright blue
		Margin(1, 0)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")). // Gray
		Margin(0, 0, 1, 0)

	fmt.Println(welcomeStyle.Render("🎯 Add New Habit"))
	fmt.Println(descStyle.Render("Let's create a new habit through guided prompts."))
}

func (gc *HabitConfigurator) displayHabitAdded(habit *models.Habit) {
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10")). // Bright green
		Margin(1, 0)

	goalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // Bright cyan
		Bold(true)

	fmt.Println(successStyle.Render("✅ Habit Added Successfully!"))
	fmt.Printf("Habit: %s\n", goalStyle.Render(habit.Title))
	fmt.Printf("Type: %s\n", habit.HabitType)
	fmt.Printf("Field: %s\n", habit.FieldType.Type)
	if habit.ScoringType != "" {
		fmt.Printf("Scoring: %s\n", habit.ScoringType)
	}
	fmt.Println()
}

// ListHabits displays all existing habits in an interactive list view
func (gc *HabitConfigurator) ListHabits(habitsFilePath string) error {
	// Load existing habits
	schema, err := gc.goalParser.LoadFromFile(habitsFilePath)
	if err != nil {
		// Handle file not found gracefully
		if strings.Contains(err.Error(), "habits file not found") {
			fmt.Println("No habits file found. Use .vice habit add. to create your first habit.")
			return nil
		}
		return fmt.Errorf("failed to load habits: %w", err)
	}

	// Handle empty habits list
	if len(schema.Habits) == 0 {
		fmt.Println("No habits configured yet. Use .vice habit add. to create your first habit.")
		return nil
	}

	// Create and run the interactive list
	for {
		listModel := NewHabitListModel(schema.Habits)
		program := tea.NewProgram(listModel)

		// Run the program
		finalModel, err := program.Run()
		if err != nil {
			return fmt.Errorf("failed to run habit list interface: %w", err)
		}

		// Check if user selected a habit for editing or deletion
		if listModel, ok := finalModel.(*HabitListModel); ok {
			if editHabitID := listModel.GetSelectedHabitForEdit(); editHabitID != "" {
				// Edit the selected habit
				if err := gc.EditHabitByID(habitsFilePath, editHabitID); err != nil {
					return fmt.Errorf("failed to edit habit: %w", err)
				}
				// Reload habits after editing and continue the loop
				schema, err = gc.goalParser.LoadFromFile(habitsFilePath)
				if err != nil {
					return fmt.Errorf("failed to reload habits after edit: %w", err)
				}
				continue // Show list again with updated habits
			} else if deleteHabitID := listModel.GetSelectedHabitForDelete(); deleteHabitID != "" {
				// Delete the selected habit
				if err := gc.RemoveHabitByID(habitsFilePath, deleteHabitID); err != nil {
					return fmt.Errorf("failed to delete habit: %w", err)
				}
				// Reload habits after deletion and continue the loop
				schema, err = gc.goalParser.LoadFromFile(habitsFilePath)
				if err != nil {
					return fmt.Errorf("failed to reload habits after delete: %w", err)
				}
				// Check if any habits remain
				if len(schema.Habits) == 0 {
					fmt.Println("No habits remaining. Use .vice habit add. to create your first habit.")
					break // Exit the loop
				}
				continue // Show list again with updated habits
			}
		}

		// No edit operation, exit normally
		break
	}

	return nil
}

// EditHabit presents an interactive UI to modify an existing habit.
// AIDEV-NOTE: habit-edit-flow; Phase 3 implementation - delegates to interactive list UI
// Public API maintains backward compatibility while ListHabits() handles selection+editing
// AIDEV-NOTE: habit-edit-integration; uses interactive list for habit selection and editing
func (gc *HabitConfigurator) EditHabit(habitsFilePath string) error {
	// Delegate to ListHabits which now handles edit operations
	return gc.ListHabits(habitsFilePath)
}

// EditHabitByID modifies a specific habit by ID (used internally by habit list UI).
// AIDEV-NOTE: position-preservation-architecture; maintains habit.Position and habit.ID during edits
// Critical for future reordering feature - habits stay in same list position after editing
func (gc *HabitConfigurator) EditHabitByID(habitsFilePath string, goalID string) error {
	// Load existing schema
	schema, err := gc.loadSchema(habitsFilePath)
	if err != nil {
		return fmt.Errorf("failed to load existing habits: %w", err)
	}

	// Find the habit to edit
	var goalToEdit *models.Habit
	var goalIndex int
	for i, habit := range schema.Habits {
		if habit.ID == goalID {
			goalToEdit = &habit
			goalIndex = i
			break
		}
	}

	if goalToEdit == nil {
		return fmt.Errorf("habit with ID %s not found", goalID)
	}

	// Display edit welcome message
	gc.displayEditHabitWelcome(goalToEdit)

	// AIDEV-NOTE: habit-edit-routing; preserve position and ID during edit operations
	// Route to appropriate habit creator based on habit type
	var editedHabit *models.Habit

	switch goalToEdit.HabitType {
	case models.InformationalHabit:
		editedHabit, err = gc.runInformationalHabitEditor(goalToEdit)
		if err != nil {
			return fmt.Errorf("informational habit editing failed: %w", err)
		}
	case models.SimpleHabit:
		editedHabit, err = gc.runSimpleHabitEditor(goalToEdit)
		if err != nil {
			return fmt.Errorf("simple habit editing failed: %w", err)
		}
	case models.ElasticHabit:
		editedHabit, err = gc.runElasticHabitEditor(goalToEdit)
		if err != nil {
			return fmt.Errorf("elastic habit editing failed: %w", err)
		}
	case models.ChecklistHabit:
		editedHabit, err = gc.runChecklistHabitEditor(goalToEdit)
		if err != nil {
			return fmt.Errorf("checklist habit editing failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported habit type for editing: %s", goalToEdit.HabitType)
	}

	// Preserve original ID and position
	editedHabit.ID = goalToEdit.ID
	editedHabit.Position = goalToEdit.Position

	// Validate the edited habit
	if err := editedHabit.Validate(); err != nil {
		return fmt.Errorf("habit validation failed: %w", err)
	}

	// Replace habit in schema at same position
	schema.Habits[goalIndex] = *editedHabit

	// Validate complete schema
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// Save updated schema
	if err := gc.saveSchema(schema, habitsFilePath); err != nil {
		return fmt.Errorf("failed to save habits: %w", err)
	}

	// Display success message
	gc.displayHabitEdited(editedHabit)

	return nil
}

// RemoveHabit presents an interactive UI to remove an existing habit.
// AIDEV-NOTE: habit-remove-flow; Phase 3 implementation - delegates to interactive list UI
// Public API maintains backward compatibility while ListHabits() handles selection+deletion
// AIDEV-NOTE: habit-remove-integration; uses interactive list for habit selection and deletion
func (gc *HabitConfigurator) RemoveHabit(habitsFilePath string) error {
	// Delegate to ListHabits which now handles delete operations
	return gc.ListHabits(habitsFilePath)
}

// RemoveHabitByID removes a specific habit by ID (used internally by habit list UI)
func (gc *HabitConfigurator) RemoveHabitByID(habitsFilePath string, goalID string) error {
	// Load existing schema
	schema, err := gc.loadSchema(habitsFilePath)
	if err != nil {
		return fmt.Errorf("failed to load existing habits: %w", err)
	}

	// Find the habit to delete
	var goalToDelete *models.Habit
	var goalIndex int
	for i, habit := range schema.Habits {
		if habit.ID == goalID {
			goalToDelete = &habit
			goalIndex = i
			break
		}
	}

	if goalToDelete == nil {
		return fmt.Errorf("habit with ID %s not found", goalID)
	}

	// Show confirmation dialog with backup option
	confirmed, createBackup, err := gc.confirmHabitDeletion(goalToDelete)
	if err != nil {
		return fmt.Errorf("failed to get deletion confirmation: %w", err)
	}

	if !confirmed {
		fmt.Println("Habit deletion cancelled.")
		return nil
	}

	// Create backup if requested
	if createBackup {
		if err := gc.createHabitsBackup(habitsFilePath); err != nil {
			// Warn but don't fail the deletion
			fmt.Printf("Warning: failed to create backup: %v\n", err)
			fmt.Println("Continuing with deletion...")
		}
	}

	// Remove habit from schema
	schema.Habits = append(schema.Habits[:goalIndex], schema.Habits[goalIndex+1:]...)

	// Validate complete schema
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("schema validation failed after removal: %w", err)
	}

	// Save updated schema
	if err := gc.saveSchema(schema, habitsFilePath); err != nil {
		return fmt.Errorf("failed to save habits after removal: %w", err)
	}

	// Display success message
	gc.displayHabitDeleted(goalToDelete)

	return nil
}

// loadSchema loads and parses the habits schema from file
func (gc *HabitConfigurator) loadSchema(habitsFilePath string) (*models.Schema, error) {
	return gc.goalParser.LoadFromFileWithIDPersistence(habitsFilePath, true)
}

// saveSchema saves the habits schema back to file
func (gc *HabitConfigurator) saveSchema(schema *models.Schema, habitsFilePath string) error {
	return gc.goalParser.SaveToFile(schema, habitsFilePath)
}

// BasicInfo holds the pre-collected basic information for all habits
type BasicInfo struct {
	Title       string
	Description string
	HabitType   models.HabitType
}

// collectBasicInformation collects title, description, and habit type upfront
func (gc *HabitConfigurator) collectBasicInformation() (*BasicInfo, error) {
	var title, description string
	var goalType models.HabitType

	// Step 1: Collect Title and Description
	basicInfoForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Habit Title").
				Description("Enter a clear, descriptive title for your habit").
				Value(&title).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return fmt.Errorf("habit title is required")
					}
					if len(s) > 100 {
						return fmt.Errorf("habit title must be 100 characters or less")
					}
					return nil
				}),

			huh.NewText().
				Title("Description (optional)").
				Description("Provide additional context about this habit").
				Value(&description),
		),
	)

	if err := basicInfoForm.Run(); err != nil {
		return nil, err
	}

	// Step 2: Habit type selection
	goalTypeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[models.HabitType]().
				Title("Habit Type").
				Description("Choose how this habit will be tracked and scored").
				Options(
					huh.NewOption("Simple (Pass/Fail)", models.SimpleHabit),
					huh.NewOption("Elastic (Mini/Midi/Maxi levels)", models.ElasticHabit),
					huh.NewOption("Informational (Data tracking only)", models.InformationalHabit),
					huh.NewOption("Checklist (Complete checklist items)", models.ChecklistHabit),
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
		HabitType:   goalType,
	}

	return basicInfo, nil
}

// runSimpleHabitCreator runs the simplified habit creator with pre-populated basic info
func (gc *HabitConfigurator) runSimpleHabitCreator(basicInfo *BasicInfo, _ []models.Habit) (*models.Habit, error) {
	// Create simple habit creator with pre-populated basic info
	creator := NewSimpleHabitCreator(basicInfo.Title, basicInfo.Description, basicInfo.HabitType)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("habit creator execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*SimpleHabitCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("habit creation was cancelled")
		}

		habit, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("habit creation error: %w", err)
		}

		if habit == nil {
			return nil, fmt.Errorf("habit creation completed without result")
		}

		// AIDEV-NOTE: Position is inferred and should not be set in habit creation
		// Position will be determined by the parser/schema based on order in habits.yml

		return habit, nil
	}

	return nil, fmt.Errorf("unexpected creator model type")
}

// AIDEV-NOTE: elastic-habit-creator-integration; follows same pattern as runSimpleHabitCreator for consistency
// runElasticHabitCreator runs the elastic habit creator with pre-populated basic info
func (gc *HabitConfigurator) runElasticHabitCreator(basicInfo *BasicInfo, _ []models.Habit) (*models.Habit, error) {
	// Create elastic habit creator with pre-populated basic info
	creator := NewElasticHabitCreator(basicInfo.Title, basicInfo.Description, basicInfo.HabitType)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("elastic habit creator execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*ElasticHabitCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("elastic habit creation was cancelled")
		}

		habit, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("elastic habit creation error: %w", err)
		}

		if habit == nil {
			return nil, fmt.Errorf("elastic habit creation completed without result")
		}

		// AIDEV-NOTE: Position is inferred and should not be set in habit creation
		// Position will be determined by the parser/schema based on order in habits.yml

		return habit, nil
	}

	return nil, fmt.Errorf("unexpected elastic creator model type")
}

// runInformationalHabitCreator runs the informational habit creator with pre-populated basic info
func (gc *HabitConfigurator) runInformationalHabitCreator(basicInfo *BasicInfo, _ []models.Habit) (*models.Habit, error) {
	// Create informational habit creator with pre-populated basic info
	creator := NewInformationalHabitCreator(basicInfo.Title, basicInfo.Description, basicInfo.HabitType)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("informational habit creator execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*InformationalHabitCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("informational habit creation was cancelled")
		}

		habit, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("informational habit creation error: %w", err)
		}

		if habit == nil {
			return nil, fmt.Errorf("informational habit creation completed without result")
		}

		// AIDEV-NOTE: Position is inferred and should not be set in habit creation
		// Position will be determined by the parser/schema based on order in habits.yml

		return habit, nil
	}

	return nil, fmt.Errorf("unexpected informational creator model type")
}

// AddHabitWithYAMLOutput creates a new habit and returns the resulting YAML without saving to file.
// This is used for dry-run operations where the user wants to preview the generated YAML.
func (gc *HabitConfigurator) AddHabitWithYAMLOutput(habitsFilePath string) (string, error) {
	// Load existing schema
	schema, err := gc.loadSchema(habitsFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to load existing habits: %w", err)
	}

	// Display welcome message
	gc.displayAddHabitWelcome()

	// Collect basic information first (title, description, habit type)
	basicInfo, err := gc.collectBasicInformation()
	if err != nil {
		return "", fmt.Errorf("basic information collection failed: %w", err)
	}

	// AIDEV-NOTE: habit-type-routing; add new habit types here with corresponding creator methods
	// Route to appropriate habit creator based on habit type
	var newHabit *models.Habit

	switch basicInfo.HabitType {
	case models.InformationalHabit:
		newHabit, err = gc.runInformationalHabitCreator(basicInfo, schema.Habits)
		if err != nil {
			return "", fmt.Errorf("informational habit creation failed: %w", err)
		}
	case models.SimpleHabit:
		// Use simple habit creator for simple habits
		newHabit, err = gc.runSimpleHabitCreator(basicInfo, schema.Habits)
		if err != nil {
			return "", fmt.Errorf("simple habit creation failed: %w", err)
		}
	case models.ElasticHabit:
		// Use elastic habit creator for elastic habits
		newHabit, err = gc.runElasticHabitCreator(basicInfo, schema.Habits)
		if err != nil {
			return "", fmt.Errorf("elastic habit creation failed: %w", err)
		}
	case models.ChecklistHabit:
		// Use checklist habit creator for checklist habits
		newHabit, err = gc.runChecklistHabitCreator(basicInfo, schema.Habits)
		if err != nil {
			return "", fmt.Errorf("checklist habit creation failed: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported habit type: %s", basicInfo.HabitType)
	}

	// Validate the new habit
	if err := newHabit.Validate(); err != nil {
		return "", fmt.Errorf("habit validation failed: %w", err)
	}

	// Add to schema (in memory only)
	schema.Habits = append(schema.Habits, *newHabit)

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
	gc.displayHabitAddedDryRun(newHabit)

	return yamlOutput, nil
}

// EditHabitWithYAMLOutput edits a habit and returns the resulting YAML without saving to file.
// This is a placeholder implementation for T006 habit management features.
func (gc *HabitConfigurator) EditHabitWithYAMLOutput(_ string) (string, error) {
	return "", fmt.Errorf("habit editing not yet implemented - see T006 for habit management features")
}

// displayHabitAddedDryRun displays success message for dry-run mode (to stderr)
func (gc *HabitConfigurator) displayHabitAddedDryRun(habit *models.Habit) {
	// Note: Using fmt.Fprintf to stderr to not interfere with YAML output to stdout
	fmt.Fprintf(os.Stderr, "✅ Habit created successfully (dry-run mode): %s\n", habit.Title)
	fmt.Fprintf(os.Stderr, "Type: %s\n", habit.HabitType)
	fmt.Fprintf(os.Stderr, "Field: %s\n", habit.FieldType.Type)
	if habit.ScoringType != "" {
		fmt.Fprintf(os.Stderr, "Scoring: %s\n", habit.ScoringType)
	}
	fmt.Fprintf(os.Stderr, "\n")
}

// runChecklistHabitCreator runs the checklist habit creator with pre-populated basic info
func (gc *HabitConfigurator) runChecklistHabitCreator(basicInfo *BasicInfo, _ []models.Habit) (*models.Habit, error) {
	// Validate that checklists file path is configured
	if gc.checklistsFilePath == "" {
		return nil, fmt.Errorf("checklists file path not configured - use WithChecklistsFile()")
	}

	// Create checklist habit creator with pre-populated basic info
	creator := NewChecklistHabitCreator(basicInfo.Title, basicInfo.Description, basicInfo.HabitType, gc.checklistsFilePath)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("checklist habit creator execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*ChecklistHabitCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("checklist habit creation was cancelled")
		}

		habit, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("checklist habit creation error: %w", err)
		}

		if habit == nil {
			return nil, fmt.Errorf("checklist habit creation completed without result")
		}

		// AIDEV-NOTE: Position is inferred and should not be set in habit creation
		// Position will be determined by the parser/schema based on order in habits.yml

		return habit, nil
	}

	return nil, fmt.Errorf("unexpected checklist creator model type")
}

// displayEditHabitWelcome displays welcome message for habit editing
func (gc *HabitConfigurator) displayEditHabitWelcome(habit *models.Habit) {
	welcomeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("11")). // Bright yellow for edit
		Margin(1, 0)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")). // Gray
		Margin(0, 0, 1, 0)

	fmt.Println(welcomeStyle.Render("✏️ Edit Habit"))
	fmt.Printf("Editing: %s\n", habit.Title)
	fmt.Println(descStyle.Render("Update habit configuration through guided prompts."))
}

// displayHabitEdited displays success message for habit editing
func (gc *HabitConfigurator) displayHabitEdited(habit *models.Habit) {
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10")). // Bright green
		Margin(1, 0)

	goalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // Bright cyan
		Bold(true)

	fmt.Println(successStyle.Render("✅ Habit Updated Successfully!"))
	fmt.Printf("Habit: %s\n", goalStyle.Render(habit.Title))
	fmt.Printf("Type: %s\n", habit.HabitType)
	fmt.Printf("Field: %s\n", habit.FieldType.Type)
	if habit.ScoringType != "" {
		fmt.Printf("Scoring: %s\n", habit.ScoringType)
	}
	fmt.Println()
}

// runSimpleHabitEditor runs the simple habit editor with pre-populated data
func (gc *HabitConfigurator) runSimpleHabitEditor(habit *models.Habit) (*models.Habit, error) {
	// Create simple habit creator with pre-populated data
	creator := NewSimpleHabitCreatorForEdit(habit)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("habit editor execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*SimpleHabitCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("habit editing was cancelled")
		}

		editedHabit, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("habit editing error: %w", err)
		}

		if editedHabit == nil {
			return nil, fmt.Errorf("habit editing completed without result")
		}

		return editedHabit, nil
	}

	return nil, fmt.Errorf("unexpected editor model type")
}

// runElasticHabitEditor runs the elastic habit editor with pre-populated data
func (gc *HabitConfigurator) runElasticHabitEditor(habit *models.Habit) (*models.Habit, error) {
	// Create elastic habit creator with pre-populated data
	creator := NewElasticHabitCreatorForEdit(habit)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("elastic habit editor execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*ElasticHabitCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("elastic habit editing was cancelled")
		}

		editedHabit, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("elastic habit editing error: %w", err)
		}

		if editedHabit == nil {
			return nil, fmt.Errorf("elastic habit editing completed without result")
		}

		return editedHabit, nil
	}

	return nil, fmt.Errorf("unexpected elastic editor model type")
}

// runInformationalHabitEditor runs the informational habit editor with pre-populated data
func (gc *HabitConfigurator) runInformationalHabitEditor(habit *models.Habit) (*models.Habit, error) {
	// Create informational habit creator with pre-populated data
	creator := NewInformationalHabitCreatorForEdit(habit)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("informational habit editor execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*InformationalHabitCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("informational habit editing was cancelled")
		}

		editedHabit, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("informational habit editing error: %w", err)
		}

		if editedHabit == nil {
			return nil, fmt.Errorf("informational habit editing completed without result")
		}

		return editedHabit, nil
	}

	return nil, fmt.Errorf("unexpected informational editor model type")
}

// runChecklistHabitEditor runs the checklist habit editor with pre-populated data
func (gc *HabitConfigurator) runChecklistHabitEditor(habit *models.Habit) (*models.Habit, error) {
	// Validate that checklists file path is configured
	if gc.checklistsFilePath == "" {
		return nil, fmt.Errorf("checklists file path not configured - use WithChecklistsFile()")
	}

	// Create checklist habit creator with pre-populated data
	creator := NewChecklistHabitCreatorForEdit(habit, gc.checklistsFilePath)

	// Run the bubbletea program
	program := tea.NewProgram(creator)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("checklist habit editor execution failed: %w", err)
	}

	// Extract result from final model
	if creatorModel, ok := finalModel.(*ChecklistHabitCreator); ok {
		if creatorModel.IsCancelled() {
			return nil, fmt.Errorf("checklist habit editing was cancelled")
		}

		editedHabit, err := creatorModel.GetResult()
		if err != nil {
			return nil, fmt.Errorf("checklist habit editing error: %w", err)
		}

		if editedHabit == nil {
			return nil, fmt.Errorf("checklist habit editing completed without result")
		}

		return editedHabit, nil
	}

	return nil, fmt.Errorf("unexpected checklist editor model type")
}

// AIDEV-NOTE: backup-protection-strategy; dual confirmation with overwrite protection prevents data loss
// Default yes for backup creation aligns with user safety expectations
// confirmHabitDeletion shows confirmation dialog for habit deletion
func (gc *HabitConfigurator) confirmHabitDeletion(habit *models.Habit) (confirmed bool, createBackup bool, err error) {
	// Display habit details
	gc.displayDeleteHabitWelcome(habit)

	// Confirmation form with backup option
	var confirmDelete bool
	backupOption := true // Default to yes

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Delete Habit?").
				Description(fmt.Sprintf("Are you sure you want to delete '%s'?", habit.Title)).
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

// createHabitsBackup creates a backup of the habits file
func (gc *HabitConfigurator) createHabitsBackup(habitsFilePath string) error {
	// Check if habits file exists
	if _, err := os.Stat(habitsFilePath); os.IsNotExist(err) {
		return fmt.Errorf("habits file not found: %s", habitsFilePath)
	}

	// Create backup filename
	backupPath := habitsFilePath + ".backup"

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
	//nolint:gosec // file path validated via file existence check
	data, err := os.ReadFile(habitsFilePath)
	if err != nil {
		return fmt.Errorf("failed to read habits file for backup: %w", err)
	}

	// Write backup file
	if err := os.WriteFile(backupPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	fmt.Printf("✅ Backup created: %s\n", backupPath)
	return nil
}

// displayDeleteHabitWelcome displays information about the habit to be deleted
func (gc *HabitConfigurator) displayDeleteHabitWelcome(habit *models.Habit) {
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

	fmt.Println(welcomeStyle.Render("🗑️ Delete Habit"))
	fmt.Printf("Habit: %s\n", goalStyle.Render(habit.Title))
	if habit.Description != "" {
		fmt.Printf("Description: %s\n", habit.Description)
	}
	fmt.Printf("Type: %s\n", habit.HabitType)
	fmt.Println(descStyle.Render("This action cannot be undone without a backup."))
}

// displayHabitDeleted displays success message for habit deletion
func (gc *HabitConfigurator) displayHabitDeleted(habit *models.Habit) {
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10")). // Bright green
		Margin(1, 0)

	goalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // Bright cyan
		Bold(true)

	fmt.Println(successStyle.Render("✅ Habit Deleted Successfully!"))
	fmt.Printf("Deleted: %s\n", goalStyle.Render(habit.Title))
	fmt.Println()
}
