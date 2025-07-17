# ZK-Vice Interoperability Design

## Executive Summary

This document outlines the design for seamless interoperability between ZK notebooks and Vice flotsam system, ensuring both systems can operate on the same files without conflicts.

## Current State Analysis

### ZK Notebook Structure (~/workbench/zk)
- **Files**: Individual `.md` files with 4-char alphanumeric IDs (`jgtt.md`, `n10k.md`, etc.)
- **Database**: SQLite database at `.zk/notebook.db` with full-text search indexing
- **Configuration**: TOML config at `.zk/config.toml` with ID generation, templates, tools
- **Frontmatter**: ZK-standard YAML fields (`id`, `title`, `created-at`, `tags`)

### Vice Flotsam Structure (Planned)
- **Files**: Individual `.md` files in `$VICE_DATA/{context}/flotsam/`
- **Database**: File-based source of truth with optional caching
- **Configuration**: Vice configuration system
- **Frontmatter**: ZK-compatible + SRS extensions (`srs` field with scheduling data)

## Key Interoperability Challenges

### 1. Directory Structure Mismatch
- **ZK**: Flat directory structure using `ZK_NOTEBOOK_DIR` environment variable
- **Vice**: Context-scoped directory structure `$VICE_DATA/{context}/flotsam/`
- **Conflict**: Same files need to exist in different directory expectations

### 2. Metadata Storage Conflicts
- **ZK**: SQLite database as authoritative index with file change detection via checksums
- **Vice**: Individual files as source of truth, potential in-memory caching
- **Conflict**: ZK rebuilds index on file changes, Vice needs to track SRS data

### 3. Frontmatter Schema Extensions
- **ZK**: Minimal schema with extensible metadata map
- **Vice**: Requires SRS scheduling data (`easiness`, `due`, `consecutive_correct`, etc.)
- **Conflict**: Vice extensions must not break ZK parsing

### 4. Link Resolution Scope
- **ZK**: Notebook-wide link resolution with SQLite-backed backlink computation
- **Vice**: Context-scoped link resolution within flotsam directories
- **Conflict**: Different scoping rules for the same content

### 5. File Modification Synchronization
- **ZK**: Direct file modification triggers SQLite reindexing
- **Vice**: File modification + SRS metadata updates
- **Conflict**: Changes by one system might not be properly handled by the other

## Design Solution: Files-First Hybrid Interoperability Layer

### Core Design Principles

1. **Files as Source of Truth**: Markdown files contain all persistent data including SRS history
2. **SQLite as Performance Cache**: Database rebuilt from files for fast queries
3. **Non-Destructive**: Never break existing ZK functionality
4. **Bidirectional**: Both systems can read/write same files
5. **Upgradeable**: On-demand upgrade path with rollback capability
6. **Transparent**: ZK remains unaware of Vice extensions
7. **Rebuildable**: SQLite can always be reconstructed from markdown files

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    ZK Notebook Directory                        │
│                    (e.g., ~/workbench/zk)                      │
├─────────────────────────────────────────────────────────────────┤
│  *.md files with ZK-compatible frontmatter                     │
│  + Vice SRS extensions (SOURCE OF TRUTH)                       │
│                                                                 │
│  Data Flow: Write to .md → Rebuild SQLite cache                │
├─────────────────────────────────────────────────────────────────┤
│  .zk/notebook.db     │  .zk/notebook.db + Vice tables          │
│  (ZK SQLite index)   │  (Vice SRS cache - rebuilt from files)  │
├─────────────────────────────────────────────────────────────────┤
│  .zk/config.toml     │  .vice/config.yaml                     │
│  (ZK configuration)  │  (Vice flotsam configuration)          │
└─────────────────────────────────────────────────────────────────┘
```

### Implementation Strategy

#### Phase 1: Files-First SRS Data Storage

**Goal**: Store all SRS data in markdown frontmatter as source of truth

**Approach**:
```yaml
---
# ZK standard fields (preserved)
id: jgtt
title: git
created-at: "2025-06-23 14:06:42"
tags: [draft, to/review, tech, versioning]

# Vice extensions (ignored by ZK, source of truth)
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
      - timestamp: 1640994900
        quality: 5
  context: "default"
  flotsam_type: "idea"
---
```

**Benefits**:
- ZK ignores unknown frontmatter fields
- All persistent data travels with markdown files
- Complete SRS history preserved in text format
- SQLite can always be rebuilt from files
- Fully backward compatible

#### Phase 2: Directory Bridge System

**Goal**: Allow Vice to operate on ZK notebooks without moving files

**Approach**:
```go
// Vice configuration can specify ZK notebook as flotsam directory
type FlotsamConfig struct {
    // Standard Vice context directory
    Directory string `yaml:"directory"`
    
    // ZK notebook integration
    ZKNotebook *ZKNotebookConfig `yaml:"zk_notebook,omitempty"`
}

