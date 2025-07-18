package entry

import (
	"github.com/charmbracelet/huh"
	"github.com/davidlee/vice/internal/models"
)

// AIDEV-NOTE: entry-field-input-interface; designed for T010 entry collection with immediate scoring feedback
// This interface abstracts field value input for entry recording with validation and feedback
// Extends patterns from habitconfig.FieldValueInput but specialized for entry collection with scoring

// EntryFieldInput provides field-type-aware input collection for entry recording
//
//revive:disable-next-line:exported
type EntryFieldInput interface {
	// CreateInputForm creates a huh form for collecting the field value during entry
	CreateInputForm(habit models.Habit) *huh.Form

	// GetValue returns the collected value in the appropriate type
	GetValue() interface{}

	// GetStringValue returns the value as a string for display/storage
	GetStringValue() string

	// GetStatus returns the entry completion status (completed/skipped/failed)
	GetStatus() models.EntryStatus

	// Validate validates the current value against field constraints
	Validate() error

	// GetFieldType returns the field type this input handles
	GetFieldType() string

	// SetExistingValue sets an existing value for editing scenarios
	SetExistingValue(value interface{}) error

	// GetValidationError returns the current validation error state
	GetValidationError() error
}

// ScoringAwareInput extends EntryFieldInput for habits with automatic scoring
type ScoringAwareInput interface {
	EntryFieldInput

	// CanShowScoring returns true if this input can display scoring feedback
	CanShowScoring() bool

	// UpdateScoringDisplay updates the form to show scoring feedback
	UpdateScoringDisplay(level *models.AchievementLevel) error
}

// EntryFieldInputConfig holds configuration for field input creation
//
//revive:disable-next-line:exported
type EntryFieldInputConfig struct {
	Habit          models.Habit
	FieldType      models.FieldType
	ExistingEntry  *ExistingEntry
	ShowScoring    bool   // Whether to show immediate scoring feedback
	ChecklistsPath string // Path to checklists.yml file (for checklist fields)
}

// ExistingEntry represents existing data for editing scenarios
type ExistingEntry struct {
	Value            interface{}
	Notes            string
	AchievementLevel *models.AchievementLevel
}

// EntryResult represents the complete result of collecting an entry for a habit
// AIDEV-NOTE: T012/2.1-enhanced; added Status field for skip functionality integration
//
//revive:disable-next-line:exported
type EntryResult struct {
	Value            interface{}              // The collected value (any type based on field type)
	AchievementLevel *models.AchievementLevel // Achievement level for elastic habits (nil for simple habits)
	Notes            string                   // Any notes collected from the user
	Status           models.EntryStatus       // Entry completion status (completed/skipped/failed)
}
