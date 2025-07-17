# ADR-002: Flotsam Files-First Architecture

**Status**: Accepted

**Date**: 2025-07-17

## Related Reading

**Related ADRs**: 
 - (Future) ADR-004: SQLite Cache Strategy - Performance caching approach for files-first data
 - (Future) ADR-006: Context Isolation Model - How context scoping interacts with files-first design

**Related Specifications**: 
 - [Flotsam Package Documentation](/doc/specifications/flotsam.md) - Complete API and architecture reference
 - [ZK Interoperability Design](/doc/zk_interoperability_design.md) - ZK notebook compatibility analysis

**Related Tasks**: 
 - [T027/1.3.1] - ZK-Vice Interoperability Research & Design
 - [T027/1.3.3] - Cross-component integration testing validation

## Context

The flotsam note system requires a storage strategy that balances several competing concerns:

1. **Data Portability**: Notes should be readable and transferable without proprietary databases
2. **ZK Compatibility**: Must work seamlessly with existing ZK notebooks without conflicts
3. **Performance**: SRS operations require fast queries for due cards and review scheduling
4. **Vice Integration**: Must integrate with Vice's context system and repository patterns
5. **Data Integrity**: SRS history and metadata must be preserved and recoverable

Two primary approaches were considered:

### Option A: Database-First Architecture
- Store all data (including SRS history) in SQLite database
- Markdown files contain minimal content, reference database records
- Fast queries but requires database migration for portability

### Option B: Files-First Architecture  
- Store all data in markdown files with YAML frontmatter as source of truth
- Optional SQLite performance cache for fast queries
- Cache can be rebuilt from source files if corrupted

## Decision

**We choose Files-First Architecture (Option B)** where individual markdown files with YAML frontmatter serve as the authoritative source of truth, with optional SQLite caching for performance.

### Key Design Principles:

1. **Source of Truth**: All persistent data lives in markdown frontmatter, including complete SRS history
2. **Performance Cache**: SQLite cache tables for fast SRS queries, rebuildable from files
3. **ZK Integration**: Add Vice cache tables to existing ZK databases (ZK ignores unknown tables)
4. **Atomic Operations**: File writes followed by cache updates in transactions
5. **Recovery Strategy**: Drop cache to completely remove Vice functionality

### Data Storage Pattern:

```yaml
---
id: abc4
title: Example Note
created-at: 2025-07-17T10:30:00Z
tags: [example, srs]
vice:
  srs:
    easiness: 2.5
    consecutive_correct: 1
    due: 1642723200
    total_reviews: 3
    review_history:
      - timestamp: 1642636800
        quality: 5
      - timestamp: 1642550400
        quality: 4
      - timestamp: 1642464000
        quality: 3
---

# Example Note Content

This note demonstrates the files-first approach.
```

## Consequences

### Positive

- **Complete Data Portability**: All data travels with markdown files, no vendor lock-in
- **ZK Compatibility**: Proven to work with existing ZK notebooks without conflicts
- **Disaster Recovery**: Source files are human-readable and always recoverable
- **Incremental Adoption**: Can start without database, add caching later
- **Development Simplicity**: Easier to debug and understand data flow
- **Version Control Friendly**: All changes visible in Git diffs
- **No Migration Required**: Works with existing ZK notebooks immediately

### Negative

- **Query Performance**: Complex SRS queries require full file scanning without cache
- **Frontmatter Size**: Large review histories increase file size and parsing time
- **Consistency Challenges**: Need careful coordination between file and cache updates
- **Parsing Overhead**: YAML parsing required for every file access
- **Concurrent Access**: File locking needed for safe concurrent updates

### Neutral

- **Cache Complexity**: Optional SQLite cache adds implementation complexity but is rebuildable
- **Storage Efficiency**: Frontmatter overhead vs database normalization trade-offs
- **Backup Strategy**: Files are easier to backup but cache rebuild takes time
- **Performance Ceiling**: File-based approach has natural performance limits for large collections

## Implementation Details

### Change Detection Strategy
Use timestamp + SHA256 checksum (following ZK's proven approach) for efficient cache invalidation:

```go
type FileMetadata struct {
    Path     string
    ModTime  time.Time
    Checksum string
}
```

### Cache Tables Schema
Add Vice-specific tables to existing databases:

```sql
-- Added to ZK's notebook.db or Vice's context database
CREATE TABLE vice_srs_cache (
    note_id TEXT PRIMARY KEY,
    due_timestamp INTEGER,
    easiness REAL,
    consecutive_correct INTEGER,
    total_reviews INTEGER,
    file_checksum TEXT,
    updated_at INTEGER
);

CREATE TABLE vice_file_cache (
    file_path TEXT PRIMARY KEY,
    checksum TEXT,
    parsed_at INTEGER,
    link_count INTEGER,
    tag_count INTEGER
);
```

### Error Recovery Process
1. Detect cache inconsistency (checksum mismatch)
2. Mark cache entry as stale
3. Reparse source file
4. Update cache with new data
5. For corruption: drop all Vice tables and rebuild from files

---
*ADR format based on [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)*