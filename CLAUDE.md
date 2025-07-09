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

## Key Design Principles

- **Low friction entry**: Efficient CLI/TUI interface using charmbracelet libraries
- **Flexibility**: Support diverse goal types and data formats
- **Resilience**: Entry data should survive schema changes; scoring reflects goals on date of entry
- **Maintainability**: Clean separation of responsibilities with well-specified interfaces
- **Interoperability**: Text-based data formats, version control friendly, editor integration
- **Privacy**: Self-hosted data with optional API authentication

## Development Commands

The project uses standard Go tooling:

```bash
# Build the application
go build

# Run tests
go test ./...

# Run specific test
go test -run TestName ./path/to/package

# Format code
go fmt ./...

# Lint (requires golangci-lint)
golangci-lint run

# Run the application
go run main.go [subcommand]
```

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

## Initial Scope Limitations

Out of scope for initial implementation:
- Schema builder (manual DSL editing)
- Visualization & analysis features
- API/MCP server
- Full-screen TUI interface
- Advanced editor integration
- Complex schema migration handling