---
title: "Goal Configuration UI"
type: ["feature"]
tags: ["ui", "goals", "configuration", "cli"]
related_tasks: ["depends-on:T004"]
context_windows: ["./CLAUDE.md", "./doc/specifications/goal_schema.md", "./internal/models/*.go", "./internal/parser/*.go", "./internal/ui/*.go", "./cmd/*.go"]
---

# Goal Configuration UI

## 1. Goal / User Story

As a user, I want to configure my goals through an interactive CLI interface rather than manually editing YAML files, so that I can easily add, modify, and manage my goals without needing to understand the YAML schema syntax.

The system should provide:
- `iter goal add` - Interactive UI to create new goals with guided prompts
- `iter goal list` - Display existing goals in a readable format
- `iter goal edit` - Select and modify existing goal definitions
- `iter goal remove` - Remove existing goals with confirmation

This eliminates the need for users to manually edit YAML and reduces configuration errors while maintaining the existing file-based storage approach.

## 2. Acceptance Criteria

### Core Functionality
- [ ] `iter goal add` command creates new goals through interactive prompts
- [ ] Goal type selection (simple, elastic, informational) with appropriate follow-up questions
- [ ] Field type selection with validation and guidance
- [ ] Scoring type configuration (manual/automatic) with criteria definition for automatic scoring
- [ ] Elastic goal criteria definition (mini/midi/maxi) with proper validation
- [ ] Goal validation using existing schema validation logic
- [ ] New goals added to existing goals.yml file preserving existing goals

### Goal Management
- [ ] `iter goal list` displays existing goals in human-readable format
- [ ] `iter goal edit` allows selection and modification of existing goals
- [ ] `iter goal remove` removes goals with confirmation prompt
- [ ] All operations preserve goal IDs and maintain data integrity
- [ ] File operations are atomic (no partial writes)

### User Experience
- [ ] Intuitive prompts with help text and examples
- [ ] Input validation with clear error messages
- [ ] Consistent styling using existing lipgloss patterns
- [ ] Graceful handling of file permission errors
- [ ] Preview of goal definition before saving

### Technical Requirements
- [ ] Loosely coupled design - separate goal configuration logic from entry collection
- [ ] Reuse existing validation logic from models package
- [ ] Reuse existing parser and file operations
- [ ] Follow established UI patterns from entry collection
- [ ] Comprehensive error handling
- [ ] Unit tests for all UI components

---
## 3. Implementation Plan & Progress

**Overall Status:** `In Progress`

**Architecture Analysis:**

Based on investigation of existing codebase:

**Existing UI Patterns:**
- Uses `github.com/charmbracelet/huh` for form building
- Established patterns: `huh.NewForm()` with `huh.NewGroup()` containing form elements
- Form types used: `NewInput()`, `NewConfirm()`, `NewSelect()`, `NewText()`
- Styling with `lipgloss` for colors and formatting
- Error handling pattern: return descriptive errors, don't panic on form failures

**Existing Infrastructure:**
- `internal/parser/GoalParser` - handles loading/saving goals.yml
- `internal/models/Goal.Validate()` - schema validation logic
- `internal/models/` - complete goal type definitions and constants
- `cmd/` structure uses cobra for command organization
- ID persistence already implemented (T004) - generated IDs automatically saved

**Planned Implementation Approach:**

### Phase 1: Command Structure & Core UI
- [ ] **1.1 Add goal subcommand structure**
  - Create `cmd/goal.go` with `goal` parent command
  - Add `add`, `list`, `edit`, `remove` subcommands
  - Follow existing cobra patterns from `cmd/entry.go`

- [ ] **1.2 Create goal configuration UI package**
  - Create `internal/ui/goalconfig/` package for separation of concerns
  - Design `GoalConfigurator` struct similar to `EntryCollector`
  - Establish form builder patterns specific to goal configuration

### Phase 2: Goal Creation (iter goal add)
- [ ] **2.1 Basic goal creation flow**
  - Title and description input with validation
  - Goal type selection (simple/elastic/informational)
  - Field type selection with contextual guidance

- [ ] **2.2 Simple goal configuration**
  - Boolean field type setup
  - Scoring type selection (manual/automatic)
  - Basic criteria definition for automatic scoring

- [ ] **2.3 Elastic goal configuration**
  - Field type selection (numeric, duration, time, text)
  - Unit configuration for numeric fields
  - Mini/midi/maxi criteria definition with validation
  - Criteria ordering validation (reuse existing logic)

- [ ] **2.4 Informational goal configuration**
  - Field type and unit setup
  - Direction specification (higher_better/lower_better/neutral)

### Phase 3: Goal Management
- [ ] **3.1 Goal listing (iter goal list)**
  - Load and display existing goals
  - Formatted output with goal type, field type, scoring info
  - Optional filtering by goal type

- [ ] **3.2 Goal editing (iter goal edit)**
  - Goal selection from existing goals
  - Pre-populate forms with current values
  - Allow modification of all goal properties
  - Preserve goal ID for data integrity

- [ ] **3.3 Goal removal (iter goal remove)**
  - Goal selection interface
  - Confirmation prompt with goal details
  - Safe removal preserving other goals

### Phase 4: Integration & Polish
- [ ] **4.1 File operations**
  - Atomic goal additions/modifications using existing parser
  - Error handling for file permissions, disk space
  - Backup existing goals.yml before modifications

- [ ] **4.2 Validation integration**
  - Leverage existing `Goal.Validate()` and `Schema.Validate()`
  - Real-time validation during form input where possible
  - Clear error messages for validation failures

- [ ] **4.3 Testing & documentation**
  - Unit tests for form builders and validation
  - Integration tests for complete workflows
  - Update CLI help text and documentation

**Key Design Decisions:**

1. **Separation of Concerns**: Goal configuration UI in separate package (`internal/ui/goalconfig/`) to avoid coupling with entry collection
2. **Reuse Existing Infrastructure**: Leverage existing parser, validation, and UI patterns rather than duplicating
3. **Progressive Disclosure**: Guide users through goal creation with conditional prompts based on selections
4. **Data Integrity**: Preserve goal IDs and use atomic file operations
5. **Extensibility**: Design forms to easily accommodate new goal types and field types

## 4. Roadblocks

*(None identified yet)*

## 5. Notes / Discussion Log

- Created to provide user-friendly goal configuration without YAML editing
- Investigation shows strong existing foundation with huh forms, validation, and file operations
- Design follows established patterns from entry collection UI
- Focus on progressive disclosure and guided configuration experience

**Technical Notes:**
- `huh.NewSelect()` perfect for goal type and field type selection
- Existing validation in models package provides solid foundation
- T004 ID persistence ensures data integrity during goal modifications
- Cobra command structure established in `cmd/` package