package entry

import (
	"fmt"

	"davidlee/iter/internal/models"
	"davidlee/iter/internal/scoring"
)

// AIDEV-NOTE: goal-collection-flows; defines specialized collection flows for each goal type with field input integration
// Integrates T010/1.2 field input components with goal type-specific behaviors and scoring patterns
// AIDEV-NOTE: T010/3.1-complete; SimpleGoalCollectionFlow fully implemented with headless testing support
// Features: pass/fail logic, field type support (all except checklist), automatic/manual scoring, notes collection

// GoalCollectionFlow defines the interface for goal type-specific collection flows
type GoalCollectionFlow interface {
	// CollectEntry orchestrates the complete entry collection for a goal type
	CollectEntry(goal models.Goal, existing *ExistingEntry) (*EntryResult, error)

	// GetFlowType returns the goal type this flow handles
	GetFlowType() string

	// RequiresScoring indicates if this flow needs scoring engine integration
	RequiresScoring() bool

	// GetExpectedFieldTypes returns field types supported by this flow
	GetExpectedFieldTypes() []string
}

// SimpleGoalCollectionFlow handles pass/fail collection with optional additional data
type SimpleGoalCollectionFlow struct {
	factory       *EntryFieldInputFactory
	scoringEngine *scoring.Engine
}

// NewSimpleGoalCollectionFlow creates a new simple goal collection flow
func NewSimpleGoalCollectionFlow(factory *EntryFieldInputFactory, scoringEngine *scoring.Engine) *SimpleGoalCollectionFlow {
	return &SimpleGoalCollectionFlow{
		factory:       factory,
		scoringEngine: scoringEngine,
	}
}

// NewSimpleGoalCollectionFlowForTesting creates a flow for testing that bypasses user interaction
func NewSimpleGoalCollectionFlowForTesting(factory *EntryFieldInputFactory, scoringEngine *scoring.Engine) *SimpleGoalCollectionFlow {
	return &SimpleGoalCollectionFlow{
		factory:       factory,
		scoringEngine: scoringEngine,
	}
}

// CollectEntryDirectly bypasses UI interaction and creates entry directly from provided value
func (f *SimpleGoalCollectionFlow) CollectEntryDirectly(goal models.Goal, value interface{}, notes string, _ *ExistingEntry) (*EntryResult, error) {
	// Handle scoring based on goal configuration
	var achievementLevel *models.AchievementLevel
	if goal.ScoringType == models.AutomaticScoring {
		// Automatic scoring for criteria-based simple goals
		level, err := f.performAutomaticScoring(goal, value)
		if err != nil {
			return nil, fmt.Errorf("automatic scoring failed: %w", err)
		}
		achievementLevel = level
	} else {
		// Manual scoring - simple goals default to pass/fail based on primary field
		level := f.determineManualAchievement(goal, value)
		achievementLevel = level
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: achievementLevel,
		Notes:            notes,
	}, nil
}

// CollectEntry collects entry for simple goals with pass/fail logic
func (f *SimpleGoalCollectionFlow) CollectEntry(goal models.Goal, existing *ExistingEntry) (*EntryResult, error) {
	// Simple goals have primary pass/fail determination
	// Additional data fields are optional supplements

	// Create field input configuration
	config := EntryFieldInputConfig{
		Goal:          goal,
		FieldType:     goal.FieldType,
		ExistingEntry: existing,
		ShowScoring:   goal.ScoringType == models.AutomaticScoring,
	}

	// Create field input component
	input, err := f.factory.CreateScoringAwareInput(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create field input: %w", err)
	}

	// Create and run the input form
	form := input.CreateInputForm(goal)
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("input form failed: %w", err)
	}

	// Get the collected value
	value := input.GetValue()

	// Handle scoring based on goal configuration
	var achievementLevel *models.AchievementLevel
	if goal.ScoringType == models.AutomaticScoring {
		// Automatic scoring for criteria-based simple goals
		level, err := f.performAutomaticScoring(goal, value)
		if err != nil {
			return nil, fmt.Errorf("automatic scoring failed: %w", err)
		}
		achievementLevel = level

		// Update input display with scoring feedback
		if input.CanShowScoring() {
			_ = input.UpdateScoringDisplay(achievementLevel) // Non-fatal error - continue without scoring display
		}
	} else {
		// Manual scoring - simple goals default to pass/fail based on primary field
		level := f.determineManualAchievement(goal, value)
		achievementLevel = level
	}

	// Collect optional notes
	notes, err := f.collectOptionalNotes(goal, value, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to collect notes: %w", err)
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: achievementLevel,
		Notes:            notes,
	}, nil
}

// GetFlowType returns the goal type
func (f *SimpleGoalCollectionFlow) GetFlowType() string {
	return string(models.SimpleGoal)
}

