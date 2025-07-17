---
title: "Flotsam Data Layer Implementation"
tags: ["data", "markdown", "models", "storage", "zk-integration"]
related_tasks: ["part-of:T026", "depends-on:T028"]
context_windows: ["internal/models/*.go", "internal/storage/*.go", "doc/specifications/*.md", "CLAUDE.md"]
---
# Flotsam Data Layer Implementation

**Context (Background)**:
Implement the core data layer for the flotsam note system using individual markdown files with YAML frontmatter, ZK-compatible parsing, and go-srs SRS integration. This is the foundational component for T026 flotsam system.

**Type**: `feature`

**Overall Status:** `In Progress`

## Reference (Relevant Files / URLs)

This task is part of the `T026_flotsam_note_system` epic.

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
- Defines Go structs compatible with ZK frontmatter schema
- Integrates with T028 Repository Pattern for context-aware persistence
- Persists to `$VICE_DATA/{context}/flotsam/*.md` with YAML frontmatter
- Provides CRUD operations through DataRepository interface extension
- Supports ZK-compatible wiki link extraction and backlink computation
- Handles SRS scheduling using adapted go-srs SM-2 algorithm
- Maintains context isolation while supporting ZK interoperability

## Acceptance Criteria (ACs)

- [ ] `internal/models/flotsam.go` with ZK-compatible data structures
- [ ] Extend DataRepository interface for flotsam operations (T028 integration)
- [ ] `internal/repository/flotsam_repository.go` with markdown file operations
- [ ] ZK frontmatter parsing and validation (copied from ZK codebase)
- [ ] Context-scoped wiki link extraction using ZK parsing logic
- [ ] ZK-compatible ID generation (4-char alphanum, configurable)
- [ ] Integration with ViceEnv for markdown directory paths
- [ ] SM-2 SRS implementation using copied go-srs algorithm
- [ ] Individual markdown file operations with atomic safety
- [ ] Comprehensive unit tests for all operations
- [ ] Integration tests with context switching scenarios

## Architecture

