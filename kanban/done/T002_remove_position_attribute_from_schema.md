---
title: "Remove position attribute from habits schema - infer from file order"
type: ["refactor"]
tags: ["schema", "parser"]
related_tasks: []
context_windows: ["doc/specifications/*.md", "CLAUDE.md", "brief.md"]
---

# Remove position attribute from habits schema - infer from file order

## Git Commit History

**All commits related to this task (newest first):**

- `06e2e86` - fix: complete T002 position field removal implementation
- `d9810fa` - refactor: [T002] (complete) - remove position attribute from habit schema

## 1. Habit / User Story

As a developer maintaining the habit schema specification, I want to remove the explicit `position` attribute from individual habits and instead infer the display order from the sequence of habits in the YAML file. This simplifies the schema by eliminating redundant positional data that can be automatically determined.

## 2. Acceptance Criteria

- [x] Remove `position` field from Habit object specification
- [x] Update schema validation rules to not require position uniqueness
- [x] Update example schema to remove position attributes
- [x] Document that habit order is determined by sequence in YAML file
- [x] Ensure historical data compatibility is maintained (position was never stored in entries)

---
## 3. Implementation Plan & Progress

**Overall Status:** `Completed`

**Sub-tasks:**

- [x] **Update Schema Specification**: Remove position attribute from formal spec
    - [x] **Remove position from Habit object definition**
        - *Design:* Remove `position: int` field from Habit YAML structure
        - *Code/Artifacts to be created or modified:* Habit schema specification document
        - *Testing Strategy:* Verify example schemas parse correctly without position
        - *AI Notes:* Completed - removed position field from Habit object specification
    - [x] **Update validation requirements**
        - *Design:* Remove position uniqueness constraint from schema validation rules
        - *Code/Artifacts to be created or modified:* Schema validation specification
        - *Testing Strategy:* Review validation checklist for position-related items
        - *AI Notes:* Completed - updated validation rules to remove position uniqueness requirement
    - [x] **Update example schema**
        - *Design:* Remove position attributes from example YAML in specification
        - *Code/Artifacts to be created or modified:* Example schema in specification document
        - *Testing Strategy:* Verify example remains valid and clear
        - *AI Notes:* Completed - removed position attributes from all example habits
    - [x] **Add habit ordering documentation**
        - *Design:* Document how habit order is determined from YAML sequence
        - *Code/Artifacts to be created or modified:* Habit schema specification document
        - *Testing Strategy:* Verify documentation is clear and comprehensive
        - *AI Notes:* Added new "Habit Ordering" section explaining position inference

## 4. Roadblocks

*(No roadblocks identified yet)*

## 5. Notes / Discussion Log

- `2025-07-11 - User:` Requested removal of position attribute since it can be inferred from file order
- `2025-07-11 - AI:` This change simplifies schema maintenance and reduces potential for position conflicts
- `2025-07-11 - AI:` Completed all sub-tasks:
  - Removed position field from Habit object specification
  - Updated validation requirements to remove position uniqueness constraint
  - Updated example schema to remove all position attributes
  - Added new "Habit Ordering" section documenting how position is inferred from YAML array sequence

## 6. Code Snippets & Artifacts 

**Commit:** `d9810fa` - refactor: [T002] (complete) - remove position attribute from habit schema

### Changes Made:

1. **Habit Object Specification** (`doc/specifications/habit_schema.md`):
   - Removed `position: 1 # Unique integer for display order` from Habit object structure
   - Added new "Habit Ordering" section explaining position inference from YAML array sequence

2. **Validation Requirements**:
   - Updated from "All habit IDs and positions must be unique" to "All habit IDs must be unique"
   - Updated Change Resilience section to reference sequence-based ordering

3. **Example Schema**:
   - Removed position attributes from all three example habits (Daily Exercise, Morning Meditation, Sleep Quality)
   - Habits maintain logical ordering through their sequence in the YAML file

The schema is now simpler and more maintainable while preserving all functionality.