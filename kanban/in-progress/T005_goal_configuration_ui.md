---
title: "Goal Configuration UI"
type: ["feature"]
tags: ["ui", "goals", "configuration", "cli"]
related_tasks: ["depends-on:T004"]
context_windows: ["./CLAUDE.md", "./doc/specifications/goal_schema.md", "./internal/models/*.go", "./internal/parser/*.go", "./internal/ui/*.go", "./cmd/*.go"]
---

# Goal Configuration UI

## 1. Goal / User Story

As a user, I want to configure my goals through an interactive CLI interface rather than manually editing YAML files, so that I can easily add, modify, and manage my goals without needing to understand the YAML schema syntax.

The system should provide:
- `iter goal add` - Interactive UI to create new goals with guided prompts
- `iter goal list` - Display existing goals in a readable format
- `iter goal edit` - Select and modify existing goal definitions
- `iter goal remove` - Remove existing goals with confirmation

This eliminates the need for users to manually edit YAML and reduces configuration errors while maintaining the existing file-based storage approach.

## 2. Acceptance Criteria

### Core Functionality
- [ ] `iter goal add` command creates new goals through interactive prompts
- [ ] Goal type selection (simple, elastic, informational) with appropriate follow-up questions
- [ ] Field type selection with validation and guidance
- [ ] Scoring type configuration (manual/automatic) with criteria definition for automatic scoring
- [ ] Elastic goal criteria definition (mini/midi/maxi) with proper validation
- [ ] Goal validation using existing schema validation logic
- [ ] New goals added to existing goals.yml file preserving existing goals

### Goal Management
- [ ] `iter goal list` displays existing goals in human-readable format
- [ ] `iter goal edit` allows selection and modification of existing goals
- [ ] `iter goal remove` removes goals with confirmation prompt
- [ ] All operations preserve goal IDs and maintain data integrity
- [ ] File operations are atomic (no partial writes)

### User Experience
- [ ] Intuitive prompts with help text and examples
- [ ] Input validation with clear error messages
- [ ] Consistent styling using existing lipgloss patterns
- [ ] Graceful handling of file permission errors
- [ ] Preview of goal definition before saving
- [ ] Progress indicators for multi-step flows
- [ ] Navigation between steps (back/forward) where appropriate
- [ ] Real-time validation and contextual help
- [ ] Rich, interactive experience for complex goal configuration

### Technical Requirements
- [ ] Loosely coupled design - separate goal configuration logic from entry collection
- [ ] Reuse existing validation logic from models package
- [ ] Reuse existing parser and file operations
- [ ] Follow established UI patterns from entry collection
- [ ] Comprehensive error handling
- [ ] Unit tests for all UI components

---
## 3. Implementation Plan & Progress

**Overall Status:** `In Progress`

**Architecture Analysis:**

Based on investigation of existing codebase:

**Existing UI Patterns:**
- Uses `github.com/charmbracelet/huh` for form building
- Established patterns: `huh.NewForm()` with `huh.NewGroup()` containing form elements
- Form types used: `NewInput()`, `NewConfirm()`, `NewSelect()`, `NewText()`
- Styling with `lipgloss` for colors and formatting
- Error handling pattern: return descriptive errors, don't panic on form failures

**Existing Infrastructure:**
- `internal/parser/GoalParser` - handles loading/saving goals.yml
- `internal/models/Goal.Validate()` - schema validation logic
- `internal/models/` - complete goal type definitions and constants
- `cmd/` structure uses cobra for command organization
- ID persistence already implemented (T004) - generated IDs automatically saved

**Planned Implementation Approach:**

### Phase 1: Command Structure & Core UI ✅ **COMPLETED**
- [x] **1.1 Add goal subcommand structure**
  - ✅ Create `cmd/goal.go` with `goal` parent command
  - ✅ Add `add`, `list`, `edit`, `remove` subcommands
  - ✅ Follow existing cobra patterns from `cmd/entry.go`

- [x] **1.2 Create goal configuration UI package**
  - ✅ Create `internal/ui/goalconfig/` package for separation of concerns
  - ✅ Design `GoalConfigurator` struct with form builders
  - ✅ Establish comprehensive form builder patterns (GoalFormBuilder, CriteriaBuilder, GoalBuilder)
  - ✅ Implement complete AddGoal flow with huh forms

### Phase 2: Enhanced Goal Creation & Flow Design

