# ADR-004: Flotsam SQLite Cache Strategy

**Status**: Accepted

**Date**: 2025-07-17

## Related Reading

**Related ADRs**: 
 - [ADR-002: Flotsam Files-First Architecture](/doc/decisions/ADR-002-flotsam-files-first-architecture.md) - Source of truth strategy that enables this cache
 - [ADR-003: ZK-go-srs Integration Strategy](/doc/decisions/ADR-003-zk-gosrs-integration-strategy.md) - Component integration requiring performance optimization
 - (Future) ADR-006: Context Isolation Model - How cache isolation works with Vice contexts

**Related Specifications**: 
 - [Flotsam Package Documentation](/doc/specifications/flotsam.md) - Performance considerations and cache architecture
 - [ZK Interoperability Design](/doc/design-artefacts/T027_zk_interoperability_design.md) - ZK database compatibility

**Related Tasks**: 
 - [T027/1.3.1] - ZK interoperability research validating cache coexistence
 - [T027/1.3.3] - Integration testing demonstrating performance gains
 - [T028] - Repository Pattern providing context-aware storage foundation

## Context

The flotsam system's files-first architecture provides excellent data portability and ZK compatibility, but introduces performance challenges for SRS operations:

### Performance Requirements
1. **SRS Queries**: Fast identification of due cards across large note collections
2. **Bulk Operations**: Efficient processing during batch review sessions
3. **Search Operations**: Quick filtering by tags, due dates, and review statistics
4. **Link Resolution**: Fast backlink computation for large note networks

### Files-First Limitations
- **Sequential Scanning**: Must parse all files to find due cards
- **YAML Overhead**: Frontmatter parsing for every query operation
- **No Indexing**: File-based storage lacks database query optimization
- **Complex Queries**: Join-like operations require in-memory processing

### ZK Integration Constraints
- **Non-Destructive**: Cannot modify existing ZK notebook structure
- **Coexistence**: Must work alongside ZK's own SQLite database
- **Rollback**: Must be completely removable without affecting ZK functionality

### Design Options Considered

#### Option A: No Cache (Pure Files)
- **Pros**: Simplest implementation, perfect consistency
- **Cons**: Poor performance for large collections, no query optimization

#### Option B: Separate Cache Database
- **Pros**: Clean separation, easy management
- **Cons**: Additional database to manage, context isolation complexity

#### Option C: Hybrid Integration (ZK Database Extension)
- **Pros**: Leverages existing infrastructure, context-aware placement
- **Cons**: More complex integration, requires careful ZK compatibility

## Decision

**We choose Hybrid Integration (Option C)** with context-aware SQLite cache placement that extends existing database infrastructure while maintaining complete reversibility.

### Cache Strategy:

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         FLOTSAM CACHE ARCHITECTURE                             │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐              │
│  │ ZK Notebooks    │    │ Vice Contexts   │    │ Cache Tables    │              │
│  │                 │    │                 │    │                 │              │
│  │ ~/notes/.zk/    │    │ $VICE_DATA/     │    │ vice_srs_cache  │              │
│  │ notebook.db     │────│ {ctx}/flotsam.db│────│ vice_file_cache │              │
│  │ (existing)      │    │ (if needed)     │    │ vice_contexts   │              │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘              │
│           │                       │                       │                     │
│           ▼                       ▼                       ▼                     │
│  ┌─────────────────────────────────────────────────────────────────┐            │
│  │                    CACHE PLACEMENT LOGIC                        │            │
│  │                                                                 │            │
│  │  IF ZK notebook detected → Add tables to existing notebook.db  │            │
│  │  ELSE → Create flotsam.db in context directory                 │            │
│  │                                                                 │            │
│  │  Cache tables are prefixed with 'vice_' and ignored by ZK      │            │
│  └─────────────────────────────────────────────────────────────────┘            │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Key Design Principles:

#### 1. Context-Aware Placement
- **ZK Notebooks**: Add Vice tables to existing `.zk/notebook.db`
- **Vice Contexts**: Create `flotsam.db` in `$VICE_DATA/{context}/` directory
- **Dynamic Detection**: Automatically detect ZK notebooks vs pure Vice contexts
- **Isolation**: Each context maintains separate cache database

#### 2. Non-Destructive Integration
- **Table Prefixing**: All Vice tables prefixed with `vice_` for clear separation
- **ZK Ignorance**: ZK ignores unknown tables, confirmed by compatibility testing
- **Clean Removal**: Drop all `vice_*` tables to completely remove Vice functionality
- **No Schema Changes**: Never modify existing ZK tables or structure

