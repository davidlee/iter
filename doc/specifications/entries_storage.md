# Entries Storage Specification

## Overview

The entries storage system manages persistent data for daily goal tracking in YAML format. This document specifies the file format, operations, and data integrity patterns.

## File Format

### Location & Structure
- **Default path**: `~/.config/vice/entries.yml`
- **Format**: YAML with 2-space indentation
- **Encoding**: UTF-8
- **Permissions**: 0600 (user read/write only)

### YAML Schema

```yaml
version: "1.0.0"
entries:
  - date: "2025-07-15"          # ISO date format (YYYY-MM-DD)
    goals:
      - goal_id: "wake_up"
        value: "08:30"          # Human-readable time format (HH:MM)
        achievement_level: mini # Optional: none|mini|midi|maxi
        notes: "slept well"     # Optional user notes
        created_at: "2025-07-15 09:11:27"  # Human-readable timestamp
        updated_at: "2025-07-15 09:15:32"  # Optional, only if modified
        status: completed       # completed|failed|skipped
```

### Value Field Types

The `value` field supports multiple data types based on goal field configuration:

| Field Type | Storage Format | Example |
|------------|----------------|---------|
| Boolean | `true`/`false` | `value: true` |
| Text | String | `value: "Completed morning routine"` |
| Time | HH:MM format | `value: "08:30"` |
| Duration | Go duration string | `value: "45m"` |
| Numbers | Numeric | `value: 12000` |
| Checklist | Object reference | `value: {checklist_id: "morning", completed: ["task1"]}` |

### Timestamp Formats

#### Human-Readable Format (Current)
- **created_at**: `"2025-07-15 09:11:27"` (YYYY-MM-DD HH:MM:SS)
- **updated_at**: `"2025-07-15 09:15:32"` (YYYY-MM-DD HH:MM:SS)

#### Legacy Formats (Backward Compatible)
- RFC3339: `"2025-07-15T09:11:27.886682863+10:00"`
- ISO 8601 variants
- Unix timestamps

## Storage Operations

### Core Pattern: Load-Modify-Save

All write operations follow the same pattern:

1. **Load** existing entries.yml (or create empty if missing)
2. **Modify** in-memory data structure
3. **Validate** entire data structure
4. **Save** entire file atomically

### Atomic Write Process

```
1. Validate data structure
2. Marshal to YAML
3. Create parent directories (mode 0750)
4. Write to temporary file (.tmp suffix)
5. Atomically rename .tmp to final file
6. Clean up .tmp on failure
```

### File Operations

| Operation | Method | Description | Load-Modify-Save |
|-----------|--------|-------------|------------------|
| **Read** | `LoadFromFile()` | Load entire file, return empty if missing | Load only |
| **Update Day** | `UpdateDayEntry()` | Most common - update single day's entries | ✅ |
| **Add Day** | `AddDayEntry()` | Create new day entry | ✅ |
| **Update Goal** | `UpdateGoalEntry()` | Update single goal within day | ✅ |
| **Add Goal** | `AddGoalEntry()` | Add goal to existing/new day | ✅ |
| **Backup** | `BackupFile()` | Copy to .backup suffix | Read only |

### Data Validation

**Parse-time validation:**
- Strict YAML parsing (unknown fields rejected)
- Schema validation via `EntryLog.Validate()`
- Date format validation (YYYY-MM-DD)
- Goal ID uniqueness within day

**Save-time validation:**
- Pre-save validation of entire data structure
- Marshal validation ensures no data corruption

## Data Resilience Features

### Current Protections

1. **Atomic Writes**: Temp file + rename prevents partial writes
2. **Directory Creation**: Ensures parent directories exist
3. **Validation**: Strict parsing and pre-save validation
4. **Error Cleanup**: Removes temp files on failure
5. **Graceful Missing Files**: Returns empty structure if file missing

### Backup Strategy

**Current:**
- Manual backup via `BackupFile()` method
- Single `.backup` file (overwrites previous)
- Not automatically called during operations

**Planned Improvements:**
- Automatic backup before modifications
- Configurable backup behavior
- Data integrity verification

## Marshaling Behavior

### Custom YAML Marshaling (T020)

Time fields use custom marshaling for human readability:

```go
// GoalEntry implements MarshalYAML/UnmarshalYAML
func (ge *GoalEntry) MarshalYAML() (interface{}, error) {
    // Formats timestamps as "2025-07-15 09:11:27"  
    // Formats time values as "08:30"
}

func (ge *GoalEntry) UnmarshalYAML(node *yaml.Node) error {
    // Permissive parsing of multiple time formats
    // Backward compatible with RFC3339
}
```

### Permissive Parsing

The system accepts multiple input formats for flexibility:

**Time Values:**
- `"08:30"` (preferred)
- `"0000-01-01T08:30:00Z"` (legacy)
- `"08:30:00"` (with seconds)

**Timestamps:**
- `"2025-07-15 09:11:27"` (preferred)
- `"2025-07-15T09:11:27.886682863+10:00"` (RFC3339)
- Unix timestamps

## Performance Characteristics

### File Size Impact
- **Small files** (< 1MB): Negligible performance impact
- **Large files** (> 10MB): Full rewrite may cause brief delays
- **Current typical size**: ~100KB for year of daily entries

### Operation Frequency
- **Most common**: `UpdateDayEntry()` - once per goal entry session
- **Less common**: Individual goal updates during entry editing
- **Rare**: Bulk operations, date range queries

## Error Handling

### File System Errors
- Missing files: Return empty data structure
- Permission errors: Propagate with context
- Disk space: Propagate write failures
- Corruption: Strict parsing catches malformed YAML

### Data Validation Errors
- Invalid dates: Reject with specific error
- Duplicate goal IDs: Validation failure
- Missing required fields: Schema validation
- Type mismatches: YAML unmarshaling errors

## Configuration

### Planned Configuration Options
- Backup behavior (automatic/manual/disabled)
- Backup retention (single/versioned)
- Validation strictness levels
- File permissions and paths

### Environment Considerations
- Respects `XDG_CONFIG_HOME` for config directory
- Handles filesystem case sensitivity
- Cross-platform path handling

## Migration & Compatibility

### Backward Compatibility
- All legacy timestamp formats supported
- Existing entries.yml files work unchanged
- No breaking changes to file format

### Forward Compatibility
- Version field enables future format changes
- Unknown fields logged but not rejected in permissive mode
- Graceful degradation for unsupported features