package entry

import (
	"fmt"

	"davidlee/vice/internal/models"
)

// AIDEV-NOTE: entry-field-input-factory; creates appropriate input components for entry collection with scoring integration
// Extends goalconfig.FieldValueInputFactory patterns for entry-specific needs with immediate feedback

// EntryFieldInputFactory creates appropriate field input components for entry collection
//
//revive:disable-next-line:exported
type EntryFieldInputFactory struct{}

// NewEntryFieldInputFactory creates a new entry field input factory
func NewEntryFieldInputFactory() *EntryFieldInputFactory {
	return &EntryFieldInputFactory{}
}

// CreateInput creates the appropriate entry input component for a given field type and configuration
func (f *EntryFieldInputFactory) CreateInput(config EntryFieldInputConfig) (EntryFieldInput, error) {
	switch config.FieldType.Type {
	case models.BooleanFieldType:
		return NewBooleanEntryInput(config), nil

	case models.TextFieldType:
		return NewTextEntryInput(config), nil

	case models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType:
		return NewNumericEntryInput(config), nil

	case models.TimeFieldType:
		return NewTimeEntryInput(config), nil

	case models.DurationFieldType:
		return NewDurationEntryInput(config), nil

	case models.ChecklistFieldType:
		return NewChecklistEntryInput(config), nil

	default:
		return nil, fmt.Errorf("unsupported field type for entry collection: %s", config.FieldType.Type)
	}
}

// CreateScoringAwareInput creates an input component that can display scoring feedback
func (f *EntryFieldInputFactory) CreateScoringAwareInput(config EntryFieldInputConfig) (ScoringAwareInput, error) {
	input, err := f.CreateInput(config)
	if err != nil {
		return nil, err
	}

	// Check if the input supports scoring feedback
	if scoringInput, ok := input.(ScoringAwareInput); ok {
		return scoringInput, nil
	}

	// Return a wrapper that provides no-op scoring methods for inputs that don't support it
	return &scoringAwareWrapper{EntryFieldInput: input}, nil
}

// GetSupportedFieldTypes returns the list of field types supported by the factory
func (f *EntryFieldInputFactory) GetSupportedFieldTypes() []string {
	return []string{
		models.BooleanFieldType,
		models.TextFieldType,
		models.UnsignedIntFieldType,
		models.UnsignedDecimalFieldType,
		models.DecimalFieldType,
		models.TimeFieldType,
		models.DurationFieldType,
		models.ChecklistFieldType,
	}
}

// IsFieldTypeSupported checks if a given field type is supported
func (f *EntryFieldInputFactory) IsFieldTypeSupported(fieldType string) bool {
	supportedTypes := f.GetSupportedFieldTypes()
	for _, supported := range supportedTypes {
		if fieldType == supported {
			return true
		}
	}
	return false
}

// scoringAwareWrapper provides no-op scoring methods for inputs that don't support scoring
type scoringAwareWrapper struct {
	EntryFieldInput
}

// CanShowScoring returns false for wrapped inputs
func (w *scoringAwareWrapper) CanShowScoring() bool {
	return false
}

// UpdateScoringDisplay is a no-op for wrapped inputs
func (w *scoringAwareWrapper) UpdateScoringDisplay(_ *models.AchievementLevel) error {
	return nil // No-op for inputs that don't support scoring
}
