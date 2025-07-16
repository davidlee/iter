---
title: "Rename Goal to Habit Comprehensive"
type: "refactor"
tags: ["refactor", "naming", "terminology", "comprehensive"]
related_tasks: []
context_windows: ["./**/*.go", "./**/*.md", "./**/*.yml", "./**/*.yaml", "./cmd/*", "./internal/**/*", "./doc/**/*", "./testdata/**/*", "./kanban/**/*"]
---

# Rename Goal to Habit Comprehensive

**Context (Background)**:
The application currently uses "goal" terminology throughout the codebase, documentation, and user interface. This task involves comprehensively renaming all instances of "goal" to "habit" to better reflect the application's purpose as a habit tracker. This includes file names, code structures, documentation, CLI commands, and user-facing strings.

**Type**: `refactor`

**Overall Status:** `In Progress`

## Reference (Relevant Files / URLs)

### Significant Code (Files / Functions)

**Files to be renamed (42 files total):**
- `cmd/goal*.go` → `cmd/habit*.go` (5 files)
- `internal/models/goal.go` → `internal/models/habit.go`
- `internal/parser/goals.go` → `internal/parser/habits.go`
- `internal/ui/goalconfig/` → `internal/ui/habitconfig/` (11 files)
- `internal/ui/entry/*goal*.go` → `internal/ui/entry/*habit*.go` (5 files)
- `internal/integration/*goal*.go` → `internal/integration/*habit*.go` (3 files)
- `testdata/goals/` → `testdata/habits/` (3 files)
- `testdata/user_patterns/*goal*.yml` → `testdata/user_patterns/*habit*.yml` (2 files)
- `doc/specifications/goal_schema.md` → `doc/specifications/habit_schema.md`
- `doc/diagrams/goal_collection_flow.*` → `doc/diagrams/habit_collection_flow.*` (2 files)
- Kanban task files referencing goals (7 files)

**Key Go Types/Structs to rename:**
- `Goal` → `Habit`
- `GoalEntry` → `HabitEntry`
- `GoalConfig` → `HabitConfig`
- `GoalManager` → `HabitManager`
- `GoalCreator` → `HabitCreator`
- `GoalValidator` → `HabitValidator`
- `GoalCollection` → `HabitCollection`

**CLI Commands affected:**
- `vice goal` → `vice habit`
- `vice goal add` → `vice habit add`
- `vice goal list` → `vice habit list`
- `vice goal edit` → `vice habit edit`
- `vice goal remove` → `vice habit remove`

**Configuration files:**
- `goals.yml` → `habits.yml`
- All YAML field names: `goal_id`, `goal_type`, etc. → `habit_id`, `habit_type`, etc.

### Relevant Documentation
- `doc/specifications/goal_schema.md` - Schema documentation
- `doc/architecture.md` - Architecture overview
- `doc/diagrams/goal_collection_flow.d2` - Flow diagrams
- `CLAUDE.md` - Project instructions and standards
- All kanban task files containing goal references

### Related Tasks / History
- T003: Implement elastic goals end to end
- T004: Ensure goal ID persistence
- T005: Goal configuration UI
- T006: Goal management UI
- T015: Goal list interactive UI
- T016: Goal type change scoring error

## Goal / User Story

**As a developer/maintainer**, I want to rename all instances of "goal" to "habit" throughout the codebase so that the terminology consistently reflects the application's purpose as a habit tracker, improving code clarity and user understanding.

**Why this task is important:**
- Improves semantic clarity - "habits" better describes what users are tracking
- Ensures consistent terminology across codebase and documentation
- Enhances user experience with more intuitive command names
- Aligns codebase with the application's actual purpose

## Acceptance Criteria (ACs)

- [ ] All files containing "goal" or "goals" in their names are renamed to use "habit"/"habits"
- [ ] All Go types, structs, and interfaces are renamed from Goal* to Habit*
- [ ] All CLI commands change from `goal` to `habit` (e.g., `vice goal add` → `vice habit add`)
- [ ] All configuration files and YAML fields use "habit" terminology
- [ ] All documentation and comments are updated to use "habit" terminology
- [ ] All user-facing strings and help text use "habit" terminology
- [ ] All import paths broken by file renames are updated
- [ ] All tests pass after the rename
- [ ] Application builds successfully after the rename
- [ ] Users can manually rename their goals.yml to habits.yml (no automatic migration)

## Architecture

This is a comprehensive refactoring that touches:

1. **File System Structure**: Renaming files and directories
2. **Go Language Elements**: Types, interfaces, functions, variables
3. **CLI Interface**: Command names, flags, help text
4. **Configuration Schema**: YAML field names and structure
5. **Documentation**: All technical and user documentation
6. **Test Code**: All test files and test data

The approach will be systematic:
- Phase 1: Rename files and directories (using simple moves/renames)
- Phase 2: Content replacement using find/replace operations  
- Phase 3: Fix any build issues and verify completeness
- Phase 4: One-time verification with shell scripts

## Implementation Plan & Progress

**Sub-tasks:**