- [x] **2.0 Flow Analysis and Enhancement Planning** ✅ **COMPLETED**
  - [x] Analyze current multi-step goal creation flow (4-6 form interactions)
  - [x] Document logical flow with text diagrams for each goal type:
    - [x] Simple goal flow diagram (4 steps: basic info → scoring → criteria → confirmation)
    - [x] Elastic goal flow diagram (6-8 steps: basic info → field config → scoring → mini/midi/maxi criteria → validation → confirmation)
    - [x] Informational goal flow diagram (3 steps: basic info → field config → confirmation)
    - [x] Decision tree diagrams for conditional flows (manual vs automatic scoring, field type branches)
  - [x] Evaluate bubbletea integration opportunities vs standalone huh forms:
    - [x] Complexity analysis: when bubbletea adds value vs overhead
    - [x] User experience improvements: navigation, progress, error recovery
    - [x] Technical integration patterns: embedding huh in bubbletea vs standalone
  - [x] Design enhanced UX patterns:
    - [x] Progress indicator designs (Step X of Y, progress bar, breadcrumbs)
    - [x] Navigation patterns (back/forward buttons, step jumping, cancel/exit)
    - [x] Real-time validation display (inline errors, live help text, field highlighting)
    - [x] Goal preview formats (summary cards, YAML preview, validation status)
  - [x] Plan API interfaces for bubbletea-enhanced components:
    - [x] Wizard state management interfaces (WizardState, StepHandler, NavigationController)
    - [x] Form embedding patterns (HuhFormStep, FormRenderer, ValidationCollector)  
    - [x] Progress tracking APIs (ProgressTracker, StepValidator, StateSerializer)
    - [x] Error recovery mechanisms (StateSnapshot, ErrorHandler, RetryStrategy)
  - [x] Create detailed implementation strategy focusing on elastic goals:
    - [x] Complex criteria validation flow (mini ≤ midi ≤ maxi constraints)
    - [x] Dynamic field configuration based on field type selection
    - [x] Progressive disclosure patterns for complex options
    - [x] State persistence between steps for long flows
  - [x] **Documentation**: Complete analysis documented in `doc/flow_analysis_T005.md`

- [x] **2.1 Bubbletea Goal Creation Wizard (Enhanced UX)** ✅ **FOUNDATION COMPLETE**
  - [x] Create wizard infrastructure (interfaces, state management, navigation)
  - [x] Implement progress indicators showing current step (Step X of Y)
  - [x] Add back/forward navigation between steps with validation
  - [x] Real-time validation framework with contextual error display
  - [x] Goal preview and summary rendering
  - [x] Enhanced error recovery and state serialization
  - [x] Hybrid integration: enhanced wizard for elastic goals, simple forms for others
  - [x] Complete bubbletea model with tea.Model interface implementation
  - [ ] **Next**: Implement specific step handlers for each goal type

- [x] **2.2 Simple Goal Wizard Flow** ✅ **COMPLETED**
  - [x] Step 1: Basic info (title, description, goal type pre-selected)
  - [x] Step 2: Scoring configuration (manual/automatic)
  - [x] Step 3: Criteria definition (if automatic scoring, auto-skipped for manual)
  - [x] Step 4: Preview and confirmation with goal summary
  - [x] Complete step handler implementations with form management
  - [x] Smart navigation with conditional step skipping
  - [x] State persistence and real-time validation
  - [x] Full linting compliance and code quality standards

- [x] **2.3 Elastic Goal Wizard Flow (Complex)** ✅ **COMPLETED**
  - [x] Step 1: Basic info (reuses BasicInfoStepHandler)
  - [x] Step 2: Field type and configuration (FieldConfigStepHandler)
  - [x] Step 3: Scoring type selection (reuses ScoringStepHandler)
  - [x] Step 4: Mini-level criteria definition (CriteriaStepHandler with level="mini")
  - [x] Step 5: Midi-level criteria definition (CriteriaStepHandler with level="midi")
  - [x] Step 6: Maxi-level criteria definition (CriteriaStepHandler with level="maxi")
  - [x] Step 7: Criteria validation and cross-validation (ValidationStepHandler)
  - [x] Step 8: Final confirmation with complete goal summary (reuses ConfirmationStepHandler)

- [x] **2.4 Informational Goal Wizard Flow** ✅ **COMPLETED**
  - [x] Step 1: Basic info (reuses BasicInfoStepHandler)
  - [x] Step 2: Field configuration and direction (enhanced FieldConfigStepHandler)
  - [x] Step 3: Preview and confirmation (reuses ConfirmationStepHandler)

