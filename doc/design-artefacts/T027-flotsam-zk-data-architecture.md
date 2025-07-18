

# DESIGN OVERVIEW: Flotsam & ZK Data Architecture

**Date**: 2025-07-18  
**Scope**: Comprehensive overview of flotsam and ZK data structures, components & integration patterns  
**Source**: Extracted from [T027 Flotsam Data Layer](/kanban/in-progress/T027_flotsam_data_layer.md) implementation

## Related Documentation

**Source Task**: [T027 Flotsam Data Layer](/kanban/in-progress/T027_flotsam_data_layer.md) - Complete flotsam data layer implementation  

**Related ADRs**:
- [ADR-002: Flotsam Files-First Architecture](/doc/decisions/ADR-002-flotsam-files-first-architecture.md) - Source of truth strategy
- [ADR-003: ZK-go-srs Integration Strategy](/doc/decisions/ADR-003-zk-gosrs-integration-strategy.md) - Component integration approach  
- [ADR-004: Flotsam SQLite Cache Strategy](/doc/decisions/ADR-004-flotsam-sqlite-cache-strategy.md) - Performance caching design
- [ADR-006: Flotsam Context Isolation](/doc/decisions/ADR-006-flotsam-context-isolation.md) - Context boundary framework

**Related Specifications**:
- [Flotsam Package Documentation](/doc/specifications/flotsam.md) - Complete API reference
- [File Paths & Runtime Environment](/doc/specifications/file_paths_runtime_env.md) - T028 context management

**Implementation Files**:
- `internal/flotsam/` - ZK & go-srs component integration
- `internal/models/flotsam.go` - Data model definitions
- `internal/repository/file_repository.go` - Repository layer implementation

### Data Flow Overview

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           VICE FLOTSAM DATA LAYER                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐              │
│  │   .md Files     │    │   .yml Files    │    │   .json Files   │              │
│  │ (ZK Compatible) │    │   (Config)      │    │   (Data)        │              │
│  │ SOURCE OF TRUTH │    │ SOURCE OF TRUTH │    │ SOURCE OF TRUTH │              │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘              │
│           │                       │                       │                     │
│           │                       │                       │                     │
│           ▼                       ▼                       ▼                     │
│  ┌─────────────────────────────────────────────────────────────────┐            │
│  │              UNIFIED FILE HANDLER                               │            │
│  │                                                                 │            │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │            │
│  │  │ Change Detection│  │ Content Parsing │  │ Atomic Updates  │  │            │
│  │  │ (Time+Checksum)│  │ (Multi-format)  │  │ (File+Cache)    │  │            │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘  │            │
│  └─────────────────────────────────────────────────────────────────┘            │
│                                  │                                              │
│                                  ▼                                              │
│  ┌─────────────────────────────────────────────────────────────────┐            │
│  │                    SQLITE PERFORMANCE CACHE                     │            │
│  │                                                                 │            │
│  │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐       │            │
│  │  │ vice_srs_cache│  │vice_file_cache│  │vice_contexts  │       │            │
│  │  │ (Fast SRS     │  │ (Change track)│  │ (Context def) │       │            │
│  │  │  queries)     │  │               │  │               │       │            │
│  │  └───────────────┘  └───────────────┘  └───────────────┘       │            │
│  │                                                                 │            │
│  │  Added to existing ZK notebook.db (ZK ignores Vice tables)     │            │
│  └─────────────────────────────────────────────────────────────────┘            │
│                                  │                                              │
│                                  ▼                                              │
│  ┌─────────────────────────────────────────────────────────────────┐            │
│  │                      APPLICATION LAYER                          │            │
│  │                                                                 │            │
│  │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐       │            │
│  │  │ SRS Operations│  │ Link Resolution│  │ Context Mgmt  │       │            │
│  │  │ (Due cards,   │  │ (Wiki links,  │  │ (Isolation,   │       │            │
│  │  │  reviews)     │  │  backlinks)   │  │  switching)   │       │            │
│  │  └───────────────┘  └───────────────┘  └───────────────┘       │            │
│  └─────────────────────────────────────────────────────────────────┘            │
│                                                                                 │
│  Write Flow: User Action → File Update → Cache Sync → Query Cache              │
│  Read Flow:  User Query → Cache Query → Fast Results                           │
│  Recovery:   Corrupt Cache → Rebuild from Files → Consistency Restored        │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

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
    // Existing methods from T028 - IMPLEMENTED ✅
    LoadHabits(ctx string) (*models.Schema, error)
    LoadEntries(ctx string, date time.Time) (*models.EntryLog, error)
    SaveEntries(ctx string, entries *models.EntryLog) error
    LoadChecklists(ctx string) (*models.ChecklistSchema, error)
    SwitchContext(newContext string) error
    
    // New flotsam methods - TO BE IMPLEMENTED
    LoadFlotsam(ctx string) (*FlotsamCollection, error)
    SaveFlotsam(ctx string, flotsam *FlotsamCollection) error
    CreateFlotsamNote(ctx string, flotsam *Flotsam) error
    GetFlotsamNote(ctx string, id string) (*Flotsam, error)
    UpdateFlotsamNote(ctx string, flotsam *Flotsam) error
    DeleteFlotsamNote(ctx string, id string) error
    SearchFlotsam(ctx string, query string) ([]*Flotsam, error)
    
    // T028 integration methods
    GetFlotsamCacheDB(ctx string) (*sql.DB, error)  // Context-aware cache DB
    EnsureFlotsamDir(ctx string) error              // Use T028 ViceEnv paths
}
```

### Key Architectural Decisions

#### Files-First Strategy (Decision Made)
**Source of Truth**: Individual markdown files with YAML frontmatter
- **Directory**: `$VICE_DATA/{context}/flotsam/` OR ZK notebooks
- **Structure**: One `.md` file per note with YAML frontmatter  
- **Filename**: `{id}.md` (e.g., `6ub6.md`) following ZK convention
- **Format**: YAML frontmatter + markdown body

**Complete SRS Data in Frontmatter**:
```yaml
vice:
  srs:
    easiness: 2.5
    consecutive_correct: 0
    due: 1640995200
    total_reviews: 3
    review_history:
      - timestamp: 1640995100
        quality: 4
      - timestamp: 1640995000
        quality: 3
