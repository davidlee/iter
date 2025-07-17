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

## Design Solution: Hybrid Interoperability Layer

### Core Design Principles

1. **Non-Destructive**: Never break existing ZK functionality
2. **Bidirectional**: Both systems can read/write same files
3. **Upgradeable**: On-demand upgrade path with rollback capability
4. **Transparent**: ZK remains unaware of Vice extensions
5. **Efficient**: Minimal performance impact on either system

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    ZK Notebook Directory                        │
│                    (e.g., ~/workbench/zk)                      │
├─────────────────────────────────────────────────────────────────┤
│  *.md files with ZK-compatible frontmatter                     │
│  + optional Vice SRS extensions                                │
├─────────────────────────────────────────────────────────────────┤
│  .zk/notebook.db     │  .vice/flotsam.db                       │
│  (ZK SQLite index)   │  (Vice SRS & context metadata)         │
├─────────────────────────────────────────────────────────────────┤
│  .zk/config.toml     │  .vice/config.yaml                     │
│  (ZK configuration)  │  (Vice flotsam configuration)          │
└─────────────────────────────────────────────────────────────────┘
```

### Implementation Strategy

#### Phase 1: Safe Frontmatter Extensions

**Goal**: Extend ZK frontmatter with Vice-specific fields that ZK ignores

**Approach**:
```yaml
---
# ZK standard fields (preserved)
id: jgtt
title: git
created-at: "2025-06-23 14:06:42"
tags: [draft, to/review, tech, versioning]

# Vice extensions (ignored by ZK)
vice:
  srs:
    easiness: 2.5
    consecutive_correct: 0
    due: 1640995200
    total_reviews: 0
  context: "default"
  flotsam_type: "idea"
---
```

**Benefits**:
- ZK ignores unknown frontmatter fields
- Vice can read/write SRS data safely
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

#### Phase 3: Metadata Synchronization

**Goal**: Maintain separate metadata stores without conflicts

**Approach**:
- **ZK Database**: Remains authoritative for ZK features (FTS, links, etc.)
- **Vice Database**: Stores SRS data and context associations
- **Synchronization**: File checksum-based change detection

```sql
-- Vice flotsam database schema
CREATE TABLE flotsam_notes (
    id TEXT PRIMARY KEY,
    file_path TEXT NOT NULL,
    file_checksum TEXT NOT NULL,
    context TEXT NOT NULL,
    srs_data TEXT, -- JSON blob of SRS metadata
    last_sync DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(file_path, context)
);
```

**Benefits**:
- No conflicts between ZK and Vice metadata
- Change detection prevents stale data
- Context isolation maintained

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

#### Vice Flotsam Database (`.vice/flotsam.db`)

```sql
-- Core note metadata
CREATE TABLE notes (
    id TEXT PRIMARY KEY,
    file_path TEXT NOT NULL,
    file_checksum TEXT NOT NULL,
    context TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    modified_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(file_path, context)
);

-- SRS scheduling data
CREATE TABLE srs_data (
    note_id TEXT PRIMARY KEY REFERENCES notes(id),
    easiness REAL DEFAULT 2.5,
    consecutive_correct INTEGER DEFAULT 0,
    due INTEGER, -- Unix timestamp
    total_reviews INTEGER DEFAULT 0,
    last_review DATETIME
);

-- Context associations
CREATE TABLE contexts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Link resolution cache (context-scoped)
CREATE TABLE links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id TEXT NOT NULL REFERENCES notes(id),
    target_id TEXT REFERENCES notes(id),
    target_title TEXT NOT NULL,
    context TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

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
- ZK-compatible frontmatter extensions
- Separate metadata stores to avoid conflicts
- Comprehensive testing and validation
- Clear migration and rollback procedures

This design ensures that users can leverage both ZK's powerful notebook features and Vice's SRS capabilities on the same content without compromising either system's functionality.