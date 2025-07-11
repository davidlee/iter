package goalconfig

import (
	"davidlee/iter/internal/models"
	"davidlee/iter/internal/parser"
)

// GoalConfigurator provides UI for managing goal configurations
type GoalConfigurator struct {
	goalParser *parser.GoalParser
}

// NewGoalConfigurator creates a new goal configurator instance
func NewGoalConfigurator() *GoalConfigurator {
	return &GoalConfigurator{
		goalParser: parser.NewGoalParser(),
	}
}

// AddGoal presents an interactive UI to create a new goal
func (gc *GoalConfigurator) AddGoal(goalsFilePath string) error {
	// TODO: Phase 2 - Implement interactive goal creation UI
	return nil
}

// ListGoals displays all existing goals in a formatted view
func (gc *GoalConfigurator) ListGoals(goalsFilePath string) error {
	// TODO: Phase 3 - Implement goal listing UI
	return nil
}

// EditGoal presents an interactive UI to modify an existing goal
func (gc *GoalConfigurator) EditGoal(goalsFilePath string) error {
	// TODO: Phase 4 - Implement goal editing UI
	return nil
}

// RemoveGoal presents an interactive UI to remove an existing goal
func (gc *GoalConfigurator) RemoveGoal(goalsFilePath string) error {
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