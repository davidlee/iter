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

- [ ] **2.0 Flow Analysis and Enhancement Planning**
  - [ ] Analyze current multi-step goal creation flow (4-6 form interactions)
  - [ ] Document logical flow with text diagrams for each goal type:
    - [ ] Simple goal flow diagram (4 steps: basic info → scoring → criteria → confirmation)
    - [ ] Elastic goal flow diagram (6-8 steps: basic info → field config → scoring → mini/midi/maxi criteria → validation → confirmation)
    - [ ] Informational goal flow diagram (3 steps: basic info → field config → confirmation)
    - [ ] Decision tree diagrams for conditional flows (manual vs automatic scoring, field type branches)
  - [ ] Evaluate bubbletea integration opportunities vs standalone huh forms:
    - [ ] Complexity analysis: when bubbletea adds value vs overhead
    - [ ] User experience improvements: navigation, progress, error recovery
    - [ ] Technical integration patterns: embedding huh in bubbletea vs standalone
  - [ ] Design enhanced UX patterns:
    - [ ] Progress indicator designs (Step X of Y, progress bar, breadcrumbs)
    - [ ] Navigation patterns (back/forward buttons, step jumping, cancel/exit)
    - [ ] Real-time validation display (inline errors, live help text, field highlighting)
    - [ ] Goal preview formats (summary cards, YAML preview, validation status)
  - [ ] Plan API interfaces for bubbletea-enhanced components:
    - [ ] Wizard state management interfaces (WizardState, StepHandler, NavigationController)
    - [ ] Form embedding patterns (HuhFormStep, FormRenderer, ValidationCollector)  
    - [ ] Progress tracking APIs (ProgressTracker, StepValidator, StateSerializer)
    - [ ] Error recovery mechanisms (StateSnapshot, ErrorHandler, RetryStrategy)
  - [ ] Create detailed implementation strategy focusing on elastic goals:
    - [ ] Complex criteria validation flow (mini ≤ midi ≤ maxi constraints)
    - [ ] Dynamic field configuration based on field type selection
    - [ ] Progressive disclosure patterns for complex options
    - [ ] State persistence between steps for long flows

- [ ] **2.1 Bubbletea Goal Creation Wizard (Enhanced UX)**
  - [ ] Convert multi-step goal creation to unified bubbletea application
  - [ ] Implement progress indicators showing current step (Step X of Y)
  - [ ] Add back/forward navigation between steps
  - [ ] Real-time validation with contextual error display
  - [ ] Goal preview and confirmation step before saving
  - [ ] Enhanced error recovery without losing progress

- [ ] **2.2 Simple Goal Wizard Flow**
  - [ ] Step 1: Basic info (title, description, goal type pre-selected)
  - [ ] Step 2: Scoring configuration (manual/automatic)
  - [ ] Step 3: Criteria definition (if automatic scoring)
  - [ ] Step 4: Preview and confirmation

- [ ] **2.3 Elastic Goal Wizard Flow (Complex)**
  - [ ] Step 1: Basic info and field type selection
  - [ ] Step 2: Field configuration (units, constraints)
  - [ ] Step 3: Scoring type selection
  - [ ] Step 4: Mini-level criteria definition
  - [ ] Step 5: Midi-level criteria definition  
  - [ ] Step 6: Maxi-level criteria definition
  - [ ] Step 7: Criteria validation and preview
  - [ ] Step 8: Final confirmation with complete goal summary

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
- **Simple Goals**: 4 steps → Good candidate for bubbletea wizard
- **Elastic Goals**: 6-8 steps → High value from enhanced navigation and progress
- **Informational Goals**: 3 steps → Moderate benefit from bubbletea
- **Goal Management**: List/edit/remove → Enhanced interaction patterns valuable

**References:**
- [huh documentation](https://github.com/charmbracelet/huh) - Forms and prompts
- [huh API reference](https://pkg.go.dev/github.com/charmbracelet/huh)
- [bubbletea documentation](https://github.com/charmbracelet/bubbletea) - CLI UI framework  
- [bubbletea API reference](https://pkg.go.dev/github.com/charmbracelet/bubbletea)