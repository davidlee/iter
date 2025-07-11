---
title: "Implement Elastic Goals End-to-End (Mini/Midi/Maxi)"
type: ["feature"]
tags: ["elastic", "goals", "ui", "parser", "scoring"]
related_tasks: ["depends-on:T001"]
context_windows: ["./CLAUDE.md", "./doc/specifications/goal_schema.md", "./internal/models/*.go", "./internal/parser/*.go", "./internal/ui/*.go"]
---

# Implement Elastic Goals End-to-End (Mini/Midi/Maxi)

## 1. Goal / User Story

As a user, I want to track elastic goals with mini/midi/maxi achievement levels so that I can set ambitious targets while still celebrating partial progress. This allows for more nuanced habit tracking where I can define minimum, target, and stretch goals for activities like exercise duration, reading time, or other measurable habits.

The system should allow me to:
- Define elastic goals in goals.yml with three achievement levels (mini/midi/maxi)
- Record values during entry and see automatic scoring based on achievement levels
- View which level I achieved for each elastic goal
- Support both manual scoring and automatic criteria-based scoring

This builds upon the boolean goal foundation from T001 to provide more sophisticated goal tracking capabilities.

## 2. Acceptance Criteria

- [ ] User can define elastic goals in goals.yml with mini/midi/maxi criteria
- [ ] Parser supports elastic goal types with proper validation
- [ ] UI presents appropriate input types for elastic goal field types (numeric, duration, etc.)
- [ ] Automatic scoring evaluates input against mini/midi/maxi criteria
- [ ] Manual scoring allows user to select achievement level during entry
- [ ] Entry storage preserves both raw values and achievement levels
- [ ] UI displays achievement level results (none/mini/midi/maxi) clearly
- [ ] Code maintains existing quality standards (formatted, linted, tested)
- [ ] Backwards compatibility with existing boolean goals maintained

---
## 3. Implementation Plan & Progress

**Overall Status:** `In Progress`

**Sub-tasks:**

- [x] **1. Model Extensions**: Extend goal and entry models for elastic goals
    - [x] **1.1 Update Goal model for elastic criteria**
        - *Design:* Add MiniCriteria, MidiCriteria, MaxiCriteria fields to Goal struct
        - *Code/Artifacts to be created or modified:* `internal/models/goal.go`, tests
        - *Testing Strategy:* Unit tests for elastic goal validation and criteria parsing
        - *AI Notes:* Completed - elastic criteria fields were already present, added validation logic and helper methods
    - [x] **1.2 Update Entry model for achievement levels**
        - *Design:* Add AchievementLevel field to GoalEntry, support "none"/"mini"/"midi"/"maxi"
        - *Code/Artifacts to be created or modified:* `internal/models/entry.go`, tests
        - *Testing Strategy:* Unit tests for achievement level serialization and validation
        - *AI Notes:* Completed - added AchievementLevel type, validation, helper methods, and convenience functions

- [ ] **2. Parser Enhancements**: Support elastic goals in YAML parsing
    - [ ] **2.1 Extend YAML parsing for elastic goal structure**
        - *Design:* Parse mini_criteria, midi_criteria, maxi_criteria from YAML
        - *Code/Artifacts to be created or modified:* `internal/parser/goals.go`, tests
        - *Testing Strategy:* Unit tests with sample elastic goal YAML configurations
        - *AI Notes:* Validate that criteria are properly ordered (mini ≤ midi ≤ maxi for numeric types)
    - [ ] **2.2 Add validation for elastic goal consistency**
        - *Design:* Ensure criteria make logical sense (e.g., mini < midi < maxi for "higher is better")
        - *Code/Artifacts to be created or modified:* `internal/parser/goals.go`, validation functions
        - *Testing Strategy:* Unit tests for invalid criteria combinations
        - *AI Notes:* Consider different field types may have different validation rules

