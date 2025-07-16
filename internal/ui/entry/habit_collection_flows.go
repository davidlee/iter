package entry

import (
	"fmt"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
	"davidlee/vice/internal/scoring"
)

// AIDEV-NOTE: habit-collection-flows; defines specialized collection flows for each habit type with field input integration
// Integrates T010/1.2 field input components with habit type-specific behaviors and scoring patterns
// AIDEV-NOTE: T010/3.1-complete; SimpleHabitCollectionFlow fully implemented with headless testing support
// Features: pass/fail logic, field type support (all except checklist), automatic/manual scoring, notes collection

// HabitCollectionFlow defines the interface for habit type-specific collection flows
type HabitCollectionFlow interface {
	// CollectEntry orchestrates the complete entry collection for a habit type
	CollectEntry(habit models.Habit, existing *ExistingEntry) (*EntryResult, error)

	// GetFlowType returns the habit type this flow handles
	GetFlowType() string

	// RequiresScoring indicates if this flow needs scoring engine integration
	RequiresScoring() bool

	// GetExpectedFieldTypes returns field types supported by this flow
	GetExpectedFieldTypes() []string
}

// SimpleHabitCollectionFlow handles pass/fail collection with optional additional data
type SimpleHabitCollectionFlow struct {
	factory       *EntryFieldInputFactory
	scoringEngine *scoring.Engine
}

// NewSimpleHabitCollectionFlow creates a new simple habit collection flow
func NewSimpleHabitCollectionFlow(factory *EntryFieldInputFactory, scoringEngine *scoring.Engine) *SimpleHabitCollectionFlow {
	return &SimpleHabitCollectionFlow{
		factory:       factory,
		scoringEngine: scoringEngine,
	}
}

// NewSimpleHabitCollectionFlowForTesting creates a flow for testing that bypasses user interaction
func NewSimpleHabitCollectionFlowForTesting(factory *EntryFieldInputFactory, scoringEngine *scoring.Engine) *SimpleHabitCollectionFlow {
	return &SimpleHabitCollectionFlow{
		factory:       factory,
		scoringEngine: scoringEngine,
	}
}

// CollectEntryDirectly bypasses UI interaction and creates entry directly from provided value
func (f *SimpleHabitCollectionFlow) CollectEntryDirectly(habit models.Habit, value interface{}, notes string, _ *ExistingEntry) (*EntryResult, error) {
	// Handle scoring based on habit configuration
	var achievementLevel *models.AchievementLevel
	if habit.ScoringType == models.AutomaticScoring {
		// Automatic scoring for criteria-based simple habits
		level, err := f.performAutomaticScoring(habit, value)
		if err != nil {
			return nil, fmt.Errorf("automatic scoring failed: %w", err)
		}
		achievementLevel = level
	} else {
		// Manual scoring - simple habits default to pass/fail based on primary field
		level := f.determineManualAchievement(habit, value)
		achievementLevel = level
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: achievementLevel,
		Notes:            notes,
		Status:           models.EntryCompleted, // Testing method defaults to completed
	}, nil
}

