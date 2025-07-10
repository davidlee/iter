// Package models defines the data structures for the iter application.
package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Schema represents the top-level goal schema structure.
type Schema struct {
	Version     string `yaml:"version"`
	CreatedDate string `yaml:"created_date"`
	Goals       []Goal `yaml:"goals"`
}

// Goal represents a single goal in the schema.
type Goal struct {
	Title       string     `yaml:"title"`
	ID          string     `yaml:"id,omitempty"`
	Position    int        `yaml:"position"`
	Description string     `yaml:"description,omitempty"`
	GoalType    GoalType   `yaml:"goal_type"`
	FieldType   FieldType  `yaml:"field_type"`
	ScoringType ScoringType `yaml:"scoring_type,omitempty"`
	Criteria    *Criteria  `yaml:"criteria,omitempty"`
	
	// Elastic goal fields (not used for simple goals)
	MiniCriteria *Criteria `yaml:"mini_criteria,omitempty"`
	MidiCriteria *Criteria `yaml:"midi_criteria,omitempty"`
	MaxiCriteria *Criteria `yaml:"maxi_criteria,omitempty"`
	
	// Informational goal fields (not used for simple goals)
	Direction string `yaml:"direction,omitempty"`
	
	// UI fields
	Prompt   string `yaml:"prompt,omitempty"`
	HelpText string `yaml:"help_text,omitempty"`
}

// GoalType represents the type of goal.
type GoalType string

// Goal types define the behavior and scoring model for goals.
const (
	SimpleGoal        GoalType = "simple"        // Boolean pass/fail goals
	ElasticGoal       GoalType = "elastic"       // Three-tier achievement goals (mini/midi/maxi)
	InformationalGoal GoalType = "informational" // Data collection without scoring
)

// ScoringType represents how the goal is scored.
type ScoringType string

// Scoring types define how goal achievement is determined.
const (
	ManualScoring    ScoringType = "manual"    // User manually marks completion
	AutomaticScoring ScoringType = "automatic" // Automatic scoring based on criteria
)

// FieldType represents the data type for goal values.
type FieldType struct {
	Type      string `yaml:"type"`
	Multiline *bool  `yaml:"multiline,omitempty"`
	Default   *bool  `yaml:"default,omitempty"`
	Unit      string `yaml:"unit,omitempty"`
	Min       *float64 `yaml:"min,omitempty"`
	Max       *float64 `yaml:"max,omitempty"`
	Format    string `yaml:"format,omitempty"`
}

// Field type constants
const (
	TextFieldType             = "text"
	BooleanFieldType          = "boolean"
	UnsignedIntFieldType      = "unsigned_int"
	UnsignedDecimalFieldType  = "unsigned_decimal"
	DecimalFieldType          = "decimal"
	TimeFieldType             = "time"
	DurationFieldType         = "duration"
)

// Criteria represents goal achievement criteria.
type Criteria struct {
	Description string     `yaml:"description,omitempty"`
	Condition   *Condition `yaml:"condition"`
}

// Condition represents the logical condition for criteria evaluation.
type Condition struct {
	// Numeric/Duration comparisons
	GreaterThan          *float64 `yaml:"greater_than,omitempty"`
	GreaterThanOrEqual   *float64 `yaml:"greater_than_or_equal,omitempty"`
	LessThan             *float64 `yaml:"less_than,omitempty"`
	LessThanOrEqual      *float64 `yaml:"less_than_or_equal,omitempty"`
	
	// Range constraints
	Range *RangeCondition `yaml:"range,omitempty"`
	
	// Time constraints
	Before string `yaml:"before,omitempty"`
	After  string `yaml:"after,omitempty"`
	
	// Boolean equality
	Equals *bool `yaml:"equals,omitempty"`
	
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

// Validate validates a goal for correctness and consistency.
func (g *Goal) Validate() error {
	// Title is required
	if strings.TrimSpace(g.Title) == "" {
		return fmt.Errorf("goal title is required")
	}
	
	// Generate ID if not provided
	if g.ID == "" {
		g.ID = generateIDFromTitle(g.Title)
	}
	
	// Validate ID format
	if !isValidID(g.ID) {
		return fmt.Errorf("goal ID '%s' is invalid: must contain only letters, numbers, and underscores", g.ID)
	}
	
	// Position must be positive
	if g.Position <= 0 {
		return fmt.Errorf("goal position must be positive, got %d", g.Position)
	}
	
	// Goal type is required
	if g.GoalType == "" {
		return fmt.Errorf("goal_type is required")
	}
	
	// Validate goal type
	if !isValidGoalType(g.GoalType) {
		return fmt.Errorf("invalid goal_type: %s", g.GoalType)
	}
	
	// Validate field type
	if err := g.FieldType.Validate(); err != nil {
		return fmt.Errorf("invalid field_type: %w", err)
	}
	
	// Validate scoring requirements for simple goals
	if g.GoalType == SimpleGoal {
		if g.ScoringType == "" {
			return fmt.Errorf("scoring_type is required for simple goals")
		}
		if g.ScoringType == AutomaticScoring && g.Criteria == nil {
			return fmt.Errorf("criteria is required for automatic scoring")
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
	positions := make(map[int]bool)
	
	// Validate each goal
	for i, goal := range s.Goals {
		if err := goal.Validate(); err != nil {
			return fmt.Errorf("goal at index %d: %w", i, err)
		}
		
		// Check ID uniqueness
		if ids[goal.ID] {
			return fmt.Errorf("duplicate goal ID: %s", goal.ID)
		}
		ids[goal.ID] = true
		
		// Check position uniqueness
		if positions[goal.Position] {
			return fmt.Errorf("duplicate goal position: %d", goal.Position)
		}
		positions[goal.Position] = true
	}
	
	return nil
}

// generateIDFromTitle creates a valid ID from a goal title.
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
		id = "unnamed_goal"
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

// isValidGoalType checks if a goal type is valid.
func isValidGoalType(gt GoalType) bool {
	switch gt {
	case SimpleGoal, ElasticGoal, InformationalGoal:
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