```

#### SQLite Performance Cache (Decision Made)
**Cache Strategy**: Context-aware SQLite database placement
- **ZK Notebooks**: Add Vice tables to existing `.zk/notebook.db` (ZK ignores them)
- **Vice Contexts**: Create `flotsam.db` in `$VICE_DATA/{context}/` directory
- **T028 Integration**: Leverage ViceEnv for context-aware cache database paths
- **Performance**: Fast queries for SRS operations
- **Consistency**: Cache rebuilt from files when checksums change
- **Recovery**: Drop cache tables/database to completely remove Vice

#### Unified File Handler (Design Innovation)
**Multi-Format Support**: Handle .md, .yml, .json files
- **Change Detection**: Timestamp + SHA256 checksum (ZK-inspired)
- **Atomic Operations**: File write → cache update in transactions
- **Error Recovery**: Cache rebuild from source files on corruption
- **ZK Integration**: Co-existence without conflicts

**Decision Rationale**:
- **Data Portability**: All data travels with markdown files
- **Performance**: SQLite cache for fast SRS queries
- **ZK Compatibility**: Works with existing ZK notebooks
- **Reliability**: Source files always recoverable
- **Extensibility**: Multi-format support for future needs

## ZK Schema Architecture Reference

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

## Code Reuse Strategy

### ZK Code Reuse Constraints
- **Cannot import directly**: ZK's useful code is in `internal/` packages (Go prohibits external imports)
- **Application module**: Would pull entire CLI application with all dependencies
- **Recommended approach**: Copy specific code (parsing, linking) with attribution
- **Target files**: `internal/core/note_parse.go`, `internal/core/link.go`, ID generation, templates

### Go-SRS Code Reuse Options
- **Can import directly**: Public API design (`algo/`, `db/`, `uid/` packages)
- **Library module**: Intended for external consumption, clean interfaces
- **Dependency concern**: Would pull BadgerDB when only SM-2 algorithm needed
- **Recommended approach**: Copy SM-2 algorithm (`algo/sm2/`) to avoid heavyweight dependencies

### Go-SRS Analysis Complete
- **Architecture**: Clean interfaces (`db.Handler`, `algo.Algo`, `uid.UID`) with loose coupling
- **Storage**: Simple schema (easiness, consecutive_correct, due_timestamp) stored as JSON in BadgerDB
- **SM-2 Algorithm**: Mature, well-tested implementation with proper SuperMemo research backing
- **Integration Complexity**: LOW-MEDIUM - only 3 methods in `db.Handler` interface to implement
- **Data Model**: `{deck_id, card_id, easiness: 2.5, consecutive_correct: 0, due: timestamp}`
- **Quality Scale**: Research-based 0-6 scale (0=no review, 1-3=incorrect, 4-6=correct)
- **Recommended Approach**: Adapt go-srs with SQLite `db.Handler` implementation
- **Integration Benefits**: Proven algorithm, clean abstractions, time savings vs reimplementation
