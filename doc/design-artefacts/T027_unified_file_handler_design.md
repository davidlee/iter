# Unified File Handler Design for Vice

## Executive Summary

This document designs a unified file handler for Vice that ensures consistency between markdown files and SQLite cache, taking inspiration from ZK's robust file synchronization approach. The handler supports both .md and .yml files with atomic operations and cache invalidation.

## ZK File Handling Analysis

### Key Insights from ZK Architecture

1. **Demand-Driven Indexing**: ZK uses comparison-based indexing rather than active file watching
2. **Dual Change Detection**: Timestamp + SHA256 checksum for robust change detection
3. **Transactional Safety**: All database operations wrapped in transactions
4. **Incremental Updates**: Only changed files processed during normal indexing
5. **Graceful Error Handling**: System continues working even if parsing fails
6. **Metadata Extensibility**: JSON metadata storage easily accommodates extensions

### ZK Parsing Pipeline

```
File → Frontmatter Detection → YAML Parsing → Metadata Normalization → 
Content Parsing → Database Transaction → SQLite Storage → Index Update
```

**Key Components**:
- **Change Detection**: `paths.Diff()` compares filesystem vs database timestamps
- **Checksum Validation**: SHA256 checksums detect content changes
- **Frontmatter Parsing**: Robust YAML parsing with error recovery
- **Atomic Operations**: `db.WithTransaction()` ensures consistency

## Unified File Handler Architecture

### Core Design Principles

1. **Files as Source of Truth**: All authoritative data in text files (.md, .yml)
2. **SQLite as Performance Cache**: Database rebuilt from files when needed
3. **Atomic Operations**: File writes followed by cache updates in transactions
4. **Multi-Format Support**: Handle both markdown and YAML files
5. **ZK Compatibility**: Work alongside ZK's indexing without conflicts
6. **Incremental Updates**: Process only changed files for performance

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    Unified File Handler                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐          │
│  │   .md       │    │   .yml      │    │   .json     │          │
│  │   Files     │    │   Files     │    │   Files     │          │
│  └─────────────┘    └─────────────┘    └─────────────┘          │
│         │                  │                  │                 │
│         └──────────────────┼──────────────────┘                 │
│                           │                                    │
│  ┌─────────────────────────▼─────────────────────────┐          │
│  │            File Change Detection                  │          │
│  │         (Timestamp + Checksum)                   │          │
│  └─────────────────────────┬─────────────────────────┘          │
│                           │                                    │
│  ┌─────────────────────────▼─────────────────────────┐          │
│  │            Content Parsing                        │          │
│  │      (Frontmatter + Body + Validation)           │          │
│  └─────────────────────────┬─────────────────────────┘          │
│                           │                                    │
│  ┌─────────────────────────▼─────────────────────────┐          │
│  │           Cache Synchronization                   │          │
│  │        (Atomic File Write + SQLite Update)       │          │
│  └─────────────────────────┬─────────────────────────┘          │
│                           │                                    │
│  ┌─────────────────────────▼─────────────────────────┐          │
│  │             SQLite Cache                          │          │
│  │    (SRS data, indexes, performance queries)       │          │
│  └─────────────────────────────────────────────────────┘          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Implementation Strategy

### 1. File Change Detection System

**Inspired by ZK's `paths.Diff()` approach**:

```go
type FileMetadata struct {
    Path     string
    Modified time.Time
    Checksum string
    Type     FileType // .md, .yml, .json
}

type FileHandler struct {
    basePath     string
    db          *sql.DB
    logger      *slog.Logger
    parsers     map[FileType]ContentParser
}

// Main synchronization method
func (h *FileHandler) Sync(force bool) error {
    // Get filesystem state
    fsFiles, err := h.walkFiles()
    if err != nil {
        return err
    }
    
    // Get database state
    dbFiles, err := h.getCachedFiles()
    if err != nil {
        return err
    }
    
    // Compare and process changes
    return h.processChanges(fsFiles, dbFiles, force)
}
```

### 2. Multi-Format Content Parsing

**Extensible parser system**:

