# ADR-001: Permissive Entry Validation for User Data Preservation

**Status**: Accepted

**Date**: 2025-07-14

**Related Tasks**: [T014/1.2] - Design preservation strategy for skip transitions

## Context

The `vice entry` command fails when users skip previously completed habits, throwing "skipped entries cannot have achievement levels" validation errors. This creates a poor user experience where changing one's mind about habit completion results in data loss or application failure.

Two validation philosophies exist:
- **Strict validation**: Enforce pure data models where skipped entries have no associated data
- **Permissive validation**: Allow "dormant" data preservation for user experience

The system handles two distinct YAML files with different validation needs:
- `entries.yml`: User-generated habit tracking data requiring preservation
- `goals.yml`: Configuration data requiring stricter validation for system integrity

## Decision

Adopt a **permissive validation approach for entries.yml** while maintaining **strict validation for goals.yml**.

For entry validation specifically:
- Remove the restriction preventing skipped entries from having achievement levels
- Allow users to change habit status without losing historical achievement data  
- Treat achievement levels on skipped entries as "dormant" - preserved but ignored during processing

## Consequences

### Positive
- Users can freely change habit status without data loss
- Simplified error handling in entry collection flows
- Reduced validation complexity for user-facing operations
- Better user experience when editing past entries

### Negative
- Slightly "impure" data model where skipped entries may contain achievement levels
- Achievement levels become contextually meaningful (active vs dormant) rather than universally meaningful

### Neutral
- Goals configuration continues to use strict validation as appropriate for system configuration
- Existing entry data remains valid without migration
- Processing logic already handles conditional achievement level usage

---
*ADR format based on [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)*