package wizard

import (
	"encoding/json"
	"fmt"

	"davidlee/iter/internal/models"
)

// GoalState implements State for goal creation  
type GoalState struct {
	CurrentStep    int                 `json:"currentStep"`
	TotalSteps     int                 `json:"totalSteps"`
	Steps          map[int]StepData    `json:"steps"`
	CompletedSteps map[int]bool        `json:"completedSteps"`
	GoalType       models.GoalType     `json:"goalType"`
}

// NewGoalState creates a new goal wizard state
func NewGoalState(goalType models.GoalType) *GoalState {
	totalSteps := calculateTotalSteps(goalType)
	return &GoalState{
		CurrentStep:    0,
		TotalSteps:     totalSteps,
		Steps:          make(map[int]StepData),
		CompletedSteps: make(map[int]bool),
		GoalType:       goalType,
	}
}

// GetStep returns the step data for the given index
func (s *GoalState) GetStep(index int) StepData {
	return s.Steps[index]
}

// SetStep sets the step data for the given index
func (s *GoalState) SetStep(index int, data StepData) {
	s.Steps[index] = data
}

// Validate validates all steps in the wizard
func (s *GoalState) Validate() []ValidationError {
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

// ToGoal converts the wizard state to a Goal model
func (s *GoalState) ToGoal() (*models.Goal, error) {
	// Get basic info (required for all goals)
	basicData := s.GetStep(0)
	if basicData == nil {
		return nil, fmt.Errorf("basic information is required")
	}
	
	basicInfo, ok := basicData.(*BasicInfoStepData)
	if !ok {
		return nil, fmt.Errorf("invalid basic information data")
	}
	
	// Create base goal
	goal := &models.Goal{
		Title:       basicInfo.Title,
		Description: basicInfo.Description,
		GoalType:    basicInfo.GoalType,
		Position:    0, // Will be set by the configurator
	}
	
	// Add field type configuration
	switch s.GoalType {
	case models.SimpleGoal:
		goal.FieldType = models.FieldType{
			Type: models.BooleanFieldType,
		}
		
		// Add scoring configuration
		if err := s.addSimpleGoalScoring(goal); err != nil {
			return nil, fmt.Errorf("failed to add scoring configuration: %w", err)
		}
		
	case models.ElasticGoal:
		// TODO: Implement elastic goal configuration
		return nil, fmt.Errorf("elastic goal configuration not yet implemented")
		
	case models.InformationalGoal:
		// TODO: Implement informational goal configuration
		return nil, fmt.Errorf("informational goal configuration not yet implemented")
	}
	
	return goal, nil
}

func (s *GoalState) addSimpleGoalScoring(goal *models.Goal) error {
	// Get scoring configuration
	scoringData := s.GetStep(1)
	if scoringData == nil {
		return fmt.Errorf("scoring configuration is required")
	}
	
	scoring, ok := scoringData.(*ScoringStepData)
	if !ok {
		return fmt.Errorf("invalid scoring configuration data")
	}
	
	goal.ScoringType = scoring.ScoringType
	
	// Add criteria if automatic scoring
	if scoring.ScoringType == models.AutomaticScoring {
		criteriaData := s.GetStep(2)
		if criteriaData == nil {
			return fmt.Errorf("criteria configuration is required for automatic scoring")
		}
		
		criteria, ok := criteriaData.(*CriteriaStepData)
		if !ok {
			return fmt.Errorf("invalid criteria configuration data")
		}
		
		goal.Criteria = &models.Criteria{
			Description: criteria.Description,
			Condition: &models.Condition{
				Equals: &criteria.BooleanValue,
			},
		}
	}
	
	return nil
}

// Serialize converts the state to JSON bytes
func (s *GoalState) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// Deserialize loads state from JSON bytes
func (s *GoalState) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}

// GetCurrentStep returns the current step index
func (s *GoalState) GetCurrentStep() int {
	return s.CurrentStep
}

// SetCurrentStep sets the current step index
func (s *GoalState) SetCurrentStep(step int) {
	if step >= 0 && step < s.TotalSteps {
		s.CurrentStep = step
	}
}

// GetTotalSteps returns the total number of steps
func (s *GoalState) GetTotalSteps() int {
	return s.TotalSteps
}

// IsStepCompleted checks if a step is completed
func (s *GoalState) IsStepCompleted(index int) bool {
	return s.CompletedSteps[index]
}

// MarkStepCompleted marks a step as completed
func (s *GoalState) MarkStepCompleted(index int) {
	s.CompletedSteps[index] = true
}

// BasicInfoStepData holds basic goal information
type BasicInfoStepData struct {
	Title       string
	Description string
	GoalType    models.GoalType
	valid       bool
}

// IsValid checks if the basic info data is valid
func (d *BasicInfoStepData) IsValid() bool {
	return d.valid && d.Title != ""
}

// GetData returns the underlying data
func (d *BasicInfoStepData) GetData() interface{} {
	return d
}

// SetData sets the underlying data
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

// IsValid checks if the field config data is valid
func (d *FieldConfigStepData) IsValid() bool {
	return d.valid && d.FieldType != ""
}

// GetData returns the underlying data
func (d *FieldConfigStepData) GetData() interface{} {
	return d
}

// SetData sets the underlying data
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

// IsValid checks if the scoring data is valid
func (d *ScoringStepData) IsValid() bool {
	return d.valid
}

// GetData returns the underlying data
func (d *ScoringStepData) GetData() interface{} {
	return d
}

// SetData sets the underlying data
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

// IsValid checks if the criteria data is valid
func (d *CriteriaStepData) IsValid() bool {
	return d.valid
}

// GetData returns the underlying data
func (d *CriteriaStepData) GetData() interface{} {
	return d
}

// SetData sets the underlying data
func (d *CriteriaStepData) SetData(data interface{}) error {
	if typedData, ok := data.(*CriteriaStepData); ok {
		*d = *typedData
		d.valid = true // Basic validation - enhance as needed
		return nil
	}
	return fmt.Errorf("invalid data type for CriteriaStepData")
}

// AIDEV-NOTE: Update step counts when adding new goal types or modifying flows
// Current step counts:
// - SimpleGoal: 4 steps (basic_info → scoring → criteria → confirmation) 
// - ElasticGoal: 8 steps (basic_info → field_config → scoring → mini → midi → maxi → validation → confirmation)
// - InformationalGoal: 3 steps (basic_info → field_config → confirmation)

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