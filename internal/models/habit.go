// Package models defines the data structures for the vice application.
package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Schema represents the top-level habit schema structure.
type Schema struct {
	Version     string  `yaml:"version"`
	CreatedDate string  `yaml:"created_date"`
	Habits      []Habit `yaml:"habits"`
}

// Habit represents a single habit in the schema.
type Habit struct {
	Title       string      `yaml:"title"`
	ID          string      `yaml:"id,omitempty"`
	Position    int         `yaml:"position"`
	Description string      `yaml:"description,omitempty"`
	HabitType   HabitType   `yaml:"habit_type"`
	FieldType   FieldType   `yaml:"field_type"`
	ScoringType ScoringType `yaml:"scoring_type,omitempty"`
	Criteria    *Criteria   `yaml:"criteria,omitempty"`

	// Elastic habit fields (not used for simple habits)
	MiniCriteria *Criteria `yaml:"mini_criteria,omitempty"`
	MidiCriteria *Criteria `yaml:"midi_criteria,omitempty"`
	MaxiCriteria *Criteria `yaml:"maxi_criteria,omitempty"`

	// Informational habit fields (not used for simple habits)
	Direction string `yaml:"direction,omitempty"`

	// UI fields
	Prompt   string `yaml:"prompt,omitempty"`
	HelpText string `yaml:"help_text,omitempty"`
}

// HabitType represents the type of habit.
type HabitType string

// Habit types define the behavior and scoring model for habits.
const (
	SimpleHabit        HabitType = "simple"        // Boolean pass/fail habits
	ElasticHabit       HabitType = "elastic"       // Three-tier achievement habits (mini/midi/maxi)
	InformationalHabit HabitType = "informational" // Data collection without scoring
	ChecklistHabit     HabitType = "checklist"     // Checklist completion habits
)

// ScoringType represents how the habit is scored.
type ScoringType string

// Scoring types define how habit achievement is determined.
const (
	ManualScoring    ScoringType = "manual"    // User manually marks completion
	AutomaticScoring ScoringType = "automatic" // Automatic scoring based on criteria
)

// FieldType represents the data type for habit values.
type FieldType struct {
	Type        string   `yaml:"type"`
	Multiline   *bool    `yaml:"multiline,omitempty"`
	Default     *bool    `yaml:"default,omitempty"`
	Unit        string   `yaml:"unit,omitempty"`
	Min         *float64 `yaml:"min,omitempty"`
	Max         *float64 `yaml:"max,omitempty"`
	Format      string   `yaml:"format,omitempty"`
	ChecklistID string   `yaml:"checklist_id,omitempty"` // Reference to checklist
}

// Field type constants
const (
	TextFieldType            = "text"
	BooleanFieldType         = "boolean"
	UnsignedIntFieldType     = "unsigned_int"
	UnsignedDecimalFieldType = "unsigned_decimal"
	DecimalFieldType         = "decimal"
	TimeFieldType            = "time"
	DurationFieldType        = "duration"
	ChecklistFieldType       = "checklist"
)

// Criteria represents habit achievement criteria.
type Criteria struct {
	Description string     `yaml:"description,omitempty"`
	Condition   *Condition `yaml:"condition"`
}

// Condition represents the logical condition for criteria evaluation.
type Condition struct {
	// Numeric/Duration comparisons
	GreaterThan        *float64 `yaml:"greater_than,omitempty"`
	GreaterThanOrEqual *float64 `yaml:"greater_than_or_equal,omitempty"`
	LessThan           *float64 `yaml:"less_than,omitempty"`
	LessThanOrEqual    *float64 `yaml:"less_than_or_equal,omitempty"`

	// Range constraints
	Range *RangeCondition `yaml:"range,omitempty"`

	// Time constraints
	Before string `yaml:"before,omitempty"`
	After  string `yaml:"after,omitempty"`

	// Boolean equality
	Equals *bool `yaml:"equals,omitempty"`

	// Checklist completion criteria
	ChecklistCompletion *ChecklistCompletionCondition `yaml:"checklist_completion,omitempty"`

	// Logical operators (for future extension)
	And []Condition `yaml:"and,omitempty"`
	Or  []Condition `yaml:"or,omitempty"`
	Not *Condition  `yaml:"not,omitempty"`
}

