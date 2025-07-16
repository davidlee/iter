---
title: "Habit Management UI"
type: ["feature"]
tags: ["ui", "habits", "management", "cli"]
related_tasks: ["depends-on:T005"]
context_windows: ["./CLAUDE.md", "./doc/specifications/habit_schema.md", "./internal/models/*.go", "./internal/parser/*.go", "./internal/ui/*.go", "./cmd/*.go"]
---

# Habit Management UI

## Git Commit History

**All commits related to this task (newest first):**

*No commits yet - task is in backlog*

## 1. Habit / User Story

As a user, I want to manage my existing habits through an interactive CLI interface, so that I can view, edit, and remove habits without needing to manually edit YAML files.

The system should provide:
- `vice habit list` - Display existing habits in a readable format with detail views
- `vice habit edit` - Select and modify existing habit definitions with validation
- `vice habit remove` - Remove existing habits with confirmation and impact analysis

This builds upon T005 (Habit Configuration UI) to provide comprehensive habit lifecycle management while maintaining the existing file-based storage approach.

## 2. Acceptance Criteria

### Habit Listing
- [ ] `vice habit list` displays existing habits in human-readable format
- [ ] Habit selection to view detailed information
- [ ] Habit detail / summary view with rich table display
- [ ] Navigation back from habit detail to list
- [ ] Filtering and sorting capabilities (future enhancement)

### Habit Editing
- [ ] `vice habit edit` allows selection and modification of existing habits
- [ ] Interactive habit selection - reuses listing interface with edit context
- [ ] Default action on [enter] is edit when in edit mode, view when in list mode
- [ ] Wizard-style editing with current values pre-populated
- [ ] Support for editing all habit types (simple, elastic, informational)
- [ ] Preview and confirm changes before saving
- [ ] Allow editing of habit ID with impact analysis
- [ ] Confirm whether to update historical data in entries.yml to use new ID
- [ ] Error recovery and validation during editing
- [ ] Preserve habit ID and data integrity

### Habit Removal
- [ ] `vice habit remove` removes habits with confirmation prompt
- [ ] Interactive habit selection with details display
- [ ] Impact analysis showing entries that reference the habit
- [ ] Confirmation prompt with habit summary
- [ ] Safe removal with backup options
- [ ] Warning if habit has historical data

### Technical Requirements
- [ ] All operations preserve habit IDs and maintain data integrity
- [ ] File operations are atomic (no partial writes)
- [ ] Comprehensive error handling for file operations
- [ ] Graceful handling of file permission errors
- [ ] Backup existing habits.yml before modifications
- [ ] Leverage existing validation logic from models package
- [ ] Reuse existing parser and file operations
- [ ] Follow established UI patterns from T005 habit configuration

---

## 3. Implementation Plan & Progress

**Overall Status:** `Backlog`

**Dependencies:** 
- T005 Habit Configuration UI must be completed first
- Requires completed habit creation infrastructure and patterns

**Planned Implementation Approach:**

### Phase 1: Habit Listing Infrastructure

- [ ] **1.1 Habit Listing Foundation**
  - [ ] Create `ListHabits()` method in HabitConfigurator
  - [ ] Design habit listing model using bubbletea following T005 patterns
  - [ ] Implement habit selection interface with keyboard navigation
  - [ ] Create habit summary display components

- [ ] **1.2 Habit Detail View**
  - [ ] Design detailed habit information display
  - [ ] Show all habit properties: type, field configuration, scoring, criteria
  - [ ] Format different habit types appropriately (simple vs elastic vs informational)
  - [ ] Navigation between list and detail views

### Phase 2: Habit Editing System

- [ ] **2.1 Habit Selection for Editing**
  - [ ] Reuse habit listing interface with edit context
  - [ ] Clear indication of edit mode vs view mode
  - [ ] Habit selection with pre-edit confirmation

- [ ] **2.2 Edit Wizard Implementation**
  - [ ] Create HabitEditWizard bubbletea model
  - [ ] Pre-populate forms with existing habit values
  - [ ] Support editing basic info (title, description)
  - [ ] Support editing field configuration
  - [ ] Support editing scoring and criteria

- [ ] **2.3 Habit ID Editing**
  - [ ] Special handling for habit ID changes
  - [ ] Impact analysis: scan entries.yml for references
  - [ ] Confirmation dialog showing impact
  - [ ] Option to update historical data or preserve old ID

- [ ] **2.4 Change Preview and Confirmation**
  - [ ] Show before/after comparison
  - [ ] Validation of modified habit
  - [ ] Atomic save operation with rollback capability

### Phase 3: Habit Removal System

- [ ] **3.1 Safe Habit Removal**
  - [ ] Create HabitRemovalConfirmation interface
  - [ ] Scan for habit usage in entries.yml
  - [ ] Display impact analysis to user
  - [ ] Multi-step confirmation for habits with data

- [ ] **3.2 Backup and Recovery**
  - [ ] Automatic backup before removal
  - [ ] Recovery instructions if accidental removal
  - [ ] Option to archive habit instead of delete

### Phase 4: Integration and Polish

- [ ] **4.1 File Operation Safety**
  - [ ] Atomic habit modifications using existing parser
  - [ ] Error handling for file permissions, disk space
  - [ ] Backup existing habits.yml before all modifications
  - [ ] Rollback mechanisms for failed operations

- [ ] **4.2 Advanced Validation**
  - [ ] Leverage existing `Habit.Validate()` and `Schema.Validate()`
  - [ ] Real-time validation during editing
  - [ ] Clear error messages for validation failures
  - [ ] Cross-habit validation (e.g., duplicate IDs)

- [ ] **4.3 Testing & Documentation**
  - [ ] Unit tests for all management operations
  - [ ] Integration tests for complete workflows
  - [ ] Edge case testing (corrupted files, permission errors)
  - [ ] Update CLI help text and documentation

**Implementation Strategy:**
- Build on T005 patterns and infrastructure
- Reuse HabitConfigurator and bubbletea models where possible
- Follow established UI patterns for consistency
- Prioritize data safety and atomic operations
- Implement comprehensive backup and recovery systems

**Key Technical Decisions:**
1. **Pattern Reuse**: Leverage T005 bubbletea patterns and form structures
2. **Data Safety**: All operations must be atomic with rollback capabilities
3. **Impact Analysis**: Always show users the consequences of their actions
4. **Progressive Enhancement**: Start with basic list/edit/remove, add advanced features later
5. **Integration**: Seamless integration with existing habit configuration flows

---

## 4. Roadblocks

*(None identified yet - depends on T005 completion)*

---

## 5. Notes / Discussion Log

- Created as continuation of T005 Habit Configuration UI
- Focuses on habit lifecycle management after creation
- Emphasizes data safety and user confirmation for destructive operations
- Plans for integration with entry system impact analysis
- Builds upon established bubbletea/huh patterns from T005

**Design Considerations:**
- Habit editing should support all habit types created in T005
- Removal operations need careful impact analysis
- File operations must be atomic to prevent corruption
- UI patterns should be consistent with T005 implementation
- Consider future integration with entry management features