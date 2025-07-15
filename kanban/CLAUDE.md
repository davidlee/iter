# Workflow 

<!-- AIDEV-NOTE: Read this file FIRST before any kanban/ operations - contains task ID assignment rules and workflow --> 

## Roles & Collaboration Model

This document outlines the collaboration model between the User (acting as Tech Lead) and the AI (Claude, acting as a senior developer/pair programmer). We will use a DIY Kanban system based on folders and Markdown files, managed with Git.

- Tech Lead (User) Responsibilities:
  - Define project architecture, goals and priorities.
  - Select which task to work on, create or refine.
  - Create initial task Markdown files or request AI to draft them.
  - Review and approve or modify implementation plans proposed by the AI within task files.
  - Review agent's work.
  - Answer agent's questions and resolve roadblocks.

- AI (Claude) Responsibilities:
  - Draft new task file in `kanban/backlog` (as per `kanban/backlog/T000_example.md`) on request
  - Refine selected tasks by proposing detailed "Implementation Plan & Progress" sections within the task's Markdown file. This includes sub-tasks, design considerations (e.g., function signatures, data models), and testing strategies.
  - Work only on tasks selected by the User.
  - Generate code, documentation, tests, or other artifacts as per the agreed sub-tasks.
  - Update the status of sub-tasks (`[ ]`, `[WIP]`, `[X]` (done), `[blocked]`) within the task's Markdown file.
  - Clearly indicate when a roadblock is encountered by marking the relevant sub-task `[blocked]` and adding a note in the "Roadblocks" section of the task file.
  - Always stop and wait for user input after modifying a task's Markdown file content (especially the plan or sub-task status) or when a stopping condition is met.
  - When submitting a task (or subtask) for user review, prepare a commit as explained below. 
  - Update task files with commit ids.

## Kanban System: Folders & Files

The project work is managed with "cards" within `kanban/` and subdirectories: 

```
  - kanban/
    - backlog/
    - in-progress/
    - in-review/
    - done/
    - archive/
```

**ALWAYS** read [[kanban/CLAUDE.md]] before any (read or write) operations on `kanban/`.

Each task is a single Markdown file (e.g., `T023_implement_option_parser.md`), where T023 is the task's unique ID (also stored in the file's frontmatter).

Tasks MUST adhere closely to the format of `../kanban/backlog/T000_example.md`.

## Task Dependencies & Relationships

Tasks often have relationships with each other that should be explicitly documented to ensure proper sequencing and coordination. These relationships include:

- **Dependency Types**: `blocks`, `depends-on`, `related`, `part-of`, `overlaps`

- Managing Dependencies:
  - Dependencies should be included in the task's Markdown frontmatter in the related_tasks field
  - The relationship type should be specified using the format: ["depends-on:T123", "blocks:T456"]
  - Dependencies should be mentioned in the task's Discussion Log when they impact progress
  - When a task or sub-task is blocked by a dependency, it should be marked as ["blocked"]

- Visualizing Dependencies:
  - Use a simple text diagram in the task file when complex dependencies exist
  - Example: T001 --> T002 --> T003

## AI Interaction with Task Markdown:

- When planning, AI will propose content for "3. Implementation Plan & Progress".
- When working, AI will update sub-task statuses (`[ ]`, `[WIP]`, `[x]`, `[blocked]`) in section 3.
- AI will add entries to "4. Roadblocks" and "5. Notes / Discussion Log".
- AI may edit task files in place, in which case it will always print the verbatim content of the changes and surrounding context to the user.

## Commit Checklist

When asked to "commit" or when work is complete on a task or subtask, follow the following checklist:
- Review the work completed since last commit (e.g. the current subtask), and
  - [ ] add any relevant notes to the Notes section of the task file, especially noting any unexpected issues encountered, human feedback, refactoring done, or key decisions made.
  - [ ] review files touched and important decisions / changes made, and add any Anchor Comments (per CLAUDE.md) as appropriate to aid future work.
  - [ ] format the code
  - [ ] lint the code and address any issues
  - [ ] stage all changes in the working directory
  - [ ] review changes to be committed; split into multiple cohesive commits if appropriate
  - [ ] perform commit(s) with commit message per convention below 
  - [ ] add commit hash to the task file for future reference.
  - [ ] note the next logical subtask to proceed with, if any

### Commit Conventions

- **Commit Conventions:**
  - ALL commits have a one line message (in conventional format) 
  - **Commit message title** (first line)
    - Format: `type(scope): [TASK-ID/SUBTASK-ID] description`
      - Type: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
      - Scope: Component or module affected (optional)
      - Description: Concise explanation in present tense
    - Examples:
      - `feat(auth)[T123/1.1]: implement login form`
      - `fix(api)[T234]: correct response format in user service`
      - `docs(readme)[T567/2.3]: update installation instructions`

### Core Workflow

**No Code Before Approved Plan:**
- You must not begin implementing a (sub)task if the "Implementation Plan &
  Progress" section for that task/sub-task has not been filled out and
  approved by the User.

**Implementation Phase (Sub-task by Sub-task):**
1. Update sub-task 1.1 status to `[WIP]` in the Markdown.
2. Focus on sub-task 1.1: ask clarifying questions if needed, check plan & context is clear and appropriate. If not, STOP.
3. When satisfied, generate code, documentation, test cases, etc., as per the design.
3. Update sub-task 1.1 status to `[x]` (done) in the Markdown.
4. Update "Overall Status" if a major phase is complete.
5. **STOP.** "Sub-task 1.1 is complete. `T123.md` has been updated. Commit with message `feat(type):[T123/1.1] brief description`?
- **User**: Reviews work, tests, confirms commit or requests changes.

**Handling Roadblocks:**
- AI (during a sub-task): "I've encountered a roadblock on sub-task X.Y: [description]."
- AI: Updates sub-task X.Y status to `[blocked]` in the Markdown. Adds details to section "Roadblocks".
- AI: **STOPS.** "Roadblock encountered and noted in `T123.md`. Here is the updated content: ... Please advise."
- User: Resolves roadblock, provides guidance. Updates task.md if necessary (e.g., unblocks sub-task, modifies plan). Commits. "Okay, you can now proceed with sub-task X.Y."

**Task Completion:**
- (after the last sub-task): "All sub-tasks for `T123.md` are complete. The overall status is now `[done]`."
- Update the task.md file accordingly.
- **STOP.** "`T123.md` is now complete. Suggested commit: `feat(T123): Complete [Task Title]`. Should I move it to DONE and commit?"

**Stopping Conditions (AI must stop and wait for User):**
- After modifying the "Implementation Plan & Progress" section in a task.md, except logging progress or observations.
- After marking any sub-task as `[WIP]`, `[x]` (done), or `[blocked]` and providing the associated output/update.
- When a roadblock is identified (once documented).
- After completing all sub-tasks for a main task.