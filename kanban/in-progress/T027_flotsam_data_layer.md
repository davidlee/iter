---
title: "Flotsam Data Layer Implementation"
tags: ["data", "yaml", "models", "storage"]
related_tasks: ["part-of:T026", "depends-on:T028"]
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
- `internal/repository/file_repository.go` - Repository Pattern implementation (T028)
- `internal/config/env.go` - ViceEnv and context-aware paths (T028)

### Relevant Documentation
- `doc/specifications/habit_schema.md` - YAML schema patterns
- `doc/specifications/entries_storage.md` - Storage specifications
- `doc/specifications/file_paths_runtime_env.md` - Repository Pattern and context-aware storage (T028)
- `doc/architecture.md` - Data architecture section (4.1-4.4)

### Related Tasks / History
- **Parent Task**: T026 - Flotsam Note System (epic)
- **Dependency**: T028 - File Paths & Runtime Environment (Repository Pattern foundation)
- T001-T025 - Established YAML persistence and model patterns

## Habit / User Story

As a developer implementing the flotsam system, I need a robust data layer that:
- Defines Go structs for flotsam notes with proper validation
- Integrates with T028 Repository Pattern for context-aware persistence
- Persists data to `$VICE_DATA/{context}/flotsam.yml` following ViceEnv conventions
- Provides CRUD operations through DataRepository interface extension
- Supports wiki link extraction and backlink computation within context boundaries
- Handles ID generation and metadata tracking with context isolation

## Acceptance Criteria (ACs)

- [ ] `internal/models/flotsam.go` with complete data structures
- [ ] Extend DataRepository interface for flotsam operations (T028 integration)
- [ ] `internal/repository/flotsam_repository.go` with context-aware CRUD operations
- [ ] YAML schema validation and marshaling
- [ ] Context-scoped wiki link extraction and backlink computation
- [ ] ID generation using sqids or ulid with context awareness
- [ ] Integration with ViceEnv for context-specific file paths
- [ ] Atomic file operations leveraging T028 Repository Pattern
- [ ] Comprehensive unit tests for all operations
- [ ] Integration tests with context switching scenarios

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

### Repository Integration (T028)
Extend DataRepository interface for flotsam operations:
```go
type DataRepository interface {
    // Existing methods from T028
    LoadHabits(ctx string) (*models.Schema, error)
    LoadEntries(ctx string, date time.Time) (*models.EntryLog, error)
    SaveEntries(ctx string, entries *models.EntryLog) error
    LoadChecklists(ctx string) (*models.ChecklistSchema, error)
    SwitchContext(newContext string) error
    
    // New flotsam methods
    LoadFlotsam(ctx string) (*FlotsamCollection, error)
    SaveFlotsam(ctx string, flotsam *FlotsamCollection) error
    CreateFlotsamNote(ctx string, flotsam *Flotsam) error
    GetFlotsamNote(ctx string, id string) (*Flotsam, error)
    UpdateFlotsamNote(ctx string, flotsam *Flotsam) error
    DeleteFlotsamNote(ctx string, id string) error
    SearchFlotsam(ctx string, query string) ([]*Flotsam, error)
}
```

### Storage Strategy Options
Two storage approaches to evaluate:

**Option A: YAML Collection** 
- **File**: `$VICE_DATA/{context}/flotsam.yml`
- **Structure**: Single YAML file with array of flotsam objects
- **Benefits**: Atomic operations, existing YAML patterns, structured metadata
- **Drawbacks**: Large file growth, less git-friendly

**Option B: Individual Markdown Files**
- **Directory**: `$VICE_DATA/{context}/flotsam/`
- **Structure**: One `.md` file per note with YAML frontmatter
- **Benefits**: Git-friendly, external editor support, natural markdown handling
- **Drawbacks**: More complex atomic operations, directory management

**Common Elements**:
- **Context isolation**: Both approaches respect context boundaries via ViceEnv
- **Repository Pattern**: Leverage existing T028 FileRepository infrastructure
- **Atomic operations**: Follow "turn off and on again" pattern for context switching

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

- [ ] **Storage Strategy Decision**: Evaluate and choose between YAML vs Markdown files
  - [ ] **Research and prototype both approaches**: Compare YAML collection vs individual markdown files
    - *Design:* Build small prototypes of both storage methods
    - *Code/Artifacts:* Proof-of-concept implementations
    - *Testing Strategy:* Benchmark performance, git-friendliness, developer experience
  - [ ] **Choose storage approach**: Based on prototype evaluation and user requirements
    - *Design:* Document decision rationale and trade-offs
    - *Code/Artifacts:* ADR (Architecture Decision Record) if needed
    - *Testing Strategy:* Validate chosen approach meets all requirements

- [ ] **Repository Integration**: Extend T028 Repository Pattern for flotsam
  - [ ] **Extend DataRepository interface**: Add flotsam methods to existing interface
    - *Design:* Context-aware methods following T028 patterns
    - *Code/Artifacts:* Update `internal/repository/repository.go`
    - *Testing Strategy:* Test interface compliance and context isolation
  - [ ] **Implement FlotsambMethods in FileRepository**: Add to existing FileRepository
    - *Design:* Use chosen storage approach (YAML file or markdown directory)
    - *Code/Artifacts:* Extend `internal/repository/file_repository.go`
    - *Testing Strategy:* Integration tests with context switching
  - [ ] **Add ViceEnv path methods**: Context-aware file/directory path resolution
    - *Design:* GetFlotsamFile() for YAML OR GetFlotsamDir() for markdown approach
    - *Code/Artifacts:* Update `internal/config/env.go`
    - *Testing Strategy:* Test path resolution for different contexts

- [ ] **Wiki Link Processing**: Context-scoped link extraction and backlink computation
  - [ ] **Implement context-aware link extraction**: Parse [[wiki links]] within context boundaries
    - *Design:* Regex-based extraction, validate link targets exist in current context
    - *Code/Artifacts:* Link processing functions with context parameter
    - *Testing Strategy:* Test link resolution across context boundaries (should fail)
  - [ ] **Build context-scoped backlink index**: Compute reverse links within context
    - *Design:* Maintain per-context index of which notes link to each note
    - *Code/Artifacts:* Backlink computation respecting context isolation
    - *Testing Strategy:* Test backlink accuracy and context isolation

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
- `2025-07-17 - AI:` Updated task architecture and dependencies based on T028 completion:
  - Added dependency on T028 (file paths & runtime environment)
  - Updated architecture to leverage Repository Pattern and ViceEnv for context isolation
  - Added storage strategy evaluation: YAML collection vs individual markdown files
  - Extended DataRepository interface design for flotsam operations
  - Updated wiki link processing to respect context boundaries
  - Modified implementation plan to integrate with existing T028 infrastructure
  - Added storage decision as first implementation step to choose optimal approach

## Git Commit History

*No commits yet - task is in backlog*