#### 3. Performance Optimization
- **Selective Caching**: Cache only performance-critical data (SRS scheduling, file metadata)
- **Change Detection**: Timestamp + SHA256 checksum for efficient cache invalidation
- **Atomic Updates**: File writes followed by cache updates in transactions
- **Query Optimization**: Indexes optimized for SRS and search operations

## Consequences

### Positive

- **Dramatic Performance Gains**: Sub-millisecond SRS queries vs full file scans
- **ZK Compatibility**: Proven to work without affecting ZK functionality
- **Context Isolation**: Each context maintains independent cache state
- **Incremental Adoption**: Can start without cache, add later for performance
- **Infrastructure Reuse**: Leverages existing database when available
- **Complete Reversibility**: Drop tables to remove all Vice functionality
- **Query Flexibility**: Enables complex SRS statistics and reporting

### Negative

- **Cache Invalidation Complexity**: Must carefully coordinate file and cache updates
- **Consistency Challenges**: Potential for cache-file drift if not properly managed
- **Storage Overhead**: Duplicates data between files and cache
- **Database Management**: Additional database administration complexity
- **ZK Coupling**: Some coupling with ZK database structure and lifecycle

### Neutral

- **Development Complexity**: More complex than pure files but enables performance
- **Backup Strategy**: Cache is rebuildable but adds to backup considerations
- **Debugging**: Cache state adds another layer to debug during issues
- **Migration**: Existing ZK notebooks gain performance without migration

## Implementation Details

### Cache Schema Design

#### Core SRS Cache Table
```sql
CREATE TABLE vice_srs_cache (
    note_id TEXT PRIMARY KEY,
    file_path TEXT NOT NULL,
    due_timestamp INTEGER NOT NULL,
    easiness REAL NOT NULL,
    consecutive_correct INTEGER NOT NULL,
    total_reviews INTEGER NOT NULL,
    last_reviewed INTEGER,
    file_checksum TEXT NOT NULL,
    cached_at INTEGER NOT NULL,
    
    -- Performance indexes
    INDEX idx_vice_srs_due (due_timestamp),
    INDEX idx_vice_srs_reviews (total_reviews),
    INDEX idx_vice_srs_checksum (file_checksum)
);
```

#### File Metadata Cache Table
```sql
CREATE TABLE vice_file_cache (
    file_path TEXT PRIMARY KEY,
    checksum TEXT NOT NULL,
    modified_at INTEGER NOT NULL,
    parsed_at INTEGER NOT NULL,
    parse_duration_ms INTEGER,
    
    -- Content metrics
    link_count INTEGER NOT NULL,
    tag_count INTEGER NOT NULL,
    word_count INTEGER,
    has_srs BOOLEAN NOT NULL,
    
    -- Performance indexes
    INDEX idx_vice_file_modified (modified_at),
    INDEX idx_vice_file_checksum (checksum),
    INDEX idx_vice_file_srs (has_srs)
);
```

#### Context Management Table
```sql
CREATE TABLE vice_contexts (
    context_id TEXT PRIMARY KEY,
    root_path TEXT NOT NULL,
    is_zk_notebook BOOLEAN NOT NULL,
    created_at INTEGER NOT NULL,
    last_sync INTEGER NOT NULL,
    note_count INTEGER DEFAULT 0,
    
    INDEX idx_vice_context_path (root_path),
    INDEX idx_vice_context_sync (last_sync)
);
```

### Cache Operations Strategy

#### 1. Change Detection Protocol
```go
type ChangeDetector struct {
    checksumCache map[string]string
    timestampCache map[string]time.Time
}

func (cd *ChangeDetector) HasChanged(filePath string) (bool, error) {
    // 1. Check file modification time
    currentMtime := getModTime(filePath)
    if cachedMtime := cd.timestampCache[filePath]; cachedMtime.Before(currentMtime) {
        return true, nil
    }
    
    // 2. Check content checksum for verification
    currentChecksum := calculateSHA256(filePath)
    if cachedChecksum := cd.checksumCache[filePath]; cachedChecksum != currentChecksum {
        return true, nil
    }
    
    return false, nil
}
```

#### 2. Atomic Update Protocol
```go
func UpdateNoteWithCache(note *FlotsamNote, newSRS *SRSData) error {
    tx, err := db.Begin()
    if err != nil { return err }
    defer tx.Rollback()
    
    // 1. Write to file (source of truth)
    if err := writeNoteToFile(note); err != nil {
        return err
    }
    
    // 2. Update cache tables
    if err := updateSRSCache(tx, note.ID, newSRS); err != nil {
        return err
    }
    
    if err := updateFileCache(tx, note.FilePath); err != nil {
        return err
    }
    
    // 3. Commit only if both succeed
    return tx.Commit()
}
```

