---
title: "Goal Configuration Change Data Type Resilience"
type: ["bug", "enhancement"]
tags: ["scoring", "goal-editing", "entry-collection", "data-resilience", "type-conversion"]
related_tasks: ["T015"]
context_windows: ["internal/scoring/**/*.go", "internal/ui/entry/**/*.go", "internal/models/goal.go", "internal/models/entry.go"]
---

# Goal Configuration Change Data Type Resilience

**Context (Background)**:
The system lacks resilience when users modify goal configurations, particularly when changing field types, scoring methods, or goal types. The immediate issue is automatic scoring failing for simple numeric goals, but this points to broader systemic issues with data type transitions and backward compatibility that could affect multiple areas of the application.

**Context (Significant Code Files)**:
- `internal/scoring/` - Scoring logic with goal type assumptions
- `internal/ui/entry/` - Entry collection and validation
- `internal/models/goal.go` - Goal type definitions and validation
- `internal/models/entry.go` - Entry data structures and type handling
- `internal/parser/` - YAML parsing and schema validation
- `internal/ui/goalconfig/` - Goal editing workflows
- User data: `/home/david/.config/iter/goals.yml` and `entries.yml`

## 1. Goal / User Story

As a user, I should be able to modify any aspect of my goal configurations (field types, scoring methods, goal types, criteria) and have the system gracefully handle these changes without breaking existing functionality or producing confusing error messages. The system should be resilient to configuration changes and provide clear guidance when data migration or cleanup is needed.

## 2. Problem Analysis

### Immediate Issue: Simple Automatic Scoring Failure

**Steps to Reproduce:**
1. Have a simple boolean goal with manual scoring
2. Edit the goal to change field type to numeric and scoring type to automatic
3. Add criteria for automatic scoring (e.g., greater_than condition)
4. Try to enter data for today's entry for that goal

**Expected Behavior:**
- Entry collection should work normally for the updated goal
- Automatic scoring should evaluate the numeric input against the criteria
- Goal should be scored as completed/failed based on the criteria

**Actual Behavior:**
- Error: "failed to collect entry for goal do_10_push_ups: automatic scoring failed: scoring failed: goal do_10_push_ups is not an elastic goal"
- Entry collection fails completely

### Broader Data Type Resilience Issues

**Potential Problem Areas:**
1. **Scoring Logic:** Hard-coded assumptions about goal type vs scoring type relationships
2. **Entry Validation:** May not handle field type changes gracefully
3. **Historical Data:** Old entry data with different types may conflict with new goal configuration
4. **Goal Validation:** May not prevent invalid combinations or guide users through transitions
5. **UI Workflows:** Goal editing may not warn about potential data compatibility issues

## 3. Technical Analysis

### Root Cause Analysis

**Primary Issue:** The scoring system has incorrect assumptions about goal type vs scoring type relationships.

**Data Evidence:**
```yaml
# Current goal configuration (correct)
- title: Do 10 push-ups
  id: do_10_push_ups
  goal_type: simple          # Simple goal, not elastic
  field_type:
    type: unsigned_int
  scoring_type: automatic    # Automatic scoring with criteria
  criteria:
    condition:
      greater_than: 10.0

# Historical entry data (shows previous boolean values)
- goal_id: do_10_push_ups
  value: false              # Old boolean data from when it was manual
  achievement_level: none
```

### Architectural Vulnerabilities

**1. Type System Assumptions:**
- Hard-coded logic assuming automatic scoring = elastic goals
- Missing validation for valid goal type + scoring type combinations
- Lack of type safety between goal configuration and entry data

**2. Data Migration Gaps:**
- No automatic handling of field type changes
- Historical entry data may have incompatible types
- No validation or cleanup of legacy data when goals change

**3. Configuration Change Handling:**
- Goal editing doesn't validate downstream impacts
- No warning system for breaking configuration changes
- Entry collection assumes static goal configurations

**4. Error Handling:**
- Misleading error messages that reference wrong goal types
- No graceful degradation when configuration conflicts occur
- Missing context about what caused the compatibility issue

## 4. Acceptance Criteria

### Immediate Bug Fixes
- [ ] Simple goals with automatic scoring should work correctly
- [ ] Entry collection should accept numeric input for converted goals
- [ ] Automatic scoring should evaluate criteria properly for simple goals
- [ ] Error messages should be accurate and specific to the actual issue

### Data Type Resilience Improvements
- [ ] System should handle all valid goal type + scoring type + field type combinations
- [ ] Historical entry data with different types should not break current operations
- [ ] Goal configuration changes should not require manual data cleanup
- [ ] Entry validation should adapt to current goal configuration, not historical data

