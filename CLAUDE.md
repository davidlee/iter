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
- **Comments**: reveals intent where the code might not (non-idiomatic, surprising, complex, handling corner cases).  

All code should be formatted, linted, and accompanied by appropriate tests. Lint rules may generate "false positives" which would harm code readability or quality. `//nolint` directives may be used sparingly, or exclusions added. These should be accompanied by concise comments with a rationale.

Code should be evaluated for quality and refactored as necessary during development activities. This includes test code - poor test maintainability is often a signal that refactoring is required. 

Concise ADRs should be added when appropriate (e.g. a decision is made with scope of impact greater than a single file).

## Development Commands

The project otherwise uses standard Go tooling:

```bash
# Build the application
go build

# Run tests
go test ./...

# Run specific test
go test -run TestName ./path/to/package

# Format code
gofumpt ./...

# Lint (requires golangci-lint)
golangci-lint run

# Run the application
go run main.go [subcommand]
```


## Development process

We use markdown files within `kanban/` to plan, break down, and track progress of work. Read and closely follow the instructions in `doc/workflow.md`.

## Planned CLI Commands

- `entry`: Submit/append to current day's entry
- `revise`: Edit current day's entry  
- `list`: Show dates with previous entries
- `edit`: Edit previous entries with schema compatibility checks
- `goals`: Display goal schema with color formatting
- `validate`: Validate goal schema with error messages

## Data Structure

- Daily entries can be partially complete and incrementally updated
- Entry fields are typed: comment (text), boolean, numeric (with units), time of day, duration
- Goals have unique identifiers for schema change resilience
- Schema defines goal types, criteria, and automatic scoring rules

See `doc/specifications/goal_structure.md` for details.


## Initial Scope Limitations

Out of scope for initial implementation:
- Schema builder (manual DSL editing)
- Visualization & analysis features
- API/MCP server
- Full-screen TUI interface
- Advanced editor integration
- Complex schema migration handling