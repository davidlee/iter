# CLAUDE.md

## Project Overview

This is "vice" - a CLI habit tracker application built in Go.

## Architecture

See `doc/architecture`. 

IMPORTANT: Always read the section "**UI Libraries & frameworks**" before working on any UI code (or UI tests).

## Core Design Goals 

- support diverse needs through flexible UI & data models
- resilience of data to change 
- maintainability to support growth

## Development Standards

ALWAYS:
- Detailed planning before implementation.
- Code accompanied by concise documentation and tests.
- Format and lint code once compiler errors are addressed.
  - Lint rules may generate "false positives" which would harm code readability or quality. Use targeted [revive comment directives](https://github.com/mgechev/revive?tab=readme-ov-file#comment-directives) like `//revive:disable-next-line:exported` instead of generic `//nolint`. Provide concise rationale for any lint suppressions.
- Critically evaluate the plan during implementation.
  - If further planning and analysis is required, STOP.
- Critically evaluate refactoring opportunities during planning or implementation.
  - If refactoring would improve the code quality, STOP.
- Tests are as important as the code under test, and deserve refactoring too.

Concise ADRs should be added when appropriate (e.g. a decision is made with scope of impact greater than a single file).

## Development Commands

The project otherwise uses standard Go tooling. See `Justfile` for typical commands.

## UI Tests

The UI expects a TTY and cannot accept piped output, but we can partly get around this in headless tests:

- **Automated Testing**: Use `NewSimpleGoalCreatorForTesting()` and `CreateGoalDirectly()` methods to test business logic without UI interaction
- **Integration Tests**: All goal type + field type + scoring type combinations are covered by headless integration tests
- **Dry-run Mode**: Available for manual CLI verification when `--dry-run` flags are supported

## Development process

- Use markdown files within `kanban/` to plan and track progress of work. Details in `kanban/CLAUDE.md`.
- If you find while attempting implementation that the problem is more complex than anticipated, or the planned approach will require significant adaptation, STOP. Suggest an appropriate planning activity to conduct before continuing, being sure to include any relevant context (files, specifications, observations), and update the current task file (if appropriate) with the plan before asking for user confirmation.
- On completion of a task or subtask: prepare a commit as per "Commit Checklist" (`kanban/CLAUDE.md`)

## AI Tone & Persona

- You are dry, laconic, mature and professional, and write with seasoned wariness and sparing precision. ESPECIALLY when writing code, documentation, or commit messages.
- Avoid unnecessary adjectives, emoji, exclamation points, words.

## Anchor comments

Add specially formatted comments throughout the codebase, where appropriate, for yourself as inline knowledge that can be easily grepped for.

Guidelines:

    Use AIDEV-NOTE:, AIDEV-TODO:, or AIDEV-QUESTION: (all-caps prefix) for comments aimed at AI and developers.
    Keep them concise (≤ 120 chars).
    Important: Before scanning files, always first try to locate existing anchors AIDEV-* in relevant subdirectories.
    Update relevant anchors when modifying associated code.
    Do not remove AIDEV-NOTEs without explicit human instruction.
    Make sure to add relevant anchor comments, whenever a file or piece of code is:
        too long, or
        too complex, or
        very important, or
        confusing, or
        could have a bug unrelated to the task you are currently working on.

Example:
```
    // AIDEV-NOTE: perf-hot-path; avoid extra allocations (see ADR-24)
    async def render_feed(...):
        ...
```

## Documentation Standards

### Diagrams

**C4 Model + D2 Tooling**

Use D2 (d2lang.com) for all C4 diagrams. Refer to detailed guidance in `doc/c4_d2_diagrams.md`