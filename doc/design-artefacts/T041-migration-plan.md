# T041 Migration Plan: T027 Code Audit & Unix Interop Strategy

**Status**: Draft - Code audit complete
**Created**: 2025-07-18
**Task**: T041 - Unix Interop Foundation & T027 Migration

## Executive Summary

T027 implemented ~2000+ lines of coupled integration code with repository patterns, in-memory backlink computation, and complex flotsam models. This migration plan identifies what to preserve vs remove while establishing Unix interop foundation with zk.

**Key Finding**: The repository abstraction layer is currently **unused** in the main codebase - only referenced in tests. This significantly reduces migration complexity.

## T027 Code Audit Results

### Repository Layer Analysis (~800 lines)

**Files**: 
- `internal/repository/interface.go` (82 lines) - DataRepository interface
- `internal/repository/file_repository.go` (858 lines) - FileRepository implementation  

**Status**: ✅ **SAFE TO ISOLATE** - Only used in tests, no production consumers

**Key Components**:
- `DataRepository` interface with 13 flotsam methods (lines 43-66)
- `FileRepository` struct implementing complete CRUD + search operations
- Context-aware file operations with atomic save patterns
- Collection-wide backlink computation (lines 418-442)

**Dependencies**: 
- `internal/flotsam/*` - ZK parser, link extraction, SRS interfaces
- `internal/models/flotsam.go` - FlotsamNote, FlotsamCollection models
- Standard file operations, YAML serialization

### Flotsam Models Analysis (~388 lines)

**File**: `internal/models/flotsam.go`

**Status**: ⚠️ **PARTIALLY PRESERVE** - Rich models needed for SRS, but can simplify

**Key Components**:
- `FlotsamNote` struct embedding `flotsam.FlotsamNote` (lines 46-51)
- `FlotsamCollection` with metadata (lines 55-64) 
- `FlotsamFrontmatter` for ZK compatibility (lines 17-29)
- Comprehensive validation (lines 178-322)

**Preserve**:
- SRS data structures and validation
- Basic note model for single-note operations
- Frontmatter parsing compatibility

**Simplify**:
- Remove collection-level operations
- Reduce validation complexity
- Eliminate in-memory collection management

### Flotsam Package Analysis (~1500+ lines)

**Files**: 15 files in `internal/flotsam/`

**Status**: ✅ **MOSTLY PRESERVE** - Core algorithms valuable for Unix interop

**Key Components**:

**ZK Integration (~600 lines)**:
- `zk_parser.go` - Frontmatter parsing, title extraction
- `zk_links.go` - Goldmark-based link extraction, AST parsing
- `zk_id.go` - ZK-compatible ID generation

**SRS Components (~500 lines)**:
- `srs_interfaces.go` - Algorithm and storage interfaces
- `srs_sm2.go` - SM-2 algorithm implementation
- `srs_review.go` - Review session management

**Utilities (~400 lines)**:
- Link extraction, backlink computation
- Content parsing, validation helpers
- Test files and integration tests

**Decision**: **PRESERVE** - These algorithms are valuable for Unix interop and hard to replace

## Migration Strategy

### Phase 1: Isolation (No Deletion)

**Approach**: Mark repository layer as deprecated, isolate from execution paths

1. **Add deprecation markers** to repository interface and implementation
2. **Create bypass mechanisms** for essential operations
3. **Preserve test coverage** but mark as legacy
4. **Document migration notes** for future reference

### Phase 2: Unix Interop Foundation

**Implement alongside existing code**:

1. **ZK Shell-out Layer**:
   ```go
   // New: internal/zk/shell.go
   type ZKTool interface {
       List(args ...string) ([]string, error)
       Edit(paths ...string) error
       Create(template, title string) (string, error)
   }
   ```

2. **SRS Database Layer**:
   ```go
   // New: internal/srs/database.go  
   type SRSDatabase interface {
       GetDueNotes(ctx string) ([]string, error)
       UpdateReview(notePath string, quality int) error
       GetSRSData(notePath string) (*SRSData, error)
   }
   ```

