---
title: "Ensure Habit ID Persistence in habits.yml"
type: ["bug", "feature"]
tags: ["parser", "habits", "data-integrity"]
related_tasks: []
context_windows: ["./CLAUDE.md", "./doc/specifications/goal_schema.md", "./internal/models/*.go", "./internal/parser/*.go"]
---

# Ensure Habit ID Persistence in habits.yml

## Git Commit History

**All commits related to this task (newest first):**

- `54f5824` - feat: [T004] implement habit ID persistence in habits.yml

## 1. Habit / User Story

As a user, I want habit IDs to be automatically added to my habits.yml file when missing, so that changing habit titles doesn't break the connection between my historical entries and current habits.

Currently, if a user doesn't specify an `id` field in habits.yml, the system generates one internally but doesn't persist it back to the file. This means:

1. If the user changes a habit's title, a new ID gets generated
2. Historical entries in entries.yml become orphaned (no longer match any habit)
3. Data integrity is compromised

The system should automatically write inferred IDs back to habits.yml after successful parsing to ensure data consistency.

## 2. Acceptance Criteria

- [x] When habits.yml is parsed and a habit lacks an `id` field, the inferred ID is written back to the file
- [x] The original file structure and formatting is preserved as much as possible
- [x] Only missing IDs are added - existing IDs are never modified
- [x] The operation is atomic (no partial writes that could corrupt the file)
- [x] Proper error handling if the file cannot be written
- [x] No changes made if habits.yml is read-only or parsing fails
- [x] Backwards compatibility maintained (existing workflows unaffected)

---
## 3. Implementation Plan & Progress

**Overall Status:** `COMPLETED` ✅

**Investigation completed:**
- [x] **Current ID generation logic** - IDs generated in `Habit.Validate()` method (`internal/models/habit.go:129`) using `generateIDFromTitle()` if missing
- [x] **Parser architecture** - Habits loaded via `GoalParser.LoadFromFile()` which calls `ParseYAML()` then `schema.Validate()` where ID generation happens
- [x] **File writing approach** - `GoalParser.SaveToFile()` uses `yaml.MarshalWithOptions()` with indent formatting, completely rewrites file
- [x] **Error handling strategy** - Current pattern returns errors, atomic writes via `os.WriteFile()`
- [x] **Integration points** - Primary entry point is `EntryCollector.CollectTodayEntries()` in `entry` command

**Implementation completed:**
- [x] **1. Add ID persistence check** - Added `ValidateAndTrackChanges()` methods to `Habit` and `Schema` that track when IDs are generated
- [x] **2. Conditional file update** - `LoadFromFileWithIDPersistence()` saves schema back only if IDs were generated during parsing
- [x] **3. Integration in LoadFromFile** - Default `LoadFromFile()` now enables ID persistence automatically
- [x] **4. Error handling** - Read-only files, permission errors handled gracefully with warnings (don't break normal operation)
- [x] **5. Testing** - Comprehensive unit tests added covering all scenarios: missing IDs, existing IDs, read-only files, mixed scenarios

## 4. Roadblocks

*(None identified yet)*

## 5. Notes / Discussion Log

- Created based on data integrity concern - title changes breaking entry associations
- Investigation complete - implementation approach identified

**Key findings:**
- ID generation happens during `schema.Validate()` call in parser
- Current `SaveToFile()` completely rewrites file with pretty formatting 
- Need to track which habits had missing IDs to trigger conditional save
- Primary integration point is `entry` command via `EntryCollector.CollectTodayEntries()`

**Implementation strategy:**
1. Modify `Habit.Validate()` to track when IDs are generated
2. Add method to check if schema was modified during validation
3. Extend `LoadFromFile()` to conditionally save back modified schema
4. Handle file permission and error cases gracefully

---

## ✅ IMPLEMENTATION COMPLETED

**Files modified:**
- `internal/models/habit.go` - Added `ValidateAndTrackChanges()` methods for `Habit` and `Schema`
- `internal/parser/habits.go` - Added `LoadFromFileWithIDPersistence()` and `ParseYAMLWithChangeTracking()`
- `internal/parser/id_persistence_test.go` - Comprehensive test suite (18 test cases)

**Key features implemented:**
1. **Automatic ID generation tracking** - `ValidateAndTrackChanges()` methods detect when IDs are generated during validation
2. **Conditional file persistence** - Only saves back to file when IDs were actually generated
3. **Graceful error handling** - Read-only files or permission errors log warnings but don't break operation
4. **Backwards compatibility** - Default `LoadFromFile()` enables persistence, but `LoadFromFileWithIDPersistence(false)` available for opt-out
5. **Atomic operations** - Uses existing `SaveToFile()` which performs atomic writes

**Data integrity benefits:**
- Habit titles can now be changed without breaking entry associations
- Generated IDs are immediately persisted for consistency
- Historical entries remain connected to habits via stable IDs
- Manual ID specification still supported and preserved

**Testing coverage:**
- ID generation and persistence for missing IDs
- Preservation of existing IDs
- Read-only file handling
- Mixed scenarios (some habits with IDs, some without)
- Persistence enable/disable functionality
- Error cases and validation failures