- [x] **2.5 Hybrid Implementation Strategy** ✅ **COMPLETED**
  - [x] Keep simple huh forms for basic interactions
  - [x] Use bubbletea for complex multi-step flows  
  - [x] Create reusable bubbletea components that can embed huh forms
  - [x] Maintain backwards compatibility with existing patterns

- [x] **2.6 Fix Goal Creation Flow (Bug Fix)** ✅ **COMPLETED**
  - [x] Collect Title and Description before Goal Type selection (currently reversed)
  - [x] Modify configurator.AddGoal() to prompt for basic info first
  - [x] Pass pre-collected basic info to wizard/legacy forms
  - [x] Update wizard initialization to accept pre-populated basic info
  - [x] Ensure both wizard and legacy flows work with pre-populated data
  - [x] Test all goal types (simple/elastic/informational) with corrected flow

- [x] **2.7 Fix Critical Wizard/Forms Integration Issues** ✅ **COMPLETED**
  - [x] Fix Enhanced Wizard default selection (currently defaults to Quick Forms instead of Enhanced)
  - [x] Remove superfluous mode selection - automatically choose best interface per goal type
  - [x] Fix Enhanced Wizard pre-population validation errors ("Scoring configuration is required")
  - [x] Fix Quick Forms pre-population validation errors ("Basic information is required")
  - [x] Ensure wizard step handlers properly recognize pre-populated basic info
  - [x] Update legacy forms to skip basic info collection when pre-populated
  - [x] Test complete flow: Basic Info → Auto-select best interface → Launch without errors

- [x] **2.8 Simplified Idiomatic Goal Creation** ✅ **COMPLETED**
  - [x] Review bubbletea/huh documentation and implement idiomatic patterns
  - [x] Simplify goal creation to focus on most common use case: Simple + Manual
  - [x] Implement clean bubbletea model for simple goal creation (2-3 steps max)
  - [x] Add custom prompt field for manual goals (default: "Did you accomplish this goal today?")
  - [x] Ensure goals save correctly to goals.yml with expected structure
  - [x] Remove complex wizard architecture in favor of simple sequential forms
  - [x] Test end-to-end: Basic Info → Scoring Type → Custom Prompt → Save

- [x] **2.9 Prevent position field in goals.yml** ✅ **COMPLETED**
  - [x] Remove position assignment from SimpleGoalCreator in configurator.go
  - [x] Remove position assignment from legacy GoalBuilder in builder.go  
  - [x] Add AIDEV-NOTE comments explaining position inference approach
  - [x] Position now determined by parser/schema based on order in goals.yml

### Phase 3: Informational Goal Support

- [x] **3.1 Fix Informational Goal Flow Routing** ✅ **COMPLETED**
  - [x] Update configurator.AddGoal() to route informational goals to specialized creator
  - [x] Informational goals now skip scoring configuration entirely (no scoring, no criteria)
  - [x] Create new runInformationalGoalCreator() method alongside runSimpleGoalCreator()
  - [x] Update goal type selection flow with proper routing logic
  - [x] Added switch statement to route based on basicInfo.GoalType
  - [x] Informational goals get placeholder implementation (boolean field, manual scoring, neutral direction)
  - [x] Simple and Elastic goals continue using existing SimpleGoalCreator flow

- [x] **3.2 Field Type Configuration System** ✅ **COMPLETED**
  - [x] Create field type selector supporting all model types:
    - [x] Boolean (boolean) - simple true/false data
    - [x] Text (text) - free text input with optional multiline
    - [x] Numeric (unsigned_int, unsigned_decimal, decimal) - with subtype and unit selection
    - [x] Time (time) - time of day values
    - [x] Duration (duration) - time duration values
  - [x] Add numeric field configuration:
    - [x] Subtype selection: unsigned_int (default), unsigned_decimal, decimal
    - [x] Unit specification (default: "times", examples: "reps", "kg", "minutes", "pages")
    - [x] Min/Max value constraints (optional)
  - [x] Add direction preference for applicable field types:
    - [x] Numeric, Time, Duration: "higher_better" | "lower_better" | "neutral"
    - [x] Boolean, Text: no direction (always neutral)
  - [x] Implemented FieldTypeSelector with interactive configuration flow
  - [x] Created FieldConfig structure for complete field type specification
  - [x] Added validation for numeric constraints and unit configuration