```go
type ContentParser interface {
    ParseContent(filepath string, content []byte) (*ParsedContent, error)
}

type ParsedContent struct {
    Frontmatter map[string]interface{}
    Body        string
    Checksum    string
    Links       []Link
    Tags        []string
}

// Markdown parser (leverages existing flotsam code)
type MarkdownParser struct {
    linkExtractor *LinkExtractor
}

func (p *MarkdownParser) ParseContent(filepath string, content []byte) (*ParsedContent, error) {
    // Use existing zk_parser.go logic
    frontmatter, body, err := parseFrontmatter(string(content))
    if err != nil {
        return nil, err
    }
    
    // Extract links using existing zk_links.go
    links := ExtractLinks(body)
    
    return &ParsedContent{
        Frontmatter: frontmatter,
        Body:        body,
        Checksum:    fmt.Sprintf("%x", sha256.Sum256(content)),
        Links:       links,
    }, nil
}

// YAML parser for config files
type YAMLParser struct{}

func (p *YAMLParser) ParseContent(filepath string, content []byte) (*ParsedContent, error) {
    var data map[string]interface{}
    if err := yaml.Unmarshal(content, &data); err != nil {
        return nil, err
    }
    
    return &ParsedContent{
        Frontmatter: data,
        Body:        string(content),
        Checksum:    fmt.Sprintf("%x", sha256.Sum256(content)),
    }, nil
}
```

### 3. Atomic Write Operations

**Inspired by ZK's transaction handling**:

```go
func (h *FileHandler) WriteFile(filepath string, content []byte) error {
    // Atomic file write with temporary file
    tempPath := filepath + ".tmp"
    
    if err := os.WriteFile(tempPath, content, 0600); err != nil {
        return err
    }
    
    // Parse content for cache update
    parsed, err := h.parseFile(filepath, content)
    if err != nil {
        os.Remove(tempPath)
        return err
    }
    
    // Atomic database transaction
    return h.db.WithTransaction(func(tx *sql.Tx) error {
        // Move temp file to final location
        if err := os.Rename(tempPath, filepath); err != nil {
            return err
        }
        
        // Update cache
        return h.updateCache(tx, filepath, parsed)
    })
}
```

### 4. Cache Synchronization Strategy

**SQLite cache schema**:

```sql
-- File metadata cache
CREATE TABLE vice_file_cache (
    path TEXT PRIMARY KEY,
    modified_timestamp INTEGER NOT NULL,
    checksum TEXT NOT NULL,
    file_type TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    last_processed INTEGER DEFAULT (strftime('%s', 'now'))
);

-- SRS data cache (rebuilt from frontmatter)
CREATE TABLE vice_srs_cache (
    note_id INTEGER PRIMARY KEY REFERENCES notes(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    context TEXT NOT NULL DEFAULT 'default',
    easiness REAL DEFAULT 2.5,
    consecutive_correct INTEGER DEFAULT 0,
    due_timestamp INTEGER,
    total_reviews INTEGER DEFAULT 0,
    review_history TEXT, -- JSON array
    file_checksum TEXT NOT NULL,
    updated_at INTEGER DEFAULT (strftime('%s', 'now'))
);

-- Indexes for performance
CREATE INDEX idx_vice_file_checksum ON vice_file_cache(checksum);
CREATE INDEX idx_vice_srs_due ON vice_srs_cache(due_timestamp);
CREATE INDEX idx_vice_srs_context ON vice_srs_cache(context);
```

### 5. ZK Integration Strategy

**Co-existence with ZK indexing**:

```go
type ZKIntegration struct {
    zkDB        *sql.DB
    viceHandler *FileHandler
}

func (z *ZKIntegration) SyncWithZK() error {
    // Trigger ZK indexing first
    if err := z.runZKIndex(); err != nil {
        z.logger.Warn("ZK indexing failed", "error", err)
    }
    
    // Then sync Vice cache
    return z.viceHandler.Sync(false)
}

func (z *ZKIntegration) runZKIndex() error {
    // Use ZK's indexing logic or shell out to `zk index`
    cmd := exec.Command("zk", "index")
    cmd.Dir = z.viceHandler.basePath
    return cmd.Run()
}
```

## Usage Patterns

### 1. Vice Operations