#### 3. Cache Invalidation Strategy
```go
func SyncCacheWithFiles(contextID string) error {
    // 1. Scan all note files in context
    noteFiles, err := scanNoteFiles(contextID)
    if err != nil { return err }
    
    // 2. Check each file for changes
    for _, filePath := range noteFiles {
        changed, err := hasFileChanged(filePath)
        if err != nil { return err }
        
        if changed {
            // 3. Reparse and update cache
            if err := updateCacheFromFile(filePath); err != nil {
                return err
            }
        }
    }
    
    // 4. Remove cache entries for deleted files
    return removeOrphanedCacheEntries(contextID)
}
```

### Database Placement Logic

#### Context Detection Algorithm
```go
func DetermineCacheLocation(contextPath string) (dbPath string, isZK bool, err error) {
    // 1. Check for ZK notebook
    zkDbPath := filepath.Join(contextPath, ".zk", "notebook.db")
    if fileExists(zkDbPath) {
        return zkDbPath, true, nil
    }
    
    // 2. Check parent directories for ZK notebook
    for parent := filepath.Dir(contextPath); parent != "/"; parent = filepath.Dir(parent) {
        zkDbPath := filepath.Join(parent, ".zk", "notebook.db")
        if fileExists(zkDbPath) {
            return zkDbPath, true, nil
        }
    }
    
    // 3. Use Vice context database
    viceDbPath := filepath.Join(contextPath, "flotsam.db")
    return viceDbPath, false, nil
}
```

#### Table Creation Strategy
```go
func EnsureCacheTables(dbPath string, isZK bool) error {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil { return err }
    defer db.Close()
    
    // Create tables with vice_ prefix
    tables := []string{
        createSRSCacheTableSQL,
        createFileCacheTableSQL,
        createContextsTableSQL,
    }
    
    for _, tableSQL := range tables {
        if _, err := db.Exec(tableSQL); err != nil {
            return fmt.Errorf("creating cache table: %w", err)
        }
    }
    
    return nil
}
```

### Performance Optimizations

#### 1. Query Patterns
```sql
-- Find due cards (most common operation)
SELECT note_id, file_path, due_timestamp 
FROM vice_srs_cache 
WHERE due_timestamp <= ? 
ORDER BY due_timestamp ASC;

-- SRS statistics
SELECT 
    COUNT(*) as total_cards,
    COUNT(CASE WHEN due_timestamp <= ? THEN 1 END) as due_cards,
    AVG(total_reviews) as avg_reviews,
    AVG(easiness) as avg_easiness
FROM vice_srs_cache;

-- Recently modified files (for incremental sync)
SELECT file_path, checksum 
FROM vice_file_cache 
WHERE parsed_at > ?;
```

#### 2. Bulk Operations
```go
func ProcessDueCards(contextID string, maxCards int) ([]*FlotsamNote, error) {
    // 1. Fast cache query for due cards
    dueCardIDs, err := queryDueCards(contextID, maxCards)
    if err != nil { return nil, err }
    
    // 2. Load only due cards from files
    var notes []*FlotsamNote
    for _, cardID := range dueCardIDs {
        note, err := loadNoteFromFile(cardID)
        if err != nil { return nil, err }
        notes = append(notes, note)
    }
    
    return notes, nil
}
```

### Error Recovery Procedures

#### 1. Cache Corruption Recovery
```go
func RecoverCorruptedCache(contextID string) error {
    // 1. Drop all vice tables
    if err := dropViceTables(contextID); err != nil {
        return err
    }
    
    // 2. Recreate table structure
    if err := createCacheTables(contextID); err != nil {
        return err
    }
    
    // 3. Rebuild cache from source files
    return rebuildCacheFromFiles(contextID)
}
```

#### 2. Inconsistency Detection
```go
func ValidateCacheConsistency(contextID string) ([]string, error) {
    var inconsistencies []string
    
    // Check file existence
    cacheEntries, err := getAllCacheEntries(contextID)
    if err != nil { return nil, err }
    
    for _, entry := range cacheEntries {
        if !fileExists(entry.FilePath) {
            inconsistencies = append(inconsistencies, 
                fmt.Sprintf("Missing file: %s", entry.FilePath))
        }
        
        if !checksumMatches(entry.FilePath, entry.Checksum) {
            inconsistencies = append(inconsistencies, 
                fmt.Sprintf("Checksum mismatch: %s", entry.FilePath))
        }
    }
    
    return inconsistencies, nil
}
```

---
*ADR format based on [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)*