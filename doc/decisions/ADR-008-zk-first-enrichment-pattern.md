# ADR-008: ZK-First Enrichment Pattern for Flotsam Commands

## Status

Accepted

## Context

Flotsam commands need to combine note metadata (title, path, tags) from ZK with scheduling data from the SRS database. Two approaches are possible:

1. **SRS-first**: Query SRS database for note paths, then query ZK for metadata
2. **ZK-first**: Query ZK for note discovery and metadata, then enrich with SRS data

## Decision

We will use the **ZK-first enrichment pattern** as the standard approach for flotsam commands.

**Pattern**: ZK query → metadata extraction → SRS enrichment → combined output

## Rationale

### Benefits of ZK-First Pattern

1. **Source of Truth**: ZK notebook is the authoritative source for note existence and metadata
2. **Consistency**: All flotsam commands follow the same data flow pattern
3. **Resilience**: Handles cases where SRS database has stale references to deleted notes
4. **Rich Metadata**: Access to full ZK metadata (title, tags, content) for enhanced UX
5. **Performance**: ZK queries are typically fast and can be filtered efficiently

### Implementation Pattern

```go
// 1. Query ZK for note discovery and metadata
notes, err := env.ZKList("--tag", "vice:type:*", "--format", "json")

// 2. Extract metadata (title, path, tags) from ZK results  
noteMetadata := parseZKResults(notes)

// 3. Enrich with SRS scheduling data
srsData, err := srsDB.GetSRSDataBulk(notePaths)

// 4. Combine and format output
enrichedResults := combineMetadataAndSRS(noteMetadata, srsData)
```

### Consistency Across Commands

- `vice flotsam list`: ZK discovery → SRS enrichment → formatted output
- `vice flotsam due`: ZK discovery → SRS filtering → formatted output  
- `vice flotsam edit`: ZK discovery → path resolution → editor delegation

## Consequences

### Positive
- Consistent data flow across all flotsam commands
- Robust handling of note lifecycle (creation, deletion, metadata changes)
- Rich user experience with full note metadata available

### Negative  
- Requires ZK availability for all flotsam operations
- Slightly more complex than pure SRS database queries
- Need graceful degradation when ZK unavailable

### Mitigation
- Implement graceful degradation with clear error messages
- Provide offline fallback modes where appropriate
- Cache ZK metadata for performance-critical scenarios

## References

- T041: Unix Interop Foundation implementation
- `cmd/flotsam_list.go`: Reference implementation of pattern
- ZK documentation: Query formats and filtering options