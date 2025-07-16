// Package storage provides functionality for persisting and loading entry data.
// AIDEV-NOTE: T021 entries-storage-patterns; centralized file operations with atomic writes and YAML marshaling
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"davidlee/vice/internal/models"
)

// BackupConfig defines backup behavior for entries storage
// AIDEV-NOTE: T021 config-stub; placeholder for future user configuration system
type BackupConfig struct {
	// Enabled controls whether automatic backups are created
	Enabled bool
	// CreateBeforeWrite creates backup before any write operation
	CreateBeforeWrite bool
}

// DefaultBackupConfig returns the default backup configuration
func DefaultBackupConfig() BackupConfig {
	return BackupConfig{
		Enabled:           true, // Default to enabled for safety
		CreateBeforeWrite: true, // Backup before each write
	}
}

// EntryStorage handles the persistent storage of entry logs.
// AIDEV-NOTE: T021 storage-with-config; future enhancement to include BackupConfig field
type EntryStorage struct{}

// NewEntryStorage creates a new entry storage instance.
func NewEntryStorage() *EntryStorage {
	return &EntryStorage{}
}

// LoadFromFile loads an entry log from the specified file path.
// If the file doesn't exist, it returns an empty entry log.
// AIDEV-NOTE: T021 file-load-pattern; graceful handling of missing files, strict YAML parsing
func (es *EntryStorage) LoadFromFile(filePath string) (*models.EntryLog, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, return empty entry log
		return models.CreateEmptyEntryLog(), nil
	}

	// Read file contents
	// #nosec G304 - filePath is provided by the application, not user input
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read entries file %s: %w", filePath, err)
	}

	// Parse YAML
	return es.ParseYAML(data)
}

// ParseYAML parses YAML data into an entry log and validates it.
// AIDEV-NOTE: T021 yaml-parsing-strategy; strict mode prevents unknown fields, validation ensures data integrity
func (es *EntryStorage) ParseYAML(data []byte) (*models.EntryLog, error) {
	var entryLog models.EntryLog

	// Parse YAML with strict mode to catch unknown fields
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true) // Strict mode
	if err := decoder.Decode(&entryLog); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate the parsed entry log
	if err := entryLog.Validate(); err != nil {
		return nil, fmt.Errorf("entry log validation failed: %w", err)
	}

	return &entryLog, nil
}

// SaveToFile saves an entry log to the specified file path with atomic writes.
// This creates a temporary file first, then renames it to prevent corruption.
// AIDEV-NOTE: T021 atomic-write-pattern; temp-file + rename ensures data consistency, validates before marshaling
func (es *EntryStorage) SaveToFile(entryLog *models.EntryLog, filePath string) error {
	// Validate before saving
	if err := entryLog.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid entry log: %w", err)
	}

	// Marshal to YAML with pretty formatting
	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(entryLog); err != nil {
		return fmt.Errorf("failed to marshal entry log to YAML: %w", err)
	}
	_ = encoder.Close()
	data := []byte(buf.String())

	// Validate marshalled data by attempting to parse it back
	// AIDEV-NOTE: T021 marshal-validation; prevents corrupted data from being written
	var testLog models.EntryLog
	if err := yaml.Unmarshal(data, &testLog); err != nil {
		return fmt.Errorf("marshalled data failed validation - would produce corrupted file: %w", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write to temporary file first for atomic operation
	tempFile := filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0o600); err != nil {
		return fmt.Errorf("failed to write temporary file %s: %w", tempFile, err)
	}

	// Atomically rename temporary file to final file
	if err := os.Rename(tempFile, filePath); err != nil {
		// Clean up temporary file on failure
		_ = os.Remove(tempFile) // Ignore error since we're already in error state
		return fmt.Errorf("failed to rename temporary file to %s: %w", filePath, err)
	}

	return nil
}

