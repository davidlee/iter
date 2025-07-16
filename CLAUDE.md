# CLAUDE.md

## Project Overview

This is "vice" - a CLI habit tracker application built in Go.

## Architecture

See `doc/architecture.md` (header doc) and other more focused `doc/` files.
See `doc/bubbletea_guide.md` for guidance on UI code or tests.

IMPORTANT: find and read relevant docs before modifying, planning or debugging UI code or tests.

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

## Documentation Standards

### Diagrams

ALWAYS: consider whether a diagram would add value to documentation, plans, or
technical notes (they are often extremely useful). 

**C4 Model Diagrams**

Use D2 (d2lang.com) for all C4 diagrams. Refer to detailed guidance in `doc/c4_d2_diagrams.md`