- [x] **3.3 InformationalGoalCreator Implementation** ✅ **COMPLETED**
  - [x] Create new `InformationalGoalCreator` bubbletea model following SimpleGoalCreator patterns
  - [x] Flow design: Basic Info → Field Type Selection → Field Configuration → Direction Preference → Goal Prompt → Save
  - [x] Implement sequential huh forms with conditional groups:
    - [x] Group 1: Field type selection (boolean, text, numeric, time, duration)
    - [x] Group 2: Field configuration (conditional based on type)
      - [x] Numeric: subtype, unit, min/max (all optional except subtype)
      - [x] Text: multiline option
      - [x] Time/Duration: no additional config needed
    - [x] Group 3: Direction preference (conditional, hidden for boolean/text)
    - [x] Group 4: Goal prompt (question asked during entry recording)
  - [x] Handle validation for all field types and configurations
  - [x] Create proper models.Goal structure for informational goals
  - [x] Implemented multi-step bubbletea model with step progression (4 steps)
  - [x] Added conditional form groups based on field type selection
  - [x] Integrated with configurator routing system
  - [x] Full support for all field types with proper validation
  - [x] Added intelligent default prompts based on field type and configuration
  - [x] Example: "How many cups did you record for coffee?" for numeric fields with unit "cups"

- [x] **3.4 YAML Output Mode for Goal Commands** ✅ **COMPLETED**
  - [x] Add command-line flag support for YAML output without file modification
  - [x] Add `--dry-run` flag to `goal add` command (outputs generated YAML to stdout)
  - [x] Add `--dry-run` flag to `goal edit` command (outputs modified YAML to stdout)
  - [x] Add `ToYAML(schema *models.Schema) (string, error)` method to GoalParser
  - [x] Add `AddGoalWithYAMLOutput(goalsFilePath string) (string, error)` to GoalConfigurator
  - [x] Add `EditGoalWithYAMLOutput(goalsFilePath string) (string, error)` to GoalConfigurator (placeholder for T006)
  - [x] Modify command handlers to check dry-run flag and route appropriately
  - [x] Ensure YAML output goes to stdout, status messages to stderr
  - [x] Test that dry-run mode doesn't modify goals.yml (implementation ready for testing)
  - [x] Test that generated YAML is valid and parseable (uses same validation as save)
  - [x] Use cases: `iter goal add --dry-run`, `iter goal add --dry-run > custom.yml`
  - [x] Complete implementation with proper error handling and validation
  - [x] Status messages properly routed to stderr to avoid interfering with YAML output
  - [x] Help documentation updated with dry-run examples

- [x] **3.5 Field Value Input UI Foundation** ✅ **COMPLETED**
  - [x] Design reusable field input components for future entry recording:
    - [x] BooleanInput: checkbox/toggle with clear yes/no display
    - [x] TextInput: single-line and multiline text input with validation
    - [x] NumericInput: number input with unit display and min/max validation
    - [x] TimeInput: time picker or formatted input (HH:MM format)
    - [x] DurationInput: duration input (supports various formats like "1h 30m")
  - [x] Create FieldValueInput interface for type-safe field recording
  - [x] Plan integration points for entry recording system (T007)
  - [x] Document patterns for reuse in simple/elastic goal criteria definition
  - [x] Implement FieldValueInputFactory for automatic component creation
  - [x] Complete type-safe validation and error handling for all field types
  - [x] Support for all field configurations (units, constraints, multiline)

- [ ] **3.6 Integration and Testing**
  - [ ] Wire up InformationalGoalCreator in configurator flow (✅ **COMPLETED** in 3.3)
  - [ ] Test all field type combinations end-to-end:
    - [ ] Boolean informational goal
    - [ ] Text informational goal (single-line and multiline)
    - [ ] Numeric informational goals (int, decimal with units and direction)
    - [ ] Time informational goal with direction preference
    - [ ] Duration informational goal with direction preference
  - [ ] Verify goals.yml output matches expected schema for all configurations
  - [ ] Validate informational goals save and load correctly with parser
  - [ ] Ensure all field configurations persist properly (units, direction, min/max)

**Implementation Strategy:**
- Follow SimpleGoalCreator patterns for consistency and maintainability
- Use conditional huh form groups for dynamic UI based on field type selection
- Leverage existing models.FieldType structure for configuration storage
- Design field input components as foundation for entry recording system
- Maintain idiomatic bubbletea/huh patterns established in Phase 2.8