3. **Simplified Note Operations**:
   ```go
   // New: internal/flotsam/note.go
   func ParseSingleNote(path string) (*FlotsamNote, error)
   func SaveSingleNote(note *FlotsamNote) error
   ```

### Phase 3: Command Implementation

**New CLI commands bypass repository layer**:

```bash
# Implemented via Unix interop
vice flotsam list      # -> zk list --tag vice:srs
vice flotsam due       # -> SRS database query
vice flotsam edit      # -> zk edit
vice doctor           # -> check zk availability
```

### Phase 4: Gradual Migration

**Only after Unix interop is proven**:

1. **Remove repository consumers** from test files
2. **Archive repository code** to separate package
3. **Update documentation** to reflect new architecture
4. **Performance validation** to ensure no regressions

## Component-by-Component Breakdown

### 🔴 MARK FOR ISOLATION (eventual removal)

**Repository Abstraction Layer**:
- `internal/repository/interface.go` - 13 flotsam methods
- `internal/repository/file_repository.go`:
  - `LoadFlotsam()` - loads entire collection (lines 297-354)
  - `SaveFlotsam()` - saves entire collection (lines 444-478)
  - `computeBacklinks()` - in-memory backlink computation (lines 418-442)
  - Search methods - `SearchFlotsam`, `GetFlotsamByType`, `GetFlotsamByTag`
  - Collection CRUD - `CreateFlotsamNote`, `GetFlotsamNote`, etc.

**In-Memory Collection Management**:
- `models.FlotsamCollection` - collection-wide operations
- Collection metadata computation
- Bulk operations and collection validation

**Rationale**: Unix interop delegates these operations to zk and individual file operations

### 🟡 SIMPLIFY (keep essential parts)

**Flotsam Models**:
- `models.FlotsamNote` - **preserve core**, remove collection methods
- `models.FlotsamFrontmatter` - **preserve** for compatibility
- Validation logic - **simplify** but keep SRS validation

**Repository Components**:
- `parseFlotsamFile()` - **preserve** for single-note operations
- `saveFlotsamNote()` - **preserve** for atomic file operations
- `serializeFlotsamNote()` - **preserve** for frontmatter serialization

### 🟢 PRESERVE (valuable algorithms)

**SRS Components**:
- `flotsam.SRSData` struct and validation
- `flotsam.SM2Calculator` - SM-2 algorithm implementation
- `flotsam.Algorithm` interface
- Review session management

**ZK Integration**:
- `flotsam.ParseFrontmatter()` - YAML frontmatter parsing
- `flotsam.ExtractLinks()` - Goldmark-based link extraction
- `flotsam.BuildBacklinkIndex()` - backlink computation algorithm
- ZK ID generation and validation

**File Operations**:
- Atomic save patterns (`temp file + rename`)
- Frontmatter serialization
- Content parsing utilities

**Rationale**: These algorithms are hard to replace and valuable for Unix interop

## Risk Assessment

### Low Risk - Repository Layer Isolation

**Finding**: Repository layer has **zero production consumers**
- Only used in test files
- No CLI commands currently use repository
- Safe to mark as deprecated without breaking changes

### Medium Risk - Model Simplification

**Consideration**: Some model methods may be used by future UI code
- **Mitigation**: Preserve essential model structure, only remove collection operations
- **Approach**: Deprecate collection methods, keep single-note operations

### Low Risk - Algorithm Preservation

**SRS and ZK algorithms are proven and well-tested**
- **SM-2 algorithm**: Mature, mathematically sound
- **Link extraction**: Robust goldmark-based implementation
- **Frontmatter parsing**: Compatible with ZK ecosystem

## Implementation Notes

### Context Isolation

**Preserve T028's context isolation** through:
- Separate SRS databases per context: `.vice/contexts/{context}/flotsam.db`
- ZK notebook boundaries per context
- File operations scoped to context directories

### Performance Considerations

**User concern**: Search-as-you-type performance
- **Current T027**: In-memory collection, fast search
- **Unix interop**: zk commands + SRS database queries
- **Mitigation**: Implement caching layer with mtime validation

