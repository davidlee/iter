package entry

import (
	"davidlee/iter/internal/models"
	"github.com/charmbracelet/huh"
)

// AIDEV-NOTE: entry-field-input-interface; designed for T010 entry collection with immediate scoring feedback
// This interface abstracts field value input for entry recording with validation and feedback
// Extends patterns from goalconfig.FieldValueInput but specialized for entry collection with scoring

// EntryFieldInput provides field-type-aware input collection for entry recording
type EntryFieldInput interface {
	// CreateInputForm creates a huh form for collecting the field value during entry
	CreateInputForm(goal models.Goal) *huh.Form

	// GetValue returns the collected value in the appropriate type
	GetValue() interface{}

	// GetStringValue returns the value as a string for display/storage
	GetStringValue() string

	// Validate validates the current value against field constraints
	Validate() error

	// GetFieldType returns the field type this input handles
	GetFieldType() string

	// SetExistingValue sets an existing value for editing scenarios
	SetExistingValue(value interface{}) error

	// GetValidationError returns the current validation error state
	GetValidationError() error
}

// ScoringAwareInput extends EntryFieldInput for goals with automatic scoring
type ScoringAwareInput interface {
	EntryFieldInput

	// CanShowScoring returns true if this input can display scoring feedback
	CanShowScoring() bool

	// UpdateScoringDisplay updates the form to show scoring feedback
	UpdateScoringDisplay(level *models.AchievementLevel) error
}

// EntryFieldInputConfig holds configuration for field input creation
type EntryFieldInputConfig struct {
	Goal          models.Goal
	FieldType     models.FieldType
	ExistingEntry *ExistingEntry
	ShowScoring   bool // Whether to show immediate scoring feedback
}

// ExistingEntry represents existing data for editing scenarios
type ExistingEntry struct {
	Value            interface{}
	Notes            string
	AchievementLevel *models.AchievementLevel
}

// EntryResult represents the complete result of collecting an entry for a goal
type EntryResult struct {
	Value            interface{}              // The collected value (any type based on field type)
	AchievementLevel *models.AchievementLevel // Achievement level for elastic goals (nil for simple goals)
	Notes            string                   // Any notes collected from the user
}
