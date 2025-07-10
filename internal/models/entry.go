package models

import (
	"fmt"
	"strings"
	"time"
)

// EntryLog represents the top-level structure for all daily entries.
type EntryLog struct {
	Version string     `yaml:"version"`
	Entries []DayEntry `yaml:"entries"`
}

// DayEntry represents all goal completions for a single day.
type DayEntry struct {
	Date  string      `yaml:"date"`  // ISO date format: YYYY-MM-DD
	Goals []GoalEntry `yaml:"goals"` // Goal completions for this day
}

// GoalEntry represents the completion data for a single goal on a specific day.
type GoalEntry struct {
	GoalID string      `yaml:"goal_id"`
	Value  interface{} `yaml:"value"`
	// Future fields for automatic scoring results, timestamps, etc.
	CompletedAt *time.Time `yaml:"completed_at,omitempty"`
	Notes       string     `yaml:"notes,omitempty"`
}

// BooleanEntry is a convenience type for boolean goal entries.
type BooleanEntry struct {
	GoalID    string `yaml:"goal_id"`
	Completed bool   `yaml:"completed"`
}

// Validate validates an entry log for correctness and consistency.
func (el *EntryLog) Validate() error {
	// Version is required
	if el.Version == "" {
		return fmt.Errorf("entry log version is required")
	}

	// Track unique dates
	dates := make(map[string]bool)

	// Validate each day entry
	for i, dayEntry := range el.Entries {
		if err := dayEntry.Validate(); err != nil {
			return fmt.Errorf("day entry at index %d: %w", i, err)
		}

		// Check date uniqueness
		if dates[dayEntry.Date] {
			return fmt.Errorf("duplicate date: %s", dayEntry.Date)
		}
		dates[dayEntry.Date] = true
	}

	return nil
}

// Validate validates a day entry for correctness.
func (de *DayEntry) Validate() error {
	// Date is required and must be valid
	if de.Date == "" {
		return fmt.Errorf("date is required")
	}

	// Validate date format
	if _, err := time.Parse("2006-01-02", de.Date); err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Track unique goal IDs for this day
	goalIDs := make(map[string]bool)

	// Validate each goal entry
	for i, goalEntry := range de.Goals {
		if err := goalEntry.Validate(); err != nil {
			return fmt.Errorf("goal entry at index %d: %w", i, err)
		}

		// Check goal ID uniqueness within this day
		if goalIDs[goalEntry.GoalID] {
			return fmt.Errorf("duplicate goal ID for date %s: %s", de.Date, goalEntry.GoalID)
		}
		goalIDs[goalEntry.GoalID] = true
	}

	return nil
}

// Validate validates a goal entry for correctness.
func (ge *GoalEntry) Validate() error {
	// Goal ID is required
	if strings.TrimSpace(ge.GoalID) == "" {
		return fmt.Errorf("goal ID is required")
	}

	// Value is required (can be false for booleans, but not nil)
	if ge.Value == nil {
		return fmt.Errorf("goal value is required")
	}

	return nil
}

// GetDayEntry finds a day entry by date. Returns the entry and true if found.
func (el *EntryLog) GetDayEntry(date string) (*DayEntry, bool) {
	for i := range el.Entries {
		if el.Entries[i].Date == date {
			return &el.Entries[i], true
		}
	}
	return nil, false
}

// AddDayEntry adds a new day entry to the log. If an entry for this date
// already exists, it returns an error.
func (el *EntryLog) AddDayEntry(dayEntry DayEntry) error {
	// Validate the day entry
	if err := dayEntry.Validate(); err != nil {
		return fmt.Errorf("invalid day entry: %w", err)
	}

	// Check if entry for this date already exists
	if _, exists := el.GetDayEntry(dayEntry.Date); exists {
		return fmt.Errorf("entry for date %s already exists", dayEntry.Date)
	}

	// Add the entry
	el.Entries = append(el.Entries, dayEntry)

	return nil
}