// CollectEntry collects entry for simple habits with pass/fail logic
func (f *SimpleHabitCollectionFlow) CollectEntry(habit models.Habit, existing *ExistingEntry) (*EntryResult, error) {
	// Simple habits have primary pass/fail determination
	// Additional data fields are optional supplements

	// Create field input configuration
	config := EntryFieldInputConfig{
		Habit:         habit,
		FieldType:     habit.FieldType,
		ExistingEntry: existing,
		ShowScoring:   habit.ScoringType == models.AutomaticScoring,
	}

	// Create field input component
	input, err := f.factory.CreateScoringAwareInput(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create field input: %w", err)
	}

	// Create and run the input form
	// AIDEV-NOTE: T024-bug2-resolved; form.Run() no longer causes looping in entry menu (bypassed by modal system)
	// AIDEV-NOTE: T024-architecture; CLI entry collection still uses this flow, entry menu uses modal system instead
	form := input.CreateInputForm(habit)
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("input form failed: %w", err)
	}

	// Get the collected value
	value := input.GetValue()

	// AIDEV-NOTE: T012/2.1-skip-integration; status-aware processing with skip detection for Boolean inputs
	// Determine entry status - check if input supports skip functionality
	status := models.EntryCompleted // Default status
	if boolInput, ok := input.(*BooleanEntryInput); ok {
		status = boolInput.GetStatus()
	} else if value == nil {
		status = models.EntrySkipped
	} else {
		// For non-boolean inputs, determine status based on value
		switch habit.FieldType.Type {
		case models.BooleanFieldType:
			if boolVal, isBool := value.(bool); isBool {
				if boolVal {
					status = models.EntryCompleted
				} else {
					status = models.EntryFailed
				}
			}
		default:
			// Other field types default to completed if value exists
			if value != nil {
				status = models.EntryCompleted
			} else {
				status = models.EntrySkipped
			}
		}
	}

	// Handle scoring based on habit configuration (skip scoring for skipped entries)
	var achievementLevel *models.AchievementLevel
	if status != models.EntrySkipped {
		if habit.ScoringType == models.AutomaticScoring {
			// Automatic scoring for criteria-based simple habits
			level, err := f.performAutomaticScoring(habit, value)
			if err != nil {
				return nil, fmt.Errorf("automatic scoring failed: %w", err)
			}
			achievementLevel = level

			// Update input display with scoring feedback
			if input.CanShowScoring() {
				_ = input.UpdateScoringDisplay(achievementLevel) // Non-fatal error - continue without scoring display
			}
		} else {
			// Manual scoring - simple habits default to pass/fail based on primary field
			level := f.determineManualAchievement(habit, value)
			achievementLevel = level
		}
	}

	// Collect optional notes (skip note prompts for skipped entries but preserve existing notes)
	var notes string
	if status == models.EntrySkipped {
		// For skipped entries, preserve existing notes but don't prompt for new ones
		if existing != nil {
			notes = existing.Notes
		}
	} else {
		collectedNotes, err := f.collectOptionalNotes(habit, value, existing)
		if err != nil {
			return nil, fmt.Errorf("failed to collect notes: %w", err)
		}
		notes = collectedNotes
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: achievementLevel,
		Notes:            notes,
		Status:           status,
	}, nil
}

// GetFlowType returns the habit type
func (f *SimpleHabitCollectionFlow) GetFlowType() string {
	return string(models.SimpleHabit)
}

// RequiresScoring indicates simple habits may use scoring
func (f *SimpleHabitCollectionFlow) RequiresScoring() bool {
	return true
}

// GetExpectedFieldTypes returns supported field types for simple habits
func (f *SimpleHabitCollectionFlow) GetExpectedFieldTypes() []string {
	// Simple habits support all field types except checklist (per T009 design)
	return []string{
		models.BooleanFieldType,
		models.TextFieldType,
		models.UnsignedIntFieldType,
		models.UnsignedDecimalFieldType,
		models.DecimalFieldType,
		models.TimeFieldType,
		models.DurationFieldType,
	}
}

// ElasticHabitCollectionFlow handles data input with mini/midi/maxi achievement feedback
type ElasticHabitCollectionFlow struct {
	factory       *EntryFieldInputFactory
	scoringEngine *scoring.Engine
}

// NewElasticHabitCollectionFlow creates a new elastic habit collection flow
func NewElasticHabitCollectionFlow(factory *EntryFieldInputFactory, scoringEngine *scoring.Engine) *ElasticHabitCollectionFlow {
	return &ElasticHabitCollectionFlow{
		factory:       factory,
		scoringEngine: scoringEngine,
	}
}

// NewElasticHabitCollectionFlowForTesting creates a flow for testing that bypasses user interaction
func NewElasticHabitCollectionFlowForTesting(factory *EntryFieldInputFactory, scoringEngine *scoring.Engine) *ElasticHabitCollectionFlow {
	return &ElasticHabitCollectionFlow{
		factory:       factory,
		scoringEngine: scoringEngine,
	}
}

