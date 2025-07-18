
# DESIGN ANALYSIS: Non-ZK Filename Support (Flotsam)

**Date**: 2025-07-18  
**Scope**: Extending flotsam to support non-ZK filename patterns in bounded contexts  
**Approach**: Additive functionality leveraging ADR-006 context isolation framework

## Related Documentation

**Source Task**: [T027 Flotsam Data Layer](/kanban/in-progress/T027_flotsam_data_layer.md) - Subtask 2.3.2  
**Related ADRs**:
- [ADR-002: Flotsam Files-First Architecture](/doc/decisions/ADR-002-flotsam-files-first-architecture.md) - Source of truth strategy
- [ADR-006: Flotsam Context Isolation](/doc/decisions/ADR-006-flotsam-context-isolation.md) - Context boundary framework
- [ADR-004: Flotsam SQLite Cache Strategy](/doc/decisions/ADR-004-flotsam-sqlite-cache-strategy.md) - Performance caching approach

**Related Specifications**:
- [Flotsam Package Documentation](/doc/specifications/flotsam.md) - Complete API reference
- [File Paths & Runtime Environment](/doc/specifications/file_paths_runtime_env.md) - T028 context management

**Implementation Dependencies**:
- T027 Flotsam Data Layer (current context isolation implementation)
- T028 Repository Pattern (context-aware file operations)

### Executive Summary

**Recommendation**: Implement context-aware filename extension using ADR-006's existing context isolation framework. This approach preserves ZK compatibility while adding support for kanban/todos contexts with alternative filename patterns.

**Implementation Complexity**: Medium - leverages existing architecture with minimal core changes  
**Risk Level**: Low - additive changes, no breaking modifications to ZK workflows  
**Performance Impact**: Minimal - context-scoped caching prevents cross-contamination

### Context-Aware Extension Strategy

#### 1. New Context Types
Extend ADR-006's ContextType enumeration:

```go
type ContextType int
const (
    ViceContext ContextType = iota  // $VICE_DATA/{context}/ (existing)
    ZKNotebook                      // .zk/ directories (existing)  
    HybridContext                   // Vice + ZK overlap (existing)
    KanbanContext                   // --kanban-dir directories (NEW)
    TodosContext                    // $VICE_DATA/todos/ (NEW)
    PeriodicDailyContext            // $VICE_DATA/periodic/daily/ (NEW)
    PeriodicWeeklyContext           // $VICE_DATA/periodic/weekly/ (NEW)
    PeriodicMonthlyContext          // $VICE_DATA/periodic/monthly/ (NEW) 
    UnknownContext                  // Fallback (existing)
)
```

#### 2. Context-Specific Filename Resolvers
**Strategy**: Interface-based resolution per context type

```go
type ContextFilenameResolver interface {
    ExtractID(filename string) (string, error)
    FormatFilename(id string, metadata map[string]string) string
    ValidateFilename(filename string) error
    BuildFileIndex(files []string) (map[string]string, error) // ID -> filename
}

// Implementations:
// - ZKFilenameResolver: {ID}.md (existing behavior)
// - KanbanFilenameResolver: T\d+_*.md -> T\d+ ID extraction  
// - FreeformFilenameResolver: filename stem as ID
// - HybridFilenameResolver: {ID}-{description}.md patterns
// - PeriodicDailyResolver: YYYY-MM-DD.md -> date as ID
// - PeriodicWeeklyResolver: YYYY-WMM.md -> week as ID  
// - PeriodicMonthlyResolver: YYYY-MM.md -> month as ID
```

#### 3. Filename Pattern Analysis

##### A. Kanban Context (`--kanban-dir kanban/`)
**Pattern**: `T027_flotsam_data_layer.md`  
**ID Extraction**: `T027` (prefix before underscore)  
**Link Resolution**: `[[T027]]` -> pattern match `T027_*.md`  
**Use Case**: Project management, task tracking

**Benefits**:
- Preserves semantic task IDs
- Human-readable filenames  
- Pattern-based link resolution
- Zero impact on ZK contexts

**Implementation Requirements**:
- Task ID regex: `^T\d+`
- Pattern matching for file discovery
- Context-scoped link resolution

##### B. Todos Context (`$VICE_DATA/todos/`)
**Pattern**: `meeting-notes-2025-07-18.md`, `project-planning.md`  
**ID Extraction**: Full filename stem as ID  
**Link Resolution**: `[[meeting-notes-2025-07-18]]` -> exact match  
**Use Case**: Personal note management, ad-hoc documentation

**Benefits**:
- Maximum naming flexibility
- Intuitive link syntax
- No ID generation required
- Clear separation from ZK workflows

**Implementation Requirements**:
- Filename stem extraction
- Exact match link resolution  
- Collision handling for duplicate names

##### C. Hybrid Context (Optional)
**Pattern**: `6ub6-arbitrary-filename.md`  
**ID Extraction**: `6ub6` (prefix before dash)  
**Link Resolution**: `[[6ub6]]` works with existing ZK tools  
**Use Case**: ZK compatibility with human-readable names