type ZKNotebookConfig struct {
    Path string `yaml:"path"`           // Path to ZK notebook
    Context string `yaml:"context"`     // Vice context association
    ReadOnly bool `yaml:"read_only"`    // Prevent Vice modifications
}
```

**Benefits**:
- Vice can read existing ZK notebooks
- No file movement required
- Configurable read-only mode for safety

#### Phase 3: SQLite Performance Cache

**Goal**: Maintain SQLite cache for fast queries while keeping files as source of truth

**Approach**:
- **ZK Database**: Remains authoritative for ZK features (FTS, links, etc.)
- **Vice Cache Tables**: Added to ZK database for performance (ZK ignores them)
- **Synchronization**: Rebuild cache from markdown files on changes

```sql
-- Vice cache tables (added to existing ZK database)
CREATE TABLE vice_srs_cache (
    note_id INTEGER PRIMARY KEY REFERENCES notes(id) ON DELETE CASCADE,
    context TEXT NOT NULL DEFAULT 'default',
    easiness REAL DEFAULT 2.5,
    consecutive_correct INTEGER DEFAULT 0,
    due_timestamp INTEGER,
    total_reviews INTEGER DEFAULT 0,
    last_review_timestamp INTEGER,
    card_type TEXT DEFAULT 'idea',
    file_checksum TEXT NOT NULL -- For cache invalidation
);

CREATE TABLE vice_review_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id INTEGER REFERENCES notes(id) ON DELETE CASCADE,
    timestamp INTEGER NOT NULL,
    quality INTEGER NOT NULL,
    FOREIGN KEY (note_id) REFERENCES notes(id)
);

-- Performance indexes
CREATE INDEX idx_vice_srs_due ON vice_srs_cache(due_timestamp);
CREATE INDEX idx_vice_srs_context ON vice_srs_cache(context);
CREATE INDEX idx_vice_history_note ON vice_review_history(note_id);
```

**Data Flow**:
1. **Write**: Update markdown file frontmatter
2. **Cache**: Rebuild SQLite cache from file
3. **Query**: Use SQLite for fast SRS queries

**Benefits**:
- Files remain authoritative source of truth
- Fast queries for "cards due today" etc.
- Cache can always be rebuilt from files
- ZK completely ignores Vice tables
- No conflicts between ZK and Vice operations

#### Phase 4: Link Resolution Compatibility

**Goal**: Respect both ZK notebook-wide and Vice context-scoped link resolution

**Approach**:
```go
type LinkResolver interface {
    ResolveLink(target string, context string) (*Note, error)
}

type HybridLinkResolver struct {
    zkResolver *ZKLinkResolver
    viceResolver *ViceContextLinkResolver
}

func (r *HybridLinkResolver) ResolveLink(target string, context string) (*Note, error) {
    // Try Vice context-scoped resolution first
    if note, err := r.viceResolver.ResolveLink(target, context); err == nil {
        return note, nil
    }
    
    // Fall back to ZK notebook-wide resolution
    return r.zkResolver.ResolveLink(target, "")
}
```

**Benefits**:
- Vice maintains context boundaries
- ZK link resolution still works
- Graceful fallback behavior

### Migration Strategy

#### On-Demand Upgrade Process

1. **Detection**: Check if directory is ZK notebook (`.zk/config.toml` exists)
2. **Backup**: Create backup of notebook state
3. **Analysis**: Scan existing notes for compatibility
4. **Migration**: Add Vice metadata directory and configuration
5. **Validation**: Verify ZK still works normally

#### Upgrade Command

```bash
# Upgrade existing ZK notebook for Vice flotsam use
vice flotsam upgrade --zk-notebook ~/workbench/zk --context default

# Verify ZK still works
zk list --path ~/workbench/zk

# Rollback if needed
vice flotsam rollback --zk-notebook ~/workbench/zk
```

#### Safety Measures

1. **Backup Creation**: Full notebook backup before upgrade
2. **Read-Only Mode**: Option to prevent Vice modifications
3. **Validation Tests**: Automated checks that ZK commands still work
4. **Rollback Capability**: Remove Vice metadata without affecting ZK

### Database Schema Design

#### Vice Cache Tables (Added to ZK's notebook.db)

```sql
-- SRS cache (rebuilt from markdown frontmatter)
CREATE TABLE vice_srs_cache (
    note_id INTEGER PRIMARY KEY REFERENCES notes(id) ON DELETE CASCADE,
    context TEXT NOT NULL DEFAULT 'default',
    easiness REAL DEFAULT 2.5,
    consecutive_correct INTEGER DEFAULT 0,
    due_timestamp INTEGER,
    total_reviews INTEGER DEFAULT 0,
    last_review_timestamp INTEGER,
    card_type TEXT DEFAULT 'idea',
    file_checksum TEXT NOT NULL, -- For cache invalidation
    updated_at INTEGER DEFAULT (strftime('%s', 'now'))
);