```go
// Writing a flotsam note
func (s *FlotsamService) UpdateNote(id string, content string, srsData SRSData) error {
    // Update frontmatter with SRS data
    frontmatter := map[string]interface{}{
        "id":    id,
        "title": extractTitle(content),
        "vice": map[string]interface{}{
            "srs": srsData,
            "context": s.context,
        },
    }
    
    // Generate markdown with frontmatter
    markdown := generateMarkdown(frontmatter, content)
    
    // Atomic write + cache update
    return s.fileHandler.WriteFile(filepath, []byte(markdown))
}

// Querying SRS data
func (s *FlotsamService) GetDueCards(context string) ([]SRSCard, error) {
    // Fast query from cache
    query := `
        SELECT note_id, easiness, consecutive_correct, due_timestamp, review_history
        FROM vice_srs_cache 
        WHERE context = ? AND due_timestamp <= ?
        ORDER BY due_timestamp
    `
    // Execute query and return results
}
```

### 2. Background Synchronization

```go
// Periodic sync service
type SyncService struct {
    fileHandler *FileHandler
    interval    time.Duration
}

func (s *SyncService) Start() {
    ticker := time.NewTicker(s.interval)
    defer ticker.Stop()
    
    for range ticker.C {
        if err := s.fileHandler.Sync(false); err != nil {
            s.logger.Error("sync failed", "error", err)
        }
    }
}
```

## Error Handling & Recovery

### 1. Graceful Degradation

**Cache corruption recovery**:
```go
func (h *FileHandler) RepairCache() error {
    // Drop and recreate cache tables
    if err := h.dropCacheTables(); err != nil {
        return err
    }
    
    if err := h.createCacheTables(); err != nil {
        return err
    }
    
    // Full rebuild from files
    return h.Sync(true)
}
```

### 2. Conflict Resolution

**Handle concurrent modifications**:
```go
func (h *FileHandler) handleConflict(filepath string, fsChecksum, dbChecksum string) error {
    // Log conflict
    h.logger.Warn("checksum mismatch detected", 
        "file", filepath, 
        "fs_checksum", fsChecksum,
        "db_checksum", dbChecksum)
    
    // Filesystem wins - rebuild cache from file
    content, err := os.ReadFile(filepath)
    if err != nil {
        return err
    }
    
    parsed, err := h.parseFile(filepath, content)
    if err != nil {
        return err
    }
    
    return h.updateCache(nil, filepath, parsed)
}
```

## Performance Optimizations

### 1. Incremental Processing

- **Timestamp comparison**: Only process files newer than cache
- **Checksum validation**: Skip unchanged files even if timestamp differs
- **Batch operations**: Process multiple files in single transaction

### 2. Index Strategy

- **Composite indexes**: Optimize common query patterns
- **Partial indexes**: Index only active/due SRS cards
- **Query optimization**: Use EXPLAIN QUERY PLAN for performance tuning

## Testing Strategy

### 1. Unit Tests

- **Parser tests**: Each file format parser
- **Cache consistency**: Verify file-cache synchronization
- **Error handling**: Test recovery from various failure modes

### 2. Integration Tests

- **ZK compatibility**: Verify ZK still works after Vice operations
- **Concurrent access**: Test file locking and transaction safety
- **Large datasets**: Performance testing with many files

### 3. End-to-End Tests

- **Round-trip**: File write → cache update → query → verify
- **Recovery**: Corrupt cache → repair → verify consistency
- **Migration**: Existing files → cache build → verify completeness

## Implementation Plan

### Phase 1: Core File Handler
1. Implement basic file change detection
2. Create markdown parser using existing flotsam code
3. Build atomic write operations
4. Add SQLite cache synchronization

### Phase 2: ZK Integration
1. Test ZK compatibility with additional tables
2. Implement ZK-aware synchronization
3. Add conflict resolution mechanisms
4. Verify round-trip operations

### Phase 3: Multi-Format Support
1. Add YAML parser for config files
2. Implement JSON parser for data files
3. Extend cache schema for different file types
4. Add format-specific validation

### Phase 4: Performance & Reliability
1. Add comprehensive error handling
2. Implement cache repair mechanisms
3. Optimize query performance
4. Add monitoring and metrics

## Conclusion

This unified file handler design provides:
- **Consistency**: Atomic operations ensure file-cache synchronization
- **Performance**: SQLite cache for fast queries
- **Reliability**: Robust error handling and recovery
- **Extensibility**: Support for multiple file formats
- **Compatibility**: Co-existence with ZK indexing

The design leverages ZK's proven patterns while extending them for Vice's specific needs, ensuring both systems can operate safely on the same files.