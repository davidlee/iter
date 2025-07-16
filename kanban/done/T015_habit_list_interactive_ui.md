---
title: "Interactive Habit List UI - vice habit list"
type: ["feature"]
tags: ["cli", "ui", "habits", "table", "interactive"]
related_tasks: []
context_windows: ["cmd/habit/**/*.go", "internal/ui/**/*.go", "internal/models/habit.go", "CLAUDE.md", "kanban/CLAUDE.md"]
---

# Interactive Habit List UI - vice habit list

**Context (Background)**:
The CLI currently has placeholder implementations for habit listing, editing, and removal. Habit creation is fully implemented using bubbletea/huh patterns. The interactive list UI should integrate with existing HabitConfigurator routing and reuse established habit loading/validation patterns.

**Context (Significant Code Files)**:
- `cmd/habit_list.go` - Current placeholder calling configurator.ListHabits()
- `internal/ui/habitconfig/configurator.go` - Main habit orchestrator with routing methods
- `internal/parser/habits.go` - Habit loading/saving with ID persistence  
- `internal/models/habit.go` - Habit data structures and validation
- `internal/ui/todo.go` - Reference table UI implementation patterns

## Git Commit History

**All commits related to this task (newest first):**

- `4327c9b` - feat(habitconfig)[T015/3]: Phase 3 complete - edit and delete operations
- Previous commits from Phases 1-2 in git history

## 1. Habit / User Story

Users need an interactive way to view, manage, and modify their habits through a command-line interface. The `vice habit list` command should provide a table-based interface where users can navigate through their habits, view details, and perform management operations (edit, delete) without leaving the interface.

This enhances the current habit management workflow by providing a unified interface for habit operations rather than requiring separate commands for each action.

## 2. Acceptance Criteria

### Core Functionality
- [x] `vice habit list` command launches an interactive list interface
- [x] List displays all habits with title, type, and description
- [x] Uses charmbracelet/bubbles list component (decision changed from table)
- [x] Interface handles empty habit lists gracefully

### Navigation & Interaction
- [x] Users can navigate through habits using vim-style keybindings (j/k, arrow keys)
- [x] Users can view detailed habit information in modal overlay (enter/space)
- [x] Users can edit habits using existing habit configuration UI (e key)
- [x] Users can exit the interface cleanly (q/escape)

### Delete Operations
- [x] Users can delete habits with confirmation prompt (d key)
- [x] Delete confirmation includes option to create backup file (default yes)
- [x] Delete operations maintain data integrity and validation

### Search & Filtering
- [x] Users can trigger fuzzy/substring search with "/" key (built into bubbles/list)
- [x] Search filters habits by title and type
- [x] Search interface provides clear feedback and escape mechanism

### Technical Requirements  
- [x] Interface uses consistent keybindings managed through bubbles/key
- [x] Keybinding system designed for future user configurability
- [x] List interface supports filtering capabilities via built-in bubbles/list functionality
- [x] All operations provide clear user feedback

## 3. Architecture

### Component Selection: bubbles/list vs bubbles/table

**Decision: Use bubbles/list component**

**Rationale:**
- **Built-in fuzzy filtering**: Essential for managing larger habit lists, eliminates custom search implementation
- **Better UX patterns**: List component follows more idiomatic bubbletea patterns for interactive selection
- **Future extensibility**: Pagination, custom delegates, and filtering capabilities align with likely future requirements
- **Implementation efficiency**: Built-in functionality (search, navigation, filtering) reduces custom development effort

**Trade-offs accepted:**
- More complex setup for tabular-style data display (requires Item interface implementation)
- Less natural for pure columnar data compared to table component
- Custom delegate needed to achieve table-like rendering within list items

**Technical approach:**
- Implement `HabitItem` type satisfying `list.Item` interface
- Custom delegate for tabular-style rendering within list items  
- Configure fuzzy filtering on habit title and type fields
- Follow existing bubbletea Model-View-Update patterns from habit creators

## 4. Implementation Plan & Progress

**Overall Status:** `Complete`

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

- [x] **Phase 1: Core List Implementation**
  - [x] **Sub-task 1.1:** Create HabitItem type implementing list.Item interface
    - *Design:* Implement FilterValue(), Title(), Description() for habit display
    - *Code/Artifacts:* `internal/ui/habitconfig/habit_list.go` - New list model
    - *Testing Strategy:* Unit tests for HabitItem formatting and filtering
    - *AI Notes:* Follow existing bubbletea patterns, reuse habit loading from parser
  - [x] **Sub-task 1.2:** Implement basic HabitListModel with bubbles/list
    - *Design:* Bubbletea Model-View-Update pattern, integration with configurator
    - *Code/Artifacts:* List model with habit loading, basic navigation
    - *Testing Strategy:* Integration tests with sample habits.yml files
    - *AI Notes:* Handle empty lists gracefully, follow styling from todo.go
  - [x] **Sub-task 1.3:** Integrate with HabitConfigurator.ListHabits()
    - *Design:* Replace placeholder implementation, maintain error handling patterns
    - *Code/Artifacts:* `internal/ui/habitconfig/configurator.go:165` update
    - *Testing Strategy:* CLI integration tests with existing habit files
    - *AI Notes:* Preserve existing path management and validation flows

