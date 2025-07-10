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

- [ ] User can define simple boolean goals in a goals.yml file with XDG-compliant default location
- [ ] CLI supports --config-dir flag to override default config location for testing
- [ ] User can run `iter entry` command to record today's habit completion
- [ ] UI uses charmbracelet libraries for a polished CLI experience
- [ ] Entries are stored in entries.yml with proper structure and validation
- [ ] Code follows project standards (formatted, linted, tested)
- [ ] Basic error handling for invalid goals or missing files
- [ ] Project includes necessary dependencies (bubbletea, huh, lipgloss, testify, etc.)

---
## 3. Implementation Plan & Progress

**Overall Status:** `Not Started`

**Sub-tasks:**

- [ ] **Project Setup & Dependencies**: Setup Go modules and required libraries
    - [ ] **Add required dependencies:** Add charmbracelet libraries, goccy/go-yaml, testify
        - *Design:* Update go.mod with bubbletea, huh, lipgloss, bubbles, goccy/go-yaml, testify
        - *Code/Artifacts to be created or modified:* `go.mod`, `go.sum`
        - *Testing Strategy:* Verify dependencies resolve correctly with `go mod tidy`
        - *AI Notes:* Follow CLAUDE.md specifications for exact library versions
    - [ ] **Setup linting and formatting:** Configure golangci-lint and gofumpt
        - *Design:* Create .golangci.yml with staticcheck, revive, gosec, errcheck, govet, gocritic, nilnil, nilerr
        - *Code/Artifacts to be created or modified:* `.golangci.yml`, potentially Makefile or scripts
        - *Testing Strategy:* Run golangci-lint and gofumpt on sample code
        - *AI Notes:* May need to adjust linting rules as code develops

- [ ] **Configuration Management**: Implement XDG-compliant config paths with CLI override
    - [ ] **XDG path resolution:** Implement XDG Base Directory specification support
        - *Design:* Function to resolve ~/.config/iter/ as default, support XDG_CONFIG_HOME
        - *Code/Artifacts to be created or modified:* `internal/config/paths.go`
        - *Testing Strategy:* Unit tests for path resolution with various XDG env vars
        - *AI Notes:* Should gracefully handle missing directories
    - [ ] **CLI flag support:** Add --config-dir flag for custom config location
        - *Design:* Use cobra or flag package for CLI parsing, override default paths
        - *Code/Artifacts to be created or modified:* `cmd/root.go`, `cmd/entry.go`
        - *Testing Strategy:* Test CLI flag parsing and path override functionality
        - *AI Notes:* Consider using cobra for future CLI extension

- [ ] **Goal Parser & Validation**: Parse simple boolean goals from goals.yml
    - [ ] **Goal structure definition:** Define Go structs for simple boolean goals
        - *Design:* Goal struct with ID, Name, Type fields; GoalSet for collection
        - *Code/Artifacts to be created or modified:* `internal/models/goal.go`
        - *Testing Strategy:* Unit tests for goal struct validation
        - *AI Notes:* Design should be extensible for future goal types
    - [ ] **YAML parsing:** Implement goals.yml parsing with validation
        - *Design:* Use goccy/go-yaml, validate required fields, handle parse errors
        - *Code/Artifacts to be created or modified:* `internal/parser/goals.go`
        - *Testing Strategy:* Unit tests with valid/invalid YAML examples
        - *AI Notes:* Should provide clear error messages for invalid YAML

- [ ] **Entry Management**: Implement entry collection and storage
    - [ ] **Entry data model:** Define entry structure for boolean goal completion
        - *Design:* Entry struct with Date, GoalID, Value fields; EntrySet for collection
        - *Code/Artifacts to be created or modified:* `internal/models/entry.go`
        - *Testing Strategy:* Unit tests for entry validation and serialization
        - *AI Notes:* Consider partial entry support for future incremental updates
    - [ ] **Entry storage:** Implement entries.yml read/write with validation
        - *Design:* YAML serialization, atomic writes, backup on corruption
        - *Code/Artifacts to be created or modified:* `internal/storage/entries.go`
        - *Testing Strategy:* Unit tests for concurrent access, corruption handling
        - *AI Notes:* Should preserve existing entries when adding new ones

- [ ] **CLI Interface**: Create polished CLI using charmbracelet libraries
    - [ ] **Entry collection UI:** Build interactive UI for today's entry
        - *Design:* Use huh for form inputs, bubbletea for app flow, lipgloss for styling
        - *Code/Artifacts to be created or modified:* `internal/ui/entry.go`
        - *Testing Strategy:* Manual testing of UI flow, unit tests for business logic
        - *AI Notes:* Should handle keyboard navigation and validation gracefully
    - [ ] **CLI command structure:** Implement entry subcommand with proper help
        - *Design:* Main command with entry subcommand, help text, error handling
        - *Code/Artifacts to be created or modified:* `cmd/entry.go`, `main.go`
        - *Testing Strategy:* Test command parsing, help output, error scenarios
        - *AI Notes:* Structure should support future subcommands (revise, list, etc.)

- [ ] **Integration & Testing**: Ensure end-to-end functionality works correctly
    - [ ] **End-to-end testing:** Test complete workflow from goals.yml to entries.yml
        - *Design:* Create test scenarios with sample goals and entries
        - *Code/Artifacts to be created or modified:* `integration_test.go` or similar
        - *Testing Strategy:* Full workflow testing with temporary directories
        - *AI Notes:* Should test both happy path and error scenarios
    - [ ] **Code quality:** Ensure formatting, linting, and test coverage
        - *Design:* Run gofumpt, golangci-lint, go test with coverage
        - *Code/Artifacts to be created or modified:* Any code quality fixes needed
        - *Testing Strategy:* Automated checks pass, reasonable test coverage
        - *AI Notes:* May need to adjust linting rules or add nolint directives

## 4. Roadblocks

*(None identified yet)*

## 5. Notes / Discussion Log

- `2025-01-10 - User:` Requested epic task card for minimal end-to-end release following workflow.md format
- `2025-01-10 - AI:` Created comprehensive task breakdown focusing on simple boolean goals as starting point, designed for extensibility to future goal types

## 6. Code Snippets & Artifacts

*(AI will place generated code blocks here during implementation)*