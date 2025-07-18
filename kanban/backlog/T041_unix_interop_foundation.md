---
title: "Unix Interop Foundation & T027 Migration"
tags: ["flotsam", "unix-interop", "architecture", "migration", "zk-integration"]
related_tasks: ["replaces:T027", "unblocks:T026", "enables:T042,T043,T044,T045"]
context_windows: ["doc/design-artefacts/unix-interop-vs-coupled-integration-analysis.md", "internal/repository/*", "internal/flotsam/*"]
---
# Unix Interop Foundation & T027 Migration

**Context (Background)**:
Implement Unix interop approach for flotsam functionality and migrate away from T027's coupled integration. This task establishes the foundation for delegating to zk while adding vice-specific SRS functionality.

**Type**: `refactoring` + `feature`

**Overall Status:** `Backlog`

## Reference (Relevant Files / URLs)

### Design Documentation
- `doc/design-artefacts/unix-interop-vs-coupled-integration-analysis.md` - Comprehensive analysis and decision rationale
- `doc/decisions/ADR-002-flotsam-files-first-architecture.md` - Original flotsam architecture decision

### T027 Code to Migrate/Remove
- `internal/repository/interface.go` - Repository abstraction layer
- `internal/repository/file_repository.go` - File-based repository implementation
- `internal/models/flotsam.go` - Flotsam data models and validation
- `internal/flotsam/` - Core flotsam functionality (may partially preserve)

### Integration Points
- `cmd/` - CLI commands that will shell out to zk
- `internal/config/` - Configuration management
- `internal/zk/` - ZK interop
- `internal/srs/` - SM-2 algo & sqlite3 storage for SRS
- `zk/` - Reference zk installation for patterns

## Habit / User Story

As a developer implementing flotsam (Markdown / Zettelkasten + SRS) functionality, I want to:
- Replace T027's coupled approach with Unix interop patterns
- Establish foundation for zk integration and SRS database
- Create a clean migration path that preserves existing functionality
- Enable future flotsam features through tool orchestration

## Acceptance Criteria (ACs)

### T027 Migration & Cleanup
- [ ] Remove T027 repository abstraction layer (note: do not naively the file repository; simplify the abstractions added during T027)
- [ ] Migrate flotsam data models to simpler structures (if needed)
- [ ] Remove backlink computation (zk will handle this)
- [ ] Preserve any essential flotsam functionality during migration
- [ ] Update existing tests to work with new approach

### Unix Interop Foundation
- [ ] Implement basic zk shell-out functionality
- [ ] Create minimal SRS database structure
- [ ] Establish mtime-based cache invalidation patterns
- [ ] Implement `vice flotsam` command stub with basic operations

### Basic CLI Integration
- [ ] `vice flotsam list` - delegates to zk with `vice:srs` tag filter
- [ ] `vice flotsam due` - queries SRS database for due notes
- [ ] `vice flotsam edit <note>` - delegates to zk editor
- [ ] `vice doctor` - checks zk availability and reports status

### Testing & Validation
- [ ] All existing flotsam tests pass or are appropriately updated
- [ ] New integration tests for zk shell-out functionality
- [ ] Performance validation: Unix interop vs T027 startup time
- [ ] Error handling for missing zk dependency

## Implementation Plan & Progress

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

### 1. Analysis & Planning
- [ ] **1.1 T027 code audit**: Identify all components to migrate or remove
  - *Scope:* Full dependency analysis of T027 implementation
  - *Deliverable:* Migration plan with component-by-component breakdown
  - *Planning:* Detailed analysis of what to preserve vs remove recorded in Implementation Plan
  - *Context:* T027 added significant complexity (~2000+ lines) with repository patterns, in-memory backlink computation, and tightly coupled flotsam models
  - *Key Files:* `internal/repository/interface.go` (DataRepository interface), `internal/repository/file_repository.go` (CRUD operations), `internal/models/flotsam.go` (rich structs with validation)
  - *Dependencies:* Repository is used by potential CLI commands, flotsam operations, and SRS scheduling
  - *Mark For Removal (Anchor Comments):* Complex repository abstraction, in-memory note loading, automatic backlink computation
  - *Preserve:* Essential flotsam data structures, basic file operations, SRS scheduling data
  - *T027 External Code Integration:* T027 imported ZK parsing components and go-srs SRS algorithms - these may be partially preserved for SRS database functionality
  - *Repository Pattern Context:* T027 extended T028's DataRepository interface for context-aware flotsam operations - this abstraction layer is primary removal target
  - *Backlink Computation:* T027 implemented context-scoped backlink computation (~400 lines) that loads all notes into memory - delegate this entirely to zk
  - *SRS Integration:* T027 integrated SM-2 algorithm from go-srs - preserve core SRS scheduling logic but simplify data structures
  - *Test Coverage:* T027 has ~20 test files covering repository operations, flotsam models, and SRS scheduling - many will need updates for Unix interop patterns
