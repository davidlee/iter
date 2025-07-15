---
title: "Entry Menu Bug Fixes"
tags: ["ui", "bug", "entry", "menu"]
related_tasks: ["related-to:T018"]
context_windows: ["internal/ui/entrymenu/**/*.go", "internal/ui/entry.go", "internal/ui/entry/**/*.go", "cmd/entry.go", "CLAUDE.md", "doc/**/*.md"]
---

# Entry Menu Bug Fixes

**Context (Background)**:
Two bugs identified in the entry menu interface (T018) that affect user experience:

1. **Incorrect task completion status display**: Entry menu shows wrong completion status for tasks, with discrepancy between what's shown in the UI and what's actually in entries.yml file
2. **Edit looping bug**: When editing a task in the entry menu, the interface loops through single task's edit screens instead of returning to the menu

**Type**: `fix`

**Overall Status:** `Not Started`

## Reference (Relevant Files / URLs)

### Significant Code (Files / Functions)

**Entry Menu Core Files**:
- `internal/ui/entrymenu/model.go:266-280` - createMenuItems() function that converts entries to menu items
- `internal/ui/entrymenu/model.go:469-514` - updateEntriesFromCollector() syncs collector state to menu  
- `internal/ui/entrymenu/model.go:304-342` - Entry selection and collection logic (lines 308-340)
- `internal/ui/entrymenu/view.go:158-180` - calculateProgressStats() determines completion status
- `internal/ui/entry.go:314-330` - CollectSingleGoalEntry() and GetGoalEntry() methods

**Entry Collection Flow Files**:
- `internal/ui/entry/goal_collection_flows.go:78-177` - CollectEntry() method for all goal types
- `internal/ui/entry/flow_implementations.go:98-99` - form.Run() call that handles user interaction
- `internal/ui/entry/flow_implementations.go:430-484` - collectStandardOptionalNotes() method

**Entry Status Files**:
- `internal/ui/entrymenu/model.go:55-70` - getGoalStatusEmoji() determines status emojis
- `internal/ui/entrymenu/model.go:72-88` - getStatusColor() determines status colors

NOTE: `charmbracelet/` contains reference checkouts of key repos. 
also refer to API docs: https://pkg.go.dev/github.com/charmbracelet/bubbletea (./huh, etc) 

### Relevant Documentation
- `kanban/done/T018_entry_menu_interface.md` - Original entry menu implementation
- `doc/specifications/entries_storage.md` - Entry storage format specification
- `doc/specifications/goal_schema.md` - Goal schema and field type definitions

### Related Tasks / History
- **T018**: Entry menu interface implementation (recently completed)
- **T010**: Entry system implementation with goal collection flows
- **T012**: Habit skip functionality with status tracking

## Goal / User Story

**Bug 1 - Incorrect completion status**: As a user, when I look at the entry menu, I want to see the correct completion status for my tasks so that I can understand my actual progress without confusion.

**Bug 2 - Edit looping**: As a user, when I edit a task in the entry menu, I want the interface to return to the menu after I complete the edit so that I can continue with other tasks efficiently.

## Acceptance Criteria (ACs)

### Bug 1 - Incorrect completion status
- [ ] Entry menu correctly displays task completion status matching entries.yml data
- [ ] Progress bar shows accurate completion statistics
- [ ] Status emojis (✓ ✗ ~ ☐) correctly reflect actual entry status
- [ ] Status colors match the actual entry status (gold/red/grey/light grey)

### Bug 2 - Edit looping  
- [ ] When editing a task, user is returned to entry menu after completion
- [ ] No infinite loops or repeated edit screens for single task
- [ ] Entry menu state properly updates after edit completion
- [ ] Auto-save functionality works correctly after edit

## Architecture

### Bug Analysis

**Bug 1 - Status Display Issue**:
Based on code analysis, the bug appears to be in the status synchronization between:
1. `entries.yml` file data (source of truth)
2. `EntryCollector` internal state (interface{} values)
3. `EntryMenuModel` display state (GoalEntry structs)

**Root Cause**: The `updateEntriesFromCollector()` method (lines 469-514) may have type conversion issues or timing problems when syncing collector state to menu display.

**Bug 2 - Edit Looping Issue**:
The looping occurs in the goal collection flow where:
1. User presses Enter → `CollectSingleGoalEntry()` called
2. `flow.CollectEntry()` launches input form via `form.Run()`
3. Instead of returning to menu, the flow continues or loops

**Root Cause**: The `form.Run()` call in goal collection flows may not properly handle completion state, or there's an issue with the return flow logic.

### Design Strategy

**Approach 1: Minimal Fix (Recommended for immediate resolution)**
- Follow T018 patterns using EntryCollector abstraction
- Fix synchronization issues in existing architecture
- Ensure proper state management between file storage, EntryCollector, and EntryMenuModel

**Approach 2: Modal/Viewport Enhancement (Architectural improvement)**
- Implement modal system for entry editing forms
- Entry menu remains active in background during edits
- Edit forms appear as overlays/modals that naturally close → return to menu
- Eliminates complex state handoff and return logic entirely

