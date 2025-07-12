---
id: T007
title: Dynamic Checklist System with Goal Integration
priority: medium
status: backlog
created: 2025-07-12
related_tasks: ["T005"]
---

# T007: Dynamic Checklist System with Goal Integration

## 1. Goal

Extend the static checklist prototype (`iter checklist`) to support configurable checklists stored in `checklists.yml` and integrated with the goal system. Enable checklists to be used as goal types with automatic or manual scoring based on completion criteria.

## 2. Acceptance Criteria

- [ ] Create `checklists.yml` configuration file format for storing checklist definitions
- [ ] Implement checklist management commands: `iter list add`, `iter list edit`, `iter list entry`
- [ ] Add `ChecklistGoal` as a new goal type in the existing goal system
- [ ] Support automatic scoring when all checklist items are completed
- [ ] Support manual scoring for partial checklist completion
- [ ] Maintain backward compatibility with existing goal types
- [ ] Provide seamless UI experience for checklist selection and completion
- [ ] Store checklist completion state as entry data

## 3. Implementation Plan & Progress

### Phase 1: Data Model & Configuration (Priority: High)
- [x] 1.1: Define checklist YAML schema and data structures
- [x] 1.2: Create checklist parser for loading/saving `checklists.yml`
- [x] 1.3: Extend goal models to support checklist goal type
  - refer to [doc/specifications/goal_schema.md]; update it as required
- [x] 1.4: Add checklist validation logic

### Phase 2: Checklist Management Commands (Priority: High)
- [x] 2.1: Implement `iter list add $id` command
  - Simple multiline text field UI with note about "# " prefix for headings
  - Parse input into checklist items array and save to checklists.yml
  - Basic validation and ID generation
- [x] 2.2: Implement `iter list edit $id` command
  - Load existing checklist and populate multiline text field
  - Reuse same UI as add command with pre-filled content
  - Update existing checklist in checklists.yml
- [x] 2.3: Implement `iter list entry $id` (direct access)
  - Adapt existing internal/ui/checklist.go prototype with minimal changes
  - Load checklist by ID and populate items from checklists.yml
  - Save completion state (item text -> boolean map) for entry recording
- [x] 2.4: Implement `iter list entry` (menu selection)
  - Present list of available checklist IDs/titles for selection
  - On selection, invoke same logic as `iter list entry $id`
  - Handle empty checklists.yml gracefully
- [x] 2.5: Review implementation and consider refactoring opportunities
  - Evaluate code reuse between add/edit commands
  - Assess checklist UI integration patterns
  - Identify any architectural improvements needed for Phase 3

### Phase 3: Checklist Entry Persistence & UX Refinements (Priority: High)
- [x] 3.1: Make checklist ID optional in `iter list add` command
  - Generate ID from title using same logic as goals
  - Update editor UI to prompt for title first, then generate ID
- [x] 3.2: Implement checklist_entries.yml for persistent completion tracking
  - Create data model for daily checklist completion by date & checklist ID
  - Add checklist entry parser for loading/saving completion state
  - Store completion state separate from goal entries to avoid clutter
- [x] 3.3: Update entry command to persist and restore completion state
  - Save completion state to checklist_entries.yml on exit
  - Restore previous completion state when re-entering same checklist on same day
  - Handle date transitions properly (new day = fresh state)
- [x] 3.4: Add ChecklistEntriesFile to config paths and initialization

### Phase 4: Goal Integration (Priority: High)
- [x] 4.1: Add ChecklistGoal support to goal configuration UI
- [ ] 4.2: Implement automatic scoring for checklist completion
- [ ] 4.3: Implement manual scoring support
- [ ] 4.4: Add checklist criteria validation

### Phase 5: Enhanced UI & Experience (Priority: Medium)
- [ ] 5.1: Add progress indicators to checklist headings (e.g., "clean station (3/5)")
- [ ] 5.2: Add entry recording integration for checklist goals
- [ ] 5.3: Add checklist completion summary and statistics

**Overall Status**: `[in_progress]`

## 4. Technical Design

### YAML Data Format