- [x] **Phase 2: Modal and Detail Views**
  - [x] **Sub-task 2.1:** Implement habit detail modal overlay
    - *Design:* Modal component showing full habit information (title, type, criteria, etc.)
    - *Code/Artifacts:* Modal view within list model, styled habit detail rendering
    - *Testing Strategy:* Test modal for all habit types (simple, elastic, checklist, informational)
    - *AI Notes:* Implemented comprehensive modal with proper styling, emoji display, and detailed criteria rendering. Added robust tests for modal functionality and keybinding behavior.
  - [x] **Sub-task 2.2:** Add keybinding management with bubbles/key
    - *Design:* Centralized keybinding definitions, vim-style navigation
    - *Code/Artifacts:* Key definitions for list navigation, modal, edit, delete operations
    - *Testing Strategy:* Test all keybinding scenarios, ensure no conflicts
    - *AI Notes:* Implemented comprehensive keybinding system with HabitListKeyMap struct, default bindings with vim-style navigation (j/k + arrows), modal controls (enter/space/ESC), prepared future operations (e/d//) with TODO placeholders, and WithKeyMap() method for user configurability. Added extensive tests for all keybinding scenarios.

- [x] **Phase 3: Edit and Delete Operations**
  - [x] **Sub-task 3.1:** Integrate habit editing with existing creators
    - *Design:* Launch appropriate habit creator (Simple/Elastic/Checklist/Informational) for selected habit
    - *Code/Artifacts:* Edit operation routing, habit pre-population in creators
    - *Testing Strategy:* Test editing for all habit types, validate data preservation
    - *AI Notes:* Implemented edit-mode constructors for all habit creators with habit-to-data conversion. Added EditHabitByID method with position preservation and ID retention. Integrated 'e' key in habit list UI with quit-and-edit pattern. Edit operations maintain habit position in list and preserve original ID as requested.
  - [x] **Sub-task 3.2:** Implement habit deletion with backup option
    - *Design:* Confirmation dialog with backup file creation option (default yes)
    - *Code/Artifacts:* Delete confirmation UI, backup file generation, habit removal
    - *Testing Strategy:* Test delete operation, backup file creation, validation after removal
    - *AI Notes:* Implemented RemoveHabitByID with confirmation dialog including backup option (defaults to yes). Added backup file handling with overwrite protection - prompts before overwriting existing .backup files. Habit removal from schema by filtering out target habit. Integrated 'd' key in habit list UI. Handles empty habit list after deletion gracefully.


## 5. Roadblocks

*(Timestamped list of any impediments. AI adds here when a sub-task is marked `[blocked]`)*

- No roadblocks encountered. All phases completed successfully with comprehensive functionality implemented and tested.

## 6. Notes / Discussion Log

*(Timestamped notes, decisions, clarifications from User or AI during the task's lifecycle)*

- `2025-07-14 - User:` Requested interactive habit list UI using table bubble and bubbles/key for consistent keybindings. Should provide view, edit, delete operations within single interface.
- `2025-07-14 - User:` Requirements clarified: Show all habits (ID, Title, Type, Status), modal overlay for details, existing UI for editing, vim keybindings, backup option for deletion, future search with "/" key for title filtering.
- `2025-07-14 - User:` Technical considerations: viewport/paginator may be needed for long lists. Consider bubbles/list component instead of table - already has fuzzy search and appropriate bindings built-in.
- `2025-07-14 - AI:` Analysis complete. Recommendation: Use bubbles/list component over table for built-in fuzzy filtering, better UX patterns, and future extensibility. Implementation plan created with 4 phases covering core list, modal views, operations, and search functionality.
- `2025-07-14 - AI:` Phase 1 complete. Implemented HabitItem and HabitListModel with bubbles/list integration. Created comprehensive unit and integration tests. Successfully integrated with HabitConfigurator.ListHabits() with proper error handling for missing/empty files.
- `2025-07-14 - AI:` UI revision complete. Changed from tabular to clean vertical list format with emojis (✅🎯📊📝), inverted title styling, item count display, and legend. Removed custom delegate in favor of default bubbles/list styling for better UX.
- `2025-07-14 - AI:` Phase 2.1 complete. Implemented habit detail modal overlay with enter/space to open and ESC to close. Modal displays comprehensive habit information including title with emoji, description, habit details (ID, type, field, scoring), habit-type specific details (achievement levels for elastic habits, direction for informational habits), criteria with proper formatting, and UI prompts. Added proper styling with purple border and organized layout. Created extensive tests for modal functionality and criteria rendering.
- `2025-07-14 - AI:` Phase 2.2 complete. Implemented centralized keybinding management using bubbles/key. Created HabitListKeyMap struct with vim-style navigation (j/k + arrows), modal controls (enter/space/ESC), and prepared future operations (e/d//) with TODO placeholders. Added WithKeyMap() method for user configurability. Enhanced Update() method to use key.Matches() for cleaner key handling. Modal keys are properly isolated to prevent interference. Created comprehensive tests for all keybinding scenarios and custom keymap functionality.
- `2025-07-14 - AI:` UX improvements: Added 'q' key to close modal for consistency. Integrated ShowDetail key into main help system via AdditionalShortHelpKeys. Made modal footer dynamic using Help().Key instead of hardcoded text. Added AIDEV anchor comments documenting key architectural patterns: keybinding-architecture (centralized management), help-integration (bubbles/list integration), dynamic-help-text (configurability), and modal-key-isolation (event handling priority).
- `2025-07-14 - AI:` Phase 3 complete. Implemented comprehensive habit editing and deletion operations. Edit functionality: Created NewXXXHabitCreatorForEdit constructors for all habit types with habit-to-data conversion logic. Edit operations preserve habit position and ID as requested. Added EditHabitByID method with routing to appropriate creators. Integrated 'e' key in habit list with quit-and-edit pattern that returns to updated list after editing. Delete functionality: Implemented RemoveHabitByID with dual confirmation dialog (delete + backup option). Backup handling includes overwrite protection for existing .backup files. Habit removal updates schema and handles empty list gracefully. Integrated 'd' key in habit list. Both operations use habit list UI for selection, maintaining consistent UX. CLI commands (vice habit edit/remove) delegate to interactive list for seamless user experience.
- `2025-07-14 - AI:` Task completion: All acceptance criteria met. Built-in bubbles/list filtering with "/" key provides complete search functionality. Help text updated to show edit/delete operations. Boolean habit editing crash fixed. Task ready for completion.

### Key Code Files Modified in Phase 3:
- `internal/ui/habitconfig/configurator.go` - EditHabitByID/RemoveHabitByID methods, confirmation dialogs, backup handling
- `internal/ui/habitconfig/habit_list.go` - Edit/delete integration with selectedHabitForEdit/Delete fields
- `internal/ui/habitconfig/simple_habit_creator.go` - NewSimpleHabitCreatorForEdit with habitToTestData conversion
- `internal/ui/habitconfig/elastic_habit_creator.go` - NewElasticHabitCreatorForEdit with habitToTestElasticData conversion
- `internal/ui/habitconfig/informational_habit_creator.go` - NewInformationalHabitCreatorForEdit with pre-population
- `internal/ui/habitconfig/checklist_habit_creator.go` - NewChecklistHabitCreatorForEdit with checklist ID preservation

### Critical Design Patterns Established:
1. **Habit-to-Data Conversion**: Reverse engineering from models.Habit to TestHabitData structures enables seamless edit mode
2. **Position Preservation Architecture**: Edit operations maintain habit.Position and habit.ID for future reordering support
3. **Quit-and-Return UI Pattern**: Operations exit list UI, perform action, then return to refreshed list for consistent UX
4. **Backup Protection Strategy**: Default yes for backups with overwrite confirmation prevents accidental data loss
5. **CLI Delegation Pattern**: Public methods delegate to interactive UI while internal ByID methods handle specific operations

### Future Improvements for Next Developer:
1. **Search Enhancement**: bubbles/list provides built-in filtering with "/" key - search functionality is complete
2. **Habit Reordering**: Architecture ready - add up/down arrow handlers and position updates
3. **Bulk Operations**: Select multiple habits for batch edit/delete operations
4. **Undo/Redo**: Leverage backup files for habit restoration functionality
5. **Export/Import**: Habit list could support exporting selected habits to new YAML files
6. **Performance**: Consider pagination or virtualization for very large habit lists (>1000 habits)

### Testing Strategy Notes:
- Edit operations should test all habit types with complex field configurations
- Delete operations should verify backup file creation and overwrite scenarios
- UI integration tests should verify quit-and-return behavior
- Error scenarios: missing files, invalid habit IDs, permission issues
- Load testing: verify performance with 100+ habits