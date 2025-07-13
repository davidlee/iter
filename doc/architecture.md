# Architecture.md

## Table of Contents

1. [References](#references)
2. [High-Level Architecture Summary](#high-level-architecture-summary)
   - 2.1 [Core Components](#21-core-components)
   - 2.2 [Data Flow Architecture](#22-data-flow-architecture)
   - 2.3 [Key Architectural Decisions](#23-key-architectural-decisions)
3. [Data Architecture](#1-data-architecture)
   - 3.1 [YAML Schema Structure](#11-yaml-schema-structure)
   - 3.2 [Goal Types and Field Types](#12-goal-types-and-field-types)
   - 3.3 [Entry Storage and Historical Preservation](#13-entry-storage-and-historical-preservation)
   - 3.4 [Specialized Storage: Checklist System](#14-specialized-storage-checklist-system)
4. [Component Architecture](#2-component-architecture)
   - 4.1 [Package Organization](#21-package-organization)
   - 4.2 [Parser Layer Architecture](#22-parser-layer-architecture)
   - 4.3 [Models Layer: Data Structures and Validation](#23-models-layer-data-structures-and-validation)
   - 4.4 [Storage Layer: Atomic Operations](#24-storage-layer-atomic-operations)
   - 4.5 [Scoring Engine Architecture](#25-scoring-engine-architecture)
5. [User Interface Architecture](#3-user-interface-architecture)
   - 5.1 [Hybrid UI Strategy](#31-hybrid-ui-strategy)
   - 5.2 [Form Generation and Field Type Adaptation](#32-form-generation-and-field-type-adaptation)
   - 5.3 [Navigation and Progress Tracking](#33-navigation-and-progress-tracking)
   - 5.4 [Error Handling and Validation Feedback](#34-error-handling-and-validation-feedback)
6. [Integration Patterns](#4-integration-patterns)
   - 6.1 [CLI Command Structure and Routing](#41-cli-command-structure-and-routing)
   - 6.2 [Configuration Management](#42-configuration-management)
   - 6.3 [File Initialization and Sample Data](#43-file-initialization-and-sample-data)
   - 6.4 [Testing Strategies](#44-testing-strategies)
7. [Extension Points](#5-extension-points)
   - 7.1 [Adding New Goal Types](#51-adding-new-goal-types)
   - 7.2 [Field Type Extensions](#52-field-type-extensions)
   - 7.3 [Scoring Criteria Extensions](#53-scoring-criteria-extensions)
   - 7.4 [Storage Format Evolution](#54-storage-format-evolution)

## References

This section provides context on the key documentation files that inform the architecture:

- **[CLAUDE.md](./CLAUDE.md)** - Primary development guide with design principles, dependencies, and standards. Essential for understanding the clean architecture approach and charmbracelet UI framework usage.

- **[initial_brief.md](./initial_brief.md)** - Original project vision and requirements. Defines core goals: low-friction entry, flexibility, resilience to schema changes, and text-based interoperability.

- **[goal_schema.md](./doc/specifications/goal_schema.md)** - Complete specification of the YAML-based goal configuration format. Critical for understanding data structures and validation rules.

- **[T001_minimal_end_to_end_release.md](./T001_minimal_end_to_end_release.md)** - Foundation implementation covering project setup, configuration management, goal parsing, entry collection, and CLI interface. Shows the core architectural decisions.

- **[T003_implement_elastic_goals_end_to_end.md](./T003_implement_elastic_goals_end_to_end.md)** - Elastic goals with mini/midi/maxi achievement levels. Demonstrates the scoring engine architecture and strategy pattern for goal handlers.

- **[T005_goal_configuration_ui.md](./T005_goal_configuration_ui.md)** - Interactive goal creation system with bubbletea wizards and huh forms. Shows the hybrid UI approach for simple vs complex interactions.

- **[T007_dynamic_checklist_system.md](./T007_dynamic_checklist_system.md)** - Checklist goals with dynamic item management. Illustrates separation between templates (checklists.yml) and instances (checklist_entries.yml).

- **[T010_iter_entry_ui_system.md](./T010_iter_entry_ui_system.md)** - Comprehensive entry collection system with field-type awareness and goal-type adaptation. Shows the strategy pattern for entry handlers and immediate scoring feedback.

- **[flow_analysis_T005.md](./flow_analysis_T005.md)** - Detailed UX flow analysis for goal configuration. Documents the evolution from simple huh forms to enhanced bubbletea wizards.

## High-Level Architecture Summary

**iter** is a CLI habit tracker designed around three core principles: **low-friction entry**, **schema resilience**, and **text-based interoperability**. The architecture follows clean separation of concerns with a focus on maintaining user data integrity as goals evolve over time.

### Core Components

1. **Schema Management Layer** - YAML-based goal definitions with validation and automatic ID persistence
2. **Entry Collection Layer** - Interactive CLI for recording daily habit data with immediate scoring feedback  
3. **Storage Layer** - Text-based persistence with atomic operations and backup strategies
4. **UI Framework** - Hybrid approach using charmbracelet libraries (huh for forms, bubbletea for complex flows)
5. **Scoring Engine** - Automatic evaluation of entries against goal criteria with achievement levels

### Data Flow Architecture

```
goals.yml (schema) → Parser → Validation → UI Generation → Entry Collection → Scoring → entries.yml (data)
                                     ↓
                             checklists.yml → checklist_entries.yml (specialized storage)
```

### Key Architectural Decisions

- **Text-First Storage**: YAML files as primary storage for version control compatibility and user transparency
- **Strategy Pattern**: Goal type and field type handlers for extensible entry collection
- **Hybrid UI**: Simple huh forms for basic interactions, bubbletea wizards for complex multi-step flows
- **Separation of Concerns**: Clear boundaries between schema definition, entry collection, scoring, and storage
- **Resilience Design**: Historical entries preserved through schema changes via stable goal IDs

## Proposed Detailed Sections

I propose organizing the detailed architecture documentation into these sections:

### 1. **Data Architecture**
   - YAML schema structure and validation
   - Goal types (simple, elastic, informational, checklist) and field types
   - Entry storage patterns and historical data preservation
   - ID generation and persistence strategies

### 2. **Component Architecture** 
   - Package organization and dependency relationships
   - Parser layer (goal schema, checklist management)
   - Models layer (data structures and validation)
   - Storage layer (atomic operations, backup strategies)
   - UI layer (form generation, wizard flows)
   - Scoring engine (criteria evaluation, achievement calculation)

### 3. **User Interface Architecture**
   - Hybrid UI strategy (huh vs bubbletea decision matrix)
   - Form generation patterns and field type adaptation
   - Navigation and progress tracking in multi-step flows
   - Error handling and validation feedback

### 4. **Integration Patterns**
   - CLI command structure and routing
   - Configuration management (XDG compliance, override flags)
   - File initialization and sample data generation
   - Testing strategies (headless testing, integration patterns)

### 5. **Extension Points**
   - Adding new goal types and field types
   - Scoring criteria extensions
   - UI component reuse patterns
   - Storage format evolution strategies

## 1. Data Architecture

### 1.1 YAML Schema Structure

The iter application uses a declarative YAML-based schema for goal definitions, designed for human readability and version control compatibility:

```yaml
version: "1.0.0"
created_date: "2024-01-01"
goals:
  - title: "Daily Exercise"
    id: "daily_exercise"  # Auto-generated if missing
    goal_type: "elastic"
    field_type:
      type: "duration"
      format: "minutes"
    scoring_type: "automatic"
    mini_criteria: { condition: { greater_than_or_equal: 15 } }
    midi_criteria: { condition: { greater_than_or_equal: 30 } }
    maxi_criteria: { condition: { greater_than_or_equal: 60 } }
```

**Schema Validation Pipeline:**
1. YAML parsing with structure validation
2. Goal type and field type compatibility checks  
3. Scoring criteria validation (mini ≤ midi ≤ maxi for numeric types)
4. Automatic ID generation and persistence for missing IDs
5. Cross-goal uniqueness validation

### 1.2 Goal Types and Field Types

**Goal Types:**
- **Simple**: Binary pass/fail goals with boolean or single-value fields
- **Elastic**: Three-tier achievement goals (mini/midi/maxi) with criteria-based scoring
- **Informational**: Data collection without pass/fail evaluation, supports direction preferences
- **Checklist**: Dynamic item completion with template/instance separation

**Field Types with Data Validation:**
- `boolean` - True/false values with flexible input parsing (yes/no, y/n, 1/0)
- `text` - String values with optional multiline support
- `unsigned_int`, `unsigned_decimal`, `decimal` - Numeric values with units and constraints
- `time` - HH:MM format with 24-hour validation
- `duration` - Flexible duration parsing (30m, 1h30m, 90m)
- `checklist` - References to external checklist definitions

### 1.3 Entry Storage and Historical Preservation

**Entry Data Structure:**
```yaml
version: "1.0.0"
entries:
  "2024-01-15":
    daily_exercise:
      value: 45  # Raw user input
      achievement_level: "midi"  # Computed achievement
      notes: "Morning run in the park"
      completed_at: "2024-01-15T07:30:00Z"
```

**Historical Data Resilience:**
- **Stable Goal IDs**: Generated once and persisted, survive title changes
- **Schema Evolution**: Orphaned fields preserved as "historical data"
- **Scoring Context**: Achievement levels reflect criteria active on entry date
- **Atomic Operations**: File writes use temporary files with atomic moves

### 1.4 Specialized Storage: Checklist System

**Template/Instance Separation:**
- `checklists.yml` - Reusable checklist templates with items and metadata
- `checklist_entries.yml` - Daily completion state by date and checklist ID

**Data Model:**
```yaml
# checklists.yml
checklists:
  - id: "morning_routine"
    title: "Morning Routine"
    items:
      - "# clean station: physical inputs (~5m)"  # Heading
      - "clear desk"                              # Item
      - "clear desk inbox, loose papers"

# checklist_entries.yml  
entries:
  "2024-01-15":
    morning_routine:
      completed_items:
        "clear desk": true
        "clear desk inbox, loose papers": false
      completion_time: "2024-01-15T08:15:00Z"
```

## 2. Component Architecture

### 2.1 Package Organization

```
iter/
├── cmd/                    # CLI commands and routing
├── internal/
│   ├── config/            # XDG-compliant path resolution
│   ├── models/            # Data structures and validation
│   ├── parser/            # YAML parsing and file operations
│   ├── storage/           # Entry persistence and atomic operations
│   ├── scoring/           # Criteria evaluation engine
│   ├── ui/                # User interface components
│   │   ├── goalconfig/    # Goal creation wizards and forms
│   │   └── checklist/     # Checklist management UI
│   └── init/              # File initialization and samples
└── doc/                   # Specifications and documentation
```

### 2.2 Parser Layer Architecture

**GoalParser** (`internal/parser/goals.go`):
- YAML marshaling/unmarshaling with goccy/go-yaml
- Schema validation with automatic ID generation
- Atomic file operations with backup on corruption
- ID persistence detection and conditional saves

**ChecklistParser** (`internal/parser/checklist_parser.go`):
- Template management (CRUD operations)
- Entry state persistence separate from templates
- Heading item filtering (prefixed with "# ")

### 2.3 Models Layer: Data Structures and Validation

**Core Models** (`internal/models/`):
```go
type Goal struct {
    Title       string      `yaml:"title"`
    ID          string      `yaml:"id,omitempty"`
    GoalType    GoalType    `yaml:"goal_type"`
    FieldType   FieldType   `yaml:"field_type"`
    ScoringType ScoringType `yaml:"scoring_type"`
    // Goal-type specific criteria fields
}

type GoalEntry struct {
    Value            interface{}        `yaml:"value"`
    AchievementLevel *AchievementLevel `yaml:"achievement_level,omitempty"`
    Notes            string            `yaml:"notes,omitempty"`
    CompletedAt      string            `yaml:"completed_at,omitempty"`
}
```

**Validation Strategy:**
- **Field-level validation**: Type checking, format validation, constraint enforcement
- **Cross-field validation**: Criteria ordering (mini ≤ midi ≤ maxi), goal type compatibility
- **Schema-level validation**: ID uniqueness, version compatibility

### 2.4 Storage Layer: Atomic Operations

**EntryStorage** (`internal/storage/entries.go`):
- **Thread-safe operations**: Mutex protection for concurrent access
- **Atomic writes**: Temporary file + atomic move pattern
- **Backup strategy**: Preserve corrupted files for recovery
- **Query interface**: Date ranges, goal-specific lookups, today helpers

**File Operation Pattern:**
```go
// Atomic write pattern used throughout
tempFile := targetFile + ".tmp"
writeToFile(tempFile, data)
os.Rename(tempFile, targetFile)  // Atomic on most filesystems
```

### 2.5 Scoring Engine Architecture

**ScoringEngine** (`internal/scoring/engine.go`):
```go
type Engine struct {
    // Stateless evaluation engine
}

func (e *Engine) ScoreElasticGoal(goal models.Goal, value interface{}) 
    (*models.AchievementLevel, error) {
    // Value conversion → Criteria evaluation → Achievement calculation
}
```

**Evaluation Pipeline:**
1. **Value Conversion**: Interface{} → typed values (numeric, duration, time, boolean, text)
2. **Criteria Evaluation**: Apply conditions (greater_than, less_than, equals, range)
3. **Achievement Calculation**: Determine highest achieved level (none/mini/midi/maxi)
4. **Error Handling**: Graceful fallback for incompatible value types

## 3. User Interface Architecture

### 3.1 Hybrid UI Strategy

**Decision Matrix for UI Framework Selection:**

| Interaction Type | Framework | Rationale |
|------------------|-----------|-----------|
| Simple confirmations | huh.NewConfirm() | Minimal overhead, perfect for yes/no |
| Basic input collection | huh.NewInput() | Built-in validation, single-step |
| Complex multi-step flows | bubbletea + huh | Navigation, progress, state management |
| Configuration wizards | bubbletea model | Enhanced UX with back/forward navigation |

**Integration Pattern - Embedding huh in bubbletea:**
```go
type FormStepModel struct {
    form    *huh.Form
    title   string
    step    int
    total   int
}

func (m FormStepModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    form, cmd := m.form.Update(msg)
    m.form = form.(*huh.Form)
    return m, cmd
}
```

### 3.2 Form Generation and Field Type Adaptation

**FieldValueInputFactory** (`internal/ui/goalconfig/field_value_input.go`):
```go
type FieldValueInput interface {
    Render() string
    GetValue() interface{}
    SetExistingValue(value interface{}) error
    Validate() error
}

func CreateFieldInput(fieldType models.FieldType) FieldValueInput {
    switch fieldType.Type {
    case "boolean":  return NewBooleanInput()
    case "numeric":  return NewNumericInput(fieldType.Unit, fieldType.Min, fieldType.Max)
    case "duration": return NewDurationInput()
    // ...
    }
}
```

**Field Type Specific Behaviors:**
- **Boolean**: `huh.NewConfirm()` with clear yes/no prompting
- **Numeric**: `huh.NewInput()` with unit display, min/max validation, subtype awareness
- **Duration**: `huh.NewInput()` with flexible parsing hints (30m, 1h30m, 90m)
- **Time**: `huh.NewInput()` with HH:MM format validation
- **Text**: `huh.NewInput()` or `huh.NewText()` based on multiline configuration

### 3.3 Navigation and Progress Tracking

**Wizard State Management:**
```go
type WizardState interface {
    GetCurrentStep() int
    GetTotalSteps() int
    CanGoBack() bool
    CanGoForward() bool
    GetStepData(index int) interface{}
}

type StepHandler interface {
    Render(state WizardState) string
    Update(msg tea.Msg, state WizardState) (WizardState, tea.Cmd)
    Validate(state WizardState) []ValidationError
}
```

**Progress Indicators:**
- Step counters: "Step 3 of 6" 
- Breadcrumb navigation with completed step checkmarks
- Real-time validation status with contextual error display
- Achievement level preview for elastic goals

### 3.4 Error Handling and Validation Feedback

**Validation Framework:**
- **Real-time validation**: As-you-type feedback for format errors
- **Cross-step validation**: Criteria ordering checks (mini ≤ midi ≤ maxi)
- **Contextual help**: Dynamic help text based on field type and current input
- **Recovery mechanisms**: State preservation during validation failures

## 4. Integration Patterns

### 4.1 CLI Command Structure and Routing

**Cobra-based Command Hierarchy:**
```
iter
├── entry                    # Entry collection
├── goal                     # Goal management
│   ├── add [--dry-run]     # Interactive goal creation
│   ├── list                # Goal display
│   ├── edit                # Goal modification
│   └── remove              # Goal deletion
├── list                     # Checklist management
│   ├── add [id]            # Checklist creation
│   ├── edit <id>           # Checklist editing
│   └── entry [id]          # Checklist completion
└── validate                 # Schema validation
```

**Command Integration Points:**
- **Dependency Injection**: Parsers, storage, and UI components injected into commands
- **Error Propagation**: Consistent error handling with user-friendly messages
- **Configuration**: XDG-compliant paths with `--config-dir` override support

### 4.2 Configuration Management

**XDG Base Directory Compliance:**
```go
type Paths struct {
    ConfigDir        string  // ~/.config/iter/
    GoalsFile        string  // goals.yml
    EntriesFile      string  // entries.yml 
    ChecklistsFile   string  // checklists.yml
    ChecklistEntries string  // checklist_entries.yml
}
```

**Override Mechanisms:**
- `--config-dir` flag for testing and alternative configurations
- Environment variables: `XDG_CONFIG_HOME` support
- Graceful fallback to sensible defaults

### 4.3 File Initialization and Sample Data

**FileInitializer** (`internal/init/files.go`):
- **Sample Goals**: "Morning Exercise" (simple), "Daily Reading" (elastic)
- **Empty Structures**: Properly formatted YAML with version headers
- **User Guidance**: Comments and examples in generated files
- **Atomic Creation**: Check-and-create pattern to avoid overwrites

### 4.4 Testing Strategies

**Headless Testing Architecture:**
```go
// UI components provide testing constructors
func NewSimpleGoalCreatorForTesting() *SimpleGoalCreator
func NewEntryCollectorForTesting(goals []Goal) *EntryCollector

// Business logic exposed for direct testing
func (c *SimpleGoalCreator) CreateGoalDirectly(data SimpleGoalData) (*Goal, error)
```

**Testing Pyramid:**
- **Unit Tests**: Individual component behavior, validation logic
- **Integration Tests**: Component collaboration, file I/O, scoring engine
- **Compatibility Tests**: Real user data patterns, schema evolution
- **Manual UI Testing**: Interactive terminal verification (limited automation)

## 5. Extension Points

### 5.1 Adding New Goal Types

**Goal Type Extension Pattern:**
1. **Model Extension**: Add new goal type constant and validation rules
2. **Parser Support**: YAML structure definition and parsing logic
3. **UI Handler**: Implement `GoalEntryHandler` interface for entry collection
4. **Scoring Integration**: Extend scoring engine for new criteria types
5. **Configuration UI**: Add goal creation wizard or form components

**Example - Checklist Goal Addition:**
```go
// 1. Model extension
const ChecklistGoal GoalType = "checklist"

// 2. UI handler implementation  
type ChecklistGoalHandler struct {
    checklistParser *parser.ChecklistParser
    scoringEngine   *scoring.Engine
}

func (h *ChecklistGoalHandler) CollectEntry(goal Goal, existing *ExistingEntry) (*EntryResult, error) {
    // Checklist-specific entry collection logic
}
```

### 5.2 Field Type Extensions

**Field Type Addition Process:**
1. **Model Definition**: Add field type constant and validation
2. **Input Component**: Implement `FieldValueInput` interface
3. **Factory Integration**: Add case to `CreateFieldInput()`
4. **Scoring Support**: Extend criteria evaluation for new type
5. **UI Integration**: Form generation and validation patterns

### 5.3 Scoring Criteria Extensions

**Criteria Extension Architecture:**
```go
type Condition struct {
    // Existing criteria
    GreaterThan           *float64 `yaml:"greater_than,omitempty"`
    LessThan             *float64 `yaml:"less_than,omitempty"`
    // New criteria types
    PeriodicityCondition *PeriodicityCondition `yaml:"periodicity,omitempty"`
    CustomCondition      *CustomCondition      `yaml:"custom,omitempty"`
}
```

**Extension Requirements:**
- YAML schema updates with backwards compatibility
- Scoring engine evaluation logic for new condition types
- Validation rules for new criteria structures
- UI components for criteria configuration

### 5.4 Storage Format Evolution

**Migration Strategy:**
- **Version Headers**: Track schema version in all YAML files
- **Backwards Compatibility**: Parser handles older versions gracefully
- **Migration Scripts**: Automated conversion for breaking changes
- **Orphaned Data Preservation**: Historical entries maintained through migrations

**Example Migration Pattern:**
```go
func MigrateSchema(fromVersion, toVersion string, data []byte) ([]byte, error) {
    switch fromVersion {
    case "1.0.0":
        return migrateFrom1_0_0(data), nil
    default:
        return data, nil  // No migration needed
    }
}
```