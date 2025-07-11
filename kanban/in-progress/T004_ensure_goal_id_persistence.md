---
title: "Ensure Goal ID Persistence in goals.yml"
type: ["bug", "feature"]
tags: ["parser", "goals", "data-integrity"]
related_tasks: []
context_windows: ["./CLAUDE.md", "./doc/specifications/goal_schema.md", "./internal/models/*.go", "./internal/parser/*.go"]
---

# Ensure Goal ID Persistence in goals.yml

## 1. Goal / User Story

As a user, I want goal IDs to be automatically added to my goals.yml file when missing, so that changing goal titles doesn't break the connection between my historical entries and current goals.

Currently, if a user doesn't specify an `id` field in goals.yml, the system generates one internally but doesn't persist it back to the file. This means:

1. If the user changes a goal's title, a new ID gets generated
2. Historical entries in entries.yml become orphaned (no longer match any goal)
3. Data integrity is compromised

The system should automatically write inferred IDs back to goals.yml after successful parsing to ensure data consistency.

## 2. Acceptance Criteria

- [ ] When goals.yml is parsed and a goal lacks an `id` field, the inferred ID is written back to the file
- [ ] The original file structure and formatting is preserved as much as possible
- [ ] Only missing IDs are added - existing IDs are never modified
- [ ] The operation is atomic (no partial writes that could corrupt the file)
- [ ] Proper error handling if the file cannot be written
- [ ] No changes made if goals.yml is read-only or parsing fails
- [ ] Backwards compatibility maintained (existing workflows unaffected)

---
## 3. Implementation Plan & Progress

**Overall Status:** `Planning`

**Investigation completed:**
- [x] **Current ID generation logic** - IDs generated in `Goal.Validate()` method (`internal/models/goal.go:129`) using `generateIDFromTitle()` if missing
- [x] **Parser architecture** - Goals loaded via `GoalParser.LoadFromFile()` which calls `ParseYAML()` then `schema.Validate()` where ID generation happens
- [x] **File writing approach** - `GoalParser.SaveToFile()` uses `yaml.MarshalWithOptions()` with indent formatting, completely rewrites file
- [x] **Error handling strategy** - Current pattern returns errors, atomic writes via `os.WriteFile()`
- [x] **Integration points** - Primary entry point is `EntryCollector.CollectTodayEntries()` in `entry` command

**Implementation approach:**
- [ ] **1. Add ID persistence check** - After successful parsing/validation, check if any goals had missing IDs
- [ ] **2. Conditional file update** - If IDs were added during validation, save schema back to file
- [ ] **3. Integration in LoadFromFile** - Extend `LoadFromFile()` to optionally write back inferred IDs
- [ ] **4. Error handling** - Handle read-only files, disk full, permission errors gracefully
- [ ] **5. Testing** - Unit tests for ID persistence, integration tests for file updates

## 4. Roadblocks

*(None identified yet)*

## 5. Notes / Discussion Log

- Created based on data integrity concern - title changes breaking entry associations
- Investigation complete - implementation approach identified

**Key findings:**
- ID generation happens during `schema.Validate()` call in parser
- Current `SaveToFile()` completely rewrites file with pretty formatting 
- Need to track which goals had missing IDs to trigger conditional save
- Primary integration point is `entry` command via `EntryCollector.CollectTodayEntries()`

**Implementation strategy:**
1. Modify `Goal.Validate()` to track when IDs are generated
2. Add method to check if schema was modified during validation
3. Extend `LoadFromFile()` to conditionally save back modified schema
4. Handle file permission and error cases gracefully