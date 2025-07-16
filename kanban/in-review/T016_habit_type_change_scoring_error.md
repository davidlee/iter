---
title: "Habit Configuration Change Data Type Resilience"
type: ["bug", "enhancement"]
tags: ["scoring", "habit-editing", "entry-collection", "data-resilience", "type-conversion"]
related_tasks: ["T015"]
context_windows: ["internal/scoring/**/*.go", "internal/ui/entry/**/*.go", "internal/models/habit.go", "internal/models/entry.go"]
---

# Habit Configuration Change Data Type Resilience

**Context (Background)**:
The system lacks resilience when users modify habit configurations, particularly when changing field types, scoring methods, or habit types. The immediate issue is automatic scoring failing for simple numeric habits, but this points to broader systemic issues with data type transitions and backward compatibility that could affect multiple areas of the application.

**Context (Significant Code Files)**:
- `internal/scoring/engine.go:ScoreSimpleHabit()` - NEW: Dedicated simple habit scoring method (T016 fix)
- `internal/ui/entry/flow_implementations.go:performAutomaticScoring()` - FIXED: Now uses proper scoring method
- `internal/integration/habit_configuration_changes_test.go` - NEW: Integration tests for config changes
- `internal/scoring/engine_test.go:TestEngine_ScoreSimpleHabit()` - NEW: Unit tests for simple habit scoring
- `internal/models/habit.go` - Habit type definitions and validation methods (IsSimple, IsElastic, etc.)
- `internal/models/entry.go` - Entry data structures and type handling
- `internal/parser/` - YAML parsing and schema validation  
- `internal/ui/habitconfig/` - Habit editing workflows (future Phase 2 enhancement target)
- User data: `/home/david/.config/vice/habits.yml` and `entries.yml`

## 1. Habit / User Story

As a user, I should be able to modify any aspect of my habit configurations (field types, scoring methods, habit types, criteria) and have the system gracefully handle these changes without breaking existing functionality or producing confusing error messages. The system should be resilient to configuration changes and provide clear guidance when data migration or cleanup is needed.

## 2. Problem Analysis

### Immediate Issue: Simple Automatic Scoring Failure

**Steps to Reproduce:**
1. Have a simple boolean habit with manual scoring
2. Edit the habit to change field type to numeric and scoring type to automatic
3. Add criteria for automatic scoring (e.g., greater_than condition)
4. Try to enter data for today's entry for that habit

**Expected Behavior:**
- Entry collection should work normally for the updated habit
- Automatic scoring should evaluate the numeric input against the criteria
- Habit should be scored as completed/failed based on the criteria

**Actual Behavior:**
- Error: "failed to collect entry for habit do_10_push_ups: automatic scoring failed: scoring failed: habit do_10_push_ups is not an elastic habit"
- Entry collection fails completely

### Broader Data Type Resilience Issues

**Potential Problem Areas:**
1. **Scoring Logic:** Hard-coded assumptions about habit type vs scoring type relationships
2. **Entry Validation:** May not handle field type changes gracefully
3. **Historical Data:** Old entry data with different types may conflict with new habit configuration
4. **Habit Validation:** May not prevent invalid combinations or guide users through transitions
5. **UI Workflows:** Habit editing may not warn about potential data compatibility issues

## 3. Technical Analysis

### Root Cause Analysis

**Primary Issue:** The scoring system has incorrect assumptions about habit type vs scoring type relationships.

**Data Evidence:**
```yaml
# Current habit configuration (correct)
- title: Do 10 push-ups
  id: do_10_push_ups
  habit_type: simple          # Simple habit, not elastic
  field_type:
    type: unsigned_int
  scoring_type: automatic    # Automatic scoring with criteria
  criteria:
    condition:
      greater_than: 10.0

# Historical entry data (shows previous boolean values)
- habit_id: do_10_push_ups
  value: false              # Old boolean data from when it was manual
  achievement_level: none
```

### Architectural Vulnerabilities

**1. Type System Assumptions:**
- Hard-coded logic assuming automatic scoring = elastic habits
- Missing validation for valid habit type + scoring type combinations
- Lack of type safety between habit configuration and entry data

**2. Data Migration Gaps:**
- No automatic handling of field type changes
- Historical entry data may have incompatible types
- No validation or cleanup of legacy data when habits change

**3. Configuration Change Handling:**
- Habit editing doesn't validate downstream impacts
- No warning system for breaking configuration changes
- Entry collection assumes static habit configurations

