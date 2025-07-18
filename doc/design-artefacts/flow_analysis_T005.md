# Habit Configuration Flow Analysis

**Task**: T005 Phase 2.0 - Flow Analysis and Enhancement Planning  
**Date**: Created during Phase 2.0 planning  
**Purpose**: Document current flows, identify enhancement opportunities, and plan bubbletea integration

## Current Implementation Analysis

### Overview
The current habit creation process uses sequential `huh.Form.Run()` calls:
- **Simple Habits**: 4-5 steps 
- **Elastic Habits**: 6-8 steps (most complex)
- **Informational Habits**: 3-4 steps

### Pain Points Identified
1. **No Progress Indication**: Users don't know how many steps remain
2. **No Back Navigation**: Can't return to previous step if error made
3. **No State Preservation**: Error recovery requires starting over
4. **Isolated Forms**: No shared context or validation across steps
5. **Static Experience**: Limited dynamic behavior within forms

---

## Flow Diagrams

### 1. Simple Habit Flow (4 steps)

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Step 1/4       │    │  Step 2/4       │    │  Step 3/4       │    │  Step 4/4       │
│  Basic Info     │───▶│  Scoring Type   │───▶│  Criteria       │───▶│  Confirmation   │
│                 │    │                 │    │  (if automatic) │    │  & Save         │
│ • Title         │    │ • Manual        │    │ • Boolean true  │    │ • Preview       │
│ • Description   │    │ • Automatic     │    │ • Description   │    │ • Validate      │
│ • Type: Simple  │    │                 │    │                 │    │ • Save to file  │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
                                   │                     │
                                   │                     │
                                   ▼                     ▼
                              Manual Scoring    Automatic Scoring
                              (Skip Step 3)     (Include Step 3)
```

**Flow Details:**
- **Step 1**: Title, Description, Habit Type (pre-selected as Simple)
- **Step 2**: Scoring Type (Manual/Automatic)
- **Step 3**: Criteria definition (only if Automatic scoring selected)
- **Step 4**: Preview complete habit configuration and save

**Current Issues:**
- Step 3 is conditional but no indication in UI
- No way to go back if wrong scoring type selected
- No preview of accumulated configuration until final step

---

### 2. Elastic Habit Flow (6-8 steps) - Most Complex

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Step 1/6-8     │    │  Step 2/6-8     │    │  Step 3/6-8     │    │  Step 4/6-8     │
│  Basic Info     │───▶│  Field Type     │───▶│  Field Config   │───▶│  Scoring Type   │
│                 │    │                 │    │  (conditional)  │    │                 │
│ • Title         │    │ • Numeric types │    │ • Unit          │    │ • Manual        │
│ • Description   │    │ • Duration      │    │ • Min/Max       │    │ • Automatic     │
│ • Type: Elastic │    │ • Time          │    │ • Multiline     │    │                 │
└─────────────────┘    │ • Text          │    │                 │    └─────────────────┘
                       └─────────────────┘    └─────────────────┘              │
                                                       │                       │
                                              ┌────────▼───────┐               │
                                              │ Field Details  │               │
                                              │ Needed?        │               │
                                              │ • Numeric: Yes │               │
                                              │ • Text: Yes    │               │
                                              │ • Boolean: No  │               │
                                              └────────────────┘               │
                                                                               │
                              ┌────────────────────────────────────────────────┴──────────┐
                              │                                                           │
                              ▼                                                           ▼
                    ┌─────────────────┐                                        ┌─────────────────┐
                    │  Step 5-7/6-8   │                                        │  Step 5/6-8     │
                    │  Criteria Def   │                                        │  Confirmation   │
                    │  (if automatic) │                                        │  & Save         │
                    │                 │                                        │                 │
                    │ • Mini Level    │──┐                                     │ • Preview       │
                    │ • Midi Level    │  │                                     │ • Validate      │
                    │ • Maxi Level    │  │    ┌─────────────────┐              │ • Save to file  │
                    └─────────────────┘  └───▶│  Step 8/8       │─────────────▶└─────────────────┘
                                              │  Validation     │
                                              │  & Preview      │
                                              │                 │
                                              │ • Check mini≤   │
                                              │   midi≤maxi     │
                                              │ • Validate all  │
                                              │ • Habit preview  │
                                              └─────────────────┘
```

