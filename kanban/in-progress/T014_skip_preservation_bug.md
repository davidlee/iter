---
title: "Skip Preservation Bug - Entry Validation Error"
type: ["fix"]
tags: ["entry", "validation", "skip"]
related_tasks: ["related-to:T012"]
context_windows: ["cmd/entry/**/*.go", "internal/entry/**/*.go", "internal/goals/**/*.go", "CLAUDE.md", "kanban/CLAUDE.md"]
---

# Skip Preservation Bug - Entry Validation Error

**Context (Background)**:
The validation error occurs in `GoalEntry.Validate()` which enforces that skipped entries cannot have achievement levels (line 186 in `internal/models/entry.go`). However, the entry collection system preserves existing achievement levels when loading entries, creating a conflict when users change status from completed to skipped.

**Context (Significant Code Files)**:
- `internal/models/entry.go:186` - Validation logic that throws the error
- `internal/ui/entry.go:91-113` - Entry loading preserves achievement levels
- `internal/ui/entry/goal_collection_flows.go` - Skip-aware processing flows
- `internal/storage/entries.go:62` - Storage operations trigger validation

## Git Commit History

**All commits related to this task (newest first):**

*No commits yet - task is in backlog*

## 1. Goal / User Story

When a user has previously recorded achievement data for a habit and later decides to skip it, the system should preserve their data and allow the skip operation without throwing validation errors. Currently, `iter entry` fails when trying to skip a habit that was previously recorded with achievement levels.

**Error Message:**
```
📊 Recorded: 0
Error: failed to save entries: failed to update day entry: failed to update day entry: invalid day entry: goal entry at index 0: skipped entries cannot have achievement levels
```

## 2. Acceptance Criteria

- [ ] User can skip a previously recorded habit without losing existing achievement data
- [ ] User can unskip a previously skipped habit and retain any preserved data
- [ ] Validation logic permits reasonable data preservation scenarios
- [ ] No user data is lost during skip state transitions
- [ ] System behavior is permissive and avoids data deletion where possible

## 3. Architecture

*AI to complete when changes are architecturally significant, or when asked, prior to implementation plan.*

## 4. Implementation Plan & Progress

**Overall Status:** `In Progress`

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

- [ ] **Phase 1: Analysis & Design**
  - [x] **Sub-task 1.1:** Analyze current validation logic and entry collection flow
    - *Design:* Understand conflict between validation rules and data preservation
    - *Code/Artifacts:* Analysis of `internal/models/entry.go:186` and collection flows
    - *Testing Strategy:* Review existing test coverage for validation
    - *AI Notes:* Core issue is strict validation vs permissive data handling
  - [WIP] **Sub-task 1.2:** Design preservation strategy for skip transitions
    - *Design:* Choose between validation relaxation vs data cleanup approaches
    - *Code/Artifacts:* Design document or architectural decision
    - *Testing Strategy:* Define test scenarios for skip/unskip transitions
    - *AI Notes:* Need to consider impact on data integrity

- [ ] **Phase 2: Implementation**
  - [ ] **Sub-task 2.1:** Implement chosen preservation strategy
    - *Design:* Modify validation logic or collection flow to handle preserved data
    - *Code/Artifacts:* `internal/models/entry.go` or collection flow modifications
    - *Testing Strategy:* Unit tests for validation, integration tests for entry flow
    - *AI Notes:* Ensure backward compatibility with existing entries
  - [ ] **Sub-task 2.2:** Add tests for skip preservation scenarios
    - *Design:* Cover skip→unskip, unskip→skip, and data preservation cases
    - *Code/Artifacts:* Test files covering validation and collection scenarios
    - *Testing Strategy:* Automated tests for all transition scenarios
    - *AI Notes:* Include edge cases like partial data preservation

- [ ] **Phase 3: Verification**
  - [ ] **Sub-task 3.1:** Manual testing of skip functionality
    - *Design:* Test actual `iter entry` command with skip scenarios
    - *Code/Artifacts:* Manual test validation
    - *Testing Strategy:* Reproduce original bug scenario and verify fix
    - *AI Notes:* Ensure user experience matches acceptance criteria

## 5. Roadblocks

*(Timestamped list of any impediments. AI adds here when a sub-task is marked `[blocked]`)*

## 6. Notes / Discussion Log

*(Timestamped notes, decisions, clarifications from User or AI during the task's lifecycle)*

- `2025-07-14 - User:` Reported bug: `iter entry` fails when skipping previously filled habit with "skipped entries cannot have achievement levels" error. Requested permissive behavior to preserve user data.
- `2025-07-14 - AI:` Completed analysis of validation logic in `internal/models/entry.go:186` and entry collection flows. Core conflict identified between strict validation rules and data preservation during status transitions. Achievement levels are preserved during loading but validation rejects them for skipped entries.