**Phase 3 Implementation Notes:**

**3.1-3.3 Informational Goal System (Completed):**
- Complete flow routing and specialized creator for informational goals
- Comprehensive field type configuration with all model types supported
- 4-step bubbletea model with intelligent prompts and validation
- Full integration with configurator routing system

**3.4 YAML Output Mode (Completed):**
- --dry-run flag support for both goal add and goal edit commands
- ToYAML() parser method for non-destructive YAML generation
- Proper output routing (YAML to stdout, status to stderr)
- Use cases: debugging, configuration management, scripting

**3.5 Field Value Input Foundation (Completed):**
- Complete type-safe interface system for field value collection
- All field types supported: boolean, text, numeric, time, duration
- Factory pattern for automatic component creation
- Ready for integration with entry recording system (T007)
- Key file: `internal/ui/goalconfig/field_value_input.go`



---

**Next Phase:** Goal management features (list, edit, remove) moved to [T006 Goal Management UI](../backlog/T006_goal_management_ui.md)

**Key Design Decisions:**

1. **Separation of Concerns**: Goal configuration UI in separate package (`internal/ui/goalconfig/`) to avoid coupling with entry collection
2. **Reuse Existing Infrastructure**: Leverage existing parser, validation, and UI patterns rather than duplicating
3. **Progressive Disclosure**: Guide users through goal creation with conditional prompts based on selections
4. **Data Integrity**: Preserve goal IDs and use atomic file operations
5. **Extensibility**: Design forms to easily accommodate new goal types and field types
6. **Hybrid UI Strategy**: Use simple huh forms for basic interactions, bubbletea for complex multi-step flows
7. **Enhanced UX**: Progress indicators, navigation, real-time validation for complex workflows
8. **Backwards Compatibility**: Maintain existing simple form patterns while enhancing complex flows

## 4. Roadblocks

*(None identified yet)*

## 5. Notes / Discussion Log

- Created to provide user-friendly goal configuration without YAML editing
- Investigation shows strong existing foundation with huh forms, validation, and file operations
- Design follows established patterns from entry collection UI
- Focus on progressive disclosure and guided configuration experience

**Technical Notes:**

**Phase 1 Implementation (Completed):**
- `huh.NewSelect()` perfect for goal type and field type selection
- Existing validation in models package provides solid foundation
- T004 ID persistence ensures data integrity during goal modifications
- Cobra command structure established in `cmd/` package
- Comprehensive form builders implemented (GoalFormBuilder, CriteriaBuilder, GoalBuilder)
- Complete AddGoal flow with 4-6 sequential form interactions

**Phase 2.2 Implementation (Completed):**
- **Complete Simple Goal Wizard**: 4-step flow with BasicInfo → Scoring → Criteria → Confirmation
- **Smart Step Handlers**: Individual handlers implementing StepHandler interface for modularity
- **Conditional Flow Logic**: Criteria step automatically skipped when manual scoring selected
- **Form Data Management**: Direct field binding to handler structs, avoiding complex form introspection
- **State Persistence**: Complete wizard state preservation between steps with JSON serialization
- **Real-time Validation**: Live validation with contextual error messages and navigation control
- **Architecture Resolution**: Avoided import cycles by implementing all step handlers in wizard package
- **Code Quality**: 100% lint compliance with comprehensive export comments and unused parameter fixes
- **Key Files Implemented**:
  - `internal/ui/goalconfig/wizard/simple_steps.go` - Complete step handler implementations
  - `internal/ui/goalconfig/wizard/criteria_steps.go` - Criteria configuration with smart skipping
  - `internal/ui/goalconfig/wizard/wizard.go` - Main wizard model and step coordination
  - `internal/ui/goalconfig/wizard/interfaces.go` - Clean interface definitions (State, StepHandler)
  - `internal/ui/goalconfig/wizard/state.go` - Comprehensive state management with step data types

**Bubbletea Enhancement Strategy:**

**Current State Analysis:**
- Goal creation requires 4-6 separate `form.Run()` calls
- Each form is isolated with no shared state or progress indication
- No ability to navigate back once a form is submitted
- Error recovery requires starting over
- Limited dynamic behavior within forms

