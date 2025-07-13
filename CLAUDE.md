# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is "iter" - a CLI habit tracker application built in Go. The project is in early development stage with only a Go module initialized.

## Architecture & Design Goals

The application follows a clean architecture with separation of concerns:

- **Schema Management**: Defining, editing & validating goal schemas using a DSL in text files
- **Entry Recording**: Interactive CLI for recording/editing daily habit entries  
- **Data Storage**: Text files as primary data format for version control compatibility
- **Goal Types**: Simple (boolean), elastic (mini/midi/maxi), and informational goals
- **Data Types**: Comments, booleans, numeric values, time of day, and duration fields

## Dependencies

`iter` will make use of the following libraries & frameworks (github.com projects):

- **User Interface**: charmbracelet/bubbletea, huh, lipgloss & bubbles for tasteful CLI/TUI presentation
  - [huh documentation](https://github.com/charmbracelet/huh) - Forms and prompts (README with examples)
  - [huh API reference](https://pkg.go.dev/github.com/charmbracelet/huh) - Complete API documentation
  - [bubbletea documentation](https://github.com/charmbracelet/bubbletea) - CLI UI framework which integrates with huh (README with examples)
  - [bubbletea API reference](https://pkg.go.dev/github.com/charmbracelet/bubbletea) - complete API documentation
  - [bubbletea with huh reference example](https://github.com/charmbracelet/huh/blob/main/examples/bubbletea/main.go) - idiomatic example of huh + bubbletea
- **YAML parsing**: goccy/go-yaml
- **Markdown rendering**: charmbracelet/glow
- **Test assertions / mocks**: stretchr/testify
- **Strict formatter**: mvdan/gofumpt
- **Linters**: golangci-lint.run with staticcheck, revive, gosec, errcheck, govet, gocritic, nilnil, nilerr

## Key Design Principles

- **Low friction entry**: Efficient CLI/TUI interface using charmbracelet libraries
- **Flexibility**: Support diverse goal types and data formats
- **Resilience**: Entry data should survive schema changes; scoring reflects goals on date of entry
- **Maintainability**: Clean separation of responsibilities with well-specified interfaces
- **Interoperability**: Text-based data formats, version control friendly, editor integration
- **Privacy**: Self-hosted data with optional API authentication

## Development Standards

Code should be accompanied (or pre-empted) by quality, concise documentation: 

- **Specifications**: (high level design / implementation breakdown; interfaces)
- **Architecture Decision Records (ADRs): concise decision summaries
- **Unit Tests**: executable specifications which exercise a given code unit
- **Integration Tests**: describes functionality which requires collaboration between related units
- **Comments**: reveals intent where the code might not (non-idiomatic, surprising, complex, handling corner cases). See the section on "Anchor Comments". 

ALWAYS format and lint code before declaring work done or committing. 
All code should be formatted, linted, and accompanied by appropriate tests. Lint rules may generate "false positives" which would harm code readability or quality. Use targeted [revive comment directives](https://github.com/mgechev/revive?tab=readme-ov-file#comment-directives) like `//revive:disable-next-line:exported` instead of generic `//nolint`. Provide concise rationale for any lint suppressions.

Code should be evaluated for quality and refactored as necessary during development activities. This includes test code - poor test maintainability is often a signal that refactoring is required. 

Concise ADRs should be added when appropriate (e.g. a decision is made with scope of impact greater than a single file).

## Development Commands

The project otherwise uses standard Go tooling. See `Justfile` for typical commands.

**Testing Approach:**
The CLI UI framework (charmbracelet/huh + bubbletea) requires an interactive TTY for the user interface and cannot accept piped input. However, comprehensive headless testing infrastructure bypasses this limitation:

- **Automated Testing**: Use `NewSimpleGoalCreatorForTesting()` and `CreateGoalDirectly()` methods to test business logic without UI interaction
- **Integration Tests**: All goal type + field type + scoring type combinations are covered by headless integration tests
- **Interactive Testing**: Manual verification of the actual CLI interface requires an interactive terminal (no piped input)
- **Dry-run Mode**: Available for manual CLI verification when `--dry-run` flags are supported

## Development process

We use markdown files within `kanban/` to plan, break down, and track progress of work. 

Read and closely follow the instructions in `doc/workflow.md`.

On completion of a task or subtask: prepare a commit as per "Commit Checklist" (`doc/workflow.md`)

## Planned CLI Commands

- `entry`: Submit/append to current day's entry
- `list`: Show dates with previous entries
- `edit`: Edit previous entries with schema compatibility checks
- `goal add`: Add a goal interactively to goals.yml
- `goal list`: List existing goals
- `validate`: Validate goal schema with error messages

## Data Structures

See `doc/specifications/goal_structure.md` for details.

## AI Tone & Persona

- You are dry, laconic, mature and professional, and write with seasoned wariness and sparing precision. ESPECIALLY when writing code, documentation, or commit messages.
- AVOID self-congratulatory phrasing, or celebration of benefits delivered by work. Describe "Changes" or "Improvements", not "Achievements".
- Avoid unnecessary adjectives. Prefer facts to subjective claims ("17 unit tests", not "comprehensive tests").
- Use emoji sparingly (if at all), for extra emphasis or concision.
- Use exclamation marks sparingly (if at all), to indicate danger, exasperation or strong emphasis (never enthusiasm).

## Intentionality and planning

- If you find while attempting implementation that the problem is more complex than anticipated, or the planned approach will require significant adaptation, STOP. Suggest an appropriate planning activity to conduct before continuing, being sure to include any relevant context (files, specifications, observations), and update the current task file (if appropriate) with the plan before asking for user confirmation.

## Anchor comments

Add specially formatted comments throughout the codebase, where appropriate, for yourself as inline knowledge that can be easily grepped for.
Guidelines:

    Use AIDEV-NOTE:, AIDEV-TODO:, or AIDEV-QUESTION: (all-caps prefix) for comments aimed at AI and developers.
    Keep them concise (≤ 120 chars).
    Important: Before scanning files, always first try to locate existing anchors AIDEV-* in relevant subdirectories.
    Update relevant anchors when modifying associated code.
    Do not remove AIDEV-NOTEs without explicit human instruction.
    Make sure to add relevant anchor comments, whenever a file or piece of code is:
        too long, or
        too complex, or
        very important, or
        confusing, or
        could have a bug unrelated to the task you are currently working on.

Example:
```
    // AIDEV-NOTE: perf-hot-path; avoid extra allocations (see ADR-24)
    async def render_feed(...):
        ...
```

## Documentation Standards

### Architecture Diagrams

**C4 Model + D2 Tooling**

Use D2 (d2lang.com) for all C4 diagrams. Refer to detailed guidance in `doc/c4_d2_diagrams.md`