- [ ] **Phase 1: File System Reorganization**
  - [ ] **1.1: Rename Go source files**
    - *Design:* Rename all Go files containing "goal" to "habit" (letting git add . handle tracking)
    - *Code/Artifacts:* 42 files renamed across cmd/, internal/, testdata/
    - *Testing Strategy:* Verify renames successful, check for broken imports
    - *AI Notes:* Simple file renames, check for any git issues afterward
  - [ ] **1.2: Rename directories**
    - *Design:* Rename `internal/ui/goalconfig/` → `internal/ui/habitconfig/`, `testdata/goals/` → `testdata/habits/`
    - *Code/Artifacts:* 2 directories renamed
    - *Testing Strategy:* Verify all files moved correctly, check import paths
  - [ ] **1.3: Rename documentation and diagram files**
    - *Design:* Rename all doc/ files containing "goal" to "habit"
    - *Code/Artifacts:* 3 files in doc/ directory
    - *Testing Strategy:* Verify links and references still work

- [ ] **Phase 2: Content Replacement**
  - [ ] **2.1: Protect this task file from replacements**
    - *Design:* Either exclude this file from replacements or git checkout afterward
    - *Code/Artifacts:* This task file (T025_rename_goal_to_habit_comprehensive.md)
    - *Testing Strategy:* Verify task file maintains historical "goal" references
    - *AI Notes:* Task file should preserve original terminology for historical context
  - [ ] **2.2: Replace Go language elements**
    - *Design:* Use sed/ripgrep to replace Goal → Habit, goal → habit, goals → habits in all cases
    - *Code/Artifacts:* ~101 Go files, ~1000+ replacements estimated
    - *Testing Strategy:* Build verification, static analysis
    - *AI Notes:* Handle compound terms like goalEntry → habitEntry, goalConfig → habitConfig
  - [ ] **2.3: Replace CLI commands and user-facing strings**
    - *Design:* Update cobra command definitions, help text, usage examples
    - *Code/Artifacts:* cmd/ files, help strings, usage examples
    - *Testing Strategy:* Manual CLI testing, help text verification
  - [ ] **2.4: Replace configuration and YAML terminology**
    - *Design:* Update YAML field names, schema definitions, sample files
    - *Code/Artifacts:* testdata/ files, schema documentation
    - *Testing Strategy:* YAML parsing tests, schema validation
  - [ ] **2.5: Replace documentation and comments**
    - *Design:* Update all markdown files, code comments, ADRs (excluding this task file)
    - *Code/Artifacts:* ~37 markdown files, inline comments
    - *Testing Strategy:* Documentation review, link verification

- [ ] **Phase 3: Import Path and Build Fixes**
  - [ ] **3.1: Fix import statements**
    - *Design:* Update all import paths affected by file/directory renames
    - *Code/Artifacts:* Go import statements across codebase
    - *Testing Strategy:* Build verification, import cycle detection
  - [ ] **3.2: Fix build issues**
    - *Design:* Resolve any compilation errors from the rename
    - *Code/Artifacts:* Build configuration, module definitions
    - *Testing Strategy:* Clean build from scratch
  - [ ] **3.3: Update kanban task references**
    - *Design:* Update task files that reference old goal terminology
    - *Code/Artifacts:* 7 kanban task files
    - *Testing Strategy:* Task file review, link verification

- [ ] **Phase 4: One-time Verification**
  - [ ] **4.1: Create verification shell scripts**
    - *Design:* Write one-off shell scripts to verify no "goal" references remain in user-facing elements
    - *Code/Artifacts:* Verification scripts for CLI commands, help text, documentation
    - *Testing Strategy:* Script execution to catch any missed references
  - [ ] **4.2: Run comprehensive tests**
    - *Design:* Execute full test suite to ensure functionality intact
    - *Code/Artifacts:* All test files
    - *Testing Strategy:* Unit tests, integration tests, CLI tests
  - [ ] **4.3: Manual spot-check**
    - *Design:* Quick manual verification of key CLI commands and help text
    - *Code/Artifacts:* CLI interface, user workflows
    - *Testing Strategy:* Manual testing of renamed commands

## Roadblocks

*No roadblocks identified yet*

## Notes / Discussion Log

- `2025-01-16 16:00 - User:` Request to rename "goals" to "habits" comprehensively throughout the codebase
- `2025-01-16 16:00 - User:` Confirmed: (1) yes - rename files, (2) yes - handle compound terms, (3) no contexts to preserve "goal", (4) yes to all scope items (goals/goal, all cases, comments/docs, user-facing messages)
- `2025-01-16 16:05 - AI:` Analysis shows 42 files need renaming, ~101 Go files contain "goal", ~37 markdown files need updates  
- `2025-01-16 16:10 - User:` Clarifications: (1) no data migration - users rename files manually, (2) yes change YAML field names, (3) no backwards compatibility, (4) yes change package names, (5) simple renames okay, (6) one-off verification scripts fine

## Git Commit History

**All commits related to this task (newest first):**

*No commits yet - task is in backlog*