- [ ] **1.2 Unix interop architecture design**: Define shell-out patterns and abstractions
  - *Scope:* Tool integration interface, error handling, configuration
  - *Deliverable:* Architecture document with Go interfaces and examples
  - *Planning:* Design tool abstraction layer for future extensibility. See `zk/docs` for zk repo's documentation 
  - *Context:* Need clean abstraction for shelling out to zk, with future extensibility for remind/taskwarrior
  - *Pattern:* Simple Tool interface with Execute method, error handling, and configuration
  - *Key Decisions:* How to handle tool discovery, command construction, output parsing, error propagation
  - *Configuration:* User-configurable flags, environment variables, tool paths via vice config
  - *Error Handling:* Distinguish between missing tools, command failures, and parsing errors
  - *ZK Command Patterns:* From analysis, key zk commands for delegation: `list --tag vice:srs --format json`, `edit <paths>`, `list --linked-by <note>`
  - *Output Parsing:* zk supports structured output (JSON, template formats) - design parsing for these formats
  - *Tool Discovery:* Standard `exec.LookPath("zk")` approach with helpful error messages referencing installation methods
  - *Environment Variables:* zk respects `ZK_EDITOR`, `VISUAL`, `EDITOR` - vice config can set these for consistent behavior
  - *Future Extensibility:* Design Tool interface to support remind (calendar/recurring tasks) and taskwarrior (GTD task management)
- [ ] **1.3 Redesign data models & update specifications**: Simplify flotsam models for Unix interop
  - *Context:* T027 created complex flotsam models with validation, serialization, and rich metadata
  - *Simplify:* Remove heavy structs, reduce to minimal data needed for SRS and file operations
  - *SRS Focus:* Primary need is SRS database schema and basic note metadata (title, path, tags)
  - *ZK Delegation:* Content, links, backlinks, search handled by zk - no need for complex models
  - *File Format:* Individual .md files with YAML frontmatter (compatible with zk)
  - *Database Schema:* Simple SRS table with note_path, due_date, quality, interval
  - *Update Specs:* Revise `doc/specifications/` to reflect Unix interop approach
  - *T027 Model Complexity:* Current flotsam models include rich metadata, validation, context-scoped operations, and complex serialization - most of this becomes unnecessary
  - *SRS Database Schema:* Based on go-srs integration, need fields for SM-2 algorithm: `note_path, last_reviewed, next_due, quality, interval_days, ease_factor, reviews_count`
  - *Tag-based Behaviors:* With `vice:srs` tag approach, no need for complex type systems - tags in frontmatter handle note behaviors
  - *Frontmatter Simplification:* Only need essential fields: `title, tags, created, srs` (for local SRS cache) - let zk handle the rest
  - *Context Isolation:* Unix interop maintains T028's context isolation through separate databases per context and zk notebook boundaries
  - *Specification Updates:* Need to update `doc/specifications/flotsam.md` to reflect Unix interop patterns and reduced model complexity

### 2. T027 Cleanup (Topic Branch)
- [ ] **2.1 Remove repository layer**: Eliminate T027 repository abstraction
  - *Scope:* `internal/repository/` package removal
  - *Impact:* Update all dependent code to use new patterns
  - *Planning:* Identify all consumers of repository interface
- [ ] **2.2 Simplify flotsam models**: Reduce complexity of flotsam data structures
  - *Scope:* `internal/models/flotsam.go` simplification
  - *Preserve:* Essential validation and serialization logic
  - *Planning:* Determine minimum viable flotsam model
- [ ] **2.3 Remove coupled backlink logic**: Eliminate in-memory backlink computation
  - *Scope:* Context-scoped backlink computation removal
  - *Replace:* Delegate to zk's link analysis capabilities
  - *Planning:* Map T027 backlink features to zk equivalents

### 3. SRS Database Foundation
- [ ] **3.1 SRS database schema**: Design minimal SRS database structure
  - *Schema:* `(note_path, last_reviewed, next_due, quality, interval_days)`
  - *Location:* `.vice/flotsam.db` separate from zk database
  - *Planning:* Consider future schema evolution and migration
- [ ] **3.2 mtime cache invalidation**: Implement cache consistency checking
  - *Approach:* Directory mtime checking on CLI invocations
  - *Scope:* Fast validation without external dependencies
  - *Planning:* Design for both CLI and future persistent processes
- [ ] **3.3 Basic SRS operations**: Implement core SRS database operations
  - *Operations:* Create, update, query due notes, review completion
  - *Interface:* Simple Go functions for CLI integration
  - *Planning:* Design for testability and future UI integration

### 4. ZK Integration Foundation
- [ ] **4.1 ZK shell-out abstraction**: Create reusable zk command execution
  - *Interface:* Tool abstraction for zk commands
  - *Features:* Error handling, output parsing, configuration
  - *Planning:* Design for extensibility to other tools (remind, taskwarrior)
- [ ] **4.2 ZK dependency detection**: Implement zk availability checking
  - *Scope:* `vice doctor` command for dependency validation
  - *Errors:* Helpful messages for missing zk installation
  - *Planning:* Consider graceful degradation strategies
- [ ] **4.3 Tag-based note detection**: Implement `vice:srs` tag integration
  - *Scope:* Filter notes by vice-specific tags
  - *Integration:* Combine zk tag queries with SRS database
  - *Planning:* Design tag naming conventions and hierarchy

