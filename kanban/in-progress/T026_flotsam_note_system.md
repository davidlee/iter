---
title: "Flotsam Note System"
tags: ["feature", "notes", "zettelkasten", "search"]
related_tasks: ["blocks:T027", "depends-on:T028"]
context_windows: ["internal/**/*.go", "CLAUDE.md", "doc/**/*.md", "kanban/**/*.md", "cmd/**/*.go"]
---
# Flotsam Note System

**Context (Background)**:
Implement a "flotsam" note system inspired by Notational Velocity, digital zettelkasten, markdown wikis, and spaced repetition systems. Notes "resurface" periodically and can be edited gradually over time, interlinked with wiki-style links, fuzzy searched, and attached to habits/entries.

**Type**: `feature`

**Overall Status:** `In Progress`

## Reference (Relevant Files / URLs)

### Significant Code (Files / Functions)
- `internal/models/` - Data model definitions for YAML schema
- `internal/storage/` - YAML file persistence layer
- `internal/ui/` - User interface components using bubbletea
- `cmd/` - CLI command structure
- `doc/specifications/` - Schema and storage specifications

### Relevant Documentation
- `doc/architecture.md` - Core architecture patterns and UI framework guidance
- `doc/specifications/habit_schema.md` - YAML schema patterns for data modeling
- `doc/specifications/entries_storage.md` - File storage patterns
- `doc/specifications/file_paths_runtime_env.md` - Context-aware file system and Repository Pattern (T028)
- `doc/bubbletea_guide.md` - UI development guidelines

### Related Tasks / History
- **Child Task**: T027 - Flotsam Data Layer Implementation
- **Dependency**: T028 - File Paths & Runtime Environment (Repository Pattern, context-aware storage)
- Previous storage and UI patterns established in T001-T025
- YAML-based data persistence patterns from habit/entry system
- Bubbletea UI patterns from entry collection system

## Habit / User Story

As a user, I want to create and manage floating notes (flotsam) that:
- Surface periodically for review and editing
- Support fuzzy search by title and content
- Allow wiki-style interlinking with backlinks
- Can be attached to habits and entries
- Support spaced repetition for learning
- Enable gradual development of ideas over time

This supports reflective practice, knowledge management, and incremental learning alongside habit tracking.

## Acceptance Criteria (ACs)

- [ ] `vice flotsam` command launches fuzzy search interface
- [ ] Create new flotsam with title and markdown body
- [ ] Edit existing flotsam with live preview
- [ ] Fuzzy search by title and content
- [ ] Support for tags and metadata
- [ ] Wiki-style [[link]] syntax with backlink indexing
- [ ] Spaced repetition scheduling for flashcards
- [ ] Attachment to habits and entries
- [ ] Data persistence in YAML format
- [ ] Support for different types (idea, flashcard, future: script, log)

## Architecture

### Data Model (flotsam.yml)
```yaml
flotsam:
  - id: "abc123"  # short ID (sqids or ulid)
    title: "Note Title"
    body: |
      Markdown content
      with [[wiki links]]
    created: "2024-01-01T10:00:00Z"
    modified: "2024-01-02T15:30:00Z"
    tags: ["idea", "literature-note", "namespaced:tag"]
    links: ["def456", "ghi789"]  # extracted from [[links]]
    backlinks: ["xyz111"]  # computed from other notes
    metadata:
      edit_history: ["2024-01-01T10:00:00Z", "2024-01-02T15:30:00Z"]
      srs:
        score: 2.5
        due: "2024-01-05T10:00:00Z"
        reviews: 3
    type: "idea"  # idea | flashcard | script | log
```

### UI Architecture
- Full-width fuzzy search input (top)
- Left pane: matching titles list
- Right pane: selected note body with markdown rendering
- Modal for editing with live preview

### Storage Strategy
- **Context-aware persistence**: Leverage T028 Repository Pattern for context isolation
- **Primary storage**: `$VICE_DATA/{context}/flotsam/*.md` (individual markdown files with YAML frontmatter)
- **Repository integration**: Extend DataRepository interface for flotsam operations
- **Wiki link processing**: Extract [[links]] and compute backlinks within context boundaries
- **Search indexing**: Context-scoped search to maintain data isolation (file content + frontmatter)
- **Cache/index store**: Badger/skate for computed metadata (backlinks, tags, SRS) with .md files as source of truth

## Scope Questions & Design Decisions

### External Dependencies & Integration Options

