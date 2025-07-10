// Package init provides file initialization functionality for the iter application.
package init

import (
	"fmt"
	"os"
	"path/filepath"

	"davidlee/iter/internal/models"
	"davidlee/iter/internal/parser"
	"davidlee/iter/internal/storage"
)

// FileInitializer handles creation of sample configuration files.
type FileInitializer struct {
	goalParser   *parser.GoalParser
	entryStorage *storage.EntryStorage
}

// NewFileInitializer creates a new file initializer instance.
func NewFileInitializer() *FileInitializer {
	return &FileInitializer{
		goalParser:   parser.NewGoalParser(),
		entryStorage: storage.NewEntryStorage(),
	}
}

// EnsureConfigFiles checks if goals.yml and entries.yml exist, creating samples if missing.
func (fi *FileInitializer) EnsureConfigFiles(goalsFile, entriesFile string) error {
	// Ensure config directory exists
	configDir := filepath.Dir(goalsFile)
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check and create goals.yml if missing
	if !fileExists(goalsFile) {
		if err := fi.createSampleGoalsFile(goalsFile); err != nil {
			return fmt.Errorf("failed to create sample goals file: %w", err)
		}
		fmt.Printf("📝 Created sample goals file: %s\n", goalsFile)
	}

	// Check and create entries.yml if missing
	if !fileExists(entriesFile) {
		if err := fi.createEmptyEntriesFile(entriesFile); err != nil {
			return fmt.Errorf("failed to create entries file: %w", err)
		}
		fmt.Printf("📋 Created empty entries file: %s\n", entriesFile)
	}

	return nil
}

// createSampleGoalsFile creates a goals.yml file with sample boolean goals.
func (fi *FileInitializer) createSampleGoalsFile(goalsFile string) error {
	schema := &models.Schema{
		Version: "1.0.0",
		Goals: []models.Goal{
			{
				Title:       "Morning Exercise",
				Position:    1,
				Description: "Get your body moving with at least 10 minutes of exercise",
				GoalType:    models.SimpleGoal,
				FieldType:   models.FieldType{Type: models.BooleanFieldType},
				ScoringType: models.ManualScoring,
				Prompt:      "Did you exercise this morning?",
				HelpText:    "Any movement counts - stretching, walking, gym, sports, etc.",
			},
			{
				Title:       "Daily Reading",
				Position:    2,
				Description: "Read for at least 15 minutes to expand your knowledge",
				GoalType:    models.SimpleGoal,
				FieldType:   models.FieldType{Type: models.BooleanFieldType},
				ScoringType: models.ManualScoring,
				Prompt:      "Did you read for at least 15 minutes today?",
				HelpText:    "Books, articles, blogs - anything that teaches you something new",
			},
		},
	}

	return fi.goalParser.SaveToFile(schema, goalsFile)
}

// createEmptyEntriesFile creates an entries.yml file with proper structure.
func (fi *FileInitializer) createEmptyEntriesFile(entriesFile string) error {
	entryLog := models.CreateEmptyEntryLog()
	return fi.entryStorage.SaveToFile(entryLog, entriesFile)
}

// fileExists checks if a file exists and is not a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
