---
title: "Entry Menu Interface"
type: ["feature"]
tags: ["ui", "entry", "menu"]
related_tasks: ["related-to:T015", "depends-on:T010"]
context_windows: ["cmd/**/*.go", "internal/ui/**/*.go", "internal/entry/**/*.go", "CLAUDE.md", "doc/**/*.md"]
---

# Entry Menu Interface

**Context (Background)**:
*AI to complete*

**Context (Significant Code Files)**:
*AI to complete*

## Git Commit History

**All commits related to this task (newest first):**

*No commits yet - task is in backlog*

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

*AI to complete when changes are architecturally significant, or when asked, prior to implementation plan.*

## 4. Implementation Plan & Progress

**Overall Status:** `Not Started`

**Sub-tasks:**

- [ ] **Analysis & Design Phase**
  - [ ] **Research existing UI patterns:** Analyze current goal list and entry implementations
    - *Design:* Review `internal/ui/goallist/` and entry command structures
    - *Code/Artifacts to be created or modified:* Understanding of existing patterns
    - *Testing Strategy:* N/A - research phase
    - *AI Notes:* Need to understand color schemes, navigation patterns, data flow
  
- [ ] **Core Menu Interface Implementation**
  - [ ] **Create entry menu UI component:** Build main interface for goal selection
    - *Design:* Interactive TUI showing goals with status colors and progress bar
    - *Code/Artifacts to be created or modified:* `internal/ui/entrymenu/` package
    - *Testing Strategy:* Unit tests for UI state management, headless integration tests
    - *AI Notes:* Consider reusing patterns from goallist UI
  
  - [ ] **Implement status indicators:** Add color-coded goal status display
    - *Design:* Visual indicators for success/failed/skipped/incomplete states
    - *Code/Artifacts to be created or modified:* Status rendering logic in menu component
    - *Testing Strategy:* Test status color mapping, accessibility considerations
    - *AI Notes:* Ensure colors work in different terminal environments
  
  - [ ] **Add completion progress bar:** Show overall daily progress
    - *Design:* Calculate and display percentage of goals completed/attempted
    - *Code/Artifacts to be created or modified:* Progress calculation logic
    - *Testing Strategy:* Test progress calculation with various goal states
    - *AI Notes:* Consider different progress calculation strategies

- [ ] **Entry Integration**
  - [ ] **Goal selection and entry flow:** Connect menu to existing entry system
    - *Design:* Launch appropriate entry interface based on goal type/configuration
    - *Code/Artifacts to be created or modified:* Integration with existing entry commands
    - *Testing Strategy:* End-to-end tests for complete entry workflow
    - *AI Notes:* Ensure compatibility with all goal types and field types
  
  - [ ] **Auto-save and state management:** Handle entries.yml updates and navigation
    - *Design:* Automatic persistence after each entry, smart next-goal selection
    - *Code/Artifacts to be created or modified:* State management and file I/O logic
    - *Testing Strategy:* Test file writing, error handling, state transitions
    - *AI Notes:* Consider graceful handling of write failures

- [ ] **Command Integration**
  - [ ] **Add --menu flag to entry command:** Implement command-line interface
    - *Design:* Extend existing entry command with new flag option
    - *Code/Artifacts to be created or modified:* `cmd/entry.go` command structure
    - *Testing Strategy:* CLI flag parsing tests, integration with existing entry commands
    - *AI Notes:* Maintain backwards compatibility with existing entry usage
  
  - [ ] **Make entry menu default command:** Configure `vice` to launch menu by default
    - *Design:* Modify root command behavior to default to entry menu
    - *Code/Artifacts to be created or modified:* `cmd/root.go` default command logic
    - *Testing Strategy:* Test default behavior, ensure other commands still accessible
    - *AI Notes:* Consider providing clear way to access help and other commands

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