**Flow Details:**
- **Step 1**: Basic info (Title, Description, Type: Elastic)
- **Step 2**: Field type selection (6 options: numeric types, duration, time, text)
- **Step 3**: Field configuration (conditional based on field type):
  - Numeric: Unit, Min/Max constraints
  - Text: Multiline option
  - Others: Skip this step
- **Step 4**: Scoring type (Manual/Automatic)
- **Steps 5-7**: Criteria definition (only if Automatic):
  - Step 5: Mini level criteria
  - Step 6: Midi level criteria  
  - Step 7: Maxi level criteria
- **Step 8**: Validation and preview (if criteria defined)
- **Final Step**: Confirmation and save

**Current Issues:**
- 6-8 steps with complex conditional flows
- Criteria validation (mini ≤ midi ≤ maxi) happens only at the end
- No way to compare criteria across levels during definition
- Long flow with no progress indication
- Easy to make mistakes with no recovery except restart

---

### 3. Informational Habit Flow (3 steps)

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Step 1/3       │    │  Step 2/3       │    │  Step 3/3       │
│  Basic Info     │───▶│  Field Config   │───▶│  Confirmation   │
│                 │    │                 │    │  & Save         │
│ • Title         │    │ • Field Type    │    │                 │
│ • Description   │    │ • Unit/Config   │    │ • Preview       │
│ • Type: Info    │    │ • Direction     │    │ • Validate      │
└─────────────────┘    └─────────────────┘    │ • Save to file  │
                                              └─────────────────┘
```

**Flow Details:**
- **Step 1**: Basic info (Title, Description, Type: Informational)
- **Step 2**: Field configuration:
  - Field type selection
  - Unit configuration (if numeric)
  - Direction (higher_better/lower_better/neutral)
- **Step 3**: Preview and save

**Current Issues:**
- Simplest flow but still lacks progress indication
- Direction selection could be better integrated with field type

---

## Decision Tree Diagrams

### 1. Habit Type Selection Decision Tree

```
                           ┌─────────────────┐
                           │   Habit Type     │
                           │   Selection     │
                           └─────────┬───────┘
                                     │
                 ┌───────────────────┼───────────────────┐
                 │                   │                   │
                 ▼                   ▼                   ▼
         ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
         │   Simple    │     │   Elastic   │     │Informational│
         │             │     │             │     │             │
         │ • Boolean   │     │ • 6 field   │     │ • 6 field   │
         │   only      │     │   types     │     │   types     │
         │ • Pass/Fail │     │ • Mini/Midi │     │ • Direction │
         │ • 4 steps   │     │   /Maxi     │     │ • 3 steps   │
         └─────────────┘     │ • 6-8 steps │     └─────────────┘
                             └─────────────┘
```

### 2. Field Type Selection Decision Tree (Elastic/Informational)

```
                    ┌─────────────────┐
                    │   Field Type    │
                    │   Selection     │
                    └─────────┬───────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌─────────────┐      ┌─────────────┐       ┌─────────────┐
│   Numeric   │      │    Time     │       │    Text     │
│             │      │             │       │             │
│ • uint      │      │ • time      │       │ • text      │
│ • decimal   │      │ • duration  │       │             │
│ • unsigned  │      │             │       │             │
└─────┬───────┘      └─────┬───────┘       └─────┬───────┘
      │                    │                     │
      ▼                    ▼                     ▼
┌─────────────┐      ┌─────────────┐       ┌─────────────┐
│Field Details│      │Field Details│       │Field Details│
│Required     │      │Skip         │       │Optional     │
│             │      │             │       │             │
│ • Unit      │      │ • No config │       │ • Multiline │
│ • Min/Max   │      │   needed    │       │   option    │
└─────────────┘      └─────────────┘       └─────────────┘
```

### 3. Scoring Type Decision Tree

```
                     ┌─────────────────┐
                     │  Scoring Type   │
                     │   Selection     │
                     └─────────┬───────┘
                               │
                    ┌──────────┼──────────┐
                    │                     │
                    ▼                     ▼
            ┌─────────────┐       ┌─────────────┐
            │   Manual    │       │  Automatic  │
            │             │       │             │
            │ • User sets │       │ • Criteria  │
            │   completion│       │   based     │
            │ • Skip      │       │ • Define    │
            │   criteria  │       │   rules     │
            └─────────────┘       └─────┬───────┘
                                        │
                        ┌───────────────┼───────────────┐
                        │                               │
                        ▼                               ▼
                ┌─────────────┐                 ┌─────────────┐
                │   Simple    │                 │   Elastic   │
                │             │                 │             │
                │ • 1 criteria│                 │ • 3 criteria│
                │   (boolean) │                 │   (mini/    │
                │             │                 │    midi/    │
                └─────────────┘                 │    maxi)    │
                                                └─────────────┘
