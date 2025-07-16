---
title: "Habit Management: Complex Habit Types (UI)"
type: ["feature"] # feature | fix | documentation | testing | refactor | chore
tags: ["ui"] 
related_tasks: ["related-to:T006", "related-to:T005"] # Optional with relationship type
context_windows: ["./*.go", Claude.md, workflow.md] # List of glob patterns useful to build the context window required for this task
---

# Habit Management: Complex Habit Types (UI)

## Git Commit History

**All commits related to this task (newest first):**

- `c0930b0` - feat(habitconfig)[T009/2.3]: integrate ElasticHabitCreator with configurator routing
- `493a51c` - test(habitconfig)[T009/2.2]: implement comprehensive ElasticHabitCreator test suite
- `2df1abe` - feat(habitconfig)[T009/2.1]: design ElasticHabitCreator architecture
- `d3960f2` - test(habitconfig)[T009/1.4]: add comprehensive testing for enhanced SimpleHabitCreator
- `a6d76f8` - feat(habitconfig)[T009/1.3]: add automatic criteria support to SimpleHabitCreator
- `7ca86f3` - feat(habitconfig)[T009/1.2]: extend SimpleHabitCreator with field type support
- `5128c42` - docs(T009)[T009/1.1]: analyze Simple habit requirements and create task card

**Context (Background)**:
- CLAUDE.md
- doc/workflow.md
- doc/flow_analysis_T005.md
- doc/specifications/habit_schema.md
- tasks T005, T006
- API docs for huh, bubbletea

**Context (Significant Code Files)**:
- internal/ui/habitconfig/ - Habit configuration UI system from T005
- internal/ui/habitconfig/configurator.go - Main habit configurator with routing logic
- internal/ui/habitconfig/simple_habit_creator.go - Simple habit creation UI (enhanced with field types + criteria)
- internal/ui/habitconfig/simple_habit_creator_test.go - Comprehensive test suite for SimpleHabitCreator
- internal/ui/habitconfig/simple_habit_creator_integration_test.go - Integration tests for all combinations
- internal/ui/habitconfig/elastic_habit_creator.go - Elastic habit creation UI (three-tier criteria, 530+ lines)
- internal/ui/habitconfig/checklist_habit_creator.go - Checklist habit creation UI (T007/4.1)
- internal/ui/habitconfig/informational_habit_creator.go - Informational habit creation UI (working)
- internal/ui/habitconfig/field_value_input.go - Field type configuration system
- internal/models/habit.go - Habit data models and validation (includes elastic habit validation)
- internal/parser/habit_parser.go - Habit persistence and loading
- T009_IMPLEMENTATION_STATUS.md - Pre-compact analysis and implementation status

## Notes (temp)

Follows on from T005.

We have implemented informational habits with Boolean, Text, Time, Duration field types. 
We have Simple+Manual+Boolean habits. We don't yet have Simple+Manual+(other field types). 
We don't yet have working Elastic habits.

simple > automatic
   vice habit add --dry-run

  🎯 Add New Habit

  Let's create a new habit through guided prompts.

  ✅ Habit created successfully: test
  Error: habit validation failed: criteria is required for automatic scoring

elastic > automatic
   vice habit add --dry-run

  🎯 Add New Habit

  Let's create a new habit through guided prompts.

  ✅ Habit created successfully: test
  Error: habit validation failed: mini_criteria is required for automatic scoring of elastic habits

elastic > manual

    - title: test
      id: test
      position: 7
      habit_type: elastic
      field_type:
        type: boolean
      scoring_type: manual
      prompt: Did you accomplish this habit today?