#### Spaced Repetition System (SRS)
- **go-srs** (github.com/revelaction/go-srs):
  - Uses SuperMemo 2 algorithm with pluggable interfaces (Algorithm, Database, UID)
  - Uses badger + ulid for storage/IDs
  - **Question**: Adapt go-srs to work with skate/vice storage? Benefits vs complexity?
  - **Decision needed**: Use go-srs directly, adapt interfaces, or implement our own SRS?

#### Key-Value Store / Caching
- **skate** (github.com/charmbracelet/skate):
  - Simple personal key-value store with badger backend
  - CLI-based with multiple database support
  - **Question**: Use skate for tag/link indexing and SRS metadata cache?
  - **Decision needed**: Direct badger, skate wrapper, or pure file-based storage?

#### Zettelkasten Compatibility
- **zk** (github.com/zk-org/zk) - Detailed Analysis:
  - **Storage**: Flexible .md files with optional YAML frontmatter, no strict structure
  - **IDs**: Optional, configurable (alphanum/hex, configurable length), used in filenames not content
  - **Links**: `[[title]]` or `[[filename]]` resolution, dynamic backlink computation
  - **Config**: TOML-based, uses `ZK_NOTEBOOK_DIR` env var for vault location
  - **CLI**: fzf-powered search, LSP integration, template system with Handlebars

**ZK Compatibility Decision Points:**
1. **Frontmatter schema**: Support zk's YAML fields (`title`, `date`, `tags`, `aliases`)
2. **Link syntax**: Use `[[wikilinks]]` with title/filename resolution
3. **File naming**: Support zk's template patterns (e.g., `{{id}}-{{slug title}}.md`)
4. **Directory structure**: Allow zk notebook directories as flotsam storage locations
5. **Environment integration**: Respect `ZK_NOTEBOOK_DIR` for interop vs vice context isolation

### Storage Strategy Decisions

#### Primary Storage Format
- **Decided**: Individual .md files with YAML frontmatter + supplemental data store/cache for indexing
- **Storage structure**: `$VICE_DATA/{context}/flotsam/*.md` with frontmatter metadata
- **Cache/index**: Separate data store for computed data (backlinks, tags, SRS metadata)

#### ID Generation Scheme
- **Options**:
  - ZK-compatible IDs (if pursuing interop)
  - ULID (what go-srs uses)
  - sqids (original plan)
- **Questions**: 
  - If ZK compatibility: what's zk's ID scheme and generation process?
  - Do we need to generate IDs or can we reuse existing zk database/index?

### Search & UI Implementation

#### Fuzzy Search Implementation
- **Options**:
  - Shell out to fzf (like zk does)
  - Use Go fuzzy search library
  - Hybrid: fzf for title search, custom for tag/link search
- **Question**: What about tag-based or link-based search? Does zk provide utility libraries we can import?

#### Editor Integration
- **Question**: How to handle opening .md files in $EDITOR from CLI/TUI?
- **Options**: Shell out to $EDITOR, embedded editor, or delegate to external tools

#### ZK Go Dependencies
- **Question**: zk is written in Go - what components can we import/reuse?
- **Candidates**: Markdown parsing, tag/link extraction, CLI patterns, config management

### Content & Templating

#### Markdown Templating
- **Question**: What templating do we need for .md files, if any?
- **Options**: None, simple templates, zk-compatible templates

### Performance & Memory Management

#### Large Note Collections
- **Questions**:
  - Memory concerns loading large .md folders into RAM?
  - Do we need to for certain features?
  - Naive approach vs lazy/JIT loading?
- **Implications**: Affects search indexing, link resolution, and SRS scheduling

#### Context Isolation vs ZK Interop
- **Question**: How to handle zk env var for vault path vs vice's context system?
- **Risk**: Referenced flotsam from habits becomes inaccessible when ENV changes
- **Options**: Copy files, symlinks, or abstraction layer

### ZK Compatibility Evaluation Steps

**Immediate Investigation Tasks:**
1. **Test ZK setup**: Install zk, create sample notebook, understand actual file structure
2. **Analyze zk Go modules**: Examine zk's source for reusable components (parsing, linking, templates)
3. **Frontmatter compatibility**: Map zk's YAML schema to flotsam requirements
4. **Link resolution**: Test zk's `[[wikilink]]` behavior with different filename patterns
5. **Template system**: Evaluate if zk's Handlebars templates could work for flotsam creation

**Compatibility Level Options:**
- **Full compatibility**: Flotsam works as zk notebook, zk commands work on flotsam files
- **Read compatibility**: Flotsam can import/read existing zk notebooks  
- **Write compatibility**: Flotsam creates zk-compatible files but may have additional metadata
- **Independent**: Learn from zk patterns but maintain vice-specific approach


