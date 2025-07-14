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
- `internal/scoring/engine.go:ScoreSimpleGoal()` - NEW: Dedicated simple goal scoring method (T016 fix)
- `internal/ui/entry/flow_implementations.go:performAutomaticScoring()` - FIXED: Now uses proper scoring method
- `internal/integration/goal_configuration_changes_test.go` - NEW: Integration tests for config changes
- `internal/scoring/engine_test.go:TestEngine_ScoreSimpleGoal()` - NEW: Unit tests for simple goal scoring
- `internal/models/goal.go` - Goal type definitions and validation methods (IsSimple, IsElastic, etc.)
- `internal/models/entry.go` - Entry data structures and type handling
- `internal/parser/` - YAML parsing and schema validation  
- `internal/ui/goalconfig/` - Goal editing workflows (future Phase 2 enhancement target)
- User data: `/home/david/.config/vice/goals.yml` and `entries.yml`

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

### Phase 1: Fix Known Scoring Issue (Immediate) ✅ COMPLETE
**Focus:** Address the specific scoring bug for simple goals with automatic scoring

- [x] **Sub-task 1.1:** Investigate scoring logic assumptions
  - *Design:* Analyze scoring system to identify hard-coded goal type assumptions
  - *Code/Artifacts:* `internal/scoring/` - Find logic that restricts automatic scoring to elastic goals
  - *Testing Strategy:* Create test cases for all goal type + scoring type combinations
  - *AI Notes:* ✅ Found root cause: `SimpleGoalCollectionFlow.performAutomaticScoring` tries to fake simple goals as elastic goals by copying goal and setting MiniCriteria, then calls `ScoreElasticGoal()`. But `ScoreElasticGoal()` checks `goal.IsElastic()` on original goal type, causing "not an elastic goal" error. Design flaw: simple goals with automatic scoring need their own scoring method, not elastic goal masquerading.

- [x] **Sub-task 1.2:** Fix automatic scoring for simple goals
  - *Design:* Remove goal type restrictions from automatic scoring logic
  - *Code/Artifacts:* Update scoring validation to allow simple goals with automatic scoring
  - *Testing Strategy:* Test simple boolean→numeric goal conversion with automatic scoring
  - *AI Notes:* ✅ Added `ScoreSimpleGoal()` method to scoring engine that properly handles simple goals with automatic scoring. Updated `SimpleGoalCollectionFlow.performAutomaticScoring()` to use the new method instead of faking elastic goals. Fixed existing tests that incorrectly used elastic goals for simple goal testing. Added comprehensive tests for the new scoring method covering numeric and boolean goals plus error cases.

- [x] **Sub-task 1.3:** Improve error messages for scoring failures
  - *Design:* Replace misleading error messages with accurate, actionable feedback
  - *Code/Artifacts:* Update error handling in scoring and entry collection
  - *Testing Strategy:* Verify error messages accurately describe actual issues
  - *AI Notes:* ✅ Error messages are now accurate with the core fix. The new `ScoreSimpleGoal()` method provides clear, specific error messages: "goal X is not a simple goal", "does not require automatic scoring", "has no criteria for automatic scoring". The misleading "not an elastic goal" error is eliminated since simple goals now use their own scoring method.

