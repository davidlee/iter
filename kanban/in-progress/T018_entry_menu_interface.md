---
title: "Entry Menu Interface"
type: ["feature"]
tags: ["ui", "entry", "menu"]
related_tasks: ["related-to:T015", "depends-on:T010", "related-to:T017"]
context_windows: ["cmd/**/*.go", "internal/ui/**/*.go", "internal/entry/**/*.go", "CLAUDE.md", "doc/**/*.md"]
---

# Entry Menu Interface

**Context (Background)**:
Entry menu interface provides streamlined daily habit tracking by combining goal browsing with direct entry capabilities. Current system requires separate `vice goal list` and `vice entry` commands. This task creates unified interface showing goals with visual status indicators and allowing immediate entry without command switching.

**Context (Significant Code Files)**:
- `cmd/entry.go` - Current entry command structure
- `cmd/root.go` - Root command configuration for default behavior  
- `internal/ui/entry.go` - EntryCollector orchestrates entry workflow
- `internal/ui/goalconfig/goal_list.go` - GoalListModel provides goal browsing patterns
- `internal/ui/entry/goal_collection_flows.go` - Goal-specific entry collection flows
- `internal/ui/entry/` - Field input components and entry UI patterns

## Git Commit History

**All commits related to this task (newest first):**

- `b1d409e` - feat(kanban)[T018]: add entry menu interface task

## 1. Goal / User Story

As a user, I want a streamlined entry interface that shows all my goals with their current status and allows me to enter values one by one, so that I can efficiently complete my daily habit tracking without navigating multiple commands.

This interface should become the primary entry point for the application, making daily use frictionless and visually informative.

## 2. Acceptance Criteria

- [ ] `vice entry --menu` flag launches interactive goal selection interface
- [ ] Interface displays goals similar to `vice goal list` but optimized for entry
- [ ] Goals show color-coded status indicators (success/failed/skipped/incomplete)
- [ ] Completion progress bar shows overall daily progress
- [ ] Selecting a goal launches entry interface for that specific goal
- [ ] entries.yml is written after each goal entry
- [ ] Next incomplete goal is auto-selected upon return to menu
- [ ] Interface provides clear navigation controls (exit, skip, etc.)
- [ ] `vice` (no arguments) defaults to entry menu interface
- [ ] All existing entry functionality remains available through existing commands

## 3. Architecture

### Design Strategy: Loose Coupling Through Existing Interfaces

**Interface Reuse Pattern**: Entry menu will leverage existing abstractions (`EntryCollector`, `GoalCollectionFlow`) rather than direct goal type coupling, supporting T017's loose coupling goals.

**Color System Integration**: Will use lipgloss with termenv for cross-platform color support:
- Success: `lipgloss.Color("214")` (gold)
- Failed: `lipgloss.Color("88")` (dark red)  
- Skipped: `lipgloss.Color("240")` (dark grey)
- Incomplete: `lipgloss.Color("250")` (light grey)

**Component Architecture**:
```
EntryMenuModel (BubbleTea UI)
├── EntryCollector (existing orchestration)
├── GoalCollectionFlow (existing entry flows)
├── GoalListModel patterns (navigation/display)
└── MenuState (progress tracking, filters)
```

**Key Architectural Decisions**:
- **Abstraction Layer**: Menu interacts with `EntryCollector` interface, not specific goal types
- **State Management**: Separate menu state from entry collection state for clean separation
- **Navigation Flow**: Bidirectional flow between menu and entry with configurable return behavior
- **Extensibility**: Design supports T017's planned multi-field goals and validation flexibility

## 4. Implementation Plan & Progress

**Overall Status:** `In Progress`

**Sub-tasks:**

- [x] **Analysis & Design Phase**
  - [x] **Research existing UI patterns:** Analyze current goal list and entry implementations
    - *Design:* GoalListModel uses bubbles/list, EntryCollector orchestrates flows, BubbleTea Model-View-Update
    - *Code/Artifacts:* Understanding of GoalListKeyMap, EntryCollectionFlow interface, styling patterns
    - *Testing Strategy:* N/A - research phase
    - *AI Notes:* Strong foundation with GoalListModel patterns and EntryCollector abstraction
  