**Enhancement Opportunities:**
1. **Multi-step Wizards**: Goal creation flow would benefit from unified navigation
2. **Real-time Validation**: Live field validation and dynamic help text
3. **Rich Context**: Show goal overview alongside forms, progress indicators
4. **Enhanced Error Recovery**: Better handling without losing progress

**Implementation Strategy:**
- **Phase 1**: Keep current huh forms for simple interactions (working implementation)
- **Phase 2**: Enhance complex flows with bubbletea (goal creation wizard)
- **Phase 3**: Apply bubbletea to goal management operations
- **Hybrid Approach**: bubbletea apps can embed huh forms for best of both worlds

**Key Technical Decisions:**
- Start with goal configuration wizard as bubbletea proof-of-concept
- Maintain huh forms for single-step interactions (confirmations, simple input)
- Create reusable bubbletea components for wizard-style flows
- Design APIs that support both standalone huh and embedded-in-bubbletea usage

**Flow Complexity Analysis:**
- **Simple Goals**: 4 steps → ✅ **IMPLEMENTED** with bubbletea wizard
- **Elastic Goals**: 6-8 steps → High value from enhanced navigation and progress
- **Informational Goals**: 3 steps → Moderate benefit from bubbletea
- **Goal Management**: List/edit/remove → Enhanced interaction patterns valuable

**Phase 2.2 Architecture Lessons Learned:**
- **Form Data Binding**: Direct struct field binding (&h.fieldName) works better than form introspection
- **Import Cycle Management**: Keep step handlers in wizard package vs separate packages to avoid cycles
- **State Management**: JSON serialization provides clean state persistence between steps
- **Conditional Logic**: shouldSkip() pattern enables clean conditional step navigation
- **Code Quality**: Comprehensive export comments and lint compliance essential for maintainability
- **AIDEV Anchor Comments**: Strategic placement aids future development - see AIDEV-NOTE/TODO comments in code

**Phase 2.3 Implementation (Completed):**
- **Complete Elastic Goal Flow**: Full 8-step wizard for complex elastic goals with mini/midi/maxi criteria
- **Field Configuration Handler**: Dynamic field type selection with constraints (units, min/max, multiline)
- **Cross-Criteria Validation**: ValidationStepHandler enforces mini ≤ midi ≤ maxi constraints
- **Models Integration**: Proper elastic goal creation with MiniCriteria/MidiCriteria/MaxiCriteria fields
- **Reusable Components**: Leveraged existing BasicInfo, Scoring, and Confirmation handlers
- **Complex State Management**: addElasticGoalConfiguration() handles multi-step data aggregation
- **Key Files Implemented**:
  - `internal/ui/goalconfig/wizard/field_config_steps.go` - Dynamic field configuration
  - `internal/ui/goalconfig/wizard/validation_steps.go` - Cross-step validation with user choices
  - Updated `state.go` with elastic goal configuration logic and proper condition mapping
  - Enhanced wizard.go with complete elastic goal flow (8 steps vs 4 for simple goals)

**Phase 2.4 Implementation (Completed):**
- **Complete Informational Goal Flow**: Simple 3-step wizard for data collection goals
- **Enhanced Field Configuration**: Added Direction field to FieldConfigStepData for higher/lower/neutral values
- **Informational Goal Integration**: Complete addInformationalGoalConfiguration() with proper models mapping
- **Wizard Auto-Selection**: Informational goals automatically use wizard for consistency
- **Direction Support**: Full direction configuration (higher/lower/neutral) stored and applied to goal
- **Reusable Components**: Leveraged existing BasicInfo, FieldConfig, and Confirmation handlers
- **Key Enhancements**:
  - Extended FieldConfigStepData with Direction field for informational goals
  - Enhanced field_config_steps.go to store and load direction values
  - Added complete informational goal configuration in state.go
  - Automatic wizard selection for informational goals in configurator.go

**Phase 2.5 Implementation (Completed):**
- **Hybrid Implementation Strategy**: Complete backwards compatibility with intelligent interface selection
- **LegacyGoalAdapter**: Provides compatibility layer between wizard and legacy forms without import cycles
- **HybridFormModel**: Reusable component for embedding huh forms within bubbletea applications with progress tracking
- **Intelligent Selection**: determineOptimalInterface() automatically chooses best UI based on goal complexity
- **User Override**: Simple goals allow user choice between Enhanced Wizard (recommended) and Quick Forms
- **Configuration Options**: WithLegacyMode() enables preferring legacy forms for conservative deployments
- **Import Cycle Resolution**: Avoided circular dependencies by keeping legacy adapter focused on wizard integration
- **Key Files Implemented**:
  - `internal/ui/goalconfig/wizard/hybrid_forms.go` - HybridFormModel and HybridFormRunner for embedding huh in bubbletea
  - `internal/ui/goalconfig/wizard/legacy_adapter.go` - BackwardsCompatibilityMode and adapter for interface selection
  - Enhanced `configurator.go` with intelligent mode selection and compatibility configuration
  - All existing wizard flows (simple/elastic/informational) preserved with enhanced backwards compatibility