**4. Error Handling:**
- Misleading error messages that reference wrong habit types
- No graceful degradation when configuration conflicts occur
- Missing context about what caused the compatibility issue

## 4. Acceptance Criteria

### Immediate Bug Fixes
- [ ] Simple habits with automatic scoring should work correctly
- [ ] Entry collection should accept numeric input for converted habits
- [ ] Automatic scoring should evaluate criteria properly for simple habits
- [ ] Error messages should be accurate and specific to the actual issue

### Data Type Resilience Improvements
- [ ] System should handle all valid habit type + scoring type + field type combinations
- [ ] Historical entry data with different types should not break current operations
- [ ] Habit configuration changes should not require manual data cleanup
- [ ] Entry validation should adapt to current habit configuration, not historical data

### Configuration Change Management
- [ ] Habit editing should validate compatibility of new configurations
- [ ] Warning system for configuration changes that may affect existing data
- [ ] Graceful handling of incompatible historical entry data
- [ ] Clear guidance when manual intervention is needed

### Systematic Improvements
- [ ] Remove hard-coded assumptions about habit type relationships
- [ ] Implement proper type validation throughout the system
- [ ] Add comprehensive error handling with actionable messages
- [ ] Create data migration strategies for common configuration changes

### Future-Proofing
- [ ] System should be extensible for new habit types and scoring methods
- [ ] Configuration changes should be backward compatible where possible
- [ ] Entry data should be forward compatible with habit configuration evolution
- [ ] Comprehensive testing for all habit type + scoring + field type combinations

## 5. Implementation Plan & Progress

**Overall Status:** `In Progress`

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

### Phase 1: Fix Known Scoring Issue (Immediate) ✅ COMPLETE
**Focus:** Address the specific scoring bug for simple habits with automatic scoring

- [x] **Sub-task 1.1:** Investigate scoring logic assumptions
  - *Design:* Analyze scoring system to identify hard-coded habit type assumptions
  - *Code/Artifacts:* `internal/scoring/` - Find logic that restricts automatic scoring to elastic habits
  - *Testing Strategy:* Create test cases for all habit type + scoring type combinations
  - *AI Notes:* ✅ Found root cause: `SimpleHabitCollectionFlow.performAutomaticScoring` tries to fake simple habits as elastic habits by copying habit and setting MiniCriteria, then calls `ScoreElasticHabit()`. But `ScoreElasticHabit()` checks `habit.IsElastic()` on original habit type, causing "not an elastic habit" error. Design flaw: simple habits with automatic scoring need their own scoring method, not elastic habit masquerading.

- [x] **Sub-task 1.2:** Fix automatic scoring for simple habits
  - *Design:* Remove habit type restrictions from automatic scoring logic
  - *Code/Artifacts:* Update scoring validation to allow simple habits with automatic scoring
  - *Testing Strategy:* Test simple boolean→numeric habit conversion with automatic scoring
  - *AI Notes:* ✅ Added `ScoreSimpleHabit()` method to scoring engine that properly handles simple habits with automatic scoring. Updated `SimpleHabitCollectionFlow.performAutomaticScoring()` to use the new method instead of faking elastic habits. Fixed existing tests that incorrectly used elastic habits for simple habit testing. Added comprehensive tests for the new scoring method covering numeric and boolean habits plus error cases.

- [x] **Sub-task 1.3:** Improve error messages for scoring failures
  - *Design:* Replace misleading error messages with accurate, actionable feedback
  - *Code/Artifacts:* Update error handling in scoring and entry collection
  - *Testing Strategy:* Verify error messages accurately describe actual issues
  - *AI Notes:* ✅ Error messages are now accurate with the core fix. The new `ScoreSimpleHabit()` method provides clear, specific error messages: "habit X is not a simple habit", "does not require automatic scoring", "has no criteria for automatic scoring". The misleading "not an elastic habit" error is eliminated since simple habits now use their own scoring method.

