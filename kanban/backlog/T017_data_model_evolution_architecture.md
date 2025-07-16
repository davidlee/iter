---
title: "Data Model Evolution Architecture"
type: ["architecture", "enhancement"]
tags: ["data-model", "validation", "extensibility", "coupling", "type-system", "reporting", "periodicity"]
related_tasks: ["T016", "depends-on:T016"]
context_windows: ["internal/models/**/*.go", "internal/parser/**/*.go", "internal/scoring/**/*.go", "internal/ui/entry/**/*.go", "internal/validation/**/*.go"]
---

# Data Model Evolution Architecture

**Context (Background)**:
Analysis of T016 revealed significant architectural vulnerabilities that will impede planned data model extensions. The current validation and type system uses hard-coded assumptions that create tight coupling between components, limiting extensibility. With major features like reporting/analysis and advanced periodicity patterns planned, the architecture must be redesigned to support loose coupling and flexible data model evolution.

**Context (Significant Code Files)**:
- `internal/models/habit.go` - Habit validation with hard-coded type assumptions and rigid coupling
- `internal/models/entry.go` - Entry validation tied to specific habit type patterns
- `internal/scoring/engine.go` - Type-specific scoring methods with proper separation (T016 fix)
- `internal/parser/habits.go` - YAML parsing with implicit type assumptions
- `internal/ui/entry/flow_factory.go` - Clean factory pattern but limited by validation constraints
- `internal/ui/entry/flow_implementations.go` - Fixed type masquerading but still constrained by validation
- Validation logic throughout codebase with restrictive type combination rules

## 1. Habit / User Story

As a developer extending the vice data model, I should be able to add new habit types, field types, scoring methods, and periodicity patterns without requiring extensive validation logic changes throughout the codebase. The architecture should support loose coupling between components and flexible combinations of data model elements to enable feature evolution without architectural refactoring.

## 2. Problem Analysis

### Core Architectural Vulnerabilities (from T016 Analysis)

**1. Overly Restrictive Validation Logic**
- **Issue**: Checklist habits artificially restricted to checklist field types only (`internal/models/habit.go:222-235`)
- **Impact**: Prevents logical combinations like checklist habits with numeric progress tracking
- **Root Cause**: Validation logic assumes one-to-one mapping between habit types and field types
- **Extensibility Impact**: Adding new habit types requires updating multiple validation paths

**2. Hard-coded Type Assumptions in Habit Validation**
- **Issue**: Different validation code paths for Simple vs Elastic habits assume different capabilities
- **Impact**: Creates artificial barriers between habit types that share common features
- **Root Cause**: Type-specific validation instead of capability-based validation
- **Extensibility Impact**: New habit types must fit existing rigid categories

**3. Tight Coupling Between Data Model Components**
- **Issue**: Field type validation tied to specific habit type assumptions
- **Impact**: Changes to one component require coordinated changes across multiple files
- **Root Cause**: Implicit dependencies rather than explicit interfaces
- **Extensibility Impact**: Feature additions become expensive architectural changes

**4. Limited Extensibility for Complex Data Structures**
- **Issue**: Current model assumes single field per habit, simple criteria structures
- **Impact**: Cannot support planned features like multi-field habits or boolean criteria combinations
- **Root Cause**: Data structures designed for current features, not future evolution
- **Extensibility Impact**: Major features require fundamental data model redesign

### Planned Extensions Requiring Architectural Support

**Periodicity and Time Patterns**:
- "X times per Y days" patterns (e.g., 3/7 for 3 times per week)
- Rolling time windows vs fixed calendar periods  
- Custom interval tracking (every N days)
- Complex recurring patterns (every Mon,Tue,Thu)

**Advanced Data Structures**:
- Habits with multiple data fields (composite data collection)
- Boolean combination criteria (AND/OR logic for complex conditions)
- Hierarchical habit relationships (sub-habits, dependencies)

**Reporting and Analysis Requirements**:
- Flexible data access patterns across all habit/field type combinations
- Historical data analysis requiring consistent interfaces
- Cross-habit correlation analysis requiring uniform data access
- Custom aggregation patterns requiring flexible data structures

### Impact on Reporting/Analysis System

**Shared Requirements with Entry Parse/Validate**:
- **Data Access Consistency**: Reporting needs same data model flexibility as entry collection
- **Type System Robustness**: Analysis algorithms must work across all valid combinations
- **Validation Logic Reuse**: Report generation must validate data using same rules as entry collection
- **Performance Requirements**: Both systems need efficient data model access patterns

**Architectural Dependencies**:
- **Loose Coupling Critical**: Reporting cannot be tightly coupled to specific habit types
- **Extensible Interfaces**: Must support new data patterns without reporting system changes
- **Consistent Data Model**: Entry validation and reporting must share consistent data interpretation

## 3. Technical Analysis

### Current Architecture Constraints