**Modal Benefits**:
- **Simplifies Bug 2**: Natural modal close cycle eliminates looping
- **Better UX**: User maintains context of menu while editing
- **Cleaner Architecture**: `Menu + Modal(Edit) → close → Menu` vs `Menu → handoff → Edit → return → Menu`
- **Eliminates handoff complexity**: No need for complex state transfer

**Modal Complexity**:
- Requires implementing modal/viewport system in BubbleTea
- Significant refactoring of current `form.Run()` approach
- More complex rendering logic (menu + modal overlay)
- Additional testing for modal interactions

**Recommendation**: Implement Approach 2 (Modal/Viewport) as architectural improvement, then verify bug resolution.

### State Management

**Current synchronization requirements**:
- File storage (entries.yml)
- EntryCollector state (interface{} values)  
- EntryMenuModel display (GoalEntry structs)

## Implementation Plan & Progress

**Sub-tasks:**

- [ ] **1. Modal/Viewport Architecture Design**
  - [ ] **1.1 Research BubbleTea modal patterns:** Study existing modal implementations
    - *Design:* Research viewport, modal overlay patterns in BubbleTea ecosystem
    - *Code/Artifacts:* Architecture document with modal system design
    - *Testing Strategy:* Proof-of-concept modal implementation
    - *AI Notes:* Look at bubbletea examples, viewport component, layered rendering
  - [ ] **1.2 Design modal system architecture:** Define modal interface and state management
    - *Design:* Modal interface, overlay rendering, focus management, keyboard navigation
    - *Code/Artifacts:* Modal system design in `internal/ui/modal/` package
    - *Testing Strategy:* Unit tests for modal state management
    - *AI Notes:* Consider how modals integrate with existing BubbleTea Model-View-Update pattern

- [ ] **2. Modal System Implementation**
  - [ ] **2.1 Implement core modal system:** Create modal base infrastructure
    - *Design:* Modal manager, overlay rendering, focus handling, keyboard routing
    - *Code/Artifacts:* `internal/ui/modal/` package with base modal system
    - *Testing Strategy:* Unit tests for modal lifecycle, integration tests for overlay
    - *AI Notes:* Focus on clean separation between modal and parent model state
  - [ ] **2.2 Create modal entry form component:** Replace form.Run() with modal approach
    - *Design:* Modal wrapper for entry forms, proper cleanup, state isolation
    - *Code/Artifacts:* Modal entry form component that works with existing flows
    - *Testing Strategy:* teatest integration tests for modal form interactions
    - *AI Notes:* Ensure modal properly isolates entry form state from menu state

- [ ] **3. Entry Menu Integration**
  - [ ] **3.1 Modify entry menu for modal integration:** Update menu to support modal overlays
    - *Design:* Menu model handles modal events, rendering with overlay support
    - *Code/Artifacts:* Modified `internal/ui/entrymenu/model.go` with modal integration
    - *Testing Strategy:* Integration tests for menu + modal rendering
    - *AI Notes:* Menu stays active, modal appears as overlay, proper event routing
  - [ ] **3.2 Refactor goal collection flows:** Update flows to use modal instead of takeover
    - *Design:* Remove form.Run() calls, use modal form component instead
    - *Code/Artifacts:* Modified `internal/ui/entry/goal_collection_flows.go`
    - *Testing Strategy:* Flow tests with modal form component
    - *AI Notes:* This should eliminate the handoff complexity causing Bug 2

- [ ] **4. Testing & Validation**
  - [ ] **4.1 Integration testing:** Verify modal-based entry workflow
    - *Design:* End-to-end tests for menu → modal edit → close → menu flow
    - *Code/Artifacts:* Comprehensive test suite for modal interactions
    - *Testing Strategy:* teatest integration tests, keyboard navigation tests
    - *AI Notes:* Focus on smooth modal transitions, proper cleanup, state isolation
  - [ ] **4.2 Bug verification:** Confirm original bugs are resolved
    - *Design:* Test scenarios that reproduce original bugs, verify they're fixed
    - *Code/Artifacts:* Regression tests for original bug scenarios
    - *Testing Strategy:* Compare with original bug reports, verify user experience
    - *AI Notes:* Modal architecture should naturally solve Bug 2, verify Bug 1 status sync

## Roadblocks

*(No roadblocks identified yet)*

## Notes / Discussion Log

- `2025-07-15 - AI:` Initial bug report analysis completed
  - **Bug 1 Context**: User reports entry menu shows incorrect task completion status
  - **Bug 2 Context**: User reports editing task loops through single task's edit screens
  - **Analysis Method**: Examined entries.yml file, entry menu code, and goal collection flows
  - **Key Finding**: Status synchronization issues likely in updateEntriesFromCollector()
  - **Key Finding**: Edit looping likely in form.Run() completion handling
  - **Next Steps**: Reproduce bugs with test cases, then implement fixes

## Git Commit History

*No commits yet - task is in backlog*