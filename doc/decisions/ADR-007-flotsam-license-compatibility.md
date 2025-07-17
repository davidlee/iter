# ADR-007: Flotsam License Compatibility

**Status**: Accepted

**Date**: 2025-07-17

## Related Reading

**Related ADRs**: 
 - [ADR-003: ZK-go-srs Integration Strategy](/doc/decisions/ADR-003-zk-gosrs-integration-strategy.md) - Component integration approach requiring license analysis
 - [ADR-002: Flotsam Files-First Architecture](/doc/decisions/ADR-002-flotsam-files-first-architecture.md) - Data model decisions affecting licensing scope

**Related Specifications**: 
 - [Flotsam Package Documentation](/doc/specifications/flotsam.md) - License attribution documentation

**Related Tasks**: 
 - [T027/1.1] - ZK component integration requiring GPLv3 compliance
 - [T027/1.2] - go-srs component integration requiring Apache-2.0 compliance
 - [T027/1.3.1] - Attribution compliance verification across all copied components

## Context

The flotsam data layer integrates code from multiple external projects with different open source licenses:

### License Matrix

| Component | Original Project | License | Integration Method |
|-----------|------------------|---------|-------------------|
| **Vice Project** | [Vice](https://github.com/vice-org/vice) | GPLv3 | Host project |
| **ZK Components** | [zk-org/zk](https://github.com/zk-org/zk) | GPLv3 | Code copying with attribution |
| **go-srs Components** | [revelaction/go-srs](https://github.com/revelaction/go-srs) | Apache-2.0 | Code copying with attribution |

### Legal Questions

1. **GPLv3 + Apache-2.0 Compatibility**: Can Apache-2.0 licensed code be integrated into a GPLv3 project?
2. **Attribution Requirements**: What attribution and copyright notices are required for each license?
3. **Distribution Requirements**: What obligations exist for source code distribution and license notices?
4. **Derivative Work Status**: How does copying and adapting code affect licensing obligations?

### Integration Approach

The flotsam package uses **component extraction** rather than library linking:
- **ZK Components**: Copied and adapted specific functions from `internal/` packages (not importable)
- **go-srs Components**: Copied and adapted SM-2 algorithm implementation (avoiding full library dependencies)
- **Modifications**: Adapted for flotsam's files-first architecture and Vice integration patterns

## Decision

**We choose GPLv3 license for the entire flotsam package** with proper attribution to upstream projects, based on the principle that GPLv3 is compatible with Apache-2.0 when Apache-2.0 code is incorporated into a GPLv3 project.

### Legal Framework:

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         FLOTSAM LICENSE ARCHITECTURE                          │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐              │
│  │ Vice Project    │    │ ZK Components   │    │ go-srs Components│              │
│  │ (Host Project)  │    │ (External)      │    │ (External)      │              │
│  │                 │    │                 │    │                 │              │
│  │ License: GPLv3  │────│ License: GPLv3  │────│ License:        │              │
│  │ Compatible: ✅   │    │ Compatible: ✅   │    │ Apache-2.0      │              │
│  │                 │    │ Same License    │    │ Compatible: ✅   │              │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘              │
│           │                       │                       │                     │
│           ▼                       ▼                       ▼                     │
│  ┌─────────────────────────────────────────────────────────────────┐            │
│  │                    LICENSE INTEGRATION                           │            │
│  │                                                                 │            │
│  │  Flotsam Package License: GPLv3                                 │            │
│  │  ├─ Vice code: GPLv3 (same license)                             │            │
│  │  ├─ ZK code: GPLv3 (same license)                               │            │
│  │  └─ go-srs code: Apache-2.0 → GPLv3 (compatible direction)     │            │
│  │                                                                 │            │
│  │  Result: Entire work licensed under GPLv3                      │            │
│  └─────────────────────────────────────────────────────────────────┘            │
│                                                                                 │
│  Distribution: GPLv3 requirements apply to entire flotsam package              │
│  Attribution: Proper copyright notices for ZK (GPLv3) and go-srs (Apache-2.0) │
│  Source Code: Must be available under GPLv3 terms                              │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Key Legal Principles:

#### 1. License Compatibility Direction
**Apache-2.0 → GPLv3**: ✅ **Compatible**
- Apache-2.0 is compatible with GPLv3 when Apache-2.0 code is incorporated into GPLv3 projects
- The combined work must be distributed under GPLv3 terms
- Apache-2.0 attribution requirements must be preserved

**GPLv3 → Apache-2.0**: ❌ **Incompatible** 
- GPLv3 code cannot be incorporated into Apache-2.0 projects
- This direction is not relevant to our integration

#### 2. Attribution Requirements
Both licenses require proper attribution with different specific requirements:

**ZK Components (GPLv3)**:
```go
// Copyright 2024 The zk-org Authors
// Copyright 2024 David Holsgrove
//
// This file contains code copied and adapted from zk (https://github.com/zk-org/zk)
// Original source: internal/core/note_parse.go
// Licensed under GPLv3
//
// SPDX-License-Identifier: GPL-3.0-only
```

**go-srs Components (Apache-2.0)**:
```go
// Copyright (c) 2025 Vice Project
// This file contains code adapted from the go-srs spaced repetition system.
// Original code: https://github.com/revelaction/go-srs
// Original license: Apache License 2.0
// 
// Portions of this file are derived from go-srs's SM-2 algorithm implementation,
// specifically from algo/sm2/sm2.go and review/review.go.
// The original go-srs code is licensed under Apache-2.0.
```

#### 3. Derivative Work Analysis
**Code Copying Classification**:
- **Substantial Copying**: Functions and algorithms copied with modifications
- **Adaptation**: Code modified for different architecture (files-first vs database-first)
- **Integration**: Combined with Vice's existing codebase patterns
- **Result**: Creates derivative work subject to GPL terms

## Consequences

### Positive

- **Legal Compliance**: Clear legal framework with established compatibility precedent
- **Proper Attribution**: All upstream contributors properly credited and acknowledged
- **License Clarity**: Users understand their rights and obligations under GPLv3
- **Upstream Compatibility**: Can contribute improvements back to ZK (same license)
- **Freedom Preservation**: GPLv3 ensures flotsam remains free and open source
- **Standard Practice**: Common pattern in open source for incorporating Apache-2.0 into GPL projects

### Negative

- **GPLv3 Obligations**: Downstream users must comply with GPLv3 copyleft requirements
- **Distribution Complexity**: Must ensure source code availability for GPLv3 compliance
- **License Mixing Documentation**: Requires careful documentation of mixed-license components
- **Contributor Agreements**: Future contributors must understand multi-license heritage

### Neutral

- **No Runtime Impact**: License compatibility doesn't affect application functionality
- **Documentation Overhead**: Requires maintaining attribution information
- **Legal Review**: May require legal review for some commercial distributions
- **Future Integration**: Need to consider licensing when adding more external components

## Implementation Details

### Copyright Header Standards

#### 1. Vice Original Code
All new code written for flotsam uses standard Vice headers:
```go
// Copyright (c) 2025 Vice Project
// SPDX-License-Identifier: GPL-3.0-only

package flotsam
```

#### 2. ZK-Derived Code
Code copied/adapted from ZK includes full attribution:
```go
// Copyright 2024 The zk-org Authors
// Copyright 2024 David Holsgrove
//
// This file contains code copied and adapted from zk (https://github.com/zk-org/zk)
// Original source: [specific source file path]
// Licensed under GPLv3
//
// Vice Project modifications:
// - [description of specific modifications made]
// - [architectural adaptations for files-first approach]
//
// SPDX-License-Identifier: GPL-3.0-only

package flotsam
```

#### 3. go-srs-Derived Code
Code copied/adapted from go-srs includes Apache-2.0 attribution:
```go
// Copyright (c) 2025 Vice Project
// This file contains code adapted from the go-srs spaced repetition system.
// Original code: https://github.com/revelaction/go-srs
// Original license: Apache License 2.0
// Original author: revelaction
//
// Portions of this file are derived from go-srs components:
// - [specific source files]
// - [algorithms and data structures used]
//
// Vice Project modifications:
// - [description of adaptations for flotsam architecture]
// - [integration with files-first storage approach]
//
// The original go-srs code is licensed under Apache-2.0.
// This derivative work is licensed under GPL-3.0-only as part of the Vice project.

package flotsam
```

### License Documentation Strategy

#### 1. Package-Level Documentation
Every package with mixed-license code includes comprehensive attribution:

```go
// Package flotsam provides a comprehensive data layer for the flotsam note system.
//
// LICENSE ATTRIBUTION:
//
// This package incorporates code from multiple open source projects:
//
// 1. ZK Components (GPLv3):
//    - Frontmatter parsing: internal/core/note_parse.go
//    - Link extraction: internal/core/link.go  
//    - ID generation: internal/core/id.go
//    - Original project: https://github.com/zk-org/zk
//    - License: GPLv3
//
// 2. go-srs Components (Apache-2.0):
//    - SM-2 algorithm: algo/sm2/sm2.go
//    - SRS interfaces: algo/algo.go, db/db.go
//    - Review system: review/review.go
//    - Original project: https://github.com/revelaction/go-srs  
//    - License: Apache-2.0
//
// 3. Vice Integration Code (GPLv3):
//    - Data model integration
//    - Repository pattern implementation
//    - Context isolation logic
//    - License: GPLv3
//
// The entire flotsam package is distributed under GPLv3 terms.
// See individual files for specific attribution details.
package flotsam
```

#### 2. NOTICE File Creation
Create `/NOTICE` file documenting all third-party components:

```
NOTICE for Vice Project - Flotsam Package

This package contains code from the following projects:

================================================================================
ZK Note-Taking System
================================================================================
Copyright 2024 The zk-org Authors
Copyright 2024 David Holsgrove
Licensed under GPLv3

Source: https://github.com/zk-org/zk
Files incorporated:
- internal/core/note_parse.go → internal/flotsam/zk_parser.go
- internal/core/link.go → internal/flotsam/zk_links.go
- internal/core/id.go → internal/flotsam/zk_id.go
- internal/util/rand/ → internal/flotsam/zk_id.go

Modifications: Adapted for files-first architecture and Vice integration

================================================================================
go-srs Spaced Repetition System  
================================================================================
Copyright (c) revelaction
Licensed under Apache License 2.0

Source: https://github.com/revelaction/go-srs
Files incorporated:
- algo/sm2/sm2.go → internal/flotsam/srs_sm2.go
- algo/algo.go → internal/flotsam/srs_interfaces.go  
- db/db.go → internal/flotsam/srs_interfaces.go
- review/review.go → internal/flotsam/srs_review.go

Modifications: Adapted for file-based storage instead of BadgerDB

================================================================================
```

### Compliance Verification Process

#### 1. Attribution Audit Checklist
- [ ] All ZK-derived files have proper GPLv3 attribution headers
- [ ] All go-srs-derived files have proper Apache-2.0 attribution headers  
- [ ] Package documentation includes comprehensive license attribution
- [ ] NOTICE file documents all third-party components
- [ ] Modified code sections are clearly identified
- [ ] Original project URLs and licenses are documented
- [ ] SPDX license identifiers are consistent

#### 2. Distribution Requirements
- [ ] Source code availability ensured (GPLv3 requirement)
- [ ] License files included in distribution (LICENSE.md for GPLv3)
- [ ] Attribution preserved in all distributed copies
- [ ] Modified versions clearly marked as changed (GPLv3 requirement)
- [ ] Appropriate Legal Notices displayed where required

#### 3. Ongoing Compliance
- [ ] New contributors informed of multi-license heritage
- [ ] Code review process checks attribution compliance
- [ ] License compatibility verified before adding new dependencies  
- [ ] Legal review process established for commercial distributions

### Future License Considerations

#### 1. Additional Component Integration
When adding new external components, verify license compatibility:

**Compatible with GPLv3**:
- GPLv3, GPLv2+ (same/compatible)
- Apache-2.0 (incorporable into GPLv3)
- MIT, BSD licenses (incorporable into GPLv3)

**Incompatible with GPLv3**:
- Proprietary/closed source licenses
- GPL-incompatible copyleft licenses
- Apache-2.0 + GPLv2 combinations (patent clause conflicts)

#### 2. Upstream Contribution Strategy
- **ZK Contributions**: Can contribute back improvements (same GPLv3 license)
- **go-srs Contributions**: Cannot directly contribute GPLv3 modifications
- **Separate Implementation**: May need clean-room reimplementation for upstream contributions

#### 3. Commercial Distribution Considerations
- **GPLv3 Compliance**: Must provide source code to all recipients
- **Patent Protection**: Apache-2.0 patent grant applies to incorporated code
- **Legal Review**: Recommend legal review for commercial distributions
- **Attribution Requirements**: All attribution must be preserved in commercial distributions

---
*ADR format based on [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)*