// UpdateDayEntry updates an existing day entry or creates a new one if it doesn't exist.
func (el *EntryLog) UpdateDayEntry(dayEntry DayEntry) error {
	// Validate the day entry
	if err := dayEntry.Validate(); err != nil {
		return fmt.Errorf("invalid day entry: %w", err)
	}

	// Find existing entry
	for i := range el.Entries {
		if el.Entries[i].Date == dayEntry.Date {
			el.Entries[i] = dayEntry
			return nil
		}
	}

	// Entry doesn't exist, add it
	el.Entries = append(el.Entries, dayEntry)
	return nil
}

// GetGoalEntry finds a goal entry within this day. Returns the entry and true if found.
func (de *DayEntry) GetGoalEntry(goalID string) (*GoalEntry, bool) {
	for i := range de.Goals {
		if de.Goals[i].GoalID == goalID {
			return &de.Goals[i], true
		}
	}
	return nil, false
}

// AddGoalEntry adds a goal entry to this day. If an entry for this goal
// already exists, it returns an error.
func (de *DayEntry) AddGoalEntry(goalEntry GoalEntry) error {
	// Validate the goal entry
	if err := goalEntry.Validate(); err != nil {
		return fmt.Errorf("invalid goal entry: %w", err)
	}

	// Check if entry for this goal already exists
	if _, exists := de.GetGoalEntry(goalEntry.GoalID); exists {
		return fmt.Errorf("entry for goal %s already exists on date %s", goalEntry.GoalID, de.Date)
	}

	// Add the entry
	de.Goals = append(de.Goals, goalEntry)

	return nil
}

// UpdateGoalEntry updates an existing goal entry or creates a new one if it doesn't exist.
func (de *DayEntry) UpdateGoalEntry(goalEntry GoalEntry) error {
	// Validate the goal entry
	if err := goalEntry.Validate(); err != nil {
		return fmt.Errorf("invalid goal entry: %w", err)
	}

	// Find existing entry
	for i := range de.Goals {
		if de.Goals[i].GoalID == goalEntry.GoalID {
			de.Goals[i] = goalEntry
			return nil
		}
	}

	// Entry doesn't exist, add it
	de.Goals = append(de.Goals, goalEntry)
	return nil
}

// GetBooleanValue safely extracts a boolean value from the goal entry.
// Returns the boolean value and true if successful, false and false if not a boolean.
func (ge *GoalEntry) GetBooleanValue() (bool, bool) {
	if boolVal, ok := ge.Value.(bool); ok {
		return boolVal, true
	}
	return false, false
}

// SetBooleanValue sets the goal entry value to a boolean.
func (ge *GoalEntry) SetBooleanValue(value bool) {
	ge.Value = value
}

// CreateTodayEntry creates a new day entry for today's date.
func CreateTodayEntry() DayEntry {
	return DayEntry{
		Date:  time.Now().Format("2006-01-02"),
		Goals: []GoalEntry{},
	}
}

// CreateBooleanGoalEntry creates a new goal entry for a boolean goal.
func CreateBooleanGoalEntry(goalID string, completed bool) GoalEntry {
	return GoalEntry{
		GoalID: goalID,
		Value:  completed,
	}
}

// IsToday checks if this day entry is for today's date.
func (de *DayEntry) IsToday() bool {
	today := time.Now().Format("2006-01-02")
	return de.Date == today
}

// GetDate parses the date string into a time.Time.
func (de *DayEntry) GetDate() (time.Time, error) {
	return time.Parse("2006-01-02", de.Date)
}

// CreateEmptyEntryLog creates a new empty entry log with the current version.
func CreateEmptyEntryLog() *EntryLog {
	return &EntryLog{
		Version: "1.0.0",
		Entries: []DayEntry{},
	}
}

// GetEntriesForDateRange returns all day entries within the specified date range (inclusive).
func (el *EntryLog) GetEntriesForDateRange(startDate, endDate string) ([]DayEntry, error) {
	// Parse dates for comparison
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %w", err)
	}

	if start.After(end) {
		return nil, fmt.Errorf("start date %s is after end date %s", startDate, endDate)
	}

	var result []DayEntry
	for _, entry := range el.Entries {
		entryDate, err := time.Parse("2006-01-02", entry.Date)
		if err != nil {
			continue // Skip invalid dates
		}

		if (entryDate.Equal(start) || entryDate.After(start)) &&
			(entryDate.Equal(end) || entryDate.Before(end)) {
			result = append(result, entry)
		}
	}

	return result, nil
}