// RequiresScoring indicates simple goals may use scoring
func (f *SimpleGoalCollectionFlow) RequiresScoring() bool {
	return true
}

// GetExpectedFieldTypes returns supported field types for simple goals
func (f *SimpleGoalCollectionFlow) GetExpectedFieldTypes() []string {
	// Simple goals support all field types except checklist (per T009 design)
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

// ElasticGoalCollectionFlow handles data input with mini/midi/maxi achievement feedback
type ElasticGoalCollectionFlow struct {
	factory       *EntryFieldInputFactory
	scoringEngine *scoring.Engine
}

// NewElasticGoalCollectionFlow creates a new elastic goal collection flow
func NewElasticGoalCollectionFlow(factory *EntryFieldInputFactory, scoringEngine *scoring.Engine) *ElasticGoalCollectionFlow {
	return &ElasticGoalCollectionFlow{
		factory:       factory,
		scoringEngine: scoringEngine,
	}
}

// NewElasticGoalCollectionFlowForTesting creates a flow for testing that bypasses user interaction
func NewElasticGoalCollectionFlowForTesting(factory *EntryFieldInputFactory, scoringEngine *scoring.Engine) *ElasticGoalCollectionFlow {
	return &ElasticGoalCollectionFlow{
		factory:       factory,
		scoringEngine: scoringEngine,
	}
}

// CollectEntry collects entry for elastic goals with immediate achievement calculation
func (f *ElasticGoalCollectionFlow) CollectEntry(goal models.Goal, existing *ExistingEntry) (*EntryResult, error) {
	// Elastic goals focus on achievement levels with immediate feedback

	// Create field input configuration with scoring enabled
	config := EntryFieldInputConfig{
		Goal:          goal,
		FieldType:     goal.FieldType,
		ExistingEntry: existing,
		ShowScoring:   true, // Always show scoring for elastic goals
	}

	// Create field input component
	input, err := f.factory.CreateScoringAwareInput(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create field input: %w", err)
	}

	// Display criteria information for motivation
	f.displayCriteriaInformation(goal)

	// Create and run the input form
	form := input.CreateInputForm(goal)
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("input form failed: %w", err)
	}

	// Get the collected value
	value := input.GetValue()

	// Perform scoring (elastic goals require achievement level determination)
	var achievementLevel *models.AchievementLevel
	if goal.ScoringType == models.AutomaticScoring {
		// Automatic scoring with three-tier criteria
		level, err := f.performElasticScoring(goal, value)
		if err != nil {
			return nil, fmt.Errorf("elastic scoring failed: %w", err)
		}
		achievementLevel = level
	} else {
		// Manual scoring with achievement level selection
		level, err := f.collectManualAchievementLevel(goal, value)
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
	f.displayAchievementResult(goal, value, achievementLevel)

	// Collect optional notes
	notes, err := f.collectOptionalNotes(goal, value, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to collect notes: %w", err)
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: achievementLevel,
		Notes:            notes,
	}, nil
}

// CollectEntryDirectly bypasses UI interaction and creates entry directly from provided value
func (f *ElasticGoalCollectionFlow) CollectEntryDirectly(goal models.Goal, value interface{}, notes string, _ *ExistingEntry) (*EntryResult, error) {
	// Handle scoring based on goal configuration
	var achievementLevel *models.AchievementLevel
	if goal.ScoringType == models.AutomaticScoring {
		// Automatic scoring with three-tier criteria
		level, err := f.performElasticScoring(goal, value)
		if err != nil {
			return nil, fmt.Errorf("elastic scoring failed: %w", err)
		}
		achievementLevel = level
	} else {
		// For testing, determine achievement level based on value patterns
		level := f.determineTestingAchievementLevel(goal, value)
		achievementLevel = level
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: achievementLevel,
		Notes:            notes,
	}, nil
}

// determineTestingAchievementLevel provides simplified achievement determination for testing
func (f *ElasticGoalCollectionFlow) determineTestingAchievementLevel(goal models.Goal, value interface{}) *models.AchievementLevel {
	// Simplified logic for testing - in real scenarios, manual selection would be used
	switch goal.FieldType.Type {
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

// GetFlowType returns the goal type
func (f *ElasticGoalCollectionFlow) GetFlowType() string {
	return string(models.ElasticGoal)
}

// RequiresScoring indicates elastic goals always use scoring
func (f *ElasticGoalCollectionFlow) RequiresScoring() bool {
	return true
}

// GetExpectedFieldTypes returns supported field types for elastic goals
func (f *ElasticGoalCollectionFlow) GetExpectedFieldTypes() []string {
	// Elastic goals support all field types
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

// InformationalGoalCollectionFlow handles data-only collection without evaluation
type InformationalGoalCollectionFlow struct {
	factory *EntryFieldInputFactory
}

// NewInformationalGoalCollectionFlow creates a new informational goal collection flow
func NewInformationalGoalCollectionFlow(factory *EntryFieldInputFactory) *InformationalGoalCollectionFlow {
	return &InformationalGoalCollectionFlow{
		factory: factory,
	}
}

// CollectEntry collects entry for informational goals without scoring
func (f *InformationalGoalCollectionFlow) CollectEntry(goal models.Goal, existing *ExistingEntry) (*EntryResult, error) {
	// Informational goals collect data without pass/fail evaluation

	// Create field input configuration without scoring
	config := EntryFieldInputConfig{
		Goal:          goal,
		FieldType:     goal.FieldType,
		ExistingEntry: existing,
		ShowScoring:   false, // No scoring for informational goals
	}

	// Create field input component
	input, err := f.factory.CreateInput(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create field input: %w", err)
	}

	// Display informational context
	f.displayInformationalContext(goal)

	// Create and run the input form
	form := input.CreateInputForm(goal)
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("input form failed: %w", err)
	}

	// Get the collected value
	value := input.GetValue()

	// Display direction-aware feedback if configured
	f.displayDirectionFeedback(goal, value)

	// Collect optional notes
	notes, err := f.collectOptionalNotes(goal, value, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to collect notes: %w", err)
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: nil, // No achievement level for informational goals
		Notes:            notes,
	}, nil
}

// GetFlowType returns the goal type
func (f *InformationalGoalCollectionFlow) GetFlowType() string {
	return string(models.InformationalGoal)
}

// RequiresScoring indicates informational goals don't use scoring
func (f *InformationalGoalCollectionFlow) RequiresScoring() bool {
	return false
}

// GetExpectedFieldTypes returns supported field types for informational goals
func (f *InformationalGoalCollectionFlow) GetExpectedFieldTypes() []string {
	// Informational goals support all field types
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

// ChecklistGoalCollectionFlow handles interactive checklist completion with progress feedback
type ChecklistGoalCollectionFlow struct {
	factory       *EntryFieldInputFactory
	scoringEngine *scoring.Engine
}

// NewChecklistGoalCollectionFlow creates a new checklist goal collection flow
func NewChecklistGoalCollectionFlow(factory *EntryFieldInputFactory, scoringEngine *scoring.Engine) *ChecklistGoalCollectionFlow {
	return &ChecklistGoalCollectionFlow{
		factory:       factory,
		scoringEngine: scoringEngine,
	}
}

// CollectEntry collects entry for checklist goals with progress tracking
func (f *ChecklistGoalCollectionFlow) CollectEntry(goal models.Goal, existing *ExistingEntry) (*EntryResult, error) {
	// Checklist goals use checklist field type exclusively
	if goal.FieldType.Type != models.ChecklistFieldType {
		return nil, fmt.Errorf("checklist goals require checklist field type, got: %s", goal.FieldType.Type)
	}

	// Create field input configuration
	config := EntryFieldInputConfig{
		Goal:          goal,
		FieldType:     goal.FieldType,
		ExistingEntry: existing,
		ShowScoring:   goal.ScoringType == models.AutomaticScoring,
	}

	// Create checklist input component
	input, err := f.factory.CreateScoringAwareInput(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create checklist input: %w", err)
	}

	// Display checklist progress context
	f.displayChecklistContext(goal, existing)

	// Create and run the checklist form
	form := input.CreateInputForm(goal)
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("checklist form failed: %w", err)
	}

	// Get the collected checklist selections
	value := input.GetValue()

	// Handle scoring based on completion percentage
	var achievementLevel *models.AchievementLevel
	if goal.ScoringType == models.AutomaticScoring {
		// Automatic scoring based on checklist completion criteria
		level, err := f.performChecklistScoring(goal, value)
		if err != nil {
			return nil, fmt.Errorf("checklist scoring failed: %w", err)
		}
		achievementLevel = level
	} else {
		// Manual scoring with achievement level selection
		level, err := f.collectManualAchievementLevel(goal, value)
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
	f.displayCompletionProgress(goal, value, achievementLevel)

	// Collect optional notes
	notes, err := f.collectOptionalNotes(goal, value, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to collect notes: %w", err)
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: achievementLevel,
		Notes:            notes,
	}, nil
}

// GetFlowType returns the goal type
func (f *ChecklistGoalCollectionFlow) GetFlowType() string {
	return string(models.ChecklistGoal)
}

// RequiresScoring indicates checklist goals may use scoring
func (f *ChecklistGoalCollectionFlow) RequiresScoring() bool {
	return true
}

// GetExpectedFieldTypes returns supported field types for checklist goals
func (f *ChecklistGoalCollectionFlow) GetExpectedFieldTypes() []string {
	// Checklist goals only support checklist field type
	return []string{
		models.ChecklistFieldType,
	}
}