### SRS Integration Points

**SM-2 algorithm preserved**:
- `flotsam.SM2Calculator` continues to work
- SRS database replaces frontmatter storage
- Review workflows maintain same quality scale

### Testing Strategy

**Preserve test coverage**:
- Mark repository tests as legacy
- Create new Unix interop tests
- Maintain SRS algorithm test coverage
- Add integration tests for zk shell-out

## Migration Checklist

### Phase 1: Isolation ✅
- [x] **Mark repository layer as deprecated**
- [x] **Document migration plan**
- [x] **Identify bypass mechanisms**

### Phase 2: Unix Interop Foundation
- [ ] **Implement ZK shell-out abstraction**
- [ ] **Create SRS database layer**
- [ ] **Implement basic CLI commands**
- [ ] **Add zk dependency checking**

### Phase 3: Command Implementation
- [ ] **`vice flotsam list`** with zk delegation
- [ ] **`vice flotsam due`** with SRS queries
- [ ] **`vice flotsam edit`** with zk delegation
- [ ] **`vice doctor`** for dependency validation

### Phase 4: Validation & Cleanup
- [ ] **Performance testing** vs T027 baseline
- [ ] **Integration testing** with zk
- [ ] **Archive repository code**
- [ ] **Update documentation**

## Conclusion

**T027 migration is low-risk** due to repository layer being unused in production. The migration can proceed incrementally:

1. **Isolate** repository layer without breaking changes
2. **Implement** Unix interop alongside existing code
3. **Validate** performance and functionality
4. **Archive** deprecated code only after proven replacement

**Key insight**: SRS algorithms and ZK integration code are valuable assets that enable sophisticated Unix interop capabilities. The repository abstraction layer is the primary removal target.

This approach allows for **safe experimentation** with Unix interop while preserving the option to fall back to coupled integration if needed.

---

# Unix Interop Architecture Design

**Status**: Design Phase
**Created**: 2025-07-18  
**Subtask**: T041/1.2 - Unix interop architecture design

## Overview

The Unix interop architecture transforms vice from a monolithic habit tracker into a **TUI orchestrator** that delegates complex operations to specialized Unix tools while maintaining vice-specific functionality through targeted integrations.

## Core Design Principles

### 1. Tool Delegation Strategy
- **Leverage best-in-class tools**: Use zk for note management, fuzzy finding, editor integration
- **Minimal reimplementation**: Only implement what's uniquely vice-specific (SRS, habit integration)
- **Composable workflows**: Enable complex operations through tool orchestration
- **Graceful degradation**: Provide useful functionality even when external tools are unavailable

### 2. Data Architecture
- **Files as primary storage**: Markdown files with YAML frontmatter  
- **Database for performance**: SQLite for SRS scheduling and caching
- **Tag-based behaviors**: Use zk tags for vice-specific note behaviors
- **Context isolation**: Separate databases and notebooks per vice context

### 3. Interface Design
- **Shell-out patterns**: Structured command execution with error handling
- **Structured output**: JSON and template-based parsing
- **Configuration management**: User-configurable tool behavior
- **Error propagation**: Clear error messages with recovery guidance

## Architecture Components

### Tool Abstraction Layer

```go
// Core tool interface for external command execution
type Tool interface {
    Name() string
    IsAvailable() bool
    Execute(cmd string, args ...string) (*Result, error)
    Version() (string, error)
}

// Specialized interface for zk operations
type ZKTool interface {
    Tool
    
    // Note operations
    List(filters ...string) ([]Note, error)
    Create(title, template string) (string, error)
    Edit(paths ...string) error
    
    // Query operations  
    FindByTag(tags ...string) ([]string, error)
    FindLinkedBy(notePath string) ([]string, error)
    FindLinkingTo(notePath string) ([]string, error)
    
    // Interactive operations
    SelectInteractive(filters ...string) ([]string, error)
}

// Command execution result
type Result struct {
    ExitCode int
    Stdout   string
    Stderr   string
    Duration time.Duration
}
```

### SRS Database Integration