**AIDEV-NOTE**: `zk/` is a symlink to the ZK source; it's also installed locally. User has a notebook at `~/workbench/zk`.

ZK Schema Architecture (SQLite):

```
┌─────────────────────────────────────────────────────────────┐
│                        NOTES                                │
├─────────────────────────────────────────────────────────────┤
│ id                PK  INTEGER  AUTOINCREMENT               │
│ path              U   TEXT     /path/to/note.md            │ 
│ sortable_path         TEXT     normalized sorting key      │
│ title                 TEXT     extracted/frontmatter       │
│ lead                  TEXT     first paragraph excerpt     │
│ body                  TEXT     main content                │
│ raw_content           TEXT     original markdown           │
│ word_count            INTEGER  content length metric       │
│ checksum              TEXT     content change detection    │
│ metadata              TEXT     JSON blob (v3+)             │
│ created               DATETIME timestamp                   │
│ modified              DATETIME timestamp                   │
└─────────────────────────────────────────────────────────────┘
             │
             │ 1:N
             ▼
┌─────────────────────────────────────────────────────────────┐
│                        LINKS                                │
├─────────────────────────────────────────────────────────────┤
│ id                PK  INTEGER  AUTOINCREMENT               │
│ source_id         FK  INTEGER  → notes(id) CASCADE         │
│ target_id         FK  INTEGER  → notes(id) SET NULL        │
│ title                 TEXT     link display text           │
│ href                  TEXT     original link target        │
│ external              INTEGER  boolean flag                │
│ rels                  TEXT     relationship types          │
│ snippet               TEXT     surrounding context         │
│ snippet_start         INTEGER  context start offset (v3+)  │
│ snippet_end           INTEGER  context end offset (v3+)    │
│ type                  TEXT     link classification (v5+)   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    COLLECTIONS                              │
├─────────────────────────────────────────────────────────────┤
│ id                PK  INTEGER  AUTOINCREMENT               │
│ kind              U   TEXT     'tag','group','type'        │
│ name              U   TEXT     collection identifier       │
└─────────────────────────────────────────────────────────────┘
             │
             │ N:M
             ▼
┌─────────────────────────────────────────────────────────────┐
│                NOTES_COLLECTIONS                            │
├─────────────────────────────────────────────────────────────┤
│ id                PK  INTEGER  AUTOINCREMENT               │
│ note_id           FK  INTEGER  → notes(id) CASCADE         │
│ collection_id     FK  INTEGER  → collections(id) CASCADE   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                     METADATA                                │
├─────────────────────────────────────────────────────────────┤
│ key               PK  TEXT     config/setting key          │
│ value                 TEXT     JSON/string value           │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   NOTES_FTS (Virtual)                       │
├─────────────────────────────────────────────────────────────┤
│ rowid             →   notes.id content linkage             │
│ path                  TEXT     indexed for search          │
│ title                 TEXT     indexed for search          │
│ body                  TEXT     indexed for search          │
└─────────────────────────────────────────────────────────────┘
```

**VIEWS:**
- `notes_with_metadata`: Notes + aggregated tags (GROUP_CONCAT)
- `resolved_links`: Links + source/target note paths & titles

**INDEXES:**
- `index_notes_checksum`: Fast content change detection
- `index_notes_path`: Unique path constraint + lookup optimization  
- `index_links_source_id_target_id`: Link relationship queries
- `index_collections`: Collection lookup by kind+name
- `index_notes_collections`: N:M association queries

**TRIGGERS (FTS Sync):**
- `trigger_notes_ai`: INSERT → update FTS index
- `trigger_notes_ad`: DELETE → remove from FTS index  
- `trigger_notes_au`: UPDATE → delete old + insert new FTS entry

**FEATURES:**
- **FTS5 Search**: Porter stemming, Unicode normalization, custom tokenizers
- **Referential Integrity**: CASCADE deletes, SET NULL for broken links
- **Versioned Schema**: 6 migration levels with reindexing support
- **JSON Metadata**: Extensible note properties in metadata column
- **Link Context**: Snippet extraction with precise offset tracking

## Implementation Plan & Progress

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