// RangeCondition represents a range constraint.
type RangeCondition struct {
	Min          float64 `yaml:"min"`
	Max          float64 `yaml:"max"`
	MinInclusive *bool   `yaml:"min_inclusive,omitempty"`
	MaxInclusive *bool   `yaml:"max_inclusive,omitempty"`
}

// ChecklistCompletionCondition defines criteria for automatic scoring of checklist habits.
type ChecklistCompletionCondition struct {
	RequiredItems string `yaml:"required_items"` // "all" (only valid option)
}

// Validate validates a habit for correctness and consistency.
func (g *Habit) Validate() error {
	// Title is required
	if strings.TrimSpace(g.Title) == "" {
		return fmt.Errorf("habit title is required")
	}

	// Generate ID if not provided
	if g.ID == "" {
		g.ID = generateIDFromTitle(g.Title)
	}

	return g.validateInternal()
}

// ValidateAndTrackChanges validates a habit and returns whether it was modified.
// Returns (wasModified, error) where wasModified indicates if ID was generated.
func (g *Habit) ValidateAndTrackChanges() (bool, error) {
	// Title is required
	if strings.TrimSpace(g.Title) == "" {
		return false, fmt.Errorf("habit title is required")
	}

	// Check if ID needs to be generated
	wasModified := false
	if g.ID == "" {
		g.ID = generateIDFromTitle(g.Title)
		wasModified = true
	}

	return wasModified, g.validateInternal()
}

// validateInternal performs the core validation logic (shared between Validate methods).
func (g *Habit) validateInternal() error {
	// Validate ID format
	if !isValidID(g.ID) {
		return fmt.Errorf("habit ID '%s' is invalid: must contain only letters, numbers, and underscores", g.ID)
	}

	// Position is auto-assigned during parsing and not validated here

	// Habit type is required
	if g.HabitType == "" {
		return fmt.Errorf("habit_type is required")
	}

	// Validate habit type
	if !isValidHabitType(g.HabitType) {
		return fmt.Errorf("invalid habit_type: %s", g.HabitType)
	}

	// Validate field type
	if err := g.FieldType.Validate(); err != nil {
		return fmt.Errorf("invalid field_type: %w", err)
	}

	// Validate scoring requirements for simple habits
	if g.HabitType == SimpleHabit {
		if g.ScoringType == "" {
			return fmt.Errorf("scoring_type is required for simple habits")
		}
		if g.ScoringType == AutomaticScoring && g.Criteria == nil {
			return fmt.Errorf("criteria is required for automatic scoring")
		}
	}

	// Validate scoring requirements for elastic habits
	if g.HabitType == ElasticHabit {
		if g.ScoringType == "" {
			return fmt.Errorf("scoring_type is required for elastic habits")
		}
		if g.ScoringType == AutomaticScoring {
			if g.MiniCriteria == nil {
				return fmt.Errorf("mini_criteria is required for automatic scoring of elastic habits")
			}
			if g.MidiCriteria == nil {
				return fmt.Errorf("midi_criteria is required for automatic scoring of elastic habits")
			}
			if g.MaxiCriteria == nil {
				return fmt.Errorf("maxi_criteria is required for automatic scoring of elastic habits")
			}

			// Validate criteria ordering for numeric field types
			if err := g.validateElasticCriteriaOrdering(); err != nil {
				return fmt.Errorf("invalid elastic criteria ordering: %w", err)
			}
		}
	}

	// Validate scoring requirements for checklist habits
	if g.HabitType == ChecklistHabit {
		if g.ScoringType == "" {
			return fmt.Errorf("scoring_type is required for checklist habits")
		}
		if g.FieldType.Type != ChecklistFieldType {
			return fmt.Errorf("checklist habits must use checklist field type")
		}
		if g.FieldType.ChecklistID == "" {
			return fmt.Errorf("checklist_id is required for checklist field type")
		}
		if g.ScoringType == AutomaticScoring && g.Criteria == nil {
			return fmt.Errorf("criteria is required for automatic scoring of checklist habits")
		}

		// Validate checklist criteria if present
		if g.Criteria != nil {
			if err := g.validateChecklistCriteria(g.Criteria); err != nil {
				return fmt.Errorf("invalid checklist criteria for habit '%s': %w", g.Title, err)
			}
		}
	}

	return nil
}