```go
// SRS database for spaced repetition scheduling
type SRSDatabase interface {
    // Review operations
    GetDueNotes(context string) ([]SRSNote, error)
    UpdateReview(notePath string, quality int) error
    GetSRSData(notePath string) (*SRSData, error)
    
    // Batch operations
    SyncFromFileSystem(contextDir string) error
    InvalidateCache(notePath string) error
    
    // Maintenance
    Vacuum() error
    GetStats(context string) (*SRSStats, error)
}

// SRS note with file system metadata
type SRSNote struct {
    Path         string
    Title        string
    NextDue      time.Time
    Quality      int
    Interval     int
    LastReviewed time.Time
    Context      string
    Tags         []string
}
```

### Command Orchestration

```go
// High-level command orchestration
type FlotsamOrchestrator struct {
    zk      ZKTool
    srs     SRSDatabase
    config  *Config
    context string
}

// Composite operations combining multiple tools
func (o *FlotsamOrchestrator) ListDueNotes() ([]SRSNote, error) {
    // 1. Query SRS database for due notes
    dueNotes, err := o.srs.GetDueNotes(o.context)
    if err != nil {
        return nil, err
    }
    
    // 2. Enrich with zk metadata (tags, links, etc.)
    for i, note := range dueNotes {
        zkNote, err := o.zk.GetNote(note.Path)
        if err != nil {
            continue // Skip missing notes
        }
        dueNotes[i].Tags = zkNote.Tags
        dueNotes[i].Title = zkNote.Title
    }
    
    return dueNotes, nil
}

func (o *FlotsamOrchestrator) EditDueNotes() error {
    // 1. Get due notes from SRS
    dueNotes, err := o.ListDueNotes()
    if err != nil {
        return err
    }
    
    // 2. Extract paths for zk
    paths := make([]string, len(dueNotes))
    for i, note := range dueNotes {
        paths[i] = note.Path
    }
    
    // 3. Delegate to zk for editing
    return o.zk.Edit(paths...)
}
```

## Command Implementation Patterns

### ZK Shell-out Implementation

```go
type ZKShell struct {
    binary string
    config *ZKConfig
}

func (z *ZKShell) List(filters ...string) ([]Note, error) {
    args := []string{"list", "--format", "json", "--no-pager", "--quiet"}
    args = append(args, filters...)
    
    result, err := z.Execute("list", args...)
    if err != nil {
        return nil, fmt.Errorf("zk list failed: %w", err)
    }
    
    var notes []Note
    if err := json.Unmarshal([]byte(result.Stdout), &notes); err != nil {
        return nil, fmt.Errorf("failed to parse zk output: %w", err)
    }
    
    return notes, nil
}

func (z *ZKShell) FindByTag(tags ...string) ([]string, error) {
    tagFilter := strings.Join(tags, " OR ")
    args := []string{"list", "--tag", tagFilter, "--format", "path", "--quiet"}
    
    result, err := z.Execute("list", args...)
    if err != nil {
        return nil, fmt.Errorf("zk tag query failed: %w", err)
    }
    
    paths := strings.Split(strings.TrimSpace(result.Stdout), "\n")
    return paths, nil
}

func (z *ZKShell) Create(title, template string) (string, error) {
    args := []string{"new", "--title", title}
    if template != "" {
        args = append(args, "--template", template)
    }
    
    result, err := z.Execute("new", args...)
    if err != nil {
        return "", fmt.Errorf("zk create failed: %w", err)
    }
    
    // Parse output to extract created file path
    path := strings.TrimSpace(result.Stdout)
    return path, nil
}
```

### SRS Database Operations