// CollectEntry collects entry for elastic habits with immediate achievement calculation
func (f *ElasticHabitCollectionFlow) CollectEntry(habit models.Habit, existing *ExistingEntry) (*EntryResult, error) {
	// Elastic habits focus on achievement levels with immediate feedback

	// Create field input configuration with scoring enabled
	config := EntryFieldInputConfig{
		Habit:         habit,
		FieldType:     habit.FieldType,
		ExistingEntry: existing,
		ShowScoring:   true, // Always show scoring for elastic habits
	}

	// Create field input component
	input, err := f.factory.CreateScoringAwareInput(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create field input: %w", err)
	}

	// Display criteria information for motivation
	f.displayCriteriaInformation(habit)

	// Create and run the input form
	// AIDEV-NOTE: T024-bug2-resolved; form.Run() no longer causes looping in entry menu (bypassed by modal system)
	// AIDEV-NOTE: T024-architecture; CLI entry collection still uses this flow, entry menu uses modal system instead
	form := input.CreateInputForm(habit)
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("input form failed: %w", err)
	}

	// Get the collected value and status
	value := input.GetValue()
	status := input.GetStatus()

	// Skip processing if entry was skipped
	var achievementLevel *models.AchievementLevel
	var notes string
	if status != models.EntrySkipped {
		// Perform scoring (elastic habits require achievement level determination)
		if habit.ScoringType == models.AutomaticScoring {
			// Automatic scoring with three-tier criteria
			level, err := f.performElasticScoring(habit, value)
			if err != nil {
				return nil, fmt.Errorf("elastic scoring failed: %w", err)
			}
			achievementLevel = level
		} else {
			// Manual scoring with achievement level selection
			level, err := f.collectManualAchievementLevel(habit, value)
			if err != nil {
				return nil, fmt.Errorf("manual achievement selection failed: %w", err)
			}
			achievementLevel = level
		}

		// Update input display with achievement feedback
		if input.CanShowScoring() && achievementLevel != nil {
			_ = input.UpdateScoringDisplay(achievementLevel) // Non-fatal error - continue without scoring display
		}

		// Display achievement result
		f.displayAchievementResult(habit, value, achievementLevel)

		// Collect optional notes
		var err error
		notes, err = f.collectOptionalNotes(habit, value, existing)
		if err != nil {
			return nil, fmt.Errorf("failed to collect notes: %w", err)
		}
	} else if existing != nil {
		// For skipped entries, preserve existing notes but don't collect new ones
		notes = existing.Notes
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: achievementLevel,
		Notes:            notes,
		Status:           status,
	}, nil
}

// CollectEntryDirectly bypasses UI interaction and creates entry directly from provided value
func (f *ElasticHabitCollectionFlow) CollectEntryDirectly(habit models.Habit, value interface{}, notes string, _ *ExistingEntry) (*EntryResult, error) {
	// Handle scoring based on habit configuration
	var achievementLevel *models.AchievementLevel
	if habit.ScoringType == models.AutomaticScoring {
		// Automatic scoring with three-tier criteria
		level, err := f.performElasticScoring(habit, value)
		if err != nil {
			return nil, fmt.Errorf("elastic scoring failed: %w", err)
		}
		achievementLevel = level
	} else {
		// For testing, determine achievement level based on value patterns
		level := f.determineTestingAchievementLevel(habit, value)
		achievementLevel = level
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: achievementLevel,
		Notes:            notes,
		Status:           models.EntryCompleted, // Testing method defaults to completed
	}, nil
}

// determineTestingAchievementLevel provides simplified achievement determination for testing
func (f *ElasticHabitCollectionFlow) determineTestingAchievementLevel(habit models.Habit, value interface{}) *models.AchievementLevel {
	// Simplified logic for testing - in real scenarios, manual selection would be used
	switch habit.FieldType.Type {
	case models.BooleanFieldType:
		if boolVal, ok := value.(bool); ok && boolVal {
			level := models.AchievementMini
			return &level
		}
	case models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType:
		// Simple numeric achievement level determination for testing
		if numVal, ok := value.(float64); ok {
			switch {
			case numVal >= 100:
				level := models.AchievementMaxi
				return &level
			case numVal >= 50:
				level := models.AchievementMidi
				return &level
			case numVal > 0:
				level := models.AchievementMini
				return &level
			}
		}
		if intVal, ok := value.(int); ok {
			switch {
			case intVal >= 100:
				level := models.AchievementMaxi
				return &level
			case intVal >= 50:
				level := models.AchievementMidi
				return &level
			case intVal > 0:
				level := models.AchievementMini
				return &level
			}
		}
	default:
		// For other field types, default to Mini if value is provided
		if value != nil {
			level := models.AchievementMini
			return &level
		}
	}

	// Default to None
	level := models.AchievementNone
	return &level
}

