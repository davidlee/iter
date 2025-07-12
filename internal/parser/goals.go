// Package parser provides functionality for parsing and loading goal schemas.
package parser

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"

	"davidlee/iter/internal/models"
)

// GoalParser handles parsing and validation of goal schemas.
type GoalParser struct{}

// NewGoalParser creates a new goal parser instance.
func NewGoalParser() *GoalParser {
	return &GoalParser{}
}

// LoadFromFile loads and parses a goals.yml file from the given path.
// It returns the parsed schema or an error if parsing or validation fails.
func (gp *GoalParser) LoadFromFile(filePath string) (*models.Schema, error) {
	return gp.LoadFromFileWithIDPersistence(filePath, true)
}

// LoadFromFileWithIDPersistence loads and parses a goals.yml file with optional ID persistence.
// If persistIDs is true and goal IDs are generated during validation, the file is updated
// with the generated IDs to maintain data integrity.
func (gp *GoalParser) LoadFromFileWithIDPersistence(filePath string, persistIDs bool) (*models.Schema, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("goals file not found: %s", filePath)
	}

	// Read file contents
	// #nosec G304 - filePath is provided by the application, not user input
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read goals file %s: %w", filePath, err)
	}

	// Parse YAML with change tracking if persistence is enabled
	var schema *models.Schema
	var wasModified bool
	if persistIDs {
		schema, wasModified, err = gp.ParseYAMLWithChangeTracking(data)
	} else {
		schema, err = gp.ParseYAML(data)
	}
	if err != nil {
		return nil, err
	}

	// If ID persistence is enabled and IDs were generated, save back to file
	if persistIDs && wasModified {
		if err := gp.saveGeneratedIDs(schema, filePath); err != nil {
			// Log the error but don't fail the load operation
			// This ensures read-only files or permission issues don't break normal usage
			fmt.Fprintf(os.Stderr, "Warning: failed to persist generated goal IDs to %s: %v\n", filePath, err)
		}
	}

	return schema, nil
}

// ParseYAMLWithChangeTracking parses YAML data and tracks whether goal IDs were generated.
func (gp *GoalParser) ParseYAMLWithChangeTracking(data []byte) (*models.Schema, bool, error) {
	var schema models.Schema

	// Parse YAML with strict mode to catch unknown fields
	if err := yaml.UnmarshalWithOptions(data, &schema, yaml.Strict()); err != nil {
		return nil, false, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate the parsed schema with change tracking
	wasModified, err := schema.ValidateAndTrackChanges()
	if err != nil {
		return nil, false, fmt.Errorf("schema validation failed: %w", err)
	}

	return &schema, wasModified, nil
}

// saveGeneratedIDs saves the schema with generated IDs back to the file.
func (gp *GoalParser) saveGeneratedIDs(schema *models.Schema, filePath string) error {
	// Check if file is writable before attempting to save
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to check file permissions: %w", err)
	}

	// Check if file is read-only
	if fileInfo.Mode()&0o200 == 0 {
		return fmt.Errorf("file is read-only, cannot persist generated IDs")
	}

	// Save the updated schema back to the file
	if err := gp.SaveToFile(schema, filePath); err != nil {
		return fmt.Errorf("failed to save schema with generated IDs: %w", err)
	}

	return nil
}

// ParseYAML parses YAML data into a goal schema and validates it.
func (gp *GoalParser) ParseYAML(data []byte) (*models.Schema, error) {
	var schema models.Schema

	// Parse YAML with strict mode to catch unknown fields
	if err := yaml.UnmarshalWithOptions(data, &schema, yaml.Strict()); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate the parsed schema
	if err := schema.Validate(); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}

	return &schema, nil
}

