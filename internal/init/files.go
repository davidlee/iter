// Package init provides file initialization functionality for the vice application.
package init

import (
	"fmt"
	"os"
	"path/filepath"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
	"davidlee/vice/internal/storage"
)

// FileInitializer handles creation of sample configuration files.
type FileInitializer struct {
	goalParser   *parser.HabitParser
	entryStorage *storage.EntryStorage
}

// NewFileInitializer creates a new file initializer instance.
func NewFileInitializer() *FileInitializer {
	return &FileInitializer{
		goalParser:   parser.NewHabitParser(),
		entryStorage: storage.NewEntryStorage(),
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

	return fi.goalParser.SaveToFile(schema, habitsFile)
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

// fileExists checks if a file exists and is not a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
