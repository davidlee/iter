# ADR-006: Flotsam Context Isolation Model

**Status**: Accepted

**Date**: 2025-07-17

## Related Reading

**Related ADRs**: 
 - [ADR-002: Flotsam Files-First Architecture](/doc/decisions/ADR-002-flotsam-files-first-architecture.md) - Source of truth strategy enabling context boundaries
 - [ADR-004: Flotsam SQLite Cache Strategy](/doc/decisions/ADR-004-flotsam-sqlite-cache-strategy.md) - Context-aware cache placement implementing this isolation
 - [ADR-003: ZK-go-srs Integration Strategy](/doc/decisions/ADR-003-zk-gosrs-integration-strategy.md) - Component integration requiring context scoping

**Related Specifications**: 
 - [File Paths & Runtime Environment](/doc/specifications/file_paths_runtime_env.md) - T028 Repository Pattern and ViceEnv context management
 - [Flotsam Package Documentation](/doc/specifications/flotsam.md) - Integration with context-aware operations

**Related Tasks**: 
 - [T028] - Repository Pattern providing context isolation foundation
 - [T027/3.1] - Repository integration requiring context scoping design
 - [T027/4.2] - Context-scoped link resolution and backlink computation

## Context

Flotsam operates within Vice's context system where users can maintain completely isolated habit tracking environments (e.g., "personal" vs "work"). This context isolation must extend to flotsam notes while supporting ZK notebook interoperability that may exist outside Vice's context structure.

### Integration Requirements

1. **Vice Context Integration**: Flotsam must respect Vice's context boundaries and data isolation
2. **ZK Notebook Compatibility**: Must work with existing ZK notebooks that predate Vice contexts
3. **Link Resolution Scoping**: Wiki links should resolve within appropriate boundaries
4. **Data Isolation**: Each context maintains separate note collections and SRS state
5. **Performance Optimization**: Context-aware caching without cross-contamination

### Design Challenges

#### Context Boundary Definition
- **Vice Contexts**: Well-defined through `$VICE_DATA/{context}/` directories
- **ZK Notebooks**: May exist anywhere in filesystem, notebook scope defined by `.zk/` directory
- **Hybrid Scenarios**: ZK notebooks that become Vice contexts, or Vice contexts with ZK integration

#### Link Resolution Scope
- **Intra-Context**: Links between notes within same context/notebook
- **Cross-Context**: Whether to allow links across context boundaries
- **External Links**: Links to notes outside any context system

#### Cache Isolation
- **Context Separation**: Prevent cache contamination between contexts
- **ZK Integration**: Respect ZK notebook boundaries when adding cache tables
- **Performance**: Efficient context switching without cache invalidation overhead

### Design Options Considered

#### Option A: Strict Context Isolation
- **Scope**: Only resolve links within current Vice context
- **Pros**: Clear boundaries, no cross-contamination, predictable behavior
- **Cons**: Inflexible for users with related contexts, breaks some ZK workflows

#### Option B: Configurable Scope Resolution
- **Scope**: User-configurable link resolution scope (context-only, global, hybrid)
- **Pros**: Maximum flexibility, supports diverse workflows
- **Cons**: Complex configuration, potential for confusion, inconsistent behavior

#### Option C: Hybrid Context Bridging
- **Scope**: Smart detection of context boundaries with bridge mechanisms
- **Pros**: Intelligent defaults, supports both Vice and ZK patterns
- **Cons**: Complex implementation, edge cases in boundary detection

## Decision

**We choose Hybrid Context Bridging (Option C)** with intelligent context boundary detection and bridge mechanisms that support both Vice context isolation and ZK notebook workflows.

### Context Isolation Strategy:

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        FLOTSAM CONTEXT ISOLATION                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐              │
│  │ Vice Contexts   │    │ ZK Notebooks    │    │ Bridge Mechanism│              │
│  │                 │    │                 │    │                 │              │
│  │ $VICE_DATA/     │    │ Any directory   │    │ Context         │              │
│  │ {context}/      │────│ with .zk/       │────│ Detection &     │              │
│  │ flotsam/        │    │ structure       │    │ Scope Rules     │              │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘              │
│           │                       │                       │                     │
│           ▼                       ▼                       ▼                     │
│  ┌─────────────────────────────────────────────────────────────────┐            │
│  │                  CONTEXT BOUNDARY DETECTION                     │            │
│  │                                                                 │            │
│  │  IF Vice context      → Scope to $VICE_DATA/{context}/flotsam  │            │
│  │  IF ZK notebook       → Scope to notebook root directory       │            │
│  │  IF Hybrid            → Respect both boundaries with bridges    │            │
│  │  IF Unknown           → Default to current directory scope      │            │
│  └─────────────────────────────────────────────────────────────────┘            │
│                                  │                                              │
│                                  ▼                                              │
│  ┌─────────────────────────────────────────────────────────────────┐            │
│  │                    SCOPED OPERATIONS                            │            │
│  │                                                                 │            │
│  │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐       │            │
│  │  │ Note Discovery│  │ Link Resolution│  │ Cache Isolation│       │            │
│  │  │ (within scope)│  │ (scoped search)│  │ (per context) │       │            │
│  │  └───────────────┘  └───────────────┘  └───────────────┘       │            │
│  └─────────────────────────────────────────────────────────────────┘            │
│                                                                                 │
│  Data Flow: Context Detection → Scope Definition → Operation Isolation        │
│  Cache Flow: Context ID → Isolated Cache Tables → Scoped Queries              │
│  Link Flow: Parse Links → Resolve within Scope → Build Scoped Backlinks       │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Key Design Principles:

