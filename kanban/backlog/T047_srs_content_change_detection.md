---
title: "SRS Content Change Detection for Idea Development"
tags: ["srs", "git", "flotsam", "change-detection", "enhancement"]
related_tasks: ["child-of:T026", "depends-on:T041", "relates-to:T042"]
context_windows: ["internal/config/env.go", "internal/srs/database.go", "cmd/flotsam_*.go", "doc/specifications/flotsam.md"]
---
# SRS Content Change Detection for Idea Development

**Context (Background)**:
Implement content change detection for SRS quality assessment in `idea` type flotsam notes. When users edit ideas, measure engagement/development through content changes rather than binary recall. Use git-based change detection with mtime fallback to automatically assess SM-2 quality scores that reflect idea development progress.

**Note Type Strategy**: This task focuses on `vice:type:idea` behavior. Different flotsam types will require different interaction modes and quality assessment strategies (e.g., flashcards use traditional recall, scripts might use execution success, logs use frequency). The implementation should be flexible and tag-based to support future type-specific behaviors.

**Type**: `feature`

**Overall Status:** `Ready` - T041 Foundation Complete

## Reference (Relevant Files / URLs)

### Design Documentation
- `doc/specifications/flotsam.md` - Complete specification with git and mtime strategies
- `doc/research/Mapping SM-2 Algorithm Parameters to Creation and.md` - SM-2 adaptation research

### Foundation Code
- `internal/config/env.go` - ViceEnv for git integration
- `internal/srs/database.go` - SRS database schema
- `cmd/flotsam_edit.go` - Edit workflow integration point
- `internal/flotsam/init.go` - Auto-initialization patterns

### Related Tasks
- **Parent**: T026 - Flotsam Note System (advanced features)
- **Foundation**: T041 - Unix Interop Foundation (completed)
- **Sibling**: T042-T045 - Other flotsam enhancements

## Habit / User Story

As a user developing ideas through flotsam notes, I want the SRS system to:
- Automatically detect when I've made meaningful changes to an idea during editing
- Schedule idea reviews based on development progress rather than recall accuracy
- Use content change magnitude to assess engagement level (stalled vs flowing ideas) 
- Provide audit trail of all vice operations for debugging and analysis
- Work reliably even when git is unavailable (mtime fallback)

This supports incremental idea development where "correctness" means progress rather than memorization.

## Acceptance Criteria (ACs)

### Git-based Change Detection
- [ ] Auto-initialize git repository in VICE_CONTEXT directory when git available
- [ ] Auto-commit after all file-modifying vice commands with standardized messages
- [ ] Detect content changes in notes via git diff between pre/post-edit commits
- [ ] Map change magnitude to SM-2 quality scores (no change=2, minor=5, major=6)
- [ ] Graceful degradation when git operations fail

### Mtime-based Fallback Detection  
- [ ] Add last_reviewed and last_content_hash columns to SRS database schema
- [ ] Compare file modification time vs last reviewed timestamp
- [ ] Use SHA256 content hash to distinguish real changes from timestamp-only changes
- [ ] Fallback to mtime detection when git unavailable or repository corrupted

### Integration and User Experience
- [ ] Update `vice flotsam edit` workflow with pre/post-edit change detection
- [ ] Automatic SRS quality assessment after edit sessions
- [ ] Configuration options for git integration and quality mapping thresholds
- [ ] Error handling and logging for all change detection edge cases

## Implementation Plan & Progress

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

### 1. Context-Level Git Integration
- [ ] **1.1 Git Repository Management**
  - *Scope:* Auto-detect git availability, initialize context repository
  - *Implementation:* Add GitEnabled bool and GitRepo string fields to ViceEnv
  - *Features:* Auto-init git in `$VICE_DATA/{context}/`, create .gitignore
  - *Error Handling:* Graceful degradation when git unavailable
  - Installation check: add check(s) to `vice doctor`
- [ ] **1.2 Application Logging System**
  - *Scope:* Add structured logging with github.com/charmbracelet/log
  - *Implementation:* Replace existing log calls with structured logging
  - *Features:* Configurable log levels, structured fields, pretty output
  - *Integration:* Log git operations, change detection, SRS updates
  - *Dependency:* Add github.com/charmbracelet/log to go.mod
- [ ] **1.3 Auto-Commit System**
  - *Scope:* Commit after all file-modifying vice commands
  - *Implementation:* `AutoCommit(command string)` method on ViceEnv
  - *Message Format:* "vice {command} - {timestamp}" for audit trail
  - *Integration:* Hook into flotsam add, edit, habit complete, etc.

### 2. Git-based Change Detection (Idea-Specific)
- [ ] **2.1 Edit Workflow Integration**
  - *Scope:* Capture pre-edit state, assess post-edit changes for `vice:type:idea` notes
  - *Implementation:* Update `runFlotsamEdit()` with change detection workflow
  - *Process:* Pre-edit commit → ZK edit → change analysis → SRS update
  - *Type Detection:* Check note tags to determine if change detection applies
  - *Fallback:* Use mtime detection when git operations fail
