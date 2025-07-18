package entry

import (
	"fmt"

	"github.com/davidlee/vice/internal/models"
	"github.com/davidlee/vice/internal/scoring"
)

// AIDEV-NOTE: flow-factory; creates appropriate habit collection flows with field input and scoring integration
// Central factory for coordinating habit type-specific collection flows with T010/1.2 field input components

// HabitCollectionFlowFactory creates appropriate collection flows for different habit types
type HabitCollectionFlowFactory struct {
	fieldInputFactory *EntryFieldInputFactory
	scoringEngine     *scoring.Engine
	checklistsPath    string
}

// NewHabitCollectionFlowFactory creates a new habit collection flow factory
func NewHabitCollectionFlowFactory(fieldInputFactory *EntryFieldInputFactory, scoringEngine *scoring.Engine, checklistsPath string) *HabitCollectionFlowFactory {
	return &HabitCollectionFlowFactory{
		fieldInputFactory: fieldInputFactory,
		scoringEngine:     scoringEngine,
		checklistsPath:    checklistsPath,
	}
}

// CreateFlow creates the appropriate collection flow for a given habit type
func (f *HabitCollectionFlowFactory) CreateFlow(habitType string) (HabitCollectionFlow, error) {
	switch habitType {
	case string(models.SimpleHabit):
		return NewSimpleHabitCollectionFlow(f.fieldInputFactory, f.scoringEngine), nil

	case string(models.ElasticHabit):
		return NewElasticHabitCollectionFlow(f.fieldInputFactory, f.scoringEngine), nil

	case string(models.InformationalHabit):
		return NewInformationalHabitCollectionFlow(f.fieldInputFactory), nil

	case string(models.ChecklistHabit):
		return NewChecklistHabitCollectionFlow(f.fieldInputFactory, f.scoringEngine, f.checklistsPath), nil

	default:
		return nil, fmt.Errorf("unsupported habit type: %s", habitType)
	}
}

// ValidateHabitForFlow validates that a habit is compatible with its intended flow
func (f *HabitCollectionFlowFactory) ValidateHabitForFlow(habit models.Habit) error {
	flow, err := f.CreateFlow(string(habit.HabitType))
	if err != nil {
		return fmt.Errorf("unable to create flow for habit type %s: %w", habit.HabitType, err)
	}

	// Check if the habit's field type is supported by the flow
	supportedTypes := flow.GetExpectedFieldTypes()
	fieldTypeSupported := false
	for _, supportedType := range supportedTypes {
		if habit.FieldType.Type == supportedType {
			fieldTypeSupported = true
			break
		}
	}

	if !fieldTypeSupported {
		return fmt.Errorf("field type %s not supported by %s habit flow", habit.FieldType.Type, habit.HabitType)
	}

	// Check if scoring requirements are met
	if flow.RequiresScoring() && habit.ScoringType == models.AutomaticScoring && f.scoringEngine == nil {
		return fmt.Errorf("automatic scoring required for %s habit but no scoring engine available", habit.HabitType)
	}

	return nil
}

// GetSupportedHabitTypes returns the list of habit types supported by the factory
func (f *HabitCollectionFlowFactory) GetSupportedHabitTypes() []string {
	return []string{
		string(models.SimpleHabit),
		string(models.ElasticHabit),
		string(models.InformationalHabit),
		string(models.ChecklistHabit),
	}
}

// GetFlowInfo returns information about a specific habit type flow
func (f *HabitCollectionFlowFactory) GetFlowInfo(habitType string) (*FlowInfo, error) {
	flow, err := f.CreateFlow(habitType)
	if err != nil {
		return nil, err
	}

	return &FlowInfo{
		HabitType:           habitType,
		RequiresScoring:     flow.RequiresScoring(),
		SupportedFieldTypes: flow.GetExpectedFieldTypes(),
		Description:         f.getFlowDescription(habitType),
	}, nil
}

// FlowInfo provides information about a habit collection flow
type FlowInfo struct {
	HabitType           string
	RequiresScoring     bool
	SupportedFieldTypes []string
	Description         string
}

// Private helper methods

func (f *HabitCollectionFlowFactory) getFlowDescription(habitType string) string {
	switch habitType {
	case string(models.SimpleHabit):
		return "Pass/fail collection with optional additional data and automatic/manual scoring support"
	case string(models.ElasticHabit):
		return "Data collection with immediate mini/midi/maxi achievement calculation and feedback"
	case string(models.InformationalHabit):
		return "Data-only collection without pass/fail evaluation, for tracking purposes"
	case string(models.ChecklistHabit):
		return "Interactive checklist completion with progress tracking and achievement scoring"
	default:
		return "Unknown habit type"
	}
}

// CollectionFlowCoordinator coordinates multiple habit collection flows for session management
type CollectionFlowCoordinator struct {
	factory     *HabitCollectionFlowFactory
	activeFlows map[string]HabitCollectionFlow
}

// NewCollectionFlowCoordinator creates a new collection flow coordinator
func NewCollectionFlowCoordinator(factory *HabitCollectionFlowFactory) *CollectionFlowCoordinator {
	return &CollectionFlowCoordinator{
		factory:     factory,
		activeFlows: make(map[string]HabitCollectionFlow),
	}
}

// GetOrCreateFlow gets an existing flow or creates a new one for the habit type
func (c *CollectionFlowCoordinator) GetOrCreateFlow(habitType string) (HabitCollectionFlow, error) {
	// Check if we already have an active flow for this habit type
	if flow, exists := c.activeFlows[habitType]; exists {
		return flow, nil
	}

	// Create new flow
	flow, err := c.factory.CreateFlow(habitType)
	if err != nil {
		return nil, err
	}

	// Cache the flow for reuse
	c.activeFlows[habitType] = flow

	return flow, nil
}

// ValidateSessionHabits validates all habits in a session for flow compatibility
func (c *CollectionFlowCoordinator) ValidateSessionHabits(habits []models.Habit) []error {
	var errors []error

	for _, habit := range habits {
		if err := c.factory.ValidateHabitForFlow(habit); err != nil {
			errors = append(errors, fmt.Errorf("habit %s validation failed: %w", habit.Title, err))
		}
	}

	return errors
}

// ClearActiveFlows clears all cached flows (useful for session cleanup)
func (c *CollectionFlowCoordinator) ClearActiveFlows() {
	c.activeFlows = make(map[string]HabitCollectionFlow)
}