// GetFlowType returns the habit type
func (f *ElasticHabitCollectionFlow) GetFlowType() string {
	return string(models.ElasticHabit)
}

// RequiresScoring indicates elastic habits always use scoring
func (f *ElasticHabitCollectionFlow) RequiresScoring() bool {
	return true
}

// GetExpectedFieldTypes returns supported field types for elastic habits
func (f *ElasticHabitCollectionFlow) GetExpectedFieldTypes() []string {
	// Elastic habits support all field types
	return []string{
		models.BooleanFieldType,
		models.TextFieldType,
		models.UnsignedIntFieldType,
		models.UnsignedDecimalFieldType,
		models.DecimalFieldType,
		models.TimeFieldType,
		models.DurationFieldType,
		models.ChecklistFieldType,
	}
}

// InformationalHabitCollectionFlow handles data-only collection without evaluation
type InformationalHabitCollectionFlow struct {
	factory *EntryFieldInputFactory
}

// NewInformationalHabitCollectionFlow creates a new informational habit collection flow
func NewInformationalHabitCollectionFlow(factory *EntryFieldInputFactory) *InformationalHabitCollectionFlow {
	return &InformationalHabitCollectionFlow{
		factory: factory,
	}
}

// NewInformationalHabitCollectionFlowForTesting creates a flow for testing that bypasses user interaction
func NewInformationalHabitCollectionFlowForTesting(factory *EntryFieldInputFactory) *InformationalHabitCollectionFlow {
	return &InformationalHabitCollectionFlow{
		factory: factory,
	}
}

// CollectEntry collects entry for informational habits without scoring
func (f *InformationalHabitCollectionFlow) CollectEntry(habit models.Habit, existing *ExistingEntry) (*EntryResult, error) {
	// Informational habits collect data without pass/fail evaluation

	// Create field input configuration without scoring
	config := EntryFieldInputConfig{
		Habit:         habit,
		FieldType:     habit.FieldType,
		ExistingEntry: existing,
		ShowScoring:   false, // No scoring for informational habits
	}

	// Create field input component
	input, err := f.factory.CreateInput(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create field input: %w", err)
	}

	// Display informational context
	f.displayInformationalContext(habit)

	// Create and run the input form
	// AIDEV-NOTE: T024-bug2-resolved; form.Run() no longer causes looping in entry menu (bypassed by modal system)
	// AIDEV-NOTE: T024-architecture; CLI entry collection still uses this flow, entry menu uses modal system instead
	form := input.CreateInputForm(habit)
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("input form failed: %w", err)
	}

	// Get the collected value and status
	value := input.GetValue()
	status := input.GetStatus()

	// Skip processing if entry was skipped
	var notes string
	if status != models.EntrySkipped {
		// Display direction-aware feedback if configured
		f.displayDirectionFeedback(habit, value)

		// Collect optional notes
		var err error
		notes, err = f.collectOptionalNotes(habit, value, existing)
		if err != nil {
			return nil, fmt.Errorf("failed to collect notes: %w", err)
		}
	} else if existing != nil {
		// For skipped entries, preserve existing notes but don't collect new ones
		notes = existing.Notes
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: nil, // No achievement level for informational habits
		Notes:            notes,
		Status:           status,
	}, nil
}

// CollectEntryDirectly bypasses UI interaction and creates entry directly from provided value
func (f *InformationalHabitCollectionFlow) CollectEntryDirectly(_ models.Habit, value interface{}, notes string, _ *ExistingEntry) (*EntryResult, error) {
	// Informational habits simply record data without any scoring or evaluation
	return &EntryResult{
		Value:            value,
		AchievementLevel: nil, // Informational habits never have achievement levels
		Notes:            notes,
		Status:           models.EntryCompleted, // Testing method defaults to completed
	}, nil
}

