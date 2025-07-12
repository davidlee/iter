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
- [ ] 2.1: Implement `iter list add $id` command
- [ ] 2.2: Implement `iter list edit $id` command  
- [ ] 2.3: Implement `iter list entry` (menu selection)
- [ ] 2.4: Implement `iter list entry $id` (direct access)

### Phase 3: Goal Integration (Priority: High)
- [ ] 3.1: Add ChecklistGoal support to goal configuration UI
- [ ] 3.2: Implement automatic scoring for checklist completion
- [ ] 3.3: Implement manual scoring support
- [ ] 3.4: Add checklist criteria validation

### Phase 4: Enhanced UI & Experience (Priority: Medium)
- [ ] 4.1: Add progress indicators to checklist headings (e.g., "clean station (3/5)")
- [ ] 4.2: Implement checklist state persistence across sessions
- [ ] 4.3: Add entry recording integration for checklist goals
- [ ] 4.4: Add checklist completion summary and statistics

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
iter list add morning_routine          # Create new checklist
iter list edit morning_routine         # Edit existing checklist
iter list rm morning_routine           # Remove checklist

# Checklist completion
iter list entry                        # Show menu of available checklists
iter list entry morning_routine        # Complete specific checklist
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

## Roadblocks

*None identified at planning stage.*