### Data Structures (ZK-Compatible)
```go
// ZK-compatible frontmatter struct
type FlotsamFrontmatter struct {
    ID       string    `yaml:"id"`           // ZK 4-char alphanum ID
    Title    string    `yaml:"title"`        // ZK standard title field
    CreatedAt string   `yaml:"created-at"`   // ZK timestamp format
    Tags     []string  `yaml:"tags"`         // ZK tag array
    Type     string    `yaml:"type"`         // flotsam: idea|flashcard|script|log
    // SRS fields (flotsam extension)
    SRS      *SRSData  `yaml:"srs,omitempty"`
}

// In-memory representation with parsed data
type Flotsam struct {
    // Frontmatter fields
    FlotsamFrontmatter
    // Parsed content
    Body      string    // Markdown body content
    Modified  time.Time // File mtime
    Links     []string  // Extracted [[wikilinks]]
    Backlinks []string  // Computed reverse links
    FilePath  string    // Absolute file path
}

// SRS data matching go-srs schema
type SRSData struct {
    Easiness              float64   `yaml:"easiness"`              // 2.5 default
    ConsecutiveCorrect    int       `yaml:"consecutive_correct"`   // 0 default
    Due                   int64     `yaml:"due"`                   // Unix timestamp
    TotalReviews          int       `yaml:"total_reviews"`         // Review count
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

### Storage Strategy (Decision Made)

**Chosen Approach: Individual Markdown Files**
- **Directory**: `$VICE_DATA/{context}/flotsam/`
- **Structure**: One `.md` file per note with YAML frontmatter
- **Filename**: `{id}.md` (e.g., `6ub6.md`) following ZK convention
- **Format**: YAML frontmatter + markdown body

**Decision Rationale** (from T026 evaluation):
- **ZK Compatibility**: Supports external ZK tools and editors
- **Git-friendly**: Individual files enable proper version control
- **Editor Support**: Can be opened in any markdown editor
- **Extensibility**: Easy to add metadata without breaking existing files

**Implementation Details**:
- **Context isolation**: Each context has separate `/flotsam/` directory
- **Repository Pattern**: Leverage T028 FileRepository for atomic operations
- **Indexing**: Maintain in-memory index for links/backlinks per context
- **Caching**: Use lazy loading with file mtime-based cache invalidation

## Implementation Plan & Progress

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

### 1. External Code Integration (from T026.1)
- [ ] **1.1 Copy ZK Components**: Extract ZK parsing components for flotsam use
  - [ ] **1.1.1 Copy ZK frontmatter parsing**: Extract parsing logic from ZK codebase
    - *Source:* `/home/david/.local/src/zk/internal/core/note_parse.go`
    - *Target:* `internal/flotsam/zk_parser.go`
    - *Dependencies:* Also copy required utility functions from `internal/util/`
    - *Modifications:* Add package header, attribution comment, remove unused functions
    - *Testing:* Create basic test to verify frontmatter parsing works
  - [ ] **1.1.2 Copy ZK wikilink extraction**: Copy link processing logic
    - *Source:* `/home/david/.local/src/zk/internal/core/link.go`
    - *Target:* `internal/flotsam/zk_links.go`
    - *Dependencies:* May need markdown parsing utilities from `internal/adapter/markdown/`
    - *Modifications:* Adapt for context-scoped link resolution, add flotsam-specific logic
    - *Testing:* Test link extraction from markdown content
  - [ ] **1.1.3 Copy ZK ID generation**: Copy ID generation utilities
    - *Source:* `/home/david/.local/src/zk/internal/core/id.go`
    - *Target:* `internal/flotsam/zk_id.go`
    - *Dependencies:* Random generation utilities from `internal/util/rand/`
    - *Modifications:* Configure for flotsam defaults (4-char alphanum, lowercase)
    - *Testing:* Test ID generation uniqueness and format compliance
- [ ] **1.2 Copy Go-SRS Components**: Extract SM-2 algorithm for SRS functionality
  - [ ] **1.2.1 Copy SM-2 algorithm core**: Copy SuperMemo 2 implementation
    - *Source:* `/home/david/.local/src/go-srs/algo/sm2/sm2.go`
    - *Target:* `internal/flotsam/srs_sm2.go`
    - *Dependencies:* Review data structures from `review/review.go`
    - *Modifications:* Remove badgerdb dependencies, adapt for frontmatter storage
    - *Testing:* Test SM-2 calculations with known input/output pairs
  - [ ] **1.2.2 Copy SRS interfaces**: Copy algorithm and database interfaces
    - *Source:* `/home/david/.local/src/go-srs/algo/algo.go`, `/home/david/.local/src/go-srs/db/db.go`
    - *Target:* `internal/flotsam/srs_interfaces.go`
    - *Dependencies:* Core SRS types and review structures
    - *Modifications:* Adapt interfaces for flotsam markdown file storage
    - *Testing:* Test interface compliance with flotsam implementations
  - [ ] **1.2.3 Copy review data structures**: Copy review and item structures
    - *Source:* `/home/david/.local/src/go-srs/review/review.go`
    - *Target:* `internal/flotsam/srs_review.go`
    - *Dependencies:* Core algorithm types
    - *Modifications:* Adapt for flotsam note review workflow
    - *Testing:* Test review data serialization and validation
- [ ] **1.3 Integration and Attribution**: Prepare copied code for vice integration
  - [ ] **1.3.1 Add proper attribution**: Add copyright headers and attribution comments
    - *Task:* Add attribution headers to all copied files
    - *Format:* Standard Go copyright header with original project attribution
    - *Requirements:* Comply with original project licenses (ZK: GPLv3, go-srs: Apache-2.0)
  - [ ] **1.3.2 Resolve package dependencies**: Update imports and package declarations
    - *Task:* Change package names from `core`/`sm2` to `flotsam`
    - *Imports:* Update all import paths to vice project structure
    - *Conflicts:* Resolve any naming conflicts between ZK and go-srs components
  - [ ] **1.3.3 Create integration tests**: Test copied components work together
    - *Task:* Create integration tests for ZK + SRS components
    - *Tests:* Parse frontmatter → extract links → schedule review → update SRS data
    - *Coverage:* Test complete flotsam note lifecycle with copied components

### 2. Data Model Definition
- [ ] **2.1 Define ZK-Compatible Structures**: Create flotsam data structures
  - [ ] **2.1.1 Define FlotsamFrontmatter struct**: ZK-compatible YAML schema
    - *Design:* ZK standard fields (id, title, created-at, tags) + flotsam extensions (srs, type)
    - *Code/Artifacts:* `internal/models/flotsam.go`
    - *Testing:* Unit tests for struct validation and YAML marshaling
  - [ ] **2.1.2 Define in-memory Flotsam struct**: Parsed content representation
    - *Design:* Embed frontmatter + parsed content (body, links, backlinks, filepath)
    - *Code/Artifacts:* Extend `internal/models/flotsam.go`
    - *Testing:* Test struct embedding and content parsing
  - [ ] **2.1.3 Add SRS data structures**: go-srs compatible SRS metadata
    - *Design:* Match go-srs schema (easiness, consecutive_correct, due, total_reviews)
    - *Code/Artifacts:* SRS structs in `internal/models/flotsam.go`
    - *Testing:* Test SRS metadata serialization and optional fields
- [ ] **2.2 Add FlotsamType Support**: Support for different note types
  - [ ] **2.2.1 Add FlotsamType enum**: Support for idea/flashcard/script/log types
    - *Design:* String-based enum with validation and defaults
    - *Code/Artifacts:* Type definitions in `internal/models/flotsam.go`
    - *Testing:* Test type validation and defaults
  - [ ] **2.2.2 Add type-specific validation**: Validate content based on type
    - *Design:* Type-specific validation rules and content requirements
    - *Code/Artifacts:* Validation functions in `internal/models/flotsam.go`
    - *Testing:* Test type-specific validation rules

### 3. Repository Integration
- [ ] **3.1 Extend DataRepository Interface**: Add flotsam methods to T028 Repository Pattern
  - [ ] **3.1.1 Extend DataRepository interface**: Add flotsam methods to existing interface
    - *Design:* Context-aware methods following T028 patterns
    - *Code/Artifacts:* Update `internal/repository/repository.go`
    - *Testing:* Test interface compliance and context isolation
  - [ ] **3.1.2 Add flotsam method signatures**: Define CRUD operations for flotsam
    - *Design:* LoadFlotsam, SaveFlotsam, CreateNote, GetNote, UpdateNote, DeleteNote, SearchFlotsam
    - *Code/Artifacts:* Interface methods in `internal/repository/repository.go`
    - *Testing:* Test method signatures and parameter validation
- [ ] **3.2 Implement FileRepository Methods**: Add markdown file operations
  - [ ] **3.2.1 Implement LoadFlotsam**: Load all flotsam notes from context directory
    - *Design:* Scan `.md` files in context flotsam directory, parse frontmatter
    - *Code/Artifacts:* `LoadFlotsam` method in `internal/repository/file_repository.go`
    - *Testing:* Test loading from different contexts and empty directories
  - [ ] **3.2.2 Implement SaveFlotsam**: Save flotsam collection to markdown files
    - *Design:* Write individual `.md` files with frontmatter + body content
    - *Code/Artifacts:* `SaveFlotsam` method in `internal/repository/file_repository.go`
    - *Testing:* Test atomic operations and error handling
  - [ ] **3.2.3 Implement individual CRUD operations**: Create, read, update, delete single notes
    - *Design:* File-based operations with atomic safety using temp files
    - *Code/Artifacts:* CRUD methods in `internal/repository/file_repository.go`
    - *Testing:* Test individual operations and concurrent access
- [ ] **3.3 Add ViceEnv Path Support**: Context-aware directory path resolution
  - [ ] **3.3.1 Add GetFlotsamDir method**: Return context-aware flotsam directory path
    - *Design:* `GetFlotsamDir()` returns `$VICE_DATA/{context}/flotsam/`
    - *Code/Artifacts:* Update `internal/config/env.go`
    - *Testing:* Test path resolution for different contexts
  - [ ] **3.3.2 Add directory initialization**: Ensure flotsam directory exists
    - *Design:* Create flotsam directory during context initialization
    - *Code/Artifacts:* Update `EnsureContextFiles` in `internal/init/files.go`
    - *Testing:* Test directory creation and permissions

### 4. Core Operations Implementation
- [ ] **4.1 Implement Flotsam Parsing**: Use copied ZK components for parsing
  - [ ] **4.1.1 Implement frontmatter parsing**: Use copied ZK parser for YAML frontmatter
    - *Design:* Parse YAML frontmatter using ZK parsing logic
    - *Code/Artifacts:* Parsing functions using `internal/flotsam/zk_parser.go`
    - *Testing:* Test frontmatter parsing with various ZK-compatible formats
  - [ ] **4.1.2 Implement markdown body parsing**: Extract body content from markdown files
    - *Design:* Separate frontmatter from markdown body content
    - *Code/Artifacts:* Content extraction functions
    - *Testing:* Test body extraction and content preservation
- [ ] **4.2 Implement Link Processing**: Use copied ZK components for wikilink extraction
  - [ ] **4.2.1 Implement context-aware link extraction**: Parse [[wiki links]] within context boundaries
    - *Design:* Use ZK link extraction with context validation
    - *Code/Artifacts:* Link processing using `internal/flotsam/zk_links.go`
    - *Testing:* Test link resolution within and across context boundaries
  - [ ] **4.2.2 Build context-scoped backlink index**: Compute reverse links within context
    - *Design:* Maintain per-context index of which notes link to each note
    - *Code/Artifacts:* Backlink computation respecting context isolation
    - *Testing:* Test backlink accuracy and context isolation
- [ ] **4.3 Implement SRS Operations**: Use copied go-srs for review scheduling
  - [ ] **4.3.1 Implement SRS scheduling**: Quality-based review scheduling using SM-2
    - *Design:* Use copied SM-2 algorithm for spaced repetition scheduling
    - *Code/Artifacts:* SRS scheduling using `internal/flotsam/srs_sm2.go`
    - *Testing:* Test review scheduling and interval calculations
  - [ ] **4.3.2 Add SRS data persistence**: Store SRS data in frontmatter
    - *Design:* Serialize SRS data to YAML frontmatter fields
    - *Code/Artifacts:* SRS persistence functions
    - *Testing:* Test SRS data round-trip serialization
- [ ] **4.4 Add Validation and Utilities**: Comprehensive validation and helper functions
  - [ ] **4.4.1 Add struct validation**: Validate flotsam data structures
    - *Design:* Input validation for user data and frontmatter
    - *Code/Artifacts:* Validation functions in `internal/models/flotsam.go`
    - *Testing:* Test validation rules and error cases
  - [ ] **4.4.2 Add utility functions**: Helper functions for common operations
    - *Design:* ID generation, timestamp formatting, sanitization
    - *Code/Artifacts:* Utility functions in flotsam package
    - *Testing:* Test utility functions and edge cases

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
- `2025-07-17 - AI:` Updated task based on T026 evaluation phase decisions:
  - **Storage Decision**: Individual markdown files with YAML frontmatter (ZK-compatible)
  - **ZK Integration**: Copy parsing, linking, and ID generation components from ZK codebase
  - **SRS Integration**: Copy SM-2 algorithm from go-srs for spaced repetition
  - **Data Model**: ZK-compatible frontmatter with flotsam SRS extensions
  - **Initialization**: Follow vice's eager initialization pattern for flotsam directories

## Git Commit History

*No commits yet - task is in backlog*