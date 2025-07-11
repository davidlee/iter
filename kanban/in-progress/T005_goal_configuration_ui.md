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

- [ ] **2.4 Informational Goal Wizard Flow**
  - [ ] Step 1: Basic info and field type selection
  - [ ] Step 2: Field configuration and direction
  - [ ] Step 3: Preview and confirmation

- [ ] **2.5 Hybrid Implementation Strategy**
  - [ ] Keep simple huh forms for basic interactions
  - [ ] Use bubbletea for complex multi-step flows
  - [ ] Create reusable bubbletea components that can embed huh forms
  - [ ] Maintain backwards compatibility with existing patterns

### Phase 3: Goal Management Enhancement
- [ ] **3.1 Enhanced Goal Listing (iter goal list)**
  - [ ] Rich table display with goal summaries
  - [ ] Interactive filtering and sorting
  - [ ] Goal status indicators (manual/automatic scoring, completeness)
  - [ ] Search functionality for large goal sets

- [ ] **3.2 Enhanced Goal Editing (iter goal edit)**
  - [ ] Interactive goal selection with preview
  - [ ] Wizard-style editing with current values pre-populated
  - [ ] Live preview of changes before saving
  - [ ] Better error recovery and validation
  - [ ] Preserve goal ID and data integrity

- [ ] **3.3 Enhanced Goal Removal (iter goal remove)**
  - [ ] Interactive goal selection with details
  - [ ] Impact analysis (entries that reference this goal)
  - [ ] Confirmation with goal summary
  - [ ] Safe removal with backup options

### Phase 4: Integration, Testing & Polish
- [ ] **4.1 File operations**
  - Atomic goal additions/modifications using existing parser
  - Error handling for file permissions, disk space
  - Backup existing goals.yml before modifications

- [ ] **4.2 Validation integration**
  - Leverage existing `Goal.Validate()` and `Schema.Validate()`
  - Real-time validation during form input where possible
  - Clear error messages for validation failures

- [ ] **4.3 Testing & documentation**
  - Unit tests for form builders and validation
  - Integration tests for complete workflows
  - Update CLI help text and documentation

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

**References:**
- [huh documentation](https://github.com/charmbracelet/huh) - Forms and prompts
- [huh API reference](https://pkg.go.dev/github.com/charmbracelet/huh)
- [bubbletea documentation](https://github.com/charmbracelet/bubbletea) - CLI UI framework  
- [bubbletea API reference](https://pkg.go.dev/github.com/charmbracelet/bubbletea)