#### 1. Intelligent Context Detection
```go
type ContextScope struct {
    Type        ContextType  // Vice, ZK, Hybrid, Unknown
    ID          string       // Unique context identifier
    RootPath    string       // Context boundary directory
    NotePaths   []string     // Directories containing notes
    CacheDBPath string       // Context-specific cache database
}

type ContextType int
const (
    ViceContext ContextType = iota  // $VICE_DATA/{context}/ structure
    ZKNotebook                      // Directory with .zk/ subdirectory
    HybridContext                   // Vice context within ZK notebook
    UnknownContext                  // Fallback to directory-based scope
)
```

#### 2. Context Boundary Rules
- **Vice Contexts**: Scope limited to `$VICE_DATA/{context}/flotsam/` directory
- **ZK Notebooks**: Scope extends to entire notebook directory (containing `.zk/`)
- **Hybrid Contexts**: When Vice context overlaps with ZK notebook, use intersection rules
- **Bridge Contexts**: Special handling for related contexts (e.g., same user, different scopes)

#### 3. Scoped Link Resolution
```go
type LinkResolver struct {
    scope ContextScope
    cache map[string]*FlotsamNote  // scoped note cache
}

func (lr *LinkResolver) ResolveLink(link string) (*FlotsamNote, error) {
    // 1. Search within current context scope
    if note := lr.searchInScope(link); note != nil {
        return note, nil
    }
    
    // 2. Check bridge contexts (if configured)
    if lr.scope.Type == ViceContext {
        return lr.searchBridgeContexts(link)
    }
    
    // 3. Return nil for out-of-scope links
    return nil, ErrLinkNotInScope
}
```

#### 4. Cache Isolation Implementation
```go
func GetContextCacheDB(scope ContextScope) (*sql.DB, error) {
    switch scope.Type {
    case ViceContext:
        // Use $VICE_DATA/{context}/flotsam.db
        dbPath := filepath.Join(scope.RootPath, "flotsam.db")
        return openCacheDB(dbPath, scope.ID)
        
    case ZKNotebook:
        // Add vice_ tables to existing .zk/notebook.db
        dbPath := filepath.Join(scope.RootPath, ".zk", "notebook.db")
        return openCacheDB(dbPath, scope.ID)
        
    case HybridContext:
        // Prefer ZK database, namespace by Vice context
        dbPath := filepath.Join(scope.RootPath, ".zk", "notebook.db")
        return openCacheDB(dbPath, scope.ID)
        
    default:
        // Fallback to in-memory or directory-based cache
        return createInMemoryCache(scope.ID)
    }
}
```

## Consequences

### Positive

- **Flexible Integration**: Supports both Vice contexts and ZK notebooks seamlessly
- **Data Isolation**: Prevents cross-contamination between unrelated contexts
- **ZK Compatibility**: Maintains full compatibility with existing ZK workflows
- **Performance Optimization**: Context-aware caching without scope violations
- **User Experience**: Intelligent defaults with predictable behavior
- **Bridge Support**: Enables workflows that span related contexts
- **Incremental Adoption**: Works with existing setups without migration

### Negative

- **Implementation Complexity**: Context detection and bridge logic adds complexity
- **Edge Case Handling**: Boundary detection may have ambiguous cases
- **Performance Overhead**: Context detection adds startup cost for each operation
- **Configuration Complexity**: Bridge rules may require user configuration
- **Debugging Difficulty**: Scoped operations harder to debug across contexts

### Neutral

- **Migration Strategy**: Existing setups work unchanged, optimization is additive
- **Memory Usage**: Context caching increases memory usage but improves performance
- **Storage Overhead**: Multiple cache databases but isolated and manageable
- **Backup Considerations**: Context isolation simplifies backup and recovery strategies

## Implementation Details

### Context Detection Algorithm