**Benefits**:
- Maintains ZK tool compatibility
- Adds descriptive context
- Gradual migration path

**Challenges**:
- Separator parsing complexity
- Potential ID collisions
- Filename validation rules

##### D. Periodic Context (`$VICE_DATA/periodic/{type}/`)
**Patterns**: 
- Daily: `2025-11-07.md` (ISO 8601 date format)
- Weekly: `2025-W22.md` (ISO 8601 week format)
**Directory Structure**: `$VICE_DATA/periodic/daily/`, `$VICE_DATA/periodic/weekly/`
**ID Extraction**: Date string as ID (`2025-11-07`, `2025-W22`)
**Use Case**: Time-based journaling, periodic reviews, habit tracking logs

**Benefits**:
- Natural chronological organization
- Standardized date formats (ISO 8601)
- Context-scoped per period type
- Calendar-based navigation support
- SRS integration for periodic reviews

**Implementation Requirements**:
- Date format validation (strict ISO 8601)
- Temporal link resolution (`[[yesterday]]`, `[[+1w]]`)
- Cross-period linking (`[[2025-W22]]` from daily notes)
- Auto-creation of future periods
- Calendar-aware file discovery

**Design Considerations**:
- **Date Validation**: Strict ISO 8601 format enforcement vs flexibility
- **Relative References**: Support for `[[yesterday]]`, `[[next-week]]`, `[[+3d]]`
- **Cross-Period Links**: Daily notes linking to weekly summaries
- **SRS Integration**: Periodic review scheduling based on note type
- **Auto-Creation**: Generate next day/week files automatically
- **Time Zone Handling**: UTC vs local time for date calculations

**Link Resolution Strategy**:
```go
func (pr *PeriodicResolver) ResolveLink(link string) (*FlotsamNote, error) {
    // 1. Absolute date (2025-11-07, 2025-W22)
    if date := parseAbsoluteDate(link); date != nil {
        return pr.findByDate(date)
    }
    
    // 2. Relative references (yesterday, +1w, -3d)
    if relativeDate := parseRelativeDate(link, time.Now()); relativeDate != nil {
        return pr.findByDate(relativeDate)
    }
    
    // 3. Cross-period links (from daily to weekly)
    if crossDate := parseCrossPeriodLink(link, pr.contextType); crossDate != nil {
        return pr.findInRelatedPeriod(crossDate)
    }
    
    return nil, ErrDateNotFound
}
```

### Implementation Impact Assessment

#### 1. Architecture Changes
**Core Changes Required**:
- Extend ContextScope with FilenameResolver interface
- Add context detection for kanban/todos directories  
- Implement context-specific resolvers

**Files Modified**:
- `internal/flotsam/context.go` (context detection)
- `internal/flotsam/filename_resolvers.go` (new file)
- `internal/repository/file_repository.go` (use resolver interface)

**ADR Compliance**: Fully compatible with ADR-002 (files-first) and ADR-006 (context isolation)

#### 2. Link Resolution Strategy
**Multi-Strategy Resolution** per context:

```go
func (resolver *KanbanResolver) ResolveLink(link string) (*FlotsamNote, error) {
    // 1. Task ID pattern (highest priority)
    if taskID := extractTaskID(link); taskID != "" {
        return resolver.findByTaskPattern(taskID)
    }
    
    // 2. Exact filename match  
    if note := resolver.findByExactFilename(link); note != nil {
        return note, nil
    }
    
    // 3. Title search (fallback)
    return resolver.findByTitleMatch(link)
}
```

**Cross-Context Resolution**: **OPEN QUESTION**
- **Option A**: Strict isolation (recommended for initial implementation)
- **Option B**: Explicit syntax (`[[kanban:T027]]`, `[[zk:6ub6]]`)  
- **Option C**: Hierarchical search (current context -> parent contexts)

**Recommendation**: Start with Option A, evaluate Option B based on user feedback

#### 3. Performance Analysis
**Context Detection Overhead**: O(1) - cached after first detection  
**File Discovery**: O(n) per context - existing pattern  
**Link Resolution**: 
- ZK context: O(1) direct lookup (unchanged)
- Kanban context: O(n) pattern matching per unresolved link
- Todos context: O(1) exact match after indexing

**Optimization Strategy**:
- Build filename indices at context load time
- Cache pattern match results
- Lazy index building for large directories

**Memory Impact**: Minimal - one resolver instance per active context

#### 4. Cache Isolation Strategy
**Context-Specific Tables**: Extend ADR-004's cache strategy

