# Project Overview

Vice is a CLI/TUI app. It provides powerful capabilities for attention and knowledge management
by composing, orchestrating and extending a number of powerful existing tools. 

## Design Goals

- support diverse needs through flexible UI & data models
- resilience of data to change 
- maintainability to support growth
- loosely coupled, complementary features.

## UI Tests

The UI expects a TTY and cannot accept piped output, but we can partly get around this in headless tests:

- **Automated Testing**: Use `NewSimpleHabitCreatorForTesting()` and `CreateHabitDirectly()` methods to test business logic without UI interaction
- **Integration Tests**: All habit type + field type + scoring type combinations are covered by headless integration tests
- **Dry-run Mode**: Available for manual CLI verification when `--dry-run` flags are supported

## Linter

Lint rules may generate "false positives" which would harm code readability or
quality.

If you believe you need to add a `nolint` directive, or avoid fixing all lint
errors before committing, you MUST:
- explain to the User why you think it is appropriate
- accompany each directive comment with a concise rationale
- NEVER use blanket `//nolint` directives - use targeted directives with only necessary scope
- NEVER modify project-wide linter configuration at `.golangci.yml`

See: [revive comment directives](https://github.com/mgechev/revive?tab=readme-ov-file#comment-directives)

## Serena

IMPORTANT: Remember to use the Serena MCP tool where appropriate:
- Use the Serena MCP Tool for file / edit operations when:
  - Exploring codebases
  - Applying regex operations, especially over multiple files
  - Symbol navigation and reference tracking
  - Understanding project architecture
  - Code refactoring with pattern-based changes
- Use Built-in Tools When:
  - Creating new files from scratch
  - Simple string replacements
  - File is in hot cache and there is no risk of edits causing syntax errors

The Serena MCP server should be running - otherwise STOP and report it.
