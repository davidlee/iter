// Package parser provides functionality for parsing and loading habit schemas.
package parser

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"

	"davidlee/vice/internal/models"
)

// HabitParser handles parsing and validation of habit schemas.
type HabitParser struct{}

// NewHabitParser creates a new habit parser instance.
func NewHabitParser() *HabitParser {
	return &HabitParser{}
}

// LoadFromFile loads and parses a habits.yml file from the given path.
// It returns the parsed schema or an error if parsing or validation fails.
func (gp *HabitParser) LoadFromFile(filePath string) (*models.Schema, error) {
	return gp.LoadFromFileWithIDPersistence(filePath, true)
}

// LoadFromFileWithIDPersistence loads and parses a habits.yml file with optional ID persistence.
// If persistIDs is true and habit IDs are generated during validation, the file is updated
// with the generated IDs to maintain data integrity.
func (gp *HabitParser) LoadFromFileWithIDPersistence(filePath string, persistIDs bool) (*models.Schema, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("habits file not found: %s", filePath)
	}

	// Read file contents
	// #nosec G304 - filePath is provided by the application, not user input
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read habits file %s: %w", filePath, err)
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
			fmt.Fprintf(os.Stderr, "Warning: failed to persist generated habit IDs to %s: %v\n", filePath, err)
		}
	}

	return schema, nil
}

// ParseYAMLWithChangeTracking parses YAML data and tracks whether habit IDs were generated.
func (gp *HabitParser) ParseYAMLWithChangeTracking(data []byte) (*models.Schema, bool, error) {
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
func (gp *HabitParser) saveGeneratedIDs(schema *models.Schema, filePath string) error {
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

// ParseYAML parses YAML data into a habit schema and validates it.
func (gp *HabitParser) ParseYAML(data []byte) (*models.Schema, error) {
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
func (gp *HabitParser) SaveToFile(schema *models.Schema, filePath string) error {
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
		return fmt.Errorf("failed to write habits file %s: %w", filePath, err)
	}

	return nil
}

// ToYAML converts a schema to YAML string without writing to file.
// This is useful for dry-run operations and debugging.
func (gp *HabitParser) ToYAML(schema *models.Schema) (string, error) {
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

// CreateSampleSchema creates a sample schema with simple boolean habits.
// This is useful for initializing new configurations.
func (gp *HabitParser) CreateSampleSchema() *models.Schema {
	return &models.Schema{
		Version:     "1.0.0",
		CreatedDate: "2024-01-01",
		Habits: []models.Habit{
			{
				Title:       "Morning Meditation",
				Position:    1,
				Description: "Daily mindfulness practice to start the day centered",
				HabitType:   models.SimpleHabit,
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
				HabitType:   models.SimpleHabit,
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
				HabitType:   models.SimpleHabit,
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

// ValidateFile checks if a habits.yml file is valid without fully loading it.
// Returns validation errors if any are found.
func (gp *HabitParser) ValidateFile(filePath string) error {
	_, err := gp.LoadFromFile(filePath)
	return err
}

// GetHabitByID finds a habit by its ID in the schema.
// Returns the habit and true if found, nil and false otherwise.
func GetHabitByID(schema *models.Schema, habitID string) (*models.Habit, bool) {
	if schema == nil {
		return nil, false
	}

	for i := range schema.Habits {
		if schema.Habits[i].ID == habitID {
			return &schema.Habits[i], true
		}
	}

	return nil, false
}

// GetHabitsByType returns all habits of a specific type from the schema.
func GetHabitsByType(schema *models.Schema, habitType models.HabitType) []models.Habit {
	if schema == nil {
		return nil
	}

	var habits []models.Habit
	for _, habit := range schema.Habits {
		if habit.HabitType == habitType {
			habits = append(habits, habit)
		}
	}

	return habits
}

// GetSimpleBooleanHabits returns all simple boolean habits from the schema.
// This is a convenience function for the MVP implementation.
func GetSimpleBooleanHabits(schema *models.Schema) []models.Habit {
	if schema == nil {
		return nil
	}

	var habits []models.Habit
	for _, habit := range schema.Habits {
		if habit.HabitType == models.SimpleHabit && habit.FieldType.Type == models.BooleanFieldType {
			habits = append(habits, habit)
		}
	}

	return habits
}