// GetFlowType returns the habit type
func (f *InformationalHabitCollectionFlow) GetFlowType() string {
	return string(models.InformationalHabit)
}

// RequiresScoring indicates informational habits don't use scoring
func (f *InformationalHabitCollectionFlow) RequiresScoring() bool {
	return false
}

// GetExpectedFieldTypes returns supported field types for informational habits
func (f *InformationalHabitCollectionFlow) GetExpectedFieldTypes() []string {
	// Informational habits support all field types
	return []string{
		models.BooleanFieldType,
		models.TextFieldType,
		models.UnsignedIntFieldType,
		models.UnsignedDecimalFieldType,
		models.DecimalFieldType,
		models.TimeFieldType,
		models.DurationFieldType,
		models.ChecklistFieldType,
	}
}

// ChecklistHabitCollectionFlow handles interactive checklist completion with progress feedback
type ChecklistHabitCollectionFlow struct {
	factory         *EntryFieldInputFactory
	scoringEngine   *scoring.Engine
	checklistParser *parser.ChecklistParser
	checklistsPath  string
}

// NewChecklistHabitCollectionFlow creates a new checklist habit collection flow
func NewChecklistHabitCollectionFlow(factory *EntryFieldInputFactory, scoringEngine *scoring.Engine, checklistsPath string) *ChecklistHabitCollectionFlow {
	return &ChecklistHabitCollectionFlow{
		factory:         factory,
		scoringEngine:   scoringEngine,
		checklistParser: parser.NewChecklistParser(),
		checklistsPath:  checklistsPath,
	}
}

// loadChecklistData loads the actual checklist data from the file based on the habit's ChecklistID
func (f *ChecklistHabitCollectionFlow) loadChecklistData(habit models.Habit) (*models.Checklist, error) {
	// Validate that the habit has a ChecklistID
	if habit.FieldType.ChecklistID == "" {
		return nil, fmt.Errorf("habit field type missing checklist_id")
	}

	// Load checklist schema from file
	schema, err := f.checklistParser.LoadFromFile(f.checklistsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load checklists from %s: %w", f.checklistsPath, err)
	}

	// Find the checklist by ID
	checklist, err := f.checklistParser.GetChecklistByID(schema, habit.FieldType.ChecklistID)
	if err != nil {
		return nil, fmt.Errorf("checklist with ID '%s' not found: %w", habit.FieldType.ChecklistID, err)
	}

	return checklist, nil
}