### 1. External Code Integration
- [ ] **1.1 Copy ZK Components**: Extract and prepare ZK parsing components
  - [ ] **1.1.1 Copy ZK frontmatter parsing**: Copy `internal/core/note_parse.go` and dependencies
    - *Source:* `/home/david/.local/src/zk/internal/core/note_parse.go`
    - *Target:* `internal/flotsam/zk_parser.go`
    - *Dependencies:* Also copy required utility functions from `internal/util/`
    - *Modifications:* Add package header, attribution comment, remove unused functions
    - *Testing:* Create basic test to verify frontmatter parsing works
  - [ ] **1.1.2 Copy ZK wikilink extraction**: Copy `internal/core/link.go` and link processing
    - *Source:* `/home/david/.local/src/zk/internal/core/link.go`
    - *Target:* `internal/flotsam/zk_links.go`
    - *Dependencies:* May need markdown parsing utilities from `internal/adapter/markdown/`
    - *Modifications:* Adapt for context-scoped link resolution, add flotsam-specific logic
    - *Testing:* Test link extraction from markdown content
  - [ ] **1.1.3 Copy ZK ID generation**: Copy `internal/core/id.go` and ID utilities
    - *Source:* `/home/david/.local/src/zk/internal/core/id.go`
    - *Target:* `internal/flotsam/zk_id.go`
    - *Dependencies:* Random generation utilities from `internal/util/rand/`
    - *Modifications:* Configure for flotsam defaults (4-char alphanum, lowercase)
    - *Testing:* Test ID generation uniqueness and format compliance
  - [ ] **1.1.4 Copy ZK template system**: Copy handlebars template engine
    - *Source:* `/home/david/.local/src/zk/internal/adapter/handlebars/`
    - *Target:* `internal/flotsam/zk_templates.go`
    - *Dependencies:* Handlebars library and helper functions
    - *Modifications:* Adapt for flotsam note creation templates
    - *Testing:* Test template rendering with flotsam data
- [ ] **1.2 Copy Go-SRS Components**: Extract and prepare SM-2 algorithm
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

### 2. Data Layer Foundation (T027)
- [ ] **2.1 Data Model Definition**: Create ZK-compatible flotsam structures
  - [ ] **2.1.1 Define FlotsamFrontmatter struct**: ZK-compatible YAML schema
  - [ ] **2.1.2 Define in-memory Flotsam struct**: Parsed content representation
  - [ ] **2.1.3 Add SRS data structures**: go-srs compatible SRS metadata
- [ ] **2.2 Repository Integration**: Extend T028 Repository Pattern
  - [ ] **2.2.1 Extend DataRepository interface**: Add flotsam methods
  - [ ] **2.2.2 Implement markdown file operations**: Individual .md file CRUD
  - [ ] **2.2.3 Add ViceEnv path methods**: Context-aware directory paths
- [ ] **2.3 Core Operations**: Build on copied components
  - [ ] **2.3.1 Implement flotsam parsing**: Use copied ZK parser for frontmatter
  - [ ] **2.3.2 Implement link processing**: Use copied ZK links for wikilink extraction
  - [ ] **2.3.3 Implement SRS operations**: Use copied go-srs for review scheduling
  - [ ] **2.3.4 Add validation helpers**: Struct validation and sanitization

### 3. Core CLI Commands
- [ ] **3.1 Flotsam Command Structure**: Base command and subcommands
  - [ ] **3.1.1 Create vice flotsam command**: Main command entry point
  - [ ] **3.1.2 Add subcommand routing**: list, new, edit, search, review
  - [ ] **3.1.3 Implement context awareness**: Use current vice context
- [ ] **3.2 Basic CRUD Operations**: Create, read, update, delete notes
  - [ ] **3.2.1 Implement flotsam new**: Create new notes with templates
  - [ ] **3.2.2 Implement flotsam edit**: Edit existing notes in $EDITOR
  - [ ] **3.2.3 Implement flotsam list**: List notes with filtering
  - [ ] **3.2.4 Implement flotsam remove**: Delete notes safely

### 4. Search and Navigation
- [ ] **4.1 Fuzzy Search Interface**: Interactive note discovery
  - [ ] **4.1.1 Implement title/content search**: Fuzzy matching on title and body
  - [ ] **4.1.2 Add tag-based filtering**: Search by tags and metadata
  - [ ] **4.1.3 Implement interactive selection**: fzf-style interface
- [ ] **4.2 Wiki Link System**: Link resolution and backlinks
  - [ ] **4.2.1 Implement link extraction**: Parse [[wikilinks]] from content
  - [ ] **4.2.2 Build backlink index**: Compute reverse link relationships
  - [ ] **4.2.3 Add link navigation**: Jump between linked notes

