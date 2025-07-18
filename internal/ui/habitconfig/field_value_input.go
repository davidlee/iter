package habitconfig

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"

	"github.com/davidlee/vice/internal/models"
)

// AIDEV-NOTE: Field value input UI foundation for entry recording system (T005/3.5)
// These components provide type-safe, validated input for different field types
// Designed for reuse in entry recording (T007) and habit criteria definition
// Each input component handles validation, formatting, and user experience
// Interface-based design allows for easy extension and testing
//
// AIDEV-TODO: Integrate with entry recording system (T007) using FieldValueInputFactory
// AIDEV-TODO: Extend for habit criteria definition in simple/elastic habits with automatic scoring
// AIDEV-TODO: Add unit tests for all input components and validation logic

// FieldValueInput provides a type-safe interface for field value input
type FieldValueInput interface {
	// CreateInputForm creates a huh form for collecting the field value
	CreateInputForm(prompt string) *huh.Form

	// GetValue returns the collected value in the appropriate type
	GetValue() interface{}

	// GetStringValue returns the value as a string for display/storage
	GetStringValue() string

	// Validate validates the current value
	Validate() error

	// GetFieldType returns the field type this input handles
	GetFieldType() string
}

// BooleanInput handles boolean field value input with clear yes/no display
type BooleanInput struct {
	value bool
}

// NewBooleanInput creates a new boolean input component
func NewBooleanInput() *BooleanInput {
	return &BooleanInput{}
}

// CreateInputForm creates a boolean input form with clear yes/no display
func (bi *BooleanInput) CreateInputForm(prompt string) *huh.Form {
	if prompt == "" {
		prompt = "Select value"
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(prompt).
				Description("Choose Yes (true) or No (false)").
				Value(&bi.value),
		),
	)
}

// GetValue returns the boolean value
func (bi *BooleanInput) GetValue() interface{} {
	return bi.value
}

// GetStringValue returns the boolean as a string
func (bi *BooleanInput) GetStringValue() string {
	if bi.value {
		return "true"
	}
	return "false"
}

// Validate validates the boolean value (always valid)
func (bi *BooleanInput) Validate() error {
	return nil // Boolean values are always valid
}

// GetFieldType returns the field type
func (bi *BooleanInput) GetFieldType() string {
	return models.BooleanFieldType
}

// TextInput handles text field value input with single-line and multiline support
type TextInput struct {
	value     string
	multiline bool
}

// NewTextInput creates a new text input component
func NewTextInput(multiline bool) *TextInput {
	return &TextInput{
		multiline: multiline,
	}
}

// CreateInputForm creates a text input form with validation
func (ti *TextInput) CreateInputForm(prompt string) *huh.Form {
	if prompt == "" {
		prompt = "Enter text value"
	}

	if ti.multiline {
		return huh.NewForm(
			huh.NewGroup(
				huh.NewText().
					Title(prompt).
					Description("Enter multiple lines of text").
					Value(&ti.value),
			),
		)
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Description("Enter text value").
				Value(&ti.value),
		),
	)
}

// GetValue returns the text value
func (ti *TextInput) GetValue() interface{} {
	return ti.value
}

// GetStringValue returns the text value
func (ti *TextInput) GetStringValue() string {
	return ti.value
}

// Validate validates the text value
func (ti *TextInput) Validate() error {
	// Text values are generally always valid, but we could add length checks
	return nil
}

// GetFieldType returns the field type
func (ti *TextInput) GetFieldType() string {
	return models.TextFieldType
}

// NumericInput handles numeric field value input with unit display and validation
type NumericInput struct {
	value       string
	numericType string
	unit        string
	min         *float64
	max         *float64
}

// NewNumericInput creates a new numeric input component
func NewNumericInput(numericType, unit string, minVal, maxVal *float64) *NumericInput {
	return &NumericInput{
		numericType: numericType,
		unit:        unit,
		min:         minVal,
		max:         maxVal,
	}
}