```

---

## Enhancement Opportunities Analysis

### 1. Bubbletea Value vs Overhead Analysis

**High Value Cases (Recommend Bubbletea):**
- **Elastic Habits with Automatic Scoring** (6-8 steps)
  - **Value**: Progress tracking, criteria comparison, validation feedback
  - **Overhead**: Justified by complexity reduction
- **Habit Editing** (variable steps based on habit type)
  - **Value**: Live preview, better error recovery
  - **Overhead**: Worth it for improved UX

**Medium Value Cases (Consider Bubbletea):**
- **Simple Habits with Automatic Scoring** (4 steps)
  - **Value**: Progress indication, back navigation
  - **Overhead**: Moderate - could go either way
- **Habit Management Operations** (list/remove)
  - **Value**: Interactive selection, rich display
  - **Overhead**: Depends on implementation complexity

**Low Value Cases (Keep Simple Huh):**
- **Confirmations and Simple Prompts**
  - Single-step interactions where bubbletea adds little value
- **Informational Habits** (3 steps)
  - Simple enough that current flow is adequate

### 2. User Experience Improvements

**Navigation Enhancements:**
- **Progress Indicators**: "Step 3 of 6" with progress bar
- **Breadcrumbs**: Show completed steps with checkmarks
- **Back/Forward Navigation**: Allow correction of previous steps
- **Step Jumping**: Direct navigation to specific steps (advanced)
- **Cancel/Exit**: Graceful exit with optional save-as-draft

**Real-time Validation Enhancements:**
- **Live Field Validation**: Immediate feedback as user types
- **Cross-step Validation**: Check criteria ordering (mini ≤ midi ≤ maxi) in real-time
- **Contextual Help**: Dynamic help text based on current input
- **Field Highlighting**: Visual indication of validation status

**Habit Preview Enhancements:**
- **Side-by-side Preview**: Show accumulated habit config alongside current step
- **YAML Preview**: Live YAML generation for advanced users
- **Validation Status**: Real-time indication of completeness and correctness
- **Summary Cards**: Compact display of configured properties

### 3. Technical Integration Patterns

**Embedding Huh in Bubbletea:**
```go
type HuhFormStep struct {
    form *huh.Form
    onComplete func(result interface{}) tea.Cmd
    onCancel func() tea.Cmd
}

type WizardModel struct {
    steps []Step
    currentStep int
    state WizardState
    navigation NavigationController
}
```

**Standalone Huh (Keep for Simple Cases):**
```go
// Current pattern - keep for simple interactions
form := huh.NewForm(...)
if err := form.Run(); err != nil {
    return err
}
```

---

## API Interface Planning

### 1. Wizard State Management

```go
type WizardState interface {
    GetStep(index int) StepData
    SetStep(index int, data StepData) 
    Validate() []ValidationError
    ToHabit() (*models.Habit, error)
    Serialize() ([]byte, error)
    Deserialize([]byte) error
}

type StepHandler interface {
    Render(state WizardState) string
    Update(msg tea.Msg, state WizardState) (WizardState, tea.Cmd)
    Validate(state WizardState) []ValidationError
    CanNavigateFrom() bool
    CanNavigateTo() bool
}

type NavigationController interface {
    CanGoBack() bool
    CanGoForward() bool
    CanGoToStep(index int) bool
    GoBack() tea.Cmd
    GoForward() tea.Cmd
    GoToStep(index int) tea.Cmd
    Cancel() tea.Cmd
}
```

### 2. Form Embedding Patterns

```go
type HuhFormStep struct {
    form *huh.Form
    title string
    description string
    validator func(interface{}) []ValidationError
}