- [x] **Sub-task 1.4:** Test habit configuration change workflows
  - *Design:* Comprehensive testing of habit editing followed by entry collection
  - *Code/Artifacts:* Integration tests covering habit type/field type/scoring changes
  - *Testing Strategy:* Test all realistic habit conversion scenarios
  - *AI Notes:* ✅ Created `internal/integration/habit_configuration_changes_test.go` with comprehensive tests covering: boolean→numeric automatic scoring (user's exact scenario), manual→automatic scoring conversions for different field types, and verification that both simple and elastic habits work correctly with automatic scoring. All tests pass, confirming the fix resolves the reported issue and prevents similar problems.

### Phase 2: Type Assumption Audit and Architectural Resilience Plan (Next)
**Focus:** Systematic audit and planning for type assumption vulnerabilities revealed by T016

**Objective:** Create comprehensive plan to identify and address type masquerading patterns and hard-coded assumptions across the system

**Planning Activity: T017 - Type Assumption Audit and Data Resilience Architecture Plan**

**Scope:**
1. **Codebase Audit**: Search for type masquerading patterns and hard-coded assumptions
2. **Architecture Analysis**: Map configuration change impacts across all systems  
3. **Testing Strategy**: Design comprehensive type compatibility testing
4. **User Experience Review**: Identify validation and warning opportunities in habit editing flows

**Key Investigation Areas:**
- Habit validation logic for type assumption patterns
- Entry parsing and validation systems  
- UI workflow configuration change handling
- Error message accuracy across type operations
- Data migration and backward compatibility strategies

**Audit Methodologies (from Sub-task 2.1):**
1. **Type Masquerading Pattern Detection**: Search for components creating fake instances of other types (e.g., SimpleHabitCollectionFlow faking elastic habits)
2. **Hard-coded Type Assumption Analysis**: Identify conditional logic with missing type combinations (e.g., automatic scoring assumed = elastic only)
3. **Cross-Component Type Dependency Mapping**: Trace data flow between components with implicit type assumptions
4. **Error Message Accuracy Audit**: Validate error messages match actual failure conditions

**Priority Target Areas:**
- **High**: Scoring system, entry collection flows, habit validation
- **Medium**: Parser/YAML handling, UI habit configuration  
- **Lower**: Data persistence and historical data compatibility

**Search Patterns:**
- Type field manipulation + method calls: `rg -A 5 -B 5 "(habit|entry).*\.(habit_type|field_type|scoring_type).*="`
- Missing case coverage: `rg -A 10 -B 5 "IsElastic\(\)" | rg -v "IsSimple"`
- Hard-coded assumptions: `rg "automatic.*elastic|elastic.*automatic"`
- Error message issues: `rg -A 3 -B 3 "not.*elastic|not.*simple|invalid.*habit"`

**Deliverables:**
1. Vulnerability assessment with prioritized findings
2. Architecture improvement recommendations  
3. Comprehensive testing strategy for type resilience
4. Implementation roadmap with effort estimates

**Success Criteria:** Clear plan to prevent similar type assumption bugs across all habit configuration change scenarios

- [x] **Sub-task 2.1:** Preliminary research - audit methods and target areas
  - *Design:* Research systematic approaches for identifying type assumptions and architectural vulnerabilities
  - *Code/Artifacts:* Document audit methodologies (static analysis, pattern matching, dependency mapping)
  - *Testing Strategy:* Define search patterns and validation approaches for type assumption detection
  - *AI Notes:* ✅ Defined 4 audit methodologies: 1) Type masquerading pattern detection (Component A faking as Component B), 2) Hard-coded type assumption analysis (missing case coverage), 3) Cross-component type dependency mapping (interface boundary assumptions), 4) Error message accuracy audit. Prioritized target areas: scoring system, entry collection flows, habit validation (high); parser/YAML handling, UI habit config (medium); data persistence (lower). Created specific search patterns and validation approaches for each methodology.

- [x] **Sub-task 2.2:** Discovery - apply audit methods to codebase
  - *Design:* Execute systematic search for type masquerading patterns, hard-coded assumptions, validation gaps
  - *Code/Artifacts:* Raw findings from codebase analysis using defined methods
  - *Testing Strategy:* Validate findings with targeted tests to confirm vulnerabilities
  - *AI Notes:* ✅ Systematic audit complete. Found no active type masquerading patterns (T016 fix eliminated main issue). Identified key vulnerabilities: 1) Overly restrictive validation (checklist habits → checklist field types only), 2) Hard-coded type assumptions in habit validation logic (different paths for Simple vs Elastic), 3) Missing case coverage (informational habit scoring, some field type combinations), 4) Limited extensibility for new habit types. Scoring engine properly separated post-T016. Flow factory shows clean architecture. Error messages now accurate.