// CollectEntry collects entry for checklist habits with progress tracking
func (f *ChecklistHabitCollectionFlow) CollectEntry(habit models.Habit, existing *ExistingEntry) (*EntryResult, error) {
	// Checklist habits use checklist field type exclusively
	if habit.FieldType.Type != models.ChecklistFieldType {
		return nil, fmt.Errorf("checklist habits require checklist field type, got: %s", habit.FieldType.Type)
	}

	// Create field input configuration
	config := EntryFieldInputConfig{
		Habit:          habit,
		FieldType:      habit.FieldType,
		ExistingEntry:  existing,
		ShowScoring:    habit.ScoringType == models.AutomaticScoring,
		ChecklistsPath: f.checklistsPath,
	}

	// Create checklist input component
	input, err := f.factory.CreateScoringAwareInput(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create checklist input: %w", err)
	}

	// Display checklist progress context
	f.displayChecklistContext(habit, existing)

	// Create and run the checklist form
	form := input.CreateInputForm(habit)
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("checklist form failed: %w", err)
	}

	// Get the collected checklist selections
	value := input.GetValue()

	// AIDEV-NOTE: T012/2.3-skip-integration; status-aware processing with skip detection for Checklist inputs
	// Determine entry status - check if input supports skip functionality
	status := models.EntryCompleted // Default status
	if checklistInput, ok := input.(*ChecklistEntryInput); ok {
		status = checklistInput.GetStatus()
	} else if value == nil {
		status = models.EntrySkipped
	}

	// Handle scoring based on completion percentage (skip scoring for skipped entries)
	var achievementLevel *models.AchievementLevel
	if status != models.EntrySkipped {
		if habit.ScoringType == models.AutomaticScoring {
			// Automatic scoring based on checklist completion criteria
			level, err := f.performChecklistScoring(habit, value)
			if err != nil {
				return nil, fmt.Errorf("checklist scoring failed: %w", err)
			}
			achievementLevel = level
		} else {
			// Manual scoring with achievement level selection
			level, err := f.collectManualAchievementLevel(habit, value)
			if err != nil {
				return nil, fmt.Errorf("manual achievement selection failed: %w", err)
			}
			achievementLevel = level
		}

		// Update input display with scoring feedback
		if input.CanShowScoring() && achievementLevel != nil {
			_ = input.UpdateScoringDisplay(achievementLevel) // Non-fatal error - continue without scoring display
		}

		// Display completion progress feedback
		f.displayCompletionProgress(habit, value, achievementLevel)
	}

	// Collect optional notes (skip note prompts for skipped entries but preserve existing notes)
	var notes string
	if status == models.EntrySkipped {
		// For skipped entries, preserve existing notes but don't prompt for new ones
		if existing != nil {
			notes = existing.Notes
		}
	} else {
		collectedNotes, err := f.collectOptionalNotes(habit, value, existing)
		if err != nil {
			return nil, fmt.Errorf("failed to collect notes: %w", err)
		}
		notes = collectedNotes
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: achievementLevel,
		Notes:            notes,
		Status:           status,
	}, nil
}

// CollectEntryDirectly provides headless entry collection for testing
func (f *ChecklistHabitCollectionFlow) CollectEntryDirectly(habit models.Habit, value interface{}, notes string, _ *ExistingEntry) (*EntryResult, error) {
	// Handle scoring based on habit configuration
	var achievementLevel *models.AchievementLevel
	if habit.ScoringType == models.AutomaticScoring {
		// Automatic scoring with checklist criteria
		level, err := f.performChecklistScoring(habit, value)
		if err != nil {
			return nil, fmt.Errorf("checklist scoring failed: %w", err)
		}
		achievementLevel = level
	} else {
		// For testing, determine achievement level based on checklist completion
		level := f.determineTestingAchievementLevel(habit, value)
		achievementLevel = level
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: achievementLevel,
		Notes:            notes,
		Status:           models.EntryCompleted, // Testing method defaults to completed
	}, nil
}

// determineTestingAchievementLevel provides simplified achievement determination for testing
func (f *ChecklistHabitCollectionFlow) determineTestingAchievementLevel(habit models.Habit, value interface{}) *models.AchievementLevel {
	// For testing, determine based on checklist completion
	if items, ok := value.([]string); ok {
		// Load actual checklist data to get total item count
		checklist, err := f.loadChecklistData(habit)
		if err != nil {
			// For testing, treat failure as none achievement
			level := models.AchievementNone
			return &level
		}

		completed := len(items)
		total := checklist.GetTotalItemCount()

		if total == 0 {
			level := models.AchievementNone
			return &level
		}

		percentage := float64(completed) / float64(total)

		switch {
		case percentage >= 1.0:
			level := models.AchievementMaxi
			return &level
		case percentage >= 0.75:
			level := models.AchievementMidi
			return &level
		case percentage >= 0.5:
			level := models.AchievementMini
			return &level
		default:
			level := models.AchievementNone
			return &level
		}
	}

	// Default to None for invalid value types
	level := models.AchievementNone
	return &level
}

// GetFlowType returns the habit type
func (f *ChecklistHabitCollectionFlow) GetFlowType() string {
	return string(models.ChecklistHabit)
}

// RequiresScoring indicates checklist habits may use scoring
func (f *ChecklistHabitCollectionFlow) RequiresScoring() bool {
	return true
}

// GetExpectedFieldTypes returns supported field types for checklist habits
func (f *ChecklistHabitCollectionFlow) GetExpectedFieldTypes() []string {
	// Checklist habits only support checklist field type
	return []string{
		models.ChecklistFieldType,
	}
}