type FormRenderer interface {
    RenderForm(step HuhFormStep, state WizardState) string
    RenderProgress(current, total int) string
    RenderNavigation(nav NavigationController) string
    RenderSummary(state WizardState) string
}

type ValidationCollector interface {
    CollectErrors(state WizardState) []ValidationError
    ValidateStep(stepIndex int, data StepData) []ValidationError
    ValidateCrossStep(state WizardState) []ValidationError
}
```

### 3. Progress Tracking APIs

```go
type ProgressTracker interface {
    GetCurrentStep() int
    GetTotalSteps() int
    GetCompletedSteps() []int
    GetStepStatus(index int) StepStatus
    MarkStepCompleted(index int)
    MarkStepInProgress(index int)
}

type StepValidator interface {
    ValidateForCompletion(data StepData) bool
    ValidateForNavigation(data StepData) bool
    GetValidationErrors(data StepData) []ValidationError
}

type StateSerializer interface {
    SaveDraft(state WizardState) error
    LoadDraft() (WizardState, error)
    ClearDraft() error
    HasDraft() bool
}
```

### 4. Error Recovery Mechanisms

```go
type StateSnapshot interface {
    TakeSnapshot(state WizardState) SnapshotID
    RestoreSnapshot(id SnapshotID) (WizardState, error)
    ListSnapshots() []SnapshotInfo
    DeleteSnapshot(id SnapshotID) error
}

type ErrorHandler interface {
    HandleValidationError(err ValidationError) tea.Cmd
    HandleSystemError(err error) tea.Cmd
    RecoverFromError(state WizardState) (WizardState, error)
}

type RetryStrategy interface {
    ShouldRetry(err error) bool
    GetRetryDelay(attempt int) time.Duration
    GetMaxRetries() int
}
```

---

## Implementation Strategy for Elastic Habits

### 1. Complex Criteria Validation Flow

**Current Problem:**
- Criteria validation (mini ≤ midi ≤ maxi) only happens at the end
- No comparison view between levels during definition
- Error recovery requires restarting entire flow

**Enhanced Solution:**
```
Step 4: Mini Level         Step 5: Midi Level         Step 6: Maxi Level
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│ Define Mini     │       │ Define Midi     │       │ Define Maxi     │
│                 │       │                 │       │                 │
│ Value: [____]   │       │ Value: [____]   │       │ Value: [____]   │
│                 │───────▶│                 │───────▶│                 │
│ ✓ Mini: Valid   │       │ ⚠ Midi ≥ Mini   │       │ ⚠ Maxi ≥ Midi   │
│                 │       │   (real-time)   │       │   (real-time)   │
└─────────────────┘       └─────────────────┘       └─────────────────┘
                                    │                         │
                                    ▼                         ▼
                          ┌─────────────────┐       ┌─────────────────┐
                          │ Sidebar:        │       │ Sidebar:        │
                          │ Mini: 10        │       │ Mini: 10        │
                          │ Midi: [current] │       │ Midi: 25        │
                          │ Maxi: [pending] │       │ Maxi: [current] │
                          └─────────────────┘       └─────────────────┘
