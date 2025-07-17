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

*Implementation plan pending scope decisions above*

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