- [x] **Sub-task 1.4:** Test goal configuration change workflows
  - *Design:* Comprehensive testing of goal editing followed by entry collection
  - *Code/Artifacts:* Integration tests covering goal type/field type/scoring changes
  - *Testing Strategy:* Test all realistic goal conversion scenarios
  - *AI Notes:* ✅ Created `internal/integration/goal_configuration_changes_test.go` with comprehensive tests covering: boolean→numeric automatic scoring (user's exact scenario), manual→automatic scoring conversions for different field types, and verification that both simple and elastic goals work correctly with automatic scoring. All tests pass, confirming the fix resolves the reported issue and prevents similar problems.

### Phase 2: Type Assumption Audit and Architectural Resilience Plan (Next)
**Focus:** Systematic audit and planning for type assumption vulnerabilities revealed by T016

**Objective:** Create comprehensive plan to identify and address type masquerading patterns and hard-coded assumptions across the system

**Planning Activity: T017 - Type Assumption Audit and Data Resilience Architecture Plan**

**Scope:**
1. **Codebase Audit**: Search for type masquerading patterns and hard-coded assumptions
2. **Architecture Analysis**: Map configuration change impacts across all systems  
3. **Testing Strategy**: Design comprehensive type compatibility testing
4. **User Experience Review**: Identify validation and warning opportunities in goal editing flows

**Key Investigation Areas:**
- Goal validation logic for type assumption patterns
- Entry parsing and validation systems  
- UI workflow configuration change handling
- Error message accuracy across type operations
- Data migration and backward compatibility strategies

**Audit Methodologies (from Sub-task 2.1):**
1. **Type Masquerading Pattern Detection**: Search for components creating fake instances of other types (e.g., SimpleGoalCollectionFlow faking elastic goals)
2. **Hard-coded Type Assumption Analysis**: Identify conditional logic with missing type combinations (e.g., automatic scoring assumed = elastic only)
3. **Cross-Component Type Dependency Mapping**: Trace data flow between components with implicit type assumptions
4. **Error Message Accuracy Audit**: Validate error messages match actual failure conditions

**Priority Target Areas:**
- **High**: Scoring system, entry collection flows, goal validation
- **Medium**: Parser/YAML handling, UI goal configuration  
- **Lower**: Data persistence and historical data compatibility

**Search Patterns:**
- Type field manipulation + method calls: `rg -A 5 -B 5 "(goal|entry).*\.(goal_type|field_type|scoring_type).*="`
- Missing case coverage: `rg -A 10 -B 5 "IsElastic\(\)" | rg -v "IsSimple"`
- Hard-coded assumptions: `rg "automatic.*elastic|elastic.*automatic"`
- Error message issues: `rg -A 3 -B 3 "not.*elastic|not.*simple|invalid.*goal"`

**Deliverables:**
1. Vulnerability assessment with prioritized findings
2. Architecture improvement recommendations  
3. Comprehensive testing strategy for type resilience
4. Implementation roadmap with effort estimates

**Success Criteria:** Clear plan to prevent similar type assumption bugs across all goal configuration change scenarios

- [x] **Sub-task 2.1:** Preliminary research - audit methods and target areas
  - *Design:* Research systematic approaches for identifying type assumptions and architectural vulnerabilities
  - *Code/Artifacts:* Document audit methodologies (static analysis, pattern matching, dependency mapping)
  - *Testing Strategy:* Define search patterns and validation approaches for type assumption detection
  - *AI Notes:* ✅ Defined 4 audit methodologies: 1) Type masquerading pattern detection (Component A faking as Component B), 2) Hard-coded type assumption analysis (missing case coverage), 3) Cross-component type dependency mapping (interface boundary assumptions), 4) Error message accuracy audit. Prioritized target areas: scoring system, entry collection flows, goal validation (high); parser/YAML handling, UI goal config (medium); data persistence (lower). Created specific search patterns and validation approaches for each methodology.

- [x] **Sub-task 2.2:** Discovery - apply audit methods to codebase
  - *Design:* Execute systematic search for type masquerading patterns, hard-coded assumptions, validation gaps
  - *Code/Artifacts:* Raw findings from codebase analysis using defined methods
  - *Testing Strategy:* Validate findings with targeted tests to confirm vulnerabilities
  - *AI Notes:* ✅ Systematic audit complete. Found no active type masquerading patterns (T016 fix eliminated main issue). Identified key vulnerabilities: 1) Overly restrictive validation (checklist goals → checklist field types only), 2) Hard-coded type assumptions in goal validation logic (different paths for Simple vs Elastic), 3) Missing case coverage (informational goal scoring, some field type combinations), 4) Limited extensibility for new goal types. Scoring engine properly separated post-T016. Flow factory shows clean architecture. Error messages now accurate.
---
● Discovery Summary: Raw Findings

  Based on systematic audit using the defined methodologies, I've identified several categories of type assumptions and potential vulnerabilities:

  1. Type Masquerading Patterns (MODERATE RISK)

  - No active patterns found: The T016 fix eliminated the main type masquerading issue where SimpleGoalCollectionFlow was faking simple goals as elastic goals
  - Scoring engine separation: Both ScoreSimpleGoal() and ScoreElasticGoal() have proper type validation preventing cross-type usage

  2. Hard-coded Type Assumptions (HIGH IMPACT)

  Validation Logic Restrictions:
  - Checklist goals → Checklist field type only: internal/models/goal.go:222-235 prevents checklist goals from using other field types
  - Goal type → Scoring requirements: Different validation paths for Simple vs Elastic goals assume different capabilities
  - Field type → Goal type coupling: Some field types implicitly tied to specific goal types

  Flow Factory Pattern:
  - Clean architecture: internal/ui/entry/flow_factory.go uses proper factory pattern without type assumptions
  - Good separation: Each goal type gets its own flow without cross-type dependencies

  3. Cross-Component Type Dependencies (LOW RISK)

  - Scoring engine interface: Proper separation between ScoreSimpleGoal() and ScoreElasticGoal() methods
  - Entry collection flows: Clean routing based on goal types without assumptions

  4. Error Message Accuracy (RESOLVED)

  - Accurate messages: Post-T016 fix, error messages correctly identify actual issues
  - Type-specific errors: Each scoring method provides appropriate error context

  5. Missing Case Coverage (MEDIUM RISK)

  Identified Gaps:
  - Informational goal scoring: No explicit handling for automatic scoring in informational goals
  - Field type combinations: Some field types may not work with all goal types despite logical compatibility
  - Configuration validation: Limited checking of field type + scoring type compatibility

  Key Architectural Vulnerabilities:

  1. Overly Restrictive Validation: Checklist goals forced to use checklist field types only
  2. Implicit Type Coupling: Field types with built-in assumptions about their usage context
  3. Missing Flexibility: Validation logic prevents potentially valid combinations
  4. Limited Extensibility: Adding new goal types requires multiple validation updates