```go
type SQLiteSRS struct {
    db      *sql.DB
    context string
}

func (s *SQLiteSRS) GetDueNotes(context string) ([]SRSNote, error) {
    query := `
        SELECT path, title, next_due, quality, interval_days, last_reviewed, tags 
        FROM srs_notes 
        WHERE context = ? AND next_due <= datetime('now')
        ORDER BY next_due ASC
    `
    
    rows, err := s.db.Query(query, context)
    if err != nil {
        return nil, fmt.Errorf("SRS query failed: %w", err)
    }
    defer rows.Close()
    
    var notes []SRSNote
    for rows.Next() {
        var note SRSNote
        var tagsJSON string
        
        err := rows.Scan(&note.Path, &note.Title, &note.NextDue, 
                        &note.Quality, &note.Interval, &note.LastReviewed, &tagsJSON)
        if err != nil {
            return nil, err
        }
        
        json.Unmarshal([]byte(tagsJSON), &note.Tags)
        note.Context = context
        notes = append(notes, note)
    }
    
    return notes, nil
}

func (s *SQLiteSRS) UpdateReview(notePath string, quality int) error {
    // 1. Get current SRS data
    currentData, err := s.GetSRSData(notePath)
    if err != nil {
        return err
    }
    
    // 2. Calculate next review using SM-2 algorithm
    sm2 := flotsam.NewSM2Calculator()
    newData, err := sm2.ProcessReview(currentData, flotsam.Quality(quality))
    if err != nil {
        return err
    }
    
    // 3. Update database
    query := `
        UPDATE srs_notes 
        SET quality = ?, interval_days = ?, next_due = ?, last_reviewed = datetime('now'),
            ease_factor = ?, reviews_count = reviews_count + 1
        WHERE path = ? AND context = ?
    `
    
    _, err = s.db.Exec(query, newData.Quality, newData.Interval, 
                      time.Unix(newData.Due, 0), newData.Easiness, notePath, s.context)
    return err
}
```

## CLI Command Implementation

### `vice flotsam list`

```bash
# Delegates to zk with vice-specific tag filtering
$ vice flotsam list
# Executes: zk list --tag "vice:srs" --format json | vice-srs-enhance

# With additional filtering  
$ vice flotsam list --tag important --due-today
# Executes: zk list --tag "vice:srs AND important" --format json | vice-srs-filter --due-today
```

```go
func (cmd *FlotsamListCmd) Run() error {
    // 1. Build zk query with vice-specific tags
    tags := []string{"vice:srs"}
    if cmd.UserTags != "" {
        tags = append(tags, cmd.UserTags)
    }
    
    // 2. Query zk for matching notes
    notes, err := cmd.zk.FindByTag(tags...)
    if err != nil {
        return fmt.Errorf("failed to list notes: %w", err)
    }
    
    // 3. Enhance with SRS data if requested
    if cmd.ShowSRS {
        for i, note := range notes {
            srsData, _ := cmd.srs.GetSRSData(note.Path)
            notes[i].SRSData = srsData
        }
    }
    
    // 4. Apply vice-specific filtering
    if cmd.DueOnly {
        notes = cmd.filterDueNotes(notes)
    }
    
    // 5. Format and display
    return cmd.formatOutput(notes)
}
```

### `vice flotsam due`

```bash
# Queries SRS database directly
$ vice flotsam due --today
# Shows notes due today with zk metadata

$ vice flotsam due --overdue --interactive
# Shows overdue notes with fzf selection
```

```go
func (cmd *FlotsamDueCmd) Run() error {
    // 1. Query SRS database for due notes
    dueNotes, err := cmd.srs.GetDueNotes(cmd.context)
    if err != nil {
        return err
    }
    
    // 2. Filter by date range
    filteredNotes := cmd.filterByDate(dueNotes)
    
    // 3. Interactive selection if requested
    if cmd.Interactive {
        selected, err := cmd.selectInteractive(filteredNotes)
        if err != nil {
            return err
        }
        filteredNotes = selected
    }
    
    // 4. Display results
    return cmd.formatOutput(filteredNotes)
}
```

### `vice flotsam edit`

```bash
# Delegates to zk for editing
$ vice flotsam edit concept-note.md
# Executes: zk edit concept-note.md

$ vice flotsam edit --due-today
# Finds due notes, passes paths to zk edit
```