**Validation System Issues**:
```go
// Current: Hard-coded type restrictions
if g.HabitType == ChecklistHabit {
    if g.FieldType.Type != ChecklistFieldType {
        return fmt.Errorf("checklist habits must use checklist field type")
    }
}

// Problem: Prevents valid combinations like:
// - Checklist habits with numeric progress tracking
// - Elastic habits with checklist field types for step-by-step scoring
```

**Type System Rigidity**:
```go
// Current: Type-specific validation paths
if g.HabitType == SimpleHabit {
    // Simple habit validation...
} else if g.HabitType == ElasticHabit {
    // Elastic habit validation...
}

// Problem: Adding new habit types requires modifying all validation switch statements
```

**Data Structure Limitations**:
```go
// Current: Single field per habit
type Habit struct {
    FieldType FieldType `yaml:"field_type"`
    // ...
}

// Needed: Multiple fields per habit
type Habit struct {
    FieldTypes []FieldType `yaml:"field_types"`
    // ...
}
```

### Required Architectural Patterns

**1. Capability-Based Validation**
- Replace type-specific validation with capability-based validation
- Habits define capabilities (supports_automatic_scoring, supports_multi_field, etc.)
- Validation logic checks capabilities rather than specific types

**2. Pluggable Validation System**
- Interface-based validation allowing custom validators per habit type
- Validation rules as configuration rather than hard-coded logic
- Composable validation for complex data structures

**3. Flexible Data Model Foundation**
- Support for variable field structures per habit type
- Extensible criteria system supporting boolean combinations
- Temporal pattern abstraction supporting all periodicity requirements

**4. Interface-Based Component Integration**
- Loose coupling between parsing, validation, scoring, and UI components
- Clear contracts for data access supporting both entry collection and reporting
- Extension points for new habit types without core system modifications

## 4. Acceptance Criteria

### Validation System Flexibility
- [ ] Habit types can use any compatible field type without artificial restrictions
- [ ] New habit types can be added without modifying existing validation logic
- [ ] Validation rules are configurable and composable rather than hard-coded
- [ ] Field type compatibility determined by capability rather than type matching

### Data Model Extensibility  
- [ ] Support for habits with multiple data fields (array of FieldType)
- [ ] Boolean combination criteria (AND/OR logic for complex conditions)
- [ ] Extensible periodicity patterns supporting all planned time-based features
- [ ] Backward compatibility with existing single-field habit configurations

### Component Loose Coupling
- [ ] Parsing, validation, scoring, and UI components interact through defined interfaces
- [ ] Changes to data model don't require coordinated changes across all components
- [ ] New habit types can be added with minimal impact on existing components
- [ ] Reporting system can access all data patterns through consistent interfaces

### Reporting/Analysis Support
- [ ] Consistent data access patterns across all habit/field type combinations
- [ ] Validation logic reusable between entry collection and report generation
- [ ] Performance-optimized data structures supporting analysis workloads
- [ ] Extensible aggregation patterns for custom reporting requirements

### Future-Proofing
- [ ] Architecture supports planned periodicity extensions without refactoring
- [ ] Data structures accommodate complex recurring patterns
- [ ] Validation system scales to boolean criteria combinations
- [ ] Component interfaces stable across data model evolution

## 5. Implementation Plan & Progress

**Overall Status:** `Planned`

### Phase 1: Validation System Redesign (Foundation)
**Focus:** Replace hard-coded validation with flexible, capability-based system

- [ ] **Sub-task 1.1:** Design capability-based validation architecture
  - *Design:* Define capability interfaces and validation rule composition patterns
  - *Code/Artifacts:* Validation framework supporting configurable rules
  - *Testing Strategy:* Validation behavior tests for all current habit/field combinations
  - *AI Notes:* Foundation for all other improvements - must be robust and well-tested

- [ ] **Sub-task 1.2:** Implement pluggable validator system
  - *Design:* Interface-based validators with registration and composition mechanisms
  - *Code/Artifacts:* Validator registry and rule composition engine
  - *Testing Strategy:* Custom validator integration tests
  - *AI Notes:* Enables new habit types without core system modifications

- [ ] **Sub-task 1.3:** Migrate existing validation logic to new system
  - *Design:* Preserve all current validation behavior using new flexible system
  - *Code/Artifacts:* Migration of all existing habit type validations
  - *Testing Strategy:* Comprehensive regression testing ensuring no behavior changes
  - *AI Notes:* Critical migration step - must maintain exact current behavior

### Phase 2: Data Model Foundation Extensions (Structure)
**Focus:** Extend data structures to support planned advanced features

- [ ] **Sub-task 2.1:** Multi-field habit support
  - *Design:* Extend Habit struct to support array of FieldType with backward compatibility
  - *Code/Artifacts:* Updated Habit model and migration logic
  - *Testing Strategy:* Single-field compatibility plus multi-field capability tests
  - *AI Notes:* Backward compatibility critical - existing habits must continue working

