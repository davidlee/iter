---
title: "Goal Management: Complex Goal Types (UI)"
type: ["feature"] # feature | fix | documentation | testing | refactor | chore
tags: ["ui"] 
related_tasks: ["related-to:T006", "related-to:T005"] # Optional with relationship type
context_windows: ["./*.go", Claude.md, workflow.md] # List of glob patterns useful to build the context window required for this task
---

# Goal Management: Complex Goal Types (UI)

**Context (Background)**:
- CLAUDE.md
- doc/workflow.md
- doc/flow_analysis_T005.md
- doc/specifications/goal_schema.md
- tasks T005, T006
- API docs for huh, bubbletea

**Context (Significant Code Files)**:
- internal/ui/goalconfig/ - Goal configuration UI system from T005
- internal/ui/goalconfig/configurator.go - Main goal configurator with routing logic
- internal/ui/goalconfig/simple_goal_creator.go - Simple goal creation UI (working)
- internal/ui/goalconfig/informational_goal_creator.go - Informational goal creation UI (working)
- internal/ui/goalconfig/field_value_input.go - Field type configuration system
- internal/models/goal.go - Goal data models and validation
- internal/parser/goal_parser.go - Goal persistence and loading

## Notes (temp)

Follows on from T005.

We have implemented informational goals with Boolean, Text, Time, Duration field types. 
We have Simple+Manual+Boolean goals. We don't yet have Simple+Manual+(other field types). 
We don't yet have working Elastic goals.

simple > automatic
   iter goal add --dry-run

  🎯 Add New Goal

  Let's create a new goal through guided prompts.

  ✅ Goal created successfully: test
  Error: goal validation failed: criteria is required for automatic scoring

elastic > automatic
   iter goal add --dry-run

  🎯 Add New Goal

  Let's create a new goal through guided prompts.

  ✅ Goal created successfully: test
  Error: goal validation failed: mini_criteria is required for automatic scoring of elastic goals

elastic > manual

    - title: test
      id: test
      position: 7
      goal_type: elastic
      field_type:
        type: boolean
      scoring_type: manual
      prompt: Did you accomplish this goal today?