```go
func (cmd *FlotsamEditCmd) Run() error {
    var paths []string
    
    if cmd.DueToday {
        // 1. Get due notes from SRS
        dueNotes, err := cmd.srs.GetDueNotes(cmd.context)
        if err != nil {
            return err
        }
        
        // 2. Extract file paths
        for _, note := range dueNotes {
            paths = append(paths, note.Path)
        }
    } else {
        // Use provided paths
        paths = cmd.Paths
    }
    
    // 3. Delegate to zk for editing
    return cmd.zk.Edit(paths...)
}
```

## Configuration Management

### User Configuration

```toml
# vice config.toml
[flotsam]
zk_binary = "zk"
zk_flags = ["--no-input", "--quiet"]
zk_editor = "nvim"
zk_fzf_options = "--height 50% --border"

[flotsam.srs]
database_path = ".vice/flotsam.db"
cache_ttl = "1h"
sync_on_startup = true

[flotsam.zk_config]
# Options written to .zk/config.toml during init
filename_template = "{{slug title}}"
template = "flotsam.md"
default_tags = ["vice:srs", "flotsam"]
```

### ZK Notebook Configuration

```toml
# .zk/config.toml (written by vice init)
[note]
filename = "{{slug title}}"
extension = "md"
template = "flotsam.md"
default-title = "New Flotsam Note"

[group.flotsam]
paths = ["flotsam"]

[group.flotsam.note]
filename = "{{slug title}}" 
template = "flotsam.md"
default-title = "New Flotsam Note"

[alias]
# Vice-specific aliases
due = "list --tag 'vice:srs' --format json"
srs = "list --tag 'vice:srs'"
```

## Cache Strategy & Performance

### mtime-based Cache Invalidation

```go
type CacheManager struct {
    srs     SRSDatabase
    zkDir   string
    cacheDB *sql.DB
}

func (c *CacheManager) ValidateCache() error {
    // 1. Get last cache update time
    lastUpdate, err := c.getLastUpdateTime()
    if err != nil {
        return err
    }
    
    // 2. Check if flotsam directory modified
    flotsamDir := filepath.Join(c.zkDir, "flotsam")
    dirInfo, err := os.Stat(flotsamDir)
    if err != nil {
        return err
    }
    
    // 3. Invalidate cache if directory newer
    if dirInfo.ModTime().After(lastUpdate) {
        return c.refreshCache()
    }
    
    return nil
}

func (c *CacheManager) refreshCache() error {
    // 1. Scan file system for changes
    files, err := filepath.Glob(filepath.Join(c.zkDir, "flotsam", "*.md"))
    if err != nil {
        return err
    }
    
    // 2. Update cache for modified files
    for _, file := range files {
        if c.isFileModified(file) {
            if err := c.updateCacheEntry(file); err != nil {
                return err
            }
        }
    }
    
    // 3. Update cache timestamp
    return c.updateCacheTimestamp()
}
```

### Combined Query Performance

```go
// Fast query combining zk metadata with SRS data
func (o *FlotsamOrchestrator) FindNotesWithSRS(filters ...string) ([]SRSNote, error) {
    // 1. Check cache validity
    if err := o.cache.ValidateCache(); err != nil {
        return nil, err
    }
    
    // 2. Query combined cache table
    query := `
        SELECT n.path, n.title, n.tags, s.next_due, s.quality, s.interval_days
        FROM note_cache n
        JOIN srs_notes s ON n.path = s.path
        WHERE n.context = ? AND s.next_due <= datetime('now')
        ORDER BY s.next_due ASC
    `
    
    // 3. Fallback to separate queries if cache miss
    return o.queryWithFallback(query, filters...)
}
```

## Error Handling & Recovery

### Dependency Checking