// CreateInputForm creates a numeric input form with unit display and validation
func (ni *NumericInput) CreateInputForm(prompt string) *huh.Form {
	if prompt == "" {
		prompt = fmt.Sprintf("Enter numeric value (%s)", ni.unit)
	}

	description := fmt.Sprintf("Enter a %s value", ni.getNumericTypeDescription())
	if ni.unit != "" && ni.unit != "times" {
		description += fmt.Sprintf(" (in %s)", ni.unit)
	}
	if ni.min != nil || ni.max != nil {
		description += ni.getConstraintsDescription()
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Description(description).
				Value(&ni.value).
				Validate(ni.validateInput),
		),
	)
}

// GetValue returns the numeric value as the appropriate type
func (ni *NumericInput) GetValue() interface{} {
	val, err := ni.parseValue()
	if err != nil {
		return nil
	}
	return val
}

// GetStringValue returns the numeric value as a string
func (ni *NumericInput) GetStringValue() string {
	return ni.value
}

// Validate validates the numeric value
func (ni *NumericInput) Validate() error {
	return ni.validateInput(ni.value)
}

// GetFieldType returns the field type
func (ni *NumericInput) GetFieldType() string {
	return ni.numericType
}

// Private methods for NumericInput

func (ni *NumericInput) validateInput(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("numeric value is required")
	}

	val, err := ni.parseValue()
	if err != nil {
		return err
	}

	floatVal := val.(float64)

	if ni.min != nil && floatVal < *ni.min {
		return fmt.Errorf("value must be at least %g", *ni.min)
	}

	if ni.max != nil && floatVal > *ni.max {
		return fmt.Errorf("value must be at most %g", *ni.max)
	}

	return nil
}

func (ni *NumericInput) parseValue() (interface{}, error) {
	trimmed := strings.TrimSpace(ni.value)

	switch ni.numericType {
	case models.UnsignedIntFieldType:
		val, err := strconv.ParseUint(trimmed, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid unsigned integer: %w", err)
		}
		return float64(val), nil

	case models.UnsignedDecimalFieldType:
		val, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid unsigned decimal: %w", err)
		}
		if val < 0 {
			return nil, fmt.Errorf("value must be positive")
		}
		return val, nil

	case models.DecimalFieldType:
		val, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid decimal: %w", err)
		}
		return val, nil

	default:
		return nil, fmt.Errorf("unknown numeric type: %s", ni.numericType)
	}
}

func (ni *NumericInput) getNumericTypeDescription() string {
	switch ni.numericType {
	case models.UnsignedIntFieldType:
		return "whole number (0, 1, 2, 3...)"
	case models.UnsignedDecimalFieldType:
		return "positive decimal (0.5, 1.25, 2.7...)"
	case models.DecimalFieldType:
		return "decimal number (including negative)"
	default:
		return "numeric"
	}
}

func (ni *NumericInput) getConstraintsDescription() string {
	switch {
	case ni.min != nil && ni.max != nil:
		return fmt.Sprintf(" (between %g and %g)", *ni.min, *ni.max)
	case ni.min != nil:
		return fmt.Sprintf(" (minimum %g)", *ni.min)
	case ni.max != nil:
		return fmt.Sprintf(" (maximum %g)", *ni.max)
	default:
		return ""
	}
}

// TimeInput handles time of day input with HH:MM format
type TimeInput struct {
	value string
}

// NewTimeInput creates a new time input component
func NewTimeInput() *TimeInput {
	return &TimeInput{}
}

// CreateInputForm creates a time input form with formatted input
func (ti *TimeInput) CreateInputForm(prompt string) *huh.Form {
	if prompt == "" {
		prompt = "Enter time of day"
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Description("Enter time in HH:MM format (e.g., 14:30, 09:15)").
				Placeholder("14:30").
				Value(&ti.value).
				Validate(ti.validateTime),
		),
	)
}