// AddDayEntry adds a day entry to the entry log file.
// This loads the existing log, adds the entry, and saves it back.
func (es *EntryStorage) AddDayEntry(filePath string, dayEntry models.DayEntry) error {
	// Load existing entry log
	entryLog, err := es.LoadFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load existing entries: %w", err)
	}

	// Add the day entry
	if err := entryLog.AddDayEntry(dayEntry); err != nil {
		return fmt.Errorf("failed to add day entry: %w", err)
	}

	// Save the updated log with automatic backup
	config := DefaultBackupConfig()
	if err := es.SaveToFileWithBackup(entryLog, filePath, config); err != nil {
		return fmt.Errorf("failed to save updated entries: %w", err)
	}

	return nil
}

// SaveToFileWithBackup saves an entry log with optional automatic backup based on configuration.
// AIDEV-NOTE: T021 resilient-save; automatic backup + validation before atomic write
func (es *EntryStorage) SaveToFileWithBackup(entryLog *models.EntryLog, filePath string, config BackupConfig) error {
	// Create automatic backup if enabled and file exists
	if config.Enabled && config.CreateBeforeWrite {
		if _, err := os.Stat(filePath); err == nil {
			// File exists, create backup
			if backupErr := es.BackupFile(filePath); backupErr != nil {
				// Log warning but don't fail - backup is best-effort
				// In a real implementation, this would use a proper logger
				fmt.Fprintf(os.Stderr, "Warning: failed to create backup before write: %v\n", backupErr)
			}
		}
	}

	// Proceed with standard atomic save (includes validation)
	return es.SaveToFile(entryLog, filePath)
}

// UpdateDayEntry updates or creates a day entry in the entry log file.
// This loads the existing log, updates the entry, and saves it back.
// AIDEV-NOTE: T021 load-modify-save-pattern; most common entry operation, full file rewrite on each save
func (es *EntryStorage) UpdateDayEntry(filePath string, dayEntry models.DayEntry) error {
	// Load existing entry log
	entryLog, err := es.LoadFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load existing entries: %w", err)
	}

	// Update the day entry
	if err := entryLog.UpdateDayEntry(dayEntry); err != nil {
		return fmt.Errorf("failed to update day entry: %w", err)
	}

	// Save the updated log with automatic backup
	// AIDEV-NOTE: T021 auto-backup-integration; uses default config for automatic backup
	config := DefaultBackupConfig()
	if err := es.SaveToFileWithBackup(entryLog, filePath, config); err != nil {
		return fmt.Errorf("failed to save updated entries: %w", err)
	}

	return nil
}

// GetDayEntry retrieves a specific day's entry from the entry log file.
func (es *EntryStorage) GetDayEntry(filePath string, date string) (*models.DayEntry, error) {
	// Load entry log
	entryLog, err := es.LoadFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load entries: %w", err)
	}

	// Find the day entry
	dayEntry, found := entryLog.GetDayEntry(date)
	if !found {
		return nil, fmt.Errorf("no entry found for date %s", date)
	}

	return dayEntry, nil
}

// GetTodayEntry retrieves today's entry from the entry log file.
func (es *EntryStorage) GetTodayEntry(filePath string) (*models.DayEntry, error) {
	today := models.CreateTodayEntry().Date
	return es.GetDayEntry(filePath, today)
}

// AddGoalEntry adds a goal entry to a specific day in the entry log file.
// If the day doesn't exist, it creates a new day entry.
func (es *EntryStorage) AddGoalEntry(filePath string, date string, goalEntry models.GoalEntry) error {
	// Load existing entry log
	entryLog, err := es.LoadFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load existing entries: %w", err)
	}

	// Find or create day entry
	dayEntry, found := entryLog.GetDayEntry(date)
	if !found {
		// Create new day entry
		newDayEntry := models.DayEntry{
			Date:  date,
			Goals: []models.GoalEntry{},
		}
		if err := entryLog.AddDayEntry(newDayEntry); err != nil {
			return fmt.Errorf("failed to create day entry for %s: %w", date, err)
		}
		dayEntry, _ = entryLog.GetDayEntry(date)
	}

	// Add the goal entry
	if err := dayEntry.AddGoalEntry(goalEntry); err != nil {
		return fmt.Errorf("failed to add goal entry: %w", err)
	}

	// Save the updated log
	if err := es.SaveToFile(entryLog, filePath); err != nil {
		return fmt.Errorf("failed to save updated entries: %w", err)
	}

	return nil
}

