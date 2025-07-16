---
title: "Flotsam Data Layer Implementation"
tags: ["data", "yaml", "models", "storage"]
related_tasks: ["part-of:T026"]
context_windows: ["internal/models/*.go", "internal/storage/*.go", "doc/specifications/*.md", "CLAUDE.md"]
---
# Flotsam Data Layer Implementation

**Context (Background)**:
Implement the core data layer for the flotsam note system, including Go structs, YAML persistence, and basic CRUD operations. This is the foundational component for T026 flotsam system.

**Type**: `feature`

**Overall Status:** `In Progress`

## Reference (Relevant Files / URLs)

### Significant Code (Files / Functions)
- `internal/models/habit.go` - Existing YAML model patterns
- `internal/models/entry.go` - Entry data structures
- `internal/storage/habits.go` - YAML persistence patterns
- `internal/storage/entries.go` - Storage layer operations
- `internal/storage/backup.go` - Backup and atomic operations

### Relevant Documentation
- `doc/specifications/habit_schema.md` - YAML schema patterns
- `doc/specifications/entries_storage.md` - Storage specifications
- `doc/architecture.md` - Data architecture section (4.1-4.4)

### Related Tasks / History
- **Parent Task**: T026 - Flotsam Note System (epic)
- T001-T025 - Established YAML persistence and model patterns

## Habit / User Story

As a developer implementing the flotsam system, I need a robust data layer that:
- Defines Go structs for flotsam notes with proper validation
- Persists data to YAML files following project conventions
- Provides CRUD operations with atomic writes and backups
- Supports wiki link extraction and backlink computation
- Handles ID generation and metadata tracking

## Acceptance Criteria (ACs)

- [ ] `internal/models/flotsam.go` with complete data structures
- [ ] `internal/storage/flotsam.go` with CRUD operations
- [ ] YAML schema validation and marshaling
- [ ] Wiki link extraction from markdown content
- [ ] Backlink computation and indexing
- [ ] ID generation using sqids or ulid
- [ ] Atomic file operations with backup
- [ ] Comprehensive unit tests for all operations
- [ ] Integration with existing storage patterns

## Architecture

### Data Structures
```go
type Flotsam struct {
    ID       string    `yaml:"id"`
    Title    string    `yaml:"title"`
    Body     string    `yaml:"body"`
    Created  time.Time `yaml:"created"`
    Modified time.Time `yaml:"modified"`
    Tags     []string  `yaml:"tags"`
    Links    []string  `yaml:"links"`
    Backlinks []string `yaml:"backlinks"`
    Metadata FlotsamMetadata `yaml:"metadata"`
    Type     FlotsamType `yaml:"type"`
}

type FlotsamMetadata struct {
    EditHistory []time.Time `yaml:"edit_history"`
    SRS         *SRSData    `yaml:"srs,omitempty"`
}

type SRSData struct {
    Score   float64   `yaml:"score"`
    Due     time.Time `yaml:"due"`
    Reviews int       `yaml:"reviews"`
}
```

### Storage Operations
- `CreateFlotsam(flotsam *Flotsam) error`
- `GetFlotsam(id string) (*Flotsam, error)`
- `UpdateFlotsam(flotsam *Flotsam) error`
- `DeleteFlotsam(id string) error`
- `ListFlotsam() ([]*Flotsam, error)`
- `SearchFlotsam(query string) ([]*Flotsam, error)`

## Implementation Plan & Progress

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

- [ ] **Data Model Definition**: Create flotsam data structures
  - [ ] **Define core Flotsam struct**: ID, title, body, timestamps, tags
    - *Design:* Follow existing model patterns from habit.go/entry.go
    - *Code/Artifacts:* `internal/models/flotsam.go`
    - *Testing Strategy:* Unit tests for struct validation and YAML marshaling
  - [ ] **Implement FlotsamMetadata**: Edit history, SRS data structures
    - *Design:* Nested structs with proper YAML tags and validation
    - *Code/Artifacts:* Extend `internal/models/flotsam.go`
    - *Testing Strategy:* Test metadata serialization and optional fields
  - [ ] **Add FlotsamType enum**: Support for idea/flashcard/script/log types
    - *Design:* String-based enum with validation
    - *Code/Artifacts:* Type definitions in `internal/models/flotsam.go`
    - *Testing Strategy:* Test type validation and defaults

- [ ] **Storage Layer Implementation**: YAML persistence and CRUD operations
  - [ ] **Create FlotsamStorage struct**: Following existing storage patterns
    - *Design:* Embed common storage functionality, file path management
    - *Code/Artifacts:* `internal/storage/flotsam.go`
    - *Testing Strategy:* Test file operations and error handling
  - [ ] **Implement CRUD operations**: Create, Read, Update, Delete
    - *Design:* Atomic operations with backup, following entries.go patterns
    - *Code/Artifacts:* CRUD methods in `internal/storage/flotsam.go`
    - *Testing Strategy:* Integration tests for file persistence
  - [ ] **Add search functionality**: Title and content search
    - *Design:* Simple string matching, future: consider indexing
    - *Code/Artifacts:* Search methods in storage layer
    - *Testing Strategy:* Test search accuracy and performance

- [ ] **Wiki Link Processing**: Extract links and compute backlinks
  - [ ] **Implement link extraction**: Parse [[wiki links]] from markdown
    - *Design:* Regex-based extraction, validate link targets
    - *Code/Artifacts:* Link processing functions
    - *Testing Strategy:* Test various link formats and edge cases
  - [ ] **Build backlink index**: Compute reverse links
    - *Design:* Maintain index of which notes link to each note
    - *Code/Artifacts:* Backlink computation and storage
    - *Testing Strategy:* Test backlink accuracy and updates

- [ ] **ID Generation and Utilities**: Short ID generation and helpers
  - [ ] **Implement ID generation**: Using sqids or ulid
    - *Design:* Research both options, choose based on requirements
    - *Code/Artifacts:* ID generation utilities
    - *Testing Strategy:* Test ID uniqueness and format
  - [ ] **Add validation helpers**: Struct validation and sanitization
    - *Design:* Input validation for user data
    - *Code/Artifacts:* Validation functions
    - *Testing Strategy:* Test validation rules and error cases

## Roadblocks

*(No roadblocks identified yet)*

## Notes / Discussion Log

- `2025-07-16 - AI:` Created child task for data layer implementation as part of T026 epic.

## Git Commit History

*No commits yet - task is in backlog*