- [ ] **Core Menu Interface Implementation**
  - [ ] **Create EntryMenuModel structure:** Build BubbleTea model for menu interface
    - *Design:* Adapt GoalListModel patterns with entry-specific state (progress, filters, return behavior)
    - *Code/Artifacts:* `internal/ui/entrymenu/model.go` with BubbleTea implementation
    - *Testing Strategy:* Unit tests for model state transitions, headless constructor for testing
    - *AI Notes:* Follow established patterns from GoalListModel, maintain loose coupling to goal types
  
  - [ ] **Implement goal status rendering:** Add color-coded goal status display with progress bar
    - *Design:* Status colors via lipgloss.Color(), progress calculation across all goals
    - *Code/Artifacts:* Status rendering logic in `internal/ui/entrymenu/view.go`
    - *Testing Strategy:* Visual rendering tests, color mapping validation, progress calculation tests
    - *AI Notes:* Use termenv-compatible colors: gold(214), dark red(88), dark grey(240), light grey(250)
  
  - [ ] **Add menu navigation and filtering:** Implement keybindings for menu operations
    - *Design:* Extend GoalListKeyMap patterns: r(return behavior), s(skip filter), p(previous filter)
    - *Code/Artifacts:* Keybinding definitions and filter state management
    - *Testing Strategy:* Navigation behavior tests, filter state persistence tests
    - *AI Notes:* Follow established keybinding patterns, maintain help text generation consistency

- [ ] **Entry Integration Layer**
  - [ ] **Create menu-entry integration:** Connect menu to existing EntryCollector system
    - *Design:* Launch EntryCollector.CollectSingleEntry() method, handle return flow
    - *Code/Artifacts:* Integration methods in EntryMenuModel, single-goal entry collection
    - *Testing Strategy:* End-to-end tests for menu→entry→menu flow across all goal types
    - *AI Notes:* Leverage existing EntryCollector abstraction, avoid direct goal type coupling
  
  - [ ] **Implement auto-save and state management:** Handle entries.yml updates and navigation
    - *Design:* Reuse EntryCollector storage mechanisms, smart next-goal selection algorithm
    - *Code/Artifacts:* State persistence logic, next-goal selection in menu model
    - *Testing Strategy:* File I/O tests, state transition tests, error handling validation
    - *AI Notes:* Build on existing storage abstraction, handle write failures gracefully

- [ ] **Command Integration & Default Behavior**
  - [ ] **Add --menu flag to entry command:** Implement command-line interface
    - *Design:* Extend cmd/entry.go with menu flag, conditional execution path
    - *Code/Artifacts:* Modified `cmd/entry.go` with menu mode selection
    - *Testing Strategy:* CLI flag parsing tests, backward compatibility validation
    - *AI Notes:* Maintain existing entry command behavior, clean flag integration
  
  - [ ] **Configure default command behavior:** Make `vice` alone launch entry menu
    - *Design:* Modify root command to detect no arguments and launch menu mode
    - *Code/Artifacts:* Updated `cmd/root.go` with default command routing
    - *Testing Strategy:* Default behavior tests, ensure other commands remain accessible
    - *AI Notes:* Preserve access to help and other commands, clear user experience

- [ ] **Enhancement Features**
  - [ ] **Implement configurable return behavior:** Add "r" key toggle for menu vs next-goal return
    - *Design:* Menu state tracking return preference, toggle keybinding
    - *Code/Artifacts:* Return behavior state management and toggle logic
    - *Testing Strategy:* Behavior toggle tests, state persistence across menu sessions
    - *AI Notes:* User preference should persist during menu session, clear visual indication
  
  - [ ] **Add goal filtering capabilities:** Implement "s" and "p" keys for status filtering
    - *Design:* Filter state management, goal list filtering based on entry status
    - *Code/Artifacts:* Filter logic in menu model, visual filter indicators
    - *Testing Strategy:* Filter behavior tests, filter state transitions
    - *AI Notes:* Filters should be clearly indicated in UI, easy to toggle on/off

## 5. Roadblocks

*(No roadblocks identified yet)*

## 6. Notes / Discussion Log

- `2025-07-14 - User:` Initial task request for entry menu interface with goal selection, status colors, progress bar, and default command behavior.
- `2025-07-14 - User:` Clarifications provided:
  - `vice` (no arguments) launches entry menu
  - Colors: success=gold, failure=dark red, skipped=dark grey, incomplete=light grey  
  - Menu shows only today's/current period goals
  - Reuse existing entry interface when goal selected
  - Navigation: [ESC] or "q" to exit (except in text entry fields)
  - Future enhancement: support editing entries for days other than today
  - Enhancement: "r" key to toggle return behavior (menu vs next goal after entry)
  - Enhancement: "s" key to filter out skipped goals, "p" key to filter out previously entered goals (successful/failed)
- `2025-07-14 - AI:` Analysis phase completed. Key findings:
  - Strong foundation with GoalListModel patterns and EntryCollector abstraction
  - Existing system supports loose coupling via EntryCollector interface
  - T017 architectural insights guide interface-based design for extensibility
  - Color system uses lipgloss with termenv: gold(214), dark red(88), dark grey(240), light grey(250)
  - Implementation plan updated with detailed sub-tasks based on existing patterns