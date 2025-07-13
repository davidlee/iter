package entry

import (
	"fmt"
	"strings"

	"davidlee/iter/internal/models"
)

// AIDEV-NOTE: entry-validation-patterns; common validation and error messaging patterns for entry collection
// Provides consistent validation experience across all field input components

// ValidationResult represents the result of field validation
type ValidationResult struct {
	IsValid bool
	Error   error
	Message string // User-friendly message
}

// FieldValidator provides common validation patterns for entry fields
type FieldValidator struct{}

// NewFieldValidator creates a new field validator
func NewFieldValidator() *FieldValidator {
	return &FieldValidator{}
}

// ValidateRequired checks if a required field has a value
func (v *FieldValidator) ValidateRequired(value interface{}, fieldType string) ValidationResult {
	isEmpty := v.isEmpty(value)

	if isEmpty {
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("%s value is required", v.getFieldDisplayName(fieldType)),
			Message: fmt.Sprintf("Please enter a %s value", v.getFieldDisplayName(fieldType)),
		}
	}

	return ValidationResult{IsValid: true}
}

// ValidateFieldConstraints validates field-specific constraints
func (v *FieldValidator) ValidateFieldConstraints(value interface{}, fieldType models.FieldType) ValidationResult {
	switch fieldType.Type {
	case models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType:
		return v.validateNumericConstraints(value, fieldType)
	case models.TextFieldType:
		return v.validateTextConstraints(value, fieldType)
	default:
		return ValidationResult{IsValid: true}
	}
}

// GetValidationErrorMessage creates a user-friendly error message
func (v *FieldValidator) GetValidationErrorMessage(err error, fieldType string) string {
	if err == nil {
		return ""
	}

	errorMsg := err.Error()
	fieldName := v.getFieldDisplayName(fieldType)

	// Convert technical errors to user-friendly messages
	switch {
	case strings.Contains(errorMsg, "invalid"):
		return fmt.Sprintf("Please enter a valid %s", fieldName)
	case strings.Contains(errorMsg, "required"):
		return fmt.Sprintf("%s is required", strings.Title(fieldName))
	case strings.Contains(errorMsg, "minimum"):
		return errorMsg // Constraint errors are already user-friendly
	case strings.Contains(errorMsg, "maximum"):
		return errorMsg
	case strings.Contains(errorMsg, "format"):
		return fmt.Sprintf("Please check the format of your %s", fieldName)
	default:
		return fmt.Sprintf("Invalid %s: %s", fieldName, errorMsg)
	}
}

// Private methods

func (v *FieldValidator) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []string:
		return len(v) == 0
	default:
		return false
	}
}

func (v *FieldValidator) getFieldDisplayName(fieldType string) string {
	switch fieldType {
	case models.BooleanFieldType:
		return "yes/no answer"
	case models.TextFieldType:
		return "text"
	case models.UnsignedIntFieldType:
		return "whole number"
	case models.UnsignedDecimalFieldType:
		return "positive number"
	case models.DecimalFieldType:
		return "number"
	case models.TimeFieldType:
		return "time"
	case models.DurationFieldType:
		return "duration"
	case models.ChecklistFieldType:
		return "checklist selection"
	default:
		return "value"
	}
}

func (v *FieldValidator) validateNumericConstraints(value interface{}, fieldType models.FieldType) ValidationResult {
	// This would implement numeric constraint validation
	// For now, return valid (constraints are handled in individual input components)
	return ValidationResult{IsValid: true}
}

func (v *FieldValidator) validateTextConstraints(value interface{}, fieldType models.FieldType) ValidationResult {
	if textVal, ok := value.(string); ok {
		// Could add text length constraints here
		if len(textVal) > 10000 { // Example max length
			return ValidationResult{
				IsValid: false,
				Error:   fmt.Errorf("text too long"),
				Message: "Text must be less than 10,000 characters",
			}
		}
	}

	return ValidationResult{IsValid: true}
}

// Common validation error messages
var (
	ErrBooleanRequired   = fmt.Errorf("please select yes or no")
	ErrTextRequired      = fmt.Errorf("text value is required")
	ErrNumericRequired   = fmt.Errorf("numeric value is required")
	ErrTimeRequired      = fmt.Errorf("time value is required")
	ErrDurationRequired  = fmt.Errorf("duration value is required")
	ErrChecklistRequired = fmt.Errorf("please select at least one item")
)

// Validation helper functions

// IsValidTimeFormat checks if a string is a valid time format
func IsValidTimeFormat(timeStr string) bool {
	trimmed := strings.TrimSpace(timeStr)
	return len(trimmed) > 0 && (len(trimmed) == 4 || len(trimmed) == 5) && strings.Contains(trimmed, ":")
}

// IsValidDurationFormat checks if a string is a valid duration format
func IsValidDurationFormat(durationStr string) bool {
	trimmed := strings.TrimSpace(durationStr)
	return len(trimmed) > 0 && (strings.Contains(trimmed, "h") || strings.Contains(trimmed, "m") || strings.Contains(trimmed, "s"))
}

// FormatValidationMessage formats a validation message for display
func FormatValidationMessage(fieldType, message string) string {
	if message == "" {
		return ""
	}

	return fmt.Sprintf("⚠️  %s", message)
}