```

### 2. Dynamic Field Configuration

**Current Problem:**
- Field configuration step is conditionally shown
- No indication of what configuration is needed until you get there

**Enhanced Solution:**
- Show preview of required configuration fields based on field type selection
- Progressive disclosure with expandable sections
- Real-time validation of field constraints

### 3. Progressive Disclosure Patterns

**Current Problem:**
- All options presented at once, can be overwhelming

**Enhanced Solution:**
```
Field Type Selection:
┌─────────────────────────────────────────────────┐
│ ○ Number (unsigned integer)                    │
│   └─ Will need: Unit, Min/Max constraints      │
│ ○ Number (decimal)                             │
│   └─ Will need: Unit, Min/Max constraints      │
│ ○ Duration (e.g., 30m, 1h30m)                 │
│   └─ Will need: Unit configuration             │
│ ○ Time (e.g., 14:30)                          │
│   └─ No additional configuration needed        │
│ ○ Text                                         │
│   └─ Will need: Multiline option              │
└─────────────────────────────────────────────────┘
```

### 4. State Persistence Between Steps

**Current Problem:**
- No way to save progress if interrupted
- Error recovery requires complete restart

**Enhanced Solution:**
```go
type ElasticHabitWizardState struct {
    // Step 1: Basic Info
    Title       string
    Description string
    HabitType    models.HabitType // Always Elastic
    
    // Step 2: Field Type
    FieldType   string
    
    // Step 3: Field Configuration
    Unit        string
    Min         *float64
    Max         *float64
    Multiline   bool
    
    // Step 4: Scoring
    ScoringType models.ScoringType
    
    // Steps 5-7: Criteria (if automatic)
    MiniCriteria *CriteriaConfig
    MidiCriteria *CriteriaConfig
    MaxiCriteria *CriteriaConfig
    
    // Metadata
    CurrentStep     int
    CompletedSteps  []int
    ValidationErrors map[int][]ValidationError
    LastSaved       time.Time
}
```

---

## Next Steps

Based on this analysis, the recommended implementation approach is:

1. **Start with Elastic Habit Wizard**: Highest complexity, highest value from bubbletea
2. **Create Hybrid Architecture**: Bubbletea for wizards, huh for simple forms
3. **Build Reusable Components**: Progress indicators, navigation, form embedding
4. **Implement Progressive Disclosure**: Show configuration requirements upfront
5. **Add Real-time Validation**: Immediate feedback and cross-step validation
6. **Create Rich Preview**: Side-by-side habit configuration display

The elastic habit flow is the most complex and will benefit most from the enhanced UX patterns, making it an ideal proof-of-concept for the bubbletea integration.

---

## Phase 2.6-2.7 Implementation Update

### Current Status (Post-User Testing)

**Phase 2.6**: Successfully implemented basic info collection upfront (Title → Description → Habit Type).

**Phase 2.7**: Critical integration issues discovered during user testing:

### User Testing Results Analysis

**Test Case 1: Enhanced Wizard Flow**
- ✅ Basic info collection works correctly (Title → Description → Habit Type)
- ❌ Mode selection defaults to "Quick Forms" instead of "Enhanced Wizard (Recommended)"
- ❌ Enhanced wizard shows validation error: "Scoring configuration is required"
- **Root Cause**: Wizard step handlers don't recognize pre-populated basic info

**Test Case 2: Quick Forms Flow**  
- ✅ Basic info collection works correctly
- ❌ Legacy forms show validation error: "Basic information is required"
- **Root Cause**: HabitBuilder.BuildHabit() still tries to collect basic info

### Corrected Implementation Plan

**Immediate Fixes Required:**

1. **Remove Mode Selection Complexity**
   - Eliminate user choice between Enhanced Wizard vs Quick Forms
   - Use `determineOptimalInterface()` automatically based on habit type
   - Simplify flow: Basic Info → Auto-select best interface → Launch

2. **Fix Wizard Pre-population Logic**
   - Ensure wizard step handlers properly validate pre-populated step 0
   - Fix step navigation to start from appropriate step based on completion status
   - Update validation logic to recognize completed basic info

3. **Fix Legacy Forms Integration**
   - Modify HabitBuilder to skip basic info collection when pre-populated
   - Ensure BuildHabitWithBasicInfo() actually uses the provided basic info
   - Maintain backwards compatibility for non-pre-populated flows

### Optimal Flow (Revised)

```
1. collectBasicInformation():
   ├─ Title (required, validated)
   ├─ Description (optional)
   └─ Habit Type (simple/elastic/informational)

2. determineOptimalInterface() automatically:
   ├─ Simple Habits → Enhanced Wizard (no user choice)
   ├─ Elastic Habits → Enhanced Wizard (complexity requires it)
   └─ Informational Habits → Enhanced Wizard (direction config needs it)

3. Launch appropriate interface:
   ├─ Enhanced Wizard: Pre-populated with basic info, start from step 1
   └─ Legacy Forms: Skip basic info collection, use pre-populated data
```

**User Experience Habits:**
- No confusing mode selection prompts
- Seamless transition from basic info to appropriate interface
- No validation errors about missing information that was already collected
- Optimal interface automatically selected based on habit complexity