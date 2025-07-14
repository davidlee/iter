# T014 - Skip Preservation Bug

## Status
- **Priority**: High
- **Status**: Ready
- **Assignee**: AI
- **Epic**: Entry UX

## Problem Statement

`iter entry` fails when skipping a habit in a previously filled card where it was not previously skipped.

**Error**: 
```
📊 Recorded: 0
Error: failed to save entries: failed to update day entry: failed to update day entry: invalid day entry: goal entry at index 0: skipped entries cannot have achievement levels
```

## Expected Behavior

Be permissive and always avoid deleting user data where possible. Allow skip and preserve as much data as reasonable without complicating the code.

## Technical Context

The validation logic appears to reject entries that have both skip status and achievement levels, but this prevents users from changing their mind about skipping after previously recording data.

## Acceptance Criteria

- [ ] User can skip a previously recorded habit without data loss
- [ ] User can unskip a previously skipped habit 
- [ ] Validation permits reasonable data preservation scenarios
- [ ] No user data is lost during skip state transitions

## Implementation Notes

Likely involves updating validation logic to handle skip/achievement level combinations more permissively.