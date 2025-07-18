# ADR-003: ZK-go-srs Integration Strategy

**Status**: Accepted

**Date**: 2025-07-17

## Related Reading

**Related ADRs**: 
 - [ADR-002: Flotsam Files-First Architecture](/doc/decisions/ADR-002-flotsam-files-first-architecture.md) - Storage strategy that enables this integration
 - (Future) ADR-005: SRS Quality Scale Adaptation - Quality scale choice for SRS reviews
 - (Future) ADR-007: License Compatibility - Legal framework for combining GPLv3 and Apache-2.0 code

**Related Specifications**: 
 - [Flotsam Package Documentation](/doc/specifications/flotsam.md) - Complete integrated API reference
 - [ZK Interoperability Design](/doc/design-artefacts/T027_zk_interoperability_design.md) - ZK compatibility analysis

**Related Tasks**: 
 - [T027/1.1] - ZK component integration (parsing, links, ID generation)
 - [T027/1.2] - go-srs component integration (SM-2, interfaces, review system)
 - [T027/1.3.3] - Cross-component integration testing validation

## Context

The flotsam system requires integrating two mature but architecturally different external systems:

### ZK Note-Taking System (GPLv3)
- **Architecture**: File-based with markdown + YAML frontmatter
- **Strengths**: Robust parsing, wikilink extraction, ID generation, ZK ecosystem compatibility
- **Components Needed**: Frontmatter parsing, goldmark AST link extraction, 4-char ID generation
- **Design Philosophy**: Decentralized, file-centric, human-readable

### go-srs Spaced Repetition (Apache-2.0) 
- **Architecture**: Algorithm-focused with pluggable storage backends
- **Strengths**: Mature SM-2 implementation, clean interfaces, proven SRS algorithms
- **Components Needed**: SM-2 algorithm, quality scale (0-6), review scheduling
- **Design Philosophy**: Centralized algorithms, database-oriented, performance-focused

### Integration Challenges
1. **Data Model Mismatch**: ZK uses file-per-note vs go-srs uses card-deck database model
2. **Storage Paradigm**: ZK's file-first vs go-srs's database-first approaches
3. **License Compatibility**: GPLv3 (ZK) and Apache-2.0 (go-srs) combination
4. **Architecture Reconciliation**: Bridging decentralized files with centralized algorithms

## Decision

**We adopt a Component Extraction and Adaptation Strategy** that copies specific components from both systems and creates a unified flotsam-specific integration layer.

### Integration Architecture:

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         FLOTSAM INTEGRATION LAYER                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐              │
│  │   ZK Components │    │ go-srs Components│    │ Flotsam Bridge  │              │
│  │   (Copied)      │    │   (Copied)      │    │   (Custom)      │              │
│  │                 │    │                 │    │                 │              │
│  │ • zk_parser.go  │    │ • srs_sm2.go    │    │ • FlotsamNote   │              │
│  │ • zk_links.go   │    │ • srs_interfaces│    │ • SRSData       │              │
│  │ • zk_id.go      │    │ • srs_review.go │    │ • Serialization │              │
│  │                 │    │                 │    │ • Validation    │              │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘              │
│           │                       │                       │                     │
│           │                       │                       │                     │
│           ▼                       ▼                       ▼                     │
│  ┌─────────────────────────────────────────────────────────────────┐            │
│  │                    UNIFIED API SURFACE                          │            │
│  │                                                                 │            │
│  │  ExtractLinks(content) []Link                                   │            │
│  │  NewIDGenerator(opts) func() string                             │            │
│  │  ProcessReview(srs, quality) (*SRSData, error)                 │            │
│  │  parseFrontmatter(content) (frontmatter, body, error)          │            │
│  └─────────────────────────────────────────────────────────────────┘            │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Key Design Decisions:

#### 1. Component Extraction Strategy
- **Copy, Don't Import**: Extract specific components rather than importing entire libraries
- **Maintain Attribution**: Preserve original copyright headers and license information
- **Adapt Interfaces**: Modify copied code to work with flotsam's data models
- **Isolated Integration**: Create clean integration boundaries between ZK and go-srs components

#### 2. Data Model Unification
- **FlotsamNote Structure**: Combine ZK's note concept with go-srs's SRS data
- **Frontmatter Schema**: Extend ZK frontmatter with go-srs-compatible SRS fields
- **File-per-Note Mapping**: Map go-srs card/deck concepts to individual note files
- **Review History Storage**: Store complete SRS history in frontmatter (not database)

#### 3. Algorithm Adaptation
- **SM-2 Implementation**: Use go-srs SM-2 algorithm with file-based storage
- **Quality Scale**: Adopt go-srs 0-6 quality scale for consistency
- **Scheduling Logic**: Adapt go-srs scheduling to work with individual note files
- **Review Sessions**: Map go-srs review sessions to flotsam note collections

#### 4. Parsing Integration
- **Goldmark AST**: Use ZK's robust AST-based link extraction (not regex)
- **Frontmatter Parsing**: Leverage ZK's YAML parsing with SRS extensions
- **ID Generation**: Use ZK's proven 4-char alphanumeric ID generation
- **Content Processing**: Maintain ZK's markdown processing capabilities

## Consequences

### Positive

