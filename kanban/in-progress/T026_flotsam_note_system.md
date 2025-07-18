---
title: "Flotsam Note System"
tags: ["feature", "notes", "zettelkasten", "search"]
related_tasks: ["blocked-by:T041", "relates-to:T042,T043,T044,T045", "supersedes:T027"]
context_windows: ["internal/**/*.go", "CLAUDE.md", "doc/**/*.md", "kanban/**/*.md", "cmd/**/*.go"]
---
# Flotsam Note System

**Context (Background)**:
Implement a "flotsam" note system inspired by Notational Velocity, digital zettelkasten, markdown wikis, and spaced repetition systems. Notes "resurface" periodically and can be edited gradually over time, interlinked with wiki-style links, fuzzy searched, and attached to habits/entries.

**Type**: `feature`

**Overall Status:** `Blocked` - Awaiting T041 Unix Interop Foundation

## ARCHITECTURAL CHANGE NOTICE

**Unix Interop Decision (2025-07-18)**: After comprehensive analysis, vice is pivoting from T027's coupled integration to Unix interop patterns. See `doc/design-artefacts/unix-interop-vs-coupled-integration-analysis.md` for full rationale.

**Impact on T026**:
- **Handled by ZK Integration**: External editor, fuzzy search, wiki links, templates, note management
- **Remains Relevant**: SRS scheduling, flotsam-specific workflows, vice integration
- **New Foundation**: T041 establishes Unix interop foundation and SRS database

**Revised Scope**: T026 now focuses on SRS workflows and flotsam-specific UX built on top of zk integration, rather than implementing note management from scratch.

## Reference (Relevant Files / URLs)

### Design Documentation
- `doc/design-artefacts/unix-interop-vs-coupled-integration-analysis.md` - Unix interop decision analysis
- `kanban/backlog/T041_unix_interop_foundation.md` - Foundation implementation task

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
- `doc/guidance/bubbletea_guide.md` - UI development guidelines

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

- [ ] vice provides a fuzzy search interface for flotsam notes
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
- **Question**: zk is written in Go - what additional components can we import/reuse?
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


## Implementation Plan & Progress

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

### 1. Data Layer Foundation (T027)
- [x] **1 Data Layer Implementation**: Complete data layer foundation in T027

### 2. Core CLI Commands

- Create vice flotsam command**: Main command entry point
  - [ ]  Create vice flotsam command**: Main command entry point
  - [ ]  Add subcommand routing: list, new, edit, search, review
    - `new`: given a string, create a new flotsam file with that title.
    - `list`: list flotsam id, title 
    - `edit`: edit flotsam in $EDITOR
    - `search`: 

  - [ ]  Implement Context awareness**: Use current vice context



```
### 2. Core CLI Commands
- [ ] **3.1 Flotsam Command Structure**: Base command and subcommands
  - [ ] **3.1.1 Create vice flotsam command**: Main command entry point
  - [ ] **3.1.2 Add subcommand routing**: list, new, edit, search, review
  - [ ] **3.1.3 Implement context awareness**: Use current vice context
- [ ] **3.2 Basic CRUD Operations**: Create, read, update, delete notes
  - [ ] **3.2.1 Implement flotsam new**: Create new notes with templates
  - [ ] **3.2.2 Implement flotsam edit**: Edit existing notes in $EDITOR
  - [ ] **3.2.3 Implement flotsam list**: List notes with filtering
  - [ ] **3.2.4 Implement flotsam remove**: Delete notes safely

### 3. Search and Navigation
- [ ] **4.1 Fuzzy Search Interface**: Interactive note discovery
  - [ ] **4.1.1 Implement title/content search**: Fuzzy matching on title and body
  - [ ] **4.1.2 Add tag-based filtering**: Search by tags and metadata
  - [ ] **4.1.3 Implement interactive selection**: fzf-style interface
- [ ] **4.2 Wiki Link System**: Link resolution and backlinks
  - [ ] **4.2.1 Implement link extraction**: Parse [[wikilinks]] from content
  - [ ] **4.2.2 Build backlink index**: Compute reverse link relationships
  - [ ] **4.2.3 Add link navigation**: Jump between linked notes

### 4. Spaced Repetition System
- [ ] **5.1 Review Interface**: SRS-based note review
  - [ ] **5.1.1 Implement flotsam review**: Show due notes for review
  - [ ] **5.1.2 Add quality rating**: 0-6 scale quality feedback
  - [ ] **5.1.3 Update SRS scheduling**: SM-2 algorithm integration
- [ ] **5.2 Flashcard Support**: Specialized review for flashcard notes
  - [ ] **5.2.1 Add flashcard templates**: Front/back card structure
  - [ ] **5.2.2 Implement flashcard review**: Specialized review interface
  - [ ] **5.2.3 Add performance tracking**: Review statistics and trends

### 5. UI and User Experience
- [ ] **6.1 Interactive Interface**: Bubbletea-based TUI
  - [ ] **6.1.1 Create main flotsam view**: Three-pane interface (search, list, preview)
  - [ ] **6.1.2 Implement note editing**: In-app editing with live preview
  - [ ] **6.1.3 Add keyboard shortcuts**: Efficient navigation and actions
- [ ] **6.2 Integration with Vice**: Seamless ecosystem integration
  - [ ] **6.2.1 Link to habit entries**: Attach flotsam to habit data
  - [ ] **6.2.2 Add flotsam references**: Reference notes from other vice commands
  - [ ] **6.2.3 Implement cross-context**: Optional cross-context note access

### 6. Advanced Features
- [ ] **7.1 Templates and Automation**: Note creation templates
  - [ ] **7.1.1 Implement template system**: ZK-compatible templates
  - [ ] **7.1.2 Add automation hooks**: Trigger creation from other commands
  - [ ] **7.1.3 Support custom templates**: User-defined note templates
- [ ] **7.2 Export and Interoperability**: Data exchange capabilities
  - [ ] **7.2.1 Implement export functions**: Various format export
  - [ ] **7.2.2 Add import capabilities**: Import from other note systems
  - [ ] **7.2.3 Support external editors**: Integration with external tools
```


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

### Implementation Progress - Data Layer

- `2025-07-17 - AI:` **T027 Data Layer Progress**:
  - **External Code Integration**: ZK parsing and go-srs components copied and adapted
  - **Data Model Definition**: ZK-compatible structures with SRS support in progress
  - **Repository Integration**: Extending T028 DataRepository interface for flotsam operations
  - **Core Operations**: Building on copied components for parsing, linking, SRS scheduling
  - **Status**: See T027 task card for detailed implementation progress

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