```sql
-- Per-context cache tables  
CREATE TABLE vice_srs_cache_kanban_{context_hash} (...);
CREATE TABLE vice_file_cache_todos_{context_hash} (...);
CREATE TABLE vice_srs_cache_daily_{context_hash} (...);
CREATE TABLE vice_file_cache_weekly_{context_hash} (...);

-- Context registry for cleanup
CREATE TABLE vice_context_registry (
    context_id TEXT PRIMARY KEY,
    context_type TEXT NOT NULL,  -- 'zk', 'kanban', 'todos', 'periodic_daily', 'periodic_weekly'
    root_path TEXT NOT NULL,
    resolver_type TEXT NOT NULL,  -- resolver implementation
    period_type TEXT,            -- 'daily', 'weekly', 'monthly' for periodic contexts
    created_at INTEGER NOT NULL
);

-- Periodic-specific tables for temporal queries
CREATE TABLE vice_periodic_index (
    context_id TEXT NOT NULL,
    period_type TEXT NOT NULL,   -- 'daily', 'weekly', 'monthly'
    date_value TEXT NOT NULL,    -- '2025-11-07', '2025-W22', '2025-11'
    file_path TEXT NOT NULL,
    note_id TEXT NOT NULL,
    
    PRIMARY KEY (context_id, period_type, date_value),
    INDEX idx_periodic_date (period_type, date_value),
    INDEX idx_periodic_context (context_id, period_type)
);
```

**Benefits**:
- Complete isolation between contexts
- No cross-contamination of filename schemes
- Easy cleanup and migration

### Risk Assessment

#### Low Risk ✅
- **ZK Compatibility**: Zero impact on existing ZK workflows
- **Data Safety**: Files-first architecture protects against data loss
- **Context Isolation**: ADR-006 framework prevents cross-contamination
- **Incremental Adoption**: Can be deployed without affecting existing users

#### Medium Risk ⚠️  
- **Implementation Complexity**: Multiple resolver implementations  
- **Pattern Matching Performance**: O(n) link resolution in kanban contexts
- **Cache Management**: Additional cache tables per context

#### High Risk ❌
None identified - additive approach minimizes risk

### Recommended Implementation Sequence

#### Phase 1: Foundation (Low Risk)
1. **Context Detection Extension**: Add kanban/todos context types
2. **Resolver Interface**: Define ContextFilenameResolver interface  
3. **ZK Resolver**: Wrap existing logic in ZKFilenameResolver
4. **Testing Framework**: Unit tests for resolver interface

#### Phase 2: Kanban Support (Medium Value)
1. **KanbanFilenameResolver**: Task ID extraction and pattern matching
2. **Kanban Context Detection**: `--kanban-dir` flag handling
3. **Link Resolution**: Pattern-based resolution for task IDs
4. **Integration Testing**: End-to-end kanban workflow tests

#### Phase 3: Todos Support (High Value)
1. **FreeformFilenameResolver**: Filename stem as ID approach
2. **Todos Context Detection**: `$VICE_DATA/todos/` directory handling  
3. **Exact Match Resolution**: Filename-based link resolution
4. **User Testing**: Real-world todos workflow validation

#### Phase 4: Periodic Support (High Value)
1. **PeriodicFilenameResolver**: Date parsing and validation (ISO 8601)
2. **Periodic Context Detection**: `$VICE_DATA/periodic/{type}/` directory handling
3. **Temporal Link Resolution**: Absolute and relative date references
4. **Cross-Period Linking**: Daily ↔ weekly ↔ monthly integration
5. **Auto-Creation**: Generate future period files on demand

#### Phase 5: Cross-Context (Optional)
1. **Cross-Context Links**: Explicit syntax (`[[context:id]]`)  
2. **Bridge Configuration**: User-configurable context bridging
3. **Performance Optimization**: Advanced caching strategies

### Open Questions for Implementation

#### General Questions
1. **Cross-Context Linking**: Should `[[T027]]` work from ZK notes to kanban tasks?
2. **Filename Validation**: How strict should validation be for todos context?
3. **ID Collision Handling**: What happens if kanban task ID matches ZK note ID?
4. **Migration Strategy**: How to convert existing kanban files to flotsam format?
5. **Cache Sharing**: Should related contexts share cache databases?

#### Periodic-Specific Questions  
6. **Date Format Flexibility**: Strict ISO 8601 (`2025-11-07`) vs flexible (`2025-7-18`, `Nov 7 2025`)?
7. **Relative Link Syntax**: Support `[[yesterday]]`, `[[+3d]]`, `[[next-week]]` or explicit dates only?
8. **Cross-Period Navigation**: Should daily notes auto-link to containing week/month?
9. **Auto-Creation Policy**: Create future files on-demand or pre-generate (e.g., tomorrow's daily)?
10. **Time Zone Handling**: UTC timestamps vs local time for date calculations?
11. **SRS Integration**: How should periodic notes participate in spaced repetition?
12. **Rollover Behavior**: Auto-archive old periods or maintain indefinitely?

### Conclusion

Context-aware filename extension is architecturally sound and low-risk. The ADR-006 framework provides the foundation needed to support alternative filename patterns without compromising ZK compatibility. 

**Key Success Factors**:
- Leverage existing context isolation architecture
- Maintain strict separation between context types  
- Implement additive changes only
- Prioritize ZK compatibility preservation

**Next Steps**: Proceed with Phase 1 implementation to establish the foundation for context-aware filename resolution.