// validateChecklistCriteria validates checklist-specific criteria
func (g *Habit) validateChecklistCriteria(criteria *Criteria) error {
	if criteria == nil {
		return nil
	}

	if criteria.Condition == nil {
		return fmt.Errorf("criteria condition is required")
	}

	// Validate checklist completion condition
	if criteria.Condition.ChecklistCompletion != nil {
		if err := criteria.Condition.ChecklistCompletion.Validate(); err != nil {
			return fmt.Errorf("invalid checklist completion condition: %w", err)
		}
	}

	return nil
}

// ValidateWithChecklistContext validates a habit with checklist context from checklists.yml
// This method provides enhanced validation that includes cross-reference checks
func (g *Habit) ValidateWithChecklistContext(checklistsExist func(string) bool) error {
	// First perform standard validation
	if err := g.Validate(); err != nil {
		return err
	}

	// Additional validation for checklist habits with context
	if g.HabitType == ChecklistHabit {
		if g.FieldType.ChecklistID != "" {
			if !checklistsExist(g.FieldType.ChecklistID) {
				return fmt.Errorf("checklist habit '%s' references non-existent checklist '%s'", g.Title, g.FieldType.ChecklistID)
			}
		}
	}

	return nil
}

// Validate validates a field type for correctness.
func (ft *FieldType) Validate() error {
	if ft.Type == "" {
		return fmt.Errorf("field type is required")
	}

	switch ft.Type {
	case TextFieldType:
		// Text fields don't need additional validation
	case BooleanFieldType:
		// Boolean fields don't need additional validation
	case UnsignedIntFieldType, UnsignedDecimalFieldType:
		if ft.Min != nil && *ft.Min < 0 {
			return fmt.Errorf("unsigned fields cannot have negative min value")
		}
	case DecimalFieldType:
		// Decimal fields can have any min/max
	case TimeFieldType:
		if ft.Format != "" && ft.Format != "HH:MM" {
			return fmt.Errorf("time fields only support HH:MM format")
		}
	case DurationFieldType:
		validFormats := []string{"HH:MM:SS", "minutes", "seconds"}
		if ft.Format != "" && !contains(validFormats, ft.Format) {
			return fmt.Errorf("duration format must be one of: %v", validFormats)
		}
	case ChecklistFieldType:
		if ft.ChecklistID == "" {
			return fmt.Errorf("checklist_id is required for checklist field type")
		}
	default:
		return fmt.Errorf("unknown field type: %s", ft.Type)
	}

	// Validate min/max constraints
	if ft.Min != nil && ft.Max != nil && *ft.Min > *ft.Max {
		return fmt.Errorf("min value (%v) cannot be greater than max value (%v)", *ft.Min, *ft.Max)
	}

	return nil
}

// Validate validates a schema for correctness and consistency.
func (s *Schema) Validate() error {
	// Version is required
	if s.Version == "" {
		return fmt.Errorf("schema version is required")
	}

	// Created date should be valid if provided
	if s.CreatedDate != "" {
		if _, err := time.Parse("2006-01-02", s.CreatedDate); err != nil {
			return fmt.Errorf("invalid created_date format, expected YYYY-MM-DD: %w", err)
		}
	}

	// Track unique constraints
	ids := make(map[string]bool)

	// Validate each habit and auto-assign positions
	for i := range s.Habits {
		// Auto-assign position based on array index (1-based)
		s.Habits[i].Position = i + 1

		if err := s.Habits[i].Validate(); err != nil {
			return fmt.Errorf("habit at index %d: %w", i, err)
		}

		// Check ID uniqueness
		if ids[s.Habits[i].ID] {
			return fmt.Errorf("duplicate habit ID: %s", s.Habits[i].ID)
		}
		ids[s.Habits[i].ID] = true
	}

	return nil
}

// ValidateAndTrackChanges validates a schema and returns whether it was modified.
// Returns (wasModified, error) where wasModified indicates if any habit IDs were generated.
func (s *Schema) ValidateAndTrackChanges() (bool, error) {
	// Version is required
	if s.Version == "" {
		return false, fmt.Errorf("schema version is required")
	}

	// Created date should be valid if provided
	if s.CreatedDate != "" {
		if _, err := time.Parse("2006-01-02", s.CreatedDate); err != nil {
			return false, fmt.Errorf("invalid created_date format, expected YYYY-MM-DD: %w", err)
		}
	}

	// Track unique constraints and modifications
	ids := make(map[string]bool)
	wasModified := false

	// Validate each habit and auto-assign positions
	for i := range s.Habits {
		// Auto-assign position based on array index (1-based)
		s.Habits[i].Position = i + 1

		habitModified, err := s.Habits[i].ValidateAndTrackChanges()
		if err != nil {
			return false, fmt.Errorf("habit at index %d: %w", i, err)
		}
		if habitModified {
			wasModified = true
		}

		// Check ID uniqueness
		if ids[s.Habits[i].ID] {
			return false, fmt.Errorf("duplicate habit ID: %s", s.Habits[i].ID)
		}
		ids[s.Habits[i].ID] = true
	}

	return wasModified, nil
}