#### checklists.yml Structure
```yaml
version: "1.0.0"
created_date: "2024-01-01"
checklists:
  - id: "morning_routine"
    title: "Morning Routine"
    description: "Daily morning checklist for productivity setup"
    items:
      - "# clean station: physical inputs (~5m)"
      - "clear desk"
      - "clear desk inbox, loose papers, notebook"
      - "# clean station: digital inputs (~10m)"
      - "process emails (inbox)"
      - "phone notifications"
      - "browsers (all devices)"
      - "editors, apps"
      - "review periodic notes"
      - "log actions"
      - "# straighten & reset (~5m)"
      - "desk"
      - "digital workspace"
    created_date: "2024-01-01"
    modified_date: "2024-01-01"
```

#### goals.yml Integration
```yaml
# Example checklist goal with automatic scoring
- title: "Morning Setup"
  goal_type: "checklist"
  field_type:
    type: "checklist"
    checklist_id: "morning_routine"
  scoring_type: "automatic"
  criteria:
    description: "All items completed"
    condition:
      checklist_completion:
        required_items: "all"  # only valid option

# Example checklist goal with manual scoring  
- title: "Weekly Review"
  goal_type: "checklist"
  field_type:
    type: "checklist"
    checklist_id: "weekly_review"
  scoring_type: "manual"
```

### Data Structures

#### Checklist Models
```go
// Checklist represents a reusable checklist template
// Items are stored as simple strings, with headings prefixed by "# "
type Checklist struct {
    ID           string   `yaml:"id"`
    Title        string   `yaml:"title"`
    Description  string   `yaml:"description,omitempty"`
    Items        []string `yaml:"items"`               // Simple array of strings
    CreatedDate  string   `yaml:"created_date"`
    ModifiedDate string   `yaml:"modified_date"`
}

// ChecklistCompletion stores completion state for entries
// Stores item text -> completion for comprehensive historical data
type ChecklistCompletion struct {
    ChecklistID     string            `yaml:"checklist_id"`
    CompletedItems  map[string]bool   `yaml:"completed_items"` // item text -> completed
    CompletionTime  string            `yaml:"completion_time,omitempty"`
    PartialComplete bool              `yaml:"partial_complete"`
}

// ChecklistSchema represents the checklists.yml file structure
type ChecklistSchema struct {
    Version     string      `yaml:"version"`
    CreatedDate string      `yaml:"created_date"`
    Checklists  []Checklist `yaml:"checklists"`
}
```

#### Goal System Extensions
```go
// Add to existing FieldType constants
const (
    ChecklistFieldType = "checklist"
)

// Add to existing GoalType constants
const (
    ChecklistGoal GoalType = "checklist"
)

// Extend FieldType struct
type FieldType struct {
    // ... existing fields ...
    ChecklistID string `yaml:"checklist_id,omitempty"` // Reference to checklist
}

// Extend Condition struct for checklist-specific criteria
type Condition struct {
    // ... existing fields ...
    ChecklistCompletion *ChecklistCompletionCondition `yaml:"checklist_completion,omitempty"`
}

type ChecklistCompletionCondition struct {
    RequiredItems string `yaml:"required_items"` // "all" (only valid option)
}
```

### File Structure

```
internal/
├── models/
│   ├── checklist.go           # Checklist data structures
│   └── goal.go                # Extended goal models (existing)
├── parser/
│   ├── checklist_parser.go    # YAML parsing for checklists
│   └── goal_parser.go         # Extended goal parser (existing)
├── ui/
│   ├── checklist/
│   │   ├── manager.go         # Checklist management UI
│   │   ├── editor.go          # Checklist creation/editing
│   │   ├── selector.go        # Checklist selection menu
│   │   └── completion.go      # Interactive checklist completion
│   ├── goalconfig/            # Extended goal configuration (existing)
│   └── checklist.go           # Enhanced checklist UI (existing)
├── commands/
│   └── list.go                # List management commands
└── storage/
    └── checklist_storage.go   # Checklist persistence layer
```

### Command Structure

```bash
# Checklist management
iter list add morning_routine          # Create new checklist with multiline text UI
iter list edit morning_routine         # Edit existing checklist (reuse add UI)
iter list rm morning_routine           # Remove checklist

# Checklist completion
iter list entry                        # Show menu of available checklists, then enter selected
iter list entry morning_routine        # Complete specific checklist (adapted from prototype)
iter list show morning_routine         # Display checklist without interaction

# Goal integration (through existing goal commands)
iter goal add                          # Extended to support checklist goals
iter entry                            # Extended to handle checklist entry recording
```

