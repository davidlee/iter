---
title: "Interactive Goal List UI - iter goal list"
type: ["feature"]
tags: ["cli", "ui", "goals", "table", "interactive"]
related_tasks: []
context_windows: ["cmd/goal/**/*.go", "internal/ui/**/*.go", "internal/models/goal.go", "CLAUDE.md", "kanban/CLAUDE.md"]
---

# Interactive Goal List UI - iter goal list

**Context (Background)**:
The CLI currently has placeholder implementations for goal listing, editing, and removal. Goal creation is fully implemented using bubbletea/huh patterns. The interactive list UI should integrate with existing GoalConfigurator routing and reuse established goal loading/validation patterns.

**Context (Significant Code Files)**:
- `cmd/goal_list.go` - Current placeholder calling configurator.ListGoals()
- `internal/ui/goalconfig/configurator.go` - Main goal orchestrator with routing methods
- `internal/parser/goals.go` - Goal loading/saving with ID persistence  
- `internal/models/goal.go` - Goal data structures and validation
- `internal/ui/todo.go` - Reference table UI implementation patterns

## Git Commit History

**All commits related to this task (newest first):**

*No commits yet - task is in backlog*

## 1. Goal / User Story

Users need an interactive way to view, manage, and modify their goals through a command-line interface. The `iter goal list` command should provide a table-based interface where users can navigate through their goals, view details, and perform management operations (edit, delete) without leaving the interface.

This enhances the current goal management workflow by providing a unified interface for goal operations rather than requiring separate commands for each action.

## 2. Acceptance Criteria

### Core Functionality
- [ ] `iter goal list` command launches an interactive table interface
- [ ] Table displays all goals with columns: ID, Title, Type, Status
- [ ] Table uses charmbracelet/bubbles table component
- [ ] Interface handles empty goal lists gracefully

### Navigation & Interaction
- [ ] Users can navigate through goals using vim-style keybindings (j/k, arrow keys)
- [ ] Users can view detailed goal information in modal overlay (enter/space)
- [ ] Users can edit goals using existing goal configuration UI (e key)
- [ ] Users can exit the interface cleanly (q/escape)

### Delete Operations
- [ ] Users can delete goals with confirmation prompt (d key)
- [ ] Delete confirmation includes option to create backup file (default yes)
- [ ] Delete operations maintain data integrity and validation

### Search & Filtering
- [ ] Users can trigger fuzzy/substring search with "/" key
- [ ] Search filters goals by title match
- [ ] Search interface provides clear feedback and escape mechanism

### Technical Requirements  
- [ ] Interface uses consistent keybindings managed through bubbles/key
- [ ] Keybinding system designed for future user configurability
- [ ] Table interface supports future filtering capabilities
- [ ] All operations provide clear user feedback

## 3. Architecture

### Component Selection: bubbles/list vs bubbles/table

**Decision: Use bubbles/list component**

**Rationale:**
- **Built-in fuzzy filtering**: Essential for managing larger goal lists, eliminates custom search implementation
- **Better UX patterns**: List component follows more idiomatic bubbletea patterns for interactive selection
- **Future extensibility**: Pagination, custom delegates, and filtering capabilities align with likely future requirements
- **Implementation efficiency**: Built-in functionality (search, navigation, filtering) reduces custom development effort

**Trade-offs accepted:**
- More complex setup for tabular-style data display (requires Item interface implementation)
- Less natural for pure columnar data compared to table component
- Custom delegate needed to achieve table-like rendering within list items

**Technical approach:**
- Implement `GoalItem` type satisfying `list.Item` interface
- Custom delegate for tabular-style rendering within list items  
- Configure fuzzy filtering on goal title and type fields
- Follow existing bubbletea Model-View-Update patterns from goal creators

## 4. Implementation Plan & Progress

**Overall Status:** `In Progress`

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

