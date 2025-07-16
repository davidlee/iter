package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// EntryLog represents the top-level structure for all daily entries.
type EntryLog struct {
	Version string     `yaml:"version"`
	Entries []DayEntry `yaml:"entries"`
}

// DayEntry represents all habit completions for a single day.
type DayEntry struct {
	Date   string       `yaml:"date"`   // ISO date format: YYYY-MM-DD
	Habits []HabitEntry `yaml:"habits"` // Habit completions for this day
}

// AchievementLevel represents the achievement level for elastic habits.
type AchievementLevel string

// Achievement levels for elastic habits.
const (
	AchievementNone AchievementLevel = "none" // No achievement level met
	AchievementMini AchievementLevel = "mini" // Minimum achievement level
	AchievementMidi AchievementLevel = "midi" // Medium achievement level
	AchievementMaxi AchievementLevel = "maxi" // Maximum achievement level
)

// EntryStatus represents the completion status of a habit entry.
type EntryStatus string

// Entry status values define the state of habit completion.
const (
	EntryCompleted EntryStatus = "completed" // Habit successfully completed
	EntrySkipped   EntryStatus = "skipped"   // Habit skipped due to circumstances
	EntryFailed    EntryStatus = "failed"    // Habit attempted but not achieved
)

// HabitEntry represents the completion data for a single habit on a specific day.
type HabitEntry struct {
	HabitID          string            `yaml:"habit_id"`
	Value            interface{}       `yaml:"value,omitempty"`             // nil for skipped entries
	AchievementLevel *AchievementLevel `yaml:"achievement_level,omitempty"` // For elastic habits
	Notes            string            `yaml:"notes,omitempty"`
	CreatedAt        time.Time         `yaml:"created_at"`           // Entry creation time
	UpdatedAt        *time.Time        `yaml:"updated_at,omitempty"` // Last modification time (nil if never updated)
	Status           EntryStatus       `yaml:"status"`               // Entry completion status
}

// BooleanEntry is a convenience type for boolean habit entries.
type BooleanEntry struct {
	HabitID   string `yaml:"habit_id"`
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

	// Track unique habit IDs for this day
	habitIDs := make(map[string]bool)

	// Validate each habit entry
	for i, habitEntry := range de.Habits {
		if err := habitEntry.Validate(); err != nil {
			return fmt.Errorf("habit entry at index %d: %w", i, err)
		}

		// Check habit ID uniqueness within this day
		if habitIDs[habitEntry.HabitID] {
			return fmt.Errorf("duplicate habit ID for date %s: %s", de.Date, habitEntry.HabitID)
		}
		habitIDs[habitEntry.HabitID] = true
	}

	return nil
}

// IsSkipped returns true if this habit entry was skipped.
func (ge *HabitEntry) IsSkipped() bool {
	return ge.Status == EntrySkipped
}

// IsCompleted returns true if this habit entry was completed successfully.
func (ge *HabitEntry) IsCompleted() bool {
	return ge.Status == EntryCompleted
}

// HasFailure returns true if this habit entry failed.
func (ge *HabitEntry) HasFailure() bool {
	return ge.Status == EntryFailed
}

// IsFinalized returns true if this habit entry has been processed (has status).
func (ge *HabitEntry) IsFinalized() bool {
	return ge.Status != ""
}

// RequiresValue returns true if this habit entry should have a value.
func (ge *HabitEntry) RequiresValue() bool {
	return ge.Status != EntrySkipped
}

// MarkCreated sets the CreatedAt timestamp to the current time.
func (ge *HabitEntry) MarkCreated() {
	ge.CreatedAt = time.Now()
}

// MarkUpdated sets the UpdatedAt timestamp to the current time.
func (ge *HabitEntry) MarkUpdated() {
	now := time.Now()
	ge.UpdatedAt = &now
}

// GetLastModified returns the most recent modification time.
// Returns UpdatedAt if set, otherwise CreatedAt.
func (ge *HabitEntry) GetLastModified() time.Time {
	if ge.UpdatedAt != nil {
		return *ge.UpdatedAt
	}
	return ge.CreatedAt
}