### 5. Spaced Repetition System
- [ ] **5.1 Review Interface**: SRS-based note review
  - [ ] **5.1.1 Implement flotsam review**: Show due notes for review
  - [ ] **5.1.2 Add quality rating**: 0-6 scale quality feedback
  - [ ] **5.1.3 Update SRS scheduling**: SM-2 algorithm integration
- [ ] **5.2 Flashcard Support**: Specialized review for flashcard notes
  - [ ] **5.2.1 Add flashcard templates**: Front/back card structure
  - [ ] **5.2.2 Implement flashcard review**: Specialized review interface
  - [ ] **5.2.3 Add performance tracking**: Review statistics and trends

### 6. UI and User Experience
- [ ] **6.1 Interactive Interface**: Bubbletea-based TUI
  - [ ] **6.1.1 Create main flotsam view**: Three-pane interface (search, list, preview)
  - [ ] **6.1.2 Implement note editing**: In-app editing with live preview
  - [ ] **6.1.3 Add keyboard shortcuts**: Efficient navigation and actions
- [ ] **6.2 Integration with Vice**: Seamless ecosystem integration
  - [ ] **6.2.1 Link to habit entries**: Attach flotsam to habit data
  - [ ] **6.2.2 Add flotsam references**: Reference notes from other vice commands
  - [ ] **6.2.3 Implement cross-context**: Optional cross-context note access

### 7. Advanced Features
- [ ] **7.1 Templates and Automation**: Note creation templates
  - [ ] **7.1.1 Implement template system**: ZK-compatible templates
  - [ ] **7.1.2 Add automation hooks**: Trigger creation from other commands
  - [ ] **7.1.3 Support custom templates**: User-defined note templates
- [ ] **7.2 Export and Interoperability**: Data exchange capabilities
  - [ ] **7.2.1 Implement export functions**: Various format export
  - [ ] **7.2.2 Add import capabilities**: Import from other note systems
  - [ ] **7.2.3 Support external editors**: Integration with external tools

## Roadblocks

*(No roadblocks identified yet)*

## Notes / Discussion Log

- `2025-07-16 - User:` Initial feature request with detailed data model and UI specifications
- `2025-07-16 - AI:` Created task card with architecture outline. Ready for implementation planning after user review.
- `2025-07-17 - AI:` Updated task dependencies and architecture based on T028 completion:
  - Added dependency on T028 (file paths & runtime environment)
  - Updated storage strategy to leverage Repository Pattern and context-aware file system
  - Added storage options: YAML collection OR individual markdown files with frontmatter
  - Repository interface will be extended to support flotsam operations
  - Wiki links and search will respect context boundaries for proper data isolation
  - T027 will evaluate both storage approaches for implementation
- `2025-07-17 - AI:` Added comprehensive scope questions and design decisions section:
  - Researched external dependencies: go-srs (SRS), skate (KV store), zk (zettelkasten)
  - Identified key decision points around storage format, ID schemes, search implementation
  - Highlighted tension between vice's context isolation and zk interoperability
  - Added questions about performance, templating, and editor integration
- `2025-07-17 - AI:` Simplified storage strategy (removed YAML option) and deep-dived ZK compatibility:
  - Decided on .md files with frontmatter + cache/index approach
  - Detailed ZK technical analysis: storage conventions, ID schemes, link handling, CLI patterns
  - Identified 5 concrete ZK compatibility decision points and evaluation steps
  - Added compatibility level options from full interop to independent approach

### Evaluation Phase - ZK Compatibility Analysis

- `2025-07-17 - AI:` **ZK Investigation Complete**:
  - **File Structure**: ZK uses 4-char alphanumeric IDs with flat directory structure (`6ub6.md`, `jgtt.md`)
  - **Frontmatter**: Minimal schema (`id`, `title`, `created-at`, `tags`) with extensible metadata map
  - **Wikilinks**: Standard `[[target]]` and `[[target | label]]` with relationship detection (`#[[uplink]]`, `[[downlink]]#`)
  - **ID Generation**: Configurable charset/length/case with cryptographically random generation
  - **Templates**: Handlebars-based with rich helper system for note creation
  - **Go Components**: High reusability potential - parser, link resolution, ID generation, template engine all importable

- `2025-07-17 - AI:` **Key Compatibility Insights**:
  - **High Compatibility**: Frontmatter schema extensible, wikilink syntax standard, ID generation adaptable
  - **Reusable Components**: Markdown parser, link resolution algorithms, template system, frontmatter handling
  - **Design Tension**: ZK's flat structure vs flotsam's potential hierarchical needs; SQLite indexing vs simpler approaches
  - **Integration Strategy**: Import core parsing/linking components while adapting storage layer for vice ecosystem