- [ ] **Phase 1: Core List Implementation**
  - [x] **Sub-task 1.1:** Create GoalItem type implementing list.Item interface
    - *Design:* Implement FilterValue(), Title(), Description() for goal display
    - *Code/Artifacts:* `internal/ui/goalconfig/goal_list.go` - New list model
    - *Testing Strategy:* Unit tests for GoalItem formatting and filtering
    - *AI Notes:* Follow existing bubbletea patterns, reuse goal loading from parser
  - [WIP] **Sub-task 1.2:** Implement basic GoalListModel with bubbles/list
    - *Design:* Bubbletea Model-View-Update pattern, integration with configurator
    - *Code/Artifacts:* List model with goal loading, basic navigation
    - *Testing Strategy:* Integration tests with sample goals.yml files
    - *AI Notes:* Handle empty lists gracefully, follow styling from todo.go
  - [ ] **Sub-task 1.3:** Integrate with GoalConfigurator.ListGoals()
    - *Design:* Replace placeholder implementation, maintain error handling patterns
    - *Code/Artifacts:* `internal/ui/goalconfig/configurator.go:165` update
    - *Testing Strategy:* CLI integration tests with existing goal files
    - *AI Notes:* Preserve existing path management and validation flows

- [ ] **Phase 2: Modal and Detail Views**
  - [ ] **Sub-task 2.1:** Implement goal detail modal overlay
    - *Design:* Modal component showing full goal information (title, type, criteria, etc.)
    - *Code/Artifacts:* Modal view within list model, styled goal detail rendering
    - *Testing Strategy:* Test modal for all goal types (simple, elastic, checklist, informational)
    - *AI Notes:* Reuse goal formatting patterns, handle long content gracefully
  - [ ] **Sub-task 2.2:** Add keybinding management with bubbles/key
    - *Design:* Centralized keybinding definitions, vim-style navigation
    - *Code/Artifacts:* Key definitions for list navigation, modal, edit, delete operations
    - *Testing Strategy:* Test all keybinding scenarios, ensure no conflicts
    - *AI Notes:* Design for future user configurability, document key mappings

- [ ] **Phase 3: Edit and Delete Operations**
  - [ ] **Sub-task 3.1:** Integrate goal editing with existing creators
    - *Design:* Launch appropriate goal creator (Simple/Elastic/Checklist/Informational) for selected goal
    - *Code/Artifacts:* Edit operation routing, goal pre-population in creators
    - *Testing Strategy:* Test editing for all goal types, validate data preservation
    - *AI Notes:* Reuse existing creator flows, handle edit vs create mode differences
  - [ ] **Sub-task 3.2:** Implement goal deletion with backup option
    - *Design:* Confirmation dialog with backup file creation option (default yes)
    - *Code/Artifacts:* Delete confirmation UI, backup file generation, goal removal
    - *Testing Strategy:* Test delete operation, backup file creation, validation after removal
    - *AI Notes:* Follow atomic file operations pattern, maintain data integrity

- [ ] **Phase 4: Search and Filtering**
  - [ ] **Sub-task 4.1:** Configure built-in fuzzy search functionality
    - *Design:* Enable "/" key for search mode, filter by goal title and description
    - *Code/Artifacts:* Search configuration, filter function customization
    - *Testing Strategy:* Test fuzzy matching, search result highlighting, search exit
    - *AI Notes:* Leverage bubbles/list built-in filtering, customize for goal-specific needs
  - [ ] **Sub-task 4.2:** Add search result feedback and navigation
    - *Design:* Clear search status display, result count, easy search reset
    - *Code/Artifacts:* Search UI feedback, result navigation enhancements
    - *Testing Strategy:* Test search UX, empty results handling, search persistence
    - *AI Notes:* Follow existing UI feedback patterns, maintain search state appropriately

## 5. Roadblocks

*(Timestamped list of any impediments. AI adds here when a sub-task is marked `[blocked]`)*

## 6. Notes / Discussion Log

*(Timestamped notes, decisions, clarifications from User or AI during the task's lifecycle)*

- `2025-07-14 - User:` Requested interactive goal list UI using table bubble and bubbles/key for consistent keybindings. Should provide view, edit, delete operations within single interface.
- `2025-07-14 - User:` Requirements clarified: Show all goals (ID, Title, Type, Status), modal overlay for details, existing UI for editing, vim keybindings, backup option for deletion, future search with "/" key for title filtering.
- `2025-07-14 - User:` Technical considerations: viewport/paginator may be needed for long lists. Consider bubbles/list component instead of table - already has fuzzy search and appropriate bindings built-in.
- `2025-07-14 - AI:` Analysis complete. Recommendation: Use bubbles/list component over table for built-in fuzzy filtering, better UX patterns, and future extensibility. Implementation plan created with 4 phases covering core list, modal views, operations, and search functionality.