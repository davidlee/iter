---
title: "Flotsam Note System"
tags: ["feature", "notes", "zettelkasten", "srs", "advanced"]
related_tasks: ["depends-on:T041", "spawned:T047", "relates-to:T042,T043,T044,T045", "supersedes:T027"]
context_windows: ["internal/**/*.go", "CLAUDE.md", "doc/**/*.md", "kanban/**/*.md", "cmd/**/*.go"]
---
# Flotsam Note System

**Context (Background)**:
Implement a "flotsam" note system inspired by Notational Velocity, digital zettelkasten, markdown wikis, and spaced repetition systems. Notes "resurface" periodically and can be edited gradually over time, interlinked with wiki-style links, fuzzy searched, and attached to habits/entries.

**Type**: `feature`

**Overall Status:** `Ready` - T041 Unix Interop Foundation Complete

## FOUNDATION STATUS UPDATE

**T041 Completion (2025-07-19)**: Unix interop foundation successfully implemented and delivered:

✅ **Delivered by T041**:
- ZK integration with auto-init (external editor, fuzzy search, wiki links, templates)
- SRS database foundation with SQLite schema
- Basic CLI commands: `vice flotsam add/list/due/edit`
- Unix interop patterns and graceful degradation
- Directory auto-creation and ZK notebook initialization

**T026 Revised Scope**: Focus on advanced SRS workflows, content change detection, and enhanced flotsam-specific UX built on the solid T041 foundation.

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

✅ **COMPLETED by T041**:
- [x] Create new flotsam with title and markdown body (`vice flotsam add`)
- [x] Edit existing flotsam with external editor (`vice flotsam edit`)
- [x] Fuzzy search by title and content (ZK `--interactive` integration)
- [x] Support for tags and metadata (ZK-compatible frontmatter)
- [x] Wiki-style [[link]] syntax with backlink indexing (ZK delegation)
- [x] Basic spaced repetition scheduling (SRS database with SM-2)
- [x] Data persistence in markdown format (individual .md files)
- [x] Support for different types (idea, flashcard, script, log via `vice:type:*` tags)

**REMAINING for T026**:
- [ ] Advanced SRS quality assessment with content change detection
- [ ] Context-level git integration for audit trails
- [ ] Enhanced SRS review workflows and statistics
- [ ] Attachment to habits and entries
- [ ] Interactive TUI interface for advanced workflows
- [ ] Bulk operations and advanced note management

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
    - WARN: Performance concern for notational velocity use case
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

### 1. Foundation (COMPLETED)
- [x] **T041 Unix Interop Foundation**: Complete foundation with ZK integration, SRS database, basic CLI
  - [x] ZK tool integration and auto-init
  - [x] SRS database schema and operations  
  - [x] Basic commands: add, list, due, edit
  - [x] Tag-based behavior system (vice:type:*)
  - [x] Directory auto-creation and notebook initialization

### 2. Content Change Detection & Advanced SRS
- [ ] **2.1 Context-level Git Integration** (Extract from T041/6.2)
  - *Scope:* Auto-init git in VICE_CONTEXT, auto-commit after file operations
  - *Purpose:* Audit trail and content change detection for SRS quality assessment
  - *Implementation:* Add GitEnabled field to ViceEnv, AutoCommit() method
  - *Integration:* Hook into all file-modifying vice commands
- [ ] **2.2 Git-based SRS Quality Assessment** (Extract from T041/6.2)
  - *Scope:* Detect content changes via git diff for idea note quality scoring
  - *Quality Mapping:* No changes = quality 2, minor = 5, major = 6 (per SM-2 adaptation research)
  - *Integration:* Update `vice flotsam edit` workflow with pre/post-edit change detection
  - *Fallback:* Graceful degradation when git unavailable
- [ ] **2.3 Mtime-based Change Detection** (Extract from T041/6.3)  
  - *Scope:* File timestamp + content hash comparison when git unavailable
  - *Database:* Add last_reviewed, last_content_hash columns to SRS schema
  - *Implementation:* Compare file mtime vs last_reviewed timestamp
  - *Purpose:* Fallback change detection for SRS quality assessment

### 3. Enhanced SRS Workflows
- [ ] **3.1 Advanced Review Interface**
  - *Scope:* Enhanced `vice flotsam review` with statistics and progress tracking
  - *Features:* Review sessions, performance metrics, overdue handling
  - *Integration:* Use content change detection for automatic quality assessment
- [ ] **3.2 SRS Statistics and Analytics**
  - *Scope:* Review performance tracking and learning analytics
  - *Implementation:* Database queries for review history, success rates, scheduling effectiveness
  - *UI:* Command-line statistics display and progress reports

### 4. Vice Ecosystem Integration
- [ ] **4.1 Habit Integration**
  - *Scope:* Attach flotsam notes to habits and habit entries
  - *Data Model:* Add flotsam_id field to habit schema
  - *Workflow:* Reference notes from habit completion and review
- [ ] **4.2 Cross-Context Operations** 
  - *Scope:* Optional access to flotsam notes across vice contexts
  - *Implementation:* Context-aware search and linking capabilities
  - *Security:* Respect context isolation by default with explicit cross-context flags

### 5. Advanced User Experience
- [ ] **5.1 Interactive TUI Interface**
  - *Scope:* Bubbletea-based interactive interface for advanced workflows
  - *Features:* Three-pane view (search, list, preview), in-app operations
  - *Integration:* Combine with ZK delegation for hybrid UX
- [ ] **5.2 Bulk Operations and Management**
  - *Scope:* Batch operations on multiple notes
  - *Features:* Bulk tagging, SRS scheduling, export/import
  - *Performance:* Efficient processing of large note collections

### 6. Templates and Automation  
- [ ] **6.1 Advanced Templates**
  - *Scope:* Enhanced note creation templates beyond basic type templates
  - *Features:* User-defined templates, automation hooks, dynamic content
  - *Integration:* ZK-compatible template system
- [ ] **6.2 Workflow Automation**
  - *Scope:* Automated note creation and management triggers
  - *Features:* Create notes from habit completion, scheduled note creation
  - *Integration:* Hook into vice's event system


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