- `2025-07-17 - AI:` **Go-SRS Analysis Complete**:
  - **Architecture**: Clean interfaces (`db.Handler`, `algo.Algo`, `uid.UID`) with loose coupling
  - **Storage**: Simple schema (easiness, consecutive_correct, due_timestamp) stored as JSON in BadgerDB
  - **SM-2 Algorithm**: Mature, well-tested implementation with proper SuperMemo research backing
  - **Integration Complexity**: LOW-MEDIUM - only 3 methods in `db.Handler` interface to implement
  - **Data Model**: `{deck_id, card_id, easiness: 2.5, consecutive_correct: 0, due: timestamp}`
  - **Quality Scale**: Research-based 0-6 scale (0=no review, 1-3=incorrect, 4-6=correct)
  - **Recommended Approach**: Adapt go-srs with SQLite `db.Handler` implementation
  - **Integration Benefits**: Proven algorithm, clean abstractions, time savings vs reimplementation

### Code Reuse Strategy

- `2025-07-17 - AI:` **ZK Code Reuse Constraints**:
  - **Cannot import directly**: ZK's useful code is in `internal/` packages (Go prohibits external imports)
  - **Application module**: Would pull entire CLI application with all dependencies
  - **Recommended approach**: Copy specific code (parsing, linking) with attribution
  - **Target files**: `internal/core/note_parse.go`, `internal/core/link.go`, ID generation, templates

- `2025-07-17 - AI:` **Go-SRS Code Reuse Options**:
  - **Can import directly**: Public API design (`algo/`, `db/`, `uid/` packages)
  - **Library module**: Intended for external consumption, clean interfaces
  - **Dependency concern**: Would pull BadgerDB when only SM-2 algorithm needed
  - **Recommended approach**: Copy SM-2 algorithm (`algo/sm2/`) to avoid heavyweight dependencies

- `2025-07-17 - AI:` **Implementation Plan Structure**:
  - **Hierarchical numbering**: 7 major phases with 2-3 subsections each
  - **External integration first**: Section 1 focuses on copying and preparing external code
  - **Systematic progression**: Data layer → CLI → Search → SRS → UI → Advanced features
  - **Detailed specifications**: Each subtask includes source paths, targets, modifications, testing

## Git Commit History

- `5c12264` - docs(tasks)[T026]: add flotsam note system task card

## original crappy user notes
<!--
Inspired by the app "notational velocity", digital zettelkasten, markdown wikis, spaced repetition and incremental writing.

Flotsam are notes which "resurface" periodically and can be edited gradually over time, interlinked (markdown wiki links, backlinks), fuzzy searched, attached to habits / entries.

---
data model (flotsam.yml)

Flotsam have:
- a short ID (https://sqids.org/go or https://github.com/oklog/ulid)
- a title (text)
- a body (text/markdown multiline)
- a date created
- a date (last) modified
- tags[]
  eg. URL, literature-note, idea, 'namespoced:tag'
- links[],
- backlinks[]
  roll up from markdown wiki link syntax for indexing
- metadata:
  - edit history (array of timestamps)
  - array of events for spaced repetition (todo: identify best algorithm & hence data reqs .. assume https://github.com/revelaction/go-srs ?)
  - SRS score / position / whatever go-srs wants
- type: idea | flashcard [future: | script | log ]
  - is it a learning / flashcard thing, or an idea being massaged over time? * how does this affect usage patterns / behaviour / SRS algo?
  - (?) is it a script (plain text with a shebang)? maybe this is what we build a plugin / hook / shell snippet thing on top of
  - a log is a series of timestamped entries - a flowtime or pomodoro log, daily / weekly / monthly note; etc
    - any additional data required?

---

UI:
top, full width text input for fuzzy search / title
left pane: list of matching titles
right pane: selected / best match record's body (markdown)

---

affinity (later ..)
- attach to an entry
- attach to a habit
- attach to a habit data field(??)
- attach to a checklist
- attach to a checklist item?

same flotsam can be attached to multiple things, at least from a data model pov (a field on the other thing's yaml pointing to flotsam ID)

^ tbd whether there's enough reason to allow an array, lets assume 0-1. Once links exist and are navigable we could fake 1-many

---

- maybe: use badger / skate for reads, yaml for source of truth (ingest > db)

-
-->