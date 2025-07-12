---
title: "Goal Management UI"
type: ["feature"]
tags: ["ui", "goals", "management", "cli"]
related_tasks: ["depends-on:T005"]
context_windows: ["./CLAUDE.md", "./doc/specifications/goal_schema.md", "./internal/models/*.go", "./internal/parser/*.go", "./internal/ui/*.go", "./cmd/*.go"]
---

# Goal Management UI

## 1. Goal / User Story

As a user, I want to manage my existing goals through an interactive CLI interface, so that I can view, edit, and remove goals without needing to manually edit YAML files.

The system should provide:
- `iter goal list` - Display existing goals in a readable format with detail views
- `iter goal edit` - Select and modify existing goal definitions with validation
- `iter goal remove` - Remove existing goals with confirmation and impact analysis

This builds upon T005 (Goal Configuration UI) to provide comprehensive goal lifecycle management while maintaining the existing file-based storage approach.

## 2. Acceptance Criteria

### Goal Listing
- [ ] `iter goal list` displays existing goals in human-readable format
- [ ] Goal selection to view detailed information
- [ ] Goal detail / summary view with rich table display
- [ ] Navigation back from goal detail to list
- [ ] Filtering and sorting capabilities (future enhancement)

### Goal Editing
- [ ] `iter goal edit` allows selection and modification of existing goals
- [ ] Interactive goal selection - reuses listing interface with edit context
- [ ] Default action on [enter] is edit when in edit mode, view when in list mode
- [ ] Wizard-style editing with current values pre-populated
- [ ] Support for editing all goal types (simple, elastic, informational)
- [ ] Preview and confirm changes before saving
- [ ] Allow editing of goal ID with impact analysis
- [ ] Confirm whether to update historical data in entries.yml to use new ID
- [ ] Error recovery and validation during editing
- [ ] Preserve goal ID and data integrity

### Goal Removal
- [ ] `iter goal remove` removes goals with confirmation prompt
- [ ] Interactive goal selection with details display
- [ ] Impact analysis showing entries that reference the goal
- [ ] Confirmation prompt with goal summary
- [ ] Safe removal with backup options
- [ ] Warning if goal has historical data

### Technical Requirements
- [ ] All operations preserve goal IDs and maintain data integrity
- [ ] File operations are atomic (no partial writes)
- [ ] Comprehensive error handling for file operations
- [ ] Graceful handling of file permission errors
- [ ] Backup existing goals.yml before modifications
- [ ] Leverage existing validation logic from models package
- [ ] Reuse existing parser and file operations
- [ ] Follow established UI patterns from T005 goal configuration

---

## 3. Implementation Plan & Progress

**Overall Status:** `Backlog`

**Dependencies:** 
- T005 Goal Configuration UI must be completed first
- Requires completed goal creation infrastructure and patterns

**Planned Implementation Approach:**

### Phase 1: Goal Listing Infrastructure

- [ ] **1.1 Goal Listing Foundation**
  - [ ] Create `ListGoals()` method in GoalConfigurator
  - [ ] Design goal listing model using bubbletea following T005 patterns
  - [ ] Implement goal selection interface with keyboard navigation
  - [ ] Create goal summary display components

- [ ] **1.2 Goal Detail View**
  - [ ] Design detailed goal information display
  - [ ] Show all goal properties: type, field configuration, scoring, criteria
  - [ ] Format different goal types appropriately (simple vs elastic vs informational)
  - [ ] Navigation between list and detail views

### Phase 2: Goal Editing System

- [ ] **2.1 Goal Selection for Editing**
  - [ ] Reuse goal listing interface with edit context
  - [ ] Clear indication of edit mode vs view mode
  - [ ] Goal selection with pre-edit confirmation

- [ ] **2.2 Edit Wizard Implementation**
  - [ ] Create GoalEditWizard bubbletea model
  - [ ] Pre-populate forms with existing goal values
  - [ ] Support editing basic info (title, description)
  - [ ] Support editing field configuration
  - [ ] Support editing scoring and criteria

- [ ] **2.3 Goal ID Editing**
  - [ ] Special handling for goal ID changes
  - [ ] Impact analysis: scan entries.yml for references
  - [ ] Confirmation dialog showing impact
  - [ ] Option to update historical data or preserve old ID

- [ ] **2.4 Change Preview and Confirmation**
  - [ ] Show before/after comparison
  - [ ] Validation of modified goal
  - [ ] Atomic save operation with rollback capability

### Phase 3: Goal Removal System

- [ ] **3.1 Safe Goal Removal**
  - [ ] Create GoalRemovalConfirmation interface
  - [ ] Scan for goal usage in entries.yml
  - [ ] Display impact analysis to user
  - [ ] Multi-step confirmation for goals with data

- [ ] **3.2 Backup and Recovery**
  - [ ] Automatic backup before removal
  - [ ] Recovery instructions if accidental removal
  - [ ] Option to archive goal instead of delete

### Phase 4: Integration and Polish

- [ ] **4.1 File Operation Safety**
  - [ ] Atomic goal modifications using existing parser
  - [ ] Error handling for file permissions, disk space
  - [ ] Backup existing goals.yml before all modifications
  - [ ] Rollback mechanisms for failed operations

- [ ] **4.2 Advanced Validation**
  - [ ] Leverage existing `Goal.Validate()` and `Schema.Validate()`
  - [ ] Real-time validation during editing
  - [ ] Clear error messages for validation failures
  - [ ] Cross-goal validation (e.g., duplicate IDs)

- [ ] **4.3 Testing & Documentation**
  - [ ] Unit tests for all management operations
  - [ ] Integration tests for complete workflows
  - [ ] Edge case testing (corrupted files, permission errors)
  - [ ] Update CLI help text and documentation

**Implementation Strategy:**
- Build on T005 patterns and infrastructure
- Reuse GoalConfigurator and bubbletea models where possible
- Follow established UI patterns for consistency
- Prioritize data safety and atomic operations
- Implement comprehensive backup and recovery systems

**Key Technical Decisions:**
1. **Pattern Reuse**: Leverage T005 bubbletea patterns and form structures
2. **Data Safety**: All operations must be atomic with rollback capabilities
3. **Impact Analysis**: Always show users the consequences of their actions
4. **Progressive Enhancement**: Start with basic list/edit/remove, add advanced features later
5. **Integration**: Seamless integration with existing goal configuration flows

---

## 4. Roadblocks

*(None identified yet - depends on T005 completion)*

---

## 5. Notes / Discussion Log

- Created as continuation of T005 Goal Configuration UI
- Focuses on goal lifecycle management after creation
- Emphasizes data safety and user confirmation for destructive operations
- Plans for integration with entry system impact analysis
- Builds upon established bubbletea/huh patterns from T005

**Design Considerations:**
- Goal editing should support all goal types created in T005
- Removal operations need careful impact analysis
- File operations must be atomic to prevent corruption
- UI patterns should be consistent with T005 implementation
- Consider future integration with entry management features