// UpdateGoalEntry updates or creates a goal entry for a specific day in the entry log file.
// If the day doesn't exist, it creates a new day entry.
func (es *EntryStorage) UpdateGoalEntry(filePath string, date string, goalEntry models.GoalEntry) error {
	// Load existing entry log
	entryLog, err := es.LoadFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load existing entries: %w", err)
	}

	// Find or create day entry
	dayEntry, found := entryLog.GetDayEntry(date)
	if !found {
		// Create new day entry
		newDayEntry := models.DayEntry{
			Date:  date,
			Goals: []models.GoalEntry{},
		}
		if err := entryLog.UpdateDayEntry(newDayEntry); err != nil {
			return fmt.Errorf("failed to create day entry for %s: %w", date, err)
		}
		dayEntry, _ = entryLog.GetDayEntry(date)
	}

	// Update the goal entry
	if err := dayEntry.UpdateGoalEntry(goalEntry); err != nil {
		return fmt.Errorf("failed to update goal entry: %w", err)
	}

	// Save the updated log
	if err := es.SaveToFile(entryLog, filePath); err != nil {
		return fmt.Errorf("failed to save updated entries: %w", err)
	}

	return nil
}

// UpdateTodayGoalEntry updates or creates a goal entry for today.
func (es *EntryStorage) UpdateTodayGoalEntry(filePath string, goalEntry models.GoalEntry) error {
	today := models.CreateTodayEntry().Date
	return es.UpdateGoalEntry(filePath, today, goalEntry)
}

// GetEntriesForDateRange retrieves all entries within the specified date range.
func (es *EntryStorage) GetEntriesForDateRange(filePath string, startDate, endDate string) ([]models.DayEntry, error) {
	// Load entry log
	entryLog, err := es.LoadFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load entries: %w", err)
	}

	// Get entries in range
	entries, err := entryLog.GetEntriesForDateRange(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries for date range: %w", err)
	}

	return entries, nil
}

// ValidateFile checks if an entries.yml file is valid without fully loading it.
func (es *EntryStorage) ValidateFile(filePath string) error {
	_, err := es.LoadFromFile(filePath)
	return err
}

// BackupFile creates a backup of the entries file with a timestamp.
// AIDEV-NOTE: T021 backup-strategy; simple .backup suffix, single backup file (no versioning)
func (es *EntryStorage) BackupFile(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("entries file not found: %s", filePath)
	}

	// Create backup filename with timestamp
	backupPath := filePath + ".backup"

	// Read original file
	// #nosec G304 - filePath is provided by the application, not user input
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read entries file for backup: %w", err)
	}

	// Write backup file
	if err := os.WriteFile(backupPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// CreateSampleEntryLog creates a sample entry log with some example data.
func (es *EntryStorage) CreateSampleEntryLog() *models.EntryLog {
	entryLog := models.CreateEmptyEntryLog()

	// Add a sample day entry
	meditation := models.CreateBooleanGoalEntry("morning_meditation", true)
	meditation.Notes = "Had a peaceful 10-minute session"

	exercise := models.CreateBooleanGoalEntry("daily_exercise", false)
	exercise.Notes = "Planned to go to gym but got busy with work"

	sampleDay := models.DayEntry{
		Date: "2024-01-01",
		Goals: []models.GoalEntry{
			meditation,
			exercise,
		},
	}

	entryLog.Entries = append(entryLog.Entries, sampleDay)
	return entryLog
}