## 5. Dependencies & Integration Points

### Depends On
- T005 (Goal Configuration UI) - for checklist goal configuration
- Existing goal system models and validation
- Existing entry recording system patterns

### Integrates With
- Goal validation system for checklist-specific validation
- Entry recording system for storing completion state
- YAML parsing infrastructure from existing parsers

### May Block
- Future entry recording tasks - checklist goals will be available as entry types

## 6. Design Considerations

### Data Persistence
- `checklists.yml` stores checklist templates (reusable)
- Entry data stores completion state (date-specific)
- Maintain separation between template and instance data

### User Experience
- Leverage existing bubbletea UI patterns from checklist prototype
- Consistent command structure with existing `iter` commands
- Progressive disclosure: simple cases work simply, complex cases supported

### Backward Compatibility
- Existing goal types remain unchanged
- New checklist field type is additive
- Graceful degradation when `checklists.yml` missing

### Scoring Flexibility
- Automatic scoring: useful for binary completion tracking
- Manual scoring: supports partial completion and subjective assessment
- Extensible criteria system for future enhancements

## 7. Testing Strategy

- Unit tests for checklist YAML parsing and validation
- Integration tests for goal system extensions
- UI testing for checklist management commands
- End-to-end tests for goal entry recording with checklists
- Edge case testing for malformed checklist data

## 8. Future Extensions

- Checklist templates and inheritance
- Time-based checklist items (scheduled completion)
- Checklist analytics and completion trends
- Import/export of checklist definitions
- Nested checklists and dependencies

---

## Notes / Discussion Log

*Initial task creation based on existing checklist prototype and TODO comments.*

**Phase 1 Complete (2025-07-12):**
- Implemented simplified checklist data structures using string arrays (matching existing UI)
- Created comprehensive checklist models with validation in `internal/models/checklist.go`
- Extended goal system to support checklist goals with new field type and goal type
- Added checklist parser with full CRUD operations in `internal/parser/checklist_parser.go`
- Simplified completion criteria to "all items complete" or manual scoring (no percentage scoring)
- Changed completion storage to map item text to boolean for better historical data
- Updated goal schema specification to document checklist field type and criteria

**Phase 2 Planning (2025-07-12):**
- Detailed UI approach: multiline text field for add/edit with "# " heading instructions
- Command sequence: 2.3 (direct entry) before 2.4 (menu selection) for logical flow
- Prototype reuse: adapt existing internal/ui/checklist.go with minimal changes for entry commands
- Added 2.5 review subtask to evaluate refactoring opportunities before Phase 3

**Phase 2 Complete (2025-07-12):**
- Implemented complete checklist management command suite: add, edit, entry (with/without ID)
- Created reusable UI components in internal/ui/checklist/ package
- Successfully adapted existing prototype with minimal changes for dynamic data
- Added ChecklistsFile to config paths for proper file management
- Clean architecture with good separation of concerns and code reuse
- Ready for Phase 3 UX refinements

**Phase 3 Complete (2025-07-12):**
- Made checklist ID optional in `iter list add` command with automatic generation from title
- Implemented comprehensive checklist_entries.yml persistence system for daily completion tracking
- Enhanced entry command to save/restore completion state on same-day re-entry
- Added ChecklistEntriesFile to config paths with proper initialization
- Separation of checklist templates (checklists.yml) from completion instances (checklist_entries.yml)
- Ready for Phase 4 goal system integration

**Phase 4.1 Complete (2025-07-12):**
- Added ChecklistGoal support to goal configuration UI following existing patterns
- Created ChecklistGoalCreator component with checklist selection and scoring configuration
- Extended GoalConfigurator to handle ChecklistGoal type with new switch case
- Added "Checklist (Complete checklist items)" option to goal type selection
- Implemented automatic and manual scoring modes for checklist goals
- Added WithChecklistsFile() method to configure checklists.yml path
- Updated goal_add command to pass ChecklistsFile path to configurator
- Comprehensive unit tests covering all functionality and edge cases
- All code properly formatted and linted according to project standards

## Roadblocks

*None identified at planning stage.*