### 5. Basic CLI Implementation
- [ ] **5.1 flotsam list command**: Implement `vice flotsam list` with zk delegation
  - *Delegation:* `zk list --tag vice:srs --format json`
  - *Enhancement:* Combine with SRS database for due date info
  - *Planning:* Design output formatting and filtering options
- [ ] **5.2 flotsam due command**: Implement `vice flotsam due` with SRS queries
  - *Query:* Direct SRS database query for due notes
  - *Output:* File paths or rich format with metadata
  - *Planning:* Consider date range filtering and priority sorting
- [ ] **5.3 flotsam edit command**: Implement `vice flotsam edit` with zk delegation
  - *Delegation:* `zk edit <note>` with proper path resolution
  - *Integration:* Work with both individual notes and filtered lists
  - *Planning:* Design for interactive selection and batch editing

### 6. Testing & Validation
- [ ] **6.1 Migration testing**: Ensure all existing functionality preserved
  - *Scope:* Run existing flotsam tests against new implementation
  - *Updates:* Modify tests to work with Unix interop patterns
  - *Planning:* Design test strategy for external tool dependencies
- [ ] **6.2 Integration testing**: Test zk shell-out functionality
  - *Scope:* Test all zk command delegations and error handling
  - *Mocking:* Consider test doubles for zk commands
  - *Planning:* Design for CI/CD environments without zk
- [ ] **6.3 Performance validation**: Compare startup time vs T027
  - *Metrics:* Cold start time, memory usage, operation latency
  - *Baseline:* Current T027 performance characteristics
  - *Planning:* Establish performance regression testing

## Relationship to T026

**T026 Status Re-evaluation**: Much of T026's scope is now handled by zk integration:

**T026 Features Now Handled by ZK**:
- External editor integration → `zk edit` delegation
- Fuzzy finder → `zk list --interactive`
- Note search and filtering → `zk list` with rich query options
- Link analysis → `zk list --linked-by`, `--link-to`

**T026 Features Still Relevant**:
- SRS scheduling and review workflows
- Flotsam-specific UI/CLI design
- Integration with vice's habit tracking
- Custom flotsam note templates and creation

**Recommendation**: 
- **T041 (this task)**: Establishes Unix interop foundation
- **T026 revision**: Focus on SRS workflows and flotsam-specific UX
- **Future tasks**: Detailed CLI/UI design for flotsam features

## Technical Design Notes

### Architecture Principles
- **Tool Delegation**: Leverage zk for complex operations
- **Simple Integration**: Minimal abstraction over external tools
- **Graceful Degradation**: Useful functionality when zk unavailable
- **Future Extensibility**: Design for remind/taskwarrior integration

### Migration Strategy
- **Topic Branch**: `feature/unix-interop-foundation`
- **Incremental**: Preserve functionality while migrating
- **Validation**: Comprehensive testing at each step
- **Documentation**: Update docs to reflect new approach

### Error Handling
- **Missing Dependencies**: Clear error messages with installation guidance
- **Command Failures**: Proper error propagation and user feedback
- **Data Consistency**: Validation of SRS database integrity

## Roadblocks

*(No roadblocks identified yet)*

## Future Improvements & Refactoring Opportunities

### **Post-Foundation Tasks**
1. **T042: ZK Dependency Management** - Enhanced installation and configuration
2. **T043: Advanced SRS Features** - Sophisticated scheduling algorithms
3. **T044: Tag-based Workflows** - Rich tag hierarchy and automation
4. **T045: ZK Configuration Management** - Notebook initialization and templates

### **Strategic Extensions**
1. **Tool Orchestration** - Framework for remind/taskwarrior integration
2. **Workflow Engine** - Cross-tool workflow automation
3. **MCP Integration** - AI-powered productivity assistance
4. **TUI Enhancement** - Rich terminal interface for tool coordination

## Notes / Discussion Log

### **Task Creation (2025-07-18 - AI)**

**Design Analysis Reference**: 
- Based on comprehensive analysis in `unix-interop-vs-coupled-integration-analysis.md`
- Decision to proceed with Unix interop approach over T027 coupled integration
- Strategic repositioning of vice as Unix tool orchestrator

**T027 Migration Scope**:
- **Repository Layer**: ~800 lines of abstraction to remove
- **Flotsam Models**: Simplify while preserving essential functionality
- **Backlink Logic**: ~400 lines of complex computation to delegate to zk
- **Tests**: ~20 test files to update or rewrite

**Unix Interop Foundation**:
- **ZK Integration**: Shell-out patterns with error handling
- **SRS Database**: Minimal SQLite schema with mtime validation
- **CLI Commands**: Basic flotsam operations with tool delegation
- **Architecture**: Extensible foundation for future tool integrations

**Risk Mitigation**:
- **Topic Branch**: Safe experimentation without breaking main
- **Incremental Migration**: Preserve functionality throughout process
- **Comprehensive Testing**: Validate each migration step
- **Performance Monitoring**: Ensure Unix interop meets performance expectations

This task establishes the foundation for vice's evolution from monolithic habit tracker to Unix productivity tool orchestrator.