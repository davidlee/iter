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
- **Primary storage options**:
  - **YAML collection**: `$VICE_DATA/{context}/flotsam.yml` (structured metadata + body)
  - **Markdown files**: `$VICE_DATA/{context}/flotsam/*.md` (individual note files with frontmatter)
- **Repository integration**: Extend DataRepository interface for flotsam operations
- **Wiki link processing**: Extract [[links]] and compute backlinks within context boundaries
- **Search indexing**: Context-scoped search to maintain data isolation (file content or YAML body)
- **Optional enhancement**: Badger/skate for read performance with chosen storage as source of truth

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
- **zk** (github.com/zk-org/zk):
  - Go-based markdown note tool with LSP, fzf, multiple link styles
  - Uses standard markdown + YAML frontmatter
  - **Question**: Support zk conventions & interop (file locations, compatible IDs)?
  - **Decision needed**: Full compatibility, partial compatibility, or independent approach?
  - **Implications**: If we use zk's storage conventions, what happens when ENV vars change and referenced files move?

### Storage Strategy Decisions

#### Primary Storage Format
- **Leaning towards**: Markdown files + supplemental data store/cache for indexing
- **Options**:
  - Individual .md files with frontmatter (zk compatible)
  - YAML collection with embedded markdown bodies
  - Hybrid: .md files + separate index/metadata cache
- **Decision needed**: Choose primary storage format

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
