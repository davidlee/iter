// Package init provides file initialization functionality for the vice application.
package init

import (
	"fmt"
	"os"
	"path/filepath"

	"davidlee/vice/internal/config"
	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
	"davidlee/vice/internal/storage"
)

// FileInitializer handles creation of sample configuration files.
type FileInitializer struct {
	habitParser            *parser.HabitParser
	entryStorage           *storage.EntryStorage
	checklistParser        *parser.ChecklistParser
	checklistEntriesParser *parser.ChecklistEntriesParser
}

// NewFileInitializer creates a new file initializer instance.
func NewFileInitializer() *FileInitializer {
	return &FileInitializer{
		habitParser:            parser.NewHabitParser(),
		entryStorage:           storage.NewEntryStorage(),
		checklistParser:        parser.NewChecklistParser(),
		checklistEntriesParser: parser.NewChecklistEntriesParser(),
	}
}

// EnsureConfigFiles checks if habits.yml and entries.yml exist, creating samples if missing.
func (fi *FileInitializer) EnsureConfigFiles(habitsFile, entriesFile string) error {
	// Ensure config directory exists
	configDir := filepath.Dir(habitsFile)
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check and create habits.yml if missing
	if !fileExists(habitsFile) {
		if err := fi.createSampleHabitsFile(habitsFile); err != nil {
			return fmt.Errorf("failed to create sample habits file: %w", err)
		}
		fmt.Printf("📝 Created sample habits file: %s\n", habitsFile)
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

// EnsureContextFiles checks if all context data files exist, creating samples if missing.
// AIDEV-NOTE: T028-context-files; context-aware file initialization replacing hardcoded paths
func (fi *FileInitializer) EnsureContextFiles(env *config.ViceEnv) error {
	// Ensure context data directory exists
	if err := env.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to ensure context directories: %w", err)
	}

	// Initialize all 4 data files for the context
	if err := fi.ensureHabitsFile(env.GetHabitsFile()); err != nil {
		return fmt.Errorf("failed to ensure habits file: %w", err)
	}

	if err := fi.ensureEntriesFile(env.GetEntriesFile()); err != nil {
		return fmt.Errorf("failed to ensure entries file: %w", err)
	}

	if err := fi.ensureChecklistsFile(env.GetChecklistsFile()); err != nil {
		return fmt.Errorf("failed to ensure checklists file: %w", err)
	}

	if err := fi.ensureChecklistEntriesFile(env.GetChecklistEntriesFile()); err != nil {
		return fmt.Errorf("failed to ensure checklist entries file: %w", err)
	}

	return nil
}

// ensureHabitsFile creates habits.yml if missing.
func (fi *FileInitializer) ensureHabitsFile(habitsFile string) error {
	if !fileExists(habitsFile) {
		if err := fi.createSampleHabitsFile(habitsFile); err != nil {
			return fmt.Errorf("failed to create sample habits file: %w", err)
		}
		fmt.Printf("📝 Created sample habits file: %s\n", habitsFile)
	}
	return nil
}

// ensureEntriesFile creates entries.yml if missing.
func (fi *FileInitializer) ensureEntriesFile(entriesFile string) error {
	if !fileExists(entriesFile) {
		if err := fi.createEmptyEntriesFile(entriesFile); err != nil {
			return fmt.Errorf("failed to create entries file: %w", err)
		}
		fmt.Printf("📋 Created empty entries file: %s\n", entriesFile)
	}
	return nil
}

// ensureChecklistsFile creates checklists.yml if missing.
func (fi *FileInitializer) ensureChecklistsFile(checklistsFile string) error {
	if !fileExists(checklistsFile) {
		if err := fi.createEmptyChecklistsFile(checklistsFile); err != nil {
			return fmt.Errorf("failed to create checklists file: %w", err)
		}
		fmt.Printf("📋 Created empty checklists file: %s\n", checklistsFile)
	}
	return nil
}

// ensureChecklistEntriesFile creates checklist_entries.yml if missing.
func (fi *FileInitializer) ensureChecklistEntriesFile(checklistEntriesFile string) error {
	if !fileExists(checklistEntriesFile) {
		if err := fi.createEmptyChecklistEntriesFile(checklistEntriesFile); err != nil {
			return fmt.Errorf("failed to create checklist entries file: %w", err)
		}
		fmt.Printf("📋 Created empty checklist entries file: %s\n", checklistEntriesFile)
	}
	return nil
}

// createSampleHabitsFile creates a habits.yml file with sample habits (simple and elastic).
func (fi *FileInitializer) createSampleHabitsFile(habitsFile string) error {
	schema := &models.Schema{
		Version: "1.0.0",
		Habits: []models.Habit{
			{
				Title:       "Morning Exercise",
				Position:    1,
				Description: "Get your body moving with at least 10 minutes of exercise",
				HabitType:   models.SimpleHabit,
				FieldType:   models.FieldType{Type: models.BooleanFieldType},
				ScoringType: models.ManualScoring,
				Prompt:      "Did you exercise this morning?",
				HelpText:    "Any movement counts - stretching, walking, gym, sports, etc.",
			},
			{
				Title:       "Daily Reading",
				Position:    2,
				Description: "Read for at least 15 minutes to expand your knowledge",
				HabitType:   models.SimpleHabit,
				FieldType:   models.FieldType{Type: models.BooleanFieldType},
				ScoringType: models.ManualScoring,
				Prompt:      "Did you read for at least 15 minutes today?",
				HelpText:    "Books, articles, blogs - anything that teaches you something new",
			},
			{
				Title:       "Exercise Duration",
				Position:    3,
				Description: "Track your exercise time with mini/midi/maxi achievement levels",
				HabitType:   models.ElasticHabit,
				FieldType:   models.FieldType{Type: models.DurationFieldType},
				ScoringType: models.AutomaticScoring,
				Prompt:      "How long did you exercise today?",
				HelpText:    "Enter duration like: 30m, 1h15m, or 1:30:00",
				MiniCriteria: &models.Criteria{
					Condition: &models.Condition{
						GreaterThanOrEqual: floatPtr(15), // 15 minutes
					},
				},
				MidiCriteria: &models.Criteria{
					Condition: &models.Condition{
						GreaterThanOrEqual: floatPtr(30), // 30 minutes
					},
				},
				MaxiCriteria: &models.Criteria{
					Condition: &models.Condition{
						GreaterThanOrEqual: floatPtr(60), // 60 minutes
					},
				},
			},
			{
				Title:       "Water Intake",
				Position:    4,
				Description: "Track daily water consumption in glasses",
				HabitType:   models.ElasticHabit,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType, Unit: "glasses"},
				ScoringType: models.AutomaticScoring,
				Prompt:      "How many glasses of water did you drink?",
				HelpText:    "Count 8oz glasses or equivalent",
				MiniCriteria: &models.Criteria{
					Condition: &models.Condition{
						GreaterThanOrEqual: floatPtr(4), // 4 glasses minimum
					},
				},
				MidiCriteria: &models.Criteria{
					Condition: &models.Condition{
						GreaterThanOrEqual: floatPtr(6), // 6 glasses target
					},
				},
				MaxiCriteria: &models.Criteria{
					Condition: &models.Condition{
						GreaterThanOrEqual: floatPtr(8), // 8 glasses optimal
					},
				},
			},
		},
	}

	return fi.habitParser.SaveToFile(schema, habitsFile)
}

// floatPtr returns a pointer to a float64 value.
func floatPtr(f float64) *float64 {
	return &f
}

// createEmptyEntriesFile creates an entries.yml file with proper structure.
func (fi *FileInitializer) createEmptyEntriesFile(entriesFile string) error {
	entryLog := models.CreateEmptyEntryLog()
	return fi.entryStorage.SaveToFile(entryLog, entriesFile)
}

// createEmptyChecklistsFile creates a checklists.yml file with empty schema.
func (fi *FileInitializer) createEmptyChecklistsFile(checklistsFile string) error {
	schema := &models.ChecklistSchema{
		Version:     "1.0.0",
		CreatedDate: "",
		Checklists:  []models.Checklist{},
	}
	return fi.checklistParser.SaveToFile(schema, checklistsFile)
}

// createEmptyChecklistEntriesFile creates a checklist_entries.yml file with empty schema.
func (fi *FileInitializer) createEmptyChecklistEntriesFile(checklistEntriesFile string) error {
	schema := fi.checklistEntriesParser.CreateEmptySchema()
	return fi.checklistEntriesParser.SaveToFile(schema, checklistEntriesFile)
}

// fileExists checks if a file exists and is not a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