### Configuration Change Management
- [ ] Goal editing should validate compatibility of new configurations
- [ ] Warning system for configuration changes that may affect existing data
- [ ] Graceful handling of incompatible historical entry data
- [ ] Clear guidance when manual intervention is needed

### Systematic Improvements
- [ ] Remove hard-coded assumptions about goal type relationships
- [ ] Implement proper type validation throughout the system
- [ ] Add comprehensive error handling with actionable messages
- [ ] Create data migration strategies for common configuration changes

### Future-Proofing
- [ ] System should be extensible for new goal types and scoring methods
- [ ] Configuration changes should be backward compatible where possible
- [ ] Entry data should be forward compatible with goal configuration evolution
- [ ] Comprehensive testing for all goal type + scoring + field type combinations

## 5. Implementation Plan & Progress

**Overall Status:** `In Progress`

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

### Phase 1: Fix Known Scoring Issue (Immediate)
**Focus:** Address the specific scoring bug for simple goals with automatic scoring

- [ ] **Sub-task 1.1:** Investigate scoring logic assumptions
  - *Design:* Analyze scoring system to identify hard-coded goal type assumptions
  - *Code/Artifacts:* `internal/scoring/` - Find logic that restricts automatic scoring to elastic goals
  - *Testing Strategy:* Create test cases for all goal type + scoring type combinations
  - *AI Notes:* Look for logic that assumes `automatic scoring == elastic goals`

- [ ] **Sub-task 1.2:** Fix automatic scoring for simple goals
  - *Design:* Remove goal type restrictions from automatic scoring logic
  - *Code/Artifacts:* Update scoring validation to allow simple goals with automatic scoring
  - *Testing Strategy:* Test simple boolean→numeric goal conversion with automatic scoring
  - *AI Notes:* Ensure scoring evaluates criteria properly regardless of goal type

- [ ] **Sub-task 1.3:** Improve error messages for scoring failures
  - *Design:* Replace misleading error messages with accurate, actionable feedback
  - *Code/Artifacts:* Update error handling in scoring and entry collection
  - *Testing Strategy:* Verify error messages accurately describe actual issues
  - *AI Notes:* Focus on user-facing error clarity and debugging information

- [ ] **Sub-task 1.4:** Test goal configuration change workflows
  - *Design:* Comprehensive testing of goal editing followed by entry collection
  - *Code/Artifacts:* Integration tests covering goal type/field type/scoring changes
  - *Testing Strategy:* Test all realistic goal conversion scenarios
  - *AI Notes:* Include testing with historical entry data of different types

### Phase 2: Systematic Data Type Resilience (Future)
**Focus:** Investigate and address broader data type change vulnerabilities across the system

- [ ] **Sub-task 2.1:** Audit system for type assumption vulnerabilities
  - *Design:* TBD - Systematic analysis of codebase for similar hard-coded assumptions
  - *Code/Artifacts:* TBD - To be planned after Phase 1 completion
  - *Testing Strategy:* TBD - Comprehensive type compatibility testing
  - *AI Notes:* Phase 2 planning will be informed by findings from Phase 1 investigation

- [ ] **Sub-task 2.2:** Implement data migration strategies
  - *Design:* TBD - Design automatic handling of configuration changes
  - *Code/Artifacts:* TBD - Data migration and validation systems
  - *Testing Strategy:* TBD - Migration testing for various change scenarios
  - *AI Notes:* Will include historical data compatibility and user guidance systems

- [ ] **Sub-task 2.3:** Configuration change management system
  - *Design:* TBD - Warning and validation system for goal editing
  - *Code/Artifacts:* TBD - Goal editing workflow enhancements
  - *Testing Strategy:* TBD - User experience testing for configuration changes
  - *AI Notes:* Focus on preventing user confusion and data loss during changes

## 6. Roadblocks

*(None currently)*

## 7. Notes / Discussion Log

- `2025-07-14 - User:` Reported bug when editing goal from simple boolean to automatic numeric scoring: "goal do_10_push_ups is not an elastic goal"
- `2025-07-14 - User:` Requested generalization to look for opportunities to improve resilience to changed data types elsewhere
- `2025-07-14 - AI:` Expanded scope from simple bug fix to comprehensive data type resilience. Issue reveals broader architectural vulnerabilities: hard-coded type assumptions, missing data migration handling, inadequate configuration change validation, and poor error messaging. This affects scoring, entry collection, goal editing, and validation systems.