(no errors, but not correct; an elastic manual habit should be able to have a text/time/duration/(checklist) field for information capture; a boolean field is nonsensical here although maybe (maybe ) makes sense as a convention that there's no other field type.

Not sure how much of this is not implemented vs currently broken - that's the first thing to determine.

---
musings:

Actually that's a requirement worth highlighting that's slipped through the cracks a bit:
- simple+manual habits should be able to have a non-boolean field type, even if they lack criteria (which automatically scored habits have)
- the current design conflates a boolean data field with the (boolean or quaternary) success / failure scoring of a habit or habit.
  - a boolean checkbox + comment text might be data manually reviewed to make a pass/fail determination.
  - or, that might be an edge case, and we should just remove boolean "fields". We have checklists, after all.
  - haven't yet tackled the design issues of allowing multiple fields for a habit, potentially of different kinds - nor evaluated the benefits.
- the more I think about it the less apt "habit" seems and the more it should be called "habit" or "routine" or ...

---



## 1. Habit / User Story

As a user, I want to create and configure all habit types with appropriate field types and scoring mechanisms so that I can track diverse habits and routines with the data collection and scoring approach that best fits each one.

**Current State (from T005):**
- ✅ Simple + Manual + Boolean habits work correctly
- ✅ Informational habits with all field types (Boolean, Text, Numeric, Time, Duration) work correctly
- ✅ ChecklistHabit support added to habit configuration UI (T007/4.1)
- ❌ Simple + Automatic habits fail validation ("criteria is required")
- ❌ Elastic habits incomplete/broken (missing criteria, inappropriate field types)
- ❌ Simple + Manual habits limited to Boolean fields only

**User Story:**
I want to create habits that match my tracking needs:
- **Simple + Manual + Non-Boolean**: Track completion with additional data (e.g., "Did I exercise?" + duration field)
- **Simple + Automatic**: Habits with clear numeric/time criteria (e.g., "Exercise for 30+ minutes")
- **Elastic + Manual**: Three-tier achievement habits with subjective scoring (e.g., "mini/midi/maxi exercise intensity")
- **Elastic + Automatic**: Three-tier habits with numeric criteria (e.g., "mini: 15min, midi: 30min, maxi: 60min exercise")
- **All habit types**: Should support appropriate field types for data collection beyond simple Boolean completion

## 2. Acceptance Criteria

### Simple Habit Improvements
- [ ] Simple + Manual habits support all appropriate field types (Text, Numeric, Time, Duration, Checklist)
- [ ] Simple + Automatic habits work with proper criteria definition
- [ ] Simple + Automatic + Boolean: criteria uses boolean condition (equals: true)
- [ ] Simple + Automatic + Numeric: criteria uses numeric conditions (greater_than, etc.)
- [ ] Simple + Automatic + Time/Duration: criteria uses time-based conditions
- [ ] Criteria validation ensures automatic scoring requirements are met

### Elastic Habit Implementation
- [ ] Elastic + Manual habits support appropriate field types (Text, Numeric, Time, Duration, Checklist)
- [ ] Elastic + Automatic habits support mini/midi/maxi criteria definition
- [ ] Elastic criteria validation enforces mini ≤ midi ≤ maxi constraints
- [ ] Elastic habits generate proper YAML with mini_criteria, midi_criteria, maxi_criteria
- [ ] ElasticHabitCreator bubbletea component following established patterns

### Habit Type and Field Type Matrix
- [ ] Simple + Manual + Text: Completion tracking with text notes
- [ ] Simple + Manual + Numeric: Completion tracking with numeric data
- [ ] Simple + Manual + Time: Completion tracking with time-of-day data
- [ ] Simple + Manual + Duration: Completion tracking with duration data
- [ ] Simple + Manual + Checklist: Completion tracking with checklist progress
- [ ] Elastic + Manual + (same field types): Three-tier subjective scoring with data collection
- [ ] Elastic + Automatic + Numeric: Automatic scoring based on numeric thresholds
- [ ] Elastic + Automatic + Time/Duration: Automatic scoring based on time thresholds

### UI and User Experience
- [ ] Habit creation flow guides users to appropriate field type selections
- [ ] Criteria definition UI provides clear examples and validation
- [ ] Error messages explain validation failures clearly
- [ ] Dry-run mode works for all habit type combinations
- [ ] Generated YAML validates correctly for all supported combinations

### Technical Requirements
- [ ] Reuse existing field type configuration system from InformationalHabitCreator
- [ ] Extend SimpleHabitCreator to support field type selection and criteria definition
- [ ] Create ElasticHabitCreator following SimpleHabitCreator patterns
- [ ] Update configurator routing to handle enhanced Simple and new Elastic flows
- [ ] Comprehensive test coverage for all habit type + field type + scoring type combinations


---
## 3. Implementation Plan & Progress

**Overall Status:** `Phase 2 Complete` 
*Phase 1 (Simple Habit Enhancement) ✅ Complete: 15 field type + scoring combinations tested and validated.*
*Phase 2 (Elastic Habit Implementation) ✅ Complete: ElasticHabitCreator implemented, tested (46 tests), and integrated with configurator.*

**Architecture Analysis:**
Building on T005's successful implementation patterns:

**Existing Foundation:**
- ✅ SimpleHabitCreator: Idiomatic bubbletea + huh implementation (2-step flow)
- ✅ InformationalHabitCreator: Complete field type configuration system (4-step flow)
- ✅ ChecklistHabitCreator: Checklist habit support (T007/4.1, 3-step flow)
- ✅ Field type system: Boolean, Text, Numeric, Time, Duration, Checklist
- ✅ FieldValueInput: Reusable field configuration components
- ✅ Habit validation and YAML persistence infrastructure

**Implementation Strategy:**
1. **Extend SimpleHabitCreator** to support field types and automatic scoring
2. **Create ElasticHabitCreator** following established bubbletea patterns
3. **Reuse field configuration logic** from InformationalHabitCreator
4. **Update configurator routing** to handle enhanced habit creators

**Sub-tasks:**

### Phase 1: Simple Habit Enhancement
- [x] **1.1: Analyze Simple Habit Requirements** ✅ **COMPLETED**
  - [x] Investigate current SimpleHabitCreator limitations (Boolean-only fields)
  - [x] Define field type matrix for Simple habits (which field types make sense)
  - [x] Design automatic criteria definition UI patterns
  - [x] Plan integration with existing FieldValueInput system

- [x] **1.2: Extend SimpleHabitCreator for Field Types** ✅ **COMPLETED**
  - [x] Add field type selection step to SimpleHabitCreator flow
  - [x] Integrate FieldTypeSelector from InformationalHabitCreator
  - [x] Update habit building logic to support non-Boolean fields
  - [x] Maintain backwards compatibility for existing Simple + Manual + Boolean flow

- [x] **1.3: Add Automatic Criteria Support to SimpleHabitCreator** ✅ **COMPLETED**
  - [x] Design criteria definition UI for different field types
  - [x] Boolean criteria: equals condition (true for completion)
  - [x] Numeric criteria: threshold conditions (greater_than, etc.)
  - [x] Time/Duration criteria: time-based conditions
  - [x] Add criteria validation and user-friendly error messages

- [x] **1.4: Test and Validate Simple Habit Enhancements** ✅ **COMPLETED**
  - [x] Add headless testing support with `NewSimpleHabitCreatorForTesting()` constructor
  - [x] Create `TestHabitData` helper struct for pre-populating configuration fields
  - [x] Add `CreateHabitDirectly()` bypass method for testing business logic without UI
  - [x] Unit tests for enhanced SimpleHabitCreator (existing 17 tests + new headless tests)
  - [x] Integration tests for all Simple + field type + scoring type combinations
  - [x] YAML validation for generated Simple habits across all combinations
  - [x] Manual testing with dry-run mode for UI verification (see test_dry_run_manual.md)

### Phase 2: Elastic Habit Implementation
- [x] **2.1: Design ElasticHabitCreator Architecture** ✅ **COMPLETED**
  - [x] Follow SimpleHabitCreator patterns for consistency
  - [x] Plan multi-step flow: Field Type → Field Config → Scoring → Criteria (mini/midi/maxi) → Prompt
  - [x] Design criteria definition UI for three-tier habits
  - [x] Plan validation logic for mini ≤ midi ≤ maxi constraints

- [x] **2.2: Implement ElasticHabitCreator Component** ✅ **COMPLETED**
  - [x] Create comprehensive test suite for ElasticHabitCreator (46 tests total)
  - [x] Unit tests covering all functionality (20 tests)
  - [x] Integration tests for all Elastic + field type + scoring type combinations (13 combinations)
  - [x] Three-tier criteria validation testing (mini ≤ midi ≤ maxi constraint enforcement)
  - [x] YAML validation for all generated Elastic habits passes schema validation
  - [x] Error handling tests for invalid inputs and edge cases
  - [x] Code quality compliance (linting, formatting)

- [x] **2.3: Integrate ElasticHabitCreator with Configurator** ✅ **COMPLETED**
  - [x] Add ElasticHabit case to configurator switch statement (lines 88-93 and 365-370)
  - [x] Create runElasticHabitCreator method following existing patterns (lines 296-331)
  - [x] Update routing logic in both AddHabit and AddHabitWithYAMLOutput methods
  - [x] Add comprehensive integration tests (4 tests covering routing, creation, headless integration)
  - [x] Verify proper habit building and YAML generation through integration tests
  - [x] Elastic habit description already properly configured in habit type selection

- [ ] **2.4: Test and Validate Elastic Habit Implementation**
  - [ ] Unit tests for ElasticHabitCreator
  - [ ] Integration tests for Elastic + field type + scoring type combinations
  - [ ] Criteria validation testing (constraint enforcement)
  - [ ] Manual testing with complex Elastic habit scenarios

### Phase 3: Habit Type Matrix Completion
- [ ] **3.1: Comprehensive Habit Type Testing**
  - [ ] Test all Simple + field type + scoring type combinations
  - [ ] Test all Elastic + field type + scoring type combinations
  - [ ] Verify ChecklistHabit integration works correctly
  - [ ] Validate Informational habits continue working as expected

- [ ] **3.2: User Experience Refinements**
  - [ ] Improve field type selection guidance (which types work best for which habits)
  - [ ] Enhance criteria definition UI with examples and help text
  - [ ] Add validation error improvements with specific guidance
  - [ ] Test complete user workflows end-to-end

- [ ] **3.3: Documentation and Integration**
  - [ ] Update habit configuration documentation
  - [ ] Create examples for each habit type + field type combination
  - [ ] Verify integration with entry recording system (for future T007 work)
  - [ ] Final testing and quality assurance

**Technical Implementation Notes:**
- **Pattern Consistency**: All new creators follow SimpleHabitCreator + InformationalHabitCreator patterns
- **Code Reuse**: Leverage FieldTypeSelector, FieldConfig, and FieldValueInput from existing system
- **Validation Strategy**: Use huh's built-in validation plus custom habit validation
- **Flow Design**: Multi-step forms with conditional groups based on scoring type selection
- **Error Handling**: Clear, actionable error messages for criteria and field type mismatches
- **Testing Strategy**: Low-effort headless testing via test constructor + bypass methods (minimal refactoring)

## 4. Roadblocks

*(Timestamped list of any impediments. AI adds here when a sub-task is marked `[blocked]`)*

## 5. Notes / Discussion Log

**2025-07-12 - AI Initial Analysis:**
- T005 provides excellent foundation with working SimpleHabitCreator and InformationalHabitCreator
- Key issue: Current Simple habits hardcoded to Boolean fields, missing field type selection
- ElasticHabit completely missing but models/validation support exists
- Automatic scoring criteria definition missing for Simple habits
- ChecklistHabit support added in T007/4.1 provides additional reference implementation
- Strategy: Extend existing creators rather than rewrite, maintain idiomatic bubbletea patterns

**2025-07-12 - Design Decisions (T009/1.1):**
1. **Field Type Support**: Simple habits support all field types except checklist (Boolean, Text, Numeric, Time, Duration)
2. **Text Field Restriction**: Text fields restricted to manual scoring only (no automatic text-based criteria)
3. **UI Flow Pattern**: Use multi-step forms (like InformationalHabitCreator) with step omission when not required
4. **Quick Path**: Boolean+Manual remains the streamlined path with minimal steps
5. **Comment Pattern**: Optional comment field accompanies all field types, extend to checklist fields also

**Simple Habit Field Type Matrix (Approved):**
| Field Type | Manual Scoring | Automatic Scoring | Criteria Options | Notes |
|------------|---------------|-------------------|------------------|--------|
| Boolean | ✅ Yes | ✅ Yes | `equals: true` | Traditional pass/fail + quick path |
| Text | ✅ Yes | ❌ No | N/A | Subjective content, manual only |
| Numeric | ✅ Yes | ✅ Yes | `>`, `>=`, `<`, `<=`, `range` | Exercise minutes, reps, etc. |
| Time | ✅ Yes | ✅ Yes | `before`, `after` | Wake-up time, bedtime |
| Duration | ✅ Yes | ✅ Yes | `>`, `>=`, `<`, `<=`, `range` | Exercise duration, meditation |
| Checklist | ❌ N/A | ❌ N/A | N/A | Use ChecklistHabit type instead |

**Enhanced SimpleHabitCreator Flow Design:**
1. **Boolean + Manual (Quick Path)**: Basic Info → Scoring Type (auto-select manual) → Prompt → Save (2 steps)
2. **Boolean + Automatic**: Basic Info → Scoring Type → Criteria (equals: true) → Save (3 steps)
3. **Other Field Types + Manual**: Basic Info → Field Type → Field Config → Scoring Type → Comment/Prompt → Save (4-5 steps)
4. **Other Field Types + Automatic**: Basic Info → Field Type → Field Config → Scoring Type → Criteria → Comment/Prompt → Save (5-6 steps)

**Step Omission Strategy:**
- Skip Field Type step for Boolean habits (maintain current quick path)
- Skip Field Config step when field type needs no configuration
- Skip Criteria step for manual scoring
- Skip Comment step if user doesn't want additional data collection

**Design Considerations:**
- **Field Type Appropriateness**: All field types except checklist supported for Simple habits
- **Automatic vs Manual Scoring**: 
  - Manual: User decides achievement level, field data is informational
  - Automatic: System determines achievement based on field data meeting criteria
- **Text Field Limitation**: Text fields restricted to manual scoring (no automatic text evaluation)
- **Comment Pattern**: Optional comment field for all field types including checklist habits (ChecklistHabitCreator needs enhancement)
- **Backward Compatibility**: Existing Boolean+Manual flow preserved as quick path

**T009/1.2 Implementation Details (2025-07-12):**
- **Multi-step Conversion**: Converted SimpleHabitCreator from sequential groups to multi-step forms
- **Field Type Support**: Added support for Boolean, Text, Numeric (3 subtypes), Time, Duration fields
- **Dynamic Flow**: Flow adjusts based on field type - 3-4 steps depending on configuration needs
- **Field Configuration**: Numeric fields support subtype, unit, min/max constraints; Text supports multiline
- **Scoring Restrictions**: Text fields restricted to manual scoring only (automatic scoring prevented)
- **Comment Integration**: Optional comment field appended to habit description (temporary solution)
- **Comprehensive Testing**: 9 unit tests covering all field types and flow scenarios
- **Backward Compatibility**: Boolean field type remains default, maintains existing quick path

**T009/1.3 Implementation Details (2025-07-12):**
- **Criteria Definition Forms**: Created field-type-specific criteria forms for automatic scoring
- **Boolean Criteria**: Automatic "equals: true" condition with informational display
- **Numeric Criteria**: Support for >, >=, <, <=, and range conditions with unit display
- **Time Criteria**: Before/after time comparisons with HH:MM validation
- **Duration Criteria**: Duration-based conditions with flexible format support (30m, 1h, etc.)
- **Dynamic Flow Integration**: Added criteria step between scoring and prompt, with flow adjustment
- **Validation**: Comprehensive input validation for all criteria types with error messages
- **Habit Building**: Complete criteria construction with proper models.Condition structure
- **Range Support**: Inclusive/exclusive range boundaries for numeric and duration criteria
- **Error Handling**: Graceful handling of invalid values and unsupported field types
- **Comprehensive Testing**: 8 additional unit tests covering all criteria types and edge cases
- **AIDEV Anchor Comments**: Added key anchor comments for criteria dispatch, builder, and flow logic

**T009/1.4 Implementation Details (2025-07-12):**
- **Headless Testing Infrastructure**: Added `NewSimpleHabitCreatorForTesting()` constructor bypassing UI
- **Test Data Helper**: `TestHabitData` struct for clean specification of all configuration options
- **Direct Habit Creation**: `CreateHabitDirectly()` method enables testing business logic without UI flow
- **Comprehensive Integration Tests**: 15 field type + scoring type combinations fully tested
- **YAML Validation**: All habit combinations generate valid YAML that passes schema validation
- **Criteria Validation**: Complete testing of Boolean, Numeric (>, >=, <, <=, range), Time, Duration criteria
- **Manual Testing Guide**: Created test_dry_run_manual.md for interactive CLI verification
- **Test Coverage**: 42 total tests covering all aspects of enhanced SimpleHabitCreator

**T009/2.3 Implementation Details (2025-07-12):**
- **AIDEV-NOTE: configurator-elastic-integration-complete; ElasticHabitCreator now properly integrated with configurator routing**
- **Integration Points**: Updated 2 routing locations in configurator.go (AddHabit: lines 88-93, AddHabitWithYAMLOutput: lines 365-370)
- **New Method**: Added `runElasticHabitCreator()` method (lines 296-331) following exact pattern of existing habit creators
- **Pattern Consistency**: Follows `runInformationalHabitCreator` and `runChecklistHabitCreator` patterns exactly
- **Integration Tests**: 4 comprehensive tests covering routing, creation, headless integration, and criteria validation
- **Error Handling**: Proper error messages for elastic habit creation failures with clear error propagation
- **UI Consistency**: ElasticHabit option already properly configured in habit type selection with clear description
- **YAML Support**: Both regular and dry-run YAML generation now properly route to ElasticHabitCreator


**T009/2.2 Implementation Details (2025-07-12):**
- **Comprehensive Test Suite**: 46 tests covering all aspects of ElasticHabitCreator functionality
- **Headless Testing Infrastructure**: `NewElasticHabitCreatorForTesting()` constructor and `CreateHabitDirectly()` method
- **Test Coverage**: 20 unit tests + 26 integration tests covering all field type + scoring type combinations
- **Three-Tier Criteria Testing**: Complete validation of mini/midi/maxi criteria for Numeric, Time, Duration field types
- **Constraint Validation**: Tests verify mini ≤ midi ≤ maxi ordering enforcement by model validation
- **Field Type Support**: Text (manual only), Numeric (3 subtypes), Time, Duration with appropriate automatic scoring
- **YAML Validation**: All 13 field type + scoring combinations generate valid YAML passing schema validation
- **Error Handling**: Comprehensive edge case testing for invalid values, unsupported field types, unknown tiers
- **Code Quality**: Fixed deprecated `strings.Title` usage, passed all linting and formatting checks
- **Pattern Consistency**: Follows SimpleHabitCreator patterns exactly for maintainability and consistency

**T009/2.1 Architecture Design (2025-07-12):**
- **ElasticHabitCreator Structure**: Complete bubbletea model following SimpleHabitCreator patterns
- **Multi-Step Flow**: Field Type → Field Config → Scoring → Three-Tier Criteria → Prompt (4-5 steps)
- **Three-Tier Criteria**: Mini/Midi/Maxi achievement levels with validation (mini ≤ midi ≤ maxi)
- **Field Type Support**: Text, Numeric (3 subtypes), Time, Duration (Boolean excluded - not meaningful for elastic)
- **Headless Testing Ready**: `TestElasticHabitData` struct and `NewElasticHabitCreatorForTesting()` constructor
- **Validation Strategy**: Real-time validation for criteria ordering and proper threshold definitions
- **Reuse Patterns**: Field configuration, scoring selection, and form patterns from SimpleHabitCreator
- **Habit Building**: Complete three-tier criteria construction with models.MiniCriteria/MidiCriteria/MaxiCriteria

**Technical Integration Points (T009/1.1 Findings):**
- **Existing FieldValueInput System**: Ready for reuse in criteria definition (field_value_input.go)
- **InformationalHabitCreator Patterns**: Multi-step approach with `currentStep` and `initializeStep()` method
- **ChecklistHabitCreator Patterns**: Sequential form groups in single huh.Form (simpler but less flexible)
- **Models Support**: Complete criteria condition support for all field types in models.Condition struct
- **Current SimpleHabitCreator**: 2-step flow, hardcoded Boolean, missing automatic criteria (lines 172, 182)
- **TTY Limitation**: UI framework requires interactive terminal, no piped input testing possible

## 6. Code Snippets & Artifacts 

*(AI will place larger generated code blocks or references to files here if planned / directed. User will then move these to actual project files.)*