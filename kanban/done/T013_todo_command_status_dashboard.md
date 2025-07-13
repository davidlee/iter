---
title: "Todo Command Status Dashboard"
type: ["feature"]
tags: ["cli", "status", "dashboard"]
related_tasks: ["depends-on:T012"]
context_windows: ["cmd/*.go", "internal/ui/*.go", "internal/storage/*.go", "CLAUDE.md", "internal/models/*.go"]
---

# Todo Command Status Dashboard

**Context (Background)**:
Users need a quick way to see their daily habit status without going through the full entry collection flow. A `todo` command would provide an at-a-glance view of today's habits showing completed, pending, and skipped statuses in a clean tabular format.

**Context (Significant Code Files)**:
- `cmd/`: CLI command structure and parsing
- `internal/ui/`: UI components and formatting utilities  
- `internal/storage/`: Data access for goals and entries
- `internal/models/`: Goal and Entry data structures with EntryStatus enum

## Git Commit History

**All commits related to this task (newest first):**

*No commits yet - task is in backlog*

## 1. Goal / User Story

As a habit tracker user, I want a quick `iter todo` command that shows me today's habit status in a clean table format, so I can see at a glance what I've completed, what's pending, and what I've skipped without entering the full entry collection workflow.

## 2. Acceptance Criteria

- [ ] `iter todo` command displays today's habits in a table format
- [ ] Status indicators: ✓ (completed), ○ (pending), ⤫ (skipped) 
- [ ] Table shows: Goal Name, Status, Value/Notes (if any)
- [ ] Command works when no entries exist for today (shows all pending)
- [ ] Command handles missing goals file gracefully
- [ ] Clean, readable output suitable for terminal display
- [ ] Optional: Color coding for different statuses
- [ ] Optional: Summary line showing completion count (e.g., "3/5 completed, 1 skipped")

## 3. Architecture

*AI to complete when changes are architecturally significant, or when asked, prior to implementation plan.*

The todo command will reuse existing storage and model infrastructure:
- Leverage existing goal loading from `internal/storage/goals.go`
- Use existing entry loading from `internal/storage/entries.go` 
- Utilize EntryStatus enum from T012 for status determination
- Create new UI formatter for tabular display
- Add new subcommand to CLI structure

Key design decisions:
- Read-only operation (no modification of data)
- Status determination logic: completed/failed/skipped from existing entries, pending for missing entries
- Table formatting should be consistent with existing UI patterns
- Consider terminal width limitations for goal names and values

## 4. Implementation Plan & Progress

**Overall Status:** `Completed`

**Sub-tasks:**

- [x] **Phase 1: Core Command Structure**
  - [x] **Sub-task 1.1: Add todo subcommand to CLI**
    - *Design:* Add `todoCmd` to cobra CLI structure in `cmd/`, wire up to main command
    - *Code/Artifacts:* `cmd/todo.go` - new file with cobra command definition
    - *Testing Strategy:* Unit test for command registration, integration test for basic execution
    - *AI Notes:* Follow existing command patterns from `cmd/entry.go` and `cmd/goal.go`

- [x] **Phase 2: Data Loading & Status Logic**
  - [x] **Sub-task 2.1: Load today's goals and entries**
    - *Design:* Function to load all goals, load today's entries, merge status information
    - *Code/Artifacts:* Add functions to existing storage layer or new `internal/ui/todo.go`
    - *Testing Strategy:* Unit tests for data loading with various entry states
    - *AI Notes:* Reuse existing goal/entry loading patterns, handle missing files gracefully
    
  - [x] **Sub-task 2.2: Status determination logic**
    - *Design:* Map goal IDs to entry status, default to pending for missing entries
    - *Code/Artifacts:* Status mapping function, handle all EntryStatus values
    - *Testing Strategy:* Unit tests covering all status combinations
    - *AI Notes:* Leverage T012 EntryStatus enum, consider future status values

- [x] **Phase 3: Table Formatting & Display**
  - [x] **Sub-task 3.1: Create table formatter**
    - *Design:* Use charmbracelet/bubbles table component for rich tabular display with status symbols
    - *Code/Artifacts:* Table formatting using bubbles.table, possibly in `internal/ui/table.go`
    - *Testing Strategy:* Unit tests for table data, visual testing for bubbles table rendering
    - *AI Notes:* Use charmbracelet/bubbles table & progress components per user guidance
    
  - [x] **Sub-task 3.2: Status symbols and summary**
    - *Design:* Unicode symbols for status, optional color coding, completion summary
    - *Code/Artifacts:* Symbol constants, color utilities, summary calculation
    - *Testing Strategy:* Visual verification across different terminals, accessibility considerations
    - *AI Notes:* Ensure symbols render well in various terminal environments

- [x] **Phase 4: Error Handling & Polish**
  - [x] **Sub-task 4.1: Graceful error handling**
    - *Design:* Handle missing files, corrupted data, empty goal sets
    - *Code/Artifacts:* Error handling in command function, user-friendly error messages
    - *Testing Strategy:* Error case testing, integration tests with missing files
    - *AI Notes:* Follow existing error handling patterns from other commands

- [x] **Phase 5: Additional Output Formats**
  - [x] **Sub-task 5.1: Add markdown output flag (-m)**
    - *Design:* Add `-m` flag for markdown todo list format (`- [ ]`, `- [x]`, `- [-]`) with inline notes
    - *Code/Artifacts:* CLI flag parsing, markdown formatter function
    - *Testing Strategy:* Unit tests for markdown output format, integration tests with flag
    - *AI Notes:* Use standard markdown checkbox format, support notes as sub-items
  
  - [x] **Sub-task 5.2: Make table output non-interactive**
    - *Design:* Remove interactive quit requirement, make output pipable to files
    - *Code/Artifacts:* Update bubbles table to static display mode
    - *Testing Strategy:* Test piping output to files, verify non-blocking behavior
    - *AI Notes:* Should work well with shell redirects and automation

## 5. Roadblocks

*(None identified yet)*

## 6. Notes / Discussion Log

- `2025-07-13 - User:` Requested new card for todo command with status dashboard functionality
- `2025-07-13 - AI:` Created task with dependency on T012 (EntryStatus enum). Table format chosen for clean display of multiple habits with status indicators.
- `2025-07-13 - User:` Use charmbracelet/bubbles table & progress components: https://github.com/charmbracelet/bubbles?tab=readme-ov-file#table
- `2025-07-13 - AI:` Initial implementation complete. Features: CLI command integration, data loading with status logic, bubbles table with TTY detection and fallback, comprehensive error handling, complete test coverage. All original acceptance criteria met.
- `2025-07-13 - User:` Add markdown output flag (-m) for plain markdown todo list format. Make table output non-interactive (pipable).
- `2025-07-13 - AI:` Phase 5 implementation complete. Added `-m` flag for markdown output with checkbox format (- [x], - [-], - [ ]). Removed interactive requirement - table now pipable. All tests passing, linter clean.