- [ ] **3. Scoring Engine**: Implement automatic scoring for elastic goals
    - [ ] **3.1 Create scoring engine for criteria evaluation**
        - *Design:* ScoreEngine that evaluates values against elastic criteria
        - *Code/Artifacts to be created or modified:* `internal/scoring/engine.go` (new package)
        - *Testing Strategy:* Comprehensive unit tests for different field types and criteria
        - *AI Notes:* Should handle numeric, duration, time field types with appropriate operators
    - [ ] **3.2 Integrate scoring with entry collection**
        - *Design:* Automatic scoring during entry creation, fallback to manual for complex cases
        - *Code/Artifacts to be created or modified:* `internal/ui/entry.go`, scoring integration
        - *Testing Strategy:* Unit tests for scoring integration, edge cases
        - *AI Notes:* User should see what level they achieved immediately after input

- [ ] **4. UI Enhancements**: Update CLI interface for elastic goals
    - [ ] **4.1 Add elastic goal input handling**
        - *Design:* Different input prompts based on field types (numeric with units, duration formats)
        - *Code/Artifacts to be created or modified:* `internal/ui/entry.go`, form builders
        - *Testing Strategy:* Manual testing of different elastic goal types, unit tests for logic
        - *AI Notes:* Show criteria thresholds to user during input for motivation
    - [ ] **4.2 Display achievement results clearly**
        - *Design:* Show achievement level with appropriate styling (colors, emojis for levels)
        - *Code/Artifacts to be created or modified:* `internal/ui/entry.go`, result display
        - *Testing Strategy:* Manual testing of different achievement scenarios
        - *AI Notes:* Use lipgloss styling to make achievement levels visually distinct

- [ ] **5. Storage Updates**: Ensure proper elastic goal entry storage
    - [ ] **5.1 Update entry storage for achievement levels**
        - *Design:* Store both raw values and computed achievement levels in entries.yml
        - *Code/Artifacts to be created or modified:* `internal/storage/entries.go`
        - *Testing Strategy:* Unit tests for elastic entry serialization/deserialization
        - *AI Notes:* Achievement level should be recomputable from raw value + criteria for audit
    - [ ] **5.2 Add sample elastic goals to file initialization**
        - *Design:* Include 1-2 elastic goal examples in sample goals.yml
        - *Code/Artifacts to be created or modified:* `internal/init/files.go`
        - *Testing Strategy:* Verify sample elastic goals parse and validate correctly
        - *AI Notes:* Good examples: exercise duration, reading time, water intake

- [ ] **6. Integration & Testing**: Ensure elastic goals work end-to-end
    - [ ] **6.1 End-to-end testing with elastic goals**
        - *Design:* Test complete workflow: define elastic goal → enter value → see achievement
        - *Code/Artifacts to be created or modified:* Integration tests, manual testing
        - *Testing Strategy:* Test automatic scoring, manual scoring, edge cases
        - *AI Notes:* Verify backwards compatibility with existing boolean goals
    - [ ] **6.2 Code quality and documentation**
        - *Design:* Ensure all new code meets project standards
        - *Code/Artifacts to be created or modified:* Code formatting, linting fixes, documentation
        - *Testing Strategy:* Run full test suite, linting, formatting checks
        - *AI Notes:* Update goal schema documentation with elastic goal examples

## 4. Roadblocks

*(No roadblocks identified yet)*

## 5. Notes / Discussion Log

- `2025-07-11 - User:` Requested implementation of elastic goals with mini/midi/maxi achievement levels
- `2025-07-11 - AI:` Created comprehensive task breakdown building on T001 foundation, focusing on scoring engine and UI enhancements for multi-level achievements
- `2025-07-11 - AI:` Subtask 1.1 completed - Updated Goal model with elastic goal validation. Added validation for required criteria fields when using automatic scoring, plus helper methods (IsElastic, RequiresAutomaticScoring, etc.). Added comprehensive unit tests for elastic goal validation and helper methods. All tests pass, no linting issues.
- `2025-07-11 - AI:` Subtask 1.2 completed - Updated Entry model with achievement levels. Added AchievementLevel type with constants (none/mini/midi/maxi), AchievementLevel field to GoalEntry struct, validation for achievement levels, helper methods (GetAchievementLevel, SetAchievementLevel, HasAchievementLevel, ClearAchievementLevel), and convenience functions (CreateElasticGoalEntry, CreateValueOnlyGoalEntry). Added 13 new unit tests covering all achievement level functionality. All tests pass, code properly formatted, no linting issues.

## 6. Code Snippets & Artifacts 

*(Generated content will be placed here during implementation)*