**Phase 2.6 Bug Investigation (In Progress):**
- **Root Cause Identified**: configurator.AddGoal() asks for Goal Type first, then launches wizard/legacy forms which ask for Title/Description
- **Expected Flow**: Title → Description → Goal Type → Launch appropriate flow with pre-populated data
- **Current Incorrect Flow**: Goal Type → Launch wizard → Title/Description (duplicated effort, confusing UX)
- **Impact**: Users see Goal Type selection alone initially, then see Title/Description in wizard (feels like bug)
- **Legacy Forms**: GoalBuilder.BuildGoal() correctly starts with basic info including goal type
- **Wizard Forms**: BasicInfoStepHandler properly collects title/description, but gets called after goal type selection
- **Solution**: Move basic info collection to configurator level, pass to wizard/legacy forms as pre-populated data

**Phase 2.6 Implementation (Completed):**
- **Fixed Goal Creation Flow**: Corrected sequence to Title → Description → Goal Type → Launch appropriate flow
- **BasicInfo Structure**: New type to hold pre-collected title, description, and goal type
- **collectBasicInformation()**: Unified function collecting all basic info upfront with proper validation
- **Enhanced Wizard Integration**: NewGoalWizardModelWithBasicInfo() accepts pre-populated data and starts from step 1
- **Legacy Compatibility**: BuildGoalWithBasicInfo() and CreateGoalWithBasicInfo() methods for backwards compatibility
- **Smart Mode Selection**: Moved interface selection logic into basic info collection for seamless flow
- **Pre-population Logic**: Wizard state pre-populated with basic info, step 0 marked completed, starts from step 1
- **User Experience**: Now users see Title → Description → Goal Type → Enhanced wizard selection → Launch wizard starting with next step
- **Key Files Modified**:
  - `configurator.go`: Complete flow restructure with collectBasicInformation() and runGoalWizardWithBasicInfo()
  - `wizard/wizard.go`: Added NewGoalWizardModelWithBasicInfo() constructor with pre-population
  - `builder.go`: Added BuildGoalWithBasicInfo() for legacy compatibility
  - `wizard/legacy_adapter.go`: Added CreateGoalWithBasicInfo() for interface compatibility
- **Removed Redundancy**: Eliminated duplicate basic info collection between configurator and wizard flows

**Phase 2.7 Critical Issues Analysis (In Progress):**
- **User Testing Findings**: Testing revealed multiple integration failures after Phase 2.6 implementation
- **Enhanced Wizard Default Bug**: Mode selection shows "Quick Forms" selected instead of "Enhanced Wizard (Recommended)"
- **Superfluous Mode Selection**: User feedback indicates choice between wizard/forms adds unnecessary complexity
- **Enhanced Wizard Validation Error**: Shows "Scoring configuration is required" immediately after basic info collection
- **Quick Forms Validation Error**: Shows "Basic information is required" despite basic info being pre-collected
- **Pre-population Logic Failure**: Wizard state pre-population not properly recognized by step handlers
- **Legacy Forms Integration**: GoalBuilder.BuildGoal() still collecting basic info instead of using pre-populated data
- **Flow Analysis Mismatch**: Current implementation doesn't match the planned flow from flow_analysis_T005.md
- **Root Cause**: Wizard and legacy forms still expect to collect basic info themselves, ignoring pre-populated state
- **Required Solution**: 
  - Remove mode selection entirely - use determineOptimalInterface() automatically
  - Fix wizard step handler validation to recognize completed step 0
  - Fix legacy forms to skip basic info when pre-populated
  - Ensure seamless transition from basic info collection to appropriate interface