#### 1. Context Scope Discovery
```go
func DetectContextScope(workingDir string, viceEnv *ViceEnv) (*ContextScope, error) {
    scope := &ContextScope{}
    
    // 1. Check if we're in a Vice context
    if isViceContext(workingDir, viceEnv) {
        scope.Type = ViceContext
        scope.ID = viceEnv.Context
        scope.RootPath = viceEnv.ContextData
        scope.NotePaths = []string{filepath.Join(viceEnv.ContextData, "flotsam")}
        scope.CacheDBPath = filepath.Join(viceEnv.ContextData, "flotsam.db")
        
        // Check if this Vice context is within a ZK notebook
        if zkRoot := findZKNotebook(workingDir); zkRoot != "" {
            scope.Type = HybridContext
            scope.CacheDBPath = filepath.Join(zkRoot, ".zk", "notebook.db")
        }
        
        return scope, nil
    }
    
    // 2. Check if we're in a ZK notebook
    if zkRoot := findZKNotebook(workingDir); zkRoot != "" {
        scope.Type = ZKNotebook
        scope.ID = calculateZKContextID(zkRoot)
        scope.RootPath = zkRoot
        scope.NotePaths = []string{zkRoot}
        scope.CacheDBPath = filepath.Join(zkRoot, ".zk", "notebook.db")
        return scope, nil
    }
    
    // 3. Fallback to directory-based context
    scope.Type = UnknownContext
    scope.ID = calculateDirContextID(workingDir)
    scope.RootPath = workingDir
    scope.NotePaths = []string{workingDir}
    scope.CacheDBPath = filepath.Join(workingDir, ".flotsam.db")
    
    return scope, nil
}
```

#### 2. ZK Notebook Detection
```go
func findZKNotebook(startDir string) string {
    dir := startDir
    for {
        zkDir := filepath.Join(dir, ".zk")
        if stat, err := os.Stat(zkDir); err == nil && stat.IsDir() {
            // Verify it's a real ZK notebook by checking for notebook.db
            dbPath := filepath.Join(zkDir, "notebook.db")
            if _, err := os.Stat(dbPath); err == nil {
                return dir
            }
        }
        
        parent := filepath.Dir(dir)
        if parent == dir {
            break // Reached root directory
        }
        dir = parent
    }
    return ""
}
```

#### 3. Context Identifier Generation
```go
func calculateZKContextID(zkRoot string) string {
    // Use notebook path hash for stable context ID
    hash := sha256.Sum256([]byte(zkRoot))
    return fmt.Sprintf("zk_%x", hash[:8])
}

func calculateViceContextID(viceContext string) string {
    return fmt.Sprintf("vice_%s", viceContext)
}

func calculateDirContextID(dir string) string {
    hash := sha256.Sum256([]byte(dir))
    return fmt.Sprintf("dir_%x", hash[:8])
}
```

### Scoped Operations Implementation

#### 1. Note Discovery Within Scope
```go
func (scope *ContextScope) DiscoverNotes() ([]*FlotsamNote, error) {
    var allNotes []*FlotsamNote
    
    for _, notePath := range scope.NotePaths {
        notes, err := scanNotesInDirectory(notePath, scope)
        if err != nil {
            return nil, err
        }
        allNotes = append(allNotes, notes...)
    }
    
    return allNotes, nil
}

func scanNotesInDirectory(dir string, scope *ContextScope) ([]*FlotsamNote, error) {
    return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        // Only process .md files
        if !strings.HasSuffix(path, ".md") {
            return nil
        }
        
        // Skip files outside scope boundaries
        if !scope.containsPath(path) {
            return nil
        }
        
        note, err := parseFlotsamNote(path)
        if err != nil {
            return err // or log and continue
        }
        
        notes = append(notes, note)
        return nil
    })
}
```

#### 2. Context-Aware Link Resolution
```go
type ScopedLinkResolver struct {
    scope     *ContextScope
    noteIndex map[string]*FlotsamNote  // filename -> note lookup
    idIndex   map[string]*FlotsamNote  // note ID -> note lookup
}

func (slr *ScopedLinkResolver) ResolveWikiLink(linkText string) (*FlotsamNote, error) {
    // 1. Try exact filename match within scope
    if note := slr.noteIndex[linkText+".md"]; note != nil {
        return note, nil
    }
    
    // 2. Try note ID match within scope
    if note := slr.idIndex[linkText]; note != nil {
        return note, nil
    }
    
    // 3. Try partial title match within scope
    for _, note := range slr.noteIndex {
        if strings.Contains(strings.ToLower(note.Title), strings.ToLower(linkText)) {
            return note, nil
        }
    }
    
    // 4. Check bridge contexts (if enabled)
    if slr.scope.Type == ViceContext && slr.scope.bridgeEnabled {
        return slr.searchBridgeContexts(linkText)
    }
    
    return nil, ErrLinkNotInScope
}
```