(no errors, but not correct; an elastic manual goal should be able to have a text/time/duration/(checklist) field for information capture; a boolean field is nonsensical here although maybe (maybe ) makes sense as a convention that there's no other field type.

Not sure how much of this is not implemented vs currently broken - that's the first thing to determine.

---
musings:

Actually that's a requirement worth highlighting that's slipped through the cracks a bit:
- simple+manual goals should be able to have a non-boolean field type, even if they lack criteria (which automatically scored goals have)
- the current design conflates a boolean data field with the (boolean or quaternary) success / failure scoring of a goal or habit.
  - a boolean checkbox + comment text might be data manually reviewed to make a pass/fail determination.
  - or, that might be an edge case, and we should just remove boolean "fields". We have checklists, after all.
  - haven't yet tackled the design issues of allowing multiple fields for a goal, potentially of different kinds - nor evaluated the benefits.
- the more I think about it the less apt "goal" seems and the more it should be called "habit" or "routine" or ...

---



## 1. Goal / User Story

As a user, I want to create and configure all goal types with appropriate field types and scoring mechanisms so that I can track diverse habits and routines with the data collection and scoring approach that best fits each one.

**Current State (from T005):**
- ✅ Simple + Manual + Boolean goals work correctly
- ✅ Informational goals with all field types (Boolean, Text, Numeric, Time, Duration) work correctly
- ✅ ChecklistGoal support added to goal configuration UI (T007/4.1)
- ❌ Simple + Automatic goals fail validation ("criteria is required")
- ❌ Elastic goals incomplete/broken (missing criteria, inappropriate field types)
- ❌ Simple + Manual goals limited to Boolean fields only

**User Story:**
I want to create goals that match my tracking needs:
- **Simple + Manual + Non-Boolean**: Track completion with additional data (e.g., "Did I exercise?" + duration field)
- **Simple + Automatic**: Goals with clear numeric/time criteria (e.g., "Exercise for 30+ minutes")
- **Elastic + Manual**: Three-tier achievement goals with subjective scoring (e.g., "mini/midi/maxi exercise intensity")
- **Elastic + Automatic**: Three-tier goals with numeric criteria (e.g., "mini: 15min, midi: 30min, maxi: 60min exercise")
- **All goal types**: Should support appropriate field types for data collection beyond simple Boolean completion

## 2. Acceptance Criteria

### Simple Goal Improvements
- [ ] Simple + Manual goals support all appropriate field types (Text, Numeric, Time, Duration, Checklist)
- [ ] Simple + Automatic goals work with proper criteria definition
- [ ] Simple + Automatic + Boolean: criteria uses boolean condition (equals: true)
- [ ] Simple + Automatic + Numeric: criteria uses numeric conditions (greater_than, etc.)
- [ ] Simple + Automatic + Time/Duration: criteria uses time-based conditions
- [ ] Criteria validation ensures automatic scoring requirements are met

### Elastic Goal Implementation
- [ ] Elastic + Manual goals support appropriate field types (Text, Numeric, Time, Duration, Checklist)
- [ ] Elastic + Automatic goals support mini/midi/maxi criteria definition
- [ ] Elastic criteria validation enforces mini ≤ midi ≤ maxi constraints
- [ ] Elastic goals generate proper YAML with mini_criteria, midi_criteria, maxi_criteria
- [ ] ElasticGoalCreator bubbletea component following established patterns

### Goal Type and Field Type Matrix
- [ ] Simple + Manual + Text: Completion tracking with text notes
- [ ] Simple + Manual + Numeric: Completion tracking with numeric data
- [ ] Simple + Manual + Time: Completion tracking with time-of-day data
- [ ] Simple + Manual + Duration: Completion tracking with duration data
- [ ] Simple + Manual + Checklist: Completion tracking with checklist progress
- [ ] Elastic + Manual + (same field types): Three-tier subjective scoring with data collection
- [ ] Elastic + Automatic + Numeric: Automatic scoring based on numeric thresholds
- [ ] Elastic + Automatic + Time/Duration: Automatic scoring based on time thresholds

### UI and User Experience
- [ ] Goal creation flow guides users to appropriate field type selections
- [ ] Criteria definition UI provides clear examples and validation
- [ ] Error messages explain validation failures clearly
- [ ] Dry-run mode works for all goal type combinations
- [ ] Generated YAML validates correctly for all supported combinations

### Technical Requirements
- [ ] Reuse existing field type configuration system from InformationalGoalCreator
- [ ] Extend SimpleGoalCreator to support field type selection and criteria definition
- [ ] Create ElasticGoalCreator following SimpleGoalCreator patterns
- [ ] Update configurator routing to handle enhanced Simple and new Elastic flows
- [ ] Comprehensive test coverage for all goal type + field type + scoring type combinations


---
## 3. Implementation Plan & Progress

**Overall Status:** `Not Started`
*AI updates this based on sub-task progress or user instruction*

**Architecture Analysis:**
Building on T005's successful implementation patterns:

**Existing Foundation:**
- ✅ SimpleGoalCreator: Idiomatic bubbletea + huh implementation (2-step flow)
- ✅ InformationalGoalCreator: Complete field type configuration system (4-step flow)
- ✅ ChecklistGoalCreator: Checklist goal support (T007/4.1, 3-step flow)
- ✅ Field type system: Boolean, Text, Numeric, Time, Duration, Checklist
- ✅ FieldValueInput: Reusable field configuration components
- ✅ Goal validation and YAML persistence infrastructure

**Implementation Strategy:**
1. **Extend SimpleGoalCreator** to support field types and automatic scoring
2. **Create ElasticGoalCreator** following established bubbletea patterns
3. **Reuse field configuration logic** from InformationalGoalCreator
4. **Update configurator routing** to handle enhanced goal creators

**Sub-tasks:**

### Phase 1: Simple Goal Enhancement
- [x] **1.1: Analyze Simple Goal Requirements** ✅ **COMPLETED**
  - [x] Investigate current SimpleGoalCreator limitations (Boolean-only fields)
  - [x] Define field type matrix for Simple goals (which field types make sense)
  - [x] Design automatic criteria definition UI patterns
  - [x] Plan integration with existing FieldValueInput system

- [x] **1.2: Extend SimpleGoalCreator for Field Types** ✅ **COMPLETED**
  - [x] Add field type selection step to SimpleGoalCreator flow
  - [x] Integrate FieldTypeSelector from InformationalGoalCreator
  - [x] Update goal building logic to support non-Boolean fields
  - [x] Maintain backwards compatibility for existing Simple + Manual + Boolean flow

- [x] **1.3: Add Automatic Criteria Support to SimpleGoalCreator** ✅ **COMPLETED**
  - [x] Design criteria definition UI for different field types
  - [x] Boolean criteria: equals condition (true for completion)
  - [x] Numeric criteria: threshold conditions (greater_than, etc.)
  - [x] Time/Duration criteria: time-based conditions
  - [x] Add criteria validation and user-friendly error messages

- [ ] **1.4: Test and Validate Simple Goal Enhancements**
  - [ ] Unit tests for enhanced SimpleGoalCreator
  - [ ] Integration tests for all Simple + field type + scoring type combinations
  - [ ] Manual testing with dry-run mode
  - [ ] YAML validation for generated Simple goals

### Phase 2: Elastic Goal Implementation
- [ ] **2.1: Design ElasticGoalCreator Architecture**
  - [ ] Follow SimpleGoalCreator patterns for consistency
  - [ ] Plan multi-step flow: Basic Info → Field Type → Scoring → Criteria (mini/midi/maxi) → Confirmation
  - [ ] Design criteria definition UI for three-tier goals
  - [ ] Plan validation logic for mini ≤ midi ≤ maxi constraints

- [ ] **2.2: Implement ElasticGoalCreator Component**
  - [ ] Create ElasticGoalCreator bubbletea model
  - [ ] Implement sequential huh forms with conditional groups
  - [ ] Add field type configuration step (reuse from InformationalGoalCreator)
  - [ ] Implement three-tier criteria definition for automatic scoring
  - [ ] Add cross-criteria validation (mini ≤ midi ≤ maxi)

- [ ] **2.3: Integrate ElasticGoalCreator with Configurator**
  - [ ] Add ElasticGoal case to configurator switch statement
  - [ ] Create runElasticGoalCreator method following existing patterns
  - [ ] Ensure proper goal building and YAML generation
  - [ ] Update goal type selection to include clear Elastic goal description

- [ ] **2.4: Test and Validate Elastic Goal Implementation**
  - [ ] Unit tests for ElasticGoalCreator
  - [ ] Integration tests for Elastic + field type + scoring type combinations
  - [ ] Criteria validation testing (constraint enforcement)
  - [ ] Manual testing with complex Elastic goal scenarios

### Phase 3: Goal Type Matrix Completion
- [ ] **3.1: Comprehensive Goal Type Testing**
  - [ ] Test all Simple + field type + scoring type combinations
  - [ ] Test all Elastic + field type + scoring type combinations
  - [ ] Verify ChecklistGoal integration works correctly
  - [ ] Validate Informational goals continue working as expected

- [ ] **3.2: User Experience Refinements**
  - [ ] Improve field type selection guidance (which types work best for which goals)
  - [ ] Enhance criteria definition UI with examples and help text
  - [ ] Add validation error improvements with specific guidance
  - [ ] Test complete user workflows end-to-end

- [ ] **3.3: Documentation and Integration**
  - [ ] Update goal configuration documentation
  - [ ] Create examples for each goal type + field type combination
  - [ ] Verify integration with entry recording system (for future T007 work)
  - [ ] Final testing and quality assurance

**Technical Implementation Notes:**
- **Pattern Consistency**: All new creators follow SimpleGoalCreator + InformationalGoalCreator patterns
- **Code Reuse**: Leverage FieldTypeSelector, FieldConfig, and FieldValueInput from existing system
- **Validation Strategy**: Use huh's built-in validation plus custom goal validation
- **Flow Design**: Multi-step forms with conditional groups based on scoring type selection
- **Error Handling**: Clear, actionable error messages for criteria and field type mismatches

## 4. Roadblocks

*(Timestamped list of any impediments. AI adds here when a sub-task is marked `[blocked]`)*

## 5. Notes / Discussion Log

**2025-07-12 - AI Initial Analysis:**
- T005 provides excellent foundation with working SimpleGoalCreator and InformationalGoalCreator
- Key issue: Current Simple goals hardcoded to Boolean fields, missing field type selection
- ElasticGoal completely missing but models/validation support exists
- Automatic scoring criteria definition missing for Simple goals
- ChecklistGoal support added in T007/4.1 provides additional reference implementation
- Strategy: Extend existing creators rather than rewrite, maintain idiomatic bubbletea patterns

**2025-07-12 - Design Decisions (T009/1.1):**
1. **Field Type Support**: Simple goals support all field types except checklist (Boolean, Text, Numeric, Time, Duration)
2. **Text Field Restriction**: Text fields restricted to manual scoring only (no automatic text-based criteria)
3. **UI Flow Pattern**: Use multi-step forms (like InformationalGoalCreator) with step omission when not required
4. **Quick Path**: Boolean+Manual remains the streamlined path with minimal steps
5. **Comment Pattern**: Optional comment field accompanies all field types, extend to checklist fields also

**Simple Goal Field Type Matrix (Approved):**
| Field Type | Manual Scoring | Automatic Scoring | Criteria Options | Notes |
|------------|---------------|-------------------|------------------|--------|
| Boolean | ✅ Yes | ✅ Yes | `equals: true` | Traditional pass/fail + quick path |
| Text | ✅ Yes | ❌ No | N/A | Subjective content, manual only |
| Numeric | ✅ Yes | ✅ Yes | `>`, `>=`, `<`, `<=`, `range` | Exercise minutes, reps, etc. |
| Time | ✅ Yes | ✅ Yes | `before`, `after` | Wake-up time, bedtime |
| Duration | ✅ Yes | ✅ Yes | `>`, `>=`, `<`, `<=`, `range` | Exercise duration, meditation |
| Checklist | ❌ N/A | ❌ N/A | N/A | Use ChecklistGoal type instead |

**Enhanced SimpleGoalCreator Flow Design:**
1. **Boolean + Manual (Quick Path)**: Basic Info → Scoring Type (auto-select manual) → Prompt → Save (2 steps)
2. **Boolean + Automatic**: Basic Info → Scoring Type → Criteria (equals: true) → Save (3 steps)
3. **Other Field Types + Manual**: Basic Info → Field Type → Field Config → Scoring Type → Comment/Prompt → Save (4-5 steps)
4. **Other Field Types + Automatic**: Basic Info → Field Type → Field Config → Scoring Type → Criteria → Comment/Prompt → Save (5-6 steps)

**Step Omission Strategy:**
- Skip Field Type step for Boolean goals (maintain current quick path)
- Skip Field Config step when field type needs no configuration
- Skip Criteria step for manual scoring
- Skip Comment step if user doesn't want additional data collection

**Design Considerations:**
- **Field Type Appropriateness**: All field types except checklist supported for Simple goals
- **Automatic vs Manual Scoring**: 
  - Manual: User decides achievement level, field data is informational
  - Automatic: System determines achievement based on field data meeting criteria
- **Text Field Limitation**: Text fields restricted to manual scoring (no automatic text evaluation)
- **Comment Pattern**: Optional comment field for all field types including checklist goals (ChecklistGoalCreator needs enhancement)
- **Backward Compatibility**: Existing Boolean+Manual flow preserved as quick path

**T009/1.2 Implementation Details (2025-07-12):**
- **Multi-step Conversion**: Converted SimpleGoalCreator from sequential groups to multi-step forms
- **Field Type Support**: Added support for Boolean, Text, Numeric (3 subtypes), Time, Duration fields
- **Dynamic Flow**: Flow adjusts based on field type - 3-4 steps depending on configuration needs
- **Field Configuration**: Numeric fields support subtype, unit, min/max constraints; Text supports multiline
- **Scoring Restrictions**: Text fields restricted to manual scoring only (automatic scoring prevented)
- **Comment Integration**: Optional comment field appended to goal description (temporary solution)
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
- **Goal Building**: Complete criteria construction with proper models.Condition structure
- **Range Support**: Inclusive/exclusive range boundaries for numeric and duration criteria
- **Error Handling**: Graceful handling of invalid values and unsupported field types
- **Comprehensive Testing**: 8 additional unit tests covering all criteria types and edge cases

**Technical Integration Points (T009/1.1 Findings):**
- **Existing FieldValueInput System**: Ready for reuse in criteria definition (field_value_input.go)
- **InformationalGoalCreator Patterns**: Multi-step approach with `currentStep` and `initializeStep()` method
- **ChecklistGoalCreator Patterns**: Sequential form groups in single huh.Form (simpler but less flexible)
- **Models Support**: Complete criteria condition support for all field types in models.Condition struct
- **Current SimpleGoalCreator**: 2-step flow, hardcoded Boolean, missing automatic criteria (lines 172, 182)
- **TTY Limitation**: UI framework requires interactive terminal, no piped input testing possible

## 6. Code Snippets & Artifacts 

*(AI will place larger generated code blocks or references to files here if planned / directed. User will then move these to actual project files.)*