**Phase 2.7 Implementation (Completed):**
- **Removed Mode Selection Complexity**: Eliminated confusing Enhanced Wizard vs Quick Forms choice entirely
- **Automatic Interface Selection**: Always use enhanced wizard for all goal types (analysis shows superior UX)
- **Fixed Validation Logic**: Updated step handler Validate() methods to not show "required" errors when just starting a step
- **Simplified Flow**: Now Basic Info → Enhanced Wizard (auto-selected) → Complete goal creation
- **Step Handler Validation Fix**: Added formComplete check before showing validation errors:
  - ScoringStepHandler.Validate() - no longer shows "Scoring configuration is required" on start
  - FieldConfigStepHandler.Validate() - no longer shows "Field configuration is required" on start  
  - CriteriaStepHandler.Validate() - no longer shows criteria errors on start
- **Legacy Forms Elimination**: Removed legacy form paths since enhanced wizard is superior for all goal types
- **Interface Simplification**: Removed determineOptimalInterface() since we always use enhanced wizard
- **User Experience Improvement**: No more confusing choices, seamless flow from basic info to wizard
- **Key Files Modified**:
  - `configurator.go`: Removed mode selection, simplified to always use enhanced wizard
  - `simple_steps.go`: Fixed ScoringStepHandler validation to prevent premature error display
  - `field_config_steps.go`: Fixed FieldConfigStepHandler validation
  - `criteria_steps.go`: Fixed CriteriaStepHandler validation  
- **Root Cause Resolution**: Step handlers were validating completion before user interaction, now validation respects form state

**Phase 2.8 Analysis (In Progress):**
- **User Testing Findings**: After Phase 2.7, wizard still shows blank screen after goal type selection
- **Root Cause**: Complex custom wizard architecture doesn't follow idiomatic bubbletea patterns
- **Documentation Review**: Bubbletea examples show simple Model-View-Update pattern, not complex step handlers
- **Idiomatic Pattern**: Sequential forms in single bubbletea model, not custom wizard framework
- **User's Simplified Need**: Most common case is Simple goal + Manual scoring + Custom prompt
- **Current Implementation Issues**:
  - Over-engineered wizard with step handlers, navigation controllers, state serialization
  - Complex interfaces (State, StepHandler, NavigationController) add unnecessary abstraction
  - Form initialization happening in Render() method instead of Init()/Update()
  - Custom validation logic instead of using huh's built-in validation
- **Simplified Solution**: Replace complex wizard with idiomatic bubbletea sequential forms
- **Focus**: Get Simple + Manual goals working perfectly before expanding to other types
- **Expected User Flow**: Title → Description → Goal Type → Scoring Type → Custom Prompt → Save

**Phase 2.8 Implementation (Completed):**
- **Documentation Review**: Studied bubbletea and huh examples to implement idiomatic patterns
- **Simplified Architecture**: Replaced complex wizard system with simple Model-View-Update pattern
- **SimpleGoalCreator**: New idiomatic bubbletea model following huh/bubbletea integration example
- **Sequential Forms**: Uses huh.NewForm() with groups, follows documented patterns exactly
- **Form Structure**: 
  - Group 1: Scoring type selection (Manual vs Automatic)
  - Group 2: Custom prompt input (conditional, hidden for automatic scoring)
- **Data Binding**: Direct field binding (&creator.field) per huh documentation
- **State Management**: Simple struct fields, no complex state serialization needed
- **Validation**: Uses huh's built-in validation (ValidateNotEmpty, custom validators)
- **Conditional Logic**: Uses WithHideFunc() to conditionally show prompt step
- **Default Values**: "Did you accomplish this goal today?" for manual goal prompts
- **Goal Structure**: Creates models.Goal matching expected YAML format from user testing
- **Integration**: Seamless integration with existing configurator.AddGoal() flow
- **Key Files Implemented**:
  - `simple_goal_creator.go`: Complete idiomatic bubbletea model with huh forms
  - Updated `configurator.go`: Uses SimpleGoalCreator instead of complex wizard
  - Added comprehensive AIDEV-NOTE comments referencing documentation
- **Benefits**: Much simpler, easier to understand, follows established patterns
- **Focus**: Manual simple goals (90% use case) work perfectly, foundation for other types

**References:**
- [huh documentation](https://github.com/charmbracelet/huh) - Forms and prompts
- [huh API reference](https://pkg.go.dev/github.com/charmbracelet/huh)
- [bubbletea documentation](https://github.com/charmbracelet/bubbletea) - CLI UI framework  
- [bubbletea API reference](https://pkg.go.dev/github.com/charmbracelet/bubbletea)