- [ ] **Sub-task 2.2:** Boolean combination criteria system
  - *Design:* Extensible criteria structures supporting AND/OR logic combinations
  - *Code/Artifacts:* Enhanced criteria evaluation engine
  - *Testing Strategy:* Complex criteria evaluation tests with edge cases
  - *AI Notes:* Foundation for advanced scoring logic in reporting and entry collection

- [ ] **Sub-task 2.3:** Temporal pattern abstraction
  - *Design:* Abstract periodicity system supporting all planned time-based patterns
  - *Code/Artifacts:* Temporal pattern framework with plugin architecture
  - *Testing Strategy:* All current and planned periodicity pattern tests
  - *AI Notes:* Must support both simple and complex recurring patterns

### Phase 3: Component Interface Decoupling (Integration)
**Focus:** Establish loose coupling between major system components

- [ ] **Sub-task 3.1:** Data access interface standardization
  - *Design:* Define consistent interfaces for data access across components
  - *Code/Artifacts:* Interface definitions and adapter implementations
  - *Testing Strategy:* Interface contract tests ensuring consistent behavior
  - *AI Notes:* Critical for reporting system integration

- [ ] **Sub-task 3.2:** Parser and validation integration
  - *Design:* Decouple parsing from validation using new flexible validation system
  - *Code/Artifacts:* Updated parser with pluggable validation integration
  - *Testing Strategy:* Parser behavior tests with various validation configurations
  - *AI Notes:* Enables custom habit types with custom validation rules

- [ ] **Sub-task 3.3:** Scoring engine interface generalization
  - *Design:* Generalize scoring interfaces to support all habit types and data structures
  - *Code/Artifacts:* Enhanced scoring engine with flexible habit type support
  - *Testing Strategy:* Scoring behavior tests across all habit type combinations
  - *AI Notes:* Build on T016 success - maintain proper type separation while enabling flexibility

### Phase 4: Reporting Foundation Integration (Application)
**Focus:** Validate architecture by implementing reporting system foundation

- [ ] **Sub-task 4.1:** Reporting data access implementation
  - *Design:* Implement reporting data access using new standardized interfaces
  - *Code/Artifacts:* Reporting data layer with flexible habit/field type support
  - *Testing Strategy:* Data access tests across all valid combinations
  - *AI Notes:* Proof of architecture - reporting should work with minimal coupling to specific types

- [ ] **Sub-task 4.2:** Cross-component validation testing
  - *Design:* Comprehensive integration tests validating loose coupling success
  - *Code/Artifacts:* Integration test suite covering all component interactions
  - *Testing Strategy:* End-to-end tests ensuring components work together with new architecture
  - *AI Notes:* Validation that architectural habits have been achieved

## 6. Roadblocks

*(None currently - planning phase)*

## 7. Notes / Discussion Log

- `2025-07-14 - User:` Requested focus on architectural vulnerabilities for data model evolution, emphasizing reporting/analysis requirements and planned extensions (periodicity patterns, multi-field habits, boolean criteria)
- `2025-07-14 - AI:` Analyzed T016 findings in context of planned features. Key insight: current validation system's hard-coded assumptions will severely impede planned extensions. Reporting system shares critical requirements with entry collection system, making loose coupling and flexible data model access essential. Designed 4-phase approach: validation redesign → data model extensions → component decoupling → reporting integration validation.

### Key Architectural Principles

**1. Capability-Based Design**:
- Replace type-specific logic with capability-based validation and behavior
- Enable arbitrary combinations of habit types, field types, and scoring methods where logically valid
- Support feature extension without modifying core component logic

**2. Interface-Driven Architecture**:
- Define clear contracts between parsing, validation, scoring, and UI components
- Enable component evolution without breaking dependent systems
- Support plugin-based extension for new habit types and validation rules

**3. Data Model Flexibility**:
- Design data structures to accommodate current and planned feature requirements
- Maintain backward compatibility while enabling advanced data patterns
- Support efficient access patterns required by both entry collection and reporting

**4. Validation as Configuration**:
- Move validation rules from hard-coded logic to configurable, composable rules
- Enable custom validation patterns for new habit types
- Support complex validation scenarios (multi-field habits, boolean criteria) through rule composition

### Implementation Success Metrics

**Extensibility Test**: Adding a new habit type should require:
- New habit type definition and capabilities declaration
- Custom validation rules (if needed) through plugin system  
- NO modifications to existing parsing, scoring, or UI component logic

**Coupling Test**: Reporting system implementation should:
- Access all habit/field type combinations through consistent interfaces
- Work with new habit types without reporting system modifications
- Reuse validation logic without duplicating business rules

**Evolution Test**: Adding multi-field habits should:
- Require only data model structure changes and validation rule updates
- Work with existing UI components through interface abstraction
- Support both simple migration (single-field) and advanced usage (multi-field)