- **Best of Both Worlds**: Combines ZK's robust parsing with go-srs's proven SRS algorithms
- **Proven Components**: Uses battle-tested code from both ecosystems
- **Clean Separation**: Clear boundaries between ZK, go-srs, and flotsam-specific code
- **License Compliance**: Proper attribution and legal framework for component mixing
- **Maintainability**: Components can be updated independently from upstream sources
- **Performance**: Inherits optimizations from both source systems
- **Compatibility**: Maintains ZK ecosystem compatibility while adding SRS capabilities

### Negative

- **Code Duplication**: Maintains copies of external code rather than using libraries
- **Update Complexity**: Manual process to incorporate upstream improvements
- **License Management**: Must track and maintain compliance with both GPLv3 and Apache-2.0
- **Integration Overhead**: Custom bridge code required to connect different architectures
- **Testing Burden**: Must test integration points between copied components
- **Documentation Debt**: Need to document deviations from upstream components

### Neutral

- **Code Ownership**: Flotsam package owns integration logic but not core algorithms
- **Dependency Management**: Reduces external dependencies but increases internal complexity
- **Version Control**: All integration code tracked in Vice repository
- **Debugging**: Easier to debug copied code but harder to benefit from upstream fixes

## Implementation Details

### Component Mapping Strategy

#### ZK → Flotsam Adaptations
```go
// ZK's note parsing → Flotsam note structure
func ParseNoteContent(content string) (*NoteContent, error) // ZK original
func parseFrontmatter(content string) (map[string]interface{}, string, error) // Flotsam adaptation

// ZK's link extraction → Flotsam link processing  
func (le *LinkExtractor) ExtractLinks(content string) ([]Link, error) // ZK original
func ExtractLinks(content string) []Link // Flotsam simplified API

// ZK's ID generation → Flotsam ID generation
func NewIDGenerator(opts IDOptions) func() string // ZK original (unchanged)
func NewFlotsamIDGenerator() IDGenerator // Flotsam convenience wrapper
```

#### go-srs → Flotsam Adaptations
```go
// go-srs algorithm → Flotsam SRS processing
type Algorithm interface { ProcessReview(...) } // go-srs original
func (calc *SM2Calculator) ProcessReview(oldData *SRSData, quality Quality) (*SRSData, error) // Flotsam adaptation

// go-srs data structures → Flotsam frontmatter
type Review struct { DeckID, CardID, ... } // go-srs original
type SRSData struct { Easiness, ConsecutiveCorrect, Due, ... } // Flotsam adaptation

// go-srs interfaces → Flotsam interfaces
type Handler interface { LoadCard, SaveCard } // go-srs original
type SRSStorage interface { LoadSRSData, SaveSRSData } // Flotsam adaptation
```

### Integration Patterns

#### 1. Unified Data Flow
```go
// Complete note processing workflow
func ProcessFlotsamNote(noteContent string) (*FlotsamNote, error) {
    // 1. ZK parsing
    frontmatter, body, err := parseFrontmatter(noteContent)
    if err != nil { return nil, err }
    
    // 2. ZK link extraction  
    links := ExtractLinks(body)
    
    // 3. Flotsam note construction
    note := &FlotsamNote{
        ID:    frontmatter["id"].(string),
        Title: frontmatter["title"].(string),
        Body:  body,
        Links: extractTargets(links),
        SRS:   extractSRSData(frontmatter),
    }
    
    return note, nil
}
```

#### 2. SRS Workflow Integration
```go
// SRS review process combining go-srs algorithm with ZK note structure
func ReviewNote(note *FlotsamNote, quality Quality) (*FlotsamNote, error) {
    // 1. go-srs processing
    calc := NewSM2Calculator()
    updatedSRS, err := calc.ProcessReview(note.SRS, quality)
    if err != nil { return nil, err }
    
    // 2. Update note structure
    note.SRS = updatedSRS
    
    // 3. Serialize back to frontmatter (files-first architecture)
    return note, nil
}
```

### Attribution Strategy

#### File Headers
Each copied component includes proper attribution:

```go
// ZK Components (GPLv3)
// Copyright 2024 The zk-org Authors
// Copyright 2024 David Holsgrove
// This file contains code copied and adapted from zk (https://github.com/zk-org/zk)
// Original source: internal/core/note_parse.go
// Licensed under GPLv3
// SPDX-License-Identifier: GPL-3.0-only

// go-srs Components (Apache-2.0)  
// Copyright (c) 2025 Vice Project
// This file contains code adapted from the go-srs spaced repetition system.
// Original code: https://github.com/revelaction/go-srs
// Original license: Apache License 2.0
// Portions of this file are derived from go-srs's SM-2 algorithm implementation
```

#### Documentation References
- Complete component mapping documented in `doc/specifications/flotsam.md`
- License compatibility analysis in future ADR-007
- Attribution compliance verification completed in T027/1.3.1

### Testing Strategy

#### Integration Validation
- **Cross-Component Tests**: Validate ZK parsing → go-srs processing workflows
- **Data Flow Tests**: Verify frontmatter ↔ SRS data ↔ review structures consistency  
- **Performance Tests**: Benchmark combined operations (19µs per note average achieved)
- **Compatibility Tests**: Ensure ZK notebook compatibility maintained

#### Component Isolation
- **Unit Tests**: Test each copied component independently
- **Boundary Tests**: Validate integration points between ZK and go-srs components
- **Regression Tests**: Detect deviations from upstream component behavior

---
*ADR format based on [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)*