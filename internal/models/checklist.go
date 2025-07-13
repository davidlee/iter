// Package models contains checklist data structures for the iter application.
package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ChecklistSchema represents the top-level checklists.yml file structure.
type ChecklistSchema struct {
	Version     string      `yaml:"version"`
	CreatedDate string      `yaml:"created_date"`
	Checklists  []Checklist `yaml:"checklists"`
}

// Checklist represents a reusable checklist template.
// Items are stored as simple strings, with headings prefixed by "# ".
type Checklist struct {
	ID           string   `yaml:"id"`
	Title        string   `yaml:"title"`
	Description  string   `yaml:"description,omitempty"`
	Items        []string `yaml:"items"` // Simple array of strings
	CreatedDate  string   `yaml:"created_date"`
	ModifiedDate string   `yaml:"modified_date"`
}

// ChecklistCompletion stores completion state for entry data.
// This represents the state of a checklist at a specific point in time.
type ChecklistCompletion struct {
	ChecklistID     string          `yaml:"checklist_id"`
	CompletedItems  map[string]bool `yaml:"completed_items"` // item text -> completed
	CompletionTime  string          `yaml:"completion_time,omitempty"`
	PartialComplete bool            `yaml:"partial_complete"`
}

// Validate validates a checklist schema for correctness and consistency.
func (cs *ChecklistSchema) Validate() error {
	// Version is required
	if cs.Version == "" {
		return fmt.Errorf("checklist schema version is required")
	}

	// Created date should be valid if provided
	if cs.CreatedDate != "" {
		if _, err := time.Parse("2006-01-02", cs.CreatedDate); err != nil {
			return fmt.Errorf("invalid created_date format, expected YYYY-MM-DD: %w", err)
		}
	}

	// Track unique checklist IDs
	ids := make(map[string]bool)

	// Validate each checklist
	for i := range cs.Checklists {
		if err := cs.Checklists[i].Validate(); err != nil {
			return fmt.Errorf("checklist at index %d: %w", i, err)
		}

		// Check ID uniqueness
		if ids[cs.Checklists[i].ID] {
			return fmt.Errorf("duplicate checklist ID: %s", cs.Checklists[i].ID)
		}
		ids[cs.Checklists[i].ID] = true
	}

	return nil
}

// ChecklistEntriesSchema represents the checklist_entries.yml file structure
// for tracking daily checklist completion state.
// AIDEV-NOTE: data-separation; templates vs instances pattern (see T007 Phase 3)
type ChecklistEntriesSchema struct {
	Version string                  `yaml:"version"`
	Entries map[string]DailyEntries `yaml:"entries"` // date -> checklist entries
}

// DailyEntries holds all checklist completions for a specific date.
type DailyEntries struct {
	Date      string                    `yaml:"date"`
	Completed map[string]ChecklistEntry `yaml:"completed"` // checklist_id -> completion state
}

// ChecklistEntry stores the completion state for a single checklist on a specific date.
// AIDEV-NOTE: completion-tracking; maps item text to bool for historical accuracy
type ChecklistEntry struct {
	ChecklistID     string          `yaml:"checklist_id"`
	CompletedItems  map[string]bool `yaml:"completed_items"` // item text -> completed status
	CompletionTime  string          `yaml:"completion_time,omitempty"`
	PartialComplete bool            `yaml:"partial_complete"`
}

// Validate validates a checklist entries schema.
func (ces *ChecklistEntriesSchema) Validate() error {
	if ces.Version == "" {
		return fmt.Errorf("schema version is required")
	}

	// Validate each date entry
	for date, dailyEntries := range ces.Entries {
		if dailyEntries.Date != date {
			return fmt.Errorf("date mismatch: key '%s' does not match entry date '%s'", date, dailyEntries.Date)
		}

		// Validate date format (YYYY-MM-DD)
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return fmt.Errorf("invalid date format '%s': expected YYYY-MM-DD", date)
		}

		// Validate checklist entries for this date
		for checklistID, entry := range dailyEntries.Completed {
			if entry.ChecklistID != checklistID {
				return fmt.Errorf("checklist ID mismatch: key '%s' does not match entry ID '%s'", checklistID, entry.ChecklistID)
			}

			if err := entry.Validate(); err != nil {
				return fmt.Errorf("invalid checklist entry for '%s' on %s: %w", checklistID, date, err)
			}
		}
	}

	return nil
}

// Validate validates a single checklist entry.
func (ce *ChecklistEntry) Validate() error {
	if strings.TrimSpace(ce.ChecklistID) == "" {
		return fmt.Errorf("checklist ID is required")
	}

	// Validate completion time if provided
	if ce.CompletionTime != "" {
		if _, err := time.Parse(time.RFC3339, ce.CompletionTime); err != nil {
			return fmt.Errorf("invalid completion time format: %w", err)
		}
	}

	return nil
}

// ValidateAndTrackChanges validates a checklist schema and returns whether it was modified.
// Returns (wasModified, error) where wasModified indicates if any checklist IDs were generated.
func (cs *ChecklistSchema) ValidateAndTrackChanges() (bool, error) {
	// Version is required
	if cs.Version == "" {
		return false, fmt.Errorf("checklist schema version is required")
	}

	// Created date should be valid if provided
	if cs.CreatedDate != "" {
		if _, err := time.Parse("2006-01-02", cs.CreatedDate); err != nil {
			return false, fmt.Errorf("invalid created_date format, expected YYYY-MM-DD: %w", err)
		}
	}

	// Track unique checklist IDs and modifications
	ids := make(map[string]bool)
	wasModified := false

	// Validate each checklist
	for i := range cs.Checklists {
		checklistModified, err := cs.Checklists[i].ValidateAndTrackChanges()
		if err != nil {
			return false, fmt.Errorf("checklist at index %d: %w", i, err)
		}
		if checklistModified {
			wasModified = true
		}

		// Check ID uniqueness
		if ids[cs.Checklists[i].ID] {
			return false, fmt.Errorf("duplicate checklist ID: %s", cs.Checklists[i].ID)
		}
		ids[cs.Checklists[i].ID] = true
	}

	return wasModified, nil
}