---
- [ ] **Sub-task 2.3:** Analysis - prioritize and categorize findings
  - *Design:* Assess impact, effort, and risk for each discovered vulnerability
  - *Code/Artifacts:* Structured analysis with prioritization matrix and architectural impact assessment
  - *Testing Strategy:* Risk assessment based on user impact and system stability
  - *AI Notes:* Group findings by architectural pattern and impact scope

- [ ] **Sub-task 2.4:** Presentation - deliver findings and recommendations to user
  - *Design:* Create clear summary of vulnerabilities, recommendations, and implementation roadmap
  - *Code/Artifacts:* Executive summary with actionable recommendations and effort estimates
  - *Testing Strategy:* Present testing strategy for each recommended improvement
  - *AI Notes:* Focus on actionable insights with clear business justification

- [ ] **Sub-task 2.5:** Update Phase 3 plan with specific implementation tasks
  - *Design:* Convert audit findings into concrete implementation sub-tasks for Phase 3
  - *Code/Artifacts:* Updated Phase 3 section with specific, prioritized implementation tasks
  - *Testing Strategy:* Define acceptance criteria and testing approach for each Phase 3 task
  - *AI Notes:* Ensure Phase 3 tasks are specific enough to execute without further planning

### Phase 3: Systematic Data Type Resilience Implementation (Future)
**Focus:** Execute improvements identified in T017 audit and planning phase

- [ ] **Sub-task 3.1:** Address prioritized type assumption vulnerabilities  
  - *Design:* TBD - Based on T017 audit findings
  - *Code/Artifacts:* TBD - Fixes for identified type masquerading patterns
  - *Testing Strategy:* TBD - Comprehensive type compatibility testing
  - *AI Notes:* Implementation guided by T017 architecture recommendations

- [ ] **Sub-task 3.2:** Implement data migration strategies
  - *Design:* TBD - Design automatic handling of configuration changes
  - *Code/Artifacts:* TBD - Data migration and validation systems  
  - *Testing Strategy:* TBD - Migration testing for various change scenarios
  - *AI Notes:* Will include historical data compatibility and user guidance systems

- [ ] **Sub-task 3.3:** Configuration change management system
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
- `2025-07-14 - AI:` Phase 1 COMPLETE: Fixed the immediate scoring issue. Root cause was `SimpleGoalCollectionFlow.performAutomaticScoring()` trying to masquerade simple goals as elastic goals, which failed because `ScoreElasticGoal()` validates the original goal type. Solution: Added dedicated `ScoreSimpleGoal()` method for simple goals with automatic scoring. Updated flow to use proper method. Added comprehensive tests. User's scenario (boolean→numeric automatic scoring) now works correctly. No more "not an elastic goal" errors for simple goals.

### Key Learnings & Implementation Guidance:

**1. Architecture Pattern Discovered:**
- Problem: Type masquerading anti-pattern where SimpleGoalCollectionFlow tried to fake simple goals as elastic goals
- Solution: Dedicated scoring methods per goal type with proper type validation
- Principle: Each goal type should have its own appropriate handling, not try to reuse others' logic

**2. Critical Code Paths:**
- `internal/scoring/engine.go:ScoreSimpleGoal()` - New method for simple goal automatic scoring
- `internal/ui/entry/flow_implementations.go:performAutomaticScoring()` - Fixed to use proper method
- Integration point: Entry collection → Flow routing → Scoring engine method selection

**3. Testing Strategy Learned:**
- Unit tests alone insufficient - integration tests covering real user scenarios crucial
- Test configuration changes, not just individual components
- Verify error messages are user-friendly and accurate

**4. Data Type Conversion Insights:**
- Historical entry data (boolean values) doesn't interfere with new goal configuration (numeric)
- System correctly adapts to current goal configuration during entry collection
- No manual data cleanup needed for goal type transitions

**5. Future Phase 2 Guidance:**
- Look for similar type assumption patterns in: goal validation, entry parsing, UI workflows
- Consider implementing configuration change warnings in goal editing UI
- Audit for other places where components might be masquerading as different types