// GetValue returns the time value as a parsed time
func (ti *TimeInput) GetValue() interface{} {
	parsedTime, err := ti.parseTime()
	if err != nil {
		return nil
	}
	return parsedTime
}

// GetStringValue returns the time value as a string
func (ti *TimeInput) GetStringValue() string {
	return ti.value
}

// Validate validates the time value
func (ti *TimeInput) Validate() error {
	return ti.validateTime(ti.value)
}

// GetFieldType returns the field type
func (ti *TimeInput) GetFieldType() string {
	return models.TimeFieldType
}

// Private methods for TimeInput

func (ti *TimeInput) validateTime(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("time value is required")
	}

	_, err := ti.parseTime()
	return err
}

func (ti *TimeInput) parseTime() (time.Time, error) {
	trimmed := strings.TrimSpace(ti.value)

	// Try parsing as HH:MM
	parsedTime, err := time.Parse("15:04", trimmed)
	if err != nil {
		// Try parsing as H:MM
		parsedTime, err = time.Parse("3:04", trimmed)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid time format, use HH:MM (e.g., 14:30)")
		}
	}

	return parsedTime, nil
}

// DurationInput handles duration input with flexible format support
type DurationInput struct {
	value string
}

// NewDurationInput creates a new duration input component
func NewDurationInput() *DurationInput {
	return &DurationInput{}
}

// CreateInputForm creates a duration input form with flexible format support
func (di *DurationInput) CreateInputForm(prompt string) *huh.Form {
	if prompt == "" {
		prompt = "Enter duration"
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Description("Enter duration (e.g., 1h 30m, 45m, 2h, 90m)").
				Placeholder("1h 30m").
				Value(&di.value).
				Validate(di.validateDuration),
		),
	)
}

// GetValue returns the duration value as a time.Duration
func (di *DurationInput) GetValue() interface{} {
	parsedDuration, err := di.parseDuration()
	if err != nil {
		return nil
	}
	return parsedDuration
}

// GetStringValue returns the duration value as a string
func (di *DurationInput) GetStringValue() string {
	return di.value
}

// Validate validates the duration value
func (di *DurationInput) Validate() error {
	return di.validateDuration(di.value)
}

// GetFieldType returns the field type
func (di *DurationInput) GetFieldType() string {
	return models.DurationFieldType
}

// Private methods for DurationInput

func (di *DurationInput) validateDuration(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("duration value is required")
	}

	_, err := di.parseDuration()
	return err
}

func (di *DurationInput) parseDuration() (time.Duration, error) {
	trimmed := strings.TrimSpace(di.value)

	// Try parsing as Go duration format first
	duration, err := time.ParseDuration(trimmed)
	if err == nil {
		return duration, nil
	}

	// If that fails, try to parse common formats manually
	// This is a simplified parser for common duration formats
	return time.ParseDuration(trimmed)
}

// FieldValueInputFactory creates appropriate input components based on field type
// AIDEV-NOTE: Factory pattern enables automatic component selection for entry recording
// Used by T007 entry system to create appropriate inputs for each habit's field type
type FieldValueInputFactory struct{}

// NewFieldValueInputFactory creates a new factory
func NewFieldValueInputFactory() *FieldValueInputFactory {
	return &FieldValueInputFactory{}
}

// CreateInput creates the appropriate input component for a given field type
func (f *FieldValueInputFactory) CreateInput(fieldType models.FieldType) (FieldValueInput, error) {
	switch fieldType.Type {
	case models.BooleanFieldType:
		return NewBooleanInput(), nil

	case models.TextFieldType:
		multiline := fieldType.Multiline != nil && *fieldType.Multiline
		return NewTextInput(multiline), nil

	case models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType:
		return NewNumericInput(fieldType.Type, fieldType.Unit, fieldType.Min, fieldType.Max), nil

	case models.TimeFieldType:
		return NewTimeInput(), nil

	case models.DurationFieldType:
		return NewDurationInput(), nil

	default:
		return nil, fmt.Errorf("unsupported field type: %s", fieldType.Type)
	}
}