---
● Discovery Summary: Raw Findings

  Based on systematic audit using the defined methodologies, I've identified several categories of type assumptions and potential vulnerabilities:

  1. Type Masquerading Patterns (MODERATE RISK)

  - No active patterns found: The T016 fix eliminated the main type masquerading issue where SimpleHabitCollectionFlow was faking simple habits as elastic habits
  - Scoring engine separation: Both ScoreSimpleHabit() and ScoreElasticHabit() have proper type validation preventing cross-type usage

  2. Hard-coded Type Assumptions (HIGH IMPACT)

  Validation Logic Restrictions:
  - Checklist habits → Checklist field type only: internal/models/habit.go:222-235 prevents checklist habits from using other field types
  - Habit type → Scoring requirements: Different validation paths for Simple vs Elastic habits assume different capabilities
  - Field type → Habit type coupling: Some field types implicitly tied to specific habit types

  Flow Factory Pattern:
  - Clean architecture: internal/ui/entry/flow_factory.go uses proper factory pattern without type assumptions
  - Good separation: Each habit type gets its own flow without cross-type dependencies

  3. Cross-Component Type Dependencies (LOW RISK)

  - Scoring engine interface: Proper separation between ScoreSimpleHabit() and ScoreElasticHabit() methods
  - Entry collection flows: Clean routing based on habit types without assumptions

  4. Error Message Accuracy (RESOLVED)

  - Accurate messages: Post-T016 fix, error messages correctly identify actual issues
  - Type-specific errors: Each scoring method provides appropriate error context

  5. Missing Case Coverage (MEDIUM RISK)

  Identified Gaps:
  - Informational habit scoring: No explicit handling for automatic scoring in informational habits
  - Field type combinations: Some field types may not work with all habit types despite logical compatibility
  - Configuration validation: Limited checking of field type + scoring type compatibility

  Key Architectural Vulnerabilities:

  1. Overly Restrictive Validation: Checklist habits forced to use checklist field types only
  2. Implicit Type Coupling: Field types with built-in assumptions about their usage context
  3. Missing Flexibility: Validation logic prevents potentially valid combinations
  4. Limited Extensibility: Adding new habit types requires multiple validation updates

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
  - *Design:* TBD - Warning and validation system for habit editing
  - *Code/Artifacts:* TBD - Habit editing workflow enhancements
  - *Testing Strategy:* TBD - User experience testing for configuration changes
  - *AI Notes:* Focus on preventing user confusion and data loss during changes

## 6. Roadblocks

*(None currently)*

## 7. Notes / Discussion Log

- `2025-07-14 - User:` Reported bug when editing habit from simple boolean to automatic numeric scoring: "habit do_10_push_ups is not an elastic habit"
- `2025-07-14 - User:` Requested generalization to look for opportunities to improve resilience to changed data types elsewhere
- `2025-07-14 - AI:` Expanded scope from simple bug fix to comprehensive data type resilience. Issue reveals broader architectural vulnerabilities: hard-coded type assumptions, missing data migration handling, inadequate configuration change validation, and poor error messaging. This affects scoring, entry collection, habit editing, and validation systems.
- `2025-07-14 - AI:` Phase 1 COMPLETE: Fixed the immediate scoring issue. Root cause was `SimpleHabitCollectionFlow.performAutomaticScoring()` trying to masquerade simple habits as elastic habits, which failed because `ScoreElasticHabit()` validates the original habit type. Solution: Added dedicated `ScoreSimpleHabit()` method for simple habits with automatic scoring. Updated flow to use proper method. Added comprehensive tests. User's scenario (boolean→numeric automatic scoring) now works correctly. No more "not an elastic habit" errors for simple habits.

### Key Learnings & Implementation Guidance:

**1. Architecture Pattern Discovered:**
- Problem: Type masquerading anti-pattern where SimpleHabitCollectionFlow tried to fake simple habits as elastic habits
- Solution: Dedicated scoring methods per habit type with proper type validation
- Principle: Each habit type should have its own appropriate handling, not try to reuse others' logic

**2. Critical Code Paths:**
- `internal/scoring/engine.go:ScoreSimpleHabit()` - New method for simple habit automatic scoring
- `internal/ui/entry/flow_implementations.go:performAutomaticScoring()` - Fixed to use proper method
- Integration point: Entry collection → Flow routing → Scoring engine method selection

**3. Testing Strategy Learned:**
- Unit tests alone insufficient - integration tests covering real user scenarios crucial
- Test configuration changes, not just individual components
- Verify error messages are user-friendly and accurate

**4. Data Type Conversion Insights:**
- Historical entry data (boolean values) doesn't interfere with new habit configuration (numeric)
- System correctly adapts to current habit configuration during entry collection
- No manual data cleanup needed for habit type transitions

**5. Future Phase 2 Guidance:**
- Look for similar type assumption patterns in: habit validation, entry parsing, UI workflows
- Consider implementing configuration change warnings in habit editing UI
- Audit for other places where components might be masquerading as different types