-- Review history cache (rebuilt from markdown frontmatter)
CREATE TABLE vice_review_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id INTEGER REFERENCES notes(id) ON DELETE CASCADE,
    timestamp INTEGER NOT NULL,
    quality INTEGER NOT NULL,
    created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

-- Context definitions
CREATE TABLE vice_contexts (
    name TEXT PRIMARY KEY,
    description TEXT,
    created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

-- Performance indexes
CREATE INDEX idx_vice_srs_due ON vice_srs_cache(due_timestamp);
CREATE INDEX idx_vice_srs_context ON vice_srs_cache(context);
CREATE INDEX idx_vice_history_note ON vice_review_history(note_id);
CREATE INDEX idx_vice_srs_checksum ON vice_srs_cache(file_checksum);
```

**Key Design Decisions**:
1. **Additive Only**: No modifications to existing ZK tables
2. **Source of Truth**: All persistent data in markdown frontmatter
3. **Cache Invalidation**: Use file checksums to detect changes
4. **Performance**: Indexes for common SRS queries
5. **Referential Integrity**: Foreign keys ensure data consistency

#### Configuration Schema (`.vice/config.yaml`)

```yaml
# Vice flotsam configuration for ZK notebook
version: 1
type: zk_notebook
source:
  path: "/home/david/workbench/zk"
  zk_config: ".zk/config.toml"
  
contexts:
  default:
    name: "Default Context"
    srs_enabled: true
    link_scope: "context"  # or "notebook"
    
integration:
  read_only: false
  sync_interval: 300  # seconds
  backup_enabled: true
  
srs:
  algorithm: "sm2"
  initial_easiness: 2.5
  quality_scale: [0, 1, 2, 3, 4, 5, 6]
```

### Testing Strategy

#### Compatibility Testing

1. **ZK Functionality**: Ensure all ZK commands work normally
2. **Vice Functionality**: Verify Vice can read/write notes with SRS data
3. **Round-trip**: Test ZK→Vice→ZK operations preserve data
4. **Link Resolution**: Verify both systems resolve links correctly

#### Test Cases

```bash
# Test ZK functionality still works
zk list --path ~/workbench/zk
zk new --title "Test Note" --path ~/workbench/zk
zk edit test-note --path ~/workbench/zk

# Test Vice functionality
vice flotsam list --context default
vice flotsam new --title "Vice Test" --context default
vice flotsam review --context default

# Test interoperability
zk edit vice-created-note --path ~/workbench/zk
vice flotsam edit zk-created-note --context default
```

### Implementation Plan

#### Subtask 1.1.3 Updated: ZK-Vice Interoperability

1. **Research Phase** ✓
   - Analyzed existing ZK notebook structure
   - Identified key interoperability challenges
   - Documented ZK database schema and configuration

2. **Design Phase** ✓
   - Created hybrid interoperability architecture
   - Designed safe frontmatter extension strategy
   - Planned directory bridge system
   - Specified metadata synchronization approach

3. **Implementation Phase** (Next)
   - Implement safe frontmatter extensions
   - Build directory bridge system
   - Create metadata synchronization layer
   - Develop migration tools

4. **Testing Phase**
   - Compatibility testing with real ZK notebook
   - Round-trip data integrity testing
   - Performance impact assessment

## Risk Assessment

### Low Risk
- **Frontmatter Extensions**: ZK ignores unknown fields
- **Separate Metadata**: No conflicts with ZK database
- **Read-Only Mode**: Zero modification risk

### Medium Risk
- **Directory Bridge**: Complexity in path resolution
- **Link Resolution**: Context vs notebook scope conflicts
- **Synchronization**: Race conditions possible

### High Risk
- **File Modification**: Both systems modifying same files
- **Database Corruption**: Concurrent access issues
- **ID Conflicts**: Different ID generation between systems

### Mitigation Strategies

1. **File Locking**: Implement file-level locks for modifications
2. **Atomic Operations**: Use atomic file operations for updates
3. **Conflict Detection**: Checksum-based change detection
4. **Backup Strategy**: Automatic backups before operations
5. **Rollback Capability**: Full rollback of Vice modifications

## Conclusion

The hybrid interoperability design provides a safe, non-destructive way to integrate ZK notebooks with Vice flotsam while maintaining full compatibility. The phased approach allows for gradual adoption with multiple safety measures and rollback capabilities.

Key success factors:
- Files-first architecture with text files as source of truth
- ZK-compatible frontmatter extensions for SRS data
- SQLite performance cache that ZK ignores
- Comprehensive testing and validation
- Clear migration and rollback procedures

This design ensures that users can leverage both ZK's powerful notebook features and Vice's SRS capabilities on the same content without compromising either system's functionality, while maintaining data portability and the ability to completely rebuild the system from text files.