- [ ] **2.2 Idea-Specific Change Assessment**
  - *Scope:* Analyze git diff statistics for idea development progress
  - *Quality Mapping:* No changes=2 (stalled), <5 lines=5 (developing), ≥5 lines=6 (flowing)
  - *Implementation:* `AssessIdeaEditQuality()` method with tag-based dispatch
  - *Strategy Pattern:* Design for future type-specific assessment strategies
  - *Edge Cases:* Handle file moves, renames, permission changes as engagement

### 3. Mtime-based Fallback System
- [ ] **3.1 Database Schema Extension**
  - *Scope:* Add timestamp and hash columns to SRS database
  - *Schema:* `last_reviewed INTEGER`, `last_content_hash TEXT` columns
  - *Migration:* Handle existing databases gracefully (no migration needed)
  - *Indexing:* Ensure performance for timestamp-based queries
- [ ] **3.2 Timestamp and Hash Comparison**
  - *Scope:* Compare file mtime vs database last_reviewed timestamp
  - *Implementation:* `assessQualityByMtime()` and `assessQualityByContent()` methods
  - *Hash Algorithm:* SHA256 for reliable content change detection
  - *Quality Logic:* Same mapping as git-based detection

### 4. Configuration and Error Handling
- [ ] **4.1 Configuration System**
  - *Scope:* User-configurable git integration and quality thresholds
  - *Config File:* Add `[flotsam]` section to vice config.toml
  - *Options:* auto_git, change_detection method, quality mapping values
  - *Logging Config:* Add log level and format configuration options
  - *Defaults:* Sensible defaults with easy override capability
- [ ] **4.2 Comprehensive Error Handling**
  - *Scope:* Handle all git and filesystem edge cases gracefully
  - *Philosophy:* Never block user workflow due to change detection failures
  - *Structured Logging:* Use charmbracelet/log for debugging and analysis
  - *Recovery:* Automatic fallback strategies for corrupted state

## Architecture Notes

### Design Philosophy
- **Type-Specific Behavior**: Different flotsam types require different interaction modes
- **Strategy Pattern**: Tag-based dispatch to type-specific quality assessment
- **Heuristic Approach**: Accept git's change detection as "good enough"
- **Never Block User**: Change detection failures don't prevent normal operation
- **Audit Trail**: Full history of vice operations for debugging and rollback
- **Flexible Fallback**: Multiple detection methods for reliability

### Quality Assessment Strategies by Type
**Ideas (`vice:type:idea`)** - Content Change Detection:
- **No changes (0-2)**: Stalled/blocked - triggers more frequent scheduling
- **Minor changes (3-4)**: Progressing - normal development cycle  
- **Major changes (5-6)**: Flowing - longer intervals, idea maturing well

**Future Type Strategies** (for reference):
- **Flashcards (`vice:type:flashcard`)**: Traditional recall-based quality (0-6 user rating)
- **Scripts (`vice:type:script`)**: Execution success, runtime errors, usage frequency
- **Logs (`vice:type:log`)**: Entry frequency, content volume, regularity patterns

### Strategy Pattern Implementation
```go
type QualityAssessor interface {
    AssessQuality(notePath string, preEditState interface{}) (srs.Quality, error)
}

func GetQualityAssessor(noteType string) QualityAssessor {
    switch noteType {
    case "idea": return &IdeaChangeDetector{}
    case "flashcard": return &FlashcardRecallAssessor{}  // future
    default: return &DefaultAssessor{}
    }
}
```

### Integration Points
- **ViceEnv**: Central git management and change detection coordination
- **SRS Database**: Extended schema for timestamp and hash storage
- **Edit Commands**: All note editing workflows trigger change detection
- **Config System**: User control over detection methods and thresholds

## Roadblocks

*(No roadblocks identified - T041 foundation provides all necessary components)*

## Notes / Discussion Log

### Task Creation (2025-07-19 - AI)

**Extraction from T041**: Originally planned as T041 subtasks 6.2 and 6.3, extracted to T047 to maintain T041's focused scope on Unix interop foundation.

**Research Integration**: Incorporates insights from SM-2 adaptation research on mapping algorithm parameters to creative work. Quality assessment reflects idea development progress rather than recall accuracy.

**Implementation Readiness**: T041 provides complete foundation - ViceEnv structure, SRS database, ZK integration, command infrastructure. Ready for immediate implementation.

**User Benefits**: 
- Automatic SRS quality assessment eliminates manual rating burden
- Git audit trail provides complete history of vice operations
- Flexible detection methods ensure reliability across different environments
- SM-2 adaptation optimizes scheduling for idea development workflows

## Git Commit History

*(To be added during implementation)*