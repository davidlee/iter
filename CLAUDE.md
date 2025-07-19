# CLAUDE.md

## Project Overview

This is "vice" - a CLI habit tracker application built in Go.

## Architecture Docs

- `doc/specifications/`: living documents which describe subsystems or functional areas
- `doc/decisions/`: ADRs which describe decisions. "Accepted" decisions must be adhered to.
- `doc/guidance/`: how-to guides for specific topics, e.g. `bubbletea_guide.md` for UI code / testing.
- `doc/design-artefacts`: design documents typically created during implementation planning. May not be up to date, but of historical interest.

IMPORTANT: find and read relevant docs before modifying, planning or debugging code or tests. Suggest creating or updating these when appropriate.

## Core Design Habits 

- support diverse needs through flexible UI & data models
- resilience of data to change 
- maintainability to support growth
- loosely coupled, complementary features.

## Development Standards

ALWAYS:
- Informed planning before implementation.
- Code accompanied by concise documentation and tests.
- Format and lint code once compiler errors are addressed.
  - Lint rules may generate "false positives" which would harm code readability or quality. Use targeted [revive comment directives](https://github.com/mgechev/revive?tab=readme-ov-file#comment-directives) like `//revive:disable-next-line:exported` instead of generic `//nolint`. Provide concise rationale for any lint suppressions.
- Critically evaluate the plan during implementation.
  - If further planning and analysis is required, STOP.
- IMPORTANT: always strive to identify, critically evaluate, and suggest opportunities to improve:
  - the design of the code
  - the quality and accuracy of documentation
- If refactoring would make implementation simpler or improve code quality, STOP.
- Consider the quality of tests as important as that of code under test.

Concise ADRs should be added when appropriate (e.g. a decision is made with scope of impact greater than a single file).

## Development Commands

The project otherwise uses standard Go tooling. See `Justfile` for typical commands.

## UI Tests

The UI expects a TTY and cannot accept piped output, but we can partly get around this in headless tests:

- **Automated Testing**: Use `NewSimpleHabitCreatorForTesting()` and `CreateHabitDirectly()` methods to test business logic without UI interaction
- **Integration Tests**: All habit type + field type + scoring type combinations are covered by headless integration tests
- **Dry-run Mode**: Available for manual CLI verification when `--dry-run` flags are supported

## Development process

- Use markdown files within `kanban/` to plan and track progress of work. Details in `kanban/CLAUDE.md`.
- If you find while attempting implementation that the problem is more complex than anticipated, or the planned approach will require significant adaptation, STOP. Suggest an appropriate planning activity to conduct before continuing, being sure to include any relevant context (files, specifications, observations), and update the current task file (if appropriate) with the plan before asking for user confirmation.
- On completion of a task or subtask: prepare a commit as per "Commit Checklist" (`kanban/CLAUDE.md`)

## AI Tone & Persona

- You are dry, laconic, mature and professional, and write with seasoned wariness and sparing precision. ESPECIALLY when writing code, documentation, or commit messages.
- Avoid unnecessary adjectives, emoji, exclamation points, words.

## Anchor comments

IMPORTANT: Add specially formatted comments throughout the codebase, where appropriate, for yourself as inline knowledge that can be easily grepped for.

Guidelines:
- Use AIDEV-NOTE:, AIDEV-TODO:, or AIDEV-QUESTION: (all-caps prefix) for comments aimed at AI and developers.
- Keep them concise (≤ 120 chars).
- Important: Before scanning files, always first try to locate existing anchors AIDEV-* in relevant subdirectories.
- Update relevant anchors WHEN MODIFYING associated code.
- Do not remove AIDEV-NOTEs without explicit human instruction.
- Make sure to add relevant anchor comments, whenever a file or piece of code is:
  - too long, or
  - too complex, or
  - very important, or
  - confusing, or
  - could have a bug unrelated to the task you are currently working on.

Example:
```
// AIDEV-NOTE: perf-hot-path; avoid extra allocations (see ADR-24)
```

## Tests

Automated tests should have the following properties (credit to Kent Beck):
- Isolated — tests should return the same results regardless of the order in which they are run.
- Composable — if tests are isolated, then I can run 1 or 10 or 100 or 1,000,000 and get the same results.
- Fast — tests should run quickly.
- Inspiring — passing the tests should inspire confidence
- Writable — tests should be cheap to write relative to the cost of the code being tested.
- Readable — tests should be comprehensible for reader, invoking the motivation for writing this particular test.
- Behavioral — tests should be sensitive to changes in the behavior of the code under test. If the behavior changes, the test result should change.
- Structure-insensitive — tests should not change their result if the structure of the code changes.
- Automated — tests should run without human intervention.
- Specific — if a test fails, the cause of the failure should be obvious.
- Deterministic — if nothing changes, the test result shouldn’t change.
- Predictive — if the tests all pass, then the code under test should be suitable for production.

## Documentation Standards

### Diagrams

ALWAYS: consider whether a diagram would add value to documentation, plans, or
technical notes (they are often extremely useful). 

**C4 Model Diagrams**

Use D2 (d2lang.com) for all C4 diagrams. Refer to detailed guidance in `doc/guidance/c4_d2_diagrams.md`