// Validate validates a checklist for correctness and consistency.
func (c *Checklist) Validate() error {
	// Title is required
	if strings.TrimSpace(c.Title) == "" {
		return fmt.Errorf("checklist title is required")
	}

	// Generate ID if not provided
	if c.ID == "" {
		c.ID = generateChecklistIDFromTitle(c.Title)
	}

	return c.validateInternal()
}

// ValidateAndTrackChanges validates a checklist and returns whether it was modified.
// Returns (wasModified, error) where wasModified indicates if ID was generated.
func (c *Checklist) ValidateAndTrackChanges() (bool, error) {
	// Title is required
	if strings.TrimSpace(c.Title) == "" {
		return false, fmt.Errorf("checklist title is required")
	}

	// Check if ID needs to be generated
	wasModified := false
	if c.ID == "" {
		c.ID = generateChecklistIDFromTitle(c.Title)
		wasModified = true
	}

	return wasModified, c.validateInternal()
}

// validateInternal performs the core checklist validation logic.
func (c *Checklist) validateInternal() error {
	// Validate ID format
	if !isValidChecklistID(c.ID) {
		return fmt.Errorf("checklist ID '%s' is invalid: must contain only letters, numbers, and underscores", c.ID)
	}

	// At least one item is required
	if len(c.Items) == 0 {
		return fmt.Errorf("checklist must contain at least one item")
	}

	// Validate each item (simple string validation)
	for i, item := range c.Items {
		if strings.TrimSpace(item) == "" {
			return fmt.Errorf("item at index %d: item text cannot be empty", i)
		}
	}

	// Validate dates if provided
	if c.CreatedDate != "" {
		if _, err := time.Parse("2006-01-02", c.CreatedDate); err != nil {
			return fmt.Errorf("invalid created_date format, expected YYYY-MM-DD: %w", err)
		}
	}

	if c.ModifiedDate != "" {
		if _, err := time.Parse("2006-01-02", c.ModifiedDate); err != nil {
			return fmt.Errorf("invalid modified_date format, expected YYYY-MM-DD: %w", err)
		}
	}

	return nil
}

// Validate validates a checklist completion condition.
func (ccc *ChecklistCompletionCondition) Validate() error {
	// RequiredItems is required
	if ccc.RequiredItems == "" {
		return fmt.Errorf("required_items field is required")
	}

	// Validate RequiredItems value
	if ccc.RequiredItems != "all" {
		return fmt.Errorf("required_items must be 'all', got: %s", ccc.RequiredItems)
	}

	return nil
}

// Validate validates checklist completion state.
func (cc *ChecklistCompletion) Validate() error {
	// ChecklistID is required
	if cc.ChecklistID == "" {
		return fmt.Errorf("checklist_id is required")
	}

	// Validate completion time format if provided
	if cc.CompletionTime != "" {
		if _, err := time.Parse(time.RFC3339, cc.CompletionTime); err != nil {
			return fmt.Errorf("invalid completion_time format, expected RFC3339: %w", err)
		}
	}

	// CompletedItems can be empty (represents no items completed)
	// No additional validation needed for the map itself

	return nil
}

// GetTotalItemCount returns the total number of items (not headings) in the checklist.
func (c *Checklist) GetTotalItemCount() int {
	count := 0
	for _, item := range c.Items {
		if !strings.HasPrefix(item, "# ") {
			count++
		}
	}
	return count
}

// GetCompletedItemCount returns the number of completed items (not headings).
func (cc *ChecklistCompletion) GetCompletedItemCount(checklist *Checklist) int {
	count := 0
	for _, item := range checklist.Items {
		if !strings.HasPrefix(item, "# ") && cc.CompletedItems[item] {
			count++
		}
	}
	return count
}

// GetCompletedTotalCount returns the total number of completed items.
func (cc *ChecklistCompletion) GetCompletedTotalCount() int {
	count := 0
	for _, completed := range cc.CompletedItems {
		if completed {
			count++
		}
	}
	return count
}

// IsComplete checks if the checklist completion meets the specified condition.
func (cc *ChecklistCompletion) IsComplete(checklist *Checklist, _ *ChecklistCompletionCondition) bool {
	// Only "all" criteria is supported - all items must be completed
	totalItems := checklist.GetTotalItemCount()
	completedItems := cc.GetCompletedItemCount(checklist)
	return completedItems >= totalItems
}

// generateChecklistIDFromTitle creates a valid ID from a checklist title.
func generateChecklistIDFromTitle(title string) string {
	// Convert to lowercase
	id := strings.ToLower(title)

	// Replace spaces and special characters with underscores
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	id = reg.ReplaceAllString(id, "_")

	// Remove consecutive underscores
	reg = regexp.MustCompile(`_+`)
	id = reg.ReplaceAllString(id, "_")

	// Trim leading/trailing underscores
	id = strings.Trim(id, "_")

	// Ensure it's not empty
	if id == "" {
		id = "unnamed_checklist"
	}

	return id
}

// isValidChecklistID checks if an ID contains only valid characters.
func isValidChecklistID(id string) bool {
	if id == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-z0-9_]+$`, id)
	return matched
}
