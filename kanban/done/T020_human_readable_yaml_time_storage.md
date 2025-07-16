# T020: Human-Readable YAML Time Storage

**Date**: 2025-07-14  
**Status**: Complete  
**Related tasks**: ["related:T019"]

## 1. Habit

Convert YAML time storage to human-readable formats while maintaining permissive parsing for backward compatibility.

## 2. Acceptance Criteria

- [x] Time field values stored as `"08:30"` instead of `"0000-01-01T08:30:00Z"`
- [x] Timestamps stored as `"2025-07-15 09:11:27"` instead of `"2025-07-15T09:11:27.886682863+10:00"`
- [x] Permissive parsing: accept both old and new formats plus common time representations
- [x] Existing entries.yml files continue to work without migration
- [x] New entries written in human-readable format
- [x] No performance degradation for typical usage

## 3. Implementation Plan & Progress

### Current State
```yaml
# Time fields (from time input)
value: 0000-01-01T08:30:00Z

# Timestamps (created_at/updated_at)
created_at: 2025-07-15T09:11:27.886682863+10:00
```

### Target State
```yaml
# Time fields
value: "08:30"

# Timestamps  
created_at: "2025-07-15 09:11:27"
```

### Sub-tasks
- [x] 1.1: Implement custom YAML marshaling for time.Time fields in GoalEntry
- [x] 1.2: Implement permissive time parsing for unmarshaling
- [x] 1.3: Update time field value serialization 
- [x] 1.4: Test backward compatibility with existing data
- [x] 1.5: Test new format generation and parsing

## 4. Roadblocks

None identified.

## 5. Notes / Discussion Log

- Core philosophy: be permissive with parsing, strict with output format
- Should accept: RFC3339, "HH:MM", "YYYY-MM-DD HH:MM:SS", Unix timestamps, etc.
- Must maintain full backward compatibility - no breaking changes
- Focus on entries.yml readability for manual inspection/editing

### Implementation Notes

- **Custom Marshaling Chain**: Required custom MarshalYAML for GoalEntry, DayEntry, and EntryLog to ensure proper nested marshaling
- **YAML Package Consistency**: Updated storage layer from github.com/goccy/go-yaml to gopkg.in/yaml.v3 for consistency with models layer
- **Time Field Detection**: Uses `isTimeFieldValue()` to distinguish time-of-day fields (year 0000) from full timestamps
- **Comprehensive Testing**: Added entry_yaml_test.go with complete coverage of marshaling, unmarshaling, and round-trip scenarios
- **Backward Compatibility**: Verified existing RFC3339 formats continue to parse correctly
- **Storage Integration**: Storage layer successfully uses custom marshaling for human-readable output

### Final Validation

Verified implementation produces expected output:
```yaml
entries:
  - date: "2025-07-15"
    habits:
      - created_at: "2025-07-15 09:11:27"
        goal_id: wake_up
        notes: slept well
        status: completed
        value: "08:30"
version: 1.0.0
```

### Commits

- `f4e0478` - feat(models,storage)[T020]: implement human-readable YAML time storage
- `01655ea` - docs(kanban): update task tracking and workflow documentation  
- `94c143c` - style: run go fmt on codebase

### Completion

All acceptance criteria met. Implementation complete and committed. Ready for production use.