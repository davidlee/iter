package parser

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/davidlee/vice/internal/models"
)

// ChecklistEntriesParser handles loading and saving checklist_entries.yml.
// AIDEV-NOTE: persistence-layer; daily completion state separate from templates
type ChecklistEntriesParser struct{}

// NewChecklistEntriesParser creates a new checklist entries parser.
func NewChecklistEntriesParser() *ChecklistEntriesParser {
	return &ChecklistEntriesParser{}
}

// LoadFromFile loads checklist entries from a YAML file.
func (cep *ChecklistEntriesParser) LoadFromFile(filePath string) (*models.ChecklistEntriesSchema, error) {
	//nolint:gosec // File path is controlled by application configuration
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var schema models.ChecklistEntriesSchema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := schema.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &schema, nil
}

// SaveToFile saves checklist entries to a YAML file.
func (cep *ChecklistEntriesParser) SaveToFile(schema *models.ChecklistEntriesSchema, filePath string) error {
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	data, err := yaml.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	//nolint:gosec // File permissions 0o644 appropriate for configuration files
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GetChecklistEntryForDate retrieves the checklist entry for a specific date and checklist ID.
// Returns nil if no entry exists for that date/checklist combination.
// AIDEV-NOTE: state-lookup; enables same-day resume functionality
func (cep *ChecklistEntriesParser) GetChecklistEntryForDate(schema *models.ChecklistEntriesSchema, date, checklistID string) *models.ChecklistEntry {
	dailyEntries, exists := schema.Entries[date]
	if !exists {
		return nil
	}

	entry, exists := dailyEntries.Completed[checklistID]
	if !exists {
		return nil
	}

	return &entry
}

// SaveChecklistEntryForDate saves or updates a checklist entry for a specific date and checklist ID.
func (cep *ChecklistEntriesParser) SaveChecklistEntryForDate(schema *models.ChecklistEntriesSchema, date, checklistID string, entry models.ChecklistEntry) error {
	// Initialize entries map if needed
	if schema.Entries == nil {
		schema.Entries = make(map[string]models.DailyEntries)
	}

	// Get or create daily entries for this date
	dailyEntries, exists := schema.Entries[date]
	if !exists {
		dailyEntries = models.DailyEntries{
			Date:      date,
			Completed: make(map[string]models.ChecklistEntry),
		}
	}

	// Ensure completed map is initialized
	if dailyEntries.Completed == nil {
		dailyEntries.Completed = make(map[string]models.ChecklistEntry)
	}

	// Save the entry
	dailyEntries.Completed[checklistID] = entry
	schema.Entries[date] = dailyEntries

	return nil
}

// CreateEmptySchema creates a new empty checklist entries schema.
func (cep *ChecklistEntriesParser) CreateEmptySchema() *models.ChecklistEntriesSchema {
	return &models.ChecklistEntriesSchema{
		Version: "1.0.0",
		Entries: make(map[string]models.DailyEntries),
	}
}

// EnsureSchemaExists loads the schema from file or creates a new empty one if the file doesn't exist.
func (cep *ChecklistEntriesParser) EnsureSchemaExists(filePath string) (*models.ChecklistEntriesSchema, error) {
	// Try to load existing file
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, create empty schema
		return cep.CreateEmptySchema(), nil
	}

	// File exists, load it
	return cep.LoadFromFile(filePath)
}

// GetTodaysDate returns today's date in YYYY-MM-DD format.
func (cep *ChecklistEntriesParser) GetTodaysDate() string {
	return time.Now().Format("2006-01-02")
}
