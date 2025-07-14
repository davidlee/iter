package entry

import (
	"fmt"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/scoring"
)

// AIDEV-NOTE: flow-factory; creates appropriate goal collection flows with field input and scoring integration
// Central factory for coordinating goal type-specific collection flows with T010/1.2 field input components

// GoalCollectionFlowFactory creates appropriate collection flows for different goal types
type GoalCollectionFlowFactory struct {
	fieldInputFactory *EntryFieldInputFactory
	scoringEngine     *scoring.Engine
	checklistsPath    string
}

// NewGoalCollectionFlowFactory creates a new goal collection flow factory
func NewGoalCollectionFlowFactory(fieldInputFactory *EntryFieldInputFactory, scoringEngine *scoring.Engine, checklistsPath string) *GoalCollectionFlowFactory {
	return &GoalCollectionFlowFactory{
		fieldInputFactory: fieldInputFactory,
		scoringEngine:     scoringEngine,
		checklistsPath:    checklistsPath,
	}
}

// CreateFlow creates the appropriate collection flow for a given goal type
func (f *GoalCollectionFlowFactory) CreateFlow(goalType string) (GoalCollectionFlow, error) {
	switch goalType {
	case string(models.SimpleGoal):
		return NewSimpleGoalCollectionFlow(f.fieldInputFactory, f.scoringEngine), nil

	case string(models.ElasticGoal):
		return NewElasticGoalCollectionFlow(f.fieldInputFactory, f.scoringEngine), nil

	case string(models.InformationalGoal):
		return NewInformationalGoalCollectionFlow(f.fieldInputFactory), nil

	case string(models.ChecklistGoal):
		return NewChecklistGoalCollectionFlow(f.fieldInputFactory, f.scoringEngine, f.checklistsPath), nil

	default:
		return nil, fmt.Errorf("unsupported goal type: %s", goalType)
	}
}

// ValidateGoalForFlow validates that a goal is compatible with its intended flow
func (f *GoalCollectionFlowFactory) ValidateGoalForFlow(goal models.Goal) error {
	flow, err := f.CreateFlow(string(goal.GoalType))
	if err != nil {
		return fmt.Errorf("unable to create flow for goal type %s: %w", goal.GoalType, err)
	}

	// Check if the goal's field type is supported by the flow
	supportedTypes := flow.GetExpectedFieldTypes()
	fieldTypeSupported := false
	for _, supportedType := range supportedTypes {
		if goal.FieldType.Type == supportedType {
			fieldTypeSupported = true
			break
		}
	}

	if !fieldTypeSupported {
		return fmt.Errorf("field type %s not supported by %s goal flow", goal.FieldType.Type, goal.GoalType)
	}

	// Check if scoring requirements are met
	if flow.RequiresScoring() && goal.ScoringType == models.AutomaticScoring && f.scoringEngine == nil {
		return fmt.Errorf("automatic scoring required for %s goal but no scoring engine available", goal.GoalType)
	}

	return nil
}

// GetSupportedGoalTypes returns the list of goal types supported by the factory
func (f *GoalCollectionFlowFactory) GetSupportedGoalTypes() []string {
	return []string{
		string(models.SimpleGoal),
		string(models.ElasticGoal),
		string(models.InformationalGoal),
		string(models.ChecklistGoal),
	}
}

// GetFlowInfo returns information about a specific goal type flow
func (f *GoalCollectionFlowFactory) GetFlowInfo(goalType string) (*FlowInfo, error) {
	flow, err := f.CreateFlow(goalType)
	if err != nil {
		return nil, err
	}

	return &FlowInfo{
		GoalType:            goalType,
		RequiresScoring:     flow.RequiresScoring(),
		SupportedFieldTypes: flow.GetExpectedFieldTypes(),
		Description:         f.getFlowDescription(goalType),
	}, nil
}

// FlowInfo provides information about a goal collection flow
type FlowInfo struct {
	GoalType            string
	RequiresScoring     bool
	SupportedFieldTypes []string
	Description         string
}

// Private helper methods

func (f *GoalCollectionFlowFactory) getFlowDescription(goalType string) string {
	switch goalType {
	case string(models.SimpleGoal):
		return "Pass/fail collection with optional additional data and automatic/manual scoring support"
	case string(models.ElasticGoal):
		return "Data collection with immediate mini/midi/maxi achievement calculation and feedback"
	case string(models.InformationalGoal):
		return "Data-only collection without pass/fail evaluation, for tracking purposes"
	case string(models.ChecklistGoal):
		return "Interactive checklist completion with progress tracking and achievement scoring"
	default:
		return "Unknown goal type"
	}
}

// CollectionFlowCoordinator coordinates multiple goal collection flows for session management
type CollectionFlowCoordinator struct {
	factory     *GoalCollectionFlowFactory
	activeFlows map[string]GoalCollectionFlow
}

// NewCollectionFlowCoordinator creates a new collection flow coordinator
func NewCollectionFlowCoordinator(factory *GoalCollectionFlowFactory) *CollectionFlowCoordinator {
	return &CollectionFlowCoordinator{
		factory:     factory,
		activeFlows: make(map[string]GoalCollectionFlow),
	}
}

// GetOrCreateFlow gets an existing flow or creates a new one for the goal type
func (c *CollectionFlowCoordinator) GetOrCreateFlow(goalType string) (GoalCollectionFlow, error) {
	// Check if we already have an active flow for this goal type
	if flow, exists := c.activeFlows[goalType]; exists {
		return flow, nil
	}

	// Create new flow
	flow, err := c.factory.CreateFlow(goalType)
	if err != nil {
		return nil, err
	}

	// Cache the flow for reuse
	c.activeFlows[goalType] = flow

	return flow, nil
}

// ValidateSessionGoals validates all goals in a session for flow compatibility
func (c *CollectionFlowCoordinator) ValidateSessionGoals(goals []models.Goal) []error {
	var errors []error

	for _, goal := range goals {
		if err := c.factory.ValidateGoalForFlow(goal); err != nil {
			errors = append(errors, fmt.Errorf("goal %s validation failed: %w", goal.Title, err))
		}
	}

	return errors
}

// ClearActiveFlows clears all cached flows (useful for session cleanup)
func (c *CollectionFlowCoordinator) ClearActiveFlows() {
	c.activeFlows = make(map[string]GoalCollectionFlow)
}
