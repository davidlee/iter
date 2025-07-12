// Package parser provides functionality for parsing and loading checklist schemas.
package parser

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"

	"davidlee/iter/internal/models"
)

// ChecklistParser handles parsing and validation of checklist schemas.
type ChecklistParser struct{}

// NewChecklistParser creates a new checklist parser instance.
func NewChecklistParser() *ChecklistParser {
	return &ChecklistParser{}
}

// LoadFromFile loads and parses a checklists.yml file from the given path.
// It returns the parsed schema or an error if parsing or validation fails.
func (cp *ChecklistParser) LoadFromFile(filePath string) (*models.ChecklistSchema, error) {
	return cp.LoadFromFileWithIDPersistence(filePath, true)
}

// LoadFromFileWithIDPersistence loads and parses a checklists.yml file with optional ID persistence.
// If persistIDs is true and checklist IDs are generated during validation, the file is updated
// with the generated IDs to maintain data integrity.
func (cp *ChecklistParser) LoadFromFileWithIDPersistence(filePath string, persistIDs bool) (*models.ChecklistSchema, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("checklists file not found: %s", filePath)
	}

	// Read file contents
	// #nosec G304 - filePath is provided by the application, not user input
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read checklists file %s: %w", filePath, err)
	}

	// Parse YAML with change tracking if persistence is enabled
	var schema *models.ChecklistSchema
	var wasModified bool
	if persistIDs {
		schema, wasModified, err = cp.ParseYAMLWithChangeTracking(data)
	} else {
		schema, err = cp.ParseYAML(data)
	}
	if err != nil {
		return nil, err
	}

	// If ID persistence is enabled and IDs were generated, save back to file
	if persistIDs && wasModified {
		if err := cp.saveGeneratedIDs(schema, filePath); err != nil {
			// Log the error but don't fail the load operation
			// This ensures read-only files or permission issues don't break normal usage
			fmt.Fprintf(os.Stderr, "Warning: failed to persist generated checklist IDs to %s: %v\n", filePath, err)
		}
	}

	return schema, nil
}

// ParseYAMLWithChangeTracking parses YAML data and tracks whether checklist IDs were generated.
func (cp *ChecklistParser) ParseYAMLWithChangeTracking(data []byte) (*models.ChecklistSchema, bool, error) {
	var schema models.ChecklistSchema

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
func (cp *ChecklistParser) saveGeneratedIDs(schema *models.ChecklistSchema, filePath string) error {
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
	if err := cp.SaveToFile(schema, filePath); err != nil {
		return fmt.Errorf("failed to save schema with generated IDs: %w", err)
	}

	return nil
}

// ParseYAML parses YAML data into a checklist schema and validates it.
func (cp *ChecklistParser) ParseYAML(data []byte) (*models.ChecklistSchema, error) {
	var schema models.ChecklistSchema

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

// SaveToFile saves a checklist schema to a YAML file at the given path.
// This is useful for creating initial schemas or saving modified ones.
func (cp *ChecklistParser) SaveToFile(schema *models.ChecklistSchema, filePath string) error {
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
		return fmt.Errorf("failed to write checklists file %s: %w", filePath, err)
	}

	return nil
}

// ToYAML converts a checklist schema to YAML string without writing to file.
// This is useful for dry-run operations and debugging.
func (cp *ChecklistParser) ToYAML(schema *models.ChecklistSchema) (string, error) {
	// Validate before conversion
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

// GetChecklistByID finds a checklist by ID in the schema.
func (cp *ChecklistParser) GetChecklistByID(schema *models.ChecklistSchema, id string) (*models.Checklist, error) {
	for i := range schema.Checklists {
		if schema.Checklists[i].ID == id {
			return &schema.Checklists[i], nil
		}
	}
	return nil, fmt.Errorf("checklist with ID '%s' not found", id)
}

// AddChecklist adds a new checklist to the schema.
func (cp *ChecklistParser) AddChecklist(schema *models.ChecklistSchema, checklist *models.Checklist) error {
	// Validate the checklist
	if err := checklist.Validate(); err != nil {
		return fmt.Errorf("invalid checklist: %w", err)
	}

	// Check for duplicate ID
	if _, err := cp.GetChecklistByID(schema, checklist.ID); err == nil {
		return fmt.Errorf("checklist with ID '%s' already exists", checklist.ID)
	}

	// Add the checklist
	schema.Checklists = append(schema.Checklists, *checklist)

	// Validate the updated schema
	if err := schema.Validate(); err != nil {
		// Remove the checklist we just added if validation fails
		schema.Checklists = schema.Checklists[:len(schema.Checklists)-1]
		return fmt.Errorf("schema validation failed after adding checklist: %w", err)
	}

	return nil
}

// UpdateChecklist updates an existing checklist in the schema.
func (cp *ChecklistParser) UpdateChecklist(schema *models.ChecklistSchema, checklist *models.Checklist) error {
	// Validate the checklist
	if err := checklist.Validate(); err != nil {
		return fmt.Errorf("invalid checklist: %w", err)
	}

	// Find the checklist to update
	found := false
	for i := range schema.Checklists {
		if schema.Checklists[i].ID == checklist.ID {
			schema.Checklists[i] = *checklist
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("checklist with ID '%s' not found", checklist.ID)
	}

	// Validate the updated schema
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("schema validation failed after updating checklist: %w", err)
	}

	return nil
}

// RemoveChecklist removes a checklist from the schema by ID.
func (cp *ChecklistParser) RemoveChecklist(schema *models.ChecklistSchema, id string) error {
	// Find and remove the checklist
	found := false
	for i := range schema.Checklists {
		if schema.Checklists[i].ID == id {
			// Remove by slicing
			schema.Checklists = append(schema.Checklists[:i], schema.Checklists[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("checklist with ID '%s' not found", id)
	}

	// Validate the updated schema
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("schema validation failed after removing checklist: %w", err)
	}

	return nil
}
