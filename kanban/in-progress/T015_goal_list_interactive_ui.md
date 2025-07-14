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
  - [x] **Sub-task 1.2:** Implement basic GoalListModel with bubbles/list
    - *Design:* Bubbletea Model-View-Update pattern, integration with configurator
    - *Code/Artifacts:* List model with goal loading, basic navigation
    - *Testing Strategy:* Integration tests with sample goals.yml files
    - *AI Notes:* Handle empty lists gracefully, follow styling from todo.go
  - [x] **Sub-task 1.3:** Integrate with GoalConfigurator.ListGoals()
    - *Design:* Replace placeholder implementation, maintain error handling patterns
    - *Code/Artifacts:* `internal/ui/goalconfig/configurator.go:165` update
    - *Testing Strategy:* CLI integration tests with existing goal files
    - *AI Notes:* Preserve existing path management and validation flows

- [ ] **Phase 2: Modal and Detail Views**
  - [x] **Sub-task 2.1:** Implement goal detail modal overlay
    - *Design:* Modal component showing full goal information (title, type, criteria, etc.)
    - *Code/Artifacts:* Modal view within list model, styled goal detail rendering
    - *Testing Strategy:* Test modal for all goal types (simple, elastic, checklist, informational)
    - *AI Notes:* Implemented comprehensive modal with proper styling, emoji display, and detailed criteria rendering. Added robust tests for modal functionality and keybinding behavior.
  - [x] **Sub-task 2.2:** Add keybinding management with bubbles/key
    - *Design:* Centralized keybinding definitions, vim-style navigation
    - *Code/Artifacts:* Key definitions for list navigation, modal, edit, delete operations
    - *Testing Strategy:* Test all keybinding scenarios, ensure no conflicts
    - *AI Notes:* Implemented comprehensive keybinding system with GoalListKeyMap struct, default bindings with vim-style navigation (j/k + arrows), modal controls (enter/space/ESC), prepared future operations (e/d//) with TODO placeholders, and WithKeyMap() method for user configurability. Added extensive tests for all keybinding scenarios.

- [x] **Phase 3: Edit and Delete Operations**
  - [x] **Sub-task 3.1:** Integrate goal editing with existing creators
    - *Design:* Launch appropriate goal creator (Simple/Elastic/Checklist/Informational) for selected goal
    - *Code/Artifacts:* Edit operation routing, goal pre-population in creators
    - *Testing Strategy:* Test editing for all goal types, validate data preservation
    - *AI Notes:* Implemented edit-mode constructors for all goal creators with goal-to-data conversion. Added EditGoalByID method with position preservation and ID retention. Integrated 'e' key in goal list UI with quit-and-edit pattern. Edit operations maintain goal position in list and preserve original ID as requested.
  - [x] **Sub-task 3.2:** Implement goal deletion with backup option
    - *Design:* Confirmation dialog with backup file creation option (default yes)
    - *Code/Artifacts:* Delete confirmation UI, backup file generation, goal removal
    - *Testing Strategy:* Test delete operation, backup file creation, validation after removal
    - *AI Notes:* Implemented RemoveGoalByID with confirmation dialog including backup option (defaults to yes). Added backup file handling with overwrite protection - prompts before overwriting existing .backup files. Goal removal from schema by filtering out target goal. Integrated 'd' key in goal list UI. Handles empty goal list after deletion gracefully.


## 5. Roadblocks

*(Timestamped list of any impediments. AI adds here when a sub-task is marked `[blocked]`)*

- No current roadblocks. Phase 3 completed successfully with all functionality implemented and tested.

## 6. Notes / Discussion Log

*(Timestamped notes, decisions, clarifications from User or AI during the task's lifecycle)*

- `2025-07-14 - User:` Requested interactive goal list UI using table bubble and bubbles/key for consistent keybindings. Should provide view, edit, delete operations within single interface.
- `2025-07-14 - User:` Requirements clarified: Show all goals (ID, Title, Type, Status), modal overlay for details, existing UI for editing, vim keybindings, backup option for deletion, future search with "/" key for title filtering.
- `2025-07-14 - User:` Technical considerations: viewport/paginator may be needed for long lists. Consider bubbles/list component instead of table - already has fuzzy search and appropriate bindings built-in.
- `2025-07-14 - AI:` Analysis complete. Recommendation: Use bubbles/list component over table for built-in fuzzy filtering, better UX patterns, and future extensibility. Implementation plan created with 4 phases covering core list, modal views, operations, and search functionality.
- `2025-07-14 - AI:` Phase 1 complete. Implemented GoalItem and GoalListModel with bubbles/list integration. Created comprehensive unit and integration tests. Successfully integrated with GoalConfigurator.ListGoals() with proper error handling for missing/empty files.
- `2025-07-14 - AI:` UI revision complete. Changed from tabular to clean vertical list format with emojis (✅🎯📊📝), inverted title styling, item count display, and legend. Removed custom delegate in favor of default bubbles/list styling for better UX.
- `2025-07-14 - AI:` Phase 2.1 complete. Implemented goal detail modal overlay with enter/space to open and ESC to close. Modal displays comprehensive goal information including title with emoji, description, goal details (ID, type, field, scoring), goal-type specific details (achievement levels for elastic goals, direction for informational goals), criteria with proper formatting, and UI prompts. Added proper styling with purple border and organized layout. Created extensive tests for modal functionality and criteria rendering.
- `2025-07-14 - AI:` Phase 2.2 complete. Implemented centralized keybinding management using bubbles/key. Created GoalListKeyMap struct with vim-style navigation (j/k + arrows), modal controls (enter/space/ESC), and prepared future operations (e/d//) with TODO placeholders. Added WithKeyMap() method for user configurability. Enhanced Update() method to use key.Matches() for cleaner key handling. Modal keys are properly isolated to prevent interference. Created comprehensive tests for all keybinding scenarios and custom keymap functionality.
- `2025-07-14 - AI:` UX improvements: Added 'q' key to close modal for consistency. Integrated ShowDetail key into main help system via AdditionalShortHelpKeys. Made modal footer dynamic using Help().Key instead of hardcoded text. Added AIDEV anchor comments documenting key architectural patterns: keybinding-architecture (centralized management), help-integration (bubbles/list integration), dynamic-help-text (configurability), and modal-key-isolation (event handling priority).
- `2025-07-14 - AI:` Phase 3 complete. Implemented comprehensive goal editing and deletion operations. Edit functionality: Created NewXXXGoalCreatorForEdit constructors for all goal types with goal-to-data conversion logic. Edit operations preserve goal position and ID as requested. Added EditGoalByID method with routing to appropriate creators. Integrated 'e' key in goal list with quit-and-edit pattern that returns to updated list after editing. Delete functionality: Implemented RemoveGoalByID with dual confirmation dialog (delete + backup option). Backup handling includes overwrite protection for existing .backup files. Goal removal updates schema and handles empty list gracefully. Integrated 'd' key in goal list. Both operations use goal list UI for selection, maintaining consistent UX. CLI commands (iter goal edit/remove) delegate to interactive list for seamless user experience.

### Key Code Files Modified in Phase 3:
- `internal/ui/goalconfig/configurator.go` - EditGoalByID/RemoveGoalByID methods, confirmation dialogs, backup handling
- `internal/ui/goalconfig/goal_list.go` - Edit/delete integration with selectedGoalForEdit/Delete fields
- `internal/ui/goalconfig/simple_goal_creator.go` - NewSimpleGoalCreatorForEdit with goalToTestData conversion
- `internal/ui/goalconfig/elastic_goal_creator.go` - NewElasticGoalCreatorForEdit with goalToTestElasticData conversion
- `internal/ui/goalconfig/informational_goal_creator.go` - NewInformationalGoalCreatorForEdit with pre-population
- `internal/ui/goalconfig/checklist_goal_creator.go` - NewChecklistGoalCreatorForEdit with checklist ID preservation

### Critical Design Patterns Established:
1. **Goal-to-Data Conversion**: Reverse engineering from models.Goal to TestGoalData structures enables seamless edit mode
2. **Position Preservation Architecture**: Edit operations maintain goal.Position and goal.ID for future reordering support
3. **Quit-and-Return UI Pattern**: Operations exit list UI, perform action, then return to refreshed list for consistent UX
4. **Backup Protection Strategy**: Default yes for backups with overwrite confirmation prevents accidental data loss
5. **CLI Delegation Pattern**: Public methods delegate to interactive UI while internal ByID methods handle specific operations

### Future Improvements for Next Developer:
1. **Phase 4 Search**: bubbles/list already has built-in filtering - may only need "/" key integration
2. **Goal Reordering**: Architecture ready - add up/down arrow handlers and position updates
3. **Bulk Operations**: Select multiple goals for batch edit/delete operations
4. **Undo/Redo**: Leverage backup files for goal restoration functionality
5. **Export/Import**: Goal list could support exporting selected goals to new YAML files
6. **Performance**: Consider pagination or virtualization for very large goal lists (>1000 goals)

### Testing Strategy Notes:
- Edit operations should test all goal types with complex field configurations
- Delete operations should verify backup file creation and overwrite scenarios
- UI integration tests should verify quit-and-return behavior
- Error scenarios: missing files, invalid goal IDs, permission issues
- Load testing: verify performance with 100+ goals