// Validate validates a habit entry for correctness.
func (ge *HabitEntry) Validate() error {
	// Habit ID is required
	if strings.TrimSpace(ge.HabitID) == "" {
		return fmt.Errorf("habit ID is required")
	}

	// Status is required
	if ge.Status == "" {
		return fmt.Errorf("entry status is required")
	}

	// Validate status value
	if !isValidEntryStatus(ge.Status) {
		return fmt.Errorf("invalid entry status: %s", ge.Status)
	}

	// AIDEV-NOTE: Permissive validation per ADR-001 - allow dormant achievement levels on skipped entries
	// Status-based validation
	switch ge.Status {
	case EntrySkipped:
		// Skipped entries should not have values (but may preserve achievement levels)
		if ge.Value != nil {
			return fmt.Errorf("skipped entries cannot have values")
		}
		// Achievement levels allowed for data preservation (see ADR-001)
	case EntryCompleted, EntryFailed:
		// Completed and failed entries must have values
		if ge.Value == nil {
			return fmt.Errorf("completed and failed entries must have values")
		}
	default:
		return fmt.Errorf("unknown entry status: %s", ge.Status)
	}

	// Validate achievement level if present
	if ge.AchievementLevel != nil {
		if !isValidAchievementLevel(*ge.AchievementLevel) {
			return fmt.Errorf("invalid achievement level: %s", *ge.AchievementLevel)
		}
	}

	// Validate timestamps
	if ge.CreatedAt.IsZero() {
		return fmt.Errorf("created_at timestamp is required")
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

// GetHabitEntry finds a habit entry within this day. Returns the entry and true if found.
func (de *DayEntry) GetHabitEntry(habitID string) (*HabitEntry, bool) {
	for i := range de.Habits {
		if de.Habits[i].HabitID == habitID {
			return &de.Habits[i], true
		}
	}
	return nil, false
}

// AddHabitEntry adds a habit entry to this day. If an entry for this habit
// already exists, it returns an error.
func (de *DayEntry) AddHabitEntry(habitEntry HabitEntry) error {
	// Validate the habit entry
	if err := habitEntry.Validate(); err != nil {
		return fmt.Errorf("invalid habit entry: %w", err)
	}

	// Check if entry for this habit already exists
	if _, exists := de.GetHabitEntry(habitEntry.HabitID); exists {
		return fmt.Errorf("entry for habit %s already exists on date %s", habitEntry.HabitID, de.Date)
	}

	// Add the entry
	de.Habits = append(de.Habits, habitEntry)

	return nil
}

// UpdateHabitEntry updates an existing habit entry or creates a new one if it doesn't exist.
func (de *DayEntry) UpdateHabitEntry(habitEntry HabitEntry) error {
	// Validate the habit entry
	if err := habitEntry.Validate(); err != nil {
		return fmt.Errorf("invalid habit entry: %w", err)
	}

	// Find existing entry
	for i := range de.Habits {
		if de.Habits[i].HabitID == habitEntry.HabitID {
			de.Habits[i] = habitEntry
			return nil
		}
	}

	// Entry doesn't exist, add it
	de.Habits = append(de.Habits, habitEntry)
	return nil
}

// GetBooleanValue safely extracts a boolean value from the habit entry.
// Returns the boolean value and true if successful, false and false if not a boolean.
func (ge *HabitEntry) GetBooleanValue() (bool, bool) {
	if boolVal, ok := ge.Value.(bool); ok {
		return boolVal, true
	}
	return false, false
}

// SetBooleanValue sets the habit entry value to a boolean.
func (ge *HabitEntry) SetBooleanValue(value bool) {
	ge.Value = value
}

// GetAchievementLevel returns the achievement level for this habit entry.
// Returns the level and true if set, or AchievementNone and false if not set.
func (ge *HabitEntry) GetAchievementLevel() (AchievementLevel, bool) {
	if ge.AchievementLevel != nil {
		return *ge.AchievementLevel, true
	}
	return AchievementNone, false
}

// SetAchievementLevel sets the achievement level for this habit entry.
func (ge *HabitEntry) SetAchievementLevel(level AchievementLevel) {
	ge.AchievementLevel = &level
}

// HasAchievementLevel returns true if this habit entry has an achievement level set.
func (ge *HabitEntry) HasAchievementLevel() bool {
	return ge.AchievementLevel != nil
}

// ClearAchievementLevel removes the achievement level from this habit entry.
func (ge *HabitEntry) ClearAchievementLevel() {
	ge.AchievementLevel = nil
}

// CreateTodayEntry creates a new day entry for today's date.
func CreateTodayEntry() DayEntry {
	return DayEntry{
		Date:   time.Now().Format("2006-01-02"),
		Habits: []HabitEntry{},
	}
}

// CreateBooleanHabitEntry creates a new habit entry for a boolean habit.
func CreateBooleanHabitEntry(habitID string, completed bool) HabitEntry {
	entry := HabitEntry{
		HabitID: habitID,
		Value:   completed,
	}
	if completed {
		entry.Status = EntryCompleted
	} else {
		entry.Status = EntryFailed
	}
	entry.MarkCreated()
	return entry
}

// CreateElasticHabitEntry creates a new habit entry for an elastic habit with achievement level.
func CreateElasticHabitEntry(habitID string, value interface{}, level AchievementLevel) HabitEntry {
	entry := HabitEntry{
		HabitID:          habitID,
		Value:            value,
		AchievementLevel: &level,
	}
	if level == AchievementNone {
		entry.Status = EntryFailed
	} else {
		entry.Status = EntryCompleted
	}
	entry.MarkCreated()
	return entry
}

// CreateValueOnlyHabitEntry creates a new habit entry with just a value (no achievement level).
func CreateValueOnlyHabitEntry(habitID string, value interface{}) HabitEntry {
	entry := HabitEntry{
		HabitID: habitID,
		Value:   value,
		Status:  EntryCompleted,
	}
	entry.MarkCreated()
	return entry
}

// CreateSkippedHabitEntry creates a new habit entry that was skipped.
func CreateSkippedHabitEntry(habitID string) HabitEntry {
	entry := HabitEntry{
		HabitID: habitID,
		Status:  EntrySkipped,
	}
	entry.MarkCreated()
	return entry
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

// isValidAchievementLevel checks if an achievement level is valid.
func isValidAchievementLevel(level AchievementLevel) bool {
	switch level {
	case AchievementNone, AchievementMini, AchievementMidi, AchievementMaxi:
		return true
	default:
		return false
	}
}

// isValidEntryStatus checks if an entry status is valid.
func isValidEntryStatus(status EntryStatus) bool {
	switch status {
	case EntryCompleted, EntrySkipped, EntryFailed:
		return true
	default:
		return false
	}
}

// IsValidAchievementLevel checks if a string represents a valid achievement level.
func IsValidAchievementLevel(level string) bool {
	return isValidAchievementLevel(AchievementLevel(level))
}

// MarshalYAML implements custom YAML marshaling for HabitEntry to format timestamps in human-readable format
// AIDEV-NOTE: T020 human-readable time storage; custom YAML marshaling for timestamps and time values with permissive parsing
func (ge *HabitEntry) MarshalYAML() (interface{}, error) {
	// Create a temporary struct with the same fields but different time formatting
	type habitEntryAlias HabitEntry

	// Convert to alias to avoid infinite recursion, then create a map for custom formatting
	alias := (*habitEntryAlias)(ge)

	// Create a map to control field ordering and custom formatting
	result := make(map[string]interface{})
	result["habit_id"] = alias.HabitID

	if alias.Value != nil {
		// Handle time field values specially - format as HH:MM if it's a time
		if timeVal, ok := alias.Value.(time.Time); ok && isTimeFieldValue(timeVal) {
			result["value"] = timeVal.Format("15:04")
		} else if floatVal, ok := alias.Value.(float64); ok {
			// Preserve float64 type by ensuring decimal notation
			if floatVal == float64(int64(floatVal)) {
				// If it's a whole number, add .0 to preserve float type
				result["value"] = fmt.Sprintf("%.1f", floatVal)
			} else {
				result["value"] = floatVal
			}
		} else {
			result["value"] = alias.Value
		}
	}

	if alias.AchievementLevel != nil {
		result["achievement_level"] = *alias.AchievementLevel
	}

	if alias.Notes != "" {
		result["notes"] = alias.Notes
	}

	// Format timestamps in human-readable format
	result["created_at"] = alias.CreatedAt.Format("2006-01-02 15:04:05")

	if alias.UpdatedAt != nil {
		result["updated_at"] = alias.UpdatedAt.Format("2006-01-02 15:04:05")
	}

	result["status"] = alias.Status

	return result, nil
}

// UnmarshalYAML implements permissive YAML unmarshaling for HabitEntry to accept various time formats
func (ge *HabitEntry) UnmarshalYAML(node *yaml.Node) error {
	// Create a temporary struct for standard unmarshaling
	type habitEntryAlias HabitEntry
	alias := (*habitEntryAlias)(ge)

	// First unmarshal into a map to handle custom field processing
	var raw map[string]interface{}
	if err := node.Decode(&raw); err != nil {
		return err
	}

	// Process each field
	if habitID, ok := raw["habit_id"].(string); ok {
		alias.HabitID = habitID
	}

	// Handle value field - could be time, string, number, boolean
	if rawValue, exists := raw["value"]; exists {
		alias.Value = parseValueField(rawValue)
	}

	if achievementLevel, exists := raw["achievement_level"]; exists {
		if levelStr, ok := achievementLevel.(string); ok {
			level := AchievementLevel(levelStr)
			alias.AchievementLevel = &level
		}
	}

	if notes, ok := raw["notes"].(string); ok {
		alias.Notes = notes
	}

	// Parse created_at with permissive time parsing
	if createdAt, exists := raw["created_at"]; exists {
		if parsedTime, err := parseTimeFlexible(createdAt); err == nil {
			alias.CreatedAt = parsedTime
		} else {
			return fmt.Errorf("invalid created_at format: %v", err)
		}
	}

	// Parse updated_at with permissive time parsing
	if updatedAt, exists := raw["updated_at"]; exists {
		if parsedTime, err := parseTimeFlexible(updatedAt); err == nil {
			alias.UpdatedAt = &parsedTime
		} else {
			return fmt.Errorf("invalid updated_at format: %v", err)
		}
	}

	if status, ok := raw["status"].(string); ok {
		alias.Status = EntryStatus(status)
	}

	return nil
}

// isTimeFieldValue determines if a time.Time value represents a time-of-day rather than a full timestamp
func isTimeFieldValue(t time.Time) bool {
	// Time field values have year 0000 (zero date with time component)
	return t.Year() == 0
}

// parseValueField handles parsing the value field which could be various types including time strings
func parseValueField(raw interface{}) interface{} {
	if str, ok := raw.(string); ok {
		// Try parsing as time first (for time field values)
		if strings.Contains(str, ":") {
			if parsedTime, err := parseTimePermissive(str); err == nil {
				return parsedTime
			}
		}
		// Try parsing as float if it looks like a decimal number
		if strings.Contains(str, ".") {
			if floatVal, err := strconv.ParseFloat(str, 64); err == nil {
				return floatVal
			}
		}
		return str
	}
	// Preserve numeric types as-is to maintain float64 vs int distinction
	return raw
}

// parseTimeFlexible accepts both string and time.Time inputs and returns time.Time
func parseTimeFlexible(raw interface{}) (time.Time, error) {
	// If it's already a time.Time (from YAML auto-conversion), use it directly
	if timeVal, ok := raw.(time.Time); ok {
		return timeVal, nil
	}

	// Otherwise parse as string
	return parseTimePermissive(raw)
}

// MarshalYAML implements custom YAML marshaling for DayEntry to ensure HabitEntry custom marshaling is applied
func (de *DayEntry) MarshalYAML() (interface{}, error) {
	// Create a map to control field ordering and ensure custom marshaling of habits
	result := make(map[string]interface{})
	result["date"] = de.Date

	// Marshal habits individually to ensure custom marshaling is applied
	if len(de.Habits) > 0 {
		var habits []interface{}
		for _, habit := range de.Habits {
			habitYAML, err := habit.MarshalYAML()
			if err != nil {
				return nil, fmt.Errorf("failed to marshal habit %s: %w", habit.HabitID, err)
			}
			habits = append(habits, habitYAML)
		}
		result["habits"] = habits
	} else {
		result["habits"] = []interface{}{}
	}

	return result, nil
}

// MarshalYAML implements custom YAML marshaling for EntryLog to ensure DayEntry custom marshaling is applied
func (el *EntryLog) MarshalYAML() (interface{}, error) {
	// Create a map to control field ordering and ensure custom marshaling of entries
	result := make(map[string]interface{})
	result["version"] = el.Version

	// Marshal entries individually to ensure custom marshaling is applied
	if len(el.Entries) > 0 {
		var entries []interface{}
		for _, entry := range el.Entries {
			entryYAML, err := entry.MarshalYAML()
			if err != nil {
				return nil, fmt.Errorf("failed to marshal day entry %s: %w", entry.Date, err)
			}
			entries = append(entries, entryYAML)
		}
		result["entries"] = entries
	} else {
		result["entries"] = []interface{}{}
	}

	return result, nil
}

// parseTimePermissive accepts various time format inputs and returns time.Time
func parseTimePermissive(raw interface{}) (time.Time, error) {
	str, ok := raw.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("time value must be string, got %T", raw)
	}

	// List of formats to try, in order of preference
	formats := []string{
		// Human-readable formats (new)
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"15:04:05",
		"15:04",
		// RFC3339 formats (existing)
		time.RFC3339,
		time.RFC3339Nano,
		// ISO 8601 variants
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05.000000000Z",
		// Legacy formats
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.000-07:00",
	}

	// Try each format
	for _, format := range formats {
		if parsed, err := time.Parse(format, str); err == nil {
			return parsed, nil
		}
	}

	// Try parsing as Unix timestamp
	if unix, err := strconv.ParseInt(str, 10, 64); err == nil {
		return time.Unix(unix, 0), nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time value: %s", str)
}