// SaveToFile saves a schema to a YAML file at the given path.
// This is useful for creating initial schemas or saving modified ones.
func (gp *GoalParser) SaveToFile(schema *models.Schema, filePath string) error {
	// Validate before saving
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid schema: %w", err)
	}

	// Marshal to YAML with pretty formatting
	data, err := yaml.MarshalWithOptions(schema,
		yaml.Indent(2),
		yaml.IndentSequence(true),
	)
	if err != nil {
		return fmt.Errorf("failed to marshal schema to YAML: %w", err)
	}

	// Write to file with appropriate permissions (0600 for security)
	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write goals file %s: %w", filePath, err)
	}

	return nil
}

// ToYAML converts a schema to YAML string without writing to file.
// This is useful for dry-run operations and debugging.
func (gp *GoalParser) ToYAML(schema *models.Schema) (string, error) {
	// Validate before converting
	if err := schema.Validate(); err != nil {
		return "", fmt.Errorf("cannot convert invalid schema to YAML: %w", err)
	}

	// Marshal to YAML with pretty formatting
	data, err := yaml.MarshalWithOptions(schema,
		yaml.Indent(2),
		yaml.IndentSequence(true),
	)
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema to YAML: %w", err)
	}

	return string(data), nil
}

// CreateSampleSchema creates a sample schema with simple boolean goals.
// This is useful for initializing new configurations.
func (gp *GoalParser) CreateSampleSchema() *models.Schema {
	return &models.Schema{
		Version:     "1.0.0",
		CreatedDate: "2024-01-01",
		Goals: []models.Goal{
			{
				Title:       "Morning Meditation",
				Position:    1,
				Description: "Daily mindfulness practice to start the day centered",
				GoalType:    models.SimpleGoal,
				FieldType: models.FieldType{
					Type: models.BooleanFieldType,
				},
				ScoringType: models.ManualScoring,
				Prompt:      "Did you meditate this morning?",
				HelpText:    "Even 5 minutes counts!",
			},
			{
				Title:       "Daily Exercise",
				Position:    2,
				Description: "Physical activity to maintain health and energy",
				GoalType:    models.SimpleGoal,
				FieldType: models.FieldType{
					Type: models.BooleanFieldType,
				},
				ScoringType: models.ManualScoring,
				Prompt:      "Did you exercise today?",
				HelpText:    "Any movement counts - walking, gym, sports, yoga, etc.",
			},
			{
				Title:       "Read for 30 Minutes",
				Position:    3,
				Description: "Daily reading for learning and personal growth",
				GoalType:    models.SimpleGoal,
				FieldType: models.FieldType{
					Type: models.BooleanFieldType,
				},
				ScoringType: models.ManualScoring,
				Prompt:      "Did you read for at least 30 minutes?",
				HelpText:    "Books, articles, or educational content",
			},
		},
	}
}

// ValidateFile checks if a goals.yml file is valid without fully loading it.
// Returns validation errors if any are found.
func (gp *GoalParser) ValidateFile(filePath string) error {
	_, err := gp.LoadFromFile(filePath)
	return err
}

// GetGoalByID finds a goal by its ID in the schema.
// Returns the goal and true if found, nil and false otherwise.
func GetGoalByID(schema *models.Schema, goalID string) (*models.Goal, bool) {
	if schema == nil {
		return nil, false
	}

	for i := range schema.Goals {
		if schema.Goals[i].ID == goalID {
			return &schema.Goals[i], true
		}
	}

	return nil, false
}

// GetGoalsByType returns all goals of a specific type from the schema.
func GetGoalsByType(schema *models.Schema, goalType models.GoalType) []models.Goal {
	if schema == nil {
		return nil
	}

	var goals []models.Goal
	for _, goal := range schema.Goals {
		if goal.GoalType == goalType {
			goals = append(goals, goal)
		}
	}

	return goals
}

// GetSimpleBooleanGoals returns all simple boolean goals from the schema.
// This is a convenience function for the MVP implementation.
func GetSimpleBooleanGoals(schema *models.Schema) []models.Goal {
	if schema == nil {
		return nil
	}

	var goals []models.Goal
	for _, goal := range schema.Goals {
		if goal.GoalType == models.SimpleGoal && goal.FieldType.Type == models.BooleanFieldType {
			goals = append(goals, goal)
		}
	}

	return goals
}
