package wizard

import (
	"encoding/json"
	"fmt"

	"davidlee/iter/internal/models"
)

// GoalWizardState implements WizardState for goal creation  
type GoalWizardState struct {
	CurrentStep    int                 `json:"currentStep"`
	TotalSteps     int                 `json:"totalSteps"`
	Steps          map[int]StepData    `json:"steps"`
	CompletedSteps map[int]bool        `json:"completedSteps"`
	GoalType       models.GoalType     `json:"goalType"`
}

// NewGoalWizardState creates a new goal wizard state
func NewGoalWizardState(goalType models.GoalType) *GoalWizardState {
	totalSteps := calculateTotalSteps(goalType)
	return &GoalWizardState{
		CurrentStep:    0,
		TotalSteps:     totalSteps,
		Steps:          make(map[int]StepData),
		CompletedSteps: make(map[int]bool),
		GoalType:       goalType,
	}
}

func (s *GoalWizardState) GetStep(index int) StepData {
	return s.Steps[index]
}

func (s *GoalWizardState) SetStep(index int, data StepData) {
	s.Steps[index] = data
}

func (s *GoalWizardState) Validate() []ValidationError {
	var errors []ValidationError
	for i := 0; i < s.TotalSteps; i++ {
		if stepData := s.Steps[i]; stepData != nil {
			if !stepData.IsValid() {
				errors = append(errors, ValidationError{
					Step:    i,
					Message: "Step data is invalid",
				})
			}
		}
	}
	return errors
}

func (s *GoalWizardState) ToGoal() (*models.Goal, error) {
	// Collect data from all steps and construct Goal
	// This will be implemented based on specific step data types
	return nil, fmt.Errorf("ToGoal not yet implemented")
}

func (s *GoalWizardState) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

func (s *GoalWizardState) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *GoalWizardState) GetCurrentStep() int {
	return s.CurrentStep
}

func (s *GoalWizardState) SetCurrentStep(step int) {
	if step >= 0 && step < s.TotalSteps {
		s.CurrentStep = step
	}
}

func (s *GoalWizardState) GetTotalSteps() int {
	return s.TotalSteps
}

func (s *GoalWizardState) IsStepCompleted(index int) bool {
	return s.CompletedSteps[index]
}

func (s *GoalWizardState) MarkStepCompleted(index int) {
	s.CompletedSteps[index] = true
}

// BasicInfoStepData holds basic goal information
type BasicInfoStepData struct {
	Title       string
	Description string
	GoalType    models.GoalType
	valid       bool
}

func (d *BasicInfoStepData) IsValid() bool {
	return d.valid && d.Title != ""
}

func (d *BasicInfoStepData) GetData() interface{} {
	return d
}

func (d *BasicInfoStepData) SetData(data interface{}) error {
	if typedData, ok := data.(*BasicInfoStepData); ok {
		*d = *typedData
		d.valid = d.Title != ""
		return nil
	}
	return fmt.Errorf("invalid data type for BasicInfoStepData")
}

// FieldConfigStepData holds field configuration
type FieldConfigStepData struct {
	FieldType string
	Unit      string
	Min       *float64
	Max       *float64
	Multiline bool
	valid     bool
}

func (d *FieldConfigStepData) IsValid() bool {
	return d.valid && d.FieldType != ""
}

func (d *FieldConfigStepData) GetData() interface{} {
	return d
}

func (d *FieldConfigStepData) SetData(data interface{}) error {
	if typedData, ok := data.(*FieldConfigStepData); ok {
		*d = *typedData
		d.valid = d.FieldType != ""
		return nil
	}
	return fmt.Errorf("invalid data type for FieldConfigStepData")
}

// ScoringStepData holds scoring configuration
type ScoringStepData struct {
	ScoringType models.ScoringType
	Direction   string // For informational goals
	valid       bool
}

func (d *ScoringStepData) IsValid() bool {
	return d.valid
}

func (d *ScoringStepData) GetData() interface{} {
	return d
}

func (d *ScoringStepData) SetData(data interface{}) error {
	if typedData, ok := data.(*ScoringStepData); ok {
		*d = *typedData
		d.valid = true
		return nil
	}
	return fmt.Errorf("invalid data type for ScoringStepData")
}

// CriteriaStepData holds criteria configuration for one level
type CriteriaStepData struct {
	Level           string // mini, midi, maxi
	Description     string
	ComparisonType  string
	Value           string
	BooleanValue    bool
	valid           bool
}

func (d *CriteriaStepData) IsValid() bool {
	return d.valid
}

func (d *CriteriaStepData) GetData() interface{} {
	return d
}

func (d *CriteriaStepData) SetData(data interface{}) error {
	if typedData, ok := data.(*CriteriaStepData); ok {
		*d = *typedData
		d.valid = true // Basic validation - enhance as needed
		return nil
	}
	return fmt.Errorf("invalid data type for CriteriaStepData")
}

// Helper function to calculate total steps based on goal type
func calculateTotalSteps(goalType models.GoalType) int {
	switch goalType {
	case models.SimpleGoal:
		return 4 // Basic info, scoring, criteria (if auto), confirmation
	case models.ElasticGoal:
		return 8 // Basic info, field config, scoring, mini/midi/maxi criteria, validation, confirmation
	case models.InformationalGoal:
		return 3 // Basic info, field config, confirmation
	default:
		return 4
	}
}