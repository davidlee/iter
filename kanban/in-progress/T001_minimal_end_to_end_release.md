---
title: "Minimal End-to-End Release with Simple Boolean Goals"
type: ["feature"]
tags: ["epic", "mvp", "cli", "parser", "ui"]
related_tasks: []
context_windows: ["./CLAUDE.md", "./go.mod", "./doc/specifications/goal_structure.md", "./*.go", "./cmd/*.go", "./internal/**/*.go"]
---

# Minimal End-to-End Release with Simple Boolean Goals

## 1. Goal / User Story

As a user, I want to track simple boolean habits (did/didn't do) using a CLI tool so that I can start building a habit tracking routine with minimal friction. This epic establishes the core foundation for the iter habit tracker by implementing the essential components needed for a working MVP.

The system should allow me to:
- Define simple boolean goals in a goals.yml file
- Run a CLI command to record today's entry for those goals
- Store entries in a structured format that can grow with future features
- Use XDG-compliant paths for configuration while supporting custom paths for testing

This task is important because it establishes the architectural foundation and core user workflow that all future features will build upon.

## 2. Acceptance Criteria

- [x] User can define simple boolean goals in a goals.yml file with XDG-compliant default location
- [x] CLI supports --config-dir flag to override default config location for testing
- [x] User can run `iter entry` command to record today's habit completion
- [x] UI uses charmbracelet libraries for a polished CLI experience
- [x] Entries are stored in entries.yml with proper structure and validation
- [x] Code follows project standards (formatted, linted, tested)
- [x] Basic error handling for invalid goals or missing files
- [x] Project includes necessary dependencies (bubbletea, huh, lipgloss, testify, etc.)

---
## 3. Implementation Plan & Progress

**Overall Status:** `Completed`

**Sub-tasks:**

- [x] **1. Project Setup & Dependencies**: Setup Go modules and required libraries
    - [x] **1.1 Add required dependencies:** Add charmbracelet libraries, goccy/go-yaml, testify
        - *Design:* Update go.mod with bubbletea, huh, lipgloss, bubbles, goccy/go-yaml, testify
        - *Code/Artifacts to be created or modified:* `go.mod`, `go.sum`
        - *Testing Strategy:* Verify dependencies resolve correctly with `go mod tidy`
        - *AI Notes:* Follow CLAUDE.md specifications for exact library versions
    - [x] **1.2 Setup linting and formatting:** Configure golangci-lint and gofumpt
        - *Design:* Create .golangci.yml with staticcheck, revive, gosec, errcheck, govet, gocritic, nilnil, nilerr
        - *Code/Artifacts to be created or modified:* `.golangci.yml`, potentially Makefile or scripts
        - *Testing Strategy:* Run golangci-lint and gofumpt on sample code
        - *AI Notes:* May need to adjust linting rules as code develops

- [x] **2. Configuration Management**: Implement XDG-compliant config paths with CLI override
    - [x] **2.1 XDG path resolution:** Implement XDG Base Directory specification support
        - *Design:* Function to resolve ~/.config/iter/ as default, support XDG_CONFIG_HOME
        - *Code/Artifacts to be created or modified:* `internal/config/paths.go`
        - *Testing Strategy:* Unit tests for path resolution with various XDG env vars
        - *AI Notes:* Should gracefully handle missing directories
    - [x] **2.2 CLI flag support:** Add --config-dir flag for custom config location
        - *Design:* Use cobra or flag package for CLI parsing, override default paths
        - *Code/Artifacts to be created or modified:* `cmd/root.go`, `cmd/entry.go`
        - *Testing Strategy:* Test CLI flag parsing and path override functionality
        - *AI Notes:* Consider using cobra for future CLI extension

- [x] **3. Goal Parser & Validation**: Parse simple boolean goals from goals.yml
    - [x] **3.1 Goal structure definition:** Define Go structs for simple boolean goals
        - *Design:* Goal struct with ID, Name, Type fields; GoalSet for collection
        - *Code/Artifacts to be created or modified:* `internal/models/goal.go`
        - *Testing Strategy:* Unit tests for goal struct validation
        - *AI Notes:* Design should be extensible for future goal types
    - [x] **3.2 YAML parsing:** Implement goals.yml parsing with validation
        - *Design:* Use goccy/go-yaml, validate required fields, handle parse errors
        - *Code/Artifacts to be created or modified:* `internal/parser/goals.go`
        - *Testing Strategy:* Unit tests with valid/invalid YAML examples
        - *AI Notes:* Should provide clear error messages for invalid YAML

- [x] **4. Entry Management**: Implement entry collection and storage
    - [x] **4.1 Entry data model:** Define entry structure for boolean goal completion
        - *Design:* Entry struct with Date, GoalID, Value fields; EntrySet for collection
        - *Code/Artifacts to be created or modified:* `internal/models/entry.go`
        - *Testing Strategy:* Unit tests for entry validation and serialization
        - *AI Notes:* Consider partial entry support for future incremental updates
    - [x] **4.2 Entry storage:** Implement entries.yml read/write with validation
        - *Design:* YAML serialization, atomic writes, backup on corruption
        - *Code/Artifacts to be created or modified:* `internal/storage/entries.go`
        - *Testing Strategy:* Unit tests for concurrent access, corruption handling
        - *AI Notes:* Should preserve existing entries when adding new ones

- [x] **5. CLI Interface**: Create polished CLI using charmbracelet libraries
    - [x] **5.1 Entry collection UI:** Build interactive UI for today's entry
        - *Design:* Use huh for form inputs, bubbletea for app flow, lipgloss for styling
        - *Code/Artifacts to be created or modified:* `internal/ui/entry.go`
        - *Testing Strategy:* Manual testing of UI flow, unit tests for business logic
        - *AI Notes:* Should handle keyboard navigation and validation gracefully
    - [x] **5.2 CLI command structure:** Implement entry subcommand with proper help
        - *Design:* Main command with entry subcommand, help text, error handling
        - *Code/Artifacts to be created or modified:* `cmd/entry.go`, `main.go`
        - *Testing Strategy:* Test command parsing, help output, error scenarios
        - *AI Notes:* Completed - CLI structure implemented with cobra, proper help, and error handling
    - [x] **5.3 File initialization:** Create sample files when goals.yml or entries.yml missing
        - *Design:* Check for file existence, create minimal sample goals and empty entries structure
        - *Code/Artifacts to be created or modified:* `cmd/entry.go`, potentially new initialization module
        - *Testing Strategy:* Test file creation, sample content validation, edge cases
        - *AI Notes:* Should create user-friendly sample goals and proper YAML structure

- [x] **6. Integration & Testing**: Ensure end-to-end functionality works correctly
    - [x] **6.1 End-to-end testing:** Test complete workflow from goals.yml to entries.yml
        - *Design:* Create test scenarios with sample goals and entries
        - *Code/Artifacts to be created or modified:* Manual testing with temporary directories
        - *Testing Strategy:* Full workflow testing with temporary directories
        - *AI Notes:* Completed - tested file initialization, CLI help, and UI startup flow
    - [x] **6.2 Code quality:** Ensure formatting, linting, and test coverage
        - *Design:* Run gofumpt, golangci-lint, go test with coverage
        - *Code/Artifacts to be created or modified:* Code quality verified
        - *Testing Strategy:* Automated checks pass, reasonable test coverage
        - *AI Notes:* Completed - all tests pass, no linting issues, proper formatting

## 4. Roadblocks

*(None identified yet)*

## 5. Notes / Discussion Log

- `2025-01-10 - User:` Requested epic task card for minimal end-to-end release following workflow.md format
- `2025-01-10 - AI:` Created comprehensive task breakdown focusing on simple boolean goals as starting point, designed for extensibility to future goal types
- `2025-01-10 - AI:` Subtask 1.1 completed - Added all required dependencies (bubbletea, huh, lipgloss, bubbles, goccy/go-yaml, testify). Created main.go with blank imports to preserve dependencies after go mod tidy.
- `2025-01-10 - AI:` Subtask 1.2 completed - Setup .golangci.yml with staticcheck, revive, gosec, errcheck, govet, gocritic, nilnil, nilerr, ineffassign, unused, misspell. Installed gofumpt and golangci-lint tools. Both run successfully on current code.
- `2025-01-10 - AI:` Subtask 2.1 completed - Implemented XDG Base Directory specification support in internal/config/paths.go. Created Paths struct with GetDefaultPaths() and GetPathsWithConfigDir() functions. Added comprehensive unit tests covering XDG_CONFIG_HOME scenarios. All tests pass and linter reports 0 issues.
- `2025-01-10 - AI:` Subtask 2.2 completed - Added --config-dir CLI flag support using cobra. Created cmd/root.go with persistent flag handling and cmd/entry.go with placeholder entry command. CLI properly resolves paths from flag or XDG defaults, creates config directories, and includes comprehensive unit tests. All tests pass and linter reports 0 issues.
- `2025-01-10 - AI:` Subtask 3.1 completed - Defined comprehensive Go structs for goal schema in internal/models/goal.go. Implemented Schema, Goal, FieldType, Criteria, and Condition structs following doc/specifications/goal_schema.md closely. Added validation methods, ID generation, and extensible design for future goal types. Created 35 unit tests covering all validation scenarios. All tests pass and linter reports 0 issues.
- `2025-01-10 - AI:` Subtask 3.2 completed - Implemented comprehensive YAML parsing in internal/parser/goals.go using goccy/go-yaml. Created GoalParser with LoadFromFile, SaveToFile, ParseYAML methods and validation. Added helper functions for goal lookup and filtering. Fixed schema validation to properly generate IDs. Created 28 unit tests with comprehensive error handling. All tests pass and linter reports 0 issues.
- `2025-01-10 - AI:` Subtask 4.1 completed - Defined comprehensive entry data structures in internal/models/entry.go. Implemented EntryLog, DayEntry, and GoalEntry structs with full CRUD operations. Added validation, date handling, boolean value helpers, and range queries. Designed for partial entry support and future incremental updates. Created 48 unit tests covering all functionality. All tests pass and linter reports 0 issues.
- `2025-01-10 - AI:` Subtask 4.2 completed - Implemented comprehensive entry storage in internal/storage/entries.go. Created EntryStorage with atomic writes, YAML parsing, validation, and backup functionality. Added high-level operations for day/goal entry management with automatic file creation. Includes date range queries and today-specific helpers. Created 37 unit tests covering all storage operations and error scenarios. All tests pass and linter reports 0 issues.
- `2025-01-10 - AI:` Subtask 5.1 completed - Built interactive UI for today's entry collection in internal/ui/entry.go. Created EntryCollector with charmbracelet libraries: huh for forms, lipgloss for styling. Features include welcome/completion messages, goal-by-goal entry collection, optional notes, existing entry loading/updating, and comprehensive error handling. Updated cmd/entry.go to integrate UI with CLI. Created comprehensive unit tests covering all functionality. All tests pass and linter reports 0 issues.
- `2025-01-10 - AI:` Subtask 5.3 completed - Added file initialization feature in internal/init/files.go. Created FileInitializer that checks for missing goals.yml/entries.yml and creates user-friendly samples. Sample goals: "Morning Exercise" and "Daily Reading" with descriptions, prompts, and help text. Empty entries.yml with proper YAML structure. Updated cmd/entry.go to call initializer before UI. Created comprehensive unit tests covering all scenarios. All tests pass and linter reports 0 issues.
- `2025-07-11 - AI:` Subtask 5.2 completed - CLI command structure was already fully implemented with cobra. Root command supports --config-dir flag, entry subcommand has proper help text and error handling.
- `2025-07-11 - AI:` Subtask 6.1 completed - End-to-end testing performed with temporary directories. Verified file initialization, CLI help functionality, and UI startup flow work correctly.
- `2025-07-11 - AI:` Subtask 6.2 completed - Code quality verified: all tests pass (148+ unit tests), gofumpt reports no formatting issues, golangci-lint reports 0 issues.
- `2025-07-11 - AI:` **T001 COMPLETE** - All acceptance criteria met. Working CLI habit tracker with boolean goals, XDG-compliant paths, polished UI using charmbracelet libraries, proper file storage, comprehensive error handling, and full test coverage.

## 6. Code Snippets & Artifacts

*(AI will place generated code blocks here during implementation)*