// generateIDFromTitle creates a valid ID from a habit title.
func generateIDFromTitle(title string) string {
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
		id = "unnamed_habit"
	}

	return id
}

// isValidID checks if an ID contains only valid characters.
func isValidID(id string) bool {
	if id == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-z0-9_]+$`, id)
	return matched
}

// isValidHabitType checks if a habit type is valid.
func isValidHabitType(gt HabitType) bool {
	switch gt {
	case SimpleHabit, ElasticHabit, InformationalHabit, ChecklistHabit:
		return true
	default:
		return false
	}
}

// contains checks if a slice contains a specific string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// IsElastic returns true if this is an elastic habit.
func (g *Habit) IsElastic() bool {
	return g.HabitType == ElasticHabit
}

// IsSimple returns true if this is a simple habit.
func (g *Habit) IsSimple() bool {
	return g.HabitType == SimpleHabit
}

// IsInformational returns true if this is an informational habit.
func (g *Habit) IsInformational() bool {
	return g.HabitType == InformationalHabit
}

// RequiresAutomaticScoring returns true if this habit uses automatic scoring.
func (g *Habit) RequiresAutomaticScoring() bool {
	return g.ScoringType == AutomaticScoring
}

// RequiresManualScoring returns true if this habit uses manual scoring.
func (g *Habit) RequiresManualScoring() bool {
	return g.ScoringType == ManualScoring
}

// IsChecklist returns true if this is a checklist habit.
func (g *Habit) IsChecklist() bool {
	return g.HabitType == ChecklistHabit
}

// validateElasticCriteriaOrdering validates that elastic habit criteria are properly ordered
// for numeric field types (mini ≤ midi ≤ maxi).
func (g *Habit) validateElasticCriteriaOrdering() error {
	// Only validate ordering for numeric field types
	switch g.FieldType.Type {
	case UnsignedIntFieldType, UnsignedDecimalFieldType, DecimalFieldType, DurationFieldType:
		// These field types should have ordered criteria
	default:
		// For other field types (text, boolean, time), ordering doesn't apply
		return nil
	}

	mini := extractNumericCriteriaValue(g.MiniCriteria)
	midi := extractNumericCriteriaValue(g.MidiCriteria)
	maxi := extractNumericCriteriaValue(g.MaxiCriteria)

	// If any value couldn't be extracted, skip ordering validation
	if mini == nil || midi == nil || maxi == nil {
		return nil
	}

	// Validate mini ≤ midi ≤ maxi
	if *mini > *midi {
		return fmt.Errorf("mini criteria value (%.2f) must be ≤ midi criteria value (%.2f)", *mini, *midi)
	}
	if *midi > *maxi {
		return fmt.Errorf("midi criteria value (%.2f) must be ≤ maxi criteria value (%.2f)", *midi, *maxi)
	}

	return nil
}

// extractNumericCriteriaValue extracts a numeric value from criteria for ordering validation.
// Returns nil if no comparable numeric value can be extracted.
func extractNumericCriteriaValue(criteria *Criteria) *float64 {
	if criteria == nil || criteria.Condition == nil {
		return nil
	}

	cond := criteria.Condition

	// Try different numeric comparison operators
	if cond.GreaterThan != nil {
		return cond.GreaterThan
	}
	if cond.GreaterThanOrEqual != nil {
		return cond.GreaterThanOrEqual
	}
	if cond.LessThan != nil {
		return cond.LessThan
	}
	if cond.LessThanOrEqual != nil {
		return cond.LessThanOrEqual
	}

	// For range conditions, use the minimum value
	if cond.Range != nil {
		return &cond.Range.Min
	}

	// Could not extract a numeric value
	return nil
}