#### 3. Backlink Computation Within Scope
```go
func (slr *ScopedLinkResolver) ComputeBacklinks() error {
    // Clear existing backlinks
    for _, note := range slr.noteIndex {
        note.Backlinks = note.Backlinks[:0]
    }
    
    // Compute backlinks within scope only
    for _, sourceNote := range slr.noteIndex {
        for _, link := range sourceNote.Links {
            targetNote, err := slr.ResolveWikiLink(link)
            if err != nil {
                continue // Skip unresolvable links
            }
            
            // Add backlink only if target is in scope
            if slr.scope.containsNote(targetNote) {
                targetNote.Backlinks = append(targetNote.Backlinks, sourceNote.ID)
            }
        }
    }
    
    return nil
}
```

### Cache Table Isolation

#### 1. Context-Specific Table Names
```sql
-- Vice context tables
CREATE TABLE vice_srs_cache_{context_id} (...);
CREATE TABLE vice_file_cache_{context_id} (...);
CREATE TABLE vice_contexts_{context_id} (...);

-- Example: vice_srs_cache_personal, vice_srs_cache_work
```

#### 2. Context Metadata Tracking
```sql
CREATE TABLE vice_context_registry (
    context_id TEXT PRIMARY KEY,
    context_type TEXT NOT NULL,  -- 'vice', 'zk', 'hybrid', 'unknown'
    root_path TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    last_accessed INTEGER NOT NULL,
    note_count INTEGER DEFAULT 0,
    
    INDEX idx_vice_registry_type (context_type),
    INDEX idx_vice_registry_path (root_path)
);
```

#### 3. Cross-Context Query Prevention
```go
func (db *ContextAwareDB) QuerySRSCache(contextID string, query string) (*sql.Rows, error) {
    // Validate context ID to prevent injection
    if !isValidContextID(contextID) {
        return nil, ErrInvalidContextID
    }
    
    // Ensure query only accesses tables for this context
    tableName := fmt.Sprintf("vice_srs_cache_%s", contextID)
    if !strings.Contains(query, tableName) {
        return nil, ErrCrossContextQuery
    }
    
    return db.Query(query)
}
```

### Bridge Context Support

#### 1. Bridge Configuration
```go
type BridgeConfig struct {
    Enabled        bool               `yaml:"enabled"`
    BridgeContexts []string           `yaml:"bridge_contexts"`  // Related contexts
    LinkResolution BridgeResolution   `yaml:"link_resolution"`
    CacheSharing   bool               `yaml:"cache_sharing"`
}

type BridgeResolution int
const (
    NoBridge BridgeResolution = iota     // No cross-context links
    ReadOnlyBridge                       // Read links from bridge contexts
    FullBridge                           // Read/write links across contexts
)
```

#### 2. Bridge Link Resolution
```go
func (slr *ScopedLinkResolver) searchBridgeContexts(linkText string) (*FlotsamNote, error) {
    if !slr.bridgeConfig.Enabled {
        return nil, ErrBridgeDisabled
    }
    
    for _, bridgeContextID := range slr.bridgeConfig.BridgeContexts {
        bridgeScope, err := loadContextScope(bridgeContextID)
        if err != nil {
            continue
        }
        
        bridgeResolver := NewScopedLinkResolver(bridgeScope)
        if note, err := bridgeResolver.ResolveWikiLink(linkText); err == nil {
            return note, nil
        }
    }
    
    return nil, ErrLinkNotFound
}
```

### Performance Optimizations

#### 1. Context Caching
```go
type ContextManager struct {
    scopeCache    map[string]*ContextScope     // directory -> scope cache
    resolverCache map[string]*ScopedLinkResolver // context -> resolver cache
    cacheTimeout  time.Duration
}

func (cm *ContextManager) GetScope(workingDir string) (*ContextScope, error) {
    // Check cache first
    if scope := cm.scopeCache[workingDir]; scope != nil {
        if time.Since(scope.CacheTime) < cm.cacheTimeout {
            return scope, nil
        }
    }
    
    // Detect and cache scope
    scope, err := DetectContextScope(workingDir, cm.viceEnv)
    if err != nil {
        return nil, err
    }
    
    scope.CacheTime = time.Now()
    cm.scopeCache[workingDir] = scope
    return scope, nil
}
```

#### 2. Lazy Link Index Building
```go
func (slr *ScopedLinkResolver) EnsureIndexes() error {
    if slr.indexesBuilt {
        return nil
    }
    
    notes, err := slr.scope.DiscoverNotes()
    if err != nil {
        return err
    }
    
    // Build filename and ID indexes
    slr.noteIndex = make(map[string]*FlotsamNote)
    slr.idIndex = make(map[string]*FlotsamNote)
    
    for _, note := range notes {
        filename := filepath.Base(note.FilePath)
        slr.noteIndex[filename] = note
        slr.idIndex[note.ID] = note
    }
    
    slr.indexesBuilt = true
    return nil
}
```

---
*ADR format based on [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)*