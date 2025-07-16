package habitconfig

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
)

// AIDEV-NOTE: Checklist habit creation following simple idiomatic bubbletea pattern
// Extends existing creator patterns for checklist-specific configuration
// Handles checklist selection and automatic/manual scoring setup

// ChecklistHabitCreator implements a bubbletea model for creating checklist habits
type ChecklistHabitCreator struct {
	form     *huh.Form
	quitting bool
	err      error
	result   *models.Habit

	// Form data - bound directly to form fields per huh documentation
	title       string
	description string
	habitType   models.HabitType
	checklistID string
	scoringType models.ScoringType
	prompt      string

	// Available checklists loaded from checklists.yml
	availableChecklists []models.Checklist
	checklistParser     *parser.ChecklistParser
}

// NewChecklistHabitCreatorForEdit creates a checklist habit creator pre-populated with existing habit data for editing
func NewChecklistHabitCreatorForEdit(habit *models.Habit, checklistsFilePath string) *ChecklistHabitCreator {
	creator := &ChecklistHabitCreator{
		title:           habit.Title,
		description:     habit.Description,
		habitType:       habit.HabitType,
		checklistID:     habit.FieldType.ChecklistID,
		scoringType:     habit.ScoringType,
		prompt:          habit.Prompt,
		checklistParser: parser.NewChecklistParser(),
	}

	// Load available checklists for selection
	if err := creator.loadAvailableChecklists(checklistsFilePath); err != nil {
		creator.err = fmt.Errorf("failed to load checklists: %w", err)
		return creator
	}

	// Create the form
	creator.createForm()

	return creator
}

// NewChecklistHabitCreator creates a new checklist habit creator with pre-populated basic info
func NewChecklistHabitCreator(title, description string, habitType models.HabitType, checklistsFilePath string) *ChecklistHabitCreator {
	creator := &ChecklistHabitCreator{
		title:           title,
		description:     description,
		habitType:       habitType,
		prompt:          "Complete your checklist items today", // Default prompt
		checklistParser: parser.NewChecklistParser(),
	}

	// Load available checklists for selection
	if err := creator.loadAvailableChecklists(checklistsFilePath); err != nil {
		creator.err = fmt.Errorf("failed to load checklists: %w", err)
		return creator
	}

	// Create the form
	creator.createForm()

	return creator
}

// loadAvailableChecklists loads checklists from checklists.yml for selection
func (cgc *ChecklistHabitCreator) loadAvailableChecklists(checklistsFilePath string) error {
	schema, err := cgc.checklistParser.LoadFromFile(checklistsFilePath)
	if err != nil {
		// If no checklists file exists, return empty list (not an error)
		// Other parsing errors should be returned
		cgc.availableChecklists = []models.Checklist{}
		if strings.Contains(err.Error(), "checklists file not found") {
			return nil // File not existing is not an error
		}
		return err // Other errors should be propagated
	}

	cgc.availableChecklists = schema.Checklists
	return nil
}

// createForm creates the multi-step form for checklist habit configuration
func (cgc *ChecklistHabitCreator) createForm() {
	// Check if we have any checklists available
	if len(cgc.availableChecklists) == 0 {
		cgc.err = fmt.Errorf("no checklists found - create checklists first using 'vice list add'")
		return
	}

	// Create checklist selection options
	checklistOptions := make([]huh.Option[string], len(cgc.availableChecklists))
	for i, checklist := range cgc.availableChecklists {
		title := checklist.Title
		if title == "" {
			title = checklist.ID
		}
		description := checklist.Description
		if description != "" {
			title = fmt.Sprintf("%s - %s", title, description)
		}
		checklistOptions[i] = huh.NewOption(title, checklist.ID)
	}

	// Create sequential form following huh patterns
	cgc.form = huh.NewForm(
		// Step 1: Checklist selection
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("checklist").
				Title("Select Checklist").
				Description("Choose which checklist this habit will track").
				Options(checklistOptions...).
				Value(&cgc.checklistID),
		),

		// Step 2: Scoring type selection
		huh.NewGroup(
			huh.NewSelect[models.ScoringType]().
				Key("scoring").
				Title("Scoring Type").
				Description("How should checklist completion be scored?").
				Options(
					huh.NewOption("Automatic (all items complete)", models.AutomaticScoring),
					huh.NewOption("Manual (partial completion allowed)", models.ManualScoring),
				).
				Value(&cgc.scoringType),
		),

		// Step 3: Custom prompt (optional)
		huh.NewGroup(
			huh.NewInput().
				Key("prompt").
				Title("Entry Prompt (optional)").
				Description("Customize the prompt shown during daily entry").
				Value(&cgc.prompt).
				Placeholder("Complete your checklist items today"),
		),
	)
}

// Init initializes the checklist habit creator model
func (cgc *ChecklistHabitCreator) Init() tea.Cmd {
	if cgc.err != nil {
		return tea.Quit
	}
	return cgc.form.Init()
}

// Update handles messages and updates the model state
func (cgc *ChecklistHabitCreator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle early exit if there was an initialization error
	if cgc.err != nil {
		cgc.quitting = true
		return cgc, tea.Quit
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c", "esc":
			cgc.quitting = true
			return cgc, tea.Quit
		}
	}

	// Let the form handle the message
	form, cmd := cgc.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		cgc.form = f
	}

	// Check if form is complete
	if cgc.form.State == huh.StateCompleted {
		// Build the habit from collected data
		if err := cgc.buildResult(); err != nil {
			cgc.err = err
		}
		cgc.quitting = true
		return cgc, tea.Quit
	}

	return cgc, cmd
}

// View renders the current view
func (cgc *ChecklistHabitCreator) View() string {
	if cgc.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress any key to exit.", cgc.err)
	}

	if cgc.quitting {
		if cgc.result != nil {
			return "✅ Checklist habit configured successfully!\n"
		}
		return "❌ Checklist habit creation cancelled.\n"
	}

	return cgc.form.View()
}

// buildResult constructs the habit from the collected form data
func (cgc *ChecklistHabitCreator) buildResult() error {
	// Validate required fields
	if cgc.checklistID == "" {
		return fmt.Errorf("checklist selection is required")
	}

	// Clean up prompt
	prompt := strings.TrimSpace(cgc.prompt)
	if prompt == "" {
		prompt = "Complete your checklist items today"
	}

	// Build the habit
	habit := &models.Habit{
		Title:       cgc.title,
		Description: cgc.description,
		HabitType:   cgc.habitType,
		FieldType: models.FieldType{
			Type:        models.ChecklistFieldType,
			ChecklistID: cgc.checklistID,
		},
		ScoringType: cgc.scoringType,
		Prompt:      prompt,
	}

	// Add automatic scoring criteria if selected
	if cgc.scoringType == models.AutomaticScoring {
		habit.Criteria = &models.Criteria{
			Description: "All checklist items completed",
			Condition: &models.Condition{
				ChecklistCompletion: &models.ChecklistCompletionCondition{
					RequiredItems: "all",
				},
			},
		}
	}

	cgc.result = habit
	return nil
}

// GetResult returns the created habit
func (cgc *ChecklistHabitCreator) GetResult() (*models.Habit, error) {
	if cgc.err != nil {
		return nil, cgc.err
	}
	return cgc.result, nil
}

// IsCancelled returns true if habit creation was cancelled
func (cgc *ChecklistHabitCreator) IsCancelled() bool {
	return cgc.quitting && cgc.result == nil && cgc.err == nil
}
