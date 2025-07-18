# Flotsam Package Documentation

## Overview

The `flotsam` package provides Unix interop functionality for the flotsam note system, implementing a **tool orchestration architecture** that delegates complex operations to zk while maintaining vice-specific SRS functionality. This package combines selective components from the [ZK note-taking system](https://github.com/zk-org/zk) and [go-srs](https://github.com/revelaction/go-srs) to create a **productivity orchestration platform**.

## Architecture

### Unix Interop Design

The flotsam system uses **tool delegation** where zk handles note management and vice handles SRS scheduling. This design ensures:

- **Tool Specialization**: zk handles search/linking/editing, vice handles SRS/habit integration
- **Clean Separation**: `.vice/` folder for vice data, `.zk/` folder for zk data
- **Performance**: SQLite SRS database for fast scheduling queries
- **Composability**: Standard Unix tool composition patterns

### Component Integration

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        UNIX INTEROP ARCHITECTURE                               │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐              │
│  │   ZK TOOL       │    │ VICE SRS        │    │ ORCHESTRATION   │              │
│  │   (External)    │    │ (Internal)      │    │ LAYER           │              │
│  │                 │    │                 │    │                 │              │
│  │ • Note Search   │    │ • SM-2 Algorithm│    │ • Tool Abstraction│            │
│  │ • Editor Integ  │    │ • SRS Database  │    │ • Command Routing│             │
│  │ • Link Analysis │    │ • Review System │    │ • Error Handling│              │
│  │ • Fuzzy Finding │    │ • Quality Scale │    │ • Cache Management│            │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘              │
│           │                       │                       │                     │
│           ▼                       ▼                       ▼                     │
│  ┌─────────────────────────────────────────────────────────────────┐            │
│  │                    FLOTSAM COMMANDS                             │            │
│  │                                                                 │            │
│  │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐       │            │
│  │  │ vice flotsam  │  │ vice flotsam  │  │ vice flotsam  │       │            │
│  │  │ list          │  │ due           │  │ edit          │       │            │
│  │  │ (zk + SRS)    │  │ (SRS query)   │  │ (zk delegate) │       │            │
│  │  └───────────────┘  └───────────────┘  └───────────────┘       │            │
│  └─────────────────────────────────────────────────────────────────┘            │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Directory Structure

**Note**: Cross-reference with [`doc/specifications/file_paths_runtime_env.md`](./file_paths_runtime_env.md) for full context directory structure.

The flotsam notebook is located within a vice context at `$VICE_DATA/{context}/flotsam/` by default:

```
$VICE_DATA/{context}/        # Vice context root (see file_paths_runtime_env.md)
├── habits.yml              # Vice habit definitions  
├── entries.yml             # Vice daily completion data
└── flotsam/                # Notebook directory (configurable, defaults to "flotsam")
    ├── .zk/                # ZK notebook data
    │   ├── config.toml     # zk configuration (managed by vice)
    │   ├── notebook.db     # zk's database (search, links, metadata)
    │   └── templates/
    │       └── flotsam.md  # vice-specific note template
    ├── .vice/              # Vice notebook data (alongside .zk/)
    │   ├── flotsam.db     # SRS database (scheduling, reviews)
    │   └── config.toml    # vice notebook-local config
    ├── concept-1.md       # Clean notes with vice:srs tag
    └── concept-2.md       # zk handles content, vice handles SRS
```

**Key Design Points**:
- **Notebook Directory**: `$VICE_DATA/{context}/flotsam/` contains both `.zk/` and `.vice/` 
- **Coexistence**: `.zk/` and `.vice/` directories are siblings within the notebook
- **SRS Database Location**: `.vice/flotsam.db` placed alongside `.zk/notebook.db`
- **Context Isolation**: Each vice context has its own flotsam notebook directory

**Future Extensibility Considerations**:
- **Custom Notebook Paths**: Notebook directory name should be configurable via `config.toml`
- **Multiple Database Types**: Support both notebook-level (SRS) and context-level (habits) databases
- **Database Placement Strategy**: Different database types may have different placement rules:
  ```
  $VICE_DATA/{context}/
  ├── .vice/
  │   └── habits.db              # Context-level database
  └── {notebook_name}/           # Configurable notebook directory
      ├── .zk/
      ├── .vice/
      │   └── flotsam.db        # Notebook-level database
      └── notes.md
  ```

## Core Data Structures

### Simplified Note Model

**Unix Interop Design**: Notes are **clean markdown files** with minimal frontmatter. Complex operations are delegated to zk.

#### Note Frontmatter (ZK-Compatible)

```yaml
---
id: abc1
title: My Concept Note
created-at: 2025-07-18T10:30:00Z
tags: [vice:srs, vice:type:flashcard, concept, important]
---

# My Concept Note

Content goes here with [[wikilinks]] to other notes.
```

#### Single Note Operations

```go
// Simplified note structure for single-note operations
type FlotsamNote struct {
    ID       string    `yaml:"id"`
    Title    string    `yaml:"title"`
    Created  time.Time `yaml:"created-at"`
    Tags     []string  `yaml:"tags"`
    
    // Runtime fields (not in frontmatter)
    Body     string    `yaml:"-"`
    FilePath string    `yaml:"-"`
    Modified time.Time `yaml:"-"`
}

// Helper methods for tag-based behavior
func (n *FlotsamNote) HasTag(tag string) bool {
    for _, t := range n.Tags {
        if t == tag {
            return true
        }
    }
    return false
}

func (n *FlotsamNote) HasType(noteType string) bool {
    return n.HasTag("vice:type:" + noteType)
}

func (n *FlotsamNote) IsFlashcard() bool {
    return n.HasType("flashcard")
}

func (n *FlotsamNote) HasSRS() bool {
    return n.HasTag("vice:srs")
}
```

### Tag-based Behavior System

**Design Principle**: Use zk's tag system for note behaviors instead of separate fields. This keeps the source of truth in markdown while leveraging zk's powerful tag query capabilities.

#### Vice-specific Tag Patterns

```bash
# Core behavior tags (all vice:type:* notes participate in SRS by default)
vice:type:flashcard   # Question/answer cards for SRS scheduling
vice:type:idea        # Free-form idea capture for SRS scheduling
vice:type:script      # Executable scripts and commands for SRS scheduling
vice:type:log         # Journal entries and logs for SRS scheduling

# Hierarchical tags (future extensibility)
vice:type:flashcard:active    # Currently being reviewed
vice:type:flashcard:suspended # Temporarily disabled
vice:habit:daily             # Daily habit integration
```

#### Composable Operations

```bash
# Find all flashcards due for review (all vice:type:flashcard are SRS-enabled)
zk list --tag "vice:type:flashcard" --format path | 
    vice flotsam due --stdin

# Interactive review of overdue flashcards  
zk list --tag "vice:type:flashcard" --format path | 
    vice flotsam due --stdin --overdue --interactive

# Edit all script notes
zk edit --tag "vice:type:script" --interactive

# Batch review all notes of a specific type
zk list --tag "vice:type:script" --format path | 
    vice flotsam review --stdin

# Get all SRS-enabled notes (any vice:type:* tag)
zk list --tag "vice:type:*" --format path |
    vice flotsam status --stdin
```

**Benefits**:
- **Discoverable**: `zk list --tag "vice:type:flashcard"`
- **Composable**: Complex queries via zk's tag system
- **Source of truth**: Lives in markdown, managed by zk
- **No sync issues**: No need to cache behavior data
- **Extensible**: New behaviors via new tags

### SRS Database Schema

**SRS data lives in SQLite** (`$VICE_DATA/{context}/flotsam/.vice/flotsam.db`), **not in frontmatter**:

```sql
CREATE TABLE srs_reviews (
    note_path TEXT PRIMARY KEY,
    note_id TEXT NOT NULL,
    context TEXT NOT NULL,
    
    -- SM-2 algorithm fields
    easiness REAL NOT NULL DEFAULT 2.5,
    consecutive_correct INTEGER NOT NULL DEFAULT 0,
    due_date INTEGER NOT NULL,
    total_reviews INTEGER NOT NULL DEFAULT 0,
    
    -- Metadata
    created_at INTEGER NOT NULL,
    last_reviewed INTEGER,
    
    -- Optional cache fields (for performance)
    title TEXT,              -- Cached from zk for display
    last_synced INTEGER,     -- Cache invalidation timestamp
    
    FOREIGN KEY (context) REFERENCES contexts(name)
);
```

**Design Notes**:
- **Minimal caching**: Only cache title for display purposes
- **No type caching**: Use zk tag queries for type-based filtering
- **Cache invalidation**: `last_synced` timestamp for cache management
- **Composition over caching**: Better to compose zk queries than duplicate data

## Performance Strategy

### Hybrid Approach: Unix Interop + In-Memory Fallback

**Design Principle**: Use Unix interop for most operations, with in-memory collection loading preserved for performance-critical UX scenarios.

#### Unix Interop (Primary)
```bash
# Most operations use zk delegation
zk list --tag "vice:srs" --format json | vice flotsam due --stdin
zk edit --tag "vice:type:flashcard" --interactive
```

**Benefits**: Tool specialization, composability, reduced maintenance
**Use Cases**: One-off searches, batch operations, scripting

#### In-Memory Collection (Performance Fallback)
```go
// For search-as-you-type and high-frequency operations
type FlotsamCollection struct {
    Notes    []FlotsamNote
    noteMap  map[string]*FlotsamNote      // Fast lookup by ID
    titleIdx map[string][]*FlotsamNote    // Fast title search
}

func LoadAllNotes(contextDir string) (*FlotsamCollection, error)
func (c *FlotsamCollection) SearchByTitle(query string) []*FlotsamNote
func (c *FlotsamCollection) FilterByTags(tags []string) []*FlotsamNote
```

**Benefits**: Sub-millisecond search, no process spawning overhead
**Use Cases**: Interactive search, real-time filtering, TUI applications

#### Adaptive Performance Selection
```go
func SearchNotes(query string, interactive bool) ([]*FlotsamNote, error) {
    if interactive && len(query) > 0 {
        // Use in-memory collection for real-time search
        collection, err := LoadAllNotes(contextDir)
        if err != nil {
            return nil, err
        }
        return collection.SearchByTitle(query), nil
    } else {
        // Use zk for one-off searches
        return searchViaZK(query)
    }
}
```

### Preserved Components from T027

**Performance-Critical (Keep)**:
- `LoadFlotsam()` - Collection loading into memory
- `parseFlotsamFile()` - Single note parsing
- In-memory search/filter operations
- `computeBacklinks()` - For zk-unavailable scenarios

**Utility Functions (Keep)**:
- `saveFlotsamNote()` - Atomic file operations
- `serializeFlotsamNote()` - Frontmatter serialization
- File path validation and security

**Abstraction Layer (Remove)**:
- `DataRepository` interface
- CRUD method abstractions
- Context switching complexity
- Complex error wrapping

**Refactored Location**: 
- `internal/flotsam/collection.go` - In-memory collection operations
- `internal/flotsam/files.go` - File I/O utilities
- `internal/flotsam/search.go` - Search operations with fallback logic

### SRS Data Structure

```go
// SRS data structure for algorithm operations
type SRSData struct {
    NotePath           string    `json:"note_path"`
    Context            string    `json:"context"`
    
    // SM-2 algorithm fields
    Easiness           float64   `json:"easiness"`
    ConsecutiveCorrect int       `json:"consecutive_correct"`
    Due                int64     `json:"due"`
    TotalReviews       int       `json:"total_reviews"`
    
    // Metadata
    CreatedAt          time.Time `json:"created_at"`
    LastReviewed       *time.Time `json:"last_reviewed,omitempty"`
    
    // Cache fields
    Title              string    `json:"title,omitempty"`
    Tags               []string  `json:"tags,omitempty"`
}
```

### Link

Represents a link found in note content:

```go
type Link struct {
    Title        string        // Display text
    Href         string        // Target reference
    Type         LinkType      // Wikilink, Markdown, or Implicit
    IsExternal   bool          // Whether link points outside the note collection
    Rels         []LinkRelation // Relationship types (up/down)
    Snippet      string        // Surrounding context
    SnippetStart int          // Context start position
    SnippetEnd   int          // Context end position
}
```

## API Reference

### ID Generation

#### `NewIDGenerator(opts IDOptions) func() string`

Creates a new ID generator function using the specified options.

```go
generator := NewIDGenerator(IDOptions{
    Length:  4,
    Case:    CaseLower,
    Charset: CharsetAlphanum,
})
noteID := generator() // Returns "a1b2"
```

#### `NewFlotsamIDGenerator() IDGenerator`

Creates an ID generator with flotsam-specific defaults (4-char alphanum lowercase, ZK-compatible).

```go
generator := NewFlotsamIDGenerator()
noteID := generator() // Returns ZK-compatible ID like "3k9z"
```

### Frontmatter Parsing

#### `parseFrontmatter(content string) (map[string]interface{}, string, error)`

Parses YAML frontmatter from markdown content.

```go
frontmatter, body, err := parseFrontmatter(noteContent)
if err != nil {
    return err
}
// frontmatter contains parsed YAML data
// body contains markdown content without frontmatter
```

### Link Extraction

#### `ExtractLinks(content string) []Link`

Extracts all links from markdown content using goldmark AST parsing.

```go
links := ExtractLinks(markdownContent)
for _, link := range links {
    fmt.Printf("Link: %s -> %s (type: %v)\n", link.Title, link.Href, link.Type)
}
```

#### `ExtractWikiLinkTargets(content string) []string`

Extracts only the target hrefs from wikilinks for backlink processing.

```go
targets := ExtractWikiLinkTargets(content)
// Returns ["target1", "target2"] for [[target1]] and [[target2]]
```

#### `BuildBacklinkIndex(notes map[string]string) map[string][]string`

Builds a map of note targets to their source notes for backlink computation.

```go
backlinks := BuildBacklinkIndex(noteContents)
// backlinks["target"] = ["source1", "source2"]
```

### SRS Processing

#### `NewSM2Calculator() *SM2Calculator`

Creates a new SM-2 algorithm calculator.

```go
calc := NewSM2Calculator()
```

#### `ProcessReview(oldData *SRSData, quality Quality) (*SRSData, error)`

Updates SRS data based on a review session.

```go
updatedSRS, err := calc.ProcessReview(currentSRS, CorrectHard)
if err != nil {
    return err
}
// updatedSRS contains new scheduling information
```

#### `IsDue(data *SRSData) bool`

Checks if a card is due for review.

```go
if calc.IsDue(note.SRS) {
    // Present note for review
}
```

### Quality Scale

The SRS system uses a 0-6 quality scale, adapted from traditional memory recall to **idea development progress**:

```go
const (
    NoReview          Quality = 0  // No review performed
    IncorrectBlackout Quality = 1  // Total creative block/needs major restructuring
    IncorrectFamiliar Quality = 2  // Stalled/stuck - needs more frequent attention  
    IncorrectEasy     Quality = 3  // Minor progress/small changes
    CorrectHard       Quality = 4  // Progressing with effort/moderate development
    CorrectEffort     Quality = 5  // Good progress/idea advancing well
    CorrectEasy       Quality = 6  // Flowing/major development/idea maturing
)
```

**Interpretation for Creative Work** (per SM-2 adaptation research):
- **Quality ≤2**: Triggers "failure" response - more frequent scheduling to prevent languishing
- **Quality 3-4**: Normal development cycle - standard SM-2 intervals  
- **Quality 5-6**: High-performing ideas - longer intervals, less frequent review needed

## Usage Examples

### Basic Note Creation and Processing

```go
// Generate a unique ID
generator := NewFlotsamIDGenerator()
noteID := generator()

// Create note content
noteContent := fmt.Sprintf(`---
id: %s
title: My First Note
created-at: %s
tags: [example, test]
---

# My First Note

This note links to [[another note]] and demonstrates SRS functionality.
`, noteID, time.Now().Format(time.RFC3339))

// Parse the note
frontmatter, body, err := parseFrontmatter(noteContent)
if err != nil {
    log.Fatal(err)
}

// Extract links
links := ExtractLinks(body)
fmt.Printf("Found %d links\n", len(links))

// Create flotsam note structure
note := &FlotsamNote{
    ID:    frontmatter["id"].(string),
    Title: frontmatter["title"].(string),
    Body:  body,
    Links: make([]string, len(links)),
}

for i, link := range links {
    note.Links[i] = link.Href
}
```

### SRS Integration Workflow

```go
// Initialize SRS for a note
srsData := &SRSData{
    Easiness:           2.5,
    ConsecutiveCorrect: 0,
    Due:                time.Now().Unix(),
    TotalReviews:       0,
    ReviewHistory:      []ReviewRecord{},
}

note.SRS = srsData

// Check if due for review
calc := NewSM2Calculator()
if calc.IsDue(note.SRS) {
    // Present note for review
    fmt.Printf("Review note: %s\n", note.Title)
    
    // Process review result
    quality := CorrectHard // User's assessment
    updatedSRS, err := calc.ProcessReview(note.SRS, quality)
    if err != nil {
        log.Fatal(err)
    }
    
    // Update note with new SRS data
    note.SRS = updatedSRS
    
    // Save back to file (implementation depends on storage layer)
}
```

### Link Processing and Backlinks

```go
// Collection of notes for backlink processing
noteContents := map[string]string{
    "note1": "This links to [[note2]] and [[note3]]",
    "note2": "This links back to [[note1]]",
    "note3": "Standalone note",
}

// Build backlink index
backlinks := BuildBacklinkIndex(noteContents)

// backlinks["note1"] = ["note2"]
// backlinks["note2"] = ["note1"]
// backlinks["note3"] = ["note1"]

// Extract specific link types
for noteID, content := range noteContents {
    links := ExtractLinks(content)
    for _, link := range links {
        if link.Type == LinkTypeWikiLink && !link.IsExternal {
            fmt.Printf("%s -> %s\n", noteID, link.Href)
        }
    }
}
```

## Performance Considerations

### Search Operations

When processing large collections of notes:

- **Batch Processing**: Process multiple notes in a single operation
- **Link Extraction**: Goldmark AST parsing is efficient but cache results for repeated operations
- **Frontmatter Parsing**: YAML parsing is lightweight but consider caching for frequently accessed notes

Performance benchmark: ~19µs per note for complete processing (ID generation + frontmatter parsing + link extraction + SRS processing).

### Bulk SRS Processing

For processing many due cards:

- **Batch Calculations**: Process multiple reviews in a single transaction
- **Due Date Queries**: Use efficient date range queries when implementing SQLite cache
- **Memory Management**: Process cards in chunks to avoid memory pressure

### Directory Scanning

When scanning large note collections:

- **Incremental Processing**: Only process changed files using timestamp + checksum
- **Parallel Processing**: Parse notes concurrently where possible
- **Cache Invalidation**: Implement efficient cache invalidation strategies

### Cache Synchronization

For SQLite performance cache:

- **Atomic Operations**: Ensure file writes and cache updates are atomic
- **Change Detection**: Use timestamp + SHA256 checksum for efficient change detection
- **Recovery**: Implement cache rebuild from source files on corruption

## Integration with Vice Architecture

### Repository Pattern Integration

The flotsam package is designed to integrate with Vice's Repository Pattern (T028):

```go
// Extend DataRepository interface
type DataRepository interface {
    // Existing methods...
    
    // Flotsam methods
    LoadFlotsam(ctx string) (*FlotsamCollection, error)
    SaveFlotsam(ctx string, flotsam *FlotsamCollection) error
    CreateFlotsamNote(ctx string, flotsam *FlotsamNote) error
    GetFlotsamNote(ctx string, id string) (*FlotsamNote, error)
    UpdateFlotsamNote(ctx string, flotsam *FlotsamNote) error
    DeleteFlotsamNote(ctx string, id string) error
    SearchFlotsam(ctx string, query string) ([]*FlotsamNote, error)
}
```

### Context Isolation

All flotsam operations respect Vice's context system:

- **Directory Structure**: `$VICE_DATA/{context}/flotsam/`
- **Link Resolution**: Wikilinks resolved within context boundaries
- **Cache Isolation**: Separate cache databases per context
- **Backlink Computation**: Computed within context scope

### ZK Notebook Integration

For ZK notebook compatibility:

- **Hybrid Architecture**: Separate metadata stores prevent conflicts
- **Directory Bridge**: Vice operates on ZK notebooks without modification
- **Frontmatter Extensions**: ZK preserves Vice-specific fields in metadata
- **Database Coexistence**: Vice tables added to ZK database are ignored by ZK

## Error Handling

The package provides comprehensive error handling:

```go
// Quality validation
if err := quality.Validate(); err != nil {
    // Handle invalid quality rating
}

// SRS processing errors
updatedSRS, err := calc.ProcessReview(srsData, quality)
if err != nil {
    // Handle SRS calculation errors
}

// Link extraction errors are handled internally
links := ExtractLinks(content) // Returns empty slice on error
```

## Testing

The package includes comprehensive test coverage:

- **Unit Tests**: Individual component testing
- **Integration Tests**: Cross-component workflow testing
- **Performance Tests**: Benchmarking for optimization
- **Compatibility Tests**: ZK interoperability validation

Integration test coverage includes:
- Complete note lifecycle (creation → parsing → linking → SRS → serialization)
- Cross-component data flow validation
- Performance benchmarking (19µs per note average)
- Round-trip data integrity verification

## License and Attribution

This package incorporates code from two external projects:

### ZK Components (GPLv3)
- **Source**: https://github.com/zk-org/zk
- **License**: GNU General Public License v3.0
- **Components**: Frontmatter parsing, link extraction, ID generation
- **Files**: `zk_parser.go`, `zk_links.go`, `zk_id.go`

### go-srs Components (Apache-2.0)
- **Source**: https://github.com/revelaction/go-srs
- **License**: Apache License 2.0
- **Components**: SM-2 algorithm, SRS interfaces, review system
- **Files**: `srs_sm2.go`, `srs_interfaces.go`, `srs_review.go`

All components are properly attributed with copyright headers and license compliance documentation.

## Tool Abstraction & Integration

### Command-Line Tool Interface Design

**Architecture Principle**: Use composition with interface segregation to support multiple external tools (zk, taskwarrior, remind) without inheritance complexity.

#### Core Tool Abstraction

```go
// CommandLineTool provides generic interface for external command-line tools
type CommandLineTool interface {
    Name() string
    Available() bool
    Execute(args ...string) (*ToolResult, error)
}

// ZKTool extends CommandLineTool with zk-specific operations
type ZKTool interface {
    CommandLineTool
    List(filters ...string) ([]Note, error)
    Edit(paths ...string) error
    GetLinkedNotes(path string) ([]string, []string, error) // backlinks, outbound
}

// Concrete implementation for zk tool
type ZKExecutable struct {
    path      string    // Resolved zk binary path
    available bool      // Runtime availability status
    warned    bool      // Track if user has been warned about missing zk
}
```

**Design Benefits**:
- **Generic Interface**: `CommandLineTool` works for zk, taskwarrior (`tw`), remind (`rem`)
- **Tool-Specific Extensions**: `ZKTool` adds zk-specific operations
- **Composition**: ViceEnv contains tool instances, no inheritance
- **Future Extensibility**: New tools implement `CommandLineTool` interface

#### ViceEnv Integration

```go
type ViceEnv struct {
    // existing fields...
    ZK *ZKExecutable  // nil if zk unavailable
}

// Usage pattern with graceful degradation
func (env *ViceEnv) ListFlotsamNotes(filters ...string) ([]Note, error) {
    if env.ZK != nil && env.ZK.Available() {
        return env.ZK.List(filters...)
    }
    
    // Fallback to in-memory collection
    collection, err := env.loadInMemoryCollection()
    if err != nil {
        return nil, fmt.Errorf("zk unavailable and fallback failed: %w", err)
    }
    return collection.FilterByTags(filters), nil
}
```

### Graceful Degradation Strategy

**Availability Levels**:
- **Full Functionality**: zk available, all Unix interop operations work
- **Degraded Mode**: zk unavailable, fallback to in-memory operations where possible
- **Failed Operations**: Operations requiring zk (edit, link analysis) return helpful errors

**Error Handling**:
- **Interactive Sessions**: Warn once per session to stdout with installation guidance
- **Non-Interactive Commands**: Return error messages directing to zk installation
- **Installation URL**: Direct users to https://github.com/zk-org/zk for installation

**Functionality Matrix**:

| Operation | ZK Available | ZK Unavailable | Notes |
|-----------|-------------|----------------|-------|
| List notes | zk delegation | In-memory fallback | Performance difference |
| Search notes | zk queries | In-memory search | Limited query capabilities |
| Edit notes | zk editor | Error + guidance | Requires external editor |
| Link analysis | zk commands | Error + guidance | No fallback available |
| SRS queries | Database only | Database only | Independent of zk |

### ZK Configuration Management

**Shared Ownership Model**: Vice and user share responsibility for `.zk/config.toml` management.

**Configuration Validation** *(Future Enhancement)*:
```go
type ZKConfig struct {
    NoteDir     string                 `toml:"note-dir"`
    NotebookDir string                 `toml:"notebook-dir"`  
    Editor      string                 `toml:"editor"`
    Custom      map[string]interface{} `toml:",remainder"`
}

func ValidateZKConfig(configPath string) (*ZKConfig, error) {
    // NOOP for now - placeholder for future validation
    // Future: Check for incompatible note-dir, ID format conflicts
    return parseZKConfig(configPath), nil
}
```

**Responsibilities**:
- **Vice**: Validate compatibility, create initial config for new notebooks
- **User**: Free to modify non-breaking settings (editor, custom fields)
- **Breaking Changes**: Unexpected note file formats, incompatible ID schemes (future detection)

### Command Pipeline Support

**Future Enhancement**: Complex tool orchestration using command pipelines.

**Reference Library**: [go-command-chain](https://github.com/rainu/go-command-chain) for building command pipelines when needed.

**Use Cases**:
```bash
# Example: Complex workflow combining zk + vice
zk list --tag "vice:srs" --format path | 
    vice flotsam due --stdin --overdue | 
    zk edit --interactive
```

**Implementation**: Add pipeline support when tool orchestration becomes complex enough to warrant abstraction.

## Content Change Detection for SRS Quality Assessment

### Problem Statement

For `idea` and other non-flashcard note types, SRS quality assessment needs to measure engagement/development rather than binary correctness. Content change detection after editing provides a meaningful metric for this assessment.

### Context-Level Git Integration

**Auto-versioning at VICE_CONTEXT level** provides comprehensive change tracking and audit trail.

#### Architecture

```
$VICE_DATA/{context}/
├── .git/                    # Auto-initialized git repository
│   ├── objects/            # Version history of all vice operations
│   ├── refs/heads/main     # Single branch for linear history  
│   └── config             # Git configuration
├── habits.yml             # Vice habit definitions
├── entries.yml            # Vice daily completion data
└── flotsam/               # Notebook directory
    ├── .zk/               # ZK notebook data
    ├── .vice/             # Vice notebook data  
    ├── note1.md           # User notes (git-tracked)
    └── note2.md
```

#### Implementation Strategy

```go
// Auto-commit after file-modifying vice commands
type ViceEnv struct {
    // existing fields...
    GitEnabled bool     // Auto-detected git availability
    GitRepo    string   // Path to context git repository
}

func (env *ViceEnv) AutoCommit(command string) error {
    if !env.GitEnabled || !env.isGitRepo() {
        return env.initContextGit()
    }
    
    // Stage all changes in context directory
    if err := env.gitAdd("."); err != nil {
        return fmt.Errorf("git add failed: %w", err)
    }
    
    // Commit with standardized message
    timestamp := time.Now().Format("15:04:05")
    message := fmt.Sprintf("vice %s - %s", command, timestamp)
    
    return env.gitCommit(message)
}

func (env *ViceEnv) initContextGit() error {
    if !hasGitCommand() {
        log.Debug("Git not available - skipping auto-versioning")
        return nil
    }
    
    if err := env.gitInit(env.ContextData); err != nil {
        return fmt.Errorf("failed to initialize context git: %w", err)
    }
    
    // Create .gitignore for non-trackable files
    gitignore := `# Vice auto-generated - temp files
.vice/cache/
*.tmp
*.lock
`
    return os.WriteFile(filepath.Join(env.ContextData, ".gitignore"), []byte(gitignore), 0644)
}
```

#### SRS Quality Assessment via Git

**Conceptual Mapping**: Following research on SM-2 adaptation to creative work, we map content changes to SM-2 quality ratings that reflect **idea development progress** rather than recall accuracy.

**Quality Interpretation for Ideas**:
- **No changes** (0-2): Stalled/blocked - idea needs restructuring or more attention
- **Minor changes** (3-4): Progressing - idea is developing but needs continued work  
- **Major changes** (5-6): Flowing - idea is mature and developing well

```go
// Detect content changes after edit operations
func (env *ViceEnv) AssessEditQuality(notePath string, preEditCommit string) (srs.Quality, error) {
    // Check if file changed since pre-edit commit
    hasChanges, err := env.gitHasChanges(notePath, preEditCommit)
    if err != nil {
        // Fallback to mtime comparison
        return env.assessQualityByMtime(notePath)
    }
    
    if !hasChanges {
        return srs.IncorrectFamiliar, nil  // Stalled idea - quality 2 (needs more frequent attention)
    }
    
    // Analyze change magnitude (future enhancement)
    changeSize, err := env.gitDiffStats(notePath, preEditCommit)
    if err != nil {
        return srs.CorrectHard, nil  // Changed but unknown magnitude - quality 4 (progressing)
    }
    
    // Map change size to development progress quality
    switch {
    case changeSize.Lines < 5:
        return srs.CorrectEffort, nil     // Minor development - quality 5 (idea advancing)
    default:
        return srs.CorrectEasy, nil      // Significant development - quality 6 (idea flowing)
    }
}
```

**Key Insight from Research**: SM-2 "failure" (quality ≤2) in creative work indicates an idea is "blocked/stuck" rather than "forgotten". This triggers more frequent review cycles to prevent ideas from languishing, which aligns perfectly with our change detection approach.

#### Integration Points

**Command Integration**:
- `vice flotsam add`: Auto-commit after note creation
- `vice flotsam edit`: Pre-edit commit capture, post-edit quality assessment  
- `vice habit complete`: Auto-commit after habit data updates
- All file-modifying commands trigger `AutoCommit()`

**User Benefits**:
- **Full Audit Trail**: Complete history of vice operations with timestamps
- **Rollback Capability**: `git log` shows all operations, `git reset` for recovery
- **Change Analysis**: Precise diff information for debugging and analysis
- **No Interference**: User can still initialize git in `flotsam/` for manual control

### Mtime-Based Fallback Detection

**File timestamp comparison** when git unavailable or repository not initialized.

#### Implementation Strategy

```go
// Extended SRS schema with last_reviewed timestamp
type SRSData struct {
    // existing fields...
    LastReviewed    *time.Time `json:"last_reviewed,omitempty"`
    LastContentHash string     `json:"last_content_hash,omitempty"` // SHA256 for change detection
}

// Database schema update
const srsSchemaV2 = `
ALTER TABLE srs_reviews ADD COLUMN last_reviewed INTEGER;
ALTER TABLE srs_reviews ADD COLUMN last_content_hash TEXT;
`

func (env *ViceEnv) assessQualityByMtime(notePath string) (srs.Quality, error) {
    // Get current file mtime
    stat, err := os.Stat(notePath)
    if err != nil {
        return srs.NoReview, fmt.Errorf("failed to stat file: %w", err)
    }
    currentMtime := stat.ModTime()
    
    // Get last reviewed time from SRS database
    srsData, err := env.getSRSData(notePath)
    if err != nil {
        return srs.NoReview, err
    }
    
    if srsData.LastReviewed == nil {
        // First review - assume engagement
        return srs.CorrectEffort, nil
    }
    
    // Compare modification times
    if currentMtime.After(*srsData.LastReviewed) {
        // File changed since last review
        return env.assessQualityByContent(notePath, srsData.LastContentHash)
    }
    
    // No changes detected
    return srs.IncorrectFamiliar, nil  // Stale idea - quality 2
}

func (env *ViceEnv) assessQualityByContent(notePath, oldHash string) (srs.Quality, error) {
    content, err := os.ReadFile(notePath)
    if err != nil {
        return srs.CorrectHard, nil  // Assume change but unknown - quality 4
    }
    
    newHash := sha256.Sum256(content)
    newHashStr := hex.EncodeToString(newHash[:])
    
    if newHashStr == oldHash {
        return srs.IncorrectFamiliar, nil  // Mtime changed but content same - quality 2
    }
    
    // Content changed - could analyze magnitude here
    return srs.CorrectHard, nil  // Content changed - quality 4
}
```

#### Workflow Integration

```go
// Edit workflow with change detection
func runFlotsamEdit(notePath string, env *config.ViceEnv) error {
    // Capture pre-edit state
    preEditTime := time.Now()
    var preEditCommit string
    
    if env.GitEnabled {
        // Commit current state before edit
        if err := env.AutoCommit("pre-edit snapshot"); err != nil {
            log.Warn("Failed to create pre-edit commit", "error", err)
        } else {
            preEditCommit = env.getHEADCommit()
        }
    }
    
    // Store pre-edit content hash for mtime fallback
    preEditHash, _ := env.getFileHash(notePath)
    
    // Execute edit via ZK
    if err := env.ZK.Edit(notePath); err != nil {
        return fmt.Errorf("edit failed: %w", err)
    }
    
    // Assess quality based on changes
    var quality srs.Quality
    var err error
    
    if env.GitEnabled && preEditCommit != "" {
        quality, err = env.AssessEditQuality(notePath, preEditCommit)
    } else {
        quality, err = env.assessQualityByMtimeAndHash(notePath, preEditTime, preEditHash)
    }
    
    if err != nil {
        log.Warn("Failed to assess edit quality", "error", err)
        quality = srs.CorrectHard  // Default assumption
    }
    
    // Update SRS database with new quality assessment
    return env.updateSRSAfterEdit(notePath, quality, time.Now())
}
```

### Configuration and User Control

#### Git Integration Settings

```toml
# vice config.toml
[flotsam]
auto_git = true              # Enable context-level git integration  
git_commit_frequency = "command"  # "command", "daily", "manual"
change_detection = "git"     # "git", "mtime", "both"

[flotsam.quality_mapping]
no_change = 2               # Quality for unchanged notes
minor_change = 4            # Quality for small changes  
major_change = 6            # Quality for significant changes
```

#### User Flexibility

**Context-Level Git**: Auto-managed by vice for audit trail and change detection
**Notebook-Level Git**: Optional user-managed git repo in `flotsam/` for detailed version control
**Coexistence**: Both can exist simultaneously without conflict

### Error Handling and Edge Cases

**Design Philosophy**: Treat change detection as a **heuristic** rather than attempting perfect accuracy. Accept git's built-in capabilities for handling edge cases.

```go
// Graceful degradation when git operations fail
func (env *ViceEnv) AutoCommit(command string) error {
    if err := env.attemptGitCommit(command); err != nil {
        log.Debug("Git commit failed, continuing without versioning", "error", err)
        // Operation continues - git failure doesn't block user workflow
        return nil
    }
    return nil
}

// Handle edge cases in change detection  
func (env *ViceEnv) AssessEditQuality(notePath string, preEditCommit string) (srs.Quality, error) {
    // Handle binary files, permission issues, corrupted git state
    if err := env.validateGitState(); err != nil {
        log.Warn("Git state invalid, falling back to mtime", "error", err)
        return env.assessQualityByMtime(notePath)
    }
    
    // Handle cases where file was deleted and recreated
    if !env.fileExistsInCommit(notePath, preEditCommit) {
        return srs.CorrectEasy, nil  // New content creation - quality 6
    }
    
    // Accept git's change detection for moves, renames, permission changes
    // Don't distinguish content vs metadata changes - it's all engagement
    return env.gitBasedQualityAssessment(notePath, preEditCommit)
}
```

#### Implementation Constraints

**Database Migration**: No existing SRS databases to migrate - ignore migration complexity
**Repository Conflicts**: Not a concern for user workflows - ignore nested git repo warnings for now
**Quality Granularity**: Fixed thresholds initially - make configurable later if needed
**Performance**: Standard git performance acceptable - no optimization needed
**Multi-user**: Not a consideration for single-user contexts
**Edge Cases**: Accept git's limitations as "good enough" heuristic

#### Heuristic Acceptance Criteria

- **File renames/moves**: Trust git's move detection
- **Permission changes**: Count as engagement if git detects change
- **Metadata-only changes**: Acceptable false positive for engagement measurement
- **Binary files**: Fallback to mtime comparison
- **Corrupted git state**: Graceful fallback to mtime without blocking workflow

## Future Enhancements

### Planned Features
- **Template System**: Note creation templates using handlebars
- **Advanced Scheduling**: Alternative SRS algorithms (SM-18, FSRS)
- **Bulk Operations**: Batch processing for large note collections
- **Query Language**: Advanced search and filtering capabilities
- **Change Magnitude Analysis**: Diff-based quality assessment with line/character counts
- **Semantic Change Detection**: Content analysis beyond simple text changes

### Performance Optimizations
- **Streaming Processing**: Process large note collections without loading all into memory
- **Incremental Updates**: More efficient change detection and processing
- **Cache Optimization**: Advanced caching strategies for frequently accessed data
- **Parallel Processing**: Concurrent note processing for bulk operations
- **Git Performance**: Shallow clones and sparse checkouts for large repositories

### Integration Enhancements
- **Task Management**: Integration with Vice's task-oriented workflows
- **Incremental Writing**: Support for progressive note development
- **Context Switching**: Efficient context migration and synchronization
- **Export/Import**: Data portability between different note systems
- **Multi-Tool Orchestration**: Pipeline composition with taskwarrior, remind, etc.