```go
func (cmd *DoctorCmd) CheckDependencies() error {
    checks := []struct {
        name string
        check func() error
    }{
        {"zk binary", cmd.checkZKBinary},
        {"zk notebook", cmd.checkZKNotebook},
        {"SRS database", cmd.checkSRSDatabase},
        {"file permissions", cmd.checkFilePermissions},
    }
    
    var errors []string
    for _, check := range checks {
        if err := check.check(); err != nil {
            errors = append(errors, fmt.Sprintf("%s: %v", check.name, err))
        } else {
            fmt.Printf("✓ %s: OK\n", check.name)
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("dependency check failed:\n%s", strings.Join(errors, "\n"))
    }
    
    return nil
}

func (cmd *DoctorCmd) checkZKBinary() error {
    if _, err := exec.LookPath("zk"); err != nil {
        return fmt.Errorf("zk binary not found. Install with: go install github.com/zk-org/zk@latest")
    }
    return nil
}
```

### Graceful Degradation

```go
func (o *FlotsamOrchestrator) ListNotes(filters ...string) ([]Note, error) {
    if o.zk.IsAvailable() {
        // Full functionality with zk
        return o.zk.List(filters...)
    } else {
        // Fallback to basic file system scanning
        fmt.Println("Warning: zk not available, using basic file scanning")
        return o.scanFileSystem(filters...)
    }
}
```

## Testing Strategy

### Integration Tests

```go
func TestZKIntegration(t *testing.T) {
    // 1. Set up test notebook
    notebook := setupTestNotebook(t)
    defer notebook.Cleanup()
    
    // 2. Create test notes
    zk := NewZKShell(notebook.Path)
    notePath, err := zk.Create("Test Note", "flotsam.md")
    require.NoError(t, err)
    
    // 3. Test tag filtering
    notes, err := zk.FindByTag("vice:srs")
    require.NoError(t, err)
    assert.Contains(t, notes, notePath)
    
    // 4. Test SRS integration
    srs := NewSQLiteSRS(notebook.Path)
    err = srs.UpdateReview(notePath, 4)
    require.NoError(t, err)
    
    dueNotes, err := srs.GetDueNotes(notebook.Context)
    require.NoError(t, err)
    assert.NotEmpty(t, dueNotes)
}
```

### Mock Implementation

```go
type MockZKTool struct {
    notes []Note
    calls []string
}

func (m *MockZKTool) List(filters ...string) ([]Note, error) {
    m.calls = append(m.calls, "list:"+strings.Join(filters, ","))
    return m.notes, nil
}

func (m *MockZKTool) Edit(paths ...string) error {
    m.calls = append(m.calls, "edit:"+strings.Join(paths, ","))
    return nil
}

func TestFlotsamOrchestrator(t *testing.T) {
    mockZK := &MockZKTool{}
    mockSRS := &MockSRSDatabase{}
    
    orchestrator := &FlotsamOrchestrator{
        zk:  mockZK,
        srs: mockSRS,
    }
    
    err := orchestrator.EditDueNotes()
    require.NoError(t, err)
    
    assert.Contains(t, mockZK.calls, "edit:")
    assert.Contains(t, mockSRS.calls, "GetDueNotes")
}
```

## Migration Benefits

### Immediate Benefits
- **Reduced complexity**: ~500 lines of shell-out code vs ~2000 lines of repository abstraction
- **Proven algorithms**: Leverage zk's battle-tested search, linking, and editor integration
- **User familiarity**: Users can use zk commands directly for advanced operations
- **Extensibility**: Easy to add new tool integrations (remind, taskwarrior, etc.)

### Strategic Benefits
- **Tool ecosystem**: Access to entire Unix productivity tool ecosystem
- **Maintenance burden**: Less code to maintain, fewer bugs to fix
- **Performance scaling**: zk's SQLite FTS handles large note collections efficiently
- **Future-proofing**: Tool orchestration architecture enables rich workflow automation

## Conclusion

The Unix interop architecture transforms vice from a monolithic application into a **productivity orchestration platform**. By delegating complex operations to specialized tools while maintaining vice-specific functionality through targeted integrations, we achieve:

1. **Reduced implementation complexity** while **increasing capability**
2. **Proven, battle-tested algorithms** for search, linking, and file management
3. **Extensible architecture** that can grow with user needs
4. **Clear separation of concerns** between tool orchestration and domain logic

This architecture provides a **sustainable path forward** for vice's evolution while